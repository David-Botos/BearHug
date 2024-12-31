package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/david-botos/BearHug/services/analysis/internal/processor"
	"github.com/david-botos/BearHug/services/analysis/internal/storage"
	"github.com/david-botos/BearHug/services/analysis/internal/types"
)

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

	var reqBody types.TranscriptsReqBody
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate service categories
	for _, category := range reqBody.ServiceCategories {
		if !category.IsValid() {
			http.Error(w, "Invalid service category provided", http.StatusBadRequest)
			return
		}
	}

	// Create storage params
	callData := types.StoreCallDataParams{
		Organization: reqBody.Organization,
		RoomURL:      reqBody.RoomURL,
		Transcript:   reqBody.Transcript,
	}

	// Create processor params
	dataForAnalysis := types.ProcessTranscriptParams{
		ServiceCategories: reqBody.ServiceCategories,
		RoomURL:           reqBody.RoomURL,
		Transcript:        reqBody.Transcript,
	}

	// Store transcript asynchronously
	go func() {
		if err := storage.StoreCallData(callData); err != nil {
			log.Printf("Error storing transcript: %v", err)
		}
	}()

	// Process transcript asynchronously
	go func() {
		if err := processor.ProcessTranscript(dataForAnalysis); err != nil {
			log.Printf("Error processing transcript: %v", err)
		}
	}()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "accepted",
		"message": "Transcript received and processing started",
	})
}
