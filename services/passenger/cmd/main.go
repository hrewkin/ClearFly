package main

import (
	"github.com/cleanair/passenger/internal/delivery"
	"github.com/cleanair/passenger/internal/repository"
	"github.com/cleanair/passenger/internal/usecase"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"log"
	"os"
)

func main() {
	// Load environment variables
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
		log.Fatalf("cannot connect to db: %v", err)
	}
	defer db.Close()

	// Initialize database schema
	schema := `
    CREATE TABLE IF NOT EXISTS passengers (
        id UUID PRIMARY KEY,
        name VARCHAR(255) NOT NULL,
        email VARCHAR(255) NOT NULL,
        phone VARCHAR(50) NOT NULL,
        passport_number VARCHAR(50) NOT NULL
    );`
	db.MustExec(schema)

	// Repository & service layer
	repo := repository.NewPostgresPassengerRepo(db)
	svc := usecase.NewPassengerService(repo)

	// HTTP layer
	r := gin.Default()
	h := delivery.NewHandler(svc)
	r.POST("/passengers", h.CreatePassenger)
	r.GET("/passengers/:id", h.GetPassenger)
	r.PUT("/passengers/:id", h.UpdatePassenger)
	r.DELETE("/passengers/:id", h.DeletePassenger)

	// Run server
	log.Println("Passenger service listening on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}
