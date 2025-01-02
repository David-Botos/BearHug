package processor

func NewTriageSchema() ToolInputSchema {
	return ToolInputSchema{
		Type: "object",
		Properties: map[string]Property{
			"detected_tables": {
				Type:        "array",
				Description: "Array of valid table names in the system",
				Items: map[string]interface{}{
					"type": "string",
					"enum": []string{
						string(ServicesTable),
						string(ServiceCapacityTable),
						string(UnitTable),
						string(ScheduleTable),
						string(ProgramTable),
						string(RequiredDocumentTable),
						string(ContactTable),
						string(PhoneTable),
					},
					"description": "Valid table name",
				},
			},
			"reasoning": {
				Type:        "array",
				Description: "Array of explanations where each index maps directly to the table at the same index in detected_tables",
				Items: map[string]interface{}{
					"type":        "string",
					"description": "Explanation for why the corresponding table in detected_tables was selected, including specific evidence from the transcript",
				},
			},
		},
		Required: []string{"detected_tables", "reasoning"},
	}
}
