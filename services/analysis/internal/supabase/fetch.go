package supabase

import (
	"encoding/json"
	"fmt"

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
            last_modified,
            created_at,
            updated_at
        `, "", false).
		Eq("organization_id", organizationID).
		Order("name", order).
		Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch services: %w", err)
	}

	// Unmarshal the byte array into the services slice
	if err := json.Unmarshal(data, &services); err != nil {
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
