package structOutputs

import (
	"fmt"
	"strings"

	"github.com/david-botos/BearHug/services/analysis/internal/processor/inference"
	"github.com/david-botos/BearHug/services/analysis/internal/supabase"
)

func GenerateServicesPrompt(organization_id string, transcript string) (string, interface{}, error) {
	// fetch and store the organization and its services from supa
	orgName, orgNameFetchErr := supabase.FetchOrganizationName(organization_id)
	if orgNameFetchErr != nil {
		return "", nil, fmt.Errorf("organization_lookup_failed: %w", orgNameFetchErr)
	}

	orgServices, orgServicesFetchErr := supabase.FetchOrganizationServices(organization_id)
	if orgServicesFetchErr != nil {
		return "", nil, fmt.Errorf("services_lookup_failed: %w", orgServicesFetchErr)
	}

	// Format existing services into a readable string
	var servicesText string
	if len(orgServices) == 0 {
		servicesText = "No services are currently documented for this organization."
	} else {
		var servicesList strings.Builder
		for i, service := range orgServices {
			servicesList.WriteString(fmt.Sprintf("%d. %s\n", i+1, service.Name))

			if service.Description != nil {
				servicesList.WriteString(fmt.Sprintf("   Description: %s\n", *service.Description))
			}

			if service.EligibilityDescription != nil {
				servicesList.WriteString(fmt.Sprintf("   Eligibility: %s\n", *service.EligibilityDescription))
			}

			if service.FeesDescription != nil {
				servicesList.WriteString(fmt.Sprintf("   Fees: %s\n", *service.FeesDescription))
			}

			if i < len(orgServices)-1 {
				servicesList.WriteString("\n")
			}
		}
		servicesText = servicesList.String()
	}

	prompt := fmt.Sprintf(`You are a service data extraction specialist that documents details about human services available to the underprivileged in your community. 
    Your task is to identify and structure information about new services mentioned in a conversation transcript. You have been provided with both the transcript and services
    that are currently documented for %s, the organization that was called to generate the transcript.

Transcript:
%s

Services that were previously documented as being offered by %s:
%s

Your task is to extract detailed information about each new service mentioned and structure it according to the provided schema. For each service:
1. Create a complete service entry with all required fields (name, status, description)
2. Include any optional fields that were explicitly mentioned or can be easily inferred in the transcript. Do not invent details that are not implied.
3. Use clear, objective language for descriptions
4. Don't worry about capturing scheduling details or capacity information about each service, that will be captured in another table 

Guidelines for extraction:
- If multiple similar services are mentioned (e.g., different medical services), create separate entries for each distinct service
- Default service status to "active" unless otherwise indicated
- Extract any eligibility requirements or application processes mentioned
- Capture specific details about fees (including if service is free)

IMPORTANT: You must ONLY respond by using the extract_services tool to output the structured data. Do not provide any explanatory text, confirmations, or additional messages. 
Simply use the tool to output the structured data following the schema exactly. Only include information that was explicitly discussed - do not make assumptions or add details 
not easily inferred from the transcript.`, orgName, transcript, orgName, servicesText)

	return prompt, ServicesSchema, nil
}

var ServicesSchema = inference.ToolInputSchema{
	Type: "object",
	Properties: map[string]inference.Property{
		"new_services": {
			Type:        "array",
			Description: "Array of new services identified in the conversation",
			Items: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					// Handled Manually
					// "organization_id": map[string]interface{}{
					// 	"type":        "string",
					// 	"description": "UUID v4 of the organization offering the service",
					// },
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Primary name of the service",
					},
					"status": map[string]interface{}{
						"type":        "string",
						"description": "Current operational status of the service",
						"enum":        []string{"active", "inactive", "defunct"},
					},
					// Handled Manually
					// "program_id": map[string]interface{}{
					// 	"type":        "string",
					// 	"description": "Optional UUID v4 of the program this service belongs to",
					// },
					"alternate_name": map[string]interface{}{
						"type":        "string",
						"description": "Alternative name or abbreviation for the service",
					},
					"description": map[string]interface{}{
						"type":        "string",
						"description": "Detailed description of what the service provides",
					},
					"url": map[string]interface{}{
						"type":        "string",
						"description": "Website or webpage for the service",
					},
					"email": map[string]interface{}{
						"type":        "string",
						"description": "Contact email for the service",
					},
					"interpretation_services": map[string]interface{}{
						"type":        "string",
						"description": "Languages and interpretation services available",
					},
					"application_process": map[string]interface{}{
						"type":        "string",
						"description": "Steps required to apply for or access the service",
					},
					"fees_description": map[string]interface{}{
						"type":        "string",
						"description": "Detailed description of any fees or costs",
					},
					"accreditations": map[string]interface{}{
						"type":        "string",
						"description": "Any professional accreditations or certifications",
					},
					"eligibility_description": map[string]interface{}{
						"type":        "string",
						"description": "Who is eligible to receive this service",
					},
					"minimum_age": map[string]interface{}{
						"type":        "number",
						"description": "Minimum age requirement for service recipients",
					},
					"maximum_age": map[string]interface{}{
						"type":        "number",
						"description": "Maximum age limit for service recipients",
					},
					"alert": map[string]interface{}{
						"type":        "string",
						"description": "Important notices or warnings about the service",
					},
				},
				"required": []string{"name", "status", "description"},
			},
		},
	},
	Required: []string{"new_services"},
}
