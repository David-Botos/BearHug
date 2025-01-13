package structOutputs

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"

	"github.com/agnivade/levenshtein"
	"github.com/david-botos/BearHug/services/analysis/internal/hsds_types"
	"github.com/david-botos/BearHug/services/analysis/internal/processor/inference"
	"github.com/david-botos/BearHug/services/analysis/pkg/logger"
)

func GenerateServiceCapacityPrompt(transcript string, serviceCtx ServiceContext) (string, inference.ToolInputSchema, error) {
	log := logger.Get()
	log.Debug().
		Int("existing_services", len(serviceCtx.ExistingServices)).
		Int("new_services", len(serviceCtx.NewServices)).
		Msg("Generating service capacity prompt")

	// Build service descriptions
	var existingServiceDesc, newServiceDesc strings.Builder

	// Process existing services
	for _, service := range serviceCtx.ExistingServices {
		writeServiceDescription(&existingServiceDesc, *service)
	}

	// Process new services
	for _, service := range serviceCtx.NewServices {
		if service != nil {
			writeServiceDescription(&newServiceDesc, *service)
		}
	}

	prompt := fmt.Sprintf(`Based on the following conversation transcript and service information, identify the available capacity for services mentioned:

Conversation Transcript:
%s

Current Service Information:
----------------------------
Existing Services (may or may not be mentioned in the transcript):
%s

New Services (extracted from the transcript directly):
%s

IMPORTANT: You must ONLY respond by using the capacities tool to output the structured data. Do not provide any explanatory text, confirmations, or additional messages. Simply use the tool to output the structured data following the schema exactly.`, transcript, existingServiceDesc.String(), newServiceDesc.String())

	log.Debug().Msg("Service capacity prompt generated successfully")
	return prompt, ServiceCapacitySchema, nil
}

var ServiceCapacitySchema = inference.ToolInputSchema{
	Type: "object",
	Properties: map[string]inference.Property{
		"capacities": {
			Type: "array",
			Items: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"serviceName": map[string]interface{}{
						"type":        "string",
						"description": "The name of the service mentioned in the prompt that this capacity describes",
					},
					"available": map[string]interface{}{
						"type":        "number",
						"description": "Current available quantity",
					},
					"maximum": map[string]interface{}{
						"type":        "number",
						"description": "Maximum possible quantity",
					},
					"unitName": map[string]interface{}{
						"type":        "string",
						"description": "Name of the unit of measurement",
					},
					"unitDescription": map[string]interface{}{
						"type":        "string",
						"description": "Human-readable description of what is being measured",
					},
				},
				"required": []string{"available", "unitName", "unitDescription"},
			},
		},
	},
	Required: []string{"capacities"},
}

func writeServiceDescription(builder *strings.Builder, service hsds_types.Service) {
	builder.WriteString(fmt.Sprintf("Service ID: %s\n", service.ID))
	builder.WriteString(fmt.Sprintf("Name: %s\n", service.Name))
	if service.Description != nil {
		builder.WriteString(fmt.Sprintf("Description: %s\n", *service.Description))
	}
	if service.Status != "" {
		builder.WriteString(fmt.Sprintf("Status: %s\n", service.Status))
	}
	builder.WriteString("\n")
}

