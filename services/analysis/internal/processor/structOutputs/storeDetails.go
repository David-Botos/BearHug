package structOutputs

import (
	"fmt"

	"github.com/david-botos/BearHug/services/analysis/internal/supabase"
	"github.com/david-botos/BearHug/services/analysis/pkg/logger"
)

// TODO: If the unit is new its uuid needs to be referenced by new capacity objects before those are uploaded...?
func StoreDetails(extractedDetails []*DetailAnalysisResult, callID string) error {
	log := logger.Get()
	for _, detail := range extractedDetails {
		if detail.Category == "CAPACITY" {
			if len(detail.CapacityData.Units) > 0 {
				unitStorageErr := supabase.StoreNewUnits(detail.CapacityData.Units, callID)
				if unitStorageErr != nil {
					log.Error().
						Err(unitStorageErr).
						Msg("Failed to store unit details in supa")
					return fmt.Errorf("error storing unit details: %w", unitStorageErr)
				}
			}
			if len(detail.CapacityData.Capacities) > 0 {
				capacityStorageErr := supabase.StoreNewCapacity(detail.CapacityData.Capacities, callID)
				if capacityStorageErr != nil {
					log.Error().
						Err(capacityStorageErr).
						Msg("Failed to store capacity details in supa")
					return fmt.Errorf("error storing capacity details: %w", capacityStorageErr)
				}
			}
		}
		if detail.Category == "CONTACT" {
			if len(detail.ContactData.Contacts) > 0 {
				contactStorageErr := supabase.StoreNewContacts(detail.ContactData.Contacts, callID)
				if contactStorageErr != nil {
					log.Error().
						Err(contactStorageErr).
						Msg("Failed to store contact details in supa")
					return fmt.Errorf("error storing contact details: %w", contactStorageErr)
				}
			}
			if len(detail.ContactData.Phones) > 0 {
				phoneStorageErr := supabase.StoreNewPhones(detail.ContactData.Phones, callID)
				if phoneStorageErr != nil {
					log.Error().
						Err(phoneStorageErr).
						Msg("Failed to store phone details in supa")
					return fmt.Errorf("error storing phone details: %w", phoneStorageErr)
				}
			}
		}
		// TODO: else if ... other detail categories
	}
	return nil
}
