package supabase

import (
	"context"
	"fmt"

	"github.com/david-botos/BearHug/services/analysis/internal/hsds_types"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// MetadataInput represents the required and optional fields for creating a metadata entry
type MetadataInput struct {
	ResourceID       string
	CallID           string
	ResourceType     string
	FieldName        string
	PreviousValue    string
	ReplacementValue string
	LastActionType   string // Optional, defaults to "UPDATE"
}

// CreateAndStoreMetadata creates and stores multiple metadata entries in Supabase
func CreateAndStoreMetadata(ctx context.Context, inputs []MetadataInput) error {
	tracer := otel.GetTracerProvider().Tracer("metadata-operations")
	ctx, span := tracer.Start(ctx, "create_and_store_metadata",
		trace.WithAttributes(
			attribute.Int("input_count", len(inputs)),
		),
	)
	defer span.End()

	// Initialize Supabase client with tracing
	ctx, clientSpan := tracer.Start(ctx, "init_supabase_client")
	client, err := InitSupabaseClient()
	if err != nil {
		clientSpan.RecordError(err)
		clientSpan.End()
		return fmt.Errorf("failed to initialize Supabase client: %w", err)
	}
	clientSpan.End()

	// Create metadata objects for each input
	ctx, createSpan := tracer.Start(ctx, "create_metadata_objects")
	var metadataRecords []hsds_types.Metadata

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
			input.CallID,
			input.ResourceType,
			actionType,
			fieldName,
			previousValue,
			input.ReplacementValue,
			"BearHug", // Hardcoded updater as specified
		)
		if err != nil {
			createSpan.RecordError(err)
			createSpan.End()
			return fmt.Errorf("failed to create metadata object: %w", err)
		}

		metadataRecords = append(metadataRecords, *metadata)
	}

	createSpan.SetAttributes(
		attribute.Int("records_created", len(metadataRecords)),
	)
	createSpan.End()

	// Store all metadata records in a single transaction
	ctx, storeSpan := tracer.Start(ctx, "store_metadata_batch",
		trace.WithAttributes(
			attribute.Int("batch_size", len(metadataRecords)),
			attribute.String("table", "metadata"),
		),
	)

	_, _, err = client.From("metadata").Insert(metadataRecords, false, "", "representation", "").Execute()
	if err != nil {
		storeSpan.RecordError(err)
		storeSpan.End()
		return fmt.Errorf("failed to store metadata records: %w", err)
	}
	storeSpan.End()

	span.SetAttributes(
		attribute.Bool("success", true),
		attribute.Int("total_records_stored", len(metadataRecords)),
	)

	return nil
}
