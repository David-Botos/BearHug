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
)

type ServiceVerificationResult struct {
	ExistingService  *hsds_types.Service    // Nil if no match found
	ExtractedService ExtractedService       // The original extracted service
	IsNew            bool                   // True if this is a new service
	HasChanges       bool                   // True if existing service needs updates
	Changes          map[string]interface{} // Fields that differ between existing and extracted
}

type ServiceVerificationResults struct {
	NewServices    []ExtractedService          // Services to be created
	UpdateServices []ServiceVerificationResult // Services that need updating
	Error          error                       // Any error that occurred during verification
}

func VerifyServiceUniqueness(services ServicesExtracted, organizationID string) (*ServiceVerificationResults, error) {
	// Fetch existing services from Supabase
	existingServices, err := supabase.FetchOrganizationServices(organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch existing services: %w", err)
	}

	results := &ServiceVerificationResults{
		NewServices:    make([]ExtractedService, 0),
		UpdateServices: make([]ServiceVerificationResult, 0),
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

	// Check each extracted service against existing services
	for _, extractedService := range services.Input.NewServices {
		found := false

		for _, existingService := range existingServices {
			// Check if service names are similar
			if isSimilarString(existingService.Name, extractedService.Name) {
				found = true

				// Detect what fields have changed
				changes := detectChanges(&existingService, &extractedService)

				if len(changes) > 0 {
					results.UpdateServices = append(results.UpdateServices, ServiceVerificationResult{
						ExistingService:  &existingService,
						ExtractedService: extractedService,
						IsNew:            false,
						HasChanges:       true,
						Changes:          changes,
					})
				}
				break
			}
		}

		if !found {
			// This is a new service
			results.NewServices = append(results.NewServices, extractedService)
		}
	}

	return results, nil
}

func UpdateExistingServices(services []ServiceVerificationResult) error {
	client, err := supabase.InitSupabaseClient()
	if err != nil {
		return fmt.Errorf("failed to initialize Supabase client: %w", err)
	}

	for _, service := range services {
		if !service.HasChanges {
			continue
		}

		// Prepare update data using the Changes map
		updateData := make(map[string]interface{})

		// Add all changed fields to the update data
		for field, value := range service.Changes {
			updateData[field] = value
		}

		// Add last_modified timestamp
		updateData["last_modified"] = time.Now()

		// Update the service in Supabase
		data, _, err := client.From("service").
			Update(updateData, "", "").
			Eq("id", service.ExistingService.ID).
			Execute()

		if err != nil {
			return fmt.Errorf("failed to update service %s: %w, data: %s",
				service.ExistingService.ID, err, string(data))
		}
	}

	return nil
}
