package structOutputs

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/david-botos/BearHug/services/analysis/internal/hsds_types"
	"github.com/david-botos/BearHug/services/analysis/internal/processor/inference"
	"github.com/david-botos/BearHug/services/analysis/internal/supabase"
	"github.com/david-botos/BearHug/services/analysis/internal/types"
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

	prompt := fmt.Sprintf(`You are a service data extraction specialist that documents details about human services available to the underprivileged in your community. Your task is to identify and structure information about new services mentioned in a conversation transcript, not currently known for %s. You have been provided with both the transcript and services that are currently documented for %s, the organization that was called to generate the transcript.

Transcript:
%s

Services that were previously documented as being offered by %s:
%s

Your task is to extract detailed information about each new service mentioned and structure it according to the provided schema. For each service:
1. Create a complete service entry with all required fields (name, status, description)
2. Include any optional fields that were explicitly mentioned or can be easily inferred in the transcript. Do not invent details that are not implied.
3. Use clear, objective language for descriptions
4. Don't worry about capturing scheduling details or capacity information about each service, that will be captured in another table
5. If you aren't confident that the service mentioned in the call was already documented by 

Guidelines for extraction:
- If multiple similar services are mentioned (e.g., different medical services), create separate entries for each distinct service
- Default service status to "active" unless otherwise indicated
- Extract any eligibility requirements or application processes mentioned
- Capture specific details about fees (if the service is free state "Free" and nothing else)

IMPORTANT: You must ONLY respond by using the new_services tool to output the structured data. Do not provide any explanatory text, confirmations, or additional messages. Simply use the tool to output the structured data following the schema exactly.`, orgName, orgName, transcript, orgName, servicesText)

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
					// Handled Manually
					// "organization_id": map[string]interface{}{
					// 	"type":        "string",
					// 	"description": "UUID v4 of the organization offering the service",
					// },
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Primary name of the service",
					},
					"status": map[string]interface{}{
						"type":        "string",
						"description": "Current operational status of the service",
						"enum":        []string{"active", "inactive", "defunct"},
					},
					// Handled Manually
					// "program_id": map[string]interface{}{
					// 	"type":        "string",
					// 	"description": "Optional UUID v4 of the program this service belongs to",
					// },
					"alternate_name": map[string]interface{}{
						"type":        "string",
						"description": "Alternative name or abbreviation for the service",
					},
					"description": map[string]interface{}{
						"type":        "string",
						"description": "Detailed description of what the service provides",
					},
					"url": map[string]interface{}{
						"type":        "string",
						"description": "Website or webpage for the service",
					},
					"email": map[string]interface{}{
						"type":        "string",
						"description": "Contact email for the service",
					},
					"interpretation_services": map[string]interface{}{
						"type":        "string",
						"description": "Languages and interpretation services available",
					},
					"application_process": map[string]interface{}{
						"type":        "string",
						"description": "Steps required to apply for or access the service",
					},
					"fees_description": map[string]interface{}{
						"type":        "string",
						"description": "Detailed description of any fees or costs",
					},
					"accreditations": map[string]interface{}{
						"type":        "string",
						"description": "Any professional accreditations or certifications",
					},
					"eligibility_description": map[string]interface{}{
						"type":        "string",
						"description": "Who is eligible to receive this service",
					},
					"minimum_age": map[string]interface{}{
						"type":        "number",
						"description": "Minimum age requirement for service recipients",
					},
					"maximum_age": map[string]interface{}{
						"type":        "number",
						"description": "Maximum age limit for service recipients",
					},
					"alert": map[string]interface{}{
						"type":        "string",
						"description": "Important notices or warnings about the service",
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
}

type ServicesInput struct {
	NewServices []ExtractedService `json:"new_services"`
}
type ServicesExtracted struct {
	Input ServicesInput `json:"input"`
}

