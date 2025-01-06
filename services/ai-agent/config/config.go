package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	OpenAIAPIKey     string
	ElevenLabsAPIKey string
}

func LoadConfig() Config {
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}

	return Config{
		OpenAIAPIKey:     getEnv("OPENAI_API_KEY"),
		ElevenLabsAPIKey: getEnv("ELEVENLABS_API_KEY"),
	}
}

func getEnv(key string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		log.Fatalf("Environment variable %s not set", key)
	}
	return value
}
