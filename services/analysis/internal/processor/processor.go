package processor

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/david-botos/BearHug/services/analysis/internal/processor/inference"
	"github.com/david-botos/BearHug/services/analysis/internal/processor/triage"
	"github.com/david-botos/BearHug/services/analysis/internal/types"
	"github.com/joho/godotenv"
)

// TODO: Check the return types based on what i end up wanting to return, right now I assume ill just return true if its successful
func ProcessTranscript(params types.ProcessTranscriptParams) (bool, error) {
	triagePrompt := triage.GenerateTriagePrompt(params.Transcript)
	fmt.Printf("Generated prompt: %s\n", triagePrompt)

	triageSchema := triage.NewTriageSchema()
	fmt.Printf("Created schema: %+v\n", triageSchema)

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

	triageInferenceResult, err := client.RunClaudeInference(inference.PromptParams{Prompt: triagePrompt, Schema: triageSchema})
	if err != nil {
		fmt.Printf("Error occurred during inference: %v\n", err)
		return false, fmt.Errorf("error reading response: %w", err)
	}

	fmt.Printf("Received result: %+v\n", triageInferenceResult)

}