func ServicesExtraction(params types.TranscriptsReqBody) (ServicesExtracted, error) {
	log := logger.Get()
	log.Info().
		Str("organization_id", params.OrganizationID).
		Int("transcript_length", len(params.Transcript)).
		Msg("Starting services extraction")

	servicesPrompt, servicesSchema, promptErr := GenerateServicesPrompt(params.OrganizationID, params.Transcript)
	if promptErr != nil {
		log.Error().
			Err(promptErr).
			Str("organization_id", params.OrganizationID).
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
		Int("extracted_services_count", len(servicesExtracted.Input.NewServices)).
		Msg("Successfully completed services extraction")

	return servicesExtracted, nil
}

func HandleExtractedServices(services ServicesExtracted, organizationID string) (*ServiceContext, error) {
	log := logger.Get()
	log.Info().
		Str("organization_id", organizationID).
		Int("services_count", len(services.Input.NewServices)).
		Msg("Starting to handle extracted services")

	if !hsds_types.ValidateUUID(organizationID) {
		log.Error().
			Str("organization_id", organizationID).
			Msg("Invalid organization ID format")
		return nil, fmt.Errorf("invalid organization ID format: must be UUIDv4")
	}

	verificationResults, err := VerifyServiceUniqueness(services, organizationID)
	if err != nil {
		log.Error().
			Err(err).
			Str("organization_id", organizationID).
			Msg("Failed to verify service uniqueness")
		return nil, fmt.Errorf("failed to verify service uniqueness: %w", err)
	}

	log.Debug().
		Int("new_services", len(verificationResults.NewServices)).
		Int("update_services", len(verificationResults.UpdateServices)).
		Msg("Service verification completed")

	serviceContext := &ServiceContext{
		ExistingServices: make([]*hsds_types.Service, 0),
		NewServices:      make([]*hsds_types.Service, 0),
	}

	// Convert new services to proper HSDS format
	if len(verificationResults.NewServices) > 0 {
		for _, extractedService := range verificationResults.NewServices {
			log.Debug().
				Str("service_name", extractedService.Name).
				Msg("Processing new service")
			opts := &hsds_types.ServiceOptions{
				AlternateName:          extractedService.AlternateName,
				Description:            &extractedService.Description,
				URL:                    extractedService.URL,
				Email:                  extractedService.Email,
				InterpretationServices: extractedService.InterpretationServices,
				ApplicationProcess:     extractedService.ApplicationProcess,
				FeesDescription:        extractedService.FeesDescription,
				Accreditations:         extractedService.Accreditations,
				EligibilityDescription: extractedService.EligibilityDescription,
				MinimumAge:             extractedService.MinimumAge,
				MaximumAge:             extractedService.MaximumAge,
				Alert:                  extractedService.Alert,
			}

			service, err := hsds_types.NewService(
				organizationID,
				extractedService.Name,
				extractedService.Status,
				opts,
			)
			if err != nil {
				return nil, fmt.Errorf("error converting new service '%s': %w", extractedService.Name, err)
			}

			hsdsService := &hsds_types.Service{
				ID:                     service.ID,
				OrganizationID:         service.OrganizationID,
				Name:                   service.Name,
				Status:                 service.Status,
				ProgramID:              service.ProgramID,
				AlternateName:          service.AlternateName,
				Description:            service.Description,
				URL:                    service.URL,
				Email:                  service.Email,
				InterpretationServices: service.InterpretationServices,
				ApplicationProcess:     service.ApplicationProcess,
				FeesDescription:        service.FeesDescription,
				EligibilityDescription: service.EligibilityDescription,
				MinimumAge:             service.MinimumAge,
				MaximumAge:             service.MaximumAge,
				Alert:                  service.Alert,
				WaitTime:               service.WaitTime,
				Fees:                   service.Fees,
				Licenses:               service.Licenses,
				Accreditations:         service.Accreditations,
				AssuredDate:            service.AssuredDate,
				AssurerEmail:           service.AssurerEmail,
				LastModified:           service.LastModified,
				CreatedAt:              service.CreatedAt,
				UpdatedAt:              service.UpdatedAt,
			}
			serviceContext.NewServices = append(serviceContext.NewServices, hsdsService)
		}

		// Store the new services
		if err := supabase.StoreNewServices(serviceContext.NewServices); err != nil {
			log.Error().
				Err(err).
				Int("services_count", len(serviceContext.NewServices)).
				Msg("Failed to store new services")
			return nil, fmt.Errorf("failed to store new services: %w", err)
		}

		log.Info().
			Int("stored_services_count", len(serviceContext.NewServices)).
			Msg("Successfully stored new services")
	}

	// Handle updates to existing services
	if len(verificationResults.UpdateServices) > 0 {
		if err := UpdateExistingServices(verificationResults.UpdateServices); err != nil {
			log.Error().
				Err(err).
				Int("update_count", len(verificationResults.UpdateServices)).
				Msg("Failed to update existing services")
			return nil, fmt.Errorf("failed to update existing services: %w", err)
		}

		log.Info().
			Int("updated_services_count", len(verificationResults.UpdateServices)).
			Msg("Successfully updated existing services")
	}

	return serviceContext, nil
}
