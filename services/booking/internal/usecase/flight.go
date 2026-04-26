package usecase

import (
	"context"
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
	BlockSeat(ctx context.Context, seatID uuid.UUID, bookingID uuid.UUID) error

	CreateTariff(ctx context.Context, t *Tariff) error
	GetTariff(ctx context.Context, flightID uuid.UUID, class string) (*Tariff, error)
}

// FlightService defines business logic for flight management.
type FlightService interface {
	CreateFlight(ctx context.Context, f *Flight) (*Flight, error)
	GetFlight(ctx context.Context, id uuid.UUID) (*Flight, error)
	SearchFlights(ctx context.Context, origin, destination, date string) ([]Flight, error)
	UpdateFlightStatus(ctx context.Context, id uuid.UUID, status, gate string) error
	GetAvailableSeats(ctx context.Context, flightID uuid.UUID) ([]Seat, error)
	CreateTariff(ctx context.Context, flightID uuid.UUID, class string, price float64, currency string) (*Tariff, error)
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
