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

type InferenceResponse struct {
	Content      []MessageContent `json:"content"`
	Role         string           `json:"role"`
	Model        string           `json:"model"`
	ID           string           `json:"id"`
	Type         string           `json:"type"`
	Usage        Usage            `json:"usage"`
	StopReason   string           `json:"stop_reason"`
	StopSequence interface{}      `json:"stop_sequence"`
}

type MessageContent struct {
	Type  string      `json:"type"`
	Text  string      `json:"text,omitempty"`
	ID    string      `json:"id,omitempty"`
	Name  string      `json:"name,omitempty"`
	Input interface{} `json:"input,omitempty"`
}

type ToolUseData struct {
	ID    string      `json:"id"`
	Name  string      `json:"name"`
	Input interface{} `json:"input"`
}

type Usage struct {
	InputTokens              int `json:"input_tokens"`
	CacheCreationInputTokens int `json:"cache_creation_input_tokens"`
	CacheReadInputTokens     int `json:"cache_read_input_tokens"`
	OutputTokens             int `json:"output_tokens"`
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
func (c *ClaudeClient) RunClaudeInference(params TriagePromptParams) (map[string]interface{}, error) {
	// Create request body
	reqBody := TriagePromptRequest{
		Model:     "claude-3-5-sonnet-20241022",
		MaxTokens: 1500,
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

	// LOG
	prettyJSON, err := json.MarshalIndent(reqBody, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %w", err)
	}
	fmt.Printf("Request body being sent to Claude:\n%s\n", string(prettyJSON))
	// END LOG

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

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	// LOG raw body first
	fmt.Printf("Raw body before processing:\n%s\n", string(body))

	// Parse response
	var inferenceResp InferenceResponse
	if err := json.Unmarshal(body, &inferenceResp); err != nil {
		return nil, fmt.Errorf("error parsing response: %w", err)
	}

	// LOG the parsed response
	fmt.Printf("Parsed response Content length: %d\n", len(inferenceResp.Content))
	for i, content := range inferenceResp.Content {
		fmt.Printf("Content[%d] Type: %s\n", i, content.Type)
		if content.Type == "tool_use" {
			fmt.Printf("Found tool_use content at index %d\n", i)
			fmt.Printf("Input: %+v\n", content.Input)
		}
	}

	var toolOutput interface{}
	for _, content := range inferenceResp.Content {
		if content.Type == "tool_use" {
			toolOutput = content.Input
			fmt.Printf("Setting toolOutput: %+v\n", content.Input)
			break
		}
	}

	if toolOutput == nil {
		fmt.Printf("toolOutput is nil after processing\n")
		return nil, fmt.Errorf("no structured output found in response")
	}

	// LOG
	var prettyToolOutput bytes.Buffer
	jsonBytes, err := json.Marshal(toolOutput)
	if err != nil {
		return nil, fmt.Errorf("error marshaling tool output: %v", err)
	}
	if err := json.Indent(&prettyToolOutput, jsonBytes, "", "  "); err != nil {
		return nil, fmt.Errorf("error formatting JSON: %v", err)
	}
	fmt.Printf("Formatted tool output:\n%s\n", prettyToolOutput.String())
	// END LOG

	// Validate the tool output against the schema
	if err := validateAgainstSchema(toolOutput, params.Schema); err != nil {
		return nil, fmt.Errorf("response validation failed: %w", err)
	}

	contentMap, ok := toolOutput.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected response format")
	}

	return contentMap, nil
}
