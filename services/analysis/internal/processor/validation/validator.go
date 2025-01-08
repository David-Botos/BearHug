package validation

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/david-botos/BearHug/services/analysis/internal/hsds_types"
	"github.com/david-botos/BearHug/services/analysis/internal/processor/inference"
	"github.com/david-botos/BearHug/services/analysis/internal/processor/structOutputs"
	"github.com/joho/godotenv"
)

// ValidationItem represents a single validation issue identified by the inference
type ValidationItem struct {
	Type       string   `json:"type"`        // "HALLUCINATION" or "DUPLICATE"
	ObjectType string   `json:"object_type"` // "SERVICE", "CAPACITY", or "UNIT"
	IDs        []string `json:"ids"`
	// Fields for Hallucinations
	IdentifiedSnippet   string  `json:"identified_snippet,omitempty"`
	Reasoning           string  `json:"reasoning,omitempty"`
	ConfidenceLevel     float64 `json:"confidence_level,omitempty"`
	SuggestedCorrection string  `json:"suggested_correction,omitempty"`
	// Fields for Duplicates
	Name              string   `json:"name,omitempty"`
	PreferredID       string   `json:"preferred_id,omitempty"`
	ConflictingFields []string `json:"conflicting_fields,omitempty"`
}

type ValidationOutput struct {
	Validation []ValidationItem `json:"validation"`
	IsValid    bool             `json:"is_valid"`
}

func ValidateExtractedInfo(extractedDetails []*structOutputs.DetailAnalysisResult, serviceCtx structOutputs.ServiceContext, transcript string) (bool, error) {

	// Extract the most valuable information to identify potential hallucinations and duplicates
	extractedDetailStrings := buildValidationString(extractedDetails, serviceCtx)

	// Create a prompt and schema for looping validation runs
	validationPrompt, validationSchema, validationPromptGenErr := generateValidationPrompt(extractedDetailStrings, transcript)
	if validationPromptGenErr != nil {
		return false, fmt.Errorf(`Error when generating the validation prompt: %w`, validationPromptGenErr)
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
	validationOutput, validationInfErr := client.RunClaudeInference(inference.PromptParams{Prompt: validationPrompt, Schema: validationSchema})
	if validationInfErr != nil {
		return false, fmt.Errorf(`Error when running validation inference: %w`, validationInfErr)
	}

	// First, get the raw JSON bytes from the inference output
	jsonData, err := json.Marshal(validationOutput)
	if err != nil {
		return false, fmt.Errorf("error marshaling validation output: %w", err)
	}

	// Create our ValidationOutput struct
	var typedOutput ValidationOutput

	// Unmarshal the JSON into our struct
	if err := json.Unmarshal(jsonData, &typedOutput); err != nil {
		return false, fmt.Errorf("error unmarshaling validation output: %w", err)
	}

	if typedOutput.IsValid {
		// submit the information to supabase
		return true, nil
	} else {
		/*
			Create a while loop with an iterator that defines a maximum number of reasoning loops before it needs to flag the result for human review

			do reasoning and output ValidationResult

			if IsValid == false {

			}
			break
		*/

		/*
			if !validationResult.IsValid {
				flag the result for human review
				return false, nil
			} else {
				submit the info to supabase
				return true, nil
			}
		*/
		return false, nil
	}
}

func buildValidationString(extractedDetails []*structOutputs.DetailAnalysisResult, serviceCtx structOutputs.ServiceContext) string {
	var builder strings.Builder

	// First, let's build a map of units by ID for easy reference
	unitMap := make(map[string]*hsds_types.Unit)

	// Find and map all units from extracted details
	for _, detail := range extractedDetails {
		if detail.CapacityData != nil {
			for _, unit := range detail.CapacityData.Units {
				unitMap[unit.ID] = unit
			}
		}
	}

	// Create a map of services that have new capacity information
	servicesWithNewInfo := make(map[string]bool)
	for _, detail := range extractedDetails {
		if detail.CapacityData != nil {
			for _, capacity := range detail.CapacityData.Capacities {
				servicesWithNewInfo[capacity.ServiceID] = true
			}
		}
	}

	// Start with existing services that have new information
	builder.WriteString("=== Existing Services with New Information ===\n")
	for _, service := range serviceCtx.ExistingServices {
		// Skip if no new information
		if !servicesWithNewInfo[service.ID] {
			continue
		}

		builder.WriteString(fmt.Sprintf("Service: %s (ID: %s)\n", service.Name, service.ID))
		if service.Description != nil {
			builder.WriteString(fmt.Sprintf("Existing Description: %s\n", *service.Description))
		}

		// Add new capacity information for this service
		builder.WriteString("New Information Extracted from Transcript:\n")
		for _, detail := range extractedDetails {
			if detail.CapacityData != nil {
				for _, capacity := range detail.CapacityData.Capacities {
					if capacity.ServiceID == service.ID {
						unit := unitMap[capacity.UnitID]
						builder.WriteString("- Capacity: ")
						builder.WriteString(fmt.Sprintf("%.2f", capacity.Available))
						if capacity.Maximum != nil {
							builder.WriteString(fmt.Sprintf(" (Maximum: %.2f)", *capacity.Maximum))
						}
						if unit != nil {
							builder.WriteString(fmt.Sprintf(" %s", unit.Name))
						}
						if capacity.Description != nil {
							builder.WriteString(fmt.Sprintf(" - %s", *capacity.Description))
						}
						builder.WriteString("\n")
					}
				}
			}
		}
		builder.WriteString("\n")
	}

	// Add new services information
	builder.WriteString("=== Newly Mentioned Services ===\n")
	for _, service := range serviceCtx.NewServices {
		builder.WriteString(fmt.Sprintf("Service: %s (ID: %s)\n", service.Name, service.ID))
		if service.Description != nil {
			builder.WriteString(fmt.Sprintf("Description: %s\n", *service.Description))
		}

		// Add capacity information for this service
		builder.WriteString("Extracted Information:\n")
		for _, detail := range extractedDetails {
			if detail.CapacityData != nil {
				for _, capacity := range detail.CapacityData.Capacities {
					if capacity.ServiceID == service.ID {
						unit := unitMap[capacity.UnitID]
						builder.WriteString("- Capacity: ")
						builder.WriteString(fmt.Sprintf("%.2f", capacity.Available))
						if capacity.Maximum != nil {
							builder.WriteString(fmt.Sprintf(" (Maximum: %.2f)", *capacity.Maximum))
						}
						if unit != nil {
							builder.WriteString(fmt.Sprintf(" %s", unit.Name))
						}
						if capacity.Description != nil {
							builder.WriteString(fmt.Sprintf(" - %s", *capacity.Description))
						}
						builder.WriteString("\n")
					}
				}
			}
		}
		builder.WriteString("\n")
	}

	return builder.String()
}
