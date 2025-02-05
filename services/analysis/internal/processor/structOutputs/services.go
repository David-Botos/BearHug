package structOutputs

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/david-botos/BearHug/services/analysis/internal/hsds_types"
	"github.com/david-botos/BearHug/services/analysis/internal/processor/inference"
	"github.com/david-botos/BearHug/services/analysis/internal/supabase"
	"github.com/david-botos/BearHug/services/analysis/pkg/logger"
)

func GenerateServicesPrompt(organization_id string, transcript string) (string, inference.ToolInputSchema, error) {
	log := logger.Get()
	log.Debug().
		Str("organization_id", organization_id).
		Msg("Generating services prompt")

	orgName, orgNameFetchErr := supabase.FetchOrganizationName(organization_id)
	if orgNameFetchErr != nil {
		log.Error().
			Err(orgNameFetchErr).
			Str("organization_id", organization_id).
			Msg("Failed to fetch organization name")
		return "", inference.ToolInputSchema{}, fmt.Errorf("organization_lookup_failed: %w", orgNameFetchErr)
	}

	log.Debug().
		Str("organization_name", orgName).
		Msg("Successfully fetched organization name")

	orgServices, orgServicesFetchErr := supabase.FetchOrganizationServices(organization_id)
	if orgServicesFetchErr != nil {
		log.Error().
			Err(orgServicesFetchErr).
			Str("organization_id", organization_id).
			Msg("Failed to fetch organization services")
		return "", inference.ToolInputSchema{}, fmt.Errorf("services_lookup_failed: %w", orgServicesFetchErr)
	}

	log.Debug().
		Int("services_count", len(orgServices)).
		Msg("Successfully fetched organization services")

	// Format existing services into a readable string
	var servicesText string
	if len(orgServices) == 0 {
		servicesText = "No services are currently documented for this organization."
	} else {
		var servicesList strings.Builder
		for i, service := range orgServices {
			if service.Status == hsds_types.ServiceStatusActive {
				if service.AlternateName != nil {
					servicesList.WriteString(fmt.Sprintf("%d. %s AKA %s\n", i+1, service.Name, *service.AlternateName))
				} else {
					servicesList.WriteString(fmt.Sprintf("%d. %s\n", i+1, service.Name))
				}
				if service.Description != nil {
					servicesList.WriteString(fmt.Sprintf("Description: %s\n", *service.Description))
				}

				if i < len(orgServices)-1 {
					servicesList.WriteString("\n")
				}
			}
		}
		servicesText = servicesList.String()
	}

	prompt := fmt.Sprintf(`You are a service data extraction specialist that documents details about human services available in your community. Your task is to identify and structure information about new services mentioned in a conversation transcript for %s.

Transcript:
%s

Previously documented services for %s:
%s

IMPORTANT EXTRACTION RULES:
1. Break down composite services into their individual components. For example:
   - If "counseling services" includes both "group counseling" and "individual counseling", create separate entries for each
   - If a program has different delivery methods (in-person vs online), create separate entries
   - Each distinct service should stand alone with its own eligibility, fees, and application process

2. For each individual service:
   - Name it specifically (e.g., "Brain Trauma Individual Coaching" instead of just "Brain Trauma Services")
   - Include only confirmed details from the transcript
   - Default status to "active" unless otherwise indicated
   - Keep descriptions focused on that specific service only

3. Do NOT combine multiple services into a single entry, even if they serve similar populations

Only respond using the new_services tool to output the structured data. Do not provide any additional text.`, orgName, transcript, orgName, servicesText)

	return prompt, ServicesSchema, nil
}

var ServicesSchema = inference.ToolInputSchema{
	Type: "object",
	Properties: map[string]inference.Property{
		"new_services": {
			Type:        "array",
			Description: "Array of new services identified in the conversation",
			Items: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Primary name of the service",
					},
					"status": map[string]interface{}{
						"type": "string",
						"enum": []string{"active", "inactive", "defunct"},
					},
					"description": map[string]interface{}{
						"type": "string",
					},
					"application_process": map[string]interface{}{
						"type": "string",
					},
					"fees_description": map[string]interface{}{
						"type": "string",
					},
					"eligibility_description": map[string]interface{}{
						"type": "string",
					},
					"wait_time": map[string]interface{}{
						"type":        "string",
						"description": "Current wait time for service access",
					},
				},
				"required": []string{"name", "status", "description"},
			},
		},
	},
	Required: []string{"new_services"},
}

type ExtractedService struct {
	// Required fields
	Name        string                       `json:"name"`
	Status      hsds_types.ServiceStatusEnum `json:"status"`
	Description string                       `json:"description"`

	// Optional fields
	AlternateName          *string  `json:"alternate_name,omitempty"`
	URL                    *string  `json:"url,omitempty"`
	Email                  *string  `json:"email,omitempty"`
	InterpretationServices *string  `json:"interpretation_services,omitempty"`
	ApplicationProcess     *string  `json:"application_process,omitempty"`
	FeesDescription        *string  `json:"fees_description,omitempty"`
	Accreditations         *string  `json:"accreditations,omitempty"`
	EligibilityDescription *string  `json:"eligibility_description,omitempty"`
	MinimumAge             *float64 `json:"minimum_age,omitempty"`
	MaximumAge             *float64 `json:"maximum_age,omitempty"`
	Alert                  *string  `json:"alert,omitempty"`
	WaitTime               *string  `json:"wait_time,omitempty"`
}
type ServicesExtracted struct {
	NewServices []ExtractedService `json:"new_services"`
}

