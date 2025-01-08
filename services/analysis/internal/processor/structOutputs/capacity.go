package structOutputs

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/agnivade/levenshtein"
	"github.com/david-botos/BearHug/services/analysis/internal/hsds_types"
	"github.com/david-botos/BearHug/services/analysis/internal/processor/inference"
	"github.com/joho/godotenv"
)

func GenerateServiceCapacityPrompt(transcript string, serviceCtx ServiceContext) (string, inference.ToolInputSchema, error) {

	// Build service descriptions
	var existingServiceDesc, newServiceDesc strings.Builder

	// Process existing services
	for _, service := range serviceCtx.ExistingServices {
		writeServiceDescription(&existingServiceDesc, service)
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

	// Generate Prompt and Schema
	capacityCategoryPrompt, capacitySchema, err := GenerateServiceCapacityPrompt(transcript, serviceCtx)
	if err != nil {
		return DetailAnalysisResult{}, fmt.Errorf(`Failure when generating service capacity prompt: %w`, err)
	}

	// Declare Claude Inference Client
	workingDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	envPath := filepath.Join(workingDir, ".env")
	if err := godotenv.Load(envPath); err != nil {
		panic(err)
	}
	fmt.Printf("envPath declared as: %s\n", envPath)
	client := inference.NewClient(os.Getenv("ANTHROPIC_API_KEY"))
	fmt.Printf("Initialized client with API key length: %d\n", len(os.Getenv("ANTHROPIC_API_KEY")))

	// Run inference
	unformattedCapacityDetails, inferenceErr := client.RunClaudeInference(inference.PromptParams{Prompt: capacityCategoryPrompt, Schema: capacitySchema})
	if inferenceErr != nil {
		return DetailAnalysisResult{}, fmt.Errorf(`Error running inference to extract capacity details: %w`, inferenceErr)
	}

	capacityDetails, unitDetails, capacityAndUnitInfConvErr := infToCapacityAndUnits(unformattedCapacityDetails, serviceCtx)
	if capacityAndUnitInfConvErr != nil {
		return DetailAnalysisResult{}, fmt.Errorf(`Error while converting the inference response to clean capacity and unit objects: %w`, capacityAndUnitInfConvErr)
	}

	var result DetailAnalysisResult = *NewCapacityResult(capacityDetails, unitDetails)

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
	// Convert the inference result to our structured type
	jsonData, err := json.Marshal(inferenceResult)
	if err != nil {
		return nil, nil, fmt.Errorf("error marshaling inference result: %w", err)
	}

	var output capacityAndUnitInfOutput
	if err := json.Unmarshal(jsonData, &output); err != nil {
		return nil, nil, fmt.Errorf("error unmarshaling to structured output: %w", err)
	}

	// Combine existing and new services into a single array
	totalServices := make([]hsds_types.Service, 0, len(serviceCtx.ExistingServices)+len(serviceCtx.NewServices))
	totalServices = append(totalServices, serviceCtx.ExistingServices...)
	for _, newService := range serviceCtx.NewServices {
		totalServices = append(totalServices, *newService)
	}

	// Track matching results
	matchResults := make([]serviceMatchResult, len(output.Capacities))
	for i, capacity := range output.Capacities {
		matchedService := findMatchingService(capacity, totalServices)
		matchResults[i] = serviceMatchResult{
			inference: capacity,
			service:   matchedService,
			matched:   matchedService != nil,
		}

		// TODO: is this right?
	}

	// Check if we have any unmatched capacities
	var unmatched []string
	for _, result := range matchResults {
		if !result.matched {
			unmatched = append(unmatched, result.inference.ServiceName)
		}
	}
	if len(unmatched) > 0 {
		// TODO: add inference to match / make sense of output
		return nil, nil, fmt.Errorf("unable to match services for: %s", strings.Join(unmatched, ", "))
	}

	// Create arrays to hold our results
	units := make([]*hsds_types.Unit, 0, len(matchResults))
	capacities := make([]*hsds_types.ServiceCapacity, 0, len(matchResults))

	// Process each matched capacity entry
	for _, match := range matchResults {
		// Create the Unit
		unitOpts := &hsds_types.UnitOptions{}

		// TODO: optionally add more unit detail extraction to the inference about schemas

		unit, err := hsds_types.NewUnit(match.inference.UnitName, unitOpts)
		if err != nil {
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
			return nil, nil, fmt.Errorf("error creating service capacity for unit %s: %w", unit.ID, err)
		}
		capacities = append(capacities, serviceCapacity)
	}

	return capacities, units, nil
}

// findMatchingService attempts to find the corresponding service for a capacity
func findMatchingService(inf capacityInference, services []hsds_types.Service) *hsds_types.Service {
	normalizedInfName := strings.ToLower(strings.TrimSpace(inf.ServiceName))

	// First try exact name match
	for _, svc := range services {
		if strings.ToLower(strings.TrimSpace(svc.Name)) == normalizedInfName {
			return &svc
		}
	}

	// Try alternate name if no exact match found
	for _, svc := range services {
		if svc.AlternateName != nil &&
			strings.ToLower(strings.TrimSpace(*svc.AlternateName)) == normalizedInfName {
			return &svc
		}
	}

	// If still no match, try fuzzy matching with a threshold
	threshold := 0.8 // 80% similarity threshold
	var bestMatch *hsds_types.Service
	highestSimilarity := 0.0

	for _, svc := range services {
		similarity := calculateStringSimilarity(normalizedInfName, strings.ToLower(strings.TrimSpace(svc.Name)))
		if similarity > threshold && similarity > highestSimilarity {
			highestSimilarity = similarity
			bestMatch = &svc
		}

		if svc.AlternateName != nil {
			altNameSimilarity := calculateStringSimilarity(normalizedInfName,
				strings.ToLower(strings.TrimSpace(*svc.AlternateName)))
			if altNameSimilarity > threshold && altNameSimilarity > highestSimilarity {
				highestSimilarity = altNameSimilarity
				bestMatch = &svc
			}
		}
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
