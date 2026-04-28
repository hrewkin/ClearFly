package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// BaggageStatus represents a single baggage tag tracked through the journey.
//
// FlightID is set when the baggage is associated with a specific flight; it
// is nullable in legacy records.
type BaggageStatus struct {
	ID          uuid.UUID  `db:"id" json:"id"`
	PassengerID uuid.UUID  `db:"passenger_id" json:"passenger_id"`
	FlightID    *uuid.UUID `db:"flight_id" json:"flight_id,omitempty"`
	Status      string     `db:"status" json:"status"`
	Location    string     `db:"location" json:"location"`
	UpdatedAt   time.Time  `db:"updated_at" json:"updated_at"`
}

type Repository interface {
	Upsert(ctx context.Context, bag *BaggageStatus) error
	GetByID(ctx context.Context, id uuid.UUID) (*BaggageStatus, error)
	List(ctx context.Context, limit int) ([]BaggageStatus, error)
	ListByPassenger(ctx context.Context, passengerID uuid.UUID) ([]BaggageStatus, error)
	ListByFlight(ctx context.Context, flightID uuid.UUID) ([]BaggageStatus, error)
}

type postgresRepo struct {
	db *sqlx.DB
}

func NewPostgresRepo(db *sqlx.DB) Repository {
	return &postgresRepo{db: db}
}

func (r *postgresRepo) Upsert(ctx context.Context, bag *BaggageStatus) error {
	query := `
		INSERT INTO baggage_tracking (id, passenger_id, flight_id, status, location, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (id) DO UPDATE SET
			flight_id = COALESCE(EXCLUDED.flight_id, baggage_tracking.flight_id),
			status = EXCLUDED.status,
			location = EXCLUDED.location,
			updated_at = EXCLUDED.updated_at
	`
	_, err := r.db.ExecContext(ctx, query, bag.ID, bag.PassengerID, bag.FlightID, bag.Status, bag.Location, bag.UpdatedAt)
	return err
}

func (r *postgresRepo) GetByID(ctx context.Context, id uuid.UUID) (*BaggageStatus, error) {
	var bag BaggageStatus
	query := `SELECT id, passenger_id, flight_id, status, location, updated_at FROM baggage_tracking WHERE id=$1`
	err := r.db.GetContext(ctx, &bag, query, id)
	if err != nil {
		return nil, err
	}
	return &bag, nil
}

func (r *postgresRepo) List(ctx context.Context, limit int) ([]BaggageStatus, error) {
	if limit <= 0 {
		limit = 50
	}
	var bags []BaggageStatus
	query := `SELECT id, passenger_id, flight_id, status, location, updated_at FROM baggage_tracking ORDER BY updated_at DESC LIMIT $1`
	err := r.db.SelectContext(ctx, &bags, query, limit)
	if err != nil {
		return nil, err
	}
	return bags, nil
}

func (r *postgresRepo) ListByFlight(ctx context.Context, flightID uuid.UUID) ([]BaggageStatus, error) {
	var bags []BaggageStatus
	query := `SELECT id, passenger_id, flight_id, status, location, updated_at FROM baggage_tracking WHERE flight_id=$1 ORDER BY updated_at DESC`
	err := r.db.SelectContext(ctx, &bags, query, flightID)
	if err != nil {
		return nil, err
	}
	return bags, nil
}

func (r *postgresRepo) ListByPassenger(ctx context.Context, passengerID uuid.UUID) ([]BaggageStatus, error) {
	var bags []BaggageStatus
	query := `SELECT id, passenger_id, flight_id, status, location, updated_at FROM baggage_tracking WHERE passenger_id=$1 ORDER BY updated_at DESC`
	err := r.db.SelectContext(ctx, &bags, query, passengerID)
	if err != nil {
		return nil, err
	}
	return bags, nil
}
