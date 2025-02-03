package supabase

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/david-botos/BearHug/services/analysis/internal/hsds_types"
	"github.com/supabase-community/postgrest-go"
)

// ServiceStatusEnum represents the status of a service
type ServiceStatusEnum string

// FetchOrganizationName retrieves an organization's name by its ID
func FetchOrganizationName(organizationID string) (string, error) {
	client, initErr := InitSupabaseClient()
	if initErr != nil {
		return "", fmt.Errorf("failed to initialize Supabase client: %w", initErr)
	}

	type Organization struct {
		Name string `json:"name"`
	}

	var org Organization
	data, _, err := client.From("organization").
		Select("name", "", false).
		Eq("id", organizationID).
		Single().
		Execute()

	if err != nil {
		return "", fmt.Errorf("failed to fetch organization: %w", err)
	}

	if err := json.Unmarshal(data, &org); err != nil {
		return "", fmt.Errorf("failed to unmarshal organization data: %w", err)
	}

	return org.Name, nil
}

// FetchOrganizationServices retrieves all services associated with an organization
func FetchOrganizationServices(organizationID string) ([]hsds_types.Service, error) {
	fmt.Printf(`Fetching organization services for org ID: %s`, organizationID)

	client, initErr := InitSupabaseClient()
	if initErr != nil {
		return nil, fmt.Errorf("failed to initialize Supabase client: %w", initErr)
	}

	var services []hsds_types.Service
	order := &postgrest.OrderOpts{
		Ascending:    true,
		NullsFirst:   false,
		ForeignTable: "",
	}

	data, _, err := client.From("service").
		Select(`
            id,
            organization_id,
            program_id,
            name,
            alternate_name,
            description,
            url,
            email,
            status,
            interpretation_services,
            application_process,
            fees_description,
            wait_time,
            fees,
            accreditations,
            eligibility_description,
            minimum_age,
            maximum_age,
            assured_date,
            assurer_email,
            licenses,
            alert,
        	last_modified',
        	created_at',
        	updated_at'
        `, "", false).
		Eq("organization_id", organizationID).
		Order("name", order).
		Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch services: %w", err)
	}

	fmt.Printf("Raw data from Supabase: %s\n", string(data))

	// Pre-process the JSON to handle timestamp format
	var rawData []map[string]interface{}
	if err := json.Unmarshal(data, &rawData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal raw data: %w", err)
	}

	// Process timestamps
	for i := range rawData {
		if lm, ok := rawData[i]["last_modified"].(string); ok && lm != "" {
			// Parse the timestamp and add UTC timezone
			t, err := time.Parse("2006-01-02T15:04:05.999999", lm)
			if err != nil {
				return nil, fmt.Errorf("failed to parse timestamp: %w", err)
			}
			rawData[i]["last_modified"] = t.UTC().Format(time.RFC3339)
		}
	}
	// Marshal back to JSON
	processedData, err := json.Marshal(rawData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal processed data: %w", err)
	}

	// Unmarshal into services slice
	if err := json.Unmarshal(processedData, &services); err != nil {
		return nil, fmt.Errorf("failed to unmarshal services data: %w", err)
	}

	return services, nil
}

func FetchUnits() ([]hsds_types.Unit, error) {
	client, initErr := InitSupabaseClient()
	if initErr != nil {
		return nil, fmt.Errorf("failed to initialize Supabase client: %w", initErr)
	}

	var units []hsds_types.Unit

	order := &postgrest.OrderOpts{
		Ascending:    true,
		NullsFirst:   false,
		ForeignTable: "",
	}

	data, _, err := client.From("unit").
		Select(`
			id,
			name,
			scheme,
			identifier,
			uri
		`, "", false).
		Order("name", order).
		Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch units from supa: %w", err)
	}

	// Unmarshal the byte array into the units slice
	if err := json.Unmarshal(data, &units); err != nil {
		return nil, fmt.Errorf("failed to unmarshal units data: %w", err)
	}

	return units, nil
}
