package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type BaggageStatus struct {
	ID          uuid.UUID `db:"id" json:"id"`
	PassengerID uuid.UUID `db:"passenger_id" json:"passenger_id"`
	Status      string    `db:"status" json:"status"`
	Location    string    `db:"location" json:"location"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}

type Repository interface {
	Upsert(ctx context.Context, bag *BaggageStatus) error
	GetByID(ctx context.Context, id uuid.UUID) (*BaggageStatus, error)
}

type postgresRepo struct {
	db *sqlx.DB
}

func NewPostgresRepo(db *sqlx.DB) Repository {
	return &postgresRepo{db: db}
}

func (r *postgresRepo) Upsert(ctx context.Context, bag *BaggageStatus) error {
	query := `
		INSERT INTO baggage_tracking (id, passenger_id, status, location, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (id) DO UPDATE SET
			status = EXCLUDED.status,
			location = EXCLUDED.location,
			updated_at = EXCLUDED.updated_at
	`
	_, err := r.db.ExecContext(ctx, query, bag.ID, bag.PassengerID, bag.Status, bag.Location, bag.UpdatedAt)
	return err
}

func (r *postgresRepo) GetByID(ctx context.Context, id uuid.UUID) (*BaggageStatus, error) {
	var bag BaggageStatus
	query := `SELECT id, passenger_id, status, location, updated_at FROM baggage_tracking WHERE id=$1`
	err := r.db.GetContext(ctx, &bag, query, id)
	if err != nil {
		return nil, err
	}
	return &bag, nil
}
