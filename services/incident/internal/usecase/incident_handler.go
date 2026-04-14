package usecase

import (
	"log"
)

type IncidentEvent struct {
	Type     string `json:"type"` // e.g., "FLIGHT_CANCELLED", "FLIGHT_DELAYED"
	FlightID string `json:"flight_id"`
	Reason   string `json:"reason"`
}

type IncidentHandler struct {
	// dependencies like repositories or internal API clients go here
}

func NewIncidentHandler() *IncidentHandler {
	return &IncidentHandler{}
}

func (h *IncidentHandler) HandleIncident(event IncidentEvent) error {
	log.Printf("Processing incident: [%s] for flight %s. Reason: %s", event.Type, event.FlightID, event.Reason)

	switch event.Type {
	case "FLIGHT_CANCELLED":
		return h.handleCancellation(event)
	case "FLIGHT_DELAYED":
		return h.handleDelay(event)
	default:
		log.Printf("Unknown incident type: %s", event.Type)
	}
	return nil
}

func (h *IncidentHandler) handleCancellation(event IncidentEvent) error {
	// Future: Fetch all bookings for this flight and dispatch rebooking commands
	log.Printf("Initiating rebooking protocol for flight %s", event.FlightID)
	return nil
}

func (h *IncidentHandler) handleDelay(event IncidentEvent) error {
	// Future: Send notifications to all passengers
	log.Printf("Initiating notification protocol for delayed flight %s", event.FlightID)
	return nil
}
