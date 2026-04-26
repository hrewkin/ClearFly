package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Flight represents an airline flight.
type Flight struct {
	ID             uuid.UUID `json:"id" db:"id"`
	FlightNumber   string    `json:"flight_number" db:"flight_number"`
	Origin         string    `json:"origin" db:"origin"`
	Destination    string    `json:"destination" db:"destination"`
	DepartureTime  time.Time `json:"departure_time" db:"departure_time"`
	ArrivalTime    time.Time `json:"arrival_time" db:"arrival_time"`
	AircraftType   string    `json:"aircraft_type" db:"aircraft_type"`
	TotalSeats     int       `json:"total_seats" db:"total_seats"`
	AvailableSeats int       `json:"available_seats" db:"available_seats"`
	Gate           string    `json:"gate" db:"gate"`
	Status         string    `json:"status" db:"status"` // SCHEDULED, DELAYED, CANCELLED, DEPARTED, ARRIVED
}

// Seat represents a single seat on a flight.
type Seat struct {
	ID         uuid.UUID  `json:"id" db:"id"`
	FlightID   uuid.UUID  `json:"flight_id" db:"flight_id"`
	SeatNumber string     `json:"seat_number" db:"seat_number"`
	Class      string     `json:"class" db:"class"` // ECONOMY, BUSINESS
	Status     string     `json:"status" db:"status"` // AVAILABLE, BLOCKED, BOOKED
	BookingID  *uuid.UUID `json:"booking_id,omitempty" db:"booking_id"`
}

// Tariff represents pricing rules for a flight class.
type Tariff struct {
	ID        uuid.UUID `json:"id" db:"id"`
	FlightID  uuid.UUID `json:"flight_id" db:"flight_id"`
	Class     string    `json:"class" db:"class"`
	BasePrice float64   `json:"base_price" db:"base_price"`
	Currency  string    `json:"currency" db:"currency"`
}

// FlightRepository defines data access methods for flights and seats.
type FlightRepository interface {
	CreateFlight(ctx context.Context, f *Flight) error
	GetFlightByID(ctx context.Context, id uuid.UUID) (*Flight, error)
	SearchFlights(ctx context.Context, origin, destination, date string) ([]Flight, error)
	UpdateFlightStatus(ctx context.Context, id uuid.UUID, status, gate string) error

	GetAvailableSeats(ctx context.Context, flightID uuid.UUID) ([]Seat, error)
	GetSeatsByFlight(ctx context.Context, flightID uuid.UUID) ([]Seat, error)
	GetSeatByID(ctx context.Context, seatID uuid.UUID) (*Seat, error)
	BlockSeat(ctx context.Context, seatID uuid.UUID, bookingID uuid.UUID) error
	CreateSeats(ctx context.Context, seats []Seat) error
	UpcomingFlights(ctx context.Context, limit int) ([]Flight, error)

	CreateTariff(ctx context.Context, t *Tariff) error
	GetTariff(ctx context.Context, flightID uuid.UUID, class string) (*Tariff, error)
	ListTariffs(ctx context.Context, flightID uuid.UUID) ([]Tariff, error)
}

// FlightService defines business logic for flight management.
type FlightService interface {
	CreateFlight(ctx context.Context, f *Flight) (*Flight, error)
	GetFlight(ctx context.Context, id uuid.UUID) (*Flight, error)
	SearchFlights(ctx context.Context, origin, destination, date string) ([]Flight, error)
	UpdateFlightStatus(ctx context.Context, id uuid.UUID, status, gate string) error
	GetAvailableSeats(ctx context.Context, flightID uuid.UUID) ([]Seat, error)
	GetSeatsByFlight(ctx context.Context, flightID uuid.UUID) ([]Seat, error)
	CreateTariff(ctx context.Context, flightID uuid.UUID, class string, price float64, currency string) (*Tariff, error)
	GetTariff(ctx context.Context, flightID uuid.UUID, class string) (*Tariff, error)
	ListTariffs(ctx context.Context, flightID uuid.UUID) ([]Tariff, error)
	CreateSeatLayout(ctx context.Context, flightID uuid.UUID, totalSeats int) error
	UpcomingFlights(ctx context.Context, limit int) ([]Flight, error)
}

