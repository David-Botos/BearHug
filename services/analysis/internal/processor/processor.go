package processor

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/david-botos/BearHug/services/analysis/internal/processor/inference"
	"github.com/david-botos/BearHug/services/analysis/internal/processor/structOutputs"
	"github.com/david-botos/BearHug/services/analysis/internal/processor/validation"
	"github.com/david-botos/BearHug/services/analysis/internal/supabase"
	"github.com/david-botos/BearHug/services/analysis/internal/types"
	"github.com/joho/godotenv"
)

func ProcessTranscript(params types.TranscriptsReqBody) (bool, error) {
	// Extract services based on the transcript
	unformattedServices, servicesExtractionErr := structOutputs.ServicesExtraction(params)
	if servicesExtractionErr != nil {
		return false, fmt.Errorf("error with service extraction: %w", servicesExtractionErr)
	}

	// Turn Services into DB format + Add Org ID FK
	services, infConvErr := structOutputs.ConvertInferenceToServices(unformattedServices, params.OrganizationID)
	if infConvErr != nil {
		return false, fmt.Errorf("error converting inference results: %w", infConvErr)
	}

	fmt.Printf("Generated services successfully: ", services != nil)

	// Generate prompt and schema to triage out what details are present
	detailTriagePrompt, detailTriageSchema := structOutputs.GenerateTriagePrompt(params.Transcript)

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

	// Run inference
	serviceDetailsRes, serviceDetailsErr := client.RunClaudeInference(inference.PromptParams{Prompt: detailTriagePrompt, Schema: detailTriageSchema})
	if serviceDetailsErr != nil {
		return false, fmt.Errorf("error with details identification: %w", serviceDetailsErr)
	}

	// Fetch existing services
	existingServices, fetchErr := supabase.FetchOrganizationServices(params.OrganizationID)
	if fetchErr != nil {
		return false, fmt.Errorf("error fetching existing services: %w", fetchErr)
	}

	// Create service context with both existing and new services
	serviceCtx := structOutputs.ServiceContext{
		ExistingServices: existingServices,
		NewServices:      services,
	}

	// Extract details about the identified detail categories
	extractedDetails, detailExtractionErr := structOutputs.HandleTriagedAnalysis(
		params.Transcript,
		serviceDetailsRes,
		serviceCtx,
	)
	if detailExtractionErr != nil {
		return false, fmt.Errorf("error fetching existing services: %w", detailExtractionErr)
	}

	// Validate the details extracted for duplicates and hallucinations
	validationResult, validatorErr := validation.ValidateExtractedInfo(extractedDetails, serviceCtx, params.Transcript)
	if validatorErr != nil {
		return false, fmt.Errorf(`Error when attempting to validate the information extracted from transcript: %w`, validatorErr)
	}

	if validationResult {
		return true, nil
	} else {
		return false, fmt.Errorf(`Ah shit.... Unhandled error... good luck homie`)
	}
}
