package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	config "github.com/david-botos/BearHug/services/go-bot/pkg"
	"github.com/david-botos/BearHug/services/go-bot/prompt"
)

// RequestBody represents the incoming request structure
type RequestBody struct {
	CBOName           string                   `json:"cbo_name"`
	ServiceCategories prompt.ServiceCategories `json:"service_categories"`
	PhoneNumber       string                   `json:"phone_number"`
}

// ErrorResponse represents a structured error response
type ErrorResponse struct {
	Error string `json:"error"`
}

func getRequiredEnvVars() (map[string]string, error) {
	required := []string{
		"DAILY_BOTS_KEY",
		"GEMINI_API_KEY",
		"AWS_ASSUME_ROLE_ARN",
		"AWS_BUCKET_NAME",
		"AWS_BUCKET_REGION",
	}

	envVars := make(map[string]string)
	var missingVars []string

	for _, varName := range required {
		if value := os.Getenv(varName); value != "" {
			envVars[varName] = value
		} else {
			missingVars = append(missingVars, varName)
		}
	}

	if len(missingVars) > 0 {
		return nil, fmt.Errorf("missing required environment variables: %v", missingVars)
	}

	return envVars, nil
}

func sendErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	response := ErrorResponse{Error: message}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

func DialOutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendErrorResponse(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	var reqBody RequestBody
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		log.Printf("Error decoding request body: %v", err)
		sendErrorResponse(w, "Failed to parse request body", http.StatusBadRequest)
		return
	}

	// Get and validate all required environment variables
	envVars, err := getRequiredEnvVars()
	if err != nil {
		log.Printf("Environment variable error: %v", err)
		sendErrorResponse(w, fmt.Sprintf("Configuration error: %v", err), http.StatusInternalServerError)
		return
	}

	// Validate phone number
	if reqBody.PhoneNumber == "" {
		sendErrorResponse(w, "Phone number is required", http.StatusBadRequest)
		return
	}

	// Generate a tailored prompt
	generatedPrompt, err := prompt.GenPrompt(reqBody.CBOName, reqBody.ServiceCategories)
	if err != nil {
		log.Printf("Error generating prompt: %v", err)
		sendErrorResponse(w, "Failed to generate prompt", http.StatusInternalServerError)
		return
	}

	// Create recording config
	recordingConfig := config.RecordingConfig{
		AssumeRoleARN: envVars["AWS_ASSUME_ROLE_ARN"],
		BucketName:    envVars["AWS_BUCKET_NAME"],
		BucketRegion:  envVars["AWS_BUCKET_REGION"],
	}

	// Build Daily API request body
	dailyReqBody, err := config.BuildRequestBody(reqBody.PhoneNumber, generatedPrompt, envVars["GEMINI_API_KEY"], recordingConfig)
	if err != nil {
		log.Printf("Error building Daily request body: %v", err)
		sendErrorResponse(w, "Failed to construct Daily API request", http.StatusInternalServerError)
		return
	}

	// Serialize payload to JSON
	dailyPayload, err := json.Marshal(dailyReqBody)
	if err != nil {
		log.Printf("Error serializing payload: %v", err)
		sendErrorResponse(w, "Failed to serialize request payload", http.StatusInternalServerError)
		return
	}

	// Create and configure the Daily API request
	apiURL := "https://api.daily.co/v1/bots/start"
	req, err := http.NewRequest(http.MethodPost, apiURL, bytes.NewBuffer(dailyPayload))
	if err != nil {
		log.Printf("Error creating Daily API request: %v", err)
		sendErrorResponse(w, "Failed to create API request", http.StatusInternalServerError)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", envVars["DAILY_BOTS_KEY"]))

	// Make the API request with timeout
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error making Daily API request: %v", err)
		sendErrorResponse(w, "Failed to contact Daily API", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Read and process the API response
	apiResponse, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading Daily API response: %v", err)
		sendErrorResponse(w, "Failed to read API response", http.StatusInternalServerError)
		return
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")

	// Handle non-200 status codes
	if resp.StatusCode != http.StatusOK {
		log.Printf("Daily API error (Status %d): %s", resp.StatusCode, string(apiResponse))
		w.WriteHeader(resp.StatusCode)
		w.Write(apiResponse)
		return
	}

	// Return success response
	w.WriteHeader(http.StatusOK)
	w.Write(apiResponse)
}
