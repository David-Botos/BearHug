package structOutputs

import (
	"strings"

	"github.com/david-botos/BearHug/services/analysis/internal/hsds_types"
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
func buildLookupMaps(orgContacts []hsds_types.Contact, relevantPhones []hsds_types.Phone, services []*hsds_types.Service) (
	phoneToContact map[string]*hsds_types.Contact,
	emailToContact map[string]*hsds_types.Contact,
	phoneDetails map[string]*hsds_types.Phone,
) {
	phoneToContact = make(map[string]*hsds_types.Contact)
	emailToContact = make(map[string]*hsds_types.Contact)
	phoneDetails = make(map[string]*hsds_types.Phone)

	// Build email to contact mapping
	for i := range orgContacts {
		contact := &orgContacts[i]
		if contact.Email != nil && *contact.Email != "" {
			emailToContact[strings.ToLower(*contact.Email)] = contact
		}
	}

	// Build service email mapping
	for _, service := range services {
		if service.Email != nil && *service.Email != "" {
			// If a service has an email, we might want to track it separately
			// or handle it differently depending on your requirements
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

	return phoneToContact, emailToContact, phoneDetails
}

// findMatches implements the core matching logic
func findMatches(mentionedContacts contactInfOutput,
	orgContacts []hsds_types.Contact,
	relevantPhones []hsds_types.Phone,
	services []*hsds_types.Service) matchResult {

	// Initialize result structure
	result := matchResult{
		Matches:         make([]contactMatch, 0),
		UnmatchedInf:    make([]contactInference, 0),
		UpdatedContacts: make([]*hsds_types.Contact, 0),
		UpdatedPhones:   make([]*hsds_types.Phone, 0),
		NewContacts:     make([]*hsds_types.Contact, 0),
		NewPhones:       make([]*hsds_types.Phone, 0),
	}

	// Build lookup maps
	phoneToContact, emailToContact, phoneDetails := buildLookupMaps(orgContacts, relevantPhones, services)

	// Process each inferred contact
	for _, mentionedContact := range mentionedContacts.Contacts {
		matched := false

		// Try phone match first if phone exists
		if mentionedContact.Phone != nil {
			normalizedPhone := normalizePhoneNumber(*mentionedContact.Phone)
			if existingPhone, exists := phoneDetails[normalizedPhone]; exists {
				if existingContact, hasContact := phoneToContact[normalizedPhone]; hasContact {
					result.Matches = append(result.Matches, contactMatch{
						InferredContact: mentionedContact,
						ExistingContact: existingContact,
						ExistingPhone:   existingPhone,
						MatchConfidence: 1.0,
						MatchType:       "phone",
					})
					matched = true
				}
			}
		}

		// Try email match if no phone match
		if !matched && mentionedContact.Email != nil {
			normalizedEmail := strings.ToLower(*mentionedContact.Email)
			if existingContact, exists := emailToContact[normalizedEmail]; exists {
				result.Matches = append(result.Matches, contactMatch{
					InferredContact: mentionedContact,
					ExistingContact: existingContact,
					MatchConfidence: 0.9,
					MatchType:       "email",
				})
				matched = true
			}
		}

		// Try name similarity if still no match
		if !matched && mentionedContact.Name != "" {
			bestMatch := findBestNameMatch(mentionedContact, orgContacts)
			if bestMatch != nil && bestMatch.MatchConfidence > 0.8 {
				result.Matches = append(result.Matches, *bestMatch)
				matched = true
			}
		}

		// If no match found, add to unmatched
		if !matched {
			result.UnmatchedInf = append(result.UnmatchedInf, mentionedContact)
		}
	}

	return result
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
