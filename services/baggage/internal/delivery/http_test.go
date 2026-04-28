package delivery

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cleanair/baggage/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type mockRepo struct {
	getByIDFunc          func(ctx context.Context, id uuid.UUID) (*repository.BaggageStatus, error)
	upsertFunc           func(ctx context.Context, b *repository.BaggageStatus) error
	listFunc             func(ctx context.Context, limit int) ([]repository.BaggageStatus, error)
	listByPassengerFunc  func(ctx context.Context, passengerID uuid.UUID) ([]repository.BaggageStatus, error)
}

func (m *mockRepo) GetByID(ctx context.Context, id uuid.UUID) (*repository.BaggageStatus, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	return nil, errors.New("not implemented")
}

func (m *mockRepo) Upsert(ctx context.Context, b *repository.BaggageStatus) error {
	if m.upsertFunc != nil {
		return m.upsertFunc(ctx, b)
	}
	return nil
}

func (m *mockRepo) List(ctx context.Context, limit int) ([]repository.BaggageStatus, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx, limit)
	}
	return nil, nil
}

func (m *mockRepo) ListByPassenger(ctx context.Context, passengerID uuid.UUID) ([]repository.BaggageStatus, error) {
	if m.listByPassengerFunc != nil {
		return m.listByPassengerFunc(ctx, passengerID)
	}
	return nil, nil
}

func (m *mockRepo) ListByFlight(ctx context.Context, flightID uuid.UUID) ([]repository.BaggageStatus, error) {
	return nil, nil
}

func TestGetBaggage_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	bagID := uuid.New()
	expectedBag := &repository.BaggageStatus{ID: bagID, Location: "GATE_A1", Status: "SCANNED"}

	repo := &mockRepo{
		getByIDFunc: func(ctx context.Context, id uuid.UUID) (*repository.BaggageStatus, error) {
			if id == bagID {
				return expectedBag, nil
			}
			return nil, errors.New("not found")
		},
	}

	handler := NewHttpHandler(repo)
	router := gin.Default()
	router.GET("/api/v1/baggage/:id", handler.GetBaggage)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/baggage/"+bagID.String(), nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.Code)
	}

	var bag repository.BaggageStatus
	json.Unmarshal(resp.Body.Bytes(), &bag)
	if bag.ID != bagID {
		t.Errorf("expected bag id %v, got %v", bagID, bag.ID)
	}
}

func TestGetBaggage_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewHttpHandler(&mockRepo{})
	router := gin.Default()
	router.GET("/api/v1/baggage/:id", handler.GetBaggage)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/baggage/invalid-uuid", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", resp.Code)
	}
}
