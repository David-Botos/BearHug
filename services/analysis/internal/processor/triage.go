package processor

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

// GenerateTriagePrompt should output what tables are worth looking into filling based on the
func GenerateTriagePrompt(transcript string) string {
	var validTablesList []string
	for tableName := range validTableNames {
		validTablesList = append(validTablesList, string(tableName))
	}

	prompt := fmt.Sprintf(`Analyze the following transcript between an AI agent and a community organization representative.
Determine which database tables may need to be populated based on the information discussed.

Valid table names are: %s

Only use valid table names when defining detected_tables in your output. For each table you identify as potentially relevant, explain why the information in the transcript suggests this table might need to be populated.

Transcript:
%s

Format your response as a JSON object with the following schema:
{
    "detected_tables": string[],  // Array of table names that may need to be populated
    "reasoning": string          // Explanation of why each table was selected
}

Remember:
1. Only include tables from the provided list
2. Only include a table if there is clear evidence in the transcript that information for that table was discussed
3. Provide specific examples from the transcript in your reasoning
4. Consider implicit references to information (e.g., if hours of operation are mentioned, that implies schedule table)`,
		strings.Join(validTablesList, ", "),
		transcript)

	return prompt
}

