// TODO: make sure this is sophisticated enough
package structOutputs

import (
	"fmt"
	"math"
	"reflect"
	"strings"
	"time"

	"github.com/agnivade/levenshtein"
	"github.com/david-botos/BearHug/services/analysis/internal/hsds_types"
	"github.com/david-botos/BearHug/services/analysis/internal/supabase"
	"github.com/david-botos/BearHug/services/analysis/pkg/logger"
)

type ServiceVerificationResult struct {
	ExistingService  *hsds_types.Service    // Nil if no match found
	ExtractedService ExtractedService       // The original extracted service
	IsNew            bool                   // True if this is a new service
	HasChanges       bool                   // True if existing service needs updates
	Changes          map[string]interface{} // Fields that differ between existing and extracted
}

type ServiceVerificationResults struct {
	NewServices       []ExtractedService          // Services to be created
	UpdateServices    []ServiceVerificationResult // Services that need updating
	UnchangedServices []*hsds_types.Service
	Error             error // Any error that occurred during verification
}

func VerifyServiceUniqueness(services ServicesExtracted, organizationID string) (ServiceVerificationResults, error) {
	log := logger.Get()
	log.Info().
		Str("organization_id", organizationID).
		Int("services_count", len(services.NewServices)).
		Msg("Starting service verification")

	// Fetch existing services from Supabase
	existingServices, err := supabase.FetchOrganizationServices(organizationID)
	if err != nil {
		log.Error().
			Err(err).
			Str("organization_id", organizationID).
			Msg("Failed to fetch existing services")
		return ServiceVerificationResults{}, fmt.Errorf("failed to fetch existing services: %w", err)
	}

	log.Debug().
		Int("existing_services_count", len(existingServices)).
		Msg("Retrieved existing services")

	results := ServiceVerificationResults{
		NewServices:       make([]ExtractedService, 0),
		UpdateServices:    make([]ServiceVerificationResult, 0),
		UnchangedServices: make([]*hsds_types.Service, 0),
	}

	// Helper function to check if two strings are similar
	isSimilarString := func(s1, s2 string) bool {
		// Convert both strings to lowercase for comparison
		s1 = strings.ToLower(strings.TrimSpace(s1))
		s2 = strings.ToLower(strings.TrimSpace(s2))

		// Direct match
		if s1 == s2 {
			return true
		}

		// Levenshtein distance check for similar names
		distance := levenshtein.ComputeDistance(s1, s2)
		maxLen := math.Max(float64(len(s1)), float64(len(s2)))
		// Allow for some fuzzy matching (e.g., if strings are 80% similar)
		return float64(distance) <= maxLen*0.2
	}

	// Helper function to detect changes between existing and extracted service
	detectChanges := func(existing *hsds_types.Service, extracted *ExtractedService) map[string]interface{} {
		changes := make(map[string]interface{})

		// Compare fields and record changes
		if existing.Description != nil && *existing.Description != extracted.Description {
			changes["description"] = extracted.Description
		}
		if !reflect.DeepEqual(existing.AlternateName, extracted.AlternateName) {
			changes["alternate_name"] = extracted.AlternateName
		}
		if !reflect.DeepEqual(existing.URL, extracted.URL) {
			changes["url"] = extracted.URL
		}
		if !reflect.DeepEqual(existing.Email, extracted.Email) {
			changes["email"] = extracted.Email
		}
		if !reflect.DeepEqual(existing.InterpretationServices, extracted.InterpretationServices) {
			changes["interpretation_services"] = extracted.InterpretationServices
		}
		if !reflect.DeepEqual(existing.ApplicationProcess, extracted.ApplicationProcess) {
			changes["application_process"] = extracted.ApplicationProcess
		}
		if !reflect.DeepEqual(existing.FeesDescription, extracted.FeesDescription) {
			changes["fees_description"] = extracted.FeesDescription
		}
		if !reflect.DeepEqual(existing.Accreditations, extracted.Accreditations) {
			changes["accreditations"] = extracted.Accreditations
		}
		if !reflect.DeepEqual(existing.EligibilityDescription, extracted.EligibilityDescription) {
			changes["eligibility_description"] = extracted.EligibilityDescription
		}
		if !reflect.DeepEqual(existing.MinimumAge, extracted.MinimumAge) {
			changes["minimum_age"] = extracted.MinimumAge
		}
		if !reflect.DeepEqual(existing.MaximumAge, extracted.MaximumAge) {
			changes["maximum_age"] = extracted.MaximumAge
		}
		if !reflect.DeepEqual(existing.Alert, extracted.Alert) {
			changes["alert"] = extracted.Alert
		}

		return changes
	}

	// Create a map to track which existing services have been processed
	processedServices := make(map[string]bool)

	// Check each extracted service against existing services
	for _, extractedService := range services.NewServices {
		found := false

		log.Debug().
			Str("service_name", extractedService.Name).
			Msg("Checking service for uniqueness")

		for _, existingService := range existingServices {
			// Check if service names are similar
			if isSimilarString(existingService.Name, extractedService.Name) {
				found = true
				processedServices[existingService.ID] = true

				// Detect what fields have changed
				changes := detectChanges(&existingService, &extractedService)

				if len(changes) > 0 {
					log.Info().
						Str("service_id", existingService.ID).
						Str("service_name", existingService.Name).
						Int("changes_count", len(changes)).
						Interface("changes", changes).
						Msg("Service needs updates")

					results.UpdateServices = append(results.UpdateServices, ServiceVerificationResult{
						ExistingService:  &existingService,
						ExtractedService: extractedService,
						IsNew:            false,
						HasChanges:       true,
						Changes:          changes,
					})
				} else {
					log.Debug().
						Str("service_id", existingService.ID).
						Str("service_name", existingService.Name).
						Msg("Service matched but no changes needed")

					// Add to UnchangedServices since we found a match with no changes
					results.UnchangedServices = append(results.UnchangedServices, &existingService)
				}
				break
			}
		}

		if !found {
			log.Info().
				Str("service_name", extractedService.Name).
				Msg("New service identified")
			results.NewServices = append(results.NewServices, extractedService)
		}
	}

	// Add any existing services that weren't processed (weren't part of the extraction)
	for _, existingService := range existingServices {
		if !processedServices[existingService.ID] {
			log.Debug().
				Str("service_id", existingService.ID).
				Str("service_name", existingService.Name).
				Msg("Adding unprocessed existing service to unchanged services")
			results.UnchangedServices = append(results.UnchangedServices, &existingService)
		}
	}

	log.Info().
		Int("new_services", len(results.NewServices)).
		Int("updated_services", len(results.UpdateServices)).
		Int("unchanged_services", len(results.UnchangedServices)).
		Msg("Service verification completed")

	return results, nil
}

