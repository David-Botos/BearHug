package validation

import (
	"fmt"
	"strings"

	"github.com/david-botos/BearHug/services/analysis/internal/hsds_types"
	"github.com/david-botos/BearHug/services/analysis/internal/processor/inference"
	"github.com/david-botos/BearHug/services/analysis/internal/processor/structOutputs"
)

func generateFixPrompt(currentState, issuesSummary, transcript string) (string, inference.ToolInputSchema, error) {
	const promptTemplate = `You are helping to fix issues identified in service capacity information extracted from a transcript. You will be given the current state of the data, identified issues, and the original transcript. Your task is to provide specific fixes for each issue.

Original Transcript:
---
%s
---

Current State of Extracted Data:
---
%s
---

Identified Issues to Fix:
---
%s
---

For each issue, provide:
1. For hallucinations:
   - The correct information based on the transcript
   - If no correct information exists in the transcript, specify that the data should be removed
   - For partial hallucinations (where some aspects are correct), specify which parts to keep

2. For duplicates:
   - Which record should be kept (using the preferred_id)
   - What values should be used for any conflicting fields
   - Whether any capacity information should be combined

Provide your response in the following JSON format:
{
    "fixes": [
        {
            "issue_type": "HALLUCINATION",
            "object_ids": ["id1"],
            "action": "MODIFY|REMOVE",
            "modification": {
                "field": "name of field",
                "new_value": "corrected value"
            }
        },
        {
            "issue_type": "DUPLICATE",
            "object_ids": ["id1", "id2"],
            "action": "MERGE",
            "keep_id": "id1",
            "field_resolutions": [
                {
                    "field": "name of field",
                    "value": "resolved value"
                }
            ]
        }
    ]
}`

	// Define the schema for the fix output
	fixSchema := inference.ToolInputSchema{
		Type: "object",
		Properties: map[string]inference.Property{
			"fixes": {
				Type: "array",
				Items: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"issue_type": map[string]interface{}{
							"type": "string",
							"enum": []string{"HALLUCINATION", "DUPLICATE"},
						},
						"object_ids": map[string]interface{}{
							"type": "array",
							"items": map[string]interface{}{
								"type": "string",
							},
						},
						"action": map[string]interface{}{
							"type": "string",
							"enum": []string{"MODIFY", "REMOVE", "MERGE"},
						},
						"modification": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"field": map[string]interface{}{
									"type": "string",
								},
								"new_value": map[string]interface{}{
									"type": "string",
								},
							},
							"required": []string{"field", "new_value"},
						},
						"keep_id": map[string]interface{}{
							"type": "string",
						},
						"field_resolutions": map[string]interface{}{
							"type": "array",
							"items": map[string]interface{}{
								"type": "object",
								"properties": map[string]interface{}{
									"field": map[string]interface{}{
										"type": "string",
									},
									"value": map[string]interface{}{
										"type": "string",
									},
								},
								"required": []string{"field", "value"},
							},
						},
					},
					"required": []string{"issue_type", "object_ids", "action"},
				},
			},
		},
		Required: []string{"fixes"},
	}

	// Format the prompt with our context strings
	prompt := fmt.Sprintf(promptTemplate, transcript, currentState, issuesSummary)
	return prompt, fixSchema, nil
}

// FixOutput represents the structured response from the fix inference
type FixOutput struct {
	Fixes []Fix `json:"fixes"`
}

type Fix struct {
	IssueType string   `json:"issue_type"`
	ObjectIDs []string `json:"object_ids"`
	Action    string   `json:"action"`

	// For hallucinations
	Modification *struct {
		Field    string `json:"field"`
		NewValue string `json:"new_value"`
	} `json:"modification,omitempty"`

	// For duplicates
	KeepID           string `json:"keep_id,omitempty"`
	FieldResolutions []struct {
		Field string `json:"field"`
		Value string `json:"value"`
	} `json:"field_resolutions,omitempty"`
}

// ContextMaps holds lookup maps for quick reference to services, units, and capacities
type ContextMaps struct {
	ServiceMap   map[string]*hsds_types.Service
	UnitMap      map[string]*hsds_types.Unit
	CapacityMap  map[string]*hsds_types.ServiceCapacity
	AffectedSvcs map[string]bool
	AffectedCaps map[string]bool
}

// buildContextMaps creates lookup maps for all objects and marks which ones are affected by issues
func buildContextMaps(
	details []*structOutputs.DetailAnalysisResult,
	serviceCtx structOutputs.ServiceContext,
	issues []ValidationItem,
) *ContextMaps {
	ctx := &ContextMaps{
		ServiceMap:   make(map[string]*hsds_types.Service),
		UnitMap:      make(map[string]*hsds_types.Unit),
		CapacityMap:  make(map[string]*hsds_types.ServiceCapacity),
		AffectedSvcs: make(map[string]bool),
		AffectedCaps: make(map[string]bool),
	}

	// Map existing services
	for i := range serviceCtx.ExistingServices {
		ctx.ServiceMap[serviceCtx.ExistingServices[i].ID] = serviceCtx.ExistingServices[i]
	}
	// Map new services
	for _, service := range serviceCtx.NewServices {
		ctx.ServiceMap[service.ID] = service
	}

	// Map units and capacities
	for _, detail := range details {
		if detail.CapacityData != nil {
			for _, unit := range detail.CapacityData.Units {
				ctx.UnitMap[unit.ID] = unit
			}
			for _, capacity := range detail.CapacityData.Capacities {
				ctx.CapacityMap[capacity.ID] = capacity
			}
		}
	}

	// Mark affected items from issues
	for _, issue := range issues {
		for _, id := range issue.IDs {
			switch issue.ObjectType {
			case "SERVICE":
				ctx.AffectedSvcs[id] = true
				// Also mark capacities belonging to affected services
				for capID, cap := range ctx.CapacityMap {
					if cap.ServiceID == id {
						ctx.AffectedCaps[capID] = true
					}
				}
			case "CAPACITY":
				ctx.AffectedCaps[id] = true
			}
		}
	}

	return ctx
}

