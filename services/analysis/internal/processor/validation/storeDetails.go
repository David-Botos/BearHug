package validation

import (
	"github.com/david-botos/BearHug/services/analysis/internal/processor/structOutputs"
	"github.com/david-botos/BearHug/services/analysis/internal/supabase"
)

// Not using this
func SubmitValidatedOutput(validatedDetails []*structOutputs.DetailAnalysisResult) (bool, error) {
	for _, item := range validatedDetails {
		switch item.Category {
		case "CAPACITY":
			if len(item.CapacityData.Capacities) > 0 {
				supabase.StoreNewCapacity(item.CapacityData.Capacities)
			}
			if len(item.CapacityData.Units) > 0 {
				supabase.StoreNewUnits(item.CapacityData.Units)
			}
			// TODO: as other cases are handled add more here
		}
	}

	return true, nil
}