// analyzeCapacityDetails processes service capacity and unit information
func analyzeCapacityDetails(transcript string, serviceCtx ServiceContext) (DetailAnalysisResult, error) {
	log := logger.Get()
	log.Debug().Msg("Starting capacity details analysis")

	// Generate Prompt and Schema
	capacityCategoryPrompt, capacitySchema, err := GenerateServiceCapacityPrompt(transcript, serviceCtx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate service capacity prompt")
		return DetailAnalysisResult{}, fmt.Errorf(`Failure when generating service capacity prompt: %w`, err)
	}

	// Declare Claude Inference Client
	client, err := inference.InitInferenceClient()
	if err != nil {
		log.Error().Err(err).Msg("Failed to initialize inference client")
		return DetailAnalysisResult{}, fmt.Errorf("failed to initialize inference client: %w", err)
	}

	log.Debug().Msg("Running Claude inference for capacity analysis")
	// Run inference
	unformattedCapacityDetails, inferenceErr := client.RunClaudeInference(inference.PromptParams{Prompt: capacityCategoryPrompt, Schema: capacitySchema})
	if inferenceErr != nil {
		log.Error().Err(inferenceErr).Msg("Error during inference execution")
		return DetailAnalysisResult{}, fmt.Errorf(`Error running inference to extract capacity details: %w`, inferenceErr)
	}

	log.Debug().Msg("Converting inference response to capacity and unit objects")
	capacityDetails, unitDetails, capacityAndUnitInfConvErr := infToCapacityAndUnits(unformattedCapacityDetails, serviceCtx)
	if capacityAndUnitInfConvErr != nil {
		log.Error().Err(capacityAndUnitInfConvErr).Msg("Failed to convert inference response")
		return DetailAnalysisResult{}, fmt.Errorf(`Error while converting the inference response to clean capacity and unit objects: %w`, capacityAndUnitInfConvErr)
	}

	var result DetailAnalysisResult = *NewCapacityResult(capacityDetails, unitDetails)
	log.Info().
		Int("capacities_count", len(capacityDetails)).
		Int("units_count", len(unitDetails)).
		Msg("Capacity analysis completed successfully")

	return result, nil
}

type capacityInference struct {
	ServiceName     string   `json:"serviceName"`
	Available       float64  `json:"available"`
	Maximum         *float64 `json:"maximum,omitempty"`
	UnitName        string   `json:"unitName"`
	UnitDescription string   `json:"unitDescription,omitempty"`
}

type capacityAndUnitInfOutput struct {
	Capacities []capacityInference `json:"capacities"`
}

type serviceMatchResult struct {
	inference capacityInference
	service   *hsds_types.Service
	matched   bool
}

func infToCapacityAndUnits(inferenceResult map[string]interface{}, serviceCtx ServiceContext) ([]*hsds_types.ServiceCapacity, []*hsds_types.Unit, error) {
	log := logger.Get()
	log.Debug().Msg("Starting inference result conversion")

	// Convert the inference result to our structured type
	jsonData, err := json.Marshal(inferenceResult)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal inference result")
		return nil, nil, fmt.Errorf("error marshaling inference result: %w", err)
	}

	var output capacityAndUnitInfOutput
	if err := json.Unmarshal(jsonData, &output); err != nil {
		log.Error().Err(err).Msg("Failed to unmarshal to structured output")
		return nil, nil, fmt.Errorf("error unmarshaling to structured output: %w", err)
	}

	// Combine existing and new services
	totalServices := make([]*hsds_types.Service, 0, len(serviceCtx.ExistingServices)+len(serviceCtx.NewServices))
	totalServices = append(totalServices, serviceCtx.ExistingServices...)
	for _, newService := range serviceCtx.NewServices {
		totalServices = append(totalServices, newService)
	}

	log.Debug().
		Int("total_services", len(totalServices)).
		Int("capacities", len(output.Capacities)).
		Msg("Starting service matching")

	// Track matching results
	matchResults := make([]serviceMatchResult, len(output.Capacities))
	for i, capacity := range output.Capacities {
		matchedService := findMatchingService(capacity, totalServices)
		matchResults[i] = serviceMatchResult{
			inference: capacity,
			service:   matchedService,
			matched:   matchedService != nil,
		}

		if matchedService != nil {
			log.Debug().
				Str("service_name", capacity.ServiceName).
				Str("matched_id", matchedService.ID).
				Msg("Service matched successfully")
		}
	}

	// Check for unmatched capacities
	var unmatched []string
	for _, result := range matchResults {
		if !result.matched {
			unmatched = append(unmatched, result.inference.ServiceName)
		}
	}
	if len(unmatched) > 0 {
		log.Error().
			Strs("unmatched_services", unmatched).
			Msg("Failed to match all services")
		return nil, nil, fmt.Errorf("unable to match services for: %s", strings.Join(unmatched, ", "))
	}

	// Create arrays for results
	units := make([]*hsds_types.Unit, 0, len(matchResults))
	capacities := make([]*hsds_types.ServiceCapacity, 0, len(matchResults))

	// Process matched capacities
	for _, match := range matchResults {
		log.Debug().
			Str("service_name", match.inference.ServiceName).
			Str("unit_name", match.inference.UnitName).
			Msg("Processing capacity entry")

		// Create the Unit
		unitOpts := &hsds_types.UnitOptions{}
		unit, err := hsds_types.NewUnit(match.inference.UnitName, unitOpts)
		if err != nil {
			log.Error().
				Err(err).
				Str("unit_name", match.inference.UnitName).
				Msg("Failed to create unit")
			return nil, nil, fmt.Errorf("error creating unit for %s: %w", match.inference.UnitName, err)
		}
		units = append(units, unit)

		// Create the ServiceCapacity
		capOpts := &hsds_types.ServiceCapacityOptions{
			Maximum: match.inference.Maximum,
		}
		if match.inference.UnitDescription != "" {
			capOpts.Description = &match.inference.UnitDescription
		}

		serviceCapacity, err := hsds_types.NewServiceCapacity(
			match.service.ID,
			unit.ID,
			match.inference.Available,
			capOpts,
		)
		if err != nil {
			log.Error().
				Err(err).
				Str("service_id", match.service.ID).
				Str("unit_id", unit.ID).
				Msg("Failed to create service capacity")
			return nil, nil, fmt.Errorf("error creating service capacity for unit %s: %w", unit.ID, err)
		}
		capacities = append(capacities, serviceCapacity)
	}

	log.Info().
		Int("units_created", len(units)).
		Int("capacities_created", len(capacities)).
		Msg("Successfully converted inference results")

	return capacities, units, nil
}

