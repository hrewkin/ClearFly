package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// NotificationPublisher publishes notification events to the message bus.
//
// The booking service uses it to notify passengers about successful PNR
// confirmation. The publisher is intentionally tiny so it can be replaced
// with a no-op in tests or unit-test scenarios.
type NotificationPublisher interface {
	Publish(ctx context.Context, evt NotificationEvent) error
}

// NotificationEvent is the payload published to the notification_events
// queue. It is intentionally generic so a single notification service can
// fan it out to push/sms/email channels.
type NotificationEvent struct {
	Type        string                 `json:"type"`
	PassengerID string                 `json:"passenger_id"`
	FlightID    string                 `json:"flight_id,omitempty"`
	Title       string                 `json:"title"`
	Content     string                 `json:"content"`
	Channels    []string               `json:"channels"`
	Meta        map[string]interface{} `json:"meta,omitempty"`
}

// BookingFlowService coordinates seat selection, tariff lookup, atomic seat
// blocking and PNR generation when a passenger books a flight.
type BookingFlowService interface {
	BookSeat(ctx context.Context, flightID, passengerID, seatID uuid.UUID) (*Booking, error)
	ListBookings(ctx context.Context, passengerID uuid.UUID) ([]Booking, error)
	ListBookingsByFlight(ctx context.Context, flightID uuid.UUID) ([]Booking, error)
	CheckIn(ctx context.Context, id uuid.UUID) (*Booking, error)
	CancelBooking(ctx context.Context, id uuid.UUID, reason, actor string) (*Booking, error)
}

type bookingFlow struct {
	bookings  BookingRepository
	flights   FlightRepository
	publisher NotificationPublisher
}

// NewBookingFlowService wires repositories and the notification publisher
// (publisher may be nil — events will be silently dropped).
func NewBookingFlowService(b BookingRepository, f FlightRepository, p NotificationPublisher) BookingFlowService {
	return &bookingFlow{bookings: b, flights: f, publisher: p}
}

// BookSeat performs the full booking flow:
//  1. Look up the seat to discover its class.
//  2. Look up the tariff for that class on the flight.
//  3. Compute price = base_price × load_factor coefficient.
//  4. Atomically block the seat (FOR UPDATE inside flight repo).
//  5. Persist the booking with PNR/price/payment_status=PAID.
//  6. Publish a BOOKING_CONFIRMED notification (best-effort).
func (s *bookingFlow) BookSeat(ctx context.Context, flightID, passengerID, seatID uuid.UUID) (*Booking, error) {
	flight, err := s.flights.GetFlightByID(ctx, flightID)
	if err != nil {
		return nil, ErrFlightNotFound
	}

	seat, err := s.flights.GetSeatByID(ctx, seatID)
	if err != nil {
		return nil, ErrSeatNotAvailable
	}
	if seat.FlightID != flightID {
		return nil, ErrSeatNotAvailable
	}

	tariff, err := s.flights.GetTariff(ctx, flightID, seat.Class)
	if err != nil {
		return nil, ErrTariffNotFound
	}

	price := tariff.BasePrice
	if flight.TotalSeats > 0 {
		taken := flight.TotalSeats - flight.AvailableSeats
		loadFactor := float64(taken) / float64(flight.TotalSeats)
		switch {
		case loadFactor > 0.8:
			price *= 1.5
		case loadFactor > 0.5:
			price *= 1.2
		}
	}

	booking := &Booking{
		ID:            uuid.New(),
		FlightID:      flightID,
		PassengerID:   passengerID,
		Status:        "CONFIRMED",
		PNRCode:       GeneratePNR(),
		SeatID:        &seatID,
		Price:         price,
		Currency:      tariff.Currency,
		PaymentStatus: "PAID",
		CreatedAt:     time.Now().UTC(),
	}

	if err := s.flights.BlockSeat(ctx, seatID, booking.ID); err != nil {
		return nil, err
	}

	if err := s.bookings.Create(ctx, booking); err != nil {
		return nil, err
	}

	if s.publisher != nil {
		_ = s.publisher.Publish(ctx, NotificationEvent{
			Type:        "BOOKING_CONFIRMED",
			PassengerID: passengerID.String(),
			FlightID:    flightID.String(),
			Title:       "Бронирование подтверждено",
			Content:     "Ваш PNR: " + booking.PNRCode + ". Рейс " + flight.FlightNumber + " " + flight.Origin + " → " + flight.Destination + ". Место " + seat.SeatNumber + ".",
			Channels:    []string{"PUSH", "EMAIL"},
			Meta: map[string]interface{}{
				"pnr":         booking.PNRCode,
				"seat_number": seat.SeatNumber,
			},
		})
	}

	return booking, nil
}

