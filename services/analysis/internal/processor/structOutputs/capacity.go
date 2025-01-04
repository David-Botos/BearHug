package structOutputs

import "fmt"

func GenerateServiceCapacityPrompt(transcript string, reasoning string) (string, interface{}) {
	prompt := fmt.Sprintf(`You are a service data extraction specialist that documents details about human services available to the underprivileged in your community. 
	Your task is to identify and structure information about the capacity that certain services have. You have been provided with both the transcript and 
	initial reasoning about services mentioned.
	
	Transcript:
	%s
	
	Initial Service Analysis:
	%s
	`, transcript, reasoning)
	return prompt, ServicesSchema
}
