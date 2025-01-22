package prompt

import (
	"fmt"
	"strings"
)

// ServiceCategory represents the type of service being requested
type ServiceCategory string

// Define constants for each service category
const (
	DisabledResources     ServiceCategory = "resources for the disabled"
	UnemploymentResources ServiceCategory = "resources for the unemployed"
	FoodResources         ServiceCategory = "food resources"
	ClothingHygiene       ServiceCategory = "clothing and hygiene resources"
	Transportation        ServiceCategory = "transportation resources"
	MentalHealth          ServiceCategory = "mental health resources"
	DomesticViolence      ServiceCategory = "assistance with domestic violence"
	Education             ServiceCategory = "education assistance"
	Financial             ServiceCategory = "financial assistance"
	Healthcare            ServiceCategory = "health care resources"
	Shelter               ServiceCategory = "shelter or housing"
	BrainInjury           ServiceCategory = "assistance with traumatic brain injuries"
)

// ServiceCategories represents a set of selected service categories
type ServiceCategories []ServiceCategory

func GenPrompt(cboName string, serviceCategories ServiceCategories) (string, error) {
	if cboName == "" {
		return "", fmt.Errorf("cbo name cannot be empty")
	}
	if len(serviceCategories) == 0 {
		return "", fmt.Errorf("at least one service category must be provided")
	}

	// Validate service categories
	for _, category := range serviceCategories {
		if !isValidCategory(category) {
			return "", fmt.Errorf("invalid service category: %s", category)
		}
	}

	// Format service categories for natural language
	serviceList := formatServiceList(serviceCategories)

	prompt := fmt.Sprintf(`Hello! I am an AI assistant working with 211 to update our resource directory. I am calling specifically about %s and your services related to %s.

I will be asking a few questions about your organizations services to help connect people in need with the right resources. This call should take about 5-7 minutes. Is this a good time to talk?

For each service, I need to gather:
- Current availability and capacity
- Eligibility requirements
- Application process
- Any fees or costs
- Schedule and hours of operation
- How often your capacity changes

I understand you provide assistance with %s. Could you tell me about your current capacity and availability for these services?

Please know that:
1. I am an AI assistant working to help 211 maintain accurate resource information
2. All information will be used to help connect people in need with available services
3. I will keep our conversation focused and respect your time

Your responses will help us direct people to the right resources when they are in need. I will adapt my questions based on your answers and keep everything brief and clear.`,
		cboName,
		serviceList,
		serviceList)

	return prompt, nil
}

// isValidCategory checks if a service category is valid
func isValidCategory(category ServiceCategory) bool {
	switch category {
	case DisabledResources, UnemploymentResources, FoodResources,
		ClothingHygiene, Transportation, MentalHealth,
		DomesticViolence, Education, Financial,
		Healthcare, Shelter, BrainInjury:
		return true
	}
	return false
}

// formatServiceList creates a grammatically correct list of services
func formatServiceList(categories ServiceCategories) string {
	if len(categories) == 0 {
		return ""
	}

	var categoryStrings []string
	for _, cat := range categories {
		categoryStrings = append(categoryStrings, string(cat))
	}

	if len(categoryStrings) == 1 {
		return categoryStrings[0]
	}

	if len(categoryStrings) == 2 {
		return categoryStrings[0] + " and " + categoryStrings[1]
	}

	lastCategory := categoryStrings[len(categoryStrings)-1]
	otherCategories := categoryStrings[:len(categoryStrings)-1]
	return strings.Join(otherCategories, ", ") + ", and " + lastCategory
}
