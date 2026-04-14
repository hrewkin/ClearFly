package repository

import (
	"context"

	"github.com/jmoiron/sqlx"
)

type AnalyticsResult struct {
	TotalBookings int `db:"total_bookings" json:"total_bookings"`
	LoadFactor    int `json:"load_factor"` // Percentage
}

type Repository interface {
	GetFlightLoad(ctx context.Context, flightID string) (*AnalyticsResult, error)
}

type postgresRepo struct {
	db *sqlx.DB
}

func NewPostgresRepo(db *sqlx.DB) Repository {
	return &postgresRepo{db: db}
}

func (r *postgresRepo) GetFlightLoad(ctx context.Context, flightID string) (*AnalyticsResult, error) {
	var total int
	// We count the confirmed bookings for this flight
	query := `SELECT COUNT(*) FROM bookings WHERE flight_id = $1 AND status = 'CONFIRMED'`
	err := r.db.GetContext(ctx, &total, query, flightID)
	if err != nil {
		// table might not exist if no booking was ever made
		return &AnalyticsResult{TotalBookings: 0, LoadFactor: 0}, nil
	}

	// Simplistic load factor formula: capacity = 150
	loadFactor := (total * 100) / 150

	return &AnalyticsResult{
		TotalBookings: total,
		LoadFactor:    loadFactor,
	}, nil
}
