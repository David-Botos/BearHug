package main

import (
	"encoding/json"
	"net/http"

	"github.com/david-botos/BearHug/services/analysis/internal/processor"
	"github.com/david-botos/BearHug/services/analysis/internal/supabase"
	"github.com/david-botos/BearHug/services/analysis/internal/types"
	"github.com/david-botos/BearHug/services/analysis/pkg/logger"
)

type genServicesPrompt struct {
	OrganizationID string `json:"organization_id"`
	Transcript     string `json:"transcript"`
}

type GenerateServicesPromptResponse struct {
	Status                  string       `json:"status"`
	Message                 string       `json:"message"`
	GeneratedServicesPrompt string       `json:"generated_services_prompt"`
	IsSchemaPresent         bool         `json:"is_schema_present"`
	Error                   *ErrorDetail `json:"error,omitempty"`
}

type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func handleTranscript(w http.ResponseWriter, r *http.Request) {
	log := logger.Get()
	requestID := r.Header.Get("X-Request-ID")

	log.Info().
		Str("method", r.Method).
		Str("path", r.URL.Path).
		Str("remote_addr", r.RemoteAddr).
		Str("request_id", requestID).
		Msg("Processing incoming transcript request")

	if r.Method != http.MethodPost {
		log.Warn().
			Str("method", r.Method).
			Str("allowed_method", http.MethodPost).
			Str("request_id", requestID).
			Msg("Request rejected due to invalid HTTP method")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var reqBody types.TranscriptsReqBody
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		log.Error().
			Err(err).
			Str("request_id", requestID).
			Msg("Failed to parse request body as JSON")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	log.Info().
		Str("organization_id", reqBody.OrganizationID).
		Str("room_url", reqBody.RoomURL).
		Str("request_id", requestID).
		Int("transcript_length", len(reqBody.Transcript)).
		Msg("Initiating parallel processing operations")

	// Store transcript asynchronously
	go func() {
		log.Info().
			Str("request_id", requestID).
			Str("organization_id", reqBody.OrganizationID).
			Msg("Starting async transcript storage")

		if err := supabase.StoreCallData(reqBody); err != nil {
			log.Error().
				Err(err).
				Str("request_id", requestID).
				Str("organization_id", reqBody.OrganizationID).
				Msg("Failed to persist transcript data in storage")
			return
		}

		log.Info().
			Str("request_id", requestID).
			Str("organization_id", reqBody.OrganizationID).
			Msg("Successfully persisted transcript data")
	}()

	// Process transcript asynchronously
	go func() {
		log.Info().
			Str("request_id", requestID).
			Str("organization_id", reqBody.OrganizationID).
			Msg("Starting async transcript analysis")

		result, err := processor.ProcessTranscript(reqBody)
		if err != nil {
			log.Error().
				Err(err).
				Str("request_id", requestID).
				Str("organization_id", reqBody.OrganizationID).
				Msg("Transcript analysis failed")

			writeErrorResponse(w, http.StatusInternalServerError, "processing_failed", err.Error())
			return
		}

		if result {
			log.Info().
				Str("request_id", requestID).
				Str("organization_id", reqBody.OrganizationID).
				Bool("success", result).
				Msg("Transcript analysis completed successfully")
		} else {
			log.Warn().
				Str("request_id", requestID).
				Str("organization_id", reqBody.OrganizationID).
				Msg("Transcript analysis completed with no actionable results")
		}
	}()

	log.Info().
		Str("request_id", requestID).
		Str("organization_id", reqBody.OrganizationID).
		Msg("Request accepted, async operations initiated")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)

	response := map[string]string{
		"status":  "accepted",
		"message": "Transcript received and processing started",
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Error().
			Err(err).
			Str("request_id", requestID).
			Msg("Failed to serialize response JSON")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func writeErrorResponse(w http.ResponseWriter, statusCode int, errorCode, message string) {
	log := logger.Get()

	response := GenerateServicesPromptResponse{
		Status:  "error",
		Message: "Failed to generate services prompt",
		Error: &ErrorDetail{
			Code:    errorCode,
			Message: message,
		},
		IsSchemaPresent: false,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Error().
			Err(err).
			Str("error_code", errorCode).
			Str("intended_message", message).
			Int("status_code", statusCode).
			Msg("Failed to encode error response")
		return
	}

	log.Error().
		Str("error_code", errorCode).
		Str("error_message", message).
		Int("status_code", statusCode).
		Msg("Request processing failed with error")
}

func main() {
	// Initialize the logger
	logger.Init()
	log := logger.Get()

	// Configure routes
	http.HandleFunc("/transcript", handleTranscript)

	port := "8080"
	log.Info().
		Str("port", port).
		Msg("Starting HTTP server")

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal().
			Err(err).
			Str("port", port).
			Msg("Server failed to start")
	}
}
