package main

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"github.com/david-botos/BearHug/services/analysis/internal/processor"
	"github.com/david-botos/BearHug/services/analysis/internal/processor/structOutputs"
	"github.com/david-botos/BearHug/services/analysis/internal/supabase"
	"github.com/david-botos/BearHug/services/analysis/internal/types"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Create a package-level logger variable
var logger zerolog.Logger

// initLogger initializes the global logger
func initLogger() {
	logger = zerolog.New(os.Stdout).With().Timestamp().Logger()
}

func main() {
	// Initialize zerolog
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	initLogger()

	// Configure routes
	http.HandleFunc("/transcript", handleTranscript)
	http.HandleFunc("/GenerateServicesPrompt", handleGenerateServicesPrompt)

	log.Info().
		Str("port", "8080").
		Msg("Server starting")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal().
			Err(err).
			Msg("Server failed to start")
	}
}

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

func handleGenerateServicesPrompt(w http.ResponseWriter, r *http.Request) {
	logger.Info().
		Str("method", r.Method).
		Str("path", r.URL.Path).
		Msg("Handling generate services prompt request")

	// Parse request body
	var req genServicesPrompt
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error().
			Err(err).
			Msg("Failed to parse request body")
		writeErrorResponse(w, http.StatusBadRequest, "invalid_request", "Failed to parse request body")
		return
	}

	// Log parsed request details
	logger.Debug().
		Str("organization_id", req.OrganizationID).
		Int("transcript_length", len(req.Transcript)).
		Msg("Parsed request body")

	// Validate required fields
	if req.OrganizationID == "" {
		logger.Error().
			Msg("Missing organization ID in request")
		writeErrorResponse(w, http.StatusBadRequest, "missing_organization_id", "Organization ID is required")
		return
	}

	if req.Transcript == "" {
		logger.Error().
			Msg("Missing transcript in request")
		writeErrorResponse(w, http.StatusBadRequest, "missing_transcript", "Transcript is required")
		return
	}

	logger.Info().
		Str("organization_id", req.OrganizationID).
		Msg("Processing request for organization")

	// Generate prompt
	prompt, schema, err := structOutputs.GenerateServicesPrompt(req.OrganizationID, req.Transcript)
	if err != nil {
		var statusCode int
		var errorCode string

		// Determine error type and log appropriately
		switch {
		case strings.Contains(err.Error(), "organization_lookup_failed"):
			statusCode = http.StatusNotFound
			errorCode = "organization_not_found"
			logger.Error().
				Err(err).
				Str("organization_id", req.OrganizationID).
				Msg("Organization lookup failed")

		case strings.Contains(err.Error(), "services_lookup_failed"):
			statusCode = http.StatusInternalServerError
			errorCode = "services_lookup_failed"
			logger.Error().
				Err(err).
				Str("organization_id", req.OrganizationID).
				Msg("Services lookup failed")

		default:
			statusCode = http.StatusInternalServerError
			errorCode = "internal_error"
			logger.Error().
				Err(err).
				Msg("Unexpected error during prompt generation")
		}

		writeErrorResponse(w, statusCode, errorCode, err.Error())
		return
	}

	// Prepare successful response
	response := GenerateServicesPromptResponse{
		Status:                  "OK",
		Message:                 "Services prompt generated successfully",
		GeneratedServicesPrompt: prompt,
		IsSchemaPresent:         schema != nil,
	}

	// Write success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Error().
			Err(err).
			Msg("Failed to write response")
		return
	}

	logger.Info().
		Bool("schema_present", schema != nil).
		Int("prompt_length", len(prompt)).
		Msg("Successfully processed request")
}

func handleTranscript(w http.ResponseWriter, r *http.Request) {
	logger.Info().
		Str("method", r.Method).
		Str("remote_addr", r.RemoteAddr).
		Msg("Received transcript request")

	if r.Method != http.MethodPost {
		logger.Warn().
			Str("method", r.Method).
			Msg("Invalid method used")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var reqBody types.TranscriptsReqBody
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		logger.Error().
			Err(err).
			Msg("Failed to decode request body")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	logger.Info().
		Str("organization_id", reqBody.OrganizationID).
		Str("room_url", reqBody.RoomURL).
		Int("service_categories_count", len(reqBody.ServiceCategories)).
		Msg("Processing transcript request")

	// Validate service categories
	for _, category := range reqBody.ServiceCategories {
		if !category.IsValid() {
			logger.Error().
				Interface("category", category).
				Msg("Invalid service category provided")
			http.Error(w, "Invalid service category provided", http.StatusBadRequest)
			return
		}
	}

	// Create storage params
	callData := types.StoreCallDataParams{
		Organization: reqBody.OrganizationID,
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
		logger.Info().Msg("Starting async storage operation")

		if err := supabase.StoreCallData(callData); err != nil {
			logger.Error().
				Err(err).
				Msg("Failed to store transcript")
			return
		}

		logger.Info().Msg("Successfully stored transcript data")
	}()

	// Process transcript asynchronously
	go func() {
		logger.Info().Msg("Starting async transcript processing")

		result, err := processor.ProcessTranscript(dataForAnalysis)
		if err != nil {
			logger.Error().
				Err(err).
				Msg("Failed to process transcript")
			return
		}

		if result {
			logger.Info().
				Interface("result", result).
				Msg("Successfully processed transcript")
		} else {
			logger.Warn().
				Msg("Transcript processing completed with no result")
		}
	}()

	logger.Info().Msg("Request accepted, async processing initiated")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)

	response := map[string]string{
		"status":  "accepted",
		"message": "Transcript received and processing started",
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Error().
			Err(err).
			Msg("Failed to encode response")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// writeErrorResponse handles writing error responses with consistent logging
func writeErrorResponse(w http.ResponseWriter, statusCode int, errorCode, message string) {
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
		logger.Error().
			Err(err).
			Msg("Failed to write error response")
		return
	}

	logger.Error().
		Str("error_code", errorCode).
		Str("error_message", message).
		Int("status_code", statusCode).
		Msg("Request failed")
}
