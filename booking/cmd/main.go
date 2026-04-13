// cmd/main.go
package main

import (
    "log"
    "os"
    "github.com/gin-gonic/gin"
    "github.com/jmoiron/sqlx"
    _ "github.com/lib/pq"
    "github.com/cleanair/booking/internal/delivery"
    "github.com/cleanair/booking/internal/repository"
    "github.com/cleanair/booking/internal/usecase"
)

func main() {
    // Load environment variables
    dbHost := os.Getenv("DB_HOST")
    dbUser := os.Getenv("DB_USER")
    dbPass := os.Getenv("DB_PASS")
    dbName := os.Getenv("DB_NAME")
    if dbName == "" { dbName = "cleanair" }
    dsn := "postgres://" + dbUser + ":" + dbPass + "@" + dbHost + "/" + dbName + "?sslmode=disable"
    db, err := sqlx.Connect("postgres", dsn)
    if err != nil {
        log.Fatalf("cannot connect to db: %v", err)
    }
    defer db.Close()

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
