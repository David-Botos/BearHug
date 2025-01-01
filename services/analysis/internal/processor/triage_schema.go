package processor

func NewTriageSchema() ToolInputSchema {
	return ToolInputSchema{
		Type: "object",
		Properties: map[string]Property{
			"detected_tables": {
				Type:        "array",
				Description: "Valid table names in the system",
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
					"description": "Valid table names in the system",
				},
			},
			"reasoning": {
				Type:        "string",
				Description: "Explanation of why each table was selected",
			},
		},
		Required: []string{"detected_tables", "reasoning"},
	}
}