func ServicesExtraction(org_id string, transcript string) (ServicesExtracted, error) {
	log := logger.Get()
	log.Info().
		Str("organization_id", org_id).
		Int("transcript_length", len(transcript)).
		Msg("Starting services extraction")

	servicesPrompt, servicesSchema, promptErr := GenerateServicesPrompt(org_id, transcript)
	if promptErr != nil {
		log.Error().
			Err(promptErr).
			Str("organization_id", org_id).
			Msg("Failed to generate services prompt")
		return ServicesExtracted{}, fmt.Errorf("failed to generate services prompt: %w", promptErr)
	}

	client, err := inference.InitInferenceClient()
	if err != nil {
		log.Error().
			Err(err).
			Msg("Failed to initialize inference client")
		return ServicesExtracted{}, fmt.Errorf("failed to initialize inference client: %w", err)
	}

	log.Debug().Msg("Running Claude inference for services extraction")
	servicesInferenceResult, servicesInferenceResultErr := client.RunClaudeInference(inference.PromptParams{Prompt: servicesPrompt, Schema: servicesSchema})
	if servicesInferenceResultErr != nil {
		log.Error().
			Err(servicesInferenceResultErr).
			Msg("Error occurred during services inference")
		return ServicesExtracted{}, fmt.Errorf("error reading response: %w", servicesInferenceResultErr)
	}

	jsonBytes, err := json.Marshal(servicesInferenceResult)
	if err != nil {
		log.Error().
			Err(err).
			Msg("Failed to marshal response map to JSON")
		return ServicesExtracted{}, fmt.Errorf("failed to marshal response map: %w", err)
	}

	var servicesExtracted ServicesExtracted
	if err := json.Unmarshal(jsonBytes, &servicesExtracted); err != nil {
		log.Error().
			Err(err).
			Msg("Failed to unmarshal Claude inference response")
		return ServicesExtracted{}, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	log.Info().
		Int("extracted_services_count", len(servicesExtracted.NewServices)).
		Msg("Successfully completed services extraction")

	return servicesExtracted, nil
}

func HandleExtractedServices(extractedServices ServicesExtracted, organizationID string, callID string) (ServiceContext, error) {
	log := logger.Get()
	log.Info().
		Str("organization_id", organizationID).
		Int("services_count", len(extractedServices.NewServices)).
		Msg("Starting to handle extracted services")

	verificationResults, err := VerifyServiceUniqueness(extractedServices, organizationID)
	if err != nil {
		return ServiceContext{}, fmt.Errorf("failed to verify service uniqueness: %w", err)
	}

	serviceContext := ServiceContext{
		ExistingServices: make([]*hsds_types.Service, 0),
		NewServices:      make([]*hsds_types.Service, 0),
	}

	// Convert new services to HSDS format
	if len(verificationResults.NewServices) > 0 {
		for _, extractedService := range verificationResults.NewServices {
			opts := &hsds_types.ServiceOptions{
				Description:            &extractedService.Description,
				ApplicationProcess:     extractedService.ApplicationProcess,
				FeesDescription:        extractedService.FeesDescription,
				EligibilityDescription: extractedService.EligibilityDescription,
				WaitTime:               extractedService.WaitTime,
			}

			service, err := hsds_types.NewService(
				organizationID,
				extractedService.Name,
				extractedService.Status,
				opts,
			)
			if err != nil {
				return ServiceContext{}, fmt.Errorf("error converting new service '%s': %w", extractedService.Name, err)
			}

			hsdsService := &hsds_types.Service{
				ID:                     service.ID,
				OrganizationID:         service.OrganizationID,
				Name:                   service.Name,
				Status:                 service.Status,
				Description:            service.Description,
				ApplicationProcess:     service.ApplicationProcess,
				FeesDescription:        service.FeesDescription,
				EligibilityDescription: service.EligibilityDescription,
				WaitTime:               service.WaitTime,
				CreatedAt:              service.CreatedAt,
			}
			serviceContext.NewServices = append(serviceContext.NewServices, hsdsService)
		}

		if err := supabase.StoreNewServices(serviceContext.NewServices, callID); err != nil {
			return ServiceContext{}, fmt.Errorf("failed to store new services: %w", err)
		}
	}

	// Handle existing service updates
	if len(verificationResults.UpdateServices) > 0 {
		if err := UpdateExistingServices(verificationResults.UpdateServices, callID); err != nil {
			return ServiceContext{}, fmt.Errorf("failed to update existing services: %w", err)
		}
		for _, updatedService := range verificationResults.UpdateServices {
			serviceContext.ExistingServices = append(serviceContext.ExistingServices, updatedService.ExistingService)
		}
	}

	serviceContext.ExistingServices = append(serviceContext.ExistingServices, verificationResults.UnchangedServices...)

	return serviceContext, nil
}
