package usecase

import (
	"context"
	"github.com/google/uuid"
)

type Booking struct {
	ID          uuid.UUID `json:"id" db:"id"`
	FlightID    uuid.UUID `json:"flight_id" db:"flight_id"`
	PassengerID uuid.UUID `json:"passenger_id" db:"passenger_id"`
	Status      string    `json:"status" db:"status"`
}

type BookingRepository interface {
	Create(ctx context.Context, b *Booking) error
	GetByID(ctx context.Context, id uuid.UUID) (*Booking, error)
}

type BookingService interface {
	CreateBooking(ctx context.Context, flightID, passengerID uuid.UUID) (*Booking, error)
	GetBooking(ctx context.Context, id uuid.UUID) (*Booking, error)
}

type bookingService struct {
	repo BookingRepository
}

func NewBookingService(repo BookingRepository) BookingService {
	return &bookingService{repo: repo}
}

func (s *bookingService) CreateBooking(ctx context.Context, flightID, passengerID uuid.UUID) (*Booking, error) {
	b := &Booking{
		ID:          uuid.New(),
		FlightID:    flightID,
		PassengerID: passengerID,
		Status:      "CONFIRMED",
	}
	if err := s.repo.Create(ctx, b); err != nil {
		return nil, err
	}
	return b, nil
}

func (s *bookingService) GetBooking(ctx context.Context, id uuid.UUID) (*Booking, error) {
	return s.repo.GetByID(ctx, id)
}
