package supabase

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/david-botos/BearHug/services/analysis/internal/hsds_types"
	"github.com/david-botos/BearHug/services/analysis/internal/types"
	"github.com/david-botos/BearHug/services/analysis/pkg/logger"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// StoreCallData stores transcript and call data in Supabase and returns the call ID
// It creates two records: one in the transcripts table and one in the calls table
func StoreCallData(ctx context.Context, params types.TranscriptsReqBody) (string, error) {
	tracer := otel.GetTracerProvider().Tracer("supabase-client")
	ctx, span := tracer.Start(ctx, "store_call_data",
		trace.WithAttributes(
			attribute.String("organization_id", params.OrganizationID),
			attribute.String("room_url", params.RoomURL),
			attribute.Int("transcript_length", len(params.Transcript)),
		),
	)
	defer span.End()

	log := logger.Get()
	log.Info().
		Str("organization_id", params.OrganizationID).
		Str("room_url", params.RoomURL).
		Str("transcript", params.Transcript).
		Msg("Storing transcript data")

	// Initialize Supabase client
	ctx, clientSpan := tracer.Start(ctx, "init_supabase_client")
	client, err := InitSupabaseClient()
	if err != nil {
		clientSpan.RecordError(err)
		clientSpan.End()
		return "", fmt.Errorf("failed to initialize Supabase client: %w", err)
	}
	clientSpan.End()

	// Create the transcript data
	var results []struct {
		ID string `json:"id"`
	}

	// Store transcript with tracing
	ctx, transcriptSpan := tracer.Start(ctx, "store_transcript")
	data, _, err := client.From("transcripts").
		Insert(map[string]interface{}{
			"full_transcript": params.Transcript,
		}, false, "", "representation", "").
		Execute()

	if err != nil {
		transcriptSpan.RecordError(err)
		transcriptSpan.End()
		log.Error().
			Err(err).
			Str("data", string(data)).
			Msg("Failed to execute Supabase query")
		return "", fmt.Errorf("failed to execute Supabase query: %w, data: %s", err, string(data))
	}

	log.Debug().
		RawJSON("response_data", data).
		Msg("Received response from Supabase")

	if err := json.Unmarshal(data, &results); err != nil {
		transcriptSpan.RecordError(err)
		transcriptSpan.End()
		log.Error().
			Err(err).
			Str("data", string(data)).
			Msg("Failed to unmarshal response")
		return "", fmt.Errorf("failed to unmarshal response: %w, data: %s", err, string(data))
	}

	if len(results) == 0 {
		err := fmt.Errorf("no results returned from insert")
		transcriptSpan.RecordError(err)
		transcriptSpan.End()
		log.Error().Msg("No results returned from insert")
		return "", err
	}

	transcriptID := results[0].ID
	transcriptSpan.SetAttributes(attribute.String("transcript_id", transcriptID))
	transcriptSpan.End()

	log.Debug().
		Str("transcript_id", transcriptID).
		Msg("Successfully created transcript record")

	// Store call data with tracing
	ctx, callSpan := tracer.Start(ctx, "store_call")
	callData := map[string]interface{}{
		"fk_organization": params.OrganizationID,
		"room_url":        params.RoomURL,
		"fk_transcript":   transcriptID,
	}

	// Insert into calls table
	data, _, err = client.From("calls").
		Insert(callData, false, "", "representation", "").
		Execute()

	if err != nil {
		callSpan.RecordError(err)
		callSpan.End()
		log.Error().
			Err(err).
			Interface("call_data", callData).
			Msg("Failed to insert call data")
		return "", fmt.Errorf("failed to insert call data: %w", err)
	}

	if err := json.Unmarshal(data, &results); err != nil {
		callSpan.RecordError(err)
		callSpan.End()
		log.Error().
			Err(err).
			Str("data", string(data)).
			Msg("Failed to unmarshal call response")
		return "", fmt.Errorf("failed to unmarshal response: %w, data: %s", err, string(data))
	}

	callID := results[0].ID
	callSpan.SetAttributes(attribute.String("call_id", callID))
	callSpan.End()

	log.Debug().
		Str("call_id", callID).
		Msg("Successfully created call record")

	span.SetAttributes(
		attribute.Bool("success", true),
		attribute.String("transcript_id", transcriptID),
		attribute.String("call_id", callID),
	)

	return callID, nil
}

