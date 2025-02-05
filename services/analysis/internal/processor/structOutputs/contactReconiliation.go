package structOutputs

import (
	"sort"
	"strings"

	"github.com/david-botos/BearHug/services/analysis/internal/hsds_types"
	"github.com/david-botos/BearHug/services/analysis/pkg/logger"
)

// contactMatch represents a match between inferred and existing data
type contactMatch struct {
	InferredContact contactInference
	ExistingContact *hsds_types.Contact
	ExistingPhone   *hsds_types.Phone
	MatchConfidence float64
	MatchType       string // "phone_and_name", "shared_phone", "orphaned_phone", "email", "name"
	NeedsNewPhone   bool   // Should create new phone record
	NeedsNewContact bool   // Should create new contact despite phone match
	UpdatePhone     bool   // Should update existing phone record with new details
}

// matchResult holds the results of the matching process
type matchResult struct {
	Matches         []contactMatch
	UnmatchedInf    []contactInference
	UpdatedContacts []*hsds_types.Contact
	UpdatedPhones   []*hsds_types.Phone
	NewContacts     []*hsds_types.Contact
	NewPhones       []*hsds_types.Phone
}

// phoneMapping represents the relationship between a phone number and its associated contacts
type phoneMapping struct {
	phone    *hsds_types.Phone     // The phone record from the database
	contacts []*hsds_types.Contact // All contacts associated with this phone number
}

// buildLookupMaps creates efficient lookup structures for contact matching operations.
// It handles cases where:
// 1. Multiple contacts share the same phone number
// 2. A single contact has multiple phone numbers
// 3. Phone numbers exist without associated contacts (orphaned)
// 4. Contacts exist with only email addresses
func buildLookupMaps(orgContacts []hsds_types.Contact, relevantPhones []hsds_types.Phone) (
	phoneToContacts map[string]phoneMapping, // Maps normalized phone numbers to their contacts
	emailToContact map[string]*hsds_types.Contact, // Maps lowercase emails to contacts
) {
	log := logger.Get()

	// Initialize maps
	phoneToContacts = make(map[string]phoneMapping)
	emailToContact = make(map[string]*hsds_types.Contact)

	// First pass: Build email-to-contact mapping
	// This is separate from phone mapping as email matches are handled differently
	for i := range orgContacts {
		contact := &orgContacts[i]
		if contact.Email != nil && *contact.Email != "" {
			normalizedEmail := strings.ToLower(*contact.Email)
			emailToContact[normalizedEmail] = contact
			log.Debug().
				Str("email", normalizedEmail).
				Str("contact_id", contact.ID).
				Str("name", getStringValue(contact.Name)).
				Msg("Added email to contact mapping")
		}
	}

	// Second pass: Build phone mappings
	// Handle both contact-to-phone relationships and orphaned phones
	for i := range relevantPhones {
		phone := &relevantPhones[i]
		normalizedPhone := normalizePhoneNumber(phone.Number)

		// Initialize or retrieve existing mapping
		mapping := phoneToContacts[normalizedPhone]
		mapping.phone = phone

		// If phone is associated with a contact, find and store the relationship
		if phone.ContactID != nil {
			for i := range orgContacts {
				if orgContacts[i].ID == *phone.ContactID {
					// Add to contacts slice if not already present
					contactExists := false
					for _, existingContact := range mapping.contacts {
						if existingContact.ID == orgContacts[i].ID {
							contactExists = true
							break
						}
					}
					if !contactExists {
						mapping.contacts = append(mapping.contacts, &orgContacts[i])
						log.Debug().
							Str("phone", normalizedPhone).
							Str("contact_id", orgContacts[i].ID).
							Str("name", getStringValue(orgContacts[i].Name)).
							Msg("Added phone to contact mapping")
					}
				}
			}
		} else {
			log.Debug().
				Str("phone", normalizedPhone).
				Msg("Found orphaned phone number")
		}

		// Update the mapping in the main map
		phoneToContacts[normalizedPhone] = mapping
	}

	// Log final mapping stats
	log.Debug().
		Int("phone_map_size", len(phoneToContacts)).
		Int("email_map_size", len(emailToContact)).
		Msg("Built lookup maps")

	// Detailed debug logging of phone mappings
	for phone, mapping := range phoneToContacts {
		contactIDs := make([]string, 0, len(mapping.contacts))
		for _, contact := range mapping.contacts {
			contactIDs = append(contactIDs, contact.ID)
		}
		log.Debug().
			Str("phone", phone).
			Strs("contact_ids", contactIDs).
			Msg("Phone mapping details")
	}

	return phoneToContacts, emailToContact
}

