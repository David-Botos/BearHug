package supabase

import (
	"fmt"
	"os"

	env "github.com/david-botos/BearHug/services/analysis/pkg/ENV"
	"github.com/david-botos/BearHug/services/analysis/pkg/logger"
	"github.com/supabase-community/supabase-go"
)

// InitSupabaseClient initializes the Supabase client
func InitSupabaseClient() (*supabase.Client, error) {
	log := logger.Get()
	if err := env.LoadEnvFile(); err != nil {
		return nil, err
	}

	log.Debug().Str("environment", os.Getenv("ENVIRONMENT")).Msg("Current environment")

	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_KEY")

	if supabaseURL == "" || supabaseKey == "" {
		return nil, fmt.Errorf("SUPABASE_URL and SUPABASE_KEY must be set in environment")
	}

	options := supabase.ClientOptions{}
	client, err := supabase.NewClient(supabaseURL, supabaseKey, &options)
	if err != nil {
		return nil, fmt.Errorf("error creating Supabase client: %w", err)
	}
	return client, nil
}
