package structOutputs

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/david-botos/BearHug/services/analysis/internal/hsds_types"
	"github.com/david-botos/BearHug/services/analysis/internal/processor/inference"
	"github.com/david-botos/BearHug/services/analysis/internal/supabase"
	"github.com/david-botos/BearHug/services/analysis/pkg/logger"
)

func GenerateContactCategoryPrompt(transcript string, serviceCtx ServiceContext) (string, inference.ToolInputSchema, error) {
	log := logger.Get()
	log.Debug().
		Int("existing_services", len(serviceCtx.ExistingServices)).
		Int("new_services", len(serviceCtx.NewServices)).
		Msg("Generating service capacity prompt")

	// Build service descriptions
	var existingServiceDesc, newServiceDesc strings.Builder

	// Process existing services
	for _, service := range serviceCtx.ExistingServices {
		writeServiceDescription(&existingServiceDesc, *service)
	}

	// Process new services
	for _, service := range serviceCtx.NewServices {
		if service != nil {
			writeServiceDescription(&newServiceDesc, *service)
		}
	}

	prompt := fmt.Sprintf(`Extract contact information from the following transcript and map contacts to any services they are responsible for or associated with. Follow these rules:

Contact Information Rules:
1. Service Responsibility Mapping:
   - Link contacts to any services they are mentioned as handling, managing, or being responsible for
   - Use exact service names from either the "Existing Services" or "New Services" sections
   - If a contact is mentioned without clear service associations, still include the contact

2. Data Requirements:
   - Include a contact entry if ANY of these are mentioned: name, email, or phone number
   - Phone numbers must be formatted with country code (default to +1 for US)
   - Extensions must be in integer format
   - Service names must match exactly with one of those listed below if it is mentioned in the transcript that the contact is responsible for specific services.  Otherwise leave this field blank.
   - Capture contextual information about phone numbers in the phoneDescription field, such as:
  	* Whether it's a front desk, direct line, or general contact
  	* Any specific guidance about when to use the number
 	* Whether it's a personal or shared line

Conversation Transcript:
%s

Current Service Information:
---------------------------
Existing Services (reference these exact names when mapping contacts to services):
%s

New Services (reference these exact names when mapping contacts to services):
%s

IMPORTANT: Respond ONLY with the structured data output. Do not include any additional text, explanations, or notes.`, transcript, existingServiceDesc.String(), newServiceDesc.String())

	return prompt, ContactInformationSchema, nil
}

var ContactInformationSchema = inference.ToolInputSchema{
	Type: "object",
	Properties: map[string]inference.Property{
		"contacts": {
			Type: "array",
			Items: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name": map[string]interface{}{
						"type":        "string",
						"description": "The contact's name, which may include first name only or both first and last names",
					},
					"title": map[string]interface{}{
						"type":        "string",
						"description": "The contact's job title",
					},
					"department": map[string]interface{}{
						"type":        "string",
						"description": "The contact's department",
					},
					"email": map[string]interface{}{
						"type":        "string",
						"description": "The contact's email address",
					},
					"phone": map[string]interface{}{
						"type":        "string",
						"description": "The contact's phone number in international format (e.g., '+12344567890'). Assume +1 for US when no country code is specified",
					},
					"phoneDescription": map[string]interface{}{
						"type":        "string",
						"description": "A description of what to expect when calling this number (e.g., 'front desk', 'direct line', 'after-hours emergency line')",
					},
					"phoneExtension": map[string]interface{}{
						"type":        "integer",
						"description": "The contact's phone extension in integer format",
					},
					"responsibleServices": map[string]interface{}{
						"type": "array",
						"items": map[string]interface{}{
							"type":        "string",
							"description": "Name of a service this contact is responsible for, matching services mentioned in the provided context",
						},
						"description": "List of services this contact is responsible for managing or supporting",
					},
				},
				"anyOf": []map[string]interface{}{
					{
						"required": []string{"name"},
					},
					{
						"required": []string{"email"},
					},
					{
						"required": []string{"phone"},
					},
				},
			},
		},
	},
	Required: []string{"contacts"},
}

type contactInference struct {
	Name                string   `json:"name"`
	Title               *string  `json:"title,omitempty"`
	Department          *string  `json:"department,omitempty"`
	Email               *string  `json:"email,omitempty"`
	Phone               *string  `json:"phone,omitempty"`
	PhoneDescription    *string  `json:"phoneDescription,omitempty"`
	PhoneExtension      *int     `json:"phoneExtension,omitempty"`
	ResponsibleServices []string `json:"responsibleServices,omitempty"`
}

type contactInfOutput struct {
	Contacts []contactInference `json:"contacts"`
}

