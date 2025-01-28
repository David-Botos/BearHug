package prompt

import (
	"fmt"
	"strings"
)

// ServiceCategory represents the type of service being requested
type ServiceCategory string

// Define constants for each service category
const (
	DisabledResources     ServiceCategory = "DisabledResources"
	UnemploymentResources ServiceCategory = "UnemploymentResources"
	FoodResources         ServiceCategory = "FoodResources"
	ClothingHygiene       ServiceCategory = "ClothingHygiene"
	Transportation        ServiceCategory = "Transportation"
	MentalHealth          ServiceCategory = "MentalHealth"
	DomesticViolence      ServiceCategory = "DomesticViolence"
	Education             ServiceCategory = "Education"
	Financial             ServiceCategory = "Financial"
	Healthcare            ServiceCategory = "Healthcare"
	Shelter               ServiceCategory = "Shelter"
	BrainInjury           ServiceCategory = "BrainInjury"
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

	serviceList := formatServiceList(serviceCategories)

	prompt := fmt.Sprintf(`You are Alex, a Two One One resource directory specialist focused on %s. Speak naturally and professionally.

When the call connects, say exactly:
"Hello! I'm Alex from Two One One's resource directory. I'm calling to update our information about %s so we can connect people with your services more effectively. This will take about 5 to 7 minutes. Is now a good time?"

If they agree, ask:
"Thank you! Our records show you provide %s. Could you confirm which of these services you currently offer?"

For each confirmed service, ask these questions in order:
"What's your current availability?"
"Who is eligible for this service?"
"What's the application process?"
"What are your hours of operation?"

End with:
"Thank you so much for your time today. This will help us better serve the community."

Listen carefully and ask for clarification if needed. Keep responses brief and conversational.`,
		serviceList,
		cboName,
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

func (sc ServiceCategory) DisplayName() string {
	switch sc {
	case DisabledResources:
		return "resources for the disabled"
	case UnemploymentResources:
		return "resources for the unemployed"
	case FoodResources:
		return "food resources"
	case ClothingHygiene:
		return "clothing and hygiene resources"
	case Transportation:
		return "transportation resources"
	case MentalHealth:
		return "mental health resources"
	case DomesticViolence:
		return "assistance with domestic violence"
	case Education:
		return "education assistance"
	case Financial:
		return "financial assistance"
	case Healthcare:
		return "health care resources"
	case Shelter:
		return "shelter or housing"
	case BrainInjury:
		return "assistance with traumatic brain injuries"
	default:
		return string(sc)
	}
}

// formatServiceList creates a grammatically correct list of services
func formatServiceList(categories ServiceCategories) string {
	if len(categories) == 0 {
		return "No service category data exists for this organization"
	}

	var categoryStrings []string
	for _, cat := range categories {
		categoryStrings = append(categoryStrings, cat.DisplayName())
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
