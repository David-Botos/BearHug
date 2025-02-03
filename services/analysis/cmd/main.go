package main

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/david-botos/BearHug/services/analysis/internal/processor"
	"github.com/david-botos/BearHug/services/analysis/internal/supabase"
	"github.com/david-botos/BearHug/services/analysis/internal/telemetry"
	"github.com/david-botos/BearHug/services/analysis/internal/types"
	"github.com/david-botos/BearHug/services/analysis/pkg/logger"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type GenerateErrorResponse struct {
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

func handleTest(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.Get()

	// Create a tracer with a specific name for testing
	tracer := otel.GetTracerProvider().Tracer("test-handler")

	// Start the parent span
	ctx, parentSpan := tracer.Start(ctx, "test_endpoint",
		trace.WithAttributes(
			attribute.String("http.method", r.Method),
			attribute.String("http.url", r.URL.String()),
			attribute.String("test.type", "phoenix_verification"),
			attribute.String("service.name", "transcript-analysis"),
			attribute.Int64("timestamp", time.Now().Unix()),
		),
	)
	defer parentSpan.End()

	// Add some events to the parent span
	parentSpan.AddEvent("test_started", trace.WithAttributes(
		attribute.String("event.type", "test"),
		attribute.String("event.status", "started"),
	))

	// Create a child span to test span hierarchy
	_, childSpan := tracer.Start(ctx, "test_subprocess",
		trace.WithAttributes(
			attribute.String("subprocess.type", "verification"),
			attribute.String("subprocess.status", "running"),
		),
	)

	// Simulate some work
	time.Sleep(100 * time.Millisecond)

	childSpan.SetAttributes(attribute.String("subprocess.status", "completed"))
	childSpan.End()

	// Add completion event to parent span
	parentSpan.AddEvent("test_completed", trace.WithAttributes(
		attribute.String("event.type", "test"),
		attribute.String("event.status", "completed"),
	))

	// Log that we're sending trace data
	log.Info().
		Str("trace_id", parentSpan.SpanContext().TraceID().String()).
		Str("span_id", parentSpan.SpanContext().SpanID().String()).
		Msg("Test spans created and sent to Phoenix")

	// Return a simple response
	w.Header().Set("Content-Type", "application/json")
	response := map[string]string{
		"status":   "success",
		"message":  "Test spans created",
		"trace_id": parentSpan.SpanContext().TraceID().String(),
	}

	json.NewEncoder(w).Encode(response)
}

func handleTranscript(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tracer := otel.GetTracerProvider().Tracer("http-handler")

	ctx, span := tracer.Start(ctx, "handle_transcript_request",
		trace.WithAttributes(
			attribute.String("http.method", r.Method),
			attribute.String("http.url", r.URL.String()),
		),
	)
	defer span.End()

	log := logger.Get()
	requestID := r.Header.Get("X-Request-ID")
	span.SetAttributes(attribute.String("request_id", requestID))

	log.Info().
		Str("method", r.Method).
		Str("path", r.URL.Path).
		Str("remote_addr", r.RemoteAddr).
		Str("request_id", requestID).
		Msg("Processing incoming transcript request")

	if r.Method != http.MethodPost {
		span.SetAttributes(attribute.String("error", "method_not_allowed"))
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
		span.RecordError(err)
		span.SetAttributes(attribute.String("error", "invalid_request_body"))
		log.Error().
			Err(err).
			Str("request_id", requestID).
			Msg("Failed to parse request body as JSON")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	span.SetAttributes(
		attribute.String("organization_id", reqBody.OrganizationID),
		attribute.String("room_url", reqBody.RoomURL),
		attribute.Int("transcript_length", len(reqBody.Transcript)),
	)

	// Store transcript synchronously
	ctx, storeSpan := tracer.Start(ctx, "store_call_data")
	callID, err := supabase.StoreCallData(ctx, reqBody)
	if err != nil {
		storeSpan.RecordError(err)
		storeSpan.End()
		span.RecordError(err)
		log.Error().
			Err(err).
			Str("request_id", requestID).
			Str("organization_id", reqBody.OrganizationID).
			Msg("Failed to persist transcript data in storage")
		writeErrorResponse(w, http.StatusInternalServerError, "storage_failed", err.Error())
		return
	}
	storeSpan.SetAttributes(attribute.String("call_id", callID))
	storeSpan.End()

	// Create a new reqBody with callID
	procTranscriptParams := types.ProcTranscriptParams{
		OrganizationID: reqBody.OrganizationID,
		RoomURL:        reqBody.RoomURL,
		Transcript:     reqBody.Transcript,
		CallID:         callID,
	}

	// Process transcript synchronously with context
	result, err := processor.ProcessTranscript(ctx, procTranscriptParams)
	if err != nil {
		span.RecordError(err)
		log.Error().
			Err(err).
			Str("request_id", requestID).
			Str("organization_id", reqBody.OrganizationID).
			Msg("Transcript analysis failed")
		writeErrorResponse(w, http.StatusInternalServerError, "processing_failed", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := map[string]interface{}{
		"status":  "success",
		"message": "Transcript processed successfully",
		"call_id": callID,
		"result":  result,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		span.RecordError(err)
		log.Error().
			Err(err).
			Str("request_id", requestID).
			Msg("Failed to serialize response JSON")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	span.SetAttributes(
		attribute.Bool("success", true),
		attribute.String("status", "success"),
	)
}

func writeErrorResponse(w http.ResponseWriter, statusCode int, errorCode, message string) {
	log := logger.Get()

	response := GenerateErrorResponse{
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

	arizeAPIKey := os.Getenv("ARIZE_API_KEY")
	if arizeAPIKey == "" {
		log.Error().Msg("ARIZE_API_KEY environment variable is required")
	}
	arizeSpaceID := os.Getenv("ARIZE_SPACEID")
	if arizeSpaceID == "" {
		log.Error().Msg("ARIZE_SPACEID environment variable is required")
	}

	cleanup, err := telemetry.InitTracer(arizeAPIKey, arizeSpaceID)
	if err != nil {
		log.Error().Str("Failed to initialize tracer: ", err.Error())
	}
	defer cleanup()

	// Configure routes
	http.HandleFunc("/transcript", handleTranscript)
	http.HandleFunc("/test", handleTest)

	port := "8500"
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
