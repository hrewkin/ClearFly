package delivery

import (
	"net/http"

	"github.com/cleanair/passenger/internal/usecase"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Handler is the HTTP handler for the passenger service.
type Handler struct {
	svc usecase.PassengerService
}

// NewHandler creates a new passenger HTTP handler.
func NewHandler(svc usecase.PassengerService) *Handler {
	return &Handler{svc: svc}
}

// CreatePassenger handles POST /passengers
func (h *Handler) CreatePassenger(c *gin.Context) {
	var req struct {
		Name           string `json:"name"`
		Email          string `json:"email"`
		Phone          string `json:"phone"`
		PassportNumber string `json:"passport_number"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	p, err := h.svc.CreatePassenger(c.Request.Context(), req.Name, req.Email, req.Phone, req.PassportNumber)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, p)
}

// GetPassenger handles GET /passengers/:id
func (h *Handler) GetPassenger(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	p, err := h.svc.GetPassenger(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, p)
}

// UpdatePassenger handles PUT /passengers/:id
func (h *Handler) UpdatePassenger(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req struct {
		Name           string `json:"name"`
		Email          string `json:"email"`
		Phone          string `json:"phone"`
		PassportNumber string `json:"passport_number"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	p, err := h.svc.UpdatePassenger(c.Request.Context(), id, req.Name, req.Email, req.Phone, req.PassportNumber)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, p)
}

// UpdatePreferences handles PATCH /passengers/:id/preferences
func (h *Handler) UpdatePreferences(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var req struct {
		LoyaltyTier    string `json:"loyalty_tier"`
		MealPreference string `json:"meal_preference"`
		SpecialNeeds   string `json:"special_needs"`
		LoyaltyPoints  int    `json:"loyalty_points"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	p, err := h.svc.UpdatePreferences(c.Request.Context(), id, req.LoyaltyTier, req.MealPreference, req.SpecialNeeds, req.LoyaltyPoints)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, p)
}

// DeletePassenger handles DELETE /passengers/:id
func (h *Handler) DeletePassenger(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.svc.DeletePassenger(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "passenger deleted"})
}
