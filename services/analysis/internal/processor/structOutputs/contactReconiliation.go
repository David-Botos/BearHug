package structOutputs

import (
	"strings"

	"github.com/david-botos/BearHug/services/analysis/internal/hsds_types"
	"github.com/david-botos/BearHug/services/analysis/pkg/logger"
)

type contactMatch struct {
	InferredContact contactInference
	ExistingContact *hsds_types.Contact
	ExistingPhone   *hsds_types.Phone
	MatchConfidence float64
	MatchType       string // "phone", "email", "name", or "multiple"
}

type matchResult struct {
	Matches         []contactMatch
	UnmatchedInf    []contactInference
	UpdatedContacts []*hsds_types.Contact
	UpdatedPhones   []*hsds_types.Phone
	NewContacts     []*hsds_types.Contact
	NewPhones       []*hsds_types.Phone
}

// buildLookupMaps creates efficient lookup structures for matching
func buildLookupMaps(orgContacts []hsds_types.Contact, relevantPhones []hsds_types.Phone) (
	phoneToContact map[string]*hsds_types.Contact,
	emailToContact map[string]*hsds_types.Contact,
	phoneDetails map[string]*hsds_types.Phone,
) {
	log := logger.Get()
	phoneToContact = make(map[string]*hsds_types.Contact)
	emailToContact = make(map[string]*hsds_types.Contact)
	phoneDetails = make(map[string]*hsds_types.Phone)

	// Build email to contact mapping
	for i := range orgContacts {
		contact := &orgContacts[i]
		if contact.Email != nil && *contact.Email != "" {
			normalizedEmail := strings.ToLower(*contact.Email)
			emailToContact[normalizedEmail] = contact
			log.Debug().
				Str("email", normalizedEmail).
				Str("contact_id", contact.ID).
				Str("name", *contact.Name).
				Msg("Added email to contact mapping")
		}
	}

	// Build phone mappings
	for i := range relevantPhones {
		phone := &relevantPhones[i]

		// Normalize phone number for consistent matching
		normalizedPhone := normalizePhoneNumber(phone.Number)
		phoneDetails[normalizedPhone] = phone

		// If phone is associated with a contact, create the mapping
		if phone.ContactID != nil {
			for i := range orgContacts {
				if orgContacts[i].ID == *phone.ContactID {
					phoneToContact[normalizedPhone] = &orgContacts[i]
					break
				}
			}
		}
	}

	// Log the final map contents
	log.Debug().
		Int("phone_map_size", len(phoneToContact)).
		Interface("phone_map", phoneToContact).
		Int("email_map_size", len(emailToContact)).
		Interface("email_map", emailToContact).
		Msg("Built mapping")

	return phoneToContact, emailToContact, phoneDetails
}

// findMatches identifies if mentioned contacts already exist in the database
// by checking phone, email, and name matches
func findMatches(mentionedContacts contactInfOutput,
	orgContacts []hsds_types.Contact,
	relevantPhones []hsds_types.Phone) matchResult {

	result := matchResult{
		Matches:      make([]contactMatch, 0),
		UnmatchedInf: make([]contactInference, 0),
	}

	phoneToContact, emailToContact, phoneDetails := buildLookupMaps(orgContacts, relevantPhones)
	log := logger.Get()

	for _, mentionedContact := range mentionedContacts.Contacts {
		// Track all matches found for this contact
		var matches []contactMatch

		// Check phone match
		if mentionedContact.Phone != nil {
			normalizedPhone := normalizePhoneNumber(*mentionedContact.Phone)
			if existingPhone, exists := phoneDetails[normalizedPhone]; exists {
				if existingContact, hasContact := phoneToContact[normalizedPhone]; hasContact {
					matches = append(matches, contactMatch{
						InferredContact: mentionedContact,
						ExistingContact: existingContact,
						ExistingPhone:   existingPhone,
						MatchType:       "phone",
					})
				}
			}
		}

		// Check email match
		if mentionedContact.Email != nil {
			normalizedEmail := strings.ToLower(*mentionedContact.Email)
			if existingContact, exists := emailToContact[normalizedEmail]; exists {
				matches = append(matches, contactMatch{
					InferredContact: mentionedContact,
					ExistingContact: existingContact,
					MatchType:       "email",
				})
			}
		}

		// Check name match if threshold met
		if match := findBestNameMatch(mentionedContact, orgContacts); match != nil {
			matches = append(matches, *match)
		}

		// Analyze matches found
		switch len(matches) {
		case 0:
			// No matches - this is a new contact
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
				Msg("Single match found")

		default:
			// Multiple matches - do they point to the same contact?
			if allMatchesSameContact(matches) {
				// All matches reference same contact - use the most complete match
				result.Matches = append(result.Matches, getBestMatch(matches))
				log.Debug().
					Interface("contact", mentionedContact).
					Int("match_count", len(matches)).
					Msg("Multiple matches to same contact")
			} else {
				// Conflicting matches - log warning and treat as new contact
				log.Warn().
					Interface("contact", mentionedContact).
					Interface("matches", matches).
					Msg("Conflicting matches found - treating as new contact")
				result.UnmatchedInf = append(result.UnmatchedInf, mentionedContact)
			}
		}
	}

	return result
}

// allMatchesSameContact checks if all matches reference the same contact
func allMatchesSameContact(matches []contactMatch) bool {
	if len(matches) <= 1 {
		return true
	}
	firstID := matches[0].ExistingContact.ID
	for _, match := range matches[1:] {
		if match.ExistingContact.ID != firstID {
			return false
		}
	}
	return true
}

// getBestMatch selects the most complete match (e.g., one with phone info if available)
func getBestMatch(matches []contactMatch) contactMatch {
	for _, match := range matches {
		if match.ExistingPhone != nil {
			return match
		}
	}
	return matches[0]
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
