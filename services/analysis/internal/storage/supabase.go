package storage

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"github.com/supabase-community/supabase-go"
)

func initSupabaseClient() (*supabase.Client, error) {
	workingDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	envPath := filepath.Join(workingDir, ".env")
	if err := godotenv.Load(envPath); err != nil {
		panic(err)
	}
	fmt.Printf("envPath declared as: %s\n", envPath)
	if err != nil {
		return nil, fmt.Errorf("error loading .env file: %w", err)
	}

	// Get environment variables
	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_KEY")

	// Validate environment variables are present
	if supabaseURL == "" || supabaseKey == "" {
		return nil, fmt.Errorf("SUPABASE_URL and SUPABASE_KEY must be set in .env file")
	}

	// Create client options with default values
	options := supabase.ClientOptions{
		// custom options here if needed
	}

	// Create Supabase client
	client, err := supabase.NewClient(supabaseURL, supabaseKey, &options)
	if err != nil {
		return nil, fmt.Errorf("error creating Supabase client: %w", err)
	}

	return client, nil
}
