package delivery

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cleanair/passenger/internal/usecase"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type mockService struct {
	createFunc func(ctx context.Context, name, email, phone, passport string) (*usecase.Passenger, error)
	getFunc    func(ctx context.Context, id uuid.UUID) (*usecase.Passenger, error)
	updateFunc func(ctx context.Context, id uuid.UUID, name, email, phone, passport string) (*usecase.Passenger, error)
	deleteFunc func(ctx context.Context, id uuid.UUID) error
}

func (m *mockService) CreatePassenger(ctx context.Context, name, email, phone, passport string) (*usecase.Passenger, error) {
	if m.createFunc != nil {
		return m.createFunc(ctx, name, email, phone, passport)
	}
	return nil, errors.New("not implemented")
}

func (m *mockService) GetPassenger(ctx context.Context, id uuid.UUID) (*usecase.Passenger, error) {
	if m.getFunc != nil {
		return m.getFunc(ctx, id)
	}
	return nil, errors.New("not implemented")
}

func (m *mockService) UpdatePassenger(ctx context.Context, id uuid.UUID, name, email, phone, passport string) (*usecase.Passenger, error) {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, id, name, email, phone, passport)
	}
	return nil, errors.New("not implemented")
}

func (m *mockService) UpdatePreferences(ctx context.Context, id uuid.UUID, loyaltyTier, meal, special string, points int) (*usecase.Passenger, error) {
	return nil, errors.New("not implemented")
}

func (m *mockService) DeletePassenger(ctx context.Context, id uuid.UUID) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, id)
	}
	return errors.New("not implemented")
}

func TestCreatePassenger_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	expectedID := uuid.New()
	svc := &mockService{
		createFunc: func(ctx context.Context, name, email, phone, passport string) (*usecase.Passenger, error) {
			return &usecase.Passenger{
				ID:             expectedID,
				Name:           name,
				Email:          email,
				Phone:          phone,
				PassportNumber: passport,
			}, nil
		},
	}

	handler := NewHandler(svc)
	router := gin.Default()
	router.POST("/passengers", handler.CreatePassenger)

	body, _ := json.Marshal(map[string]string{
		"name":            "John Doe",
		"email":           "john@example.com",
		"phone":           "+79001234567",
		"passport_number": "AB1234567",
	})
	req, _ := http.NewRequest(http.MethodPost, "/passengers", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", resp.Code)
	}

	var p usecase.Passenger
	json.Unmarshal(resp.Body.Bytes(), &p)
	if p.Name != "John Doe" {
		t.Errorf("expected name John Doe, got %s", p.Name)
	}
}

func TestGetPassenger_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	pID := uuid.New()
	svc := &mockService{
		getFunc: func(ctx context.Context, id uuid.UUID) (*usecase.Passenger, error) {
			if id == pID {
				return &usecase.Passenger{ID: pID, Name: "Jane Doe", Email: "jane@example.com"}, nil
			}
			return nil, errors.New("not found")
		},
	}

	handler := NewHandler(svc)
	router := gin.Default()
	router.GET("/passengers/:id", handler.GetPassenger)

	req, _ := http.NewRequest(http.MethodGet, "/passengers/"+pID.String(), nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.Code)
	}

	var p usecase.Passenger
	json.Unmarshal(resp.Body.Bytes(), &p)
	if p.ID != pID {
		t.Errorf("expected id %v, got %v", pID, p.ID)
	}
}

func TestGetPassenger_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewHandler(&mockService{})
	router := gin.Default()
	router.GET("/passengers/:id", handler.GetPassenger)

	req, _ := http.NewRequest(http.MethodGet, "/passengers/invalid-uuid", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", resp.Code)
	}
}

func TestDeletePassenger_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	pID := uuid.New()
	svc := &mockService{
		deleteFunc: func(ctx context.Context, id uuid.UUID) error {
			if id == pID {
				return nil
			}
			return errors.New("not found")
		},
	}

	handler := NewHandler(svc)
	router := gin.Default()
	router.DELETE("/passengers/:id", handler.DeletePassenger)

	req, _ := http.NewRequest(http.MethodDelete, "/passengers/"+pID.String(), nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.Code)
	}
}
