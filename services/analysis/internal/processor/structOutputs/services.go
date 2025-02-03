package structOutputs

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/david-botos/BearHug/services/analysis/internal/hsds_types"
	"github.com/david-botos/BearHug/services/analysis/internal/processor/inference"
	"github.com/david-botos/BearHug/services/analysis/internal/supabase"
	"github.com/david-botos/BearHug/services/analysis/pkg/logger"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
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

func ServicesExtraction(ctx context.Context, org_id string, transcript string) (ServicesExtracted, error) {
	tracer := otel.GetTracerProvider().Tracer("services-extraction")
	ctx, span := tracer.Start(ctx, "services_extraction",
		trace.WithAttributes(
			attribute.String("organization_id", org_id),
			attribute.Int("transcript_length", len(transcript)),
		),
	)
	defer span.End()

	log := logger.Get()
	log.Info().
		Str("organization_id", org_id).
		Int("transcript_length", len(transcript)).
		Msg("Starting services extraction")

	// Generate prompt with tracing
	ctx, promptSpan := tracer.Start(ctx, "generate_services_prompt")
	servicesPrompt, servicesSchema, promptErr := GenerateServicesPrompt(org_id, transcript)
	if promptErr != nil {
		promptSpan.RecordError(promptErr)
		promptSpan.End()
		log.Error().
			Err(promptErr).
			Str("organization_id", org_id).
			Msg("Failed to generate services prompt")
		return ServicesExtracted{}, fmt.Errorf("failed to generate services prompt: %w", promptErr)
	}
	promptSpan.SetAttributes(
		attribute.Int("prompt_length", len(servicesPrompt)),
	)
	promptSpan.End()

	// Initialize client with tracing
	ctx, clientSpan := tracer.Start(ctx, "init_inference_client")
	client, err := inference.InitInferenceClient()
	if err != nil {
		clientSpan.RecordError(err)
		clientSpan.End()
		log.Error().
			Err(err).
			Msg("Failed to initialize inference client")
		return ServicesExtracted{}, fmt.Errorf("failed to initialize inference client: %w", err)
	}
	clientSpan.End()

	log.Debug().Msg("Running Claude inference for services extraction")
	servicesInferenceResult, servicesInferenceResultErr := client.RunClaudeInference(ctx, inference.PromptParams{
		Prompt: servicesPrompt,
		Schema: servicesSchema,
	})
	if servicesInferenceResultErr != nil {
		span.RecordError(servicesInferenceResultErr)
		log.Error().
			Err(servicesInferenceResultErr).
			Msg("Error occurred during services inference")
		return ServicesExtracted{}, fmt.Errorf("error reading response: %w", servicesInferenceResultErr)
	}

	// Process results with tracing
	ctx, processSpan := tracer.Start(ctx, "process_inference_results")
	jsonBytes, err := json.Marshal(servicesInferenceResult)
	if err != nil {
		processSpan.RecordError(err)
		processSpan.End()
		log.Error().
			Err(err).
			Msg("Failed to marshal response map to JSON")
		return ServicesExtracted{}, fmt.Errorf("failed to marshal response map: %w", err)
	}

	var servicesExtracted ServicesExtracted
	if err := json.Unmarshal(jsonBytes, &servicesExtracted); err != nil {
		processSpan.RecordError(err)
		processSpan.End()
		log.Error().
			Err(err).
			Msg("Failed to unmarshal Claude inference response")
		return ServicesExtracted{}, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	processSpan.SetAttributes(
		attribute.Int("extracted_services_count", len(servicesExtracted.NewServices)),
	)
	processSpan.End()

	// Set final success attributes on main span
	span.SetAttributes(
		attribute.Int("total_services_extracted", len(servicesExtracted.NewServices)),
		attribute.Bool("success", true),
	)

	log.Info().
		Int("extracted_services_count", len(servicesExtracted.NewServices)).
		Msg("Successfully completed services extraction")

	return servicesExtracted, nil
}

func HandleExtractedServices(ctx context.Context, extractedServices ServicesExtracted, organizationID string, callID string) (ServiceContext, error) {
	tracer := otel.GetTracerProvider().Tracer("services-handler")
	ctx, span := tracer.Start(ctx, "handle_extracted_services",
		trace.WithAttributes(
			attribute.String("organization_id", organizationID),
			attribute.String("call_id", callID),
			attribute.Int("extracted_services_count", len(extractedServices.NewServices)),
		),
	)
	defer span.End()

	log := logger.Get()
	log.Info().
		Str("organization_id", organizationID).
		Int("services_count", len(extractedServices.NewServices)).
		Msg("Starting to handle extracted services")

	// Verify service uniqueness with tracing
	ctx, verifySpan := tracer.Start(ctx, "verify_service_uniqueness")
	verificationResults, err := VerifyServiceUniqueness(ctx, extractedServices, organizationID)
	if err != nil {
		verifySpan.RecordError(err)
		verifySpan.End()
		return ServiceContext{}, fmt.Errorf("failed to verify service uniqueness: %w", err)
	}
	verifySpan.SetAttributes(
		attribute.Int("new_services_count", len(verificationResults.NewServices)),
		attribute.Int("update_services_count", len(verificationResults.UpdateServices)),
		attribute.Int("unchanged_services_count", len(verificationResults.UnchangedServices)),
	)
	verifySpan.End()

	serviceContext := ServiceContext{
		ExistingServices: make([]*hsds_types.Service, 0),
		NewServices:      make([]*hsds_types.Service, 0),
	}

	// Convert new services to HSDS format with tracing
	if len(verificationResults.NewServices) > 0 {
		ctx, convertSpan := tracer.Start(ctx, "convert_new_services")
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
				convertSpan.RecordError(err)
				convertSpan.End()
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
				LastModified:           service.LastModified,
				CreatedAt:              service.CreatedAt,
				UpdatedAt:              service.UpdatedAt,
			}
			serviceContext.NewServices = append(serviceContext.NewServices, hsdsService)
		}
		convertSpan.SetAttributes(
			attribute.Int("converted_services_count", len(serviceContext.NewServices)),
		)
		convertSpan.End()

		// Store new services with tracing
		ctx, storeSpan := tracer.Start(ctx, "store_new_services")
		if err := supabase.StoreNewServices(ctx, serviceContext.NewServices, callID); err != nil {
			storeSpan.RecordError(err)
			storeSpan.End()
			return ServiceContext{}, fmt.Errorf("failed to store new services: %w", err)
		}
		storeSpan.SetAttributes(
			attribute.Int("stored_services_count", len(serviceContext.NewServices)),
		)
		storeSpan.End()
	}

	// Handle existing service updates with tracing
	if len(verificationResults.UpdateServices) > 0 {
		ctx, updateSpan := tracer.Start(ctx, "update_existing_services")
		if err := UpdateExistingServices(ctx, verificationResults.UpdateServices, callID); err != nil {
			updateSpan.RecordError(err)
			updateSpan.End()
			return ServiceContext{}, fmt.Errorf("failed to update existing services: %w", err)
		}
		for _, updatedService := range verificationResults.UpdateServices {
			serviceContext.ExistingServices = append(serviceContext.ExistingServices, updatedService.ExistingService)
		}
		updateSpan.SetAttributes(
			attribute.Int("updated_services_count", len(verificationResults.UpdateServices)),
		)
		updateSpan.End()
	}

	serviceContext.ExistingServices = append(serviceContext.ExistingServices, verificationResults.UnchangedServices...)

	// Set final metrics on the main span
	span.SetAttributes(
		attribute.Int("total_services_processed", len(serviceContext.NewServices)+len(serviceContext.ExistingServices)),
		attribute.Int("new_services_created", len(serviceContext.NewServices)),
		attribute.Int("existing_services_updated", len(verificationResults.UpdateServices)),
		attribute.Int("services_unchanged", len(verificationResults.UnchangedServices)),
	)

	return serviceContext, nil
}
