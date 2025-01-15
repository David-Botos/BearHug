package structOutputs

import (
	"fmt"

	"github.com/david-botos/BearHug/services/analysis/internal/supabase"
	"github.com/david-botos/BearHug/services/analysis/pkg/logger"
)

func StoreDetails(extractedDetails []*DetailAnalysisResult) error {
	log := logger.Get()
	for _, detail := range extractedDetails {
		if detail.Category == "CAPACITY" {
			if len(detail.CapacityData.Capacities) > 0 {
				capacityStorageErr := supabase.StoreNewCapacity(detail.CapacityData.Capacities)
				if capacityStorageErr != nil {
					log.Error().
						Err(capacityStorageErr).
						Msg("Failed to store capacity details in supa")
					return fmt.Errorf("error storing capacity details: %w", capacityStorageErr)
				}
			}
			if len(detail.CapacityData.Units) > 0 {
				unitStorageErr := supabase.StoreNewUnits(detail.CapacityData.Units)
				if unitStorageErr != nil {
					log.Error().
						Err(unitStorageErr).
						Msg("Failed to store unit details in supa")
					return fmt.Errorf("error storing unit details: %w", unitStorageErr)
				}
			}
		}
		// TODO: else if ... other detail categories
	}
	return nil
}
