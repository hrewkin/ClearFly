package usecase

import "errors"

// Sentinel errors for domain logic.
var (
	ErrSeatNotAvailable = errors.New("seat is not available")
	ErrFlightNotFound   = errors.New("flight not found")
	ErrTariffNotFound   = errors.New("tariff not found")
)
