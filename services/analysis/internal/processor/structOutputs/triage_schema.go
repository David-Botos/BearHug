package structOutputs

import "github.com/david-botos/BearHug/services/analysis/internal/processor/inference"

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
					// string(SchedulingCategory),
					// string(ProgramCategory),
					// string(ReqDocsCategory),
					// string(ContactCategory),
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
