package processor

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/david-botos/BearHug/services/analysis/internal/processor/inference"
	"github.com/david-botos/BearHug/services/analysis/internal/processor/structOutputs"
	"github.com/david-botos/BearHug/services/analysis/internal/types"
	"github.com/joho/godotenv"
)

// TODO: Check the return types based on what i end up wanting to return, right now I assume ill just return true if its successful
func ProcessTranscript(params types.ProcessTranscriptParams) (bool, error) {
	// triagePrompt := triage.GenerateTriagePrompt(params.Transcript)

	// TODO: an error is the third variable, if I wanna propagate that up
	servicesPrompt, servicesSchema, _ := structOutputs.GenerateServicesPrompt("1770599e-2fdd-4e62-83d4-caf6456d5d15", params.Transcript)
	fmt.Printf("Generated prompt: %s\n", servicesPrompt)

	// triageSchema := triage.NewTriageSchema()
	fmt.Printf("Created schema: %+v\n", servicesSchema)

	// TODO: couch schema in tool format for api
	servicesSchemaTool := inference.ToolInputHelper(servicesSchema)

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

	triageInferenceResult, err := client.RunClaudeInference(inference.PromptParams{Prompt: servicesPrompt, Schema: servicesSchemaTool})
	if err != nil {
		fmt.Printf("Error occurred during inference: %v\n", err)
		return false, fmt.Errorf("error reading response: %w", err)
	}

	fmt.Printf("Received result: %+v\n", triageInferenceResult)

	// TODO: incomplete
	return true, nil
}
