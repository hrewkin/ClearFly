package delivery

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/cleanair/passenger/internal/audit"
	"github.com/cleanair/passenger/internal/auth"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// StaffHandler implements operator-level actions (refunds, audit log
// access). All endpoints require an authenticated user with the staff or
// admin role; the audit endpoint additionally requires admin.
type StaffHandler struct {
	auth        *auth.Service
	audit       *audit.Repository
	bookingHost string
	httpClient  *http.Client
}

// NewStaffHandler wires the dependencies. bookingHost is the base URL of
// the booking service (e.g. "http://booking:8080"). When empty it is read
// from the BOOKING_SERVICE_URL environment variable, falling back to the
// docker-compose service name.
func NewStaffHandler(authSvc *auth.Service, auditRepo *audit.Repository, bookingHost string) *StaffHandler {
	if bookingHost == "" {
		bookingHost = os.Getenv("BOOKING_SERVICE_URL")
	}
	if bookingHost == "" {
		bookingHost = "http://booking:8080"
	}
	return &StaffHandler{
		auth:        authSvc,
		audit:       auditRepo,
		bookingHost: bookingHost,
		httpClient:  &http.Client{Timeout: 5 * time.Second},
	}
}

// authorize parses the bearer token from the request and ensures the
// caller has one of the allowed roles. On failure it writes the response
// and returns an empty user.
func (h *StaffHandler) authorize(c *gin.Context, allowed ...string) (*auth.User, bool) {
	token := extractToken(c)
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "no token"})
		return nil, false
	}
	uid, role, err := h.auth.ParseToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return nil, false
	}
	roleAllowed := false
	for _, a := range allowed {
		if role == a {
			roleAllowed = true
			break
		}
	}
	if !roleAllowed {
		c.JSON(http.StatusForbidden, gin.H{"error": "недостаточно прав"})
		return nil, false
	}
	u, err := h.auth.GetByID(c.Request.Context(), uid)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
		return nil, false
	}
	return u, true
}

type refundReq struct {
	BookingID uuid.UUID `json:"booking_id"`
	Reason    string    `json:"reason"`
}

// Refund cancels a booking on behalf of staff/admin and writes an audit
// entry. The actual booking mutation is delegated to the booking service.
func (h *StaffHandler) Refund(c *gin.Context) {
	u, ok := h.authorize(c, auth.RoleStaff, auth.RoleAdmin)
	if !ok {
		return
	}
	var req refundReq
	if err := c.ShouldBindJSON(&req); err != nil || req.BookingID == uuid.Nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}

	body, _ := json.Marshal(map[string]string{
		"reason": req.Reason,
		"actor":  staffActorLabel(u),
	})
	resp, err := h.httpClient.Post(
		h.bookingHost+"/bookings/"+req.BookingID.String()+"/cancel",
		"application/json", bytes.NewReader(body))
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "booking service unavailable"})
		return
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 400 {
		c.Data(resp.StatusCode, "application/json", respBody)
		return
	}

	h.audit.LogJSON(c.Request.Context(), audit.Entry{
		ActorID:    u.ID,
		ActorRole:  u.Role,
		ActorName:  u.FullName,
		Action:     "BOOKING_REFUND",
		TargetType: "booking",
		TargetID:   req.BookingID.String(),
		IPAddress:  c.ClientIP(),
	}, map[string]string{"reason": req.Reason})

	c.Data(resp.StatusCode, "application/json", respBody)
}

// SelfCancel lets a passenger cancel their own booking (the booking is
// cross-checked against the caller's passenger_id). Audit is written with
// actor=passenger.
func (h *StaffHandler) SelfCancel(c *gin.Context) {
	u, ok := h.authorize(c, auth.RolePassenger, auth.RoleAdmin)
	if !ok {
		return
	}
	bookingID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid booking id"})
		return
	}

	booking, err := h.fetchBooking(bookingID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "booking not found"})
		return
	}
	if u.Role == auth.RolePassenger {
		if u.PassengerID == nil || *u.PassengerID != booking.PassengerID {
			c.JSON(http.StatusForbidden, gin.H{"error": "это не ваше бронирование"})
			return
		}
	}

	body, _ := json.Marshal(map[string]string{
		"reason": "self-cancel",
		"actor":  "passenger:" + u.ID.String(),
	})
	resp, err := h.httpClient.Post(
		h.bookingHost+"/bookings/"+bookingID.String()+"/cancel",
		"application/json", bytes.NewReader(body))
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "booking service unavailable"})
		return
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode < 400 {
		h.audit.LogJSON(c.Request.Context(), audit.Entry{
			ActorID:    u.ID,
			ActorRole:  u.Role,
			ActorName:  u.FullName,
			Action:     "BOOKING_SELF_CANCEL",
			TargetType: "booking",
			TargetID:   bookingID.String(),
			IPAddress:  c.ClientIP(),
		}, nil)
	}
	c.Data(resp.StatusCode, "application/json", respBody)
}

type miniBooking struct {
	ID          uuid.UUID `json:"id"`
	PassengerID uuid.UUID `json:"passenger_id"`
	Status      string    `json:"status"`
}

func (h *StaffHandler) fetchBooking(id uuid.UUID) (*miniBooking, error) {
	resp, err := h.httpClient.Get(h.bookingHost + "/bookings/" + id.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("booking not found")
	}
	var b miniBooking
	if err := json.NewDecoder(resp.Body).Decode(&b); err != nil {
		return nil, err
	}
	return &b, nil
}

// Audit returns the latest audit entries (admin only).
func (h *StaffHandler) Audit(c *gin.Context) {
	if _, ok := h.authorize(c, auth.RoleAdmin); !ok {
		return
	}
	limit := 100
	entries, err := h.audit.List(c.Request.Context(), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if entries == nil {
		entries = []audit.Entry{}
	}
	c.JSON(http.StatusOK, entries)
}

func staffActorLabel(u *auth.User) string {
	if u.Role == auth.RoleStaff && u.EmployeeID != "" {
		return "staff:" + u.EmployeeID
	}
	return u.Role + ":" + u.ID.String()
}