// StoreNewServices stores multiple service records in Supabase and creates corresponding metadata
func StoreNewServices(ctx context.Context, services []*hsds_types.Service, callID string) error {
	log := logger.Get()

	client, err := InitSupabaseClient()
	if err != nil {
		log.Error().
			Err(err).
			Msg("Failed to initialize Supabase client")
		return fmt.Errorf("failed to initialize Supabase client: %w", err)
	}

	// Create a slice to collect metadata entries
	var metadataInputs []MetadataInput

	for _, service := range services {
		serviceData := map[string]interface{}{
			"id":                      service.ID,
			"organization_id":         service.OrganizationID,
			"name":                    service.Name,
			"status":                  service.Status,
			"program_id":              service.ProgramID,
			"alternate_name":          service.AlternateName,
			"description":             service.Description,
			"url":                     service.URL,
			"email":                   service.Email,
			"interpretation_services": service.InterpretationServices,
			"application_process":     service.ApplicationProcess,
			"fees_description":        service.FeesDescription,
			"eligibility_description": service.EligibilityDescription,
			"minimum_age":             service.MinimumAge,
			"maximum_age":             service.MaximumAge,
			"alert":                   service.Alert,
			"wait_time":               service.WaitTime,
			"fees":                    service.Fees,
			"licenses":                service.Licenses,
			"accreditations":          service.Accreditations,
			"assured_date":            service.AssuredDate,
			"assurer_email":           service.AssurerEmail,
			"last_modified":           service.LastModified,
		}

		data, _, err := client.From("service").
			Insert(serviceData, false, "", "representation", "").
			Execute()
		if err != nil {
			log.Error().
				Err(err).
				Str("service_id", service.ID).
				Interface("service_data", serviceData).
				Msg("Failed to insert service data")
			return fmt.Errorf("failed to insert service data: %w, data: %s", err, string(data))
		}

		log.Debug().
			Str("service_id", service.ID).
			Msg("Successfully created service record")

		metadataInputs = append(metadataInputs, MetadataInput{
			ResourceID:       service.ID,
			ResourceType:     "service",
			ReplacementValue: "new entry",
			LastActionType:   "CREATE",
			CallID:           callID,
		})
	}

	// Create metadata for all the new services
	if len(metadataInputs) > 0 {
		if err := CreateAndStoreMetadata(ctx, metadataInputs); err != nil {
			log.Error().
				Err(err).
				Int("metadata_count", len(metadataInputs)).
				Msg("Failed to create metadata for services")
			return fmt.Errorf("failed to create metadata for services: %w", err)
		}

		log.Info().
			Int("metadata_count", len(metadataInputs)).
			Msg("Successfully created metadata for services")
	}

	return nil
}

// StoreNewCapacity stores multiple service capacity records in Supabase and creates corresponding metadata
func StoreNewCapacity(ctx context.Context, capacityObjects []*hsds_types.ServiceCapacity, callID string) error {
	tracer := otel.GetTracerProvider().Tracer("supabase-operations")
	ctx, span := tracer.Start(ctx, "store_new_capacity",
		trace.WithAttributes(
			attribute.String("call_id", callID),
			attribute.Int("capacity_count", len(capacityObjects)),
		),
	)
	defer span.End()

	log := logger.Get()

	client, err := InitSupabaseClient()
	if err != nil {
		span.RecordError(err)
		log.Error().
			Err(err).
			Msg("Failed to initialize Supabase client")
		return fmt.Errorf("failed to initialize Supabase client: %w", err)
	}

	// Create a slice to collect metadata entries
	var metadataInputs []MetadataInput
	successCount := 0

	// Store all capacity records
	ctx, storageSpan := tracer.Start(ctx, "store_capacity_batch")
	for _, capObj := range capacityObjects {
		capacityData := map[string]interface{}{
			"id":         capObj.ID,
			"service_id": capObj.ServiceID,
			"unit_id":    capObj.UnitID,
			"available":  capObj.Available,
			"updated":    capObj.Updated,
		}

		if capObj.Maximum != nil {
			capacityData["maximum"] = *capObj.Maximum
		}
		if capObj.Description != nil {
			capacityData["description"] = *capObj.Description
		}

		data, _, err := client.From("service_capacity").
			Insert(capacityData, false, "", "representation", "").
			Execute()
		if err != nil {
			storageSpan.RecordError(err)
			storageSpan.End()
			log.Error().
				Err(err).
				Str("capacity_id", capObj.ID).
				Msg("Failed to insert capacity data")
			return fmt.Errorf("failed to insert capacity data: %w, data: %s", err, string(data))
		}

		metadataInputs = append(metadataInputs, MetadataInput{
			ResourceID:       capObj.ID,
			CallID:           callID,
			ResourceType:     "service_capacity",
			ReplacementValue: "new entry",
			LastActionType:   "CREATE",
		})
		successCount++
	}

	storageSpan.SetAttributes(
		attribute.Int("records_stored", successCount),
	)
	storageSpan.End()

	// Create metadata if needed
	if len(metadataInputs) > 0 {
		ctx, metadataSpan := tracer.Start(ctx, "store_metadata")
		if err := CreateAndStoreMetadata(ctx, metadataInputs); err != nil {
			metadataSpan.RecordError(err)
			metadataSpan.End()
			log.Error().
				Err(err).
				Int("metadata_count", len(metadataInputs)).
				Msg("Failed to create metadata for capacity objects")
			return fmt.Errorf("failed to create metadata for capacity objects: %w", err)
		}
		metadataSpan.SetAttributes(attribute.Int("metadata_count", len(metadataInputs)))
		metadataSpan.End()
	}

	span.SetAttributes(
		attribute.Bool("success", true),
		attribute.Int("capacities_stored", successCount),
		attribute.Int("metadata_created", len(metadataInputs)),
	)

	return nil
}

