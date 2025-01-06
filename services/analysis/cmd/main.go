package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/david-botos/BearHug/services/analysis/internal/processor"
	"github.com/david-botos/BearHug/services/analysis/internal/supabase"
	"github.com/david-botos/BearHug/services/analysis/internal/types"
	"github.com/rs/zerolog"
)

// Create a package-level logger variable
var logger zerolog.Logger

// initLogger initializes the global logger
func initLogger() {
	output := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: "15:04:05",
		FormatLevel: func(i interface{}) string {
			switch i.(string) {
			case "info":
				return "🟢 INFO"
			case "debug":
				return "🔍 DEBUG"
			case "warn":
				return "⚠️  WARN"
			case "error":
				return "❌ ERROR"
			case "fatal":
				return "💀 FATAL"
			default:
				return "   " + strings.ToUpper(fmt.Sprint(i))
			}
		},
		FormatMessage: func(i interface{}) string {
			return fmt.Sprintf("| %s |", i)
		},
		FormatFieldName: func(i interface{}) string {
			return fmt.Sprintf("%s:", i)
		},
		FormatFieldValue: func(i interface{}) string {
			return strings.ToUpper(fmt.Sprint(i))
		},
	}

	logger = zerolog.New(output).
		With().
		Timestamp().
		Logger()
}

func main() {
	initLogger()

	// Configure routes
	http.HandleFunc("/transcript", handleTranscript)
	// http.HandleFunc("/GenerateServicesPrompt", handleGenerateServicesPrompt)

	logger.Info().
		Str("port", "8080").
		Msg("Server starting")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		logger.Fatal().
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
		Msg("Processing transcript request")

	// Store transcript asynchronously
	go func() {
		logger.Info().Msg("Starting async storage operation")

		if err := supabase.StoreCallData(reqBody); err != nil {
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
		result, err := processor.ProcessTranscript(reqBody)
		if err != nil {
			logger.Error().
				Err(err).
				Msg("Failed to process transcript")

			// Write error response
			writeErrorResponse(w, http.StatusInternalServerError, "processing_failed", err.Error())
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

// func handleGenerateServicesPrompt(w http.ResponseWriter, r *http.Request) {
// 	logger.Info().
// 		Str("method", r.Method).
// 		Str("path", r.URL.Path).
// 		Msg("Handling generate services prompt request")

// 	// Parse request body
// 	var req genServicesPrompt
// 	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
// 		logger.Error().
// 			Err(err).
// 			Msg("Failed to parse request body")
// 		writeErrorResponse(w, http.StatusBadRequest, "invalid_request", "Failed to parse request body")
// 		return
// 	}

// 	// Log parsed request details
// 	logger.Debug().
// 		Str("organization_id", req.OrganizationID).
// 		Int("transcript_length", len(req.Transcript)).
// 		Msg("Parsed request body")

// 	// Validate required fields
// 	if req.OrganizationID == "" {
// 		logger.Error().
// 			Msg("Missing organization ID in request")
// 		writeErrorResponse(w, http.StatusBadRequest, "missing_organization_id", "Organization ID is required")
// 		return
// 	}

// 	if req.Transcript == "" {
// 		logger.Error().
// 			Msg("Missing transcript in request")
// 		writeErrorResponse(w, http.StatusBadRequest, "missing_transcript", "Transcript is required")
// 		return
// 	}

// 	logger.Info().
// 		Str("organization_id", req.OrganizationID).
// 		Msg("Processing request for organization")

// 	// Generate prompt
// 	prompt, schema, err := structOutputs.GenerateServicesPrompt(req.OrganizationID, req.Transcript)
// 	if err != nil {
// 		var statusCode int
// 		var errorCode string

// 		// Determine error type and log appropriately
// 		switch {
// 		case strings.Contains(err.Error(), "organization_lookup_failed"):
// 			statusCode = http.StatusNotFound
// 			errorCode = "organization_not_found"
// 			logger.Error().
// 				Err(err).
// 				Str("organization_id", req.OrganizationID).
// 				Msg("Organization lookup failed")

// 		case strings.Contains(err.Error(), "services_lookup_failed"):
// 			statusCode = http.StatusInternalServerError
// 			errorCode = "services_lookup_failed"
// 			logger.Error().
// 				Err(err).
// 				Str("organization_id", req.OrganizationID).
// 				Msg("Services lookup failed")

// 		default:
// 			statusCode = http.StatusInternalServerError
// 			errorCode = "internal_error"
// 			logger.Error().
// 				Err(err).
// 				Msg("Unexpected error during prompt generation")
// 		}

// 		writeErrorResponse(w, statusCode, errorCode, err.Error())
// 		return
// 	}

// 	// Prepare successful response
// 	response := GenerateServicesPromptResponse{
// 		Status:                  "OK",
// 		Message:                 "Services prompt generated successfully",
// 		GeneratedServicesPrompt: prompt,
// 		IsSchemaPresent:         schema != nil,
// 	}

// 	// Write success response
// 	w.Header().Set("Content-Type", "application/json")
// 	w.WriteHeader(http.StatusOK)

// 	if err := json.NewEncoder(w).Encode(response); err != nil {
// 		logger.Error().
// 			Err(err).
// 			Msg("Failed to write response")
// 		return
// 	}

// 	logger.Info().
// 		Bool("schema_present", schema != nil).
// 		Int("prompt_length", len(prompt)).
// 		Msg("Successfully processed request")
// }
