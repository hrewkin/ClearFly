package usecase

import (
	"context"
	"fmt"
	"log"
)

// IncidentEvent describes a non-routine event that affects a flight.
//
// Type values:
//   - FLIGHT_DELAYED   — flight is delayed by Reason minutes / cause.
//   - FLIGHT_CANCELLED — flight is cancelled, rebooking required.
//   - GATE_CHANGED     — gate has changed; passengers must be informed.
type IncidentEvent struct {
	Type     string `json:"type"`
	FlightID string `json:"flight_id"`
	Reason   string `json:"reason"`
	NewGate  string `json:"new_gate,omitempty"`
}

// NotificationPublisher decouples the handler from a specific transport.
// In production this is a RabbitMQ publisher; in tests it can be a fake.
type NotificationPublisher interface {
	Publish(ctx context.Context, payload Notification) error
}

// Notification is the payload published to the notification_events queue.
//
// PassengerID is left empty for broadcast events — the notification service
// will fan it out to all subscribers of the flight in production. For the
// demo we record it on the broadcast feed.
type Notification struct {
	Type        string                 `json:"type"`
	PassengerID string                 `json:"passenger_id,omitempty"`
	FlightID    string                 `json:"flight_id,omitempty"`
	Title       string                 `json:"title"`
	Content     string                 `json:"content"`
	Channels    []string               `json:"channels"`
	Meta        map[string]interface{} `json:"meta,omitempty"`
}

// IncidentHandler processes incoming incident events and dispatches
// passenger notifications.
type IncidentHandler struct {
	publisher NotificationPublisher
}

// NewIncidentHandler creates a handler. publisher may be nil — events will
// be logged but not published.
func NewIncidentHandler(publisher NotificationPublisher) *IncidentHandler {
	return &IncidentHandler{publisher: publisher}
}

// HandleIncident routes the event to the appropriate handler based on Type.
func (h *IncidentHandler) HandleIncident(event IncidentEvent) error {
	log.Printf("Processing incident: [%s] for flight %s. Reason: %s", event.Type, event.FlightID, event.Reason)

	switch event.Type {
	case "FLIGHT_CANCELLED":
		return h.handleCancellation(event)
	case "FLIGHT_DELAYED":
		return h.handleDelay(event)
	case "GATE_CHANGED":
		return h.handleGateChange(event)
	default:
		log.Printf("Unknown incident type: %s", event.Type)
	}
	return nil
}

func (h *IncidentHandler) handleCancellation(event IncidentEvent) error {
	log.Printf("Initiating rebooking protocol for flight %s", event.FlightID)
	return h.publish(Notification{
		Type:     "FLIGHT_CANCELLED",
		FlightID: event.FlightID,
		Title:    "Рейс отменён",
		Content:  fmt.Sprintf("К сожалению, рейс %s отменён. Причина: %s. Мы предложим перебронирование на ближайший рейс.", event.FlightID, event.Reason),
		Channels: []string{"PUSH", "SMS", "EMAIL"},
		Meta:     map[string]interface{}{"reason": event.Reason},
	})
}

func (h *IncidentHandler) handleDelay(event IncidentEvent) error {
	log.Printf("Initiating notification protocol for delayed flight %s", event.FlightID)
	return h.publish(Notification{
		Type:     "FLIGHT_DELAYED",
		FlightID: event.FlightID,
		Title:    "Рейс задержан",
		Content:  fmt.Sprintf("Рейс %s задержан. Причина: %s. Следите за обновлениями расписания.", event.FlightID, event.Reason),
		Channels: []string{"PUSH", "SMS"},
		Meta:     map[string]interface{}{"reason": event.Reason},
	})
}

func (h *IncidentHandler) handleGateChange(event IncidentEvent) error {
	gate := event.NewGate
	if gate == "" {
		gate = event.Reason // accept reason as the new gate label as well
	}
	log.Printf("Notifying passengers about gate change for flight %s -> %s", event.FlightID, gate)
	return h.publish(Notification{
		Type:     "GATE_CHANGED",
		FlightID: event.FlightID,
		Title:    "Изменён выход на посадку",
		Content:  fmt.Sprintf("Выход на посадку для рейса %s изменён на %s.", event.FlightID, gate),
		Channels: []string{"PUSH"},
		Meta:     map[string]interface{}{"new_gate": gate},
	})
}

func (h *IncidentHandler) publish(n Notification) error {
	if h.publisher == nil {
		return nil
	}
	return h.publisher.Publish(context.Background(), n)
}
