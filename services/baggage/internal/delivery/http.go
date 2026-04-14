package delivery

import (
	"net/http"

	"github.com/cleanair/baggage/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type HttpHandler struct {
	repo repository.Repository
}

func NewHttpHandler(repo repository.Repository) *HttpHandler {
	return &HttpHandler{repo: repo}
}

func (h *HttpHandler) GetBaggage(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id format"})
		return
	}

	bag, err := h.repo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "baggage not found"})
		return
	}

	c.JSON(http.StatusOK, bag)
}
