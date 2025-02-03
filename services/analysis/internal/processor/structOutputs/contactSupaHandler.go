package structOutputs

import (
	"fmt"
	"strconv"
	"time"

	"github.com/david-botos/BearHug/services/analysis/internal/hsds_types"
	"github.com/david-botos/BearHug/services/analysis/internal/supabase"
	"github.com/david-botos/BearHug/services/analysis/pkg/logger"
)

func CreateNewContactAndPhoneRecords(unmatchedResults []contactInference, org_id, call_id string) ([]*hsds_types.Contact, []*hsds_types.Phone, error) {
	log := logger.Get()
	log.Info().
		Int("unmatched_count", len(unmatchedResults)).
		Msg("Starting creation of new contacts and phones")

	client, err := supabase.InitSupabaseClient()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize Supabase client: %w", err)
	}

	var newContacts []*hsds_types.Contact
	var newPhones []*hsds_types.Phone
	var metadataInputs []supabase.MetadataInput

	// First create all contacts
	for _, inference := range unmatchedResults {
		// Create new contact
		contactOpts := &hsds_types.ContactOptions{
			OrganizationID: &org_id,
			Name:           &inference.Name,
			Title:          inference.Title,
			Department:     inference.Department,
			Email:          inference.Email,
		}

		contact, err := hsds_types.NewContact(contactOpts)
		if err != nil {
			log.Error().
				Err(err).
				Interface("inference", inference).
				Msg("Failed to create new contact")
			return nil, nil, fmt.Errorf("failed to create new contact: %w", err)
		}

		// Store the contact in Supabase
		data, _, err := client.From("contact").
			Insert(contact, false, "", "representation", "").
			Execute()
		if err != nil {
			log.Error().
				Err(err).
				Str("contact_name", inference.Name).
				Str("response_data", string(data)).
				Msg("Failed to store new contact")
			return nil, nil, fmt.Errorf("failed to store new contact: %w", err)
		}

		newContacts = append(newContacts, contact)

		// Create metadata for the new contact
		metadataInputs = append(metadataInputs,
			supabase.MetadataInput{
				ResourceID:       contact.ID,
				CallID:           call_id,
				ResourceType:     "contact",
				LastActionType:   "CREATE",
				FieldName:        "name",
				PreviousValue:    "",
				ReplacementValue: inference.Name,
			})

		// Create metadata for other non-nil fields
		if inference.Title != nil {
			metadataInputs = append(metadataInputs,
				supabase.MetadataInput{
					ResourceID:       contact.ID,
					CallID:           call_id,
					ResourceType:     "contact",
					LastActionType:   "CREATE",
					FieldName:        "title",
					PreviousValue:    "",
					ReplacementValue: *inference.Title,
				})
		}

		if inference.Department != nil {
			metadataInputs = append(metadataInputs,
				supabase.MetadataInput{
					ResourceID:       contact.ID,
					CallID:           call_id,
					ResourceType:     "contact",
					LastActionType:   "CREATE",
					FieldName:        "department",
					PreviousValue:    "",
					ReplacementValue: *inference.Department,
				})
		}

		if inference.Email != nil {
			metadataInputs = append(metadataInputs,
				supabase.MetadataInput{
					ResourceID:       contact.ID,
					CallID:           call_id,
					ResourceType:     "contact",
					LastActionType:   "CREATE",
					FieldName:        "email",
					PreviousValue:    "",
					ReplacementValue: *inference.Email,
				})
		}

		// If there's a phone number, create a phone record
		if inference.Phone != nil {
			var extension *float64
			if inference.PhoneExtension != nil {
				floatVal := float64(*inference.PhoneExtension)
				extension = &floatVal
			}

			phoneOpts := &hsds_types.PhoneOptions{
				OrganizationID: &org_id,
				ContactID:      &contact.ID,
				Extension:      extension,
				Description:    inference.PhoneDescription,
				Type:           nil, // Could be added to contactInference if needed
			}

			phone, err := hsds_types.NewPhone(*inference.Phone, phoneOpts)
			if err != nil {
				log.Error().
					Err(err).
					Str("phone_number", *inference.Phone).
					Msg("Failed to create new phone")
				return nil, nil, fmt.Errorf("failed to create new phone: %w", err)
			}

			// Store the phone in Supabase
			phoneData, _, err := client.From("phone").
				Insert(phone, false, "", "representation", "").
				Execute()
			if err != nil {
				log.Error().
					Err(err).
					Str("phone_number", *inference.Phone).
					Str("response_data", string(phoneData)).
					Msg("Failed to store new phone")
				return nil, nil, fmt.Errorf("failed to store new phone: %w", err)
			}

			newPhones = append(newPhones, phone)

			// Create metadata for the new phone
			metadataInputs = append(metadataInputs,
				supabase.MetadataInput{
					ResourceID:       phone.ID,
					CallID:           call_id,
					ResourceType:     "phone",
					LastActionType:   "CREATE",
					FieldName:        "number",
					PreviousValue:    "",
					ReplacementValue: *inference.Phone,
				})

			if extension != nil {
				metadataInputs = append(metadataInputs,
					supabase.MetadataInput{
						ResourceID:       phone.ID,
						CallID:           call_id,
						ResourceType:     "phone",
						LastActionType:   "CREATE",
						FieldName:        "extension",
						PreviousValue:    "",
						ReplacementValue: strconv.Itoa(*inference.PhoneExtension),
					})
			}

			if inference.PhoneDescription != nil {
				metadataInputs = append(metadataInputs,
					supabase.MetadataInput{
						ResourceID:       phone.ID,
						CallID:           call_id,
						ResourceType:     "phone",
						LastActionType:   "CREATE",
						FieldName:        "description",
						PreviousValue:    "",
						ReplacementValue: *inference.PhoneDescription,
					})
			}
		}
	}

	// Create metadata entries for all new records
	if len(metadataInputs) > 0 {
		if err := supabase.CreateAndStoreMetadata(metadataInputs); err != nil {
			log.Error().
				Err(err).
				Msg("Failed to create metadata entries")
			return nil, nil, fmt.Errorf("failed to create metadata entries: %w", err)
		}
	}

	log.Info().
		Int("new_contacts", len(newContacts)).
		Int("new_phones", len(newPhones)).
		Msg("Successfully created new contacts and phones")

	return newContacts, newPhones, nil
}

