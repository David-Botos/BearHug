package inference

import (
	"fmt"
	"net/http"
	"reflect"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// Client represents an Anthropic API client
type ClaudeClient struct {
	apiKey     string
	httpClient *http.Client
	tracer     trace.Tracer
}

// NewClient creates a new Anthropic API client
func NewClient(apiKey string) *ClaudeClient {
	tracer := otel.GetTracerProvider().Tracer("claude-client")
	return &ClaudeClient{
		apiKey:     apiKey,
		httpClient: &http.Client{},
		tracer:     tracer,
	}
}

// validateAgainstSchema checks if the data matches the schema definition
func validateAgainstSchema(data interface{}, schema ToolInputSchema) error {
	// Convert data to a map for easier validation
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return fmt.Errorf("response data is not a JSON object")
	}

	// Validate required fields
	for _, required := range schema.Required {
		if _, exists := dataMap[required]; !exists {
			return fmt.Errorf("missing required field: %s", required)
		}
	}

	// Validate property types
	for fieldName, property := range schema.Properties {
		value, exists := dataMap[fieldName]
		if !exists {
			continue // Skip if field is not required
		}

		// Check type
		valueType := reflect.TypeOf(value)
		switch property.Type {
		case "string":
			if valueType.Kind() != reflect.String {
				return fmt.Errorf("field %s: expected string, got %v", fieldName, valueType)
			}
		case "number":
			if valueType.Kind() != reflect.Float64 && valueType.Kind() != reflect.Int {
				return fmt.Errorf("field %s: expected number, got %v", fieldName, valueType)
			}
		case "boolean":
			if valueType.Kind() != reflect.Bool {
				return fmt.Errorf("field %s: expected boolean, got %v", fieldName, valueType)
			}
		case "array":
			if valueType.Kind() != reflect.Slice {
				return fmt.Errorf("field %s: expected array, got %v", fieldName, valueType)
			}
		case "object":
			if valueType.Kind() != reflect.Map {
				return fmt.Errorf("field %s: expected object, got %v", fieldName, valueType)
			}
		}
	}

	return nil
}
