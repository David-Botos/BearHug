package supabase

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"github.com/supabase-community/supabase-go"
)

// InitSupabaseClient initializes the Supabase client
func InitSupabaseClient() (*supabase.Client, error) {
	if os.Getenv("ENVIRONMENT") == "development" {
		workingDir, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("error getting working directory: %w", err)
		}
		if err := godotenv.Load(filepath.Join(workingDir, ".env")); err != nil {
			return nil, fmt.Errorf("error loading .env file: %w", err)
		}
	}

	// Get environment variables
	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_KEY")

	// Validate environment variables are present
	if supabaseURL == "" || supabaseKey == "" {
		return nil, fmt.Errorf("SUPABASE_URL and SUPABASE_KEY must be set in environment")
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
