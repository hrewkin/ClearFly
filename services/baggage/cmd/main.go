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
		status VARCHAR(50) NOT NULL,
		location VARCHAR(100) NOT NULL,
		updated_at TIMESTAMP NOT NULL
	);`
	db.MustExec(schema)

	repo := repository.NewPostgresRepo(db)

	rabbitURL := os.Getenv("RABBITMQ_URL")
	if rabbitURL == "" {
		rabbitURL = "amqp://guest:guest@rabbitmq:5672/"
	}

	rabbitConsumer, err := delivery.NewRabbitConsumer(rabbitURL, repo)
	if err != nil {
		log.Fatalf("failed to connect to rabbitmq: %v", err)
	}
	if err := rabbitConsumer.Start(context.Background()); err != nil {
		log.Fatalf("failed to start consumer: %v", err)
	}

	r := gin.Default()
	h := delivery.NewHttpHandler(repo)
	r.GET("/baggage/:id", h.GetBaggage)

	if err := r.Run(":8080"); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}
