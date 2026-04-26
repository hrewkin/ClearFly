package repository

import (
	"context"

	"github.com/cleanair/booking/internal/usecase"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type postgresFlightRepo struct {
	db *sqlx.DB
}

// NewPostgresFlightRepo creates a new PostgreSQL-backed flight repository.
func NewPostgresFlightRepo(db *sqlx.DB) usecase.FlightRepository {
	return &postgresFlightRepo{db: db}
}

func (r *postgresFlightRepo) CreateFlight(ctx context.Context, f *usecase.Flight) error {
	query := `INSERT INTO flights (id, flight_number, origin, destination, departure_time, arrival_time,
	           aircraft_type, total_seats, available_seats, gate, status)
	           VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`
	_, err := r.db.ExecContext(ctx, query, f.ID, f.FlightNumber, f.Origin, f.Destination,
		f.DepartureTime, f.ArrivalTime, f.AircraftType, f.TotalSeats, f.AvailableSeats, f.Gate, f.Status)
	return err
}

func (r *postgresFlightRepo) GetFlightByID(ctx context.Context, id uuid.UUID) (*usecase.Flight, error) {
	var f usecase.Flight
	query := `SELECT id, flight_number, origin, destination, departure_time, arrival_time,
	           aircraft_type, total_seats, available_seats, gate, status FROM flights WHERE id=$1`
	err := r.db.GetContext(ctx, &f, query, id)
	if err != nil {
		return nil, err
	}
	return &f, nil
}

func (r *postgresFlightRepo) SearchFlights(ctx context.Context, origin, destination, date string) ([]usecase.Flight, error) {
	var flights []usecase.Flight
	query := `SELECT id, flight_number, origin, destination, departure_time, arrival_time,
	           aircraft_type, total_seats, available_seats, gate, status
	           FROM flights WHERE origin=$1 AND destination=$2
	           AND departure_time::date = $3::date AND status='SCHEDULED'
	           ORDER BY departure_time`
	err := r.db.SelectContext(ctx, &flights, query, origin, destination, date)
	if err != nil {
		return nil, err
	}
	return flights, nil
}

func (r *postgresFlightRepo) UpdateFlightStatus(ctx context.Context, id uuid.UUID, status, gate string) error {
	query := `UPDATE flights SET status=$2, gate=$3 WHERE id=$1`
	_, err := r.db.ExecContext(ctx, query, id, status, gate)
	return err
}

func (r *postgresFlightRepo) GetAvailableSeats(ctx context.Context, flightID uuid.UUID) ([]usecase.Seat, error) {
	var seats []usecase.Seat
	query := `SELECT id, flight_id, seat_number, class, status, booking_id
	           FROM seats WHERE flight_id=$1 AND status='AVAILABLE' ORDER BY seat_number`
	err := r.db.SelectContext(ctx, &seats, query, flightID)
	if err != nil {
		return nil, err
	}
	return seats, nil
}

func (r *postgresFlightRepo) BlockSeat(ctx context.Context, seatID uuid.UUID, bookingID uuid.UUID) error {
	// Atomic seat blocking with SELECT FOR UPDATE
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var status string
	err = tx.GetContext(ctx, &status, `SELECT status FROM seats WHERE id=$1 FOR UPDATE`, seatID)
	if err != nil {
		return err
	}
	if status != "AVAILABLE" {
		return usecase.ErrSeatNotAvailable
	}

	_, err = tx.ExecContext(ctx, `UPDATE seats SET status='BOOKED', booking_id=$2 WHERE id=$1`, seatID, bookingID)
	if err != nil {
		return err
	}

	// Decrement available seats on the flight
	_, err = tx.ExecContext(ctx,
		`UPDATE flights SET available_seats = available_seats - 1
		 WHERE id = (SELECT flight_id FROM seats WHERE id=$1)`, seatID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *postgresFlightRepo) CreateTariff(ctx context.Context, t *usecase.Tariff) error {
	query := `INSERT INTO tariffs (id, flight_id, class, base_price, currency)
	           VALUES ($1,$2,$3,$4,$5)
	           ON CONFLICT (flight_id, class) DO UPDATE SET base_price=EXCLUDED.base_price, currency=EXCLUDED.currency`
	_, err := r.db.ExecContext(ctx, query, t.ID, t.FlightID, t.Class, t.BasePrice, t.Currency)
	return err
}

func (r *postgresFlightRepo) GetTariff(ctx context.Context, flightID uuid.UUID, class string) (*usecase.Tariff, error) {
	var t usecase.Tariff
	query := `SELECT id, flight_id, class, base_price, currency FROM tariffs WHERE flight_id=$1 AND class=$2`
	err := r.db.GetContext(ctx, &t, query, flightID, class)
	if err != nil {
		return nil, err
	}
	return &t, nil
}
