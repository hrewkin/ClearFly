package repository

import (
	"context"
	"github.com/cleanair/booking/internal/usecase"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type postgresBookingRepo struct {
	db *sqlx.DB
}

func NewPostgresBookingRepo(db *sqlx.DB) usecase.BookingRepository {
	return &postgresBookingRepo{db: db}
}

func (r *postgresBookingRepo) Create(ctx context.Context, b *usecase.Booking) error {
	query := `INSERT INTO bookings (id, flight_id, passenger_id, status) VALUES ($1, $2, $3, $4)`
	_, err := r.db.ExecContext(ctx, query, b.ID, b.FlightID, b.PassengerID, b.Status)
	return err
}

func (r *postgresBookingRepo) GetByID(ctx context.Context, id uuid.UUID) (*usecase.Booking, error) {
	var b usecase.Booking
	query := `SELECT id, flight_id, passenger_id, status FROM bookings WHERE id=$1`
	err := r.db.GetContext(ctx, &b, query, id)
	if err != nil {
		return nil, err
	}
	return &b, nil
}
