package main

import (
	"context"
	"log"
	"os"

	"github.com/cleanair/baggage/internal/delivery"
	"github.com/cleanair/baggage/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func main() {
	dbHost := os.Getenv("DB_HOST")
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASS")
	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "cleanair"
	}

	dsn := "postgres://" + dbUser + ":" + dbPass + "@" + dbHost + "/" + dbName + "?sslmode=disable"
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		log.Fatalf("cannot connect to postgres: %v", err)
	}
	defer db.Close()

	schema := `
	CREATE TABLE IF NOT EXISTS baggage_tracking (
		id UUID PRIMARY KEY,
		passenger_id UUID NOT NULL,
		flight_id UUID,
		status VARCHAR(50) NOT NULL,
		location VARCHAR(100) NOT NULL,
		updated_at TIMESTAMP NOT NULL
	);`
	db.MustExec(schema)
	if _, err := db.Exec(`ALTER TABLE baggage_tracking ADD COLUMN IF NOT EXISTS flight_id UUID;`); err != nil {
		log.Printf("baggage migration warning: %v", err)
	}

	repo := repository.NewPostgresRepo(db)

	rabbitURL := os.Getenv("RABBITMQ_URL")
	if rabbitURL == "" {
		rabbitURL = "amqp://guest:guest@rabbitmq:5672/"
	}

	rabbitConsumer, err := delivery.NewRabbitConsumer(rabbitURL, repo)
	if err != nil {
		log.Printf("baggage: rabbitmq unavailable, running HTTP-only: %v", err)
	} else if err := rabbitConsumer.Start(context.Background()); err != nil {
		log.Printf("baggage: failed to start consumer: %v", err)
	}

	r := gin.Default()
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "baggage"})
	})
	h := delivery.NewHttpHandler(repo)
	r.GET("/baggage", h.ListBaggage)
	r.POST("/baggage", h.CreateBaggage)
	r.GET("/baggage/:id", h.GetBaggage)
	r.POST("/baggage/:id/scan", h.Scan)

	if err := r.Run(":8080"); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}
