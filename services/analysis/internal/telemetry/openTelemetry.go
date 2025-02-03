package telemetry

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/david-botos/BearHug/services/analysis/pkg/logger"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

func InitTracer(phoenixAPIKey string, spaceID string) (func(), error) {
	ctx := context.Background()
	log := logger.Get()

	log.Info().Msg("Initializing OpenTelemetry tracer...")

	// Create resource with detailed attributes
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String("transcript-analysis-service"),
			semconv.ServiceVersionKey.String("1.0.0"),
			semconv.ServiceInstanceIDKey.String("test-instance"),
			semconv.DeploymentEnvironmentKey.String("development"),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	log.Info().Msg("Created OpenTelemetry resource")

	// Configure HTTP exporter
	traceExporter, err := otlptracehttp.New(
		ctx,
		otlptracehttp.WithEndpoint("app.phoenix.arize.com"),
		otlptracehttp.WithURLPath("/v1/traces"),
		otlptracehttp.WithHeaders(map[string]string{
			"Authorization": phoenixAPIKey,
			"space-id":      spaceID,
			"Content-Type":  "application/x-protobuf",
		}),
		otlptracehttp.WithCompression(otlptracehttp.GzipCompression),
		otlptracehttp.WithTimeout(30*time.Second),
		otlptracehttp.WithRetry(otlptracehttp.RetryConfig{
			Enabled:         true,
			InitialInterval: 1 * time.Second,
			MaxInterval:     5 * time.Second,
			MaxElapsedTime:  30 * time.Second,
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	log.Info().Msg("Created OTLP HTTP exporter")

	// Configure batch span processor with conservative settings
	bsp := sdktrace.NewBatchSpanProcessor(traceExporter,
		sdktrace.WithBatchTimeout(5*time.Second),
		sdktrace.WithMaxExportBatchSize(256),
		sdktrace.WithMaxQueueSize(2048),
		sdktrace.WithExportTimeout(20*time.Second),
	)

	// Create and set tracer provider
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
	)
	otel.SetTracerProvider(tracerProvider)

	// Set error handler for debugging
	otel.SetErrorHandler(otel.ErrorHandlerFunc(func(err error) {
		log.Error().Err(err).Msg("OpenTelemetry error occurred")
	}))

	// Return cleanup function
	return func() {
		cctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := tracerProvider.Shutdown(cctx); err != nil {
			log.Error().Err(err).Msg("Error shutting down tracer provider")
		}
	}, nil
}

// Custom transport for debugging
type debugTransport struct {
	underlying http.RoundTripper
	logger     *zerolog.Logger
}

func (t *debugTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Log request details
	t.logger.Info().
		Str("method", req.Method).
		Str("url", req.URL.String()).
		Interface("headers", req.Header).
		Msg("Sending request to Phoenix")

	// Read and log request body if present
	if req.Body != nil {
		bodyBytes, err := io.ReadAll(req.Body)
		if err == nil {
			t.logger.Info().
				Int("body_size", len(bodyBytes)).
				Msg("Request body size")
			// Reset the body for the actual request
			req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}
	}

	// Make the request
	resp, err := t.underlying.RoundTrip(req)
	if err != nil {
		return resp, err
	}

	// Log response details
	t.logger.Info().
		Int("status", resp.StatusCode).
		Interface("headers", resp.Header).
		Msg("Received response from Phoenix")

	// Read and log response body if present
	if resp.Body != nil {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err == nil {
			t.logger.Info().
				Str("body", string(bodyBytes)).
				Msg("Response body")
			// Reset the body for the actual response
			resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}
	}

	return resp, err
}
