package delivery

import (
	"net/http"

	"github.com/cleanair/incident/internal/usecase"
	"github.com/gin-gonic/gin"
)

// HttpHandler handles HTTP requests for the incident service.
type HttpHandler struct {
	handler *usecase.IncidentHandler
}

// NewHttpHandler creates a new HTTP handler for incidents.
func NewHttpHandler(handler *usecase.IncidentHandler) *HttpHandler {
	return &HttpHandler{handler: handler}
}

// CreateIncident handles POST /incidents — allows creating incidents via HTTP (manual trigger).
func (h *HttpHandler) CreateIncident(c *gin.Context) {
	var req usecase.IncidentEvent
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.handler.HandleIncident(req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"message":   "incident processed",
		"type":      req.Type,
		"flight_id": req.FlightID,
	})
}
