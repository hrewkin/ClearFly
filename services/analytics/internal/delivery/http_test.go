package delivery

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cleanair/analytics/internal/repository"
	"github.com/gin-gonic/gin"
)

type mockRepo struct {
	getFlightLoadFunc func(ctx context.Context, flightID string) (*repository.AnalyticsResult, error)
}

func (m *mockRepo) GetFlightLoad(ctx context.Context, flightID string) (*repository.AnalyticsResult, error) {
	if m.getFlightLoadFunc != nil {
		return m.getFlightLoadFunc(ctx, flightID)
	}
	return nil, errors.New("not implemented")
}

func TestGetLoadFactor_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	flightID := "FL123"
	expectedAnalytics := &repository.AnalyticsResult{
		TotalBookings: 135, // 90% of 150
		LoadFactor:    90,
	}

	repo := &mockRepo{
		getFlightLoadFunc: func(ctx context.Context, id string) (*repository.AnalyticsResult, error) {
			if id == flightID {
				return expectedAnalytics, nil
			}
			return nil, errors.New("not found")
		},
	}

	handler := NewHttpHandler(repo)
	router := gin.Default()
	router.GET("/api/v1/analytics/:flight_id", handler.GetLoadFactor)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/analytics/"+flightID, nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.Code)
	}

	var result map[string]interface{}
	json.Unmarshal(resp.Body.Bytes(), &result)

	if result["flight_id"] != flightID {
		t.Errorf("expected flight_id %v, got %v", flightID, result["flight_id"])
	}

	// 90 load factor should trigger 1.5 multiplier (100 * 1.5 = 150)
	price, ok := result["suggested_price"].(float64)
	if !ok || price != 150.0 {
		t.Errorf("expected suggested_price 150.0, got %v", result["suggested_price"])
	}
}

func TestGetLoadFactor_MediumLoad(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := &mockRepo{
		getFlightLoadFunc: func(ctx context.Context, id string) (*repository.AnalyticsResult, error) {
			return &repository.AnalyticsResult{
				TotalBookings: 90, // 60% of 150
				LoadFactor:    60,
			}, nil
		},
	}

	handler := NewHttpHandler(repo)
	router := gin.Default()
	router.GET("/api/v1/analytics/:flight_id", handler.GetLoadFactor)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/analytics/FL456", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	var result map[string]interface{}
	json.Unmarshal(resp.Body.Bytes(), &result)

	// 60 load factor should trigger 1.2 multiplier (100 * 1.2 = 120)
	if price, _ := result["suggested_price"].(float64); price != 120.0 {
		t.Errorf("expected suggested_price 120.0, got %v", result["suggested_price"])
	}
}
