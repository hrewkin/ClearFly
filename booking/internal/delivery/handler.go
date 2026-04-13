package delivery

import (
    "net/http"
    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
    "github.com/cleanair/booking/internal/usecase"
)

type Handler struct {
    svc usecase.BookingService
}

func NewHandler(svc usecase.BookingService) *Handler {
    return &Handler{svc: svc}
}

// CreateBooking handles POST /bookings
func (h *Handler) CreateBooking(c *gin.Context) {
    var req struct {
        FlightID    uuid.UUID `json:"flight_id"`
        PassengerID uuid.UUID `json:"passenger_id"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    booking, err := h.svc.CreateBooking(c.Request.Context(), req.FlightID, req.PassengerID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusCreated, booking)
}

// GetBooking handles GET /bookings/:id
func (h *Handler) GetBooking(c *gin.Context) {
    idStr := c.Param("id")
    id, err := uuid.Parse(idStr)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
        return
    }
    booking, err := h.svc.GetBooking(c.Request.Context(), id)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, booking)
}
