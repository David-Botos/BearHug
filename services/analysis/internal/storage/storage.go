package storage

import (
	"encoding/json"
	"fmt"

	"github.com/david-botos/BearHug/services/analysis/internal/types"
)

func StoreCallData(params types.StoreCallDataParams) error {
	// Initialize Supabase client
	client, err := initSupabaseClient()
	if err != nil {
		return fmt.Errorf("failed to initialize Supabase client: %w", err)
	}

	// Create the transcript data
	var results []struct {
		UUID string `json:"uuid"`
	}

	data, _, err := client.From("transcripts").
		Insert(map[string]interface{}{
			"full_transcript": params.Transcript,
		}, false, "", "uuid", "").
		Execute()

	if err != nil {
		// handle error
		return err
	}

	if err := json.Unmarshal(data, &results); err != nil {
		// handle unmarshal error
		return err
	}

	transcriptID := results[0].UUID

	// Create the call data
	callData := map[string]interface{}{
		"fk_organization": params.Organization,
		"room_url":        params.RoomURL,
		"fk_transcript":   transcriptID,
	}

	// Insert into calls table
	_, _, err = client.From("calls").
		Insert(callData, false, "", "", "").
		Execute()

	if err != nil {
		return fmt.Errorf("failed to insert call data: %w", err)
	}

	return nil
}
