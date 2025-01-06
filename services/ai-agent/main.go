package main

import (
	"context"
	"log"

	"github.com/david-botos/BearHug/services/ai-agent/config"
	"github.com/david-botos/BearHug/services/ai-agent/features/assistant"
	"github.com/david-botos/BearHug/services/ai-agent/features/elevenlabs"
)

func main() {

	// Load env varibles
	config.LoadConfig()

	// Instantiate an aiAssistant
	aiAssistant := assistant.New()

	// Instantiate ElevenLabs TTS service
	elevenLabs := elevenlabs.New(config.LoadConfig().ElevenLabsAPIKey)

	// Start a conversation thread
	thread := aiAssistant.CreateThread("You are a technical support AI", "Hello, I need help with my PC.")

	// LLM response thread
	response, err := aiAssistant.ResponseThread(context.Background(), thread)

	// Process AI response through EL service

	// Handle errors accordingly
	if err != nil {
		log.Fatalf("Error generating AI response: %v", err)
	}

	// Print response from aiAssistant
	log.Println("AI Response: ", response)

}
