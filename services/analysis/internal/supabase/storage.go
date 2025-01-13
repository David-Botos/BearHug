package supabase

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/david-botos/BearHug/services/analysis/internal/hsds_types"
	"github.com/david-botos/BearHug/services/analysis/internal/types"
)

func StoreCallData(params types.TranscriptsReqBody) error {
	client, err := InitSupabaseClient()
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
		}, false, "", "representation", "").
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

	log.Printf("DEBUG: Transcript ID: %s", transcriptID)

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
		return fmt.Errorf("failed to insert call data: %w", err)
	}

	return nil
}

func StoreNewServices(services []*hsds_types.Service) error {
	client, err := InitSupabaseClient()
	if err != nil {
		return fmt.Errorf("failed to initialize Supabase client: %w", err)
	}

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
			return fmt.Errorf("failed to insert service data: %w, data: %s", err, string(data))
		}
	}

	return nil
}

func StoreNewCapacity(capacityObjects []*hsds_types.ServiceCapacity) error {
	client, err := InitSupabaseClient()
	if err != nil {
		return fmt.Errorf("failed to initialize Supabase client: %w", err)
	}

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
			return fmt.Errorf("failed to insert capacity data: %w, data: %s", err, string(data))
		}
	}
	return nil
}

func StoreNewUnits(unitObjects []*hsds_types.Unit) error {
	client, err := InitSupabaseClient()
	if err != nil {
		return fmt.Errorf("failed to initialize Supabase client: %w", err)
	}

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
			return fmt.Errorf("failed to insert unit data: %w, data: %s", err, string(data))
		}
	}
	return nil
}
