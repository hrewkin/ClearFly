package usecase

import (
	"context"

	"github.com/google/uuid"
)

// Passenger represents a passenger profile in the system.
//
// LoyaltyTier values: STANDARD, SILVER, GOLD, PLATINUM.
// MealPreference values: STANDARD, VEGETARIAN, VEGAN, KOSHER, HALAL,
// GLUTEN_FREE, DIABETIC.
// SpecialNeeds is a free-form short label such as "WHEELCHAIR" or
// "EXTRA_LEGROOM". An empty string means none.
type Passenger struct {
	ID             uuid.UUID `json:"id" db:"id"`
	Name           string    `json:"name" db:"name"`
	Email          string    `json:"email" db:"email"`
	Phone          string    `json:"phone" db:"phone"`
	PassportNumber string    `json:"passport_number" db:"passport_number"`
	LoyaltyTier    string    `json:"loyalty_tier" db:"loyalty_tier"`
	LoyaltyPoints  int       `json:"loyalty_points" db:"loyalty_points"`
	MealPreference string    `json:"meal_preference" db:"meal_preference"`
	SpecialNeeds   string    `json:"special_needs" db:"special_needs"`
}

// PassengerRepository defines data access methods for passengers.
type PassengerRepository interface {
	Create(ctx context.Context, p *Passenger) error
	GetByID(ctx context.Context, id uuid.UUID) (*Passenger, error)
	Update(ctx context.Context, p *Passenger) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// PassengerService defines business logic methods for passenger management.
type PassengerService interface {
	CreatePassenger(ctx context.Context, name, email, phone, passport string) (*Passenger, error)
	GetPassenger(ctx context.Context, id uuid.UUID) (*Passenger, error)
	UpdatePassenger(ctx context.Context, id uuid.UUID, name, email, phone, passport string) (*Passenger, error)
	UpdatePreferences(ctx context.Context, id uuid.UUID, loyaltyTier, mealPreference, specialNeeds string, loyaltyPoints int) (*Passenger, error)
	DeletePassenger(ctx context.Context, id uuid.UUID) error
}

type passengerService struct {
	repo PassengerRepository
}

// NewPassengerService creates a new PassengerService instance.
func NewPassengerService(repo PassengerRepository) PassengerService {
	return &passengerService{repo: repo}
}

func (s *passengerService) CreatePassenger(ctx context.Context, name, email, phone, passport string) (*Passenger, error) {
	p := &Passenger{
		ID:             uuid.New(),
		Name:           name,
		Email:          email,
		Phone:          phone,
		PassportNumber: passport,
	}
	if err := s.repo.Create(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *passengerService) GetPassenger(ctx context.Context, id uuid.UUID) (*Passenger, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *passengerService) UpdatePassenger(ctx context.Context, id uuid.UUID, name, email, phone, passport string) (*Passenger, error) {
	p := &Passenger{
		ID:             id,
		Name:           name,
		Email:          email,
		Phone:          phone,
		PassportNumber: passport,
	}
	if err := s.repo.Update(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *passengerService) UpdatePreferences(ctx context.Context, id uuid.UUID, loyaltyTier, mealPreference, specialNeeds string, loyaltyPoints int) (*Passenger, error) {
	p, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if loyaltyTier != "" {
		p.LoyaltyTier = loyaltyTier
	}
	if mealPreference != "" {
		p.MealPreference = mealPreference
	}
	p.SpecialNeeds = specialNeeds
	if loyaltyPoints != 0 {
		p.LoyaltyPoints = loyaltyPoints
	}
	if err := s.repo.Update(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *passengerService) DeletePassenger(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}