// getStringValue safely retrieves string value from pointer
func getStringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// FindMatches identifies relationships between inferred contacts from conversation transcripts
// and existing database records. It handles several complex scenarios:
// 1. Multiple contacts sharing the same phone number
// 2. Single contacts having multiple phone numbers
// 3. Phone numbers without associated contacts (orphaned)
// 4. Pure name matches without phone/email correlation
// 5. Email-only contacts
func FindMatches(mentionedContacts contactInfOutput,
	orgContacts []hsds_types.Contact,
	relevantPhones []hsds_types.Phone) matchResult {

	result := matchResult{
		Matches:      make([]contactMatch, 0),
		UnmatchedInf: make([]contactInference, 0),
	}

	// Build lookup structures for efficient matching
	phoneToContacts, emailToContact := buildLookupMaps(orgContacts, relevantPhones)

	// Process each contact mentioned in the transcript
	for _, mentionedContact := range mentionedContacts.Contacts {
		matches := findMatchesForContact(mentionedContact, phoneToContacts, emailToContact, orgContacts)
		processMatches(matches, mentionedContact, &result)
	}

	return result
}

// findMatchesForContact attempts to find all possible matches for a single mentioned contact
// Returns matches ordered by confidence (highest first)
func findMatchesForContact(
	mentionedContact contactInference,
	phoneToContacts map[string]phoneMapping,
	emailToContact map[string]*hsds_types.Contact,
	orgContacts []hsds_types.Contact) []contactMatch {

	var matches []contactMatch
	log := logger.Get()

	// Step 1: Check for phone matches
	if mentionedContact.Phone != nil {
		normalizedPhone := normalizePhoneNumber(*mentionedContact.Phone)
		if mapping, exists := phoneToContacts[normalizedPhone]; exists {
			// Phone exists in system - check all associated contacts
			if len(mapping.contacts) > 0 {
				for _, existingContact := range mapping.contacts {
					similarity := 0.0
					if existingContact.Name != nil {
						similarity = calculateNameSimilarity(mentionedContact.Name, *existingContact.Name)
					}

					match := contactMatch{
						InferredContact: mentionedContact,
						ExistingContact: existingContact,
						ExistingPhone:   mapping.phone,
						MatchConfidence: similarity,
					}

					// Determine match type and confidence based on name similarity
					if similarity > 0.8 {
						match.MatchType = "phone_and_name"
						match.UpdatePhone = true // Update phone details as this is recent data
					} else {
						// Phone matches but name differs significantly
						// Flag for new contact creation despite shared phone
						match.MatchType = "shared_phone"
						match.NeedsNewContact = true
					}

					matches = append(matches, match)
					log.Debug().
						Str("phone", normalizedPhone).
						Str("existing_name", getStringValue(existingContact.Name)).
						Str("mentioned_name", mentionedContact.Name).
						Float64("similarity", similarity).
						Str("match_type", match.MatchType).
						Msg("Found phone-based match")
				}
			} else {
				// Orphaned phone case
				matches = append(matches, contactMatch{
					InferredContact: mentionedContact,
					ExistingPhone:   mapping.phone,
					MatchType:       "orphaned_phone",
					MatchConfidence: 1.0,
					NeedsNewContact: true,
				})
				log.Debug().
					Str("phone", normalizedPhone).
					Msg("Found orphaned phone match")
			}
		}
	}

	// Step 2: Check for email matches if no high-confidence phone matches
	if mentionedContact.Email != nil && !hasHighConfidenceMatch(matches) {
		normalizedEmail := strings.ToLower(*mentionedContact.Email)
		if existingContact, exists := emailToContact[normalizedEmail]; exists {
			matches = append(matches, contactMatch{
				InferredContact: mentionedContact,
				ExistingContact: existingContact,
				MatchType:       "email",
				MatchConfidence: 1.0,
				NeedsNewPhone:   mentionedContact.Phone != nil, // Create new phone if mentioned
			})
			log.Debug().
				Str("email", normalizedEmail).
				Str("name", getStringValue(existingContact.Name)).
				Msg("Found email match")
		}
	}

	// Step 3: Check for name-only matches if no other matches found
	if len(matches) == 0 {
		if nameMatch := findBestNameMatch(mentionedContact, orgContacts); nameMatch != nil {
			nameMatch.NeedsNewPhone = mentionedContact.Phone != nil
			matches = append(matches, *nameMatch)
			log.Debug().
				Str("mentioned_name", mentionedContact.Name).
				Str("matched_name", getStringValue(nameMatch.ExistingContact.Name)).
				Float64("confidence", nameMatch.MatchConfidence).
				Msg("Found name-only match")
		}
	}

	// Sort matches by confidence, descending
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].MatchConfidence > matches[j].MatchConfidence
	})

	return matches
}

