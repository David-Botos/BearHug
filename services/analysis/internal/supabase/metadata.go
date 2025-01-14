package supabase

import (
	"fmt"

	"github.com/david-botos/BearHug/services/analysis/internal/hsds_types"
)

// MetadataInput represents the required and optional fields for creating a metadata entry
type MetadataInput struct {
	ResourceID       string
	ResourceType     string
	FieldName        string
	PreviousValue    string
	ReplacementValue string
	LastActionType   string // Optional, defaults to "UPDATE"
}

// CreateAndStoreMetadata creates and stores multiple metadata entries in Supabase
func CreateAndStoreMetadata(inputs []MetadataInput) error {
	client, err := InitSupabaseClient()
	if err != nil {
		return fmt.Errorf("failed to initialize Supabase client: %w", err)
	}

	var metadataRecords []hsds_types.Metadata

	// Create metadata objects for each input
	for _, input := range inputs {
		// Set default values
		actionType := input.LastActionType
		if actionType == "" {
			actionType = "UPDATE"
		}

		previousValue := input.PreviousValue
		if previousValue == "" {
			previousValue = "none"
		}

		fieldName := input.FieldName
		if fieldName == "" {
			fieldName = "new"
		}

		metadata, err := hsds_types.NewMetadata(
			input.ResourceID,
			input.ResourceType,
			actionType,
			input.FieldName,
			previousValue,
			input.ReplacementValue,
			"BearHug", // Hardcoded updater as specified
		)
		if err != nil {
			return fmt.Errorf("failed to create metadata object: %w", err)
		}

		metadataRecords = append(metadataRecords, *metadata)
	}

	// Store all metadata records in a single transaction
	_, _, err = client.From("metadata").Insert(metadataRecords, false, "", "representation", "").Execute()
	if err != nil {
		return fmt.Errorf("failed to store metadata records: %w", err)
	}

	return nil
}
