package usecase

import (
	"context"

	"github.com/cleanair/baggage/internal/repository"
	"github.com/google/uuid"
)

// BaggageService defines the business logic interface for baggage operations.
type BaggageService interface {
	GetBaggageStatus(ctx context.Context, id uuid.UUID) (*repository.BaggageStatus, error)
	UpdateBaggageStatus(ctx context.Context, bag *repository.BaggageStatus) error
}

type baggageService struct {
	repo repository.Repository
}

// NewBaggageService creates a new BaggageService instance.
func NewBaggageService(repo repository.Repository) BaggageService {
	return &baggageService{repo: repo}
}

func (s *baggageService) GetBaggageStatus(ctx context.Context, id uuid.UUID) (*repository.BaggageStatus, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *baggageService) UpdateBaggageStatus(ctx context.Context, bag *repository.BaggageStatus) error {
	return s.repo.Upsert(ctx, bag)
}
