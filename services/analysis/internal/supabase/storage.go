package supabase

import (
	"encoding/json"
	"fmt"

	"github.com/david-botos/BearHug/services/analysis/internal/hsds_types"
	"github.com/david-botos/BearHug/services/analysis/internal/types"
	"github.com/david-botos/BearHug/services/analysis/pkg/logger"
)

// StoreCallData stores transcript and call data in Supabase and returns the call ID
// It creates two records: one in the transcripts table and one in the calls table
func StoreCallData(params types.TranscriptsReqBody) (string, error) {
	log := logger.Get() // Get instance of custom logger

	// Log the incoming request with structured fields
	log.Info().
		Str("organization_id", params.OrganizationID).
		Str("room_url", params.RoomURL).
		Str("transcript", params.Transcript).
		Msg("Storing transcript data")

	client, err := InitSupabaseClient()
	if err != nil {
		return "", fmt.Errorf("failed to initialize Supabase client: %w", err)
	}

	// Create the transcript data
	var results []struct {
		ID string `json:"id"`
	}

	data, _, err := client.From("transcripts").
		Insert(map[string]interface{}{
			"full_transcript": params.Transcript,
		}, false, "", "representation", "").
		Execute()

	if err != nil {
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
		log.Error().
			Err(err).
			Str("data", string(data)).
			Msg("Failed to unmarshal response")
		return "", fmt.Errorf("failed to unmarshal response: %w, data: %s", err, string(data))
	}

	if len(results) == 0 {
		log.Error().Msg("No results returned from insert")
		return "", fmt.Errorf("no results returned from insert")
	}

	transcriptID := results[0].ID

	log.Debug().
		Str("transcript_id", transcriptID).
		Msg("Successfully created transcript record")

	// Create the call data
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
		log.Error().
			Err(err).
			Interface("call_data", callData).
			Msg("Failed to insert call data")
		return "", fmt.Errorf("failed to insert call data: %w", err)
	}

	if err := json.Unmarshal(data, &results); err != nil {
		log.Error().
			Err(err).
			Str("data", string(data)).
			Msg("Failed to unmarshal call response")
		return "", fmt.Errorf("failed to unmarshal response: %w, data: %s", err, string(data))
	}

	callID := results[0].ID

	log.Debug().
		Str("call_id", callID).
		Msg("Successfully created call record")

	return callID, nil
}

// StoreNewServices stores multiple service records in Supabase and creates corresponding metadata
func StoreNewServices(services []*hsds_types.Service, callID string) error {
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
		if err := CreateAndStoreMetadata(metadataInputs); err != nil {
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
func StoreNewCapacity(capacityObjects []*hsds_types.ServiceCapacity, callID string) error {
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

	for _, capObj := range capacityObjects {
		capacityData := map[string]interface{}{
			"id":         capObj.ID,
			"service_id": capObj.ServiceID,
			"unit_id":    capObj.UnitID,
			"available":  capObj.Available,
			"updated":    capObj.Updated,
		}

		// Add optional fields only if they're not nil
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
			log.Error().
				Err(err).
				Str("capacity_id", capObj.ID).
				Interface("capacity_data", capacityData).
				Msg("Failed to insert capacity data")
			return fmt.Errorf("failed to insert capacity data: %w, data: %s", err, string(data))
		}

		log.Debug().
			Str("capacity_id", capObj.ID).
			Msg("Successfully created capacity record")

		metadataInputs = append(metadataInputs, MetadataInput{
			ResourceID:       capObj.ID,
			CallID:           callID,
			ResourceType:     "service_capacity",
			ReplacementValue: "new entry",
			LastActionType:   "CREATE",
		})
	}

	// Create metadata for all the new capacity data
	if len(metadataInputs) > 0 {
		if err := CreateAndStoreMetadata(metadataInputs); err != nil {
			log.Error().
				Err(err).
				Int("metadata_count", len(metadataInputs)).
				Msg("Failed to create metadata for capacity objects")
			return fmt.Errorf("failed to create metadata for capacity objs: %w", err)
		}

		log.Info().
			Int("metadata_count", len(metadataInputs)).
			Msg("Successfully created metadata for capacity objects")
	}

	return nil
}

// StoreNewUnits stores multiple unit records in Supabase and creates corresponding metadata
func StoreNewUnits(unitObjects []*hsds_types.Unit, callID string) error {
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
			log.Error().
				Err(err).
				Str("unit_id", unitObj.ID).
				Interface("unit_data", unitsData).
				Msg("Failed to insert unit data")
			return fmt.Errorf("failed to insert unit data: %w, data: %s", err, string(data))
		}

		log.Debug().
			Str("unit_id", unitObj.ID).
			Msg("Successfully created unit record")

		metadataInputs = append(metadataInputs, MetadataInput{
			ResourceID:       unitObj.ID,
			CallID:           callID,
			ResourceType:     "unit",
			ReplacementValue: "new entry",
			LastActionType:   "CREATE",
		})
	}

	// Create metadata for all the new unit data
	if len(metadataInputs) > 0 {
		if err := CreateAndStoreMetadata(metadataInputs); err != nil {
			log.Error().
				Err(err).
				Int("metadata_count", len(metadataInputs)).
				Msg("Failed to create metadata for unit objects")
			return fmt.Errorf("failed to create metadata for unit objs: %w", err)
		}

		log.Info().
			Int("metadata_count", len(metadataInputs)).
			Msg("Successfully created metadata for unit objects")
	}

	return nil
}

