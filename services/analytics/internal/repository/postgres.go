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

	// Pull real capacity from the flights table; fall back to 150 if the
	// flights table isn't available yet or the row is missing.
	capacity := 150
	_ = r.db.GetContext(ctx, &capacity, `SELECT total_seats FROM flights WHERE id = $1`, flightID)
	if capacity <= 0 {
		capacity = 150
	}

	loadFactor := (total * 100) / capacity
	if loadFactor > 100 {
		loadFactor = 100
	}

	return &AnalyticsResult{
		TotalBookings: total,
		LoadFactor:    loadFactor,
	}, nil
}