// StoreNewUnits stores multiple unit records in Supabase and creates corresponding metadata
func StoreNewUnits(ctx context.Context, unitObjects []*hsds_types.Unit, callID string) error {
	tracer := otel.GetTracerProvider().Tracer("supabase-operations")
	ctx, span := tracer.Start(ctx, "store_new_units",
		trace.WithAttributes(
			attribute.String("call_id", callID),
			attribute.Int("unit_count", len(unitObjects)),
		),
	)
	defer span.End()

	log := logger.Get()

	client, err := InitSupabaseClient()
	if err != nil {
		span.RecordError(err)
		log.Error().
			Err(err).
			Msg("Failed to initialize Supabase client")
		return fmt.Errorf("failed to initialize Supabase client: %w", err)
	}

	// Create a slice to collect metadata entries
	var metadataInputs []MetadataInput
	successCount := 0

	// Store all unit records
	ctx, storageSpan := tracer.Start(ctx, "store_units_batch")
	for _, unitObj := range unitObjects {
		unitsData := map[string]interface{}{
			"id":   unitObj.ID,
			"name": unitObj.Name,
		}

		// Add optional fields only if they're not nil
		if unitObj.Scheme != nil {
			unitsData["scheme"] = *unitObj.Scheme
		}
		if unitObj.Identifier != nil {
			unitsData["identifier"] = *unitObj.Identifier
		}
		if unitObj.URI != nil {
			unitsData["uri"] = *unitObj.URI
		}

		data, _, err := client.From("unit").
			Insert(unitsData, false, "", "representation", "").
			Execute()
		if err != nil {
			storageSpan.RecordError(err)
			storageSpan.End()
			log.Error().
				Err(err).
				Str("unit_id", unitObj.ID).
				Msg("Failed to insert unit data")
			return fmt.Errorf("failed to insert unit data: %w, data: %s", err, string(data))
		}

		metadataInputs = append(metadataInputs, MetadataInput{
			ResourceID:       unitObj.ID,
			CallID:           callID,
			ResourceType:     "unit",
			ReplacementValue: "new entry",
			LastActionType:   "CREATE",
		})
		successCount++
	}

	storageSpan.SetAttributes(
		attribute.Int("records_stored", successCount),
	)
	storageSpan.End()

	// Create metadata if needed
	if len(metadataInputs) > 0 {
		ctx, metadataSpan := tracer.Start(ctx, "store_metadata")
		if err := CreateAndStoreMetadata(ctx, metadataInputs); err != nil {
			metadataSpan.RecordError(err)
			metadataSpan.End()
			log.Error().
				Err(err).
				Int("metadata_count", len(metadataInputs)).
				Msg("Failed to create metadata for unit objects")
			return fmt.Errorf("failed to create metadata for unit objects: %w", err)
		}
		metadataSpan.SetAttributes(attribute.Int("metadata_count", len(metadataInputs)))
		metadataSpan.End()
	}

	span.SetAttributes(
		attribute.Bool("success", true),
		attribute.Int("units_stored", successCount),
		attribute.Int("metadata_created", len(metadataInputs)),
	)

	return nil
}
