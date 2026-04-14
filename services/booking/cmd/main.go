// cmd/main.go
package main

import (
	"github.com/cleanair/booking/internal/delivery"
	"github.com/cleanair/booking/internal/repository"
	"github.com/cleanair/booking/internal/usecase"
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
    CREATE TABLE IF NOT EXISTS bookings (
        id UUID PRIMARY KEY,
        flight_id UUID NOT NULL,
        passenger_id UUID NOT NULL,
        status VARCHAR(50) NOT NULL
    );`
	db.MustExec(schema)

	// Repository & service layer
	repo := repository.NewPostgresBookingRepo(db)
	svc := usecase.NewBookingService(repo)

	// HTTP layer
	r := gin.Default()
	h := delivery.NewHandler(svc)
	r.POST("/bookings", h.CreateBooking)
	r.GET("/bookings/:id", h.GetBooking)

	// Run server
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}
