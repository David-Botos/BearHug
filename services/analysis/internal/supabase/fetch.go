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
	if err := hsds_types.UnmarshalJSONWithTime(data, &services); err != nil {
		return nil, fmt.Errorf("failed to unmarshal services: %w", err)
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
	if err := hsds_types.UnmarshalJSONWithTime(data, &units); err != nil {
		return nil, fmt.Errorf("failed to unmarshal units: %w", err)
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

	if err := hsds_types.UnmarshalJSONWithTime(data, &contacts); err != nil {
		return nil, fmt.Errorf("failed to unmarshal contacts: %w", err)
	}

	return contacts, nil
}

func fetchPhones(field string, value interface{}) ([]byte, error) {
	phoneFields := `
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
    `

	client, initErr := InitSupabaseClient()
	if initErr != nil {
		return nil, fmt.Errorf("error when initializing supa: %w", initErr)
	}

	query := client.From("phone").Select(phoneFields, "", false)

	switch v := value.(type) {
	case string:
		query = query.Eq(field, v)
	case []string:
		query = query.In(field, v)
	default:
		return nil, fmt.Errorf("unsupported value type for field %s", field)
	}

	data, _, err := query.Execute()
	if err != nil {
		return nil, fmt.Errorf("fetching phones by %s: %w", field, err)
	}
	return data, nil
}

func FetchRelevantPhones(org_id string, contactIDs []string, serviceIDs []string) ([]hsds_types.Phone, error) {

	orgPhones, err := fetchPhones("organization_id", org_id)
	if err != nil {
		return nil, err
	}

	contactPhones, err := fetchPhones("contact_id", contactIDs)
	if err != nil {
		return nil, err
	}

	servicePhones, err := fetchPhones("service_id", serviceIDs)
	if err != nil {
		return nil, err
	}

	return hsds_types.UnmarshalMultipleJSONResponses[hsds_types.Phone](
		[][]byte{orgPhones, contactPhones, servicePhones},
	)
}
