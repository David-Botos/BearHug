package validation

import (
	"fmt"

	"github.com/david-botos/BearHug/services/analysis/internal/processor/inference"
)

func generateValidationPrompt(extractedDetails string, transcript string) (string, inference.ToolInputSchema, error) {
	prompt := fmt.Sprintf(`You are validating information extracted from a transcript about service capacity information. Your task is to identify potential errors in two categories:

1. Hallucinations: Information that appears in our extraction but isn't supported by the transcript
2. Duplicates: Services or capacities that appear to be referring to the same thing

Here is the transcript from which information was extracted:
---
%s
---

Here is the information that was extracted:
---
%s
---

For each potential HALLUCINATION you identify, provide:
- The specific IDs and object type involved (SERVICE, CAPACITY, or UNIT)
- The problematic snippet of extracted information
- Your explanation of why this appears to be a hallucination
- Your confidence level (0.0-1.0) that this is a hallucination
- A suggested correction if possible

For each potential DUPLICATE you identify, provide:
- The IDs involved
- The name of the duplicated entity
- Which ID should be preferred
- Which fields have conflicting values between the duplicates

If you find no issues, indicate that the information appears valid.

Respond in the following JSON format:
{
    "validation": [
        {
            "type": "HALLUCINATION",
            "object_type": "SERVICE|CAPACITY|UNIT",
            "ids": ["id1"],
            "identified_snippet": "the problematic text",
            "reasoning": "detailed explanation",
            "confidence_level": 0.95,
            "suggested_correction": "optional correction"
        },
        {
            "type": "DUPLICATE",
            "object_type": "SERVICE|CAPACITY|UNIT",
            "ids": ["id1", "id2"],
            "name": "name of duplicated entity",
            "preferred_id": "id1",
            "conflicting_fields": ["field1", "field2"]
        }
    ],
    "is_valid": false
}`, transcript, extractedDetails)

	return prompt, ValidationSchema, nil
}

type ValidationResult struct {
	IsValid                 bool
	PotentialHallucinations []Hallucination
	PotentialDuplicates     []Duplicate
}

type DetailObjectType string

const (
	ServiceObj         DetailObjectType = "SERVICE"
	ServiceCapacityObj DetailObjectType = "CAPACITY"
	CapacityUnitObj    DetailObjectType = "UNIT"
	// TODO: Add more as more details are implemented
)

type Hallucination struct {
	DetailObjectType    DetailObjectType
	IdentifiedSnippet   string
	IdentifiedReasoning string
	SuggestedCorrection string
	ConfidenceLevel     float64
}

type Duplicate struct {
	Type              DetailObjectType
	Ids               []string
	Name              string
	PreferredID       string
	ConflictingFields []string
}

var validationItemSchema = map[string]interface{}{
	"type": "object",
	"properties": map[string]interface{}{
		"type": map[string]interface{}{
			"type": "string",
			"enum": []string{"HALLUCINATION", "DUPLICATE"},
		},
		"object_type": map[string]interface{}{
			"type": "string",
			"enum": []string{"SERVICE", "CAPACITY", "UNIT"},
		},
		"ids": map[string]interface{}{
			"type": "array",
			"items": map[string]interface{}{
				"type": "string",
			},
		},
		"identified_snippet": map[string]interface{}{
			"type": "string",
		},
		"reasoning": map[string]interface{}{
			"type": "string",
		},
		"confidence_level": map[string]interface{}{
			"type":    "number",
			"minimum": 0,
			"maximum": 1,
		},
		"suggested_correction": map[string]interface{}{
			"type": "string",
		},
		"name": map[string]interface{}{
			"type": "string",
		},
		"preferred_id": map[string]interface{}{
			"type": "string",
		},
		"conflicting_fields": map[string]interface{}{
			"type": "array",
			"items": map[string]interface{}{
				"type": "string",
			},
		},
	},
	"required": []string{
		"type",
		"object_type",
		"ids",
	},
	"if": map[string]interface{}{
		"properties": map[string]interface{}{
			"type": map[string]interface{}{
				"const": "HALLUCINATION",
			},
		},
	},
	"then": map[string]interface{}{
		"required": []string{
			"identified_snippet",
			"reasoning",
			"confidence_level",
		},
	},
	"else": map[string]interface{}{
		"required": []string{
			"name",
			"preferred_id",
			"conflicting_fields",
		},
	},
}

var ValidationSchema = inference.ToolInputSchema{
	Type: "object",
	Properties: map[string]inference.Property{
		"validation": {
			Type:  "array",
			Items: validationItemSchema,
		},
		"is_valid": {
			Type: "boolean",
		},
	},
	Required: []string{"validation", "is_valid"},
}
