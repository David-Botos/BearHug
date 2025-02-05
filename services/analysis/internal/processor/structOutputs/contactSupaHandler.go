package structOutputs

import (
	"fmt"
	"strconv"

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

// UpdateExistingContact handles updating contact records and associated phone records
// based on matches found between inferred data and existing database records.
func UpdateExistingContact(match contactMatch, call_id string) error {
	log := logger.Get()
	log.Info().
		Str("contact_id", match.ExistingContact.ID).
		Str("match_type", match.MatchType).
		Bool("needs_new_phone", match.NeedsNewPhone).
		Msg("Starting contact update process")

	client, err := supabase.InitSupabaseClient()
	if err != nil {
		return fmt.Errorf("failed to initialize Supabase client: %w", err)
	}

	// Track all metadata changes
	var metadataInputs []supabase.MetadataInput

	// 1. Handle Contact Updates
	if updateData := buildContactUpdateData(match); len(updateData) > 0 {
		// Prepare metadata for contact changes
		metadataInputs = append(metadataInputs, buildContactMetadata(match, call_id, updateData)...)

		// Update contact in database
		data, _, err := client.From("contact").
			Update(updateData, "", "").
			Eq("id", match.ExistingContact.ID).
			Execute()
		if err != nil {
			log.Error().
				Err(err).
				Str("contact_id", match.ExistingContact.ID).
				Interface("update_data", updateData).
				Str("response", string(data)).
				Msg("Failed to update contact")
			return fmt.Errorf("failed to update contact %s: %w", match.ExistingContact.ID, err)
		}

		log.Info().
			Str("contact_id", match.ExistingContact.ID).
			Int("fields_updated", len(updateData)).
			Msg("Successfully updated contact fields")
	}

	// 2. Handle Phone Record Updates
	if err := handlePhoneUpdates(match, call_id, &metadataInputs); err != nil {
		return fmt.Errorf("failed to handle phone updates: %w", err)
	}

	// 3. Create all metadata entries
	if len(metadataInputs) > 0 {
		if err := supabase.CreateAndStoreMetadata(metadataInputs); err != nil {
			return fmt.Errorf("failed to create metadata entries: %w", err)
		}
	}

	return nil
}

// buildContactUpdateData determines which contact fields need to be updated
func buildContactUpdateData(match contactMatch) map[string]interface{} {
	updateData := make(map[string]interface{})

	// Check each field for updates
	if match.InferredContact.Name != "" &&
		(match.ExistingContact.Name == nil || match.InferredContact.Name != *match.ExistingContact.Name) {
		updateData["name"] = match.InferredContact.Name
	}

	if match.InferredContact.Title != nil &&
		(match.ExistingContact.Title == nil || *match.InferredContact.Title != *match.ExistingContact.Title) {
		updateData["title"] = *match.InferredContact.Title
	}

	if match.InferredContact.Department != nil &&
		(match.ExistingContact.Department == nil || *match.InferredContact.Department != *match.ExistingContact.Department) {
		updateData["department"] = *match.InferredContact.Department
	}

	if match.InferredContact.Email != nil &&
		(match.ExistingContact.Email == nil || *match.InferredContact.Email != *match.ExistingContact.Email) {
		updateData["email"] = *match.InferredContact.Email
	}

	return updateData
}

// buildContactMetadata creates metadata entries for contact field updates
func buildContactMetadata(match contactMatch, call_id string, updateData map[string]interface{}) []supabase.MetadataInput {
	var metadataInputs []supabase.MetadataInput

	// Helper to safely get string value from pointer
	getStringValue := func(s *string) string {
		if s == nil {
			return ""
		}
		return *s
	}

	// Create metadata for each updated field
	for field, newValue := range updateData {
		var oldValue string
		switch field {
		case "name":
			oldValue = getStringValue(match.ExistingContact.Name)
		case "title":
			oldValue = getStringValue(match.ExistingContact.Title)
		case "department":
			oldValue = getStringValue(match.ExistingContact.Department)
		case "email":
			oldValue = getStringValue(match.ExistingContact.Email)
		}

		metadataInputs = append(metadataInputs, supabase.MetadataInput{
			ResourceID:       match.ExistingContact.ID,
			CallID:           call_id,
			ResourceType:     "contact",
			LastActionType:   "UPDATE",
			FieldName:        field,
			PreviousValue:    oldValue,
			ReplacementValue: newValue.(string),
		})
	}

	return metadataInputs
}

// handlePhoneUpdates manages phone record creation and updates
func handlePhoneUpdates(match contactMatch, call_id string, metadataInputs *[]supabase.MetadataInput) error {
	log := logger.Get()

	client, err := supabase.InitSupabaseClient()
	if err != nil {
		return fmt.Errorf("failed to initialize Supabase client: %w", err)
	}

	// Case 1: Update existing phone record with new details
	if match.UpdatePhone && match.ExistingPhone != nil && match.InferredContact.Phone != nil {
		updateData := make(map[string]interface{})

		// Update description if provided
		if match.InferredContact.PhoneDescription != nil {
			updateData["description"] = *match.InferredContact.PhoneDescription
			*metadataInputs = append(*metadataInputs, supabase.MetadataInput{
				ResourceID:       match.ExistingPhone.ID,
				CallID:           call_id,
				ResourceType:     "phone",
				LastActionType:   "UPDATE",
				FieldName:        "description",
				PreviousValue:    getStringValue(match.ExistingPhone.Description),
				ReplacementValue: *match.InferredContact.PhoneDescription,
			})
		}

		// Update extension if provided
		if match.InferredContact.PhoneExtension != nil {
			floatVal := float64(*match.InferredContact.PhoneExtension)
			updateData["extension"] = floatVal
			*metadataInputs = append(*metadataInputs, supabase.MetadataInput{
				ResourceID:       match.ExistingPhone.ID,
				CallID:           call_id,
				ResourceType:     "phone",
				LastActionType:   "UPDATE",
				FieldName:        "extension",
				PreviousValue:    fmt.Sprintf("%v", match.ExistingPhone.Extension),
				ReplacementValue: fmt.Sprintf("%d", *match.InferredContact.PhoneExtension),
			})
		}

		if len(updateData) > 0 {
			_, _, err := client.From("phone").
				Update(updateData, "", "").
				Eq("id", match.ExistingPhone.ID).
				Execute()
			if err != nil {
				return fmt.Errorf("failed to update phone record: %w", err)
			}

			log.Info().
				Str("phone_id", match.ExistingPhone.ID).
				Int("fields_updated", len(updateData)).
				Msg("Updated existing phone record")
		}
	}

	// Case 2: Create new phone record
	// This handles both NeedsNewPhone=true and shared_phone cases
	if (match.NeedsNewPhone || match.MatchType == "shared_phone") &&
		match.InferredContact.Phone != nil {

		var extension *float64
		if match.InferredContact.PhoneExtension != nil {
			floatVal := float64(*match.InferredContact.PhoneExtension)
			extension = &floatVal
		}

		// Create new phone record
		phoneOpts := &hsds_types.PhoneOptions{
			OrganizationID: match.ExistingContact.OrganizationID,
			ContactID:      &match.ExistingContact.ID,
			Extension:      extension,
			Description:    match.InferredContact.PhoneDescription,
		}

		phone, err := hsds_types.NewPhone(*match.InferredContact.Phone, phoneOpts)
		if err != nil {
			return fmt.Errorf("failed to create new phone record: %w", err)
		}

		// Store the new phone record
		data, _, err := client.From("phone").
			Insert(phone, false, "", "representation", "").
			Execute()
		if err != nil {
			return fmt.Errorf("failed to store new phone record: %w, data: %s", err, string(data))
		}

		// Create metadata for the new phone
		createPhoneMetadata(phone, match.InferredContact, call_id, metadataInputs)

		log.Info().
			Str("contact_id", match.ExistingContact.ID).
			Str("phone_id", phone.ID).
			Str("match_type", match.MatchType).
			Msg("Created new phone record")
	}

	return nil
}

// Helper function to create phone metadata
func createPhoneMetadata(phone *hsds_types.Phone, inf contactInference, call_id string, metadataInputs *[]supabase.MetadataInput) {
	*metadataInputs = append(*metadataInputs, supabase.MetadataInput{
		ResourceID:       phone.ID,
		CallID:           call_id,
		ResourceType:     "phone",
		LastActionType:   "CREATE",
		FieldName:        "number",
		PreviousValue:    "",
		ReplacementValue: *inf.Phone,
	})

	if inf.PhoneExtension != nil {
		*metadataInputs = append(*metadataInputs, supabase.MetadataInput{
			ResourceID:       phone.ID,
			CallID:           call_id,
			ResourceType:     "phone",
			LastActionType:   "CREATE",
			FieldName:        "extension",
			PreviousValue:    "",
			ReplacementValue: strconv.Itoa(*inf.PhoneExtension),
		})
	}

	if inf.PhoneDescription != nil {
		*metadataInputs = append(*metadataInputs, supabase.MetadataInput{
			ResourceID:       phone.ID,
			CallID:           call_id,
			ResourceType:     "phone",
			LastActionType:   "CREATE",
			FieldName:        "description",
			PreviousValue:    "",
			ReplacementValue: *inf.PhoneDescription,
		})
	}
}
