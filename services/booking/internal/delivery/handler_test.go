package delivery

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cleanair/booking/internal/usecase"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type mockService struct {
	createFunc func(ctx context.Context, flightID, passengerID uuid.UUID) (*usecase.Booking, error)
	getFunc    func(ctx context.Context, id uuid.UUID) (*usecase.Booking, error)
}

func (m *mockService) CreateBooking(ctx context.Context, flightID, passengerID uuid.UUID) (*usecase.Booking, error) {
	if m.createFunc != nil {
		return m.createFunc(ctx, flightID, passengerID)
	}
	return nil, errors.New("not implemented")
}

func (m *mockService) GetBooking(ctx context.Context, id uuid.UUID) (*usecase.Booking, error) {
	if m.getFunc != nil {
		return m.getFunc(ctx, id)
	}
	return nil, errors.New("not implemented")
}

func TestCreateBooking_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	expectedID := uuid.New()
	flightID := uuid.New()
	passengerID := uuid.New()

	svc := &mockService{
		createFunc: func(ctx context.Context, fid, pid uuid.UUID) (*usecase.Booking, error) {
			return &usecase.Booking{
				ID:          expectedID,
				FlightID:    fid,
				PassengerID: pid,
				Status:      "CONFIRMED",
			}, nil
		},
	}

	handler := NewHandler(svc)
	router := gin.Default()
	router.POST("/bookings", handler.CreateBooking)

	body, _ := json.Marshal(map[string]string{
		"flight_id":    flightID.String(),
		"passenger_id": passengerID.String(),
	})
	req, _ := http.NewRequest(http.MethodPost, "/bookings", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", resp.Code)
	}

	var b usecase.Booking
	json.Unmarshal(resp.Body.Bytes(), &b)
	if b.Status != "CONFIRMED" {
		t.Errorf("expected status CONFIRMED, got %s", b.Status)
	}
}

func TestCreateBooking_BadRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewHandler(&mockService{})
	router := gin.Default()
	router.POST("/bookings", handler.CreateBooking)

	req, _ := http.NewRequest(http.MethodPost, "/bookings", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", resp.Code)
	}
}

func TestGetBooking_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	bookingID := uuid.New()
	svc := &mockService{
		getFunc: func(ctx context.Context, id uuid.UUID) (*usecase.Booking, error) {
			if id == bookingID {
				return &usecase.Booking{ID: bookingID, Status: "CONFIRMED"}, nil
			}
			return nil, errors.New("not found")
		},
	}

	handler := NewHandler(svc)
	router := gin.Default()
	router.GET("/bookings/:id", handler.GetBooking)

	req, _ := http.NewRequest(http.MethodGet, "/bookings/"+bookingID.String(), nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.Code)
	}
}

func TestGetBooking_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewHandler(&mockService{})
	router := gin.Default()
	router.GET("/bookings/:id", handler.GetBooking)

	req, _ := http.NewRequest(http.MethodGet, "/bookings/not-a-uuid", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", resp.Code)
	}
}

func TestGetBooking_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := &mockService{
		getFunc: func(ctx context.Context, id uuid.UUID) (*usecase.Booking, error) {
			return nil, errors.New("not found")
		},
	}

	handler := NewHandler(svc)
	router := gin.Default()
	router.GET("/bookings/:id", handler.GetBooking)

	req, _ := http.NewRequest(http.MethodGet, "/bookings/"+uuid.New().String(), nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", resp.Code)
	}
}
