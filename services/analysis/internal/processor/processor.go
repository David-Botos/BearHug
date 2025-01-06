package processor

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/david-botos/BearHug/services/analysis/internal/hsds_types"
	"github.com/david-botos/BearHug/services/analysis/internal/processor/inference"
	"github.com/david-botos/BearHug/services/analysis/internal/processor/structOutputs"
	"github.com/david-botos/BearHug/services/analysis/internal/types"
	"github.com/joho/godotenv"
)

// TODO: Check the return types based on what i end up wanting to return, right now I assume ill just return true if its successful
func ProcessTranscript(params types.TranscriptsReqBody) (bool, error) {
	/* Tranche 1: */
	// Extract services based on the transcript
	t1Services, t1ServicesErr := t1ServicesExtraction(params)
	if t1ServicesErr != nil {
		return false, fmt.Errorf("error with service extraction: %w", t1ServicesErr)
	}

	// Turn Services into DB format + Add Org ID FK
	services, err := convertInferenceToServices(t1Services, params.OrganizationID)
	if err != nil {
		return false, fmt.Errorf("error converting inference results: %w", err)
	}

	fmt.Printf("Generated services successfully: ", services != nil)

	// Insert new services into Supa

	/* Tranche 2: */

	// Analyze for details on old and new services

	// TODO: incomplete return
	return true, nil
}

func t1ServicesExtraction(params types.TranscriptsReqBody) (map[string]interface{}, error) {
	// Generate Prompt and Schema for Services
	servicesPrompt, servicesSchema, promptErr := structOutputs.GenerateServicesPrompt("1770599e-2fdd-4e62-83d4-caf6456d5d15", params.Transcript)
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

func convertInferenceToServices(inferenceResult map[string]interface{}, orgID string) ([]*hsds_types.Service, error) {
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
