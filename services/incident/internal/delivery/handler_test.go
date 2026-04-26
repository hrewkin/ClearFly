package delivery

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cleanair/incident/internal/usecase"
	"github.com/gin-gonic/gin"
)

func TestCreateIncident_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := usecase.NewIncidentHandler()
	httpHandler := NewHttpHandler(handler)

	router := gin.Default()
	router.POST("/incidents", httpHandler.CreateIncident)

	body, _ := json.Marshal(map[string]string{
		"type":      "FLIGHT_CANCELLED",
		"flight_id": "FL999",
		"reason":    "Bad weather",
	})

	req, _ := http.NewRequest(http.MethodPost, "/incidents", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusAccepted {
		t.Fatalf("expected status 202, got %d", resp.Code)
	}

	var result map[string]interface{}
	json.Unmarshal(resp.Body.Bytes(), &result)

	if result["type"] != "FLIGHT_CANCELLED" {
		t.Errorf("expected type FLIGHT_CANCELLED, got %v", result["type"])
	}
}

func TestCreateIncident_BadRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := usecase.NewIncidentHandler()
	httpHandler := NewHttpHandler(handler)

	router := gin.Default()
	router.POST("/incidents", httpHandler.CreateIncident)

	req, _ := http.NewRequest(http.MethodPost, "/incidents", bytes.NewBuffer([]byte("invalid")))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", resp.Code)
	}
}