func UpdateExistingContact(match contactMatch, call_id string) error {
	log := logger.Get()
	log.Info().
		Str("contact_id", match.ExistingContact.ID).
		Str("contact_name", getStringValue(match.ExistingContact.Name)).
		Msg("Starting contact update")

	client, err := supabase.InitSupabaseClient()
	if err != nil {
		log.Error().
			Err(err).
			Msg("Failed to initialize Supabase client")
		return fmt.Errorf("failed to initialize Supabase client: %w", err)
	}

	// Initialize update data and metadata inputs
	updateData := make(map[string]interface{})
	var metadataInputs []supabase.MetadataInput

	// Check and update Name if different
	if match.InferredContact.Name != "" &&
		(match.ExistingContact.Name == nil || match.InferredContact.Name != *match.ExistingContact.Name) {
		updateData["name"] = match.InferredContact.Name
		metadataInputs = append(metadataInputs, supabase.MetadataInput{
			ResourceID:       match.ExistingContact.ID,
			CallID:           call_id,
			ResourceType:     "contact",
			FieldName:        "name",
			PreviousValue:    getStringValue(match.ExistingContact.Name),
			ReplacementValue: match.InferredContact.Name,
			LastActionType:   "UPDATE",
		})
	}

	// Check and update Title if different
	if match.InferredContact.Title != nil &&
		(match.ExistingContact.Title == nil || *match.InferredContact.Title != *match.ExistingContact.Title) {
		updateData["title"] = *match.InferredContact.Title
		metadataInputs = append(metadataInputs, supabase.MetadataInput{
			ResourceID:       match.ExistingContact.ID,
			CallID:           call_id,
			ResourceType:     "contact",
			FieldName:        "title",
			PreviousValue:    getStringValue(match.ExistingContact.Title),
			ReplacementValue: *match.InferredContact.Title,
			LastActionType:   "UPDATE",
		})
	}

	// Check and update Department if different
	if match.InferredContact.Department != nil &&
		(match.ExistingContact.Department == nil || *match.InferredContact.Department != *match.ExistingContact.Department) {
		updateData["department"] = *match.InferredContact.Department
		metadataInputs = append(metadataInputs, supabase.MetadataInput{
			ResourceID:       match.ExistingContact.ID,
			CallID:           call_id,
			ResourceType:     "contact",
			FieldName:        "department",
			PreviousValue:    getStringValue(match.ExistingContact.Department),
			ReplacementValue: *match.InferredContact.Department,
			LastActionType:   "UPDATE",
		})
	}

	// Check and update Email if different
	if match.InferredContact.Email != nil &&
		(match.ExistingContact.Email == nil || *match.InferredContact.Email != *match.ExistingContact.Email) {
		updateData["email"] = *match.InferredContact.Email
		metadataInputs = append(metadataInputs, supabase.MetadataInput{
			ResourceID:       match.ExistingContact.ID,
			CallID:           call_id,
			ResourceType:     "contact",
			FieldName:        "email",
			PreviousValue:    getStringValue(match.ExistingContact.Email),
			ReplacementValue: *match.InferredContact.Email,
			LastActionType:   "UPDATE",
		})
	}

	// If there are no updates needed, return early
	if len(updateData) == 0 {
		log.Info().
			Str("contact_id", match.ExistingContact.ID).
			Msg("No updates needed for contact")
		return nil
	}

	// Update the UpdatedAt timestamp
	now := time.Now()
	updateData["updated_at"] = now

	// Update the contact in Supabase
	data, _, err := client.From("contact").
		Update(updateData, "", "").
		Eq("id", match.ExistingContact.ID).
		Execute()
	if err != nil {
		log.Error().
			Err(err).
			Str("contact_id", match.ExistingContact.ID).
			Str("response_data", string(data)).
			Msg("Failed to update contact")
		return fmt.Errorf("failed to update contact %s: %w, data: %s",
			match.ExistingContact.ID, err, string(data))
	}

	// Create metadata entries for all changes if there are any
	if len(metadataInputs) > 0 {
		if err := supabase.CreateAndStoreMetadata(metadataInputs); err != nil {
			log.Error().
				Err(err).
				Str("contact_id", match.ExistingContact.ID).
				Msg("Failed to create metadata entries")
			return fmt.Errorf("failed to create metadata entries for contact %s: %w",
				match.ExistingContact.ID, err)
		}
	}

	log.Info().
		Str("contact_id", match.ExistingContact.ID).
		Str("contact_name", getStringValue(match.ExistingContact.Name)).
		Int("fields_updated", len(updateData)-1). // Subtract 1 for updated_at
		Msg("Successfully updated contact")

	return nil
}

// Helper function to safely get string value from pointer
func getStringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