func (s *bookingFlow) ListBookings(ctx context.Context, passengerID uuid.UUID) ([]Booking, error) {
	return s.bookings.ListByPassenger(ctx, passengerID)
}

func (s *bookingFlow) ListBookingsByFlight(ctx context.Context, flightID uuid.UUID) ([]Booking, error) {
	return s.bookings.ListByFlight(ctx, flightID)
}

// CancelBooking sets the booking status to CANCELLED, releases the
// associated seat (if any) and publishes a notification. It is safe to call
// on an already-cancelled booking — the operation is a no-op in that case.
//
// `actor` is a free-form label ("passenger", "staff:EMP-001", "admin")
// stored in the notification meta and helpful for debugging.
func (s *bookingFlow) CancelBooking(ctx context.Context, id uuid.UUID, reason, actor string) (*Booking, error) {
	b, err := s.bookings.GetByID(ctx, id)
	if err != nil {
		return nil, ErrBookingNotFound
	}
	if b.Status == "CANCELLED" {
		return b, nil
	}

	if b.SeatID != nil {
		if err := s.flights.ReleaseSeat(ctx, *b.SeatID); err != nil {
			return nil, err
		}
	}
	if err := s.bookings.UpdateStatus(ctx, b.ID, "CANCELLED"); err != nil {
		return nil, err
	}
	if err := s.bookings.UpdatePaymentStatus(ctx, b.ID, "REFUNDED"); err != nil {
		return nil, err
	}
	b.Status = "CANCELLED"
	b.PaymentStatus = "REFUNDED"

	if s.publisher != nil {
		flight, _ := s.flights.GetFlightByID(ctx, b.FlightID)
		title := "Бронирование отменено"
		content := "Возврат оформлен по PNR " + b.PNRCode + "."
		if flight != nil {
			content = "Возврат оформлен по PNR " + b.PNRCode + ". Рейс " + flight.FlightNumber + " " + flight.Origin + " → " + flight.Destination + "."
		}
		_ = s.publisher.Publish(ctx, NotificationEvent{
			Type:        "BOOKING_CANCELLED",
			PassengerID: b.PassengerID.String(),
			FlightID:    b.FlightID.String(),
			Title:       title,
			Content:     content,
			Channels:    []string{"PUSH", "EMAIL"},
			Meta: map[string]interface{}{
				"pnr":    b.PNRCode,
				"reason": reason,
				"actor":  actor,
			},
		})
	}

	return b, nil
}

func (s *bookingFlow) CheckIn(ctx context.Context, id uuid.UUID) (*Booking, error) {
	b, err := s.bookings.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	b.Status = "CHECKED_IN"
	if err := s.bookings.UpdateStatus(ctx, id, "CHECKED_IN"); err != nil {
		return nil, err
	}
	if err := s.bookings.UpdatePaymentStatus(ctx, id, "PAID"); err != nil {
		return nil, err
	}
	if s.publisher != nil {
		_ = s.publisher.Publish(ctx, NotificationEvent{
			Type:        "CHECKED_IN",
			PassengerID: b.PassengerID.String(),
			FlightID:    b.FlightID.String(),
			Title:       "Регистрация на рейс выполнена",
			Content:     "Ваш посадочный талон готов. PNR: " + b.PNRCode + ".",
			Channels:    []string{"PUSH", "EMAIL"},
		})
	}
	return b, nil
}
