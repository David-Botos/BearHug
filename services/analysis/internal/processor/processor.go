package processor

import (
	"context"
	"fmt"

	"github.com/david-botos/BearHug/services/analysis/internal/processor/structOutputs"
	"github.com/david-botos/BearHug/services/analysis/internal/types"
	"github.com/david-botos/BearHug/services/analysis/pkg/logger"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func ProcessTranscript(traceCtx context.Context, params types.ProcTranscriptParams) (bool, error) {
	tracer := otel.GetTracerProvider().Tracer("transcript-processor")
	traceCtx, rootSpan := tracer.Start(traceCtx, "process_transcript",
		trace.WithAttributes(
			attribute.String("organization_id", params.OrganizationID),
			attribute.String("call_id", params.CallID),
			attribute.Int("transcript_length", len(params.Transcript)),
		),
	)
	defer rootSpan.End()

	log := logger.Get()
	log.Info().
		Str("organization_id", params.OrganizationID).
		Int("transcript_length", len(params.Transcript)).
		Msg("Starting transcript processing")

	///* --- Extract services based on the transcript --- *///
	ctx, servicesSpan := tracer.Start(traceCtx, "extract_services")
	log.Debug().Msg("Beginning service extraction from transcript")
	extractedServices, servicesExtractionErr := structOutputs.ServicesExtraction(ctx, params.OrganizationID, params.Transcript)
	if servicesExtractionErr != nil {
		servicesSpan.RecordError(servicesExtractionErr)
		servicesSpan.End()
		log.Error().
			Err(servicesExtractionErr).
			Str("organization_id", params.OrganizationID).
			Msg("Service extraction failed")
		return false, fmt.Errorf("error with service extraction: %w", servicesExtractionErr)
	}
	servicesSpan.SetAttributes(
		attribute.Int("services_extracted", len(extractedServices.NewServices)),
	)
	servicesSpan.End()

	// /* --- Verify Service Uniqueness -> Upload or Update --- *///
	ctx, serviceHandlingSpan := tracer.Start(ctx, "handle_extracted_services")
	log.Debug().Msg("Beginning to reason on extracted services compared to DB")
	serviceCtx, serviceUpdateAndUploadErr := structOutputs.HandleExtractedServices(ctx, extractedServices, params.OrganizationID, params.CallID)
	if serviceUpdateAndUploadErr != nil {
		serviceHandlingSpan.RecordError(serviceUpdateAndUploadErr)
		serviceHandlingSpan.End()
		log.Error().
			Err(serviceUpdateAndUploadErr).
			Msg("Extracted service reasoning and upload failed")
		return false, fmt.Errorf(`An error occurred when doing reasoning and upload on extracted services: %w`, serviceUpdateAndUploadErr)
	}

	serviceHandlingSpan.SetAttributes(
		attribute.Int("existing_services", len(serviceCtx.ExistingServices)),
		attribute.Int("new_services", len(serviceCtx.NewServices)),
		attribute.Int("total_services", len(serviceCtx.ExistingServices)+len(serviceCtx.NewServices)),
		attribute.Float64("new_services_ratio", float64(len(serviceCtx.NewServices))/float64(len(serviceCtx.ExistingServices)+len(serviceCtx.NewServices))),
	)
	serviceHandlingSpan.End()

	///* --- Identify details for triaged analysis --- *///
	ctx, detailIdentificationSpan := tracer.Start(ctx, "identify_details")
	log.Debug().Msg("Beginning to identify what details exist for further triaged analysis")
	identifiedDetailTypes, detailIdentificationErr := structOutputs.IdentifyDetailsForTriagedAnalysis(ctx, params.Transcript)
	if detailIdentificationErr != nil {
		detailIdentificationSpan.RecordError(detailIdentificationErr)
		detailIdentificationSpan.End()
		log.Error().
			Err(detailIdentificationErr).
			Msg("Failed to identify what details exist in the transcript")
		return false, fmt.Errorf(`an error occurred when identifying details that are present in the transcript for further detailed analysis: %w`, detailIdentificationErr)
	}
	detailIdentificationSpan.SetAttributes(
		attribute.Int("detail_categories_found", len(identifiedDetailTypes.DetectedCategories)),
	)
	detailIdentificationSpan.End()

	// /* --- Conduct Triaged Analyses for Details --- *///
	if len(identifiedDetailTypes.DetectedCategories) > 0 {
		ctx, detailExtractionSpan := tracer.Start(ctx, "extract_details")
		log.Debug().Msg("Starting detail extraction from triaged analysis")
		extractedDetails, detailExtractionErr := structOutputs.HandleTriagedAnalysis(
			ctx,
			params.Transcript,
			identifiedDetailTypes,
			serviceCtx,
		)
		if detailExtractionErr != nil {
			detailExtractionSpan.RecordError(detailExtractionErr)
			detailExtractionSpan.End()
			log.Error().
				Err(detailExtractionErr).
				Msg("Failed to extract details from triaged analysis")
			return false, fmt.Errorf("error extracting details: %w", detailExtractionErr)
		}

		// Count different types of results
		var capacityCount int
		var totalCapacities int
		var totalUnits int
		for _, detail := range extractedDetails {
			if detail.Category == structOutputs.CapacityCategory {
				capacityCount++
				if detail.CapacityData != nil {
					totalCapacities += len(detail.CapacityData.Capacities)
					totalUnits += len(detail.CapacityData.Units)
				}
			}
			// Add counters for other categories as they are implemented
			// if detail.Category == SchedulingCategory { ... }
			// if detail.Category == ProgramCategory { ... }
			// etc.
		}

		detailExtractionSpan.SetAttributes(
			attribute.Int("total_detail_results", len(extractedDetails)),
			attribute.Int("capacity_results", capacityCount),
			attribute.Int("total_service_capacities", totalCapacities),
			attribute.Int("total_units", totalUnits),
			// Future attributes as other categories are implemented:
			// attribute.Int("scheduling_results", schedulingCount),
			// attribute.Int("program_results", programCount),
			// etc.
		)
		detailExtractionSpan.End()

		///* --- Store all the details --- *///
		ctx, storageSpan := tracer.Start(ctx, "store_details")
		storageSpan.SetAttributes(
			attribute.Int("details_to_store", len(extractedDetails)),
			attribute.Int("service_capacities_to_store", totalCapacities),
			attribute.Int("units_to_store", totalUnits),
		)

		storageFailureErr := structOutputs.StoreDetails(ctx, extractedDetails, params.CallID)
		if storageFailureErr != nil {
			storageSpan.RecordError(storageFailureErr)
			storageSpan.End()
			log.Error().
				Err(storageFailureErr).
				Msg("Failed to store all details")
			return false, fmt.Errorf("error storing details: %w", storageFailureErr)
		}
		storageSpan.End()
	}

	rootSpan.SetAttributes(
		attribute.Bool("processing_completed", true),
		attribute.Bool("details_processed", len(identifiedDetailTypes.DetectedCategories) > 0),
	)

	log.Info().Msg("it worked yipeeeeee 🎉🥳")
	return true, nil
}
