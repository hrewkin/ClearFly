package repository

import (
	"context"

	"github.com/cleanair/passenger/internal/usecase"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type postgresPassengerRepo struct {
	db *sqlx.DB
}

// NewPostgresPassengerRepo creates a new PostgreSQL-backed passenger repository.
func NewPostgresPassengerRepo(db *sqlx.DB) usecase.PassengerRepository {
	return &postgresPassengerRepo{db: db}
}

const passengerColumns = `id, name, email, phone, passport_number,
	COALESCE(loyalty_tier, 'STANDARD') AS loyalty_tier,
	COALESCE(loyalty_points, 0) AS loyalty_points,
	COALESCE(meal_preference, 'STANDARD') AS meal_preference,
	COALESCE(special_needs, '') AS special_needs`

func (r *postgresPassengerRepo) Create(ctx context.Context, p *usecase.Passenger) error {
	if p.LoyaltyTier == "" {
		p.LoyaltyTier = "STANDARD"
	}
	if p.MealPreference == "" {
		p.MealPreference = "STANDARD"
	}
	query := `INSERT INTO passengers (id, name, email, phone, passport_number,
	          loyalty_tier, loyalty_points, meal_preference, special_needs)
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	_, err := r.db.ExecContext(ctx, query,
		p.ID, p.Name, p.Email, p.Phone, p.PassportNumber,
		p.LoyaltyTier, p.LoyaltyPoints, p.MealPreference, p.SpecialNeeds)
	return err
}

func (r *postgresPassengerRepo) GetByID(ctx context.Context, id uuid.UUID) (*usecase.Passenger, error) {
	var p usecase.Passenger
	query := `SELECT ` + passengerColumns + ` FROM passengers WHERE id=$1`
	err := r.db.GetContext(ctx, &p, query, id)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *postgresPassengerRepo) Update(ctx context.Context, p *usecase.Passenger) error {
	if p.LoyaltyTier == "" {
		p.LoyaltyTier = "STANDARD"
	}
	if p.MealPreference == "" {
		p.MealPreference = "STANDARD"
	}
	query := `UPDATE passengers SET
	          name=$2, email=$3, phone=$4, passport_number=$5,
	          loyalty_tier=$6, loyalty_points=$7, meal_preference=$8, special_needs=$9
	          WHERE id=$1`
	_, err := r.db.ExecContext(ctx, query,
		p.ID, p.Name, p.Email, p.Phone, p.PassportNumber,
		p.LoyaltyTier, p.LoyaltyPoints, p.MealPreference, p.SpecialNeeds)
	return err
}

func (r *postgresPassengerRepo) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM passengers WHERE id=$1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}