// buildFixContext creates structured context strings for the fix inference using shared context maps
func buildFixContext(
	details []*structOutputs.DetailAnalysisResult,
	serviceCtx structOutputs.ServiceContext,
	issues []ValidationItem,
	ctx *ContextMaps,
) (currentStateContext, issuesContext string) {
	var currentState, issuesSummary strings.Builder

	// Build current state context
	currentState.WriteString("=== Current State ===\n")

	// Add services and their capacities
	for serviceID, service := range ctx.ServiceMap {
		// Highlight affected services
		prefix := ""
		if ctx.AffectedSvcs[serviceID] {
			prefix = "* " // Mark affected services with an asterisk
		}

		currentState.WriteString(fmt.Sprintf("%sService: %s (ID: %s)\n", prefix, service.Name, serviceID))
		if service.Description != nil {
			currentState.WriteString(fmt.Sprintf("Description: %s\n", *service.Description))
		}

		// Add capacities for this service
		for capID, capacity := range ctx.CapacityMap {
			if capacity.ServiceID == serviceID {
				unit := ctx.UnitMap[capacity.UnitID]
				prefix := "- "
				if ctx.AffectedCaps[capID] {
					prefix = "* - " // Mark affected capacities with an asterisk
				}

				currentState.WriteString(prefix + "Capacity: ")
				currentState.WriteString(fmt.Sprintf("%.2f", capacity.Available))
				if capacity.Maximum != nil {
					currentState.WriteString(fmt.Sprintf(" (Maximum: %.2f)", *capacity.Maximum))
				}
				if unit != nil {
					currentState.WriteString(fmt.Sprintf(" %s", unit.Name))
				}
				if capacity.Description != nil {
					currentState.WriteString(fmt.Sprintf(" - %s", *capacity.Description))
				}
				currentState.WriteString("\n")
			}
		}
		currentState.WriteString("\n")
	}

	// Build issues context (same as before)
	issuesSummary.WriteString("=== Identified Issues ===\n")

	var hallucinations, duplicates []ValidationItem
	for _, issue := range issues {
		switch issue.Type {
		case "HALLUCINATION":
			hallucinations = append(hallucinations, issue)
		case "DUPLICATE":
			duplicates = append(duplicates, issue)
		}
	}

	// Add hallucinations
	if len(hallucinations) > 0 {
		issuesSummary.WriteString("\nHallucinations:\n")
		for _, h := range hallucinations {
			issuesSummary.WriteString(fmt.Sprintf("- Found in: %s (%s)\n", h.IdentifiedSnippet, h.ObjectType))
			issuesSummary.WriteString(fmt.Sprintf("  Reasoning: %s\n", h.Reasoning))
			if h.SuggestedCorrection != "" {
				issuesSummary.WriteString(fmt.Sprintf("  Suggested Correction: %s\n", h.SuggestedCorrection))
			}
			issuesSummary.WriteString(fmt.Sprintf("  Confidence: %.2f\n", h.ConfidenceLevel))
			issuesSummary.WriteString("\n")
		}
	}

	// Add duplicates
	if len(duplicates) > 0 {
		issuesSummary.WriteString("\nDuplicates:\n")
		for _, d := range duplicates {
			issuesSummary.WriteString(fmt.Sprintf("- %s: %s\n", d.ObjectType, d.Name))
			issuesSummary.WriteString(fmt.Sprintf("  IDs involved: %s\n", strings.Join(d.IDs, ", ")))
			issuesSummary.WriteString(fmt.Sprintf("  Preferred ID: %s\n", d.PreferredID))
			if len(d.ConflictingFields) > 0 {
				issuesSummary.WriteString(fmt.Sprintf("  Conflicting Fields: %s\n", strings.Join(d.ConflictingFields, ", ")))
			}
			issuesSummary.WriteString("\n")
		}
	}

	return currentState.String(), issuesSummary.String()
}

// getAffectedItems now just returns relevant subsets from the context maps
func getAffectedItems(ctx *ContextMaps) (affectedServices map[string]*hsds_types.Service, affectedCapacities map[string]*hsds_types.ServiceCapacity) {
	affectedServices = make(map[string]*hsds_types.Service)
	affectedCapacities = make(map[string]*hsds_types.ServiceCapacity)

	for id := range ctx.AffectedSvcs {
		if service, exists := ctx.ServiceMap[id]; exists {
			affectedServices[id] = service
		}
	}

	for id := range ctx.AffectedCaps {
		if capacity, exists := ctx.CapacityMap[id]; exists {
			affectedCapacities[id] = capacity
		}
	}

	return affectedServices, affectedCapacities
}
