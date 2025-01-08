package structOutputs

import (
	"fmt"
	"strings"

	"github.com/david-botos/BearHug/services/analysis/internal/processor/inference"
)

// TableName represents valid table names in the system
type TableName string

const (
	ServicesTable         TableName = "services"
	ServiceCapacityTable  TableName = "service_capacity"
	UnitTable             TableName = "unit"
	ScheduleTable         TableName = "schedule"
	ProgramTable          TableName = "program"
	RequiredDocumentTable TableName = "required_document"
	ContactTable          TableName = "contact"
	PhoneTable            TableName = "phone"
)

// TableDescription contains information about what data belongs in each table
type TableDescription struct {
	Name        TableName
	Description string
}

// Define descriptions for each table
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
	{
		Name:        ContactTable,
		Description: "Stores contact information for organization representatives",
	},
	{
		Name:        PhoneTable,
		Description: "Stores phone numbers for follow-up or additional information",
	},
}

// Define descriptions for each category
var categoryDescriptions = []CategoryDescription{
	{
		Category:    CapacityCategory,
		Tables:      []TableName{ServiceCapacityTable, UnitTable},
		Description: "Information about service capacity limits (e.g., number of beds) and their associated units of measurement",
	},
	{
		Category:    SchedulingCategory,
		Tables:      []TableName{ScheduleTable},
		Description: "Service timing information including hours of operation, frequency, and duration",
	},
	{
		Category:    ProgramCategory,
		Tables:      []TableName{ProgramTable},
		Description: "Organizational groupings of related services under a common program",
	},
	{
		Category:    ReqDocsCategory,
		Tables:      []TableName{RequiredDocumentTable},
		Description: "Documentation requirements for service participation",
	},
	{
		Category:    ContactCategory,
		Tables:      []TableName{ContactTable, PhoneTable},
		Description: "Contact information for service representatives including phone numbers",
	},
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