func infToContactsAndPhones(inferenceResult map[string]interface{}, serviceCtx ServiceContext, org_id string, call_id string) ([]*hsds_types.Contact, []*hsds_types.Phone, error) {
	log := logger.Get()

	// Unmarshal inference result
	jsonData, err := json.Marshal(inferenceResult)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal inference result")
		return nil, nil, fmt.Errorf("error marshaling inference result: %w", err)
	}

	var mentionedContacts contactInfOutput
	if err := json.Unmarshal(jsonData, &mentionedContacts); err != nil {
		log.Error().Err(err).Msg("Failed to unmarshal to structured output")
		return nil, nil, fmt.Errorf("error unmarshaling to structured output: %w", err)
	}

	/* Step 1: Fetch all the relevant data from supabase */

	// Fetch all the contacts for the organization
	orgContacts, contactFetchErr := supabase.FetchOrgContacts(org_id)
	if contactFetchErr != nil {
		return nil, nil, fmt.Errorf("error fetching organization contacts: %w", err)
	}

	// Make a simple slice of all the Services for the org
	services := make([]*hsds_types.Service, 0, len(serviceCtx.ExistingServices)+len(serviceCtx.NewServices))
	services = append(services, serviceCtx.ExistingServices...)
	services = append(services, serviceCtx.NewServices...)

	// Isolate contact ids and service ids so relevant phones can be pulled
	contactIDs := make([]string, 0, len(orgContacts))
	serviceIDs := make([]string, 0, len(services))

	for _, contact := range orgContacts {
		contactIDs = append(contactIDs, contact.ID)
	}

	for _, service := range services {
		serviceIDs = append(serviceIDs, service.ID)
	}

	// Go through the Phone table and select any entries linked via foreign key to the organization, contacts, or services
	relevantPhones, phoneFetchErr := supabase.FetchRelevantPhones(org_id, contactIDs, serviceIDs)
	if phoneFetchErr != nil {
		return nil, nil, fmt.Errorf("error fetching relevant phones: %w", phoneFetchErr)
	}

	/* Step 2: Match observations to existing data */
	matchResults := findMatches(mentionedContacts, orgContacts, relevantPhones, services)

	/* Step 3: Process Updates */
	for _, match := range matchResults.Matches {
		err := UpdateExistingContact(match, call_id)
		if err != nil {
			return nil, nil, fmt.Errorf("Error when updating existing contact: %w", err)
		}
	}

	/* Step 4: Create new records for unmmatched mentions */
	newContacts, newPhones, recordCreationErr := CreateNewContactAndPhoneRecords(matchResults.UnmatchedInf, org_id, call_id)
	if recordCreationErr != nil {
		return nil, nil, fmt.Errorf("Error when creating new records for unmmatched data: %w", err)
	}

	return newContacts, newPhones, nil
}

func analyzeContactCategoryDetails(transcript string, org_id string, serviceCtx ServiceContext, call_id string) (DetailAnalysisResult, error) {
	log := logger.Get()
	log.Debug().Msg("Starting contact details analysis")

	// Generate Prompt and Schema
	prompt, schema, err := GenerateContactCategoryPrompt(transcript, serviceCtx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate service capacity prompt")
		return DetailAnalysisResult{}, fmt.Errorf(`Failure when generating service capacity prompt: %w`, err)
	}

	// Declare Claude Inference Client
	client, err := inference.InitInferenceClient()
	if err != nil {
		log.Error().Err(err).Msg("Failed to initialize inference client")
		return DetailAnalysisResult{}, fmt.Errorf("failed to initialize inference client: %w", err)
	}

	log.Debug().Msg("Running Claude inference for capacity analysis")
	// Run Inference
	unformmattedContactDetails, inferenceErr := client.RunClaudeInference(inference.PromptParams{Prompt: prompt, Schema: schema})
	if inferenceErr != nil {
		log.Error().Err(inferenceErr).Msg("Error during inference execution")
		return DetailAnalysisResult{}, fmt.Errorf(`Error running inference to extract contact details: %w`, inferenceErr)
	}

	log.Debug().Msg("Converting inference response to contact and phone objects")
	contactDetails, phoneDetails, infConvErr := infToContactsAndPhones(unformmattedContactDetails, serviceCtx, org_id, call_id)
	if infConvErr != nil {
		log.Error().Err(infConvErr).Msg("Failed to convert inference response")
		return DetailAnalysisResult{}, fmt.Errorf(`Error while converting the inference response to clean contact and phone objects: %w`, infConvErr)
	}

	var result DetailAnalysisResult = NewContactResult(contactDetails, phoneDetails)
	log.Info().Int("contact_count", len(result.ContactData.Contacts)).Int("phone_count", len(result.ContactData.Phones)).Msg("Contact analysis completed successfully")
	return result, nil
}
