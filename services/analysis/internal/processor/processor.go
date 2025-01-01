package processor

import (
	"fmt"
	"os"

	"github.com/david-botos/BearHug/services/analysis/internal/types"
)

// TODO: Check the return types based on what i end up wanting to return, right now I assume ill just return true if its successful
func ProcessTranscript(params types.ProcessTranscriptParams) (bool, error) {
	prompt := GenerateTriagePrompt(params.Transcript)
	schema := NewTriageSchema()
	client := NewClient(os.Getenv("ANTHROPIC_API_KEY"))
	result, err := client.RunClaudeInference(TriagePromptParams{prompt, schema})

	// TODO: based on the tables identified that could be affected by the transcript, launch a goroutine to export to each standardized output
	// TODO: erase ts below i just dont like the color red when im pushing
	if err != nil {
		return false, fmt.Errorf("error reading response: %w", err)
	}
	if result != nil {
		return true, nil
	}
	return true, nil
}
