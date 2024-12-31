package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/david-botos/bearhug/services/analysis/internal/processor"
)

// ProcessedTranscript represents the processed transcript data
type ProcessedTranscript struct {
	RoomURL string `json:"room_url"`
	Status  string `json:"status"`
}

func main() {
	http.HandleFunc("/transcript", handleTranscript)

	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func handleTranscript(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var transcript processor.Transcript
	if err := json.NewDecoder(r.Body).Decode(&transcript); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Launch goroutine to process transcript
	go processor.processTranscript(transcript)

	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "accepted",
		"message": "Transcript received and processing started",
	})
}
