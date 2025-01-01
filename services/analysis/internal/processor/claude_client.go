package processor

import "net/http"

// Client represents an Anthropic API client
type Client struct {
	apiKey     string
	httpClient *http.Client
}

// NewClient creates a new Anthropic API client
func NewClient(apiKey string) *Client {
	return &Client{
		apiKey:     apiKey,
		httpClient: &http.Client{},
	}
}
