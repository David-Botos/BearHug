package structOutputs

import (
	"context"
	"fmt"

	"github.com/david-botos/BearHug/services/analysis/internal/supabase"
	"github.com/david-botos/BearHug/services/analysis/pkg/logger"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// TODO: If the unit is new its uuid needs to be referenced by new capacity objects before those are uploaded...?
func StoreDetails(ctx context.Context, extractedDetails []*DetailAnalysisResult, callID string) error {
	tracer := otel.GetTracerProvider().Tracer("details-storage")
	ctx, span := tracer.Start(ctx, "store_details",
		trace.WithAttributes(
			attribute.String("call_id", callID),
			attribute.Int("total_details", len(extractedDetails)),
		),
	)
	defer span.End()

	log := logger.Get()

	for _, detail := range extractedDetails {
		if detail.Category == "CAPACITY" {
			// Create a span for capacity-type detail processing
			ctx, capacitySpan := tracer.Start(ctx, "process_capacity_detail")

			if len(detail.CapacityData.Units) > 0 {
				ctx, unitsSpan := tracer.Start(ctx, "store_units")
				unitsSpan.SetAttributes(
					attribute.Int("unit_count", len(detail.CapacityData.Units)),
				)

				unitStorageErr := supabase.StoreNewUnits(ctx, detail.CapacityData.Units, callID)
				if unitStorageErr != nil {
					unitsSpan.RecordError(unitStorageErr)
					unitsSpan.End()
					capacitySpan.RecordError(unitStorageErr)
					capacitySpan.End()
					log.Error().
						Err(unitStorageErr).
						Msg("Failed to store unit details in supa")
					return fmt.Errorf("error storing unit details: %w", unitStorageErr)
				}
				unitsSpan.End()
			}

			if len(detail.CapacityData.Capacities) > 0 {
				ctx, capacitiesSpan := tracer.Start(ctx, "store_capacities")
				capacitiesSpan.SetAttributes(
					attribute.Int("capacity_count", len(detail.CapacityData.Capacities)),
				)

				capacityStorageErr := supabase.StoreNewCapacity(ctx, detail.CapacityData.Capacities, callID)
				if capacityStorageErr != nil {
					capacitiesSpan.RecordError(capacityStorageErr)
					capacitiesSpan.End()
					capacitySpan.RecordError(capacityStorageErr)
					capacitySpan.End()
					log.Error().
						Err(capacityStorageErr).
						Msg("Failed to store capacity details in supa")
					return fmt.Errorf("error storing capacity details: %w", capacityStorageErr)
				}
				capacitiesSpan.End()
			}

			capacitySpan.SetAttributes(
				attribute.Int("units_stored", len(detail.CapacityData.Units)),
				attribute.Int("capacities_stored", len(detail.CapacityData.Capacities)),
			)
			capacitySpan.End()
		}
		// TODO: else if ... other detail categories
	}

	span.SetAttributes(
		attribute.Bool("success", true),
	)
	return nil
}
