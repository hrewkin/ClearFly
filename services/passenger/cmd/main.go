package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/cleanair/passenger/internal/audit"
	"github.com/cleanair/passenger/internal/auth"
	"github.com/cleanair/passenger/internal/delivery"
	"github.com/cleanair/passenger/internal/repository"
	"github.com/cleanair/passenger/internal/usecase"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
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

	schema := `
    CREATE TABLE IF NOT EXISTS passengers (
        id UUID PRIMARY KEY,
        name VARCHAR(255) NOT NULL,
        email VARCHAR(255) NOT NULL,
        phone VARCHAR(50) NOT NULL,
        passport_number VARCHAR(50) NOT NULL,
        loyalty_tier VARCHAR(20) DEFAULT 'STANDARD',
        loyalty_points INT DEFAULT 0,
        meal_preference VARCHAR(20) DEFAULT 'STANDARD',
        special_needs VARCHAR(50) DEFAULT ''
    );`
	db.MustExec(schema)
	migrations := []string{
		`ALTER TABLE passengers ADD COLUMN IF NOT EXISTS loyalty_tier VARCHAR(20) DEFAULT 'STANDARD';`,
		`ALTER TABLE passengers ADD COLUMN IF NOT EXISTS loyalty_points INT DEFAULT 0;`,
		`ALTER TABLE passengers ADD COLUMN IF NOT EXISTS meal_preference VARCHAR(20) DEFAULT 'STANDARD';`,
		`ALTER TABLE passengers ADD COLUMN IF NOT EXISTS special_needs VARCHAR(50) DEFAULT '';`,
	}
	for _, m := range migrations {
		if _, err := db.Exec(m); err != nil {
			log.Printf("passenger migration warning: %v", err)
		}
	}

	repo := repository.NewPostgresPassengerRepo(db)
	svc := usecase.NewPassengerService(repo)

	authRepo := auth.NewRepository(db)
	if err := authRepo.Migrate(context.Background()); err != nil {
		log.Fatalf("auth migration failed: %v", err)
	}
	secret := os.Getenv("AUTH_SECRET")
	if secret == "" {
		secret = "clearfly-dev-secret-change-me"
	}
	authSvc := auth.NewService(authRepo, repo, []byte(secret), 24*time.Hour)
	if err := authSvc.EnsureAdmin(context.Background(), "admin", "admin"); err != nil {
		log.Printf("ensure admin warning: %v", err)
	}

	r := gin.Default()
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "passenger"})
	})
	h := delivery.NewHandler(svc)
	r.POST("/passengers", h.CreatePassenger)
	r.GET("/passengers/:id", h.GetPassenger)
	r.PUT("/passengers/:id", h.UpdatePassenger)
	r.PATCH("/passengers/:id/preferences", h.UpdatePreferences)
	r.DELETE("/passengers/:id", h.DeletePassenger)

	authH := delivery.NewAuthHandler(authSvc)
	r.POST("/auth/register", authH.Register)
	r.POST("/auth/register-staff", authH.RegisterStaff)
	r.POST("/auth/login", authH.Login)
	r.GET("/auth/me", authH.Me)

	auditRepo := audit.NewRepository(db)
	if err := auditRepo.Migrate(context.Background()); err != nil {
		log.Fatalf("audit migration failed: %v", err)
	}
	staffH := delivery.NewStaffHandler(authSvc, auditRepo, os.Getenv("BOOKING_SERVICE_URL"))
	r.POST("/staff/refund", staffH.Refund)
	r.POST("/staff/bookings/:id/cancel", staffH.SelfCancel)
	r.GET("/staff/audit", staffH.Audit)

	// Run server
	log.Println("Passenger service listening on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}
