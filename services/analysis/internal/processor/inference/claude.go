package inference

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	env "github.com/david-botos/BearHug/services/analysis/pkg/ENV"
	"github.com/david-botos/BearHug/services/analysis/pkg/logger"
	"github.com/rs/zerolog"
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
func (c *ClaudeClient) RunClaudeInference(params PromptParams) (map[string]interface{}, error) {
	log := logger.Get()
	maxRetries := 2
	retryDelay := 10 * time.Second

	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			log.Info().
				Int("attempt", attempt).
				Dur("delay", retryDelay).
				Msg("Retrying Claude API request after delay")
			time.Sleep(retryDelay)
		}

		response, err := c.makeInferenceRequest(params, log)
		if err != nil {
			lastErr = err
			if isOverburdenedError(err) && attempt < maxRetries {
				continue
			}
			return nil, err
		}
		return response, nil
	}

	return nil, fmt.Errorf("anthropic API is overburdened: %w", lastErr)
}

func (c *ClaudeClient) makeInferenceRequest(params PromptParams, log *zerolog.Logger) (map[string]interface{}, error) {
	log.Debug().
		Str("model", "claude-3-5-sonnet-20241022").
		Int("max_tokens", 1500).
		Msg("Starting Claude inference request")

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

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal request body")
		return nil, fmt.Errorf("error marshaling request: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Error().Err(err).Msg("Failed to create HTTP request")
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	log.Debug().
		Str("method", req.Method).
		Str("url", req.URL.String()).
		Msg("Sending request to Claude API")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		log.Error().Err(err).Msg("Failed to execute HTTP request")
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 529 {
		return nil, fmt.Errorf("received status code 529")
	}

	log.Debug().
		Int("status_code", resp.StatusCode).
		Msg("Received response from Claude API")

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error().Err(err).Msg("Failed to read response body")
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	var inferenceResp InferenceResponse
	if err := json.Unmarshal(body, &inferenceResp); err != nil {
		log.Error().
			Err(err).
			Str("body", string(body)).
			Msg("Failed to parse response body")
		return nil, fmt.Errorf("error parsing response: %w", err)
	}

	log.Debug().
		Int("content_length", len(inferenceResp.Content)).
		Str("model", inferenceResp.Model).
		Str("stop_reason", inferenceResp.StopReason).
		Interface("usage", inferenceResp.Usage).
		Msg("Successfully parsed inference response")

	var toolOutput interface{}
	for _, content := range inferenceResp.Content {
		if content.Type == "tool_use" {
			toolOutput = content.Input
			log.Debug().
				Interface("tool_output", toolOutput).
				Msg("Found tool output in response")
			break
		}
	}

	if toolOutput == nil {
		log.Error().Msg("No structured output found in response")
		return nil, fmt.Errorf("no structured output found in response")
	}

	if err := validateAgainstSchema(toolOutput, params.Schema); err != nil {
		log.Error().
			Err(err).
			Interface("tool_output", toolOutput).
			Msg("Response validation failed")
		return nil, fmt.Errorf("response validation failed: %w", err)
	}

	contentMap, ok := toolOutput.(map[string]interface{})
	if !ok {
		log.Error().
			Interface("tool_output", toolOutput).
			Msg("Unexpected response format")
		return nil, fmt.Errorf("unexpected response format")
	}

	log.Info().
		Int("fields_count", len(contentMap)).
		Msg("Successfully processed Claude inference request")

	return contentMap, nil
}

func isOverburdenedError(err error) bool {
	return strings.Contains(err.Error(), "529")
}
