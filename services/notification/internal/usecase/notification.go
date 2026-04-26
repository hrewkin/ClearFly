package usecase

import (
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Notification is a single record delivered to a passenger across one or
// more channels (push, sms, email).
type Notification struct {
	ID          uuid.UUID `json:"id"`
	PassengerID string    `json:"passenger_id"`
	FlightID    string    `json:"flight_id,omitempty"`
	Type        string    `json:"type"`
	Title       string    `json:"title"`
	Content     string    `json:"content"`
	Channel     string    `json:"channel"`
	Read        bool      `json:"read"`
	SentAt      time.Time `json:"sent_at"`
}

// Store keeps notifications in-memory. It's intentionally simple: in a
// production system this would be a real database, but for the demo a
// thread-safe in-memory store is enough to drive the WebUI and survives
// the lifetime of the running container.
type Store struct {
	mu sync.RWMutex
	// Indexed by passenger_id for fast retrieval; "*" key holds the
	// broadcast feed used by the demo WebUI when no passenger is given.
	byPassenger map[string][]Notification
}

// NewStore creates an empty in-memory notification store.
func NewStore() *Store {
	return &Store{byPassenger: make(map[string][]Notification)}
}

// Add stores a notification and (always) appends it to the broadcast feed.
func (s *Store) Add(n Notification) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if n.ID == uuid.Nil {
		n.ID = uuid.New()
	}
	if n.SentAt.IsZero() {
		n.SentAt = time.Now().UTC()
	}
	if n.PassengerID != "" {
		s.byPassenger[n.PassengerID] = append([]Notification{n}, s.byPassenger[n.PassengerID]...)
	}
	s.byPassenger["*"] = append([]Notification{n}, s.byPassenger["*"]...)
	if len(s.byPassenger["*"]) > 200 {
		s.byPassenger["*"] = s.byPassenger["*"][:200]
	}
}

// ListByPassenger returns notifications for the given passenger sorted by
// most recent first.
func (s *Store) ListByPassenger(passengerID string) []Notification {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.byPassenger[passengerID]
	out := make([]Notification, len(src))
	copy(out, src)
	sort.Slice(out, func(i, j int) bool { return out[i].SentAt.After(out[j].SentAt) })
	return out
}

// ListAll returns the broadcast feed.
func (s *Store) ListAll(limit int) []Notification {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.byPassenger["*"]
	if limit <= 0 || limit > len(src) {
		limit = len(src)
	}
	out := make([]Notification, limit)
	copy(out, src[:limit])
	return out
}

// MarkRead marks a notification as read for a given passenger feed.
func (s *Store) MarkRead(passengerID string, id uuid.UUID) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	updated := false
	for key := range s.byPassenger {
		if passengerID != "" && passengerID != "*" && key != passengerID && key != "*" {
			continue
		}
		for i := range s.byPassenger[key] {
			if s.byPassenger[key][i].ID == id {
				s.byPassenger[key][i].Read = true
				updated = true
			}
		}
	}
	return updated
}
