package usecase

import (
	"context"
	"crypto/rand"
	"math/big"
	"time"

	"github.com/google/uuid"
)

// Booking represents a single confirmed reservation.
//
// Fields PNRCode, SeatID, Price, Currency, PaymentStatus and CreatedAt are
// optional in the original BookingService.CreateBooking flow (kept for
// backwards compatibility) and required in the extended seat-aware flow used
// by the new flight search / seat selection UI.
type Booking struct {
	ID            uuid.UUID  `json:"id" db:"id"`
	FlightID      uuid.UUID  `json:"flight_id" db:"flight_id"`
	PassengerID   uuid.UUID  `json:"passenger_id" db:"passenger_id"`
	Status        string     `json:"status" db:"status"`
	PNRCode       string     `json:"pnr_code,omitempty" db:"pnr_code"`
	SeatID        *uuid.UUID `json:"seat_id,omitempty" db:"seat_id"`
	Price         float64    `json:"price" db:"price"`
	Currency      string     `json:"currency,omitempty" db:"currency"`
	PaymentStatus string     `json:"payment_status,omitempty" db:"payment_status"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
}

// BookingRepository defines data access methods for bookings.
type BookingRepository interface {
	Create(ctx context.Context, b *Booking) error
	GetByID(ctx context.Context, id uuid.UUID) (*Booking, error)
	GetByPNR(ctx context.Context, pnr string) (*Booking, error)
	ListByPassenger(ctx context.Context, passengerID uuid.UUID) ([]Booking, error)
	UpdatePaymentStatus(ctx context.Context, id uuid.UUID, status string) error
}

// BookingService defines business logic for bookings.
type BookingService interface {
	CreateBooking(ctx context.Context, flightID, passengerID uuid.UUID) (*Booking, error)
	GetBooking(ctx context.Context, id uuid.UUID) (*Booking, error)
}

type bookingService struct {
	repo BookingRepository
}

// NewBookingService creates a new BookingService instance.
func NewBookingService(repo BookingRepository) BookingService {
	return &bookingService{repo: repo}
}

func (s *bookingService) CreateBooking(ctx context.Context, flightID, passengerID uuid.UUID) (*Booking, error) {
	b := &Booking{
		ID:            uuid.New(),
		FlightID:      flightID,
		PassengerID:   passengerID,
		Status:        "CONFIRMED",
		PNRCode:       GeneratePNR(),
		PaymentStatus: "PAID",
		Currency:      "RUB",
		CreatedAt:     time.Now().UTC(),
	}
	if err := s.repo.Create(ctx, b); err != nil {
		return nil, err
	}
	return b, nil
}

func (s *bookingService) GetBooking(ctx context.Context, id uuid.UUID) (*Booking, error) {
	return s.repo.GetByID(ctx, id)
}

// GeneratePNR returns a 6-character upper-case alphanumeric PNR code.
//
// Letters I/O and digits 0/1 are excluded to keep it readable.
func GeneratePNR() string {
	const alphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	b := make([]byte, 6)
	for i := range b {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(alphabet))))
		if err != nil {
			// Fall back to a deterministic char on error. Should never happen.
			b[i] = alphabet[i%len(alphabet)]
			continue
		}
		b[i] = alphabet[n.Int64()]
	}
	return string(b)
}
