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
		// TODO: else if ... other detail categories
	}
	return nil
}
