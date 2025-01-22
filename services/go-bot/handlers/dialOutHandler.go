package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

type ServiceConfig struct {
	Service string         `json:"service"`
	Options []ConfigOption `json:"options"`
}

type ConfigOption struct {
	Name  string      `json:"name"`
	Value interface{} `json:"value"`
}

type WebhookTools struct {
	GetEvents struct {
		URL       string `json:"url"`
		Streaming bool   `json:"streaming"`
	} `json:"get_events"`
}

type RequestBody struct {
	Services     map[string]interface{} `json:"services"`
	Config       []ServiceConfig        `json:"config"`
	WebhookTools WebhookTools           `json:"webhook_tools"`
}

type Payload struct {
	BotProfile  string                 `json:"bot_profile"`
	MaxDuration int                    `json:"max_duration"`
	DialOut     []DialOutSettings      `json:"dialout_settings"`
	Services    map[string]interface{} `json:"services"`
	APIKeys     APIKeys                `json:"api_keys"`
	Config      []ServiceConfig        `json:"config"` // Change to an array
}

type DialOutSettings struct {
	PhoneNumber string `json:"phoneNumber"`
}

type APIKeys struct {
	OpenAI   string `json:"openai"`
	Deepgram string `json:"deepgram"`
	Cartesia string `json:"cartesia"`
}

func DialOutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	var reqBody RequestBody
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		log.Printf("Error decoding request body: %v", err)
		http.Error(w, "Failed to parse request body", http.StatusBadRequest)
		return
	}

	// Validate request body and environment variables
	dailyBotsKey := os.Getenv("DAILY_BOTS_KEY")
	if reqBody.Services == nil || len(reqBody.Config) == 0 || dailyBotsKey == "" {
		http.Error(w, `{"error": "Services, config, or DAILY_BOTS_KEY missing."}`, http.StatusBadRequest)
		return
	}

	// Prepare payload
	payload := Payload{
		BotProfile:  "voice_2024_10",
		MaxDuration: 600,
		DialOut: []DialOutSettings{
			{PhoneNumber: "+13522333930"},
		},
		Services: reqBody.Services,
		APIKeys: APIKeys{
			OpenAI:   os.Getenv("OPENAI_API_KEY"),
			Deepgram: os.Getenv("DEEPGRAM_API_KEY"),
			Cartesia: os.Getenv("CARTESIA_API_KEY"),
		},
		Config: reqBody.Config, // Ensure this is an array
	}

	// Serialize payload to JSON
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error serializing payload: %v", err)
		http.Error(w, "Failed to serialize payload", http.StatusInternalServerError)
		return
	}

	// Make the API request
	apiURL := "https://api.daily.co/v1/bots/start"
	req, err := http.NewRequest(http.MethodPost, apiURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		log.Printf("Error creating request: %v", err)
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", dailyBotsKey))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error contacting Daily API: %v", err)
		http.Error(w, "Failed to contact Daily API", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Read the API response
	apiResponse, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading API response: %v", err)
		http.Error(w, "Failed to read API response", http.StatusInternalServerError)
		return
	}

	// Handle non-200 status codes
	if resp.StatusCode != http.StatusOK {
		log.Printf("Daily API error: %s", string(apiResponse))
		w.WriteHeader(resp.StatusCode)
		w.Write(apiResponse)
		return
	}

	// Return success response
	w.WriteHeader(http.StatusOK)
	w.Write(apiResponse)
}