func StoreNewContacts(contactObjects []*hsds_types.Contact, callID string) error {
	log := logger.Get()

	client, err := InitSupabaseClient()
	if err != nil {
		log.Error().
			Err(err).
			Msg("Failed to initialize Supabase client")
		return fmt.Errorf("failed to initialize Supabase client: %w", err)
	}

	var metadataInputs []MetadataInput

	for _, contactObj := range contactObjects {
		contactData := map[string]interface{}{
			"id": contactObj.ID,
		}

		// Add optional fields only if they're not nil
		if contactObj.OrganizationID != nil {
			contactData["organization_id"] = *contactObj.OrganizationID
		}
		if contactObj.ServiceID != nil {
			contactData["service_id"] = *contactObj.ServiceID
		}
		if contactObj.ServiceAtLocationID != nil {
			contactData["service_at_location_id"] = *contactObj.ServiceAtLocationID
		}
		if contactObj.LocationID != nil {
			contactData["location_id"] = *contactObj.LocationID
		}
		if contactObj.Name != nil {
			contactData["name"] = *contactObj.Name
		}
		if contactObj.Title != nil {
			contactData["title"] = *contactObj.Title
		}
		if contactObj.Department != nil {
			contactData["department"] = *contactObj.Department
		}
		if contactObj.Email != nil {
			contactData["email"] = *contactObj.Email
		}

		data, _, err := client.From("contact").
			Insert(contactData, false, "", "representation", "").
			Execute()
		if err != nil {
			log.Error().
				Err(err).
				Str("contact_id", contactObj.ID).
				Interface("contact_data", contactData).
				Msg("Failed to insert contact data")
			return fmt.Errorf("failed to insert contact data: %w, data: %s", err, string(data))
		}

		log.Debug().
			Str("contact_id", contactObj.ID).
			Msg("Successfully created contact record")

		metadataInputs = append(metadataInputs, MetadataInput{
			ResourceID:       contactObj.ID,
			CallID:           callID,
			ResourceType:     "contact",
			ReplacementValue: "new entry",
			LastActionType:   "CREATE",
		})
	}

	if len(metadataInputs) > 0 {
		if err := CreateAndStoreMetadata(metadataInputs); err != nil {
			log.Error().
				Err(err).
				Int("metadata_count", len(metadataInputs)).
				Msg("Failed to create metadata for contact objects")
			return fmt.Errorf("failed to create metadata for contact objs: %w", err)
		}

		log.Info().
			Int("metadata_count", len(metadataInputs)).
			Msg("Successfully created metadata for contact objects")
	}

	return nil
}

func StoreNewPhones(phoneObjects []*hsds_types.Phone, callID string) error {
	log := logger.Get()

	client, err := InitSupabaseClient()
	if err != nil {
		log.Error().
			Err(err).
			Msg("Failed to initialize Supabase client")
		return fmt.Errorf("failed to initialize Supabase client: %w", err)
	}

	var metadataInputs []MetadataInput

	for _, phoneObj := range phoneObjects {
		phoneData := map[string]interface{}{
			"id":     phoneObj.ID,
			"number": phoneObj.Number, // Required field
		}

		// Add optional fields only if they're not nil
		if phoneObj.LocationID != nil {
			phoneData["location_id"] = *phoneObj.LocationID
		}
		if phoneObj.ServiceID != nil {
			phoneData["service_id"] = *phoneObj.ServiceID
		}
		if phoneObj.OrganizationID != nil {
			phoneData["organization_id"] = *phoneObj.OrganizationID
		}
		if phoneObj.ContactID != nil {
			phoneData["contact_id"] = *phoneObj.ContactID
		}
		if phoneObj.ServiceAtLocationID != nil {
			phoneData["service_at_location_id"] = *phoneObj.ServiceAtLocationID
		}
		if phoneObj.Extension != nil {
			phoneData["extension"] = *phoneObj.Extension
		}
		if phoneObj.Type != nil {
			phoneData["type"] = *phoneObj.Type
		}
		if phoneObj.Description != nil {
			phoneData["description"] = *phoneObj.Description
		}

		data, _, err := client.From("phone").
			Insert(phoneData, false, "", "representation", "").
			Execute()
		if err != nil {
			log.Error().
				Err(err).
				Str("phone_id", phoneObj.ID).
				Interface("phone_data", phoneData).
				Msg("Failed to insert phone data")
			return fmt.Errorf("failed to insert phone data: %w, data: %s", err, string(data))
		}

		log.Debug().
			Str("phone_id", phoneObj.ID).
			Msg("Successfully created phone record")

		metadataInputs = append(metadataInputs, MetadataInput{
			ResourceID:       phoneObj.ID,
			CallID:           callID,
			ResourceType:     "phone",
			ReplacementValue: "new entry",
			LastActionType:   "CREATE",
		})
	}

	if len(metadataInputs) > 0 {
		if err := CreateAndStoreMetadata(metadataInputs); err != nil {
			log.Error().
				Err(err).
				Int("metadata_count", len(metadataInputs)).
				Msg("Failed to create metadata for phone objects")
			return fmt.Errorf("failed to create metadata for phone objs: %w", err)
		}

		log.Info().
			Int("metadata_count", len(metadataInputs)).
			Msg("Successfully created metadata for phone objects")
	}

	return nil
}
