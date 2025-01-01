package processor

import "net/http"

// Client represents an Anthropic API client
type ClaudeClient struct {
	apiKey     string
	httpClient *http.Client
}

// NewClient creates a new Anthropic API client
func NewClient(apiKey string) *ClaudeClient {
	return &ClaudeClient{
		apiKey:     apiKey,
		httpClient: &http.Client{},
	}
}
