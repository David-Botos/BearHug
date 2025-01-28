package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
)

// WebhookPayload represents the inner payload structure of the event.
type WebhookPayload struct {
	StartTs   float64 `json:"start_ts"`
	EndTs     float64 `json:"end_ts"`
	MeetingID string  `json:"meeting_id"`
	Room      string  `json:"room"`
}

// WebhookEvent represents the overall structure of the incoming webhook event.
type WebhookEvent struct {
	Version string         `json:"version"`
	Type    string         `json:"type"`
	ID      string         `json:"id"`
	Payload WebhookPayload `json:"payload"`
	EventTs float64        `json:"event_ts"`
}

// WebhookHandler processes the /webhook route.
func WebhookHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	var event WebhookEvent
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	if err := json.Unmarshal(body, &event); err != nil {
		http.Error(w, "Invalid JSON data", http.StatusBadRequest)
		return
	}

	// Log the received event
	log.Printf("Received event ID: %s, Type: %s, Version: %s", event.ID, event.Type, event.Version)
	log.Printf("Event payload: %+v", event.Payload)

	// Perform specific logic based on event type
	switch event.Type {
	case "meeting.started":
		log.Printf("Meeting started. Meeting ID: %s, Room: %s", event.Payload.MeetingID, event.Payload.Room)
	case "meeting.ended":
		log.Printf("Meeting ended. Meeting ID: %s, Room: %s, Duration: %.2f seconds",
			event.Payload.MeetingID, event.Payload.Room, event.Payload.EndTs-event.Payload.StartTs)

	case "waiting-participant.joined":
		log.Printf("Participant joined waiting room. Meeting ID: %s, Room: %s", event.Payload.MeetingID, event.Payload.Room)
	case "waiting-participant.left":
		log.Printf("Participant left waiting room. Meeting ID: %s, Room: %s", event.Payload.MeetingID, event.Payload.Room)

	case "transcript.started":
		log.Printf("Transcription started. Meeting ID: %s, Room: %s", event.Payload.MeetingID, event.Payload.Room)
	case "transcript.ready-to-download":
		log.Printf("Transcription ready to download. Meeting ID: %s, Room: %s", event.Payload.MeetingID, event.Payload.Room)

	case "dialout.started":
		log.Printf("Dial-out started. Meeting ID: %s, Room: %s", event.Payload.MeetingID, event.Payload.Room)
	case "dialout.answered":
		log.Printf("Dial-out answered. Meeting ID: %s, Room: %s", event.Payload.MeetingID, event.Payload.Room)
	case "dialout.stopped":
		log.Printf("Dial-out stopped. Meeting ID: %s, Room: %s", event.Payload.MeetingID, event.Payload.Room)

	default:
		log.Printf("Unhandled event type: %s", event.Type)
	}

	// Return success response
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "success"}`))
}
