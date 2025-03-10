package structOutputs

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/david-botos/BearHug/services/analysis/internal/processor/inference"
	"github.com/rs/zerolog/log"
)

// TableName represents valid table names in the system
type TableName string

const (
	ServicesTable        TableName = "services"
	ServiceCapacityTable TableName = "service_capacity"
	UnitTable            TableName = "unit"
	ContactTable         TableName = "contact"
	PhoneTable           TableName = "phone"
	// ScheduleTable         TableName = "schedule"
	// ProgramTable          TableName = "program"
	// RequiredDocumentTable TableName = "required_document"
)

// TableDescription contains information about what data belongs in each table
type TableDescription struct {
	Name        TableName
	Description string
}

// Define descriptions for each table
/*
var tableDescriptions = []TableDescription{
	{
		Name:        ServiceCapacityTable,
		Description: "Defines capacity limits for services (e.g., number of available beds, quantity of monetary assistance awards available)",
	},
	{
		Name:        UnitTable,
		Description: "Defines custom units that give context to numerical values in service_capacity (e.g., beds, monetary awards)",
	},
	{
		Name:        ContactTable,
		Description: "Stores contact information for organization representatives",
	},
	{
		Name:        PhoneTable,
		Description: "Stores phone numbers for follow-up or additional information",
	},
	{
		Name:        ScheduleTable,
		Description: "Defines service timing including start/end times, duration, and frequency (daily/weekly/monthly)",
	},
	{
		Name:        ProgramTable,
		Description: "Groups related services under a common program (e.g., employment assistance program containing multiple related services)",
	},
	{
		Name:        RequiredDocumentTable,
		Description: "Lists required documentation for services (e.g., government ID)",
	},
}
*/

// Define descriptions for each category
var categoryDescriptions = []CategoryDescription{
	{
		Category:    CapacityCategory,
		Tables:      []TableName{ServiceCapacityTable, UnitTable},
		Description: "Information about service capacity limits (e.g., number of beds) and their associated units of measurement",
	},
	{
		Category:    ContactCategory,
		Tables:      []TableName{ContactTable, PhoneTable},
		Description: "Contact information for service representatives including phone numbers",
	},
	// {
	// 	Category:    SchedulingCategory,
	// 	Tables:      []TableName{ScheduleTable},
	// 	Description: "Service timing information including hours of operation, frequency, and duration",
	// },
	// {
	// 	Category:    ProgramCategory,
	// 	Tables:      []TableName{ProgramTable},
	// 	Description: "Organizational groupings of related services under a common program",
	// },
	// {
	// 	Category:    ReqDocsCategory,
	// 	Tables:      []TableName{RequiredDocumentTable},
	// 	Description: "Documentation requirements for service participation",
	// },
}

// GenerateTriagePrompt should output what tables are worth looking into filling based on the transcript
func GenerateTriagePrompt(transcript string) (string, inference.ToolInputSchema) {
	var categoryDescriptionStrings []string
	for _, desc := range categoryDescriptions {
		tableNames := make([]string, len(desc.Tables))
		for i, table := range desc.Tables {
			tableNames[i] = string(table)
		}
		categoryDescriptionStrings = append(categoryDescriptionStrings,
			fmt.Sprintf("%s (%s): %s",
				desc.Category,
				strings.Join(tableNames, ", "),
				desc.Description))
	}

	prompt := fmt.Sprintf(`Using the provided tool schema, analyze this transcript and output only a JSON object containing detected detail categories and their corresponding reasoning.

    Detail Categories:
    %s
    
    Transcript:
    %s
    
    Return a JSON object:
    {
        "detected_categories": string[],  // Categories that need population based on transcript
        "reasoning": string[]            // Index-matched explanations with transcript evidence
    }
    
    Guidelines:
    1. Only include categories with clear transcript evidence
    2. Use specific quotes/examples in reasoning
    3. Consider implicit references (e.g., hours mentioned → SCHEDULING category)
    4. For CAPACITY category, look for both the quantity AND its unit of measurement
    5. For CONTACT category, consider both contact names and associated phone numbers
    IMPORTANT: You must ONLY respond by using the triage_details tool to output the structured data. Do not provide any explanatory text, confirmations, or additional messages. Simply use the tool to output the structured data following the schema exactly.`,
		strings.Join(categoryDescriptionStrings, "\n"),
		transcript)
	return prompt, TriageDetailsTool
}

var TriageDetailsTool = inference.ToolInputSchema{
	Type: "object",
	Properties: map[string]inference.Property{
		"detected_categories": {
			Type:        "array",
			Description: "Array of detail categories detected in the transcript",
			Items: map[string]interface{}{
				"type": "string",
				"enum": []string{
					string(CapacityCategory),
					string(ContactCategory),
					// string(SchedulingCategory),
					// string(ProgramCategory),
					// string(ReqDocsCategory),
				},
				"description": "Valid detail category name",
			},
		},
		"reasoning": {
			Type:        "array",
			Description: "Array of explanations where each index maps directly to the category at the same index in detected_categories",
			Items: map[string]interface{}{
				"type":        "string",
				"description": "Explanation for why the corresponding category was selected, including specific evidence from the transcript",
			},
		},
	},
	Required: []string{"detected_categories", "reasoning"},
}

type IdentifiedDetails struct {
	DetectedCategories []string `json:"detected_categories"`
	Reasoning          []string `json:"reasoning"`
}

// TODO: no need for a pointer here
func IdentifyDetailsForTriagedAnalysis(transcript string) (*IdentifiedDetails, error) {
	log.Debug().Msg("Generating triage prompt and schema")
	detailTriagePrompt, detailTriageSchema := GenerateTriagePrompt(transcript)

	// Initialize Claude Inference Client
	client, err := inference.InitInferenceClient()
	if err != nil {
		log.Error().
			Err(err).
			Msg("Failed to initialize inference client")
		return nil, fmt.Errorf("failed to initialize inference client: %w", err)
	}

	// Run inference
	log.Debug().
		Int("prompt_length", len(detailTriagePrompt)).
		Bool("schema_present", len(detailTriageSchema.Properties) > 0).
		Msg("Running Claude inference for detail identification")

	serviceDetailsRes, serviceDetailsErr := client.RunClaudeInference(inference.PromptParams{
		Prompt: detailTriagePrompt,
		Schema: detailTriageSchema,
	})
	if serviceDetailsErr != nil {
		log.Error().
			Err(serviceDetailsErr).
			Msg("Claude inference failed during detail identification")
		return nil, fmt.Errorf("error with details identification: %w", serviceDetailsErr)
	}

	// Convert the map to JSON bytes first
	jsonBytes, err := json.Marshal(serviceDetailsRes)
	if err != nil {
		log.Error().
			Err(err).
			Msg("Failed to marshal response map to JSON")
		return nil, fmt.Errorf("failed to marshal response map: %w", err)
	}

	// Unmarshal JSON bytes into our struct
	var details IdentifiedDetails
	if err := json.Unmarshal(jsonBytes, &details); err != nil {
		log.Error().
			Err(err).
			Msg("Failed to unmarshal Claude inference response")
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &details, nil
}
