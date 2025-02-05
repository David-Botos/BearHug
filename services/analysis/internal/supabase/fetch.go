package supabase

import (
	"encoding/json"
	"fmt"

	"github.com/david-botos/BearHug/services/analysis/internal/hsds_types"
	"github.com/david-botos/BearHug/services/analysis/pkg/logger"
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

func FetchOrgContacts(org_id string) ([]hsds_types.Contact, error) {
	log := logger.Get()
	log.Info().Str("org_id", org_id).Msg("Fetching organization contacts")
	client, initErr := InitSupabaseClient()
	if initErr != nil {
		return nil, fmt.Errorf("failed to initialize Supabase client: %w", initErr)
	}
	var contacts []hsds_types.Contact

	order := &postgrest.OrderOpts{
		Ascending:    true,
		NullsFirst:   false,
		ForeignTable: "",
	}

	data, count, err := client.From("contact").Select(`
		id,
		organization_id,
		service_id,
		service_at_location_id,
		location_id,
		name,
		title,
		department,
		email,
		created_at,
    	updated_at
	`, "", false).Eq("organization_id", org_id).Order("name", order).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch orgContacts from supa: %w", err)
	}

	log.Debug().
		Str("org_id", org_id).
		Int64("count", count).
		Str("raw_data", string(data)).
		Msg("Retrieved contacts data")

	if err := json.Unmarshal(data, &contacts); err != nil {
		return nil, fmt.Errorf("failed to unmarshal orgContacts data: %w", err)
	}

	return contacts, nil
}

func FetchRelevantPhones(org_id string, contactIDs []string, serviceIDs []string) ([]hsds_types.Phone, error) {
	client, initErr := InitSupabaseClient()
	if initErr != nil {
		return nil, fmt.Errorf("failed to initialize Supabase client: %w", initErr)
	}

	// get phones by org id
	orgPhones, _, err := client.From("phone").Select(`
        id,
        location_id,
        service_id,
        organization_id,
        contact_id,
        service_at_location_id,
        number,
        extension,
        type,
        description,
		created_at,
    	updated_at
        `, "", false).Eq("organization_id", org_id).Execute()

	if err != nil {
		return nil, fmt.Errorf("failed to fetch phones: %w", err)
	}

	// get phones by contact IDs
	contactPhones, _, err := client.From("phone").Select(`
        id,
        location_id,
        service_id,
        organization_id,
        contact_id,
        service_at_location_id,
        number,
        extension,
        type,
        description,
		created_at,
    	updated_at
        `, "", false).In("contact_id", contactIDs).Execute()

	if err != nil {
		return nil, fmt.Errorf("failed to fetch phones by contact IDs: %w", err)
	}

	// get phones by service IDs
	servicePhones, _, err := client.From("phone").Select(`
        id,
        location_id,
        service_id,
        organization_id,
        contact_id,
        service_at_location_id,
        number,
        extension,
        type,
        description,
		created_at,
    	updated_at
        `, "", false).In("service_id", serviceIDs).Execute()

	if err != nil {
		return nil, fmt.Errorf("failed to fetch phones by service IDs: %w", err)
	}

	// Create a map to store unique phone entries
	uniquePhones := make(map[string]hsds_types.Phone)

	// Helper function to unmarshal and add phones to the map
	addToUniquePhones := func(data []byte) error {
		var phones []hsds_types.Phone
		if err := json.Unmarshal(data, &phones); err != nil {
			return fmt.Errorf("failed to unmarshal phones: %w", err)
		}
		for _, phone := range phones {
			uniquePhones[phone.ID] = phone
		}
		return nil
	}

	// Add all phone results to the map
	if err := addToUniquePhones(orgPhones); err != nil {
		return nil, err
	}
	if err := addToUniquePhones(contactPhones); err != nil {
		return nil, err
	}
	if err := addToUniquePhones(servicePhones); err != nil {
		return nil, err
	}

	// Convert map values back to slice
	result := make([]hsds_types.Phone, 0, len(uniquePhones))
	for _, phone := range uniquePhones {
		result = append(result, phone)
	}

	return result, nil
}
