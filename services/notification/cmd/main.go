package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cleanair/notification/internal/delivery"
	"github.com/cleanair/notification/internal/usecase"
	"github.com/gin-gonic/gin"
)

func main() {
	store := usecase.NewStore()

	rabbitURL := os.Getenv("RABBITMQ_URL")
	if rabbitURL == "" {
		rabbitURL = "amqp://guest:guest@rabbitmq:5672/"
	}

	// RabbitMQ may not be ready immediately when the container starts —
	// retry a handful of times before giving up. The HTTP API still
	// serves an empty feed in the meantime.
	var consumer *delivery.RabbitConsumer
	var err error
	for attempt := 1; attempt <= 10; attempt++ {
		consumer, err = delivery.NewRabbitConsumer(rabbitURL, store)
		if err == nil {
			break
		}
		log.Printf("notification: rabbit not ready (attempt %d): %v", attempt, err)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		log.Printf("notification: giving up on rabbit, running with HTTP only: %v", err)
	} else {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		if err := consumer.Start(ctx); err != nil {
			log.Printf("notification: consumer start failed: %v", err)
		} else {
			log.Println("notification: rabbit consumer started")
		}
	}

	r := gin.Default()
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "notification"})
	})

	h := delivery.NewHttpHandler(store)
	r.GET("/notifications", h.ListAll)
	r.GET("/notifications/:passenger_id", h.ListByPassenger)
	r.POST("/notifications/:id/read", h.MarkRead)

	go func() {
		log.Println("Notification HTTP API listening on :8080")
		if err := r.Run(":8080"); err != nil {
			log.Fatalf("notification http failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("notification: shutting down")
}
