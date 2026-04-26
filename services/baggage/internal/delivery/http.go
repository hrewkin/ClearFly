package delivery

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/cleanair/baggage/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// BaggageStages is the canonical ordered list of stages the baggage moves
// through. It is used by the UI and by the "next scan" simulator to advance
// a tag to the next logical stage.
var BaggageStages = []string{
	"CHECKED_IN",
	"SCREENED",
	"LOADED",
	"IN_FLIGHT",
	"UNLOADED",
	"CLAIMED",
}

// stageLocation is a demo-friendly default location for each stage; the UI
// overrides it when the user wants to provide something custom.
var stageLocation = map[string]string{
	"CHECKED_IN": "Стойка регистрации",
	"SCREENED":   "Интроскоп",
	"LOADED":     "Багажный люк",
	"IN_FLIGHT":  "На борту",
	"UNLOADED":   "Багажная лента",
	"CLAIMED":    "Выдан пассажиру",
}

type HttpHandler struct {
	repo repository.Repository
}

func NewHttpHandler(repo repository.Repository) *HttpHandler {
	return &HttpHandler{repo: repo}
}

type createBaggageRequest struct {
	PassengerID string `json:"passenger_id" binding:"required"`
	FlightID    string `json:"flight_id"`
	Status      string `json:"status"`
	Location    string `json:"location"`
}

type scanRequest struct {
	Status   string `json:"status"`
	Location string `json:"location"`
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

func (h *HttpHandler) ListBaggage(c *gin.Context) {
	if pid := c.Query("passenger_id"); pid != "" {
		pu, err := uuid.Parse(pid)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid passenger_id"})
			return
		}
		bags, err := h.repo.ListByPassenger(c.Request.Context(), pu)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, bags)
		return
	}

	limit := 50
	if l := c.Query("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil {
			limit = v
		}
	}
	bags, err := h.repo.List(c.Request.Context(), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, bags)
}

func (h *HttpHandler) CreateBaggage(c *gin.Context) {
	var req createBaggageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	pid, err := uuid.Parse(req.PassengerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid passenger_id"})
		return
	}

	var fid *uuid.UUID
	if strings.TrimSpace(req.FlightID) != "" {
		parsed, err := uuid.Parse(req.FlightID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid flight_id"})
			return
		}
		fid = &parsed
	}

	status := strings.ToUpper(strings.TrimSpace(req.Status))
	if status == "" {
		status = "CHECKED_IN"
	}
	location := strings.TrimSpace(req.Location)
	if location == "" {
		location = defaultLocation(status)
	}

	bag := &repository.BaggageStatus{
		ID:          uuid.New(),
		PassengerID: pid,
		FlightID:    fid,
		Status:      status,
		Location:    location,
		UpdatedAt:   time.Now().UTC(),
	}
	if err := h.repo.Upsert(c.Request.Context(), bag); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, bag)
}

func (h *HttpHandler) Scan(c *gin.Context) {
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

	var req scanRequest
	_ = c.ShouldBindJSON(&req)

	status := strings.ToUpper(strings.TrimSpace(req.Status))
	if status == "" {
		status = nextStage(bag.Status)
	}
	location := strings.TrimSpace(req.Location)
	if location == "" {
		location = defaultLocation(status)
	}

	bag.Status = status
	bag.Location = location
	bag.UpdatedAt = time.Now().UTC()

	if err := h.repo.Upsert(c.Request.Context(), bag); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, bag)
}

func defaultLocation(stage string) string {
	if v, ok := stageLocation[stage]; ok {
		return v
	}
	return "—"
}

func nextStage(current string) string {
	for i, s := range BaggageStages {
		if s == current && i+1 < len(BaggageStages) {
			return BaggageStages[i+1]
		}
	}
	return BaggageStages[len(BaggageStages)-1]
}
