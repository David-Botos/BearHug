package processor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
)

type TriagePromptParams struct {
	Prompt string          `json:"prompt"`
	Schema ToolInputSchema `json:"schema"`
}

type Tool struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	InputSchema ToolInputSchema `json:"input_schema"`
}

type ToolInputSchema struct {
	Type       string              `json:"type"`
	Properties map[string]Property `json:"properties"`
	Required   []string            `json:"required"`
}

type Property struct {
	Type        string                 `json:"type"`
	Description string                 `json:"description"`
	Items       map[string]interface{} `json:"items,omitempty"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type TriagePromptRequest struct {
	Model     string    `json:"model"`
	MaxTokens int       `json:"max_tokens"`
	Tools     []Tool    `json:"tools"`
	Messages  []Message `json:"messages"`
}

// InferenceResponse represents the response from the Anthropic API
type InferenceResponse struct {
	Content    interface{} `json:"content"`
	Role       string      `json:"role"`
	StopReason string      `json:"stop_reason"`
	Model      string      `json:"model"`
	ResponseID string      `json:"response_id"`
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

// RunClaudeInference performs inference with structured output validation
func (c *Client) RunClaudeInference(params TriagePromptParams) (map[string]interface{}, error) {
	// Create request body
	reqBody := TriagePromptRequest{
		Model:     "claude-3-5-sonnet-20241022",
		MaxTokens: 1024,
		Tools: []Tool{
			{
				Name:        "structured_output",
				Description: "Output should conform to the provided JSON schema",
				InputSchema: params.Schema,
			},
		},
		Messages: []Message{
			{
				Role:    "user",
				Content: params.Prompt,
			},
		},
	}

	// Marshal the request body
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %w", err)
	}

	// Create request
	req, err := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	// Make request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var inferenceResp InferenceResponse
	if err := json.Unmarshal(body, &inferenceResp); err != nil {
		return nil, fmt.Errorf("error parsing response: %w", err)
	}

	// Validate the response against the schema
	if err := validateAgainstSchema(inferenceResp.Content, params.Schema); err != nil {
		return nil, fmt.Errorf("response validation failed: %w", err)
	}

	// Convert content to map[string]interface{} since we've validated it
	contentMap, ok := inferenceResp.Content.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected response format")
	}

	return contentMap, nil
}
