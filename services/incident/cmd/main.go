package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cleanair/incident/internal/consumer"
	"github.com/cleanair/incident/internal/delivery"
	"github.com/cleanair/incident/internal/usecase"
	"github.com/gin-gonic/gin"
)

func main() {
	rabbitURL := os.Getenv("RABBITMQ_URL")
	if rabbitURL == "" {
		rabbitURL = "amqp://guest:guest@rabbitmq:5672/"
	}

	// Wait for RabbitMQ to be ready before fully wiring publishers/consumers.
	publisher := waitForPublisher(rabbitURL)
	if publisher != nil {
		defer publisher.Close()
	}

	var pub usecase.NotificationPublisher
	if publisher != nil {
		pub = publisher
	}
	handler := usecase.NewIncidentHandler(pub)

	rabbitConsumer, err := consumer.NewRabbitConsumer(rabbitURL, handler)
	if err != nil {
		log.Printf("incident: rabbit consumer unavailable, HTTP-only mode: %v", err)
	} else {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		if err := rabbitConsumer.Start(ctx); err != nil {
			log.Printf("incident: failed to start consumer: %v", err)
		}
	}

	log.Println("Incident service is up and running. Listening for events...")

	r := gin.Default()
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "incident"})
	})

	httpHandler := delivery.NewHttpHandler(handler)
	r.POST("/incidents", httpHandler.CreateIncident)

	go func() {
		log.Println("Incident HTTP API listening on :8080")
		if err := r.Run(":8080"); err != nil {
			log.Fatalf("failed to run HTTP server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down Incident service...")
}

// waitForPublisher retries connecting to RabbitMQ for up to ~30 seconds before
// giving up. Returns nil on failure so the service can still serve HTTP.
func waitForPublisher(rabbitURL string) *consumer.RabbitPublisher {
	for attempt := 1; attempt <= 15; attempt++ {
		p, err := consumer.NewRabbitPublisher(rabbitURL)
		if err == nil {
			return p
		}
		log.Printf("incident: rabbit publisher not ready (attempt %d): %v", attempt, err)
		time.Sleep(2 * time.Second)
	}
	return nil
}
