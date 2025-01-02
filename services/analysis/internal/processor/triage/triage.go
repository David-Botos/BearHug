package triage

import (
	"fmt"
	"strings"
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
		Name:        ServicesTable,
		Description: "Contains individual services offered to the community, including service name, description, active status, application process, and eligibility specifications",
	},
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

// Define a map of valid table names for validation
var validTableNames = map[TableName]bool{
	ServicesTable:         true,
	ServiceCapacityTable:  true,
	UnitTable:             true,
	ScheduleTable:         true,
	ProgramTable:          true,
	RequiredDocumentTable: true,
	ContactTable:          true,
	PhoneTable:            true,
}

// GenerateTriagePrompt should output what tables are worth looking into filling based on the transcript
func GenerateTriagePrompt(transcript string) string {
	var tableDescriptionStrings []string
	for _, desc := range tableDescriptions {
		tableDescriptionStrings = append(tableDescriptionStrings,
			fmt.Sprintf("%s: %s", desc.Name, desc.Description))
	}
	prompt := fmt.Sprintf(`Using the provided tool schema, analyze this transcript and output only a JSON object containing detected tables and their corresponding reasoning.

	Tables:
	%s
	
	Transcript:
	%s
	
	Return a JSON object:
	{
		"detected_tables": string[],  // Tables that need population based on transcript
		"reasoning": string[]         // Index-matched explanations with transcript evidence
	}
	
	Guidelines:
	1. Only include tables with clear transcript evidence
	2. Use specific quotes/examples in reasoning
	3. Consider implicit references (e.g., hours mentioned → schedule table)`,
		strings.Join(tableDescriptionStrings, "\n"),
		transcript)
	return prompt
}
