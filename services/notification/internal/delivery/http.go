package delivery

import (
	"net/http"
	"strconv"

	"github.com/cleanair/notification/internal/usecase"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// HttpHandler exposes the notification feed.
type HttpHandler struct {
	store *usecase.Store
}

// NewHttpHandler creates a new notification HTTP handler.
func NewHttpHandler(store *usecase.Store) *HttpHandler {
	return &HttpHandler{store: store}
}

// ListByPassenger handles GET /notifications/:passenger_id.
func (h *HttpHandler) ListByPassenger(c *gin.Context) {
	pid := c.Param("passenger_id")
	c.JSON(http.StatusOK, h.store.ListByPassenger(pid))
}

// ListAll handles GET /notifications — broadcast feed used by the demo UI.
func (h *HttpHandler) ListAll(c *gin.Context) {
	limit := 50
	if l := c.Query("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 {
			limit = v
		}
	}
	c.JSON(http.StatusOK, h.store.ListAll(limit))
}

// MarkRead handles POST /notifications/:id/read.
func (h *HttpHandler) MarkRead(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if !h.store.MarkRead(c.Query("passenger_id"), id) {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
