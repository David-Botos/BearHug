package validation

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/david-botos/BearHug/services/analysis/internal/hsds_types"
	"github.com/david-botos/BearHug/services/analysis/internal/processor/inference"
	"github.com/david-botos/BearHug/services/analysis/internal/processor/structOutputs"
	"github.com/joho/godotenv"
)

type ValidationItemType string

const (
	HallucinationFlag ValidationItemType = "HALLUCINATION"
	DuplicateFlag     ValidationItemType = "DUPLICATE"
)

// ValidationItem represents a single validation issue identified by the inference
type ValidationItem struct {
	Type       ValidationItemType `json:"type"`        // "HALLUCINATION" or "DUPLICATE"
	ObjectType DetailObjectType   `json:"object_type"` // "SERVICE", "CAPACITY", or "UNIT"
	IDs        []string           `json:"ids"`
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

type FixAttempt struct {
	ValidationItems []ValidationItem
	FixedItems      []ValidationItem
	Successful      bool
	Iteration       int
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
		submitValidatedOutputRes, submitValidatedOutputErr := SubmitValidatedOutput(extractedDetails)
		if submitValidatedOutputErr != nil {
			return false, fmt.Errorf(`Error occurred when submitting validated output in supa: %w`, &submitValidatedOutputErr)
		}
		return submitValidatedOutputRes, nil
	} else {
		var fixAttempts []FixAttempt
		maxIterations := 3
		currentIteration := 0
		currentDetails := extractedDetails
		currentServiceCtx := serviceCtx

		for currentIteration < maxIterations {
			attempt := FixAttempt{
				ValidationItems: typedOutput.Validation,
				Iteration:       currentIteration,
			}

			sortedIssues := prioritizeIssues(typedOutput.Validation)

			fixedDetails, fixedServiceCtx, fixErr := fixOutputWithInference(
				currentDetails,
				currentServiceCtx,
				sortedIssues,
				transcript,
				client,
			)
			if fixErr != nil {
				return false, fmt.Errorf("error during fix attempt %d: %w", currentIteration, fixErr)
			}

			// Validate the fixed output
			fixedDetailString := buildValidationString(fixedDetails, fixedServiceCtx)
			validationPrompt, validationSchema, err := generateValidationPrompt(fixedDetailString, transcript)
			if err != nil {
				return false, fmt.Errorf("error generating validation prompt after fix: %w", err)
			}
			// Run validation on fixed output
			newValidationOutput, valErr := client.RunClaudeInference(inference.PromptParams{
				Prompt: validationPrompt,
				Schema: validationSchema,
			})
			if valErr != nil {
				return false, fmt.Errorf("error validating fixed output: %w", valErr)
			}

			// Parse validation results
			var newTypedOutput ValidationOutput
			newJsonData, _ := json.Marshal(newValidationOutput)
			if err := json.Unmarshal(newJsonData, &newTypedOutput); err != nil {
				return false, fmt.Errorf("error parsing validation after fix: %w", err)
			}

			// Check if we've improved
			if newTypedOutput.IsValid {
				submitValidatedOutputRes, submitValidatedOutputErr := SubmitValidatedOutput(fixedDetails)
				if submitValidatedOutputErr != nil {
					return false, fmt.Errorf(`Error occurred when submitting validated output in supa: %w`, submitValidatedOutputErr)
				}
				return submitValidatedOutputRes, nil
			}

			// Check if we've made progress
			if len(newTypedOutput.Validation) >= len(typedOutput.Validation) {
				// We haven't improved or have made things worse
				// TODO: Break the loop and flag for human review
				break
			}

			// Update for next iteration
			currentDetails = fixedDetails
			currentServiceCtx = fixedServiceCtx
			typedOutput = newTypedOutput
			currentIteration++

			// Store attempt results
			attempt.FixedItems = newTypedOutput.Validation
			attempt.Successful = false
			fixAttempts = append(fixAttempts, attempt)
		}
		// If we get here, we've exceeded max iterations or haven't improved
		// Flag for human review
		// TODO: Store fixAttempts history for human review
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

// prioritizeIssues sorts validation items by confidence level and complexity
func prioritizeIssues(items []ValidationItem) []ValidationItem {
	// Create a copy to sort
	sortedItems := make([]ValidationItem, len(items))
	copy(sortedItems, items)

	// Sort by confidence level for hallucinations and by number of conflicting fields for duplicates
	sort.Slice(sortedItems, func(i, j int) bool {
		if sortedItems[i].Type == "HALLUCINATION" && sortedItems[j].Type == "HALLUCINATION" {
			return sortedItems[i].ConfidenceLevel > sortedItems[j].ConfidenceLevel
		}
		if sortedItems[i].Type == "DUPLICATE" && sortedItems[j].Type == "DUPLICATE" {
			return len(sortedItems[i].ConflictingFields) < len(sortedItems[j].ConflictingFields)
		}
		// Prioritize hallucinations over duplicates
		return sortedItems[i].Type == "HALLUCINATION"
	})

	return sortedItems
}

// fixOutputWithInference attempts to fix validation issues using Claude inference
func fixOutputWithInference(
	details []*structOutputs.DetailAnalysisResult,
	serviceCtx structOutputs.ServiceContext,
	issues []ValidationItem,
	transcript string,
	client *inference.ClaudeClient,
) ([]*structOutputs.DetailAnalysisResult, structOutputs.ServiceContext, error) {
	// 1. Build shared context maps
	contextMaps := buildContextMaps(details, serviceCtx, issues)

	// 2. Build context strings for the prompt (now using shared maps)
	currentState, issuesSummary := buildFixContext(details, serviceCtx, issues, contextMaps)

	// 3. Generate and run the fix prompt
	fixPrompt, fixSchema, err := generateFixPrompt(currentState, issuesSummary, transcript)
	if err != nil {
		return nil, structOutputs.ServiceContext{}, fmt.Errorf("error generating fix prompt: %w", err)
	}

	// 4. Get fixes from Claude
	fixOutput, err := client.RunClaudeInference(inference.PromptParams{
		Prompt: fixPrompt,
		Schema: fixSchema,
	})
	if err != nil {
		return nil, structOutputs.ServiceContext{}, fmt.Errorf("error running fix inference: %w", err)
	}

	// 5. Parse the fix output
	jsonData, err := json.Marshal(fixOutput)
	if err != nil {
		return nil, structOutputs.ServiceContext{}, fmt.Errorf("error marshaling inference result: %w", err)
	}
	var fixes FixOutput
	if err := json.Unmarshal(jsonData, &fixes); err != nil {
		return nil, structOutputs.ServiceContext{}, fmt.Errorf("error parsing fix output: %w", err)
	}

	// 6. Get maps of affected items using shared context
	affectedServices, affectedCapacities := getAffectedItems(contextMaps)

	// 7. Apply the fixes using our lookup maps
	newDetails, newServiceCtx, err := applyFixes(fixes, affectedServices, affectedCapacities, details, serviceCtx)
	if err != nil {
		return nil, structOutputs.ServiceContext{}, fmt.Errorf("error applying fixes: %w", err)
	}

	return newDetails, newServiceCtx, nil
}
