package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
)

type mockRepo struct {
	bookings map[uuid.UUID]*Booking
}

func newMockRepo() *mockRepo {
	return &mockRepo{
		bookings: make(map[uuid.UUID]*Booking),
	}
}

func (m *mockRepo) Create(ctx context.Context, b *Booking) error {
	m.bookings[b.ID] = b
	return nil
}

func (m *mockRepo) GetByID(ctx context.Context, id uuid.UUID) (*Booking, error) {
	b, ok := m.bookings[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return b, nil
}

func (m *mockRepo) GetByPNR(ctx context.Context, pnr string) (*Booking, error) {
	for _, b := range m.bookings {
		if b.PNRCode == pnr {
			return b, nil
		}
	}
	return nil, errors.New("not found")
}

func (m *mockRepo) ListByPassenger(ctx context.Context, passengerID uuid.UUID) ([]Booking, error) {
	var out []Booking
	for _, b := range m.bookings {
		if b.PassengerID == passengerID {
			out = append(out, *b)
		}
	}
	return out, nil
}

func (m *mockRepo) UpdatePaymentStatus(ctx context.Context, id uuid.UUID, status string) error {
	b, ok := m.bookings[id]
	if !ok {
		return errors.New("not found")
	}
	b.PaymentStatus = status
	return nil
}

func (m *mockRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	b, ok := m.bookings[id]
	if !ok {
		return errors.New("not found")
	}
	b.Status = status
	return nil
}

func (m *mockRepo) ListByFlight(ctx context.Context, flightID uuid.UUID) ([]Booking, error) {
	var out []Booking
	for _, b := range m.bookings {
		if b.FlightID == flightID {
			out = append(out, *b)
		}
	}
	return out, nil
}

func TestCreateBooking(t *testing.T) {
	repo := newMockRepo()
	svc := NewBookingService(repo)

	flightID := uuid.New()
	passengerID := uuid.New()

	b, err := svc.CreateBooking(context.Background(), flightID, passengerID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if b.FlightID != flightID {
		t.Errorf("expected flightID %s, got %s", flightID, b.FlightID)
	}
	if b.PassengerID != passengerID {
		t.Errorf("expected passengerID %s, got %s", passengerID, b.PassengerID)
	}
	if b.Status != "CONFIRMED" {
		t.Errorf("expected status CONFIRMED, got %s", b.Status)
	}
}

func TestGetBooking(t *testing.T) {
	repo := newMockRepo()
	svc := NewBookingService(repo)

	flightID := uuid.New()
	passengerID := uuid.New()

	b, _ := svc.CreateBooking(context.Background(), flightID, passengerID)

	retrieved, err := svc.GetBooking(context.Background(), b.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if retrieved.ID != b.ID {
		t.Errorf("expected ID %s, got %s", b.ID, retrieved.ID)
	}

	// Test not found
	_, err = svc.GetBooking(context.Background(), uuid.New())
	if err == nil {
		t.Error("expected error for non-existent booking, got nil")
	}
}
