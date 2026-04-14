package main

import (
	"log"
	"os"

	"github.com/cleanair/analytics/internal/delivery"
	"github.com/cleanair/analytics/internal/repository"
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
		log.Fatalf("cannot connect to database: %v", err)
	}
	defer db.Close()

	repo := repository.NewPostgresRepo(db)

	r := gin.Default()
	h := delivery.NewHttpHandler(repo)

	r.GET("/analytics/load-factor/:flight_id", h.GetLoadFactor)

	log.Println("Analytics service listening on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}
