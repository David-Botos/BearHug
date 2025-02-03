package inference

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	env "github.com/david-botos/BearHug/services/analysis/pkg/ENV"
	"github.com/david-botos/BearHug/services/analysis/pkg/logger"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type PromptParams struct {
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

// InitInferenceClient initializes the Claude inference client
func InitInferenceClient() (*ClaudeClient, error) {
	log := logger.Get()
	log.Debug().Msg("Initializing Claude inference client")

	if err := env.LoadEnvFile(); err != nil {
		log.Error().Err(err).Msg("Failed to load development environment")
		return nil, err
	}

	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		log.Error().Msg("ANTHROPIC_API_KEY not found in environment")
		return nil, fmt.Errorf("ANTHROPIC_API_KEY not found in environment")
	}
	return NewClient(apiKey), nil
}

// RunClaudeInference performs inference with structured output validation
func (c *ClaudeClient) RunClaudeInference(ctx context.Context, params PromptParams) (map[string]interface{}, error) {
	ctx, span := c.tracer.Start(ctx, "claude.inference",
		trace.WithAttributes(
			attribute.String("model", "claude-3-5-sonnet-20241022"),
			attribute.Int("max_tokens", 1500),
			attribute.String("inference_type", params.Schema.Type),
		),
	)
	defer span.End()

	log := logger.Get()
	log.Debug().
		Str("model", "claude-3-5-sonnet-20241022").
		Int("max_tokens", 1500).
		Msg("Starting Claude inference request")

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

	// Add request preparation span
	_, prepSpan := c.tracer.Start(ctx, "prepare_request")
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		prepSpan.RecordError(err)
		prepSpan.End()
		return nil, fmt.Errorf("error marshaling request: %w", err)
	}
	prepSpan.End()

	// Create and execute request with tracing
	_, httpSpan := c.tracer.Start(ctx, "http_request")
	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.anthropic.com/v1/messages", bytes.NewBuffer(jsonData))
	if err != nil {
		httpSpan.RecordError(err)
		httpSpan.End()
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		httpSpan.RecordError(err)
		httpSpan.End()
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()
	httpSpan.End()

	// Process response with tracing
	_, processSpan := c.tracer.Start(ctx, "process_response")
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		processSpan.RecordError(err)
		processSpan.End()
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	var inferenceResp InferenceResponse
	if err := json.Unmarshal(body, &inferenceResp); err != nil {
		processSpan.RecordError(err)
		processSpan.End()
		return nil, fmt.Errorf("error parsing response: %w", err)
	}

	// Add response metadata to span
	processSpan.SetAttributes(
		attribute.Int("response.token_count", inferenceResp.Usage.InputTokens+inferenceResp.Usage.OutputTokens),
		attribute.String("response.stop_reason", inferenceResp.StopReason),
	)

	var toolOutput interface{}
	for _, content := range inferenceResp.Content {
		if content.Type == "tool_use" {
			toolOutput = content.Input
			break
		}
	}

	if toolOutput == nil {
		processSpan.RecordError(fmt.Errorf("no structured output found"))
		processSpan.End()
		return nil, fmt.Errorf("no structured output found in response")
	}

	// Validate output with tracing
	_, validateSpan := c.tracer.Start(ctx, "validate_output")
	if err := validateAgainstSchema(toolOutput, params.Schema); err != nil {
		validateSpan.RecordError(err)
		validateSpan.End()
		processSpan.End()
		return nil, fmt.Errorf("response validation failed: %w", err)
	}
	validateSpan.End()

	contentMap, ok := toolOutput.(map[string]interface{})
	if !ok {
		processSpan.RecordError(fmt.Errorf("unexpected response format"))
		processSpan.End()
		return nil, fmt.Errorf("unexpected response format")
	}

	processSpan.End()
	return contentMap, nil
}