// processMatches handles the results of finding matches for a contact
func processMatches(matches []contactMatch, mentionedContact contactInference, result *matchResult) {
	log := logger.Get()

	switch len(matches) {
	case 0:
		// No matches - treat as new contact
		result.UnmatchedInf = append(result.UnmatchedInf, mentionedContact)
		log.Debug().
			Interface("contact", mentionedContact).
			Msg("No matches found - treating as new contact")

	case 1:
		// Single match - straightforward case
		result.Matches = append(result.Matches, matches[0])
		log.Debug().
			Interface("contact", mentionedContact).
			Str("match_type", matches[0].MatchType).
			Float64("confidence", matches[0].MatchConfidence).
			Bool("needs_new_phone", matches[0].NeedsNewPhone).
			Bool("needs_new_contact", matches[0].NeedsNewContact).
			Msg("Single match processed")

	default:
		// Multiple matches - use highest confidence match if it's significantly better
		bestMatch := matches[0]
		secondBestConfidence := matches[1].MatchConfidence

		if bestMatch.MatchConfidence > secondBestConfidence+0.2 { // Significant confidence gap
			result.Matches = append(result.Matches, bestMatch)
			log.Debug().
				Interface("contact", mentionedContact).
				Float64("best_confidence", bestMatch.MatchConfidence).
				Float64("second_best", secondBestConfidence).
				Msg("Selected best match from multiple")
		} else {
			// Confidence difference too small - treat as new contact
			result.UnmatchedInf = append(result.UnmatchedInf, mentionedContact)
			log.Debug().
				Interface("contact", mentionedContact).
				Interface("matches", matches).
				Msg("Ambiguous matches - treating as new contact")
		}
	}
}

// hasHighConfidenceMatch checks if there's already a high-confidence match
func hasHighConfidenceMatch(matches []contactMatch) bool {
	for _, match := range matches {
		if match.MatchConfidence > 0.8 {
			return true
		}
	}
	return false
}

// calculateNameSimilarity implements the Levenshtein distance calculation
func calculateNameSimilarity(name1, name2 string) float64 {
	// Convert to lowercase for comparison
	s1 := strings.ToLower(name1)
	s2 := strings.ToLower(name2)

	// Create matrix
	d := make([][]int, len(s1)+1)
	for i := range d {
		d[i] = make([]int, len(s2)+1)
	}

	// Initialize first row and column
	for i := range d {
		d[i][0] = i
	}
	for j := range d[0] {
		d[0][j] = j
	}

	// Fill in the rest of the matrix
	for i := 1; i <= len(s1); i++ {
		for j := 1; j <= len(s2); j++ {
			if s1[i-1] == s2[j-1] {
				d[i][j] = d[i-1][j-1]
			} else {
				min := d[i-1][j]
				if d[i][j-1] < min {
					min = d[i][j-1]
				}
				if d[i-1][j-1] < min {
					min = d[i-1][j-1]
				}
				d[i][j] = min + 1
			}
		}
	}

	// Calculate similarity score
	maxLen := float64(len(s1))
	if len(s2) > len(s1) {
		maxLen = float64(len(s2))
	}

	if maxLen == 0 {
		return 0
	}

	distance := float64(d[len(s1)][len(s2)])
	return 1 - (distance / maxLen)
}

// Helper functions

func normalizePhoneNumber(phone string) string {
	// Remove all non-numeric characters
	normalized := strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, phone)
	return normalized
}

func findBestNameMatch(inferredContact contactInference, orgContacts []hsds_types.Contact) *contactMatch {
	var bestMatch *contactMatch
	highestConfidence := 0.8 // Minimum threshold for name matching

	for i := range orgContacts {
		if orgContacts[i].Name == nil {
			continue
		}

		similarity := calculateNameSimilarity(inferredContact.Name, *orgContacts[i].Name)
		if similarity > highestConfidence {
			highestConfidence = similarity
			bestMatch = &contactMatch{
				InferredContact: inferredContact,
				ExistingContact: &orgContacts[i],
				MatchConfidence: similarity,
				MatchType:       "name",
			}
		}
	}

	return bestMatch
}
