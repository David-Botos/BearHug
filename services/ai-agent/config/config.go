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
		OpenAIAPIKey:     GetEnv("OPENAI_API_KEY"),
		ElevenLabsAPIKey: GetEnv("ELEVENLABS_API_KEY"),
	}
}

func GetEnv(key string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		log.Fatalf("Environment variable %s not set", key)
	}
	return value
}
