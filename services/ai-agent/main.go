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
	key := config.LoadConfig().ElevenLabsAPIKey

	// Instantiate an aiAssistant
	aiAssistant := assistant.New()

	// Instantiate ElevenLabs TTS service
	elevenLabs := elevenlabs.New(key)

	// Start a conversation thread
	thread := aiAssistant.CreateThread("You are a technical support AI", "Hello, I need help with my PC.")

	// LLM response thread
	response, err := aiAssistant.ResponseThread(context.Background(), thread)

	// Process AI response through EL service
	elevenLabs.StreamAudio(response)

	// Handle errors accordingly
	if err != nil {
		log.Fatalf("Error generating AI response: %v", err)
	}

}
