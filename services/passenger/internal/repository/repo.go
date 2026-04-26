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

func (r *postgresPassengerRepo) Create(ctx context.Context, p *usecase.Passenger) error {
	query := `INSERT INTO passengers (id, name, email, phone, passport_number) VALUES ($1, $2, $3, $4, $5)`
	_, err := r.db.ExecContext(ctx, query, p.ID, p.Name, p.Email, p.Phone, p.PassportNumber)
	return err
}

func (r *postgresPassengerRepo) GetByID(ctx context.Context, id uuid.UUID) (*usecase.Passenger, error) {
	var p usecase.Passenger
	query := `SELECT id, name, email, phone, passport_number FROM passengers WHERE id=$1`
	err := r.db.GetContext(ctx, &p, query, id)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *postgresPassengerRepo) Update(ctx context.Context, p *usecase.Passenger) error {
	query := `UPDATE passengers SET name=$2, email=$3, phone=$4, passport_number=$5 WHERE id=$1`
	_, err := r.db.ExecContext(ctx, query, p.ID, p.Name, p.Email, p.Phone, p.PassportNumber)
	return err
}

func (r *postgresPassengerRepo) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM passengers WHERE id=$1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}
