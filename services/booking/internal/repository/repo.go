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

// NewPostgresBookingRepo creates a PostgreSQL-backed booking repository.
func NewPostgresBookingRepo(db *sqlx.DB) usecase.BookingRepository {
	return &postgresBookingRepo{db: db}
}

func (r *postgresBookingRepo) Create(ctx context.Context, b *usecase.Booking) error {
	query := `INSERT INTO bookings (id, flight_id, passenger_id, status, pnr_code, seat_id, price, currency, payment_status, created_at)
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`
	_, err := r.db.ExecContext(ctx, query,
		b.ID, b.FlightID, b.PassengerID, b.Status,
		b.PNRCode, b.SeatID, b.Price, b.Currency, b.PaymentStatus, b.CreatedAt)
	return err
}

func (r *postgresBookingRepo) GetByID(ctx context.Context, id uuid.UUID) (*usecase.Booking, error) {
	var b usecase.Booking
	query := `SELECT id, flight_id, passenger_id, status,
	          COALESCE(pnr_code, '') AS pnr_code,
	          seat_id,
	          COALESCE(price, 0) AS price,
	          COALESCE(currency, '') AS currency,
	          COALESCE(payment_status, '') AS payment_status,
	          COALESCE(created_at, NOW()) AS created_at
	          FROM bookings WHERE id=$1`
	err := r.db.GetContext(ctx, &b, query, id)
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *postgresBookingRepo) GetByPNR(ctx context.Context, pnr string) (*usecase.Booking, error) {
	var b usecase.Booking
	query := `SELECT id, flight_id, passenger_id, status,
	          COALESCE(pnr_code, '') AS pnr_code,
	          seat_id,
	          COALESCE(price, 0) AS price,
	          COALESCE(currency, '') AS currency,
	          COALESCE(payment_status, '') AS payment_status,
	          COALESCE(created_at, NOW()) AS created_at
	          FROM bookings WHERE pnr_code=$1`
	err := r.db.GetContext(ctx, &b, query, pnr)
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *postgresBookingRepo) ListByPassenger(ctx context.Context, passengerID uuid.UUID) ([]usecase.Booking, error) {
	var bs []usecase.Booking
	query := `SELECT id, flight_id, passenger_id, status,
	          COALESCE(pnr_code, '') AS pnr_code,
	          seat_id,
	          COALESCE(price, 0) AS price,
	          COALESCE(currency, '') AS currency,
	          COALESCE(payment_status, '') AS payment_status,
	          COALESCE(created_at, NOW()) AS created_at
	          FROM bookings WHERE passenger_id=$1
	          ORDER BY created_at DESC`
	err := r.db.SelectContext(ctx, &bs, query, passengerID)
	if err != nil {
		return nil, err
	}
	return bs, nil
}

func (r *postgresBookingRepo) UpdatePaymentStatus(ctx context.Context, id uuid.UUID, status string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE bookings SET payment_status=$2 WHERE id=$1`, id, status)
	return err
}

func (r *postgresBookingRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE bookings SET status=$2 WHERE id=$1`, id, status)
	return err
}

func (r *postgresBookingRepo) ListByFlight(ctx context.Context, flightID uuid.UUID) ([]usecase.Booking, error) {
	var bs []usecase.Booking
	query := `SELECT id, flight_id, passenger_id, status,
	          COALESCE(pnr_code, '') AS pnr_code,
	          seat_id,
	          COALESCE(price, 0) AS price,
	          COALESCE(currency, '') AS currency,
	          COALESCE(payment_status, '') AS payment_status,
	          COALESCE(created_at, NOW()) AS created_at
	          FROM bookings WHERE flight_id=$1
	          ORDER BY created_at`
	err := r.db.SelectContext(ctx, &bs, query, flightID)
	if err != nil {
		return nil, err
	}
	return bs, nil
}
