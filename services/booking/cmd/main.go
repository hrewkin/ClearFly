// cmd/main.go
package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/cleanair/booking/internal/delivery"
	"github.com/cleanair/booking/internal/repository"
	"github.com/cleanair/booking/internal/usecase"
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
		log.Fatalf("cannot connect to db: %v", err)
	}
	defer db.Close()

	if err := initSchema(db); err != nil {
		log.Fatalf("schema init failed: %v", err)
	}

	bookingRepo := repository.NewPostgresBookingRepo(db)
	flightRepo := repository.NewPostgresFlightRepo(db)

	bookingSvc := usecase.NewBookingService(bookingRepo)
	flightSvc := usecase.NewFlightService(flightRepo)

	rabbitURL := os.Getenv("RABBITMQ_URL")
	if rabbitURL == "" {
		rabbitURL = "amqp://guest:guest@rabbitmq:5672/"
	}
	publisher, perr := delivery.NewRabbitPublisher(rabbitURL)
	if perr != nil {
		log.Printf("rabbit publisher unavailable: %v (continuing without notifications)", perr)
	} else {
		defer publisher.Close()
	}

	var pub usecase.NotificationPublisher
	if publisher != nil {
		pub = publisher
	}

	flowSvc := usecase.NewBookingFlowService(bookingRepo, flightRepo, pub)

	if err := seedDemoData(context.Background(), flightSvc); err != nil {
		log.Printf("seed warning: %v", err)
	}

	r := gin.Default()
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "booking"})
	})

	bookingHandler := delivery.NewHandler(bookingSvc)
	flightHandler := delivery.NewFlightHandler(flightSvc)
	flowHandler := delivery.NewFlowHandler(flowSvc)

	// Classic stub bookings (kept for compatibility with old clients/tests)
	r.POST("/bookings", bookingHandler.CreateBooking)
	r.GET("/bookings/:id", bookingHandler.GetBooking)

	// Seat-aware booking flow
	r.POST("/bookings/book", flowHandler.BookSeat)
	r.POST("/bookings/:id/checkin", flowHandler.CheckIn)
	r.GET("/bookings/passenger/:id", flowHandler.ListByPassenger)

	// Flight management
	r.POST("/flights", flightHandler.CreateFlight)
	r.GET("/flights/search", flightHandler.SearchFlights)
	r.GET("/flights/upcoming", flightHandler.UpcomingFlights)
	r.GET("/flights/:id", flightHandler.GetFlight)
	r.PATCH("/flights/:id/status", flightHandler.UpdateFlightStatus)
	r.GET("/flights/:id/seats", flightHandler.GetSeats)
	r.GET("/flights/:id/tariffs", flightHandler.ListTariffs)
	r.POST("/flights/:id/tariff", flightHandler.CreateTariff)

	log.Println("Booking service listening on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}

func initSchema(db *sqlx.DB) error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS flights (
			id UUID PRIMARY KEY,
			flight_number VARCHAR(20) NOT NULL,
			origin VARCHAR(10) NOT NULL,
			destination VARCHAR(10) NOT NULL,
			departure_time TIMESTAMP NOT NULL,
			arrival_time TIMESTAMP NOT NULL,
			aircraft_type VARCHAR(40) NOT NULL DEFAULT '',
			total_seats INT NOT NULL DEFAULT 0,
			available_seats INT NOT NULL DEFAULT 0,
			gate VARCHAR(10) NOT NULL DEFAULT '',
			status VARCHAR(20) NOT NULL DEFAULT 'SCHEDULED'
		);`,
		`CREATE TABLE IF NOT EXISTS seats (
			id UUID PRIMARY KEY,
			flight_id UUID NOT NULL REFERENCES flights(id) ON DELETE CASCADE,
			seat_number VARCHAR(5) NOT NULL,
			class VARCHAR(20) NOT NULL DEFAULT 'ECONOMY',
			status VARCHAR(20) NOT NULL DEFAULT 'AVAILABLE',
			booking_id UUID,
			UNIQUE(flight_id, seat_number)
		);`,
		`CREATE TABLE IF NOT EXISTS tariffs (
			id UUID PRIMARY KEY,
			flight_id UUID NOT NULL REFERENCES flights(id) ON DELETE CASCADE,
			class VARCHAR(20) NOT NULL,
			base_price DOUBLE PRECISION NOT NULL,
			currency VARCHAR(3) NOT NULL DEFAULT 'RUB',
			UNIQUE(flight_id, class)
		);`,
		`CREATE TABLE IF NOT EXISTS bookings (
			id UUID PRIMARY KEY,
			flight_id UUID NOT NULL,
			passenger_id UUID NOT NULL,
			status VARCHAR(50) NOT NULL,
			pnr_code VARCHAR(10),
			seat_id UUID,
			price DOUBLE PRECISION DEFAULT 0,
			currency VARCHAR(3) DEFAULT 'RUB',
			payment_status VARCHAR(20) DEFAULT 'PAID',
			created_at TIMESTAMP DEFAULT NOW()
		);`,
		`CREATE UNIQUE INDEX IF NOT EXISTS bookings_pnr_unique ON bookings(pnr_code) WHERE pnr_code IS NOT NULL AND pnr_code <> '';`,
	}
	for _, s := range statements {
		if _, err := db.Exec(s); err != nil {
			return err
		}
	}
	// Add columns to legacy bookings table if migrating from earlier schema.
	migrations := []string{
		`ALTER TABLE bookings ADD COLUMN IF NOT EXISTS pnr_code VARCHAR(10);`,
		`ALTER TABLE bookings ADD COLUMN IF NOT EXISTS seat_id UUID;`,
		`ALTER TABLE bookings ADD COLUMN IF NOT EXISTS price DOUBLE PRECISION DEFAULT 0;`,
		`ALTER TABLE bookings ADD COLUMN IF NOT EXISTS currency VARCHAR(3) DEFAULT 'RUB';`,
		`ALTER TABLE bookings ADD COLUMN IF NOT EXISTS payment_status VARCHAR(20) DEFAULT 'PAID';`,
		`ALTER TABLE bookings ADD COLUMN IF NOT EXISTS created_at TIMESTAMP DEFAULT NOW();`,
	}
	for _, s := range migrations {
		if _, err := db.Exec(s); err != nil {
			log.Printf("migration warning (%s): %v", s, err)
		}
	}
	return nil
}

// seedDemoData populates the database with a handful of flights with seats
// and tariffs so the WebUI demo "just works" out of the box.
//
// Seeding is idempotent based on flight_number + departure_date.
func seedDemoData(ctx context.Context, svc usecase.FlightService) error {
	type demo struct {
		number     string
		origin     string
		destination string
		dep         time.Duration
		arr         time.Duration
		aircraft    string
		seats       int
		gate        string
		eco         float64
		biz         float64
	}
	now := time.Now().UTC().Truncate(time.Hour)
	// Anchor demo flights at 09:00 UTC. If today's 09:00 is already in the
	// past, roll the schedule forward to tomorrow so the demo always has
	// upcoming flights to show.
	day := time.Date(now.Year(), now.Month(), now.Day(), 9, 0, 0, 0, time.UTC)
	if now.After(day.Add(8 * time.Hour)) {
		day = day.Add(24 * time.Hour)
	}
	demos := []demo{
		{"CN101", "SVO", "LED", 0, 1*time.Hour + 30*time.Minute, "Boeing 737-800", 30, "B7", 5500, 14000},
		{"CN204", "SVO", "AER", 3 * time.Hour, 5*time.Hour + 30*time.Minute, "Airbus A320neo", 30, "C2", 8200, 22000},
		{"CN318", "LED", "KJA", 5 * time.Hour, 9*time.Hour + 15*time.Minute, "Sukhoi Superjet 100", 30, "A4", 9300, 25000},
		{"CN401", "SVO", "KZN", 7 * time.Hour, 8*time.Hour + 40*time.Minute, "Boeing 737-800", 30, "D9", 4900, 12500},
		{"CN512", "AER", "SVO", 9 * time.Hour, 11*time.Hour + 30*time.Minute, "Airbus A320neo", 30, "F1", 7800, 20000},
	}

	for _, d := range demos {
		// Skip if a flight with this number on the same calendar date already exists.
		existing, err := svc.SearchFlights(ctx, d.origin, d.destination, day.Format("2006-01-02"))
		if err == nil {
			alreadySeeded := false
			for _, f := range existing {
				if f.FlightNumber == d.number {
					alreadySeeded = true
					break
				}
			}
			if alreadySeeded {
				continue
			}
		}

		f := &usecase.Flight{
			FlightNumber:  d.number,
			Origin:        d.origin,
			Destination:   d.destination,
			DepartureTime: day.Add(d.dep),
			ArrivalTime:   day.Add(d.arr),
			AircraftType:  d.aircraft,
			TotalSeats:    d.seats,
			Gate:          d.gate,
		}
		created, err := svc.CreateFlight(ctx, f)
		if err != nil {
			return err
		}

		if err := svc.CreateSeatLayout(ctx, created.ID, created.TotalSeats); err != nil {
			return err
		}

		if _, err := svc.CreateTariff(ctx, created.ID, "ECONOMY", d.eco, "RUB"); err != nil {
			return err
		}
		if _, err := svc.CreateTariff(ctx, created.ID, "BUSINESS", d.biz, "RUB"); err != nil {
			return err
		}
	}
	return nil
}