func UpdateExistingServices(services []ServiceVerificationResult, callID string) error {
	log := logger.Get()
	log.Info().
		Int("services_count", len(services)).
		Msg("Starting service updates")

	client, err := supabase.InitSupabaseClient()
	if err != nil {
		log.Error().
			Err(err).
			Msg("Failed to initialize Supabase client")
		return fmt.Errorf("failed to initialize Supabase client: %w", err)
	}

	for _, service := range services {
		if !service.HasChanges {
			continue
		}

		log.Debug().
			Str("service_id", service.ExistingService.ID).
			Str("service_name", service.ExistingService.Name).
			Interface("changes", service.Changes).
			Msg("Updating service")

		// Prepare update data using the Changes map
		updateData := make(map[string]interface{})

		// Prepare metadata inputs for each changed field
		var metadataInputs []supabase.MetadataInput

		// Add all changed fields to the update data and create metadata
		for field, newValue := range service.Changes {
			updateData[field] = newValue

			// Get previous value from existing service
			var previousValue string
			if existingValue, ok := getFieldValue(service.ExistingService, field); ok {
				previousValue = fmt.Sprintf("%v", existingValue)
			} else {
				previousValue = "new data"
			}

			metadataInput := supabase.MetadataInput{
				ResourceID:       service.ExistingService.ID,
				CallID:           callID,
				ResourceType:     "service",
				FieldName:        field,
				PreviousValue:    previousValue,
				ReplacementValue: fmt.Sprintf("%v", newValue),
				LastActionType:   "UPDATE",
			}
			metadataInputs = append(metadataInputs, metadataInput)
		}

		// Add last_modified timestamp
		updateData["last_modified"] = time.Now()

		// Update the service in Supabase
		data, _, err := client.From("service").
			Update(updateData, "", "").
			Eq("id", service.ExistingService.ID).
			Execute()
		if err != nil {
			log.Error().
				Err(err).
				Str("service_id", service.ExistingService.ID).
				Str("service_name", service.ExistingService.Name).
				Str("response_data", string(data)).
				Msg("Failed to update service")
			return fmt.Errorf("failed to update service %s: %w, data: %s",
				service.ExistingService.ID, err, string(data))
		}

		// Create metadata entries for all changes
		if err := supabase.CreateAndStoreMetadata(metadataInputs); err != nil {
			log.Error().
				Err(err).
				Str("service_id", service.ExistingService.ID).
				Msg("Failed to create metadata entries")
			return fmt.Errorf("failed to create metadata entries for service %s: %w",
				service.ExistingService.ID, err)
		}

		log.Info().
			Str("service_id", service.ExistingService.ID).
			Str("service_name", service.ExistingService.Name).
			Msg("Successfully updated service")
	}

	log.Info().Msg("All service updates completed successfully")
	return nil
}

// Helper function to get field value from service using reflection
func getFieldValue(service *hsds_types.Service, fieldName string) (interface{}, bool) {
	val := reflect.ValueOf(service).Elem()
	field := val.FieldByName(fieldName)
	if !field.IsValid() {
		return nil, false
	}
	return field.Interface(), true
}
