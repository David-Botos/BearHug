package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/david-botos/BearHug/services/analysis/internal/processor"
	"github.com/david-botos/BearHug/services/analysis/internal/storage"
	"github.com/david-botos/BearHug/services/analysis/internal/types"
)

func main() {
	http.HandleFunc("/transcript", handleTranscript)
	log.Printf("INFO: Server starting on port :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("FATAL: Server failed to start: %v", err)
	}
}

func handleTranscript(w http.ResponseWriter, r *http.Request) {
	requestID := generateRequestID()
	logger := createContextLogger(requestID)

	logger.Printf("INFO: Received transcript request from %s", r.RemoteAddr)

	if r.Method != http.MethodPost {
		logger.Printf("WARN: Invalid method %s used", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var reqBody types.TranscriptsReqBody
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		logger.Printf("ERROR: Failed to decode request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	logger.Printf("INFO: Processing request for organization: %s, room: %s",
		reqBody.Organization, reqBody.RoomURL)

	// Validate service categories
	for _, category := range reqBody.ServiceCategories {
		if !category.IsValid() {
			logger.Printf("ERROR: Invalid service category provided: %v", category)
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
		logger.Print("INFO: Starting async storage operation")
		if err := storage.StoreCallData(callData); err != nil {
			logger.Printf("ERROR: Failed to store transcript: %v", err)
			return
		}
		logger.Print("INFO: Successfully stored transcript data")
	}()

	// Process transcript asynchronously
	go func() {
		logger.Print("INFO: Starting async transcript processing")
		result, err := processor.ProcessTranscript(dataForAnalysis)
		if err != nil {
			logger.Printf("ERROR: Failed to process transcript: %v", err)
			return
		}

		if result {
			logger.Printf("INFO: Successfully processed transcript with result: %v", result)
			// TODO: Actually handle result
		} else {
			logger.Print("WARN: Transcript processing completed with no result")
		}
	}()

	logger.Print("INFO: Request accepted, async processing initiated")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	response := map[string]string{
		"status":     "accepted",
		"message":    "Transcript received and processing started",
		"request_id": requestID,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Printf("ERROR: Failed to encode response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// Helper functions for logging

func generateRequestID() string {
	// Generate 6 random bytes
	b := make([]byte, 3) // Will become 6 hex chars
	if _, err := rand.Read(b); err != nil {
		// Fallback to timestamp only if crypto rand fails
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return fmt.Sprintf("%d-%x", time.Now().UnixNano(), b)
}

func createContextLogger(requestID string) *log.Logger {
	return log.New(os.Stdout,
		fmt.Sprintf("[RequestID: %s] ", requestID),
		log.Ldate|log.Ltime|log.Lmicroseconds)
}