// findMatchingService attempts to find the corresponding service for a capacity
func findMatchingService(inf capacityInference, services []*hsds_types.Service) *hsds_types.Service {
	log := logger.Get()
	log.Debug().
		Str("service_name", inf.ServiceName).
		Int("services_to_check", len(services)).
		Msg("Finding matching service")

	normalizedInfName := strings.ToLower(strings.TrimSpace(inf.ServiceName))

	// Try exact name match
	for _, svc := range services {
		if strings.ToLower(strings.TrimSpace(svc.Name)) == normalizedInfName {
			log.Debug().
				Str("service_name", inf.ServiceName).
				Str("matched_id", svc.ID).
				Msg("Found exact name match")
			return svc
		}
	}

	// Try alternate name match
	for _, svc := range services {
		if svc.AlternateName != nil &&
			strings.ToLower(strings.TrimSpace(*svc.AlternateName)) == normalizedInfName {
			log.Debug().
				Str("service_name", inf.ServiceName).
				Str("matched_id", svc.ID).
				Msg("Found alternate name match")
			return svc
		}
	}

	// Try fuzzy matching
	threshold := 0.8
	var bestMatch *hsds_types.Service
	highestSimilarity := 0.0

	for _, svc := range services {
		similarity := calculateStringSimilarity(normalizedInfName, strings.ToLower(strings.TrimSpace(svc.Name)))
		if similarity > threshold && similarity > highestSimilarity {
			highestSimilarity = similarity
			bestMatch = svc
		}

		if svc.AlternateName != nil {
			altNameSimilarity := calculateStringSimilarity(normalizedInfName,
				strings.ToLower(strings.TrimSpace(*svc.AlternateName)))
			if altNameSimilarity > threshold && altNameSimilarity > highestSimilarity {
				highestSimilarity = altNameSimilarity
				bestMatch = svc
			}
		}
	}

	if bestMatch != nil {
		log.Debug().
			Str("service_name", inf.ServiceName).
			Str("matched_id", bestMatch.ID).
			Float64("similarity", highestSimilarity).
			Msg("Found fuzzy match")
	} else {
		log.Debug().
			Str("service_name", inf.ServiceName).
			Msg("No matching service found")
	}

	return bestMatch
}

// calculateStringSimilarity implements Levenshtein distance based similarity
func calculateStringSimilarity(s1, s2 string) float64 {
	d := levenshtein.ComputeDistance(s1, s2)
	maxLen := math.Max(float64(len(s1)), float64(len(s2)))
	if maxLen == 0 {
		return 0
	}
	return 1 - float64(d)/maxLen
}
