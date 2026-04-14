package delivery

import (
	"net/http"

	"github.com/cleanair/analytics/internal/repository"
	"github.com/gin-gonic/gin"
)

type HttpHandler struct {
	repo repository.Repository
}

func NewHttpHandler(repo repository.Repository) *HttpHandler {
	return &HttpHandler{repo: repo}
}

func (h *HttpHandler) GetLoadFactor(c *gin.Context) {
	flightID := c.Param("flight_id")
	if flightID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "flight_id is required"})
		return
	}

	res, err := h.repo.GetFlightLoad(c.Request.Context(), flightID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	suggestedPrice := 100.0 // Base price
	if res.LoadFactor > 80 {
		suggestedPrice *= 1.5 // High demand
	} else if res.LoadFactor > 50 {
		suggestedPrice *= 1.2
	}

	c.JSON(http.StatusOK, gin.H{
		"flight_id":       flightID,
		"analytics":       res,
		"suggested_price": suggestedPrice,
	})
}
