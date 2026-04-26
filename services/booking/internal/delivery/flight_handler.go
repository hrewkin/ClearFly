package delivery

import (
	"net/http"

	"github.com/cleanair/booking/internal/usecase"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// FlightHandler handles flight-related HTTP requests.
type FlightHandler struct {
	svc usecase.FlightService
}

// NewFlightHandler creates a new flight handler.
func NewFlightHandler(svc usecase.FlightService) *FlightHandler {
	return &FlightHandler{svc: svc}
}

// CreateFlight handles POST /flights
func (h *FlightHandler) CreateFlight(c *gin.Context) {
	var req usecase.Flight
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	f, err := h.svc.CreateFlight(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, f)
}

// GetFlight handles GET /flights/:id
func (h *FlightHandler) GetFlight(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid flight id"})
		return
	}
	f, err := h.svc.GetFlight(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "flight not found"})
		return
	}
	c.JSON(http.StatusOK, f)
}

// SearchFlights handles GET /flights/search?origin=X&destination=Y&date=2026-05-01
func (h *FlightHandler) SearchFlights(c *gin.Context) {
	origin := c.Query("origin")
	destination := c.Query("destination")
	date := c.Query("date")
	if origin == "" || destination == "" || date == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "origin, destination and date are required"})
		return
	}
	flights, err := h.svc.SearchFlights(c.Request.Context(), origin, destination, date)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, flights)
}

// UpdateFlightStatus handles PATCH /flights/:id/status
func (h *FlightHandler) UpdateFlightStatus(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid flight id"})
		return
	}
	var req struct {
		Status string `json:"status"`
		Gate   string `json:"gate"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.UpdateFlightStatus(c.Request.Context(), id, req.Status, req.Gate); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "flight status updated"})
}

// GetSeats handles GET /flights/:id/seats — returns all seats with status
// so clients can render a visual seat map (available + booked).
func (h *FlightHandler) GetSeats(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid flight id"})
		return
	}
	seats, err := h.svc.GetSeatsByFlight(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, seats)
}

// ListTariffs handles GET /flights/:id/tariffs.
func (h *FlightHandler) ListTariffs(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid flight id"})
		return
	}
	ts, err := h.svc.ListTariffs(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, ts)
}

// UpcomingFlights handles GET /flights/upcoming?limit=N — list of next
// scheduled flights, used as default content on the search page.
func (h *FlightHandler) UpcomingFlights(c *gin.Context) {
	limit := 10
	flights, err := h.svc.UpcomingFlights(c.Request.Context(), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, flights)
}

// CreateTariff handles POST /flights/:id/tariff
func (h *FlightHandler) CreateTariff(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid flight id"})
		return
	}
	var req struct {
		Class    string  `json:"class"`
		Price    float64 `json:"base_price"`
		Currency string  `json:"currency"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	t, err := h.svc.CreateTariff(c.Request.Context(), id, req.Class, req.Price, req.Currency)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, t)
}
