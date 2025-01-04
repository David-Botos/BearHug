package supabase

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/david-botos/BearHug/services/analysis/internal/types"
)

func StoreCallData(params types.StoreCallDataParams) error {
	client, err := initSupabaseClient()
	if err != nil {
		return fmt.Errorf("failed to initialize Supabase client: %w", err)
	}

	// Create the transcript data
	var results []struct {
		ID string `json:"id"`
	}

	data, _, err := client.From("transcripts").
		Insert(map[string]interface{}{
			"full_transcript": params.Transcript,
		}, false, "", "representation", ""). // Changed returning to "representation"
		Execute()

	if err != nil {
		return fmt.Errorf("failed to execute Supabase query: %w, data: %s", err, string(data))
	}

	log.Printf("DEBUG: Raw response data: %s", string(data))

	if err := json.Unmarshal(data, &results); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w, data: %s", err, string(data))
	}

	if len(results) == 0 {
		return fmt.Errorf("no results returned from insert")
	}

	transcriptID := results[0].ID

	// Debug log
	log.Printf("DEBUG: Transcript ID: %s", transcriptID)

	// Create the call data
	callData := map[string]interface{}{
		"fk_organization": params.Organization,
		"room_url":        params.RoomURL,
		"fk_transcript":   transcriptID,
	}

	// Insert into calls table with explicit returning
	data, _, err = client.From("calls").
		Insert(callData, false, "", "representation", "").
		Execute()

	if err != nil {
		return fmt.Errorf("failed to insert call data: %w", err)
	}

	return nil
}
