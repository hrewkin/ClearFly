package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

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

	handler := usecase.NewIncidentHandler()

	rabbitConsumer, err := consumer.NewRabbitConsumer(rabbitURL, handler)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := rabbitConsumer.Start(ctx); err != nil {
		log.Fatalf("Failed to start RabbitMq consumer: %v", err)
	}

	log.Println("Incident service is up and running. Listening for events...")

	// HTTP server for manual incident creation
	r := gin.Default()
	httpHandler := delivery.NewHttpHandler(handler)
	r.POST("/incidents", httpHandler.CreateIncident)

	// Start HTTP server in a goroutine
	go func() {
		log.Println("Incident HTTP API listening on :8080")
		if err := r.Run(":8080"); err != nil {
			log.Fatalf("failed to run HTTP server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down Incident service...")
}
