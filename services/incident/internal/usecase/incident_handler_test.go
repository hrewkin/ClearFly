package usecase

import (
	"log"
	"strings"
	"testing"
)

// A simple mock writer to capture logs
type logWriter struct {
	Captured []string
}

func (w *logWriter) Write(p []byte) (n int, err error) {
	w.Captured = append(w.Captured, string(p))
	return len(p), nil
}

func TestHandleIncident_Cancelled(t *testing.T) {
	w := &logWriter{}
	log.SetOutput(w)

	handler := NewIncidentHandler()
	event := IncidentEvent{
		Type:     "FLIGHT_CANCELLED",
		FlightID: "FL999",
		Reason:   "Bad weather",
	}

	err := handler.HandleIncident(event)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	found := false
	for _, l := range w.Captured {
		if strings.Contains(l, "Initiating rebooking protocol") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected rebooking protocol log, captured: %v", w.Captured)
	}
}

func TestHandleIncident_Delayed(t *testing.T) {
	w := &logWriter{}
	log.SetOutput(w)

	handler := NewIncidentHandler()
	event := IncidentEvent{
		Type:     "FLIGHT_DELAYED",
		FlightID: "FL888",
		Reason:   "Maintenance",
	}

	err := handler.HandleIncident(event)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	found := false
	for _, l := range w.Captured {
		if strings.Contains(l, "Initiating notification protocol") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected notification protocol log, captured: %v", w.Captured)
	}
}

func TestHandleIncident_Unknown(t *testing.T) {
	w := &logWriter{}
	log.SetOutput(w)

	handler := NewIncidentHandler()
	event := IncidentEvent{
		Type:     "UNKNOWN_EVENT",
		FlightID: "FL777",
	}

	err := handler.HandleIncident(event)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	found := false
	for _, l := range w.Captured {
		if strings.Contains(l, "Unknown incident type: UNKNOWN_EVENT") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected unknown type log, captured: %v", w.Captured)
	}
}
