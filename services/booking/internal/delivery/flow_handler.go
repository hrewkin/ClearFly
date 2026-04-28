package delivery

import (
	"errors"
	"net/http"

	"github.com/cleanair/booking/internal/usecase"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// FlowHandler exposes the seat-aware booking and check-in endpoints.
type FlowHandler struct {
	flow usecase.BookingFlowService
}

// NewFlowHandler creates a new flow handler.
func NewFlowHandler(flow usecase.BookingFlowService) *FlowHandler {
	return &FlowHandler{flow: flow}
}

// BookSeat handles POST /bookings/book — full booking flow with seat selection.
func (h *FlowHandler) BookSeat(c *gin.Context) {
	var req struct {
		FlightID    uuid.UUID `json:"flight_id"`
		PassengerID uuid.UUID `json:"passenger_id"`
		SeatID      uuid.UUID `json:"seat_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	booking, err := h.flow.BookSeat(c.Request.Context(), req.FlightID, req.PassengerID, req.SeatID)
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrSeatNotAvailable):
			c.JSON(http.StatusConflict, gin.H{"error": "seat is not available"})
		case errors.Is(err, usecase.ErrFlightNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "flight not found"})
		case errors.Is(err, usecase.ErrTariffNotFound):
			c.JSON(http.StatusBadRequest, gin.H{"error": "tariff not configured for this seat class"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusCreated, booking)
}

// ListByPassenger handles GET /passengers/:id/bookings.
func (h *FlowHandler) ListByPassenger(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid passenger id"})
		return
	}
	bookings, err := h.flow.ListBookings(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, bookings)
}

// ListByFlight handles GET /bookings/flight/:id and returns all bookings
// (including cancelled ones) for a given flight. Used by the staff
// dashboard to render the passenger manifest.
func (h *FlowHandler) ListByFlight(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid flight id"})
		return
	}
	bookings, err := h.flow.ListBookingsByFlight(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, bookings)
}

// Cancel handles POST /bookings/:id/cancel. Body may carry an optional
// reason/actor; on success the seat is released and a BOOKING_CANCELLED
// notification is published.
func (h *FlowHandler) Cancel(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid booking id"})
		return
	}
	var req struct {
		Reason string `json:"reason"`
		Actor  string `json:"actor"`
	}
	_ = c.ShouldBindJSON(&req)
	b, err := h.flow.CancelBooking(c.Request.Context(), id, req.Reason, req.Actor)
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrBookingNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "booking not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, b)
}

// CheckIn handles POST /bookings/:id/checkin.
func (h *FlowHandler) CheckIn(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid booking id"})
		return
	}
	b, err := h.flow.CheckIn(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, b)
}
