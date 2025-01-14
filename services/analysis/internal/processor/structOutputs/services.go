package structOutputs

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/david-botos/BearHug/services/analysis/internal/hsds_types"
	"github.com/david-botos/BearHug/services/analysis/internal/processor/inference"
	"github.com/david-botos/BearHug/services/analysis/internal/supabase"
	"github.com/david-botos/BearHug/services/analysis/internal/types"
	"github.com/joho/godotenv"
)

func GenerateServicesPrompt(organization_id string, transcript string) (string, inference.ToolInputSchema, error) {
	// fetch and store the organization and its services from supa
	orgName, orgNameFetchErr := supabase.FetchOrganizationName(organization_id)
	if orgNameFetchErr != nil {
		return "", inference.ToolInputSchema{}, fmt.Errorf("organization_lookup_failed: %w", orgNameFetchErr)
	}

	orgServices, orgServicesFetchErr := supabase.FetchOrganizationServices(organization_id)
	if orgServicesFetchErr != nil {
		return "", inference.ToolInputSchema{}, fmt.Errorf("services_lookup_failed: %w", orgServicesFetchErr)
	}

	// Format existing services into a readable string
	var servicesText string
	if len(orgServices) == 0 {
		servicesText = "No services are currently documented for this organization."
	} else {
		var servicesList strings.Builder
		for i, service := range orgServices {
			if service.Status == hsds_types.ServiceStatusActive {
				if service.AlternateName != nil {
					servicesList.WriteString(fmt.Sprintf("%d. %s AKA %s\n", i+1, service.Name, *service.AlternateName))
				} else {
					servicesList.WriteString(fmt.Sprintf("%d. %s\n", i+1, service.Name))
				}
				if service.Description != nil {
					servicesList.WriteString(fmt.Sprintf("Description: %s\n", *service.Description))
				}

				if i < len(orgServices)-1 {
					servicesList.WriteString("\n")
				}
			}
		}
		servicesText = servicesList.String()
	}

	prompt := fmt.Sprintf(`You are a service data extraction specialist that documents details about human services available to the underprivileged in your community. Your task is to identify and structure information about new services mentioned in a conversation transcript, not currently known for %s. You have been provided with both the transcript and services that are currently documented for %s, the organization that was called to generate the transcript.

Transcript:
%s

Services that were previously documented as being offered by %s:
%s

Your task is to extract detailed information about each new service mentioned and structure it according to the provided schema. For each service:
1. Create a complete service entry with all required fields (name, status, description)
2. Include any optional fields that were explicitly mentioned or can be easily inferred in the transcript. Do not invent details that are not implied.
3. Use clear, objective language for descriptions
4. Don't worry about capturing scheduling details or capacity information about each service, that will be captured in another table
5. If you aren't confident that the service mentioned in the call was already documented by 

Guidelines for extraction:
- If multiple similar services are mentioned (e.g., different medical services), create separate entries for each distinct service
- Default service status to "active" unless otherwise indicated
- Extract any eligibility requirements or application processes mentioned
- Capture specific details about fees (if the service is free state "Free" and nothing else)

IMPORTANT: You must ONLY respond by using the new_services tool to output the structured data. Do not provide any explanatory text, confirmations, or additional messages. Simply use the tool to output the structured data following the schema exactly.`, orgName, orgName, transcript, orgName, servicesText)

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

func ServicesExtraction(params types.TranscriptsReqBody) (map[string]interface{}, error) {
	// Generate Prompt and Schema for Services
	servicesPrompt, servicesSchema, promptErr := GenerateServicesPrompt(params.OrganizationID, params.Transcript)
	if promptErr != nil {
		return nil, fmt.Errorf("failed to generate services prompt: %w", promptErr)
	}
	fmt.Printf("Generated prompt: %s\n", servicesPrompt)
	fmt.Printf("Created schema: %+v\n", servicesSchema)

	// Declare Claude Inference Client
	workingDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	envPath := filepath.Join(workingDir, ".env")
	if err := godotenv.Load(envPath); err != nil {
		panic(err)
	}
	fmt.Printf("envPath declared as: %s\n", envPath)
	client := inference.NewClient(os.Getenv("ANTHROPIC_API_KEY"))
	fmt.Printf("Initialized client with API key length: %d\n", len(os.Getenv("ANTHROPIC_API_KEY")))

	// Get Services Inference Result
	servicesInferenceResult, servicesInferenceResultErr := client.RunClaudeInference(inference.PromptParams{Prompt: servicesPrompt, Schema: servicesSchema})
	if servicesInferenceResultErr != nil {
		fmt.Printf("Error occurred during inference: %v\n", err)
		return nil, fmt.Errorf("error reading response: %w", err)
	}
	return servicesInferenceResult, nil
}

func ConvertInferenceToServices(inferenceResult map[string]interface{}, orgID string) ([]*hsds_types.Service, error) {
	rawServices, ok := inferenceResult["new_services"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid services format in inference result")
	}

	var services []*hsds_types.Service
	for _, raw := range rawServices {
		rawMap, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}

		status, err := parseServiceStatus(rawMap["status"].(string))
		if err != nil {
			continue
		}

		opts := &hsds_types.ServiceOptions{
			Description:            getStringPtr(rawMap["description"]),
			AlternateName:          getStringPtr(rawMap["alternate_name"]),
			URL:                    getStringPtr(rawMap["url"]),
			Email:                  getStringPtr(rawMap["email"]),
			InterpretationServices: getStringPtr(rawMap["interpretation_services"]),
			ApplicationProcess:     getStringPtr(rawMap["application_process"]),
			FeesDescription:        getStringPtr(rawMap["fees_description"]),
			Accreditations:         getStringPtr(rawMap["accreditations"]),
			EligibilityDescription: getStringPtr(rawMap["eligibility_description"]),
			MinimumAge:             getFloat64Ptr(rawMap["minimum_age"]),
			MaximumAge:             getFloat64Ptr(rawMap["maximum_age"]),
			Alert:                  getStringPtr(rawMap["alert"]),
		}

		service, err := hsds_types.NewService(orgID, rawMap["name"].(string), status, opts)
		if err != nil {
			continue
		}
		services = append(services, service)
	}

	return services, nil
}

func getStringPtr(v interface{}) *string {
	if v == nil {
		return nil
	}
	s, ok := v.(string)
	if !ok {
		return nil
	}
	return &s
}

func getFloat64Ptr(v interface{}) *float64 {
	if v == nil {
		return nil
	}
	switch n := v.(type) {
	case float64:
		return &n
	case int:
		f := float64(n)
		return &f
	default:
		return nil
	}
}

func parseServiceStatus(status string) (hsds_types.ServiceStatusEnum, error) {
	switch status {
	case "active":
		return hsds_types.ServiceStatusActive, nil
	case "inactive":
		return hsds_types.ServiceStatusInactive, nil
	case "defunct":
		return hsds_types.ServiceStatusDefunct, nil
	default:
		return "", fmt.Errorf("invalid service status: %s", status)
	}
}