type flightService struct {
	repo FlightRepository
}

func NewFlightService(repo FlightRepository) FlightService {
	return &flightService{repo: repo}
}

func (s *flightService) CreateFlight(ctx context.Context, f *Flight) (*Flight, error) {
	f.ID = uuid.New()
	f.Status = "SCHEDULED"
	f.AvailableSeats = f.TotalSeats
	if err := s.repo.CreateFlight(ctx, f); err != nil {
		return nil, err
	}
	return f, nil
}

func (s *flightService) GetFlight(ctx context.Context, id uuid.UUID) (*Flight, error) {
	return s.repo.GetFlightByID(ctx, id)
}

func (s *flightService) SearchFlights(ctx context.Context, origin, destination, date string) ([]Flight, error) {
	return s.repo.SearchFlights(ctx, origin, destination, date)
}

func (s *flightService) UpdateFlightStatus(ctx context.Context, id uuid.UUID, status, gate string) error {
	return s.repo.UpdateFlightStatus(ctx, id, status, gate)
}

func (s *flightService) GetAvailableSeats(ctx context.Context, flightID uuid.UUID) ([]Seat, error) {
	return s.repo.GetAvailableSeats(ctx, flightID)
}

func (s *flightService) GetSeatsByFlight(ctx context.Context, flightID uuid.UUID) ([]Seat, error) {
	return s.repo.GetSeatsByFlight(ctx, flightID)
}

func (s *flightService) GetTariff(ctx context.Context, flightID uuid.UUID, class string) (*Tariff, error) {
	return s.repo.GetTariff(ctx, flightID, class)
}

func (s *flightService) ListTariffs(ctx context.Context, flightID uuid.UUID) ([]Tariff, error) {
	return s.repo.ListTariffs(ctx, flightID)
}

// CreateSeatLayout generates a deterministic seat layout for the given
// flight: 2 BUSINESS rows (4 seats per row, columns A,B,C,D) followed by
// ECONOMY rows of 6 seats each (A,B,C,D,E,F) until totalSeats is reached.
func (s *flightService) CreateSeatLayout(ctx context.Context, flightID uuid.UUID, totalSeats int) error {
	if totalSeats <= 0 {
		return nil
	}
	const businessRows = 2
	const businessCols = "ABCD"
	const economyCols = "ABCDEF"

	seats := make([]Seat, 0, totalSeats)
	row := 1
	count := 0

	for r := 0; r < businessRows && count < totalSeats; r++ {
		for _, col := range businessCols {
			if count >= totalSeats {
				break
			}
			seats = append(seats, Seat{
				ID:         uuid.New(),
				FlightID:   flightID,
				SeatNumber: seatNumber(row, string(col)),
				Class:      "BUSINESS",
				Status:     "AVAILABLE",
			})
			count++
		}
		row++
	}
	// Skip a row to leave a gap between business and economy.
	row++
	for count < totalSeats {
		for _, col := range economyCols {
			if count >= totalSeats {
				break
			}
			seats = append(seats, Seat{
				ID:         uuid.New(),
				FlightID:   flightID,
				SeatNumber: seatNumber(row, string(col)),
				Class:      "ECONOMY",
				Status:     "AVAILABLE",
			})
			count++
		}
		row++
	}
	return s.repo.CreateSeats(ctx, seats)
}

// UpcomingFlights returns next scheduled flights ordered by departure time.
func (s *flightService) UpcomingFlights(ctx context.Context, limit int) ([]Flight, error) {
	if limit <= 0 {
		limit = 20
	}
	return s.repo.UpcomingFlights(ctx, limit)
}

func seatNumber(row int, col string) string {
	return fmt.Sprintf("%02d%s", row, col)
}

func (s *flightService) CreateTariff(ctx context.Context, flightID uuid.UUID, class string, price float64, currency string) (*Tariff, error) {
	t := &Tariff{
		ID:        uuid.New(),
		FlightID:  flightID,
		Class:     class,
		BasePrice: price,
		Currency:  currency,
	}
	if err := s.repo.CreateTariff(ctx, t); err != nil {
		return nil, err
	}
	return t, nil
}
