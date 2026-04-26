package usecase

import (
	"context"

	"github.com/cleanair/analytics/internal/repository"
)

// AnalyticsService defines the business logic interface for analytics operations.
type AnalyticsService interface {
	GetFlightLoadFactor(ctx context.Context, flightID string) (*LoadFactorResult, error)
}

// LoadFactorResult contains the analytics data and pricing suggestion.
type LoadFactorResult struct {
	FlightID       string  `json:"flight_id"`
	TotalBookings  int     `json:"total_bookings"`
	LoadFactor     int     `json:"load_factor"`
	SuggestedPrice float64 `json:"suggested_price"`
}

type analyticsService struct {
	repo repository.Repository
}

// NewAnalyticsService creates a new AnalyticsService instance.
func NewAnalyticsService(repo repository.Repository) AnalyticsService {
	return &analyticsService{repo: repo}
}

func (s *analyticsService) GetFlightLoadFactor(ctx context.Context, flightID string) (*LoadFactorResult, error) {
	res, err := s.repo.GetFlightLoad(ctx, flightID)
	if err != nil {
		return nil, err
	}

	// Base price calculation logic
	suggestedPrice := 100.0
	if res.LoadFactor > 80 {
		suggestedPrice *= 1.5 // High demand — +50%
	} else if res.LoadFactor > 50 {
		suggestedPrice *= 1.2 // Medium demand — +20%
	}

	return &LoadFactorResult{
		FlightID:       flightID,
		TotalBookings:  res.TotalBookings,
		LoadFactor:     res.LoadFactor,
		SuggestedPrice: suggestedPrice,
	}, nil
}
