package processor

import (
	"fmt"

	"github.com/david-botos/BearHug/services/analysis/internal/processor/structOutputs"
	"github.com/david-botos/BearHug/services/analysis/internal/types"
	"github.com/david-botos/BearHug/services/analysis/pkg/logger"
)

func ProcessTranscript(params types.ProcTranscriptParams) (bool, error) {
	log := logger.Get()

	log.Info().
		Str("organization_id", params.OrganizationID).
		Int("transcript_length", len(params.Transcript)).
		Msg("Starting transcript processing")

	///* --- Extract services based on the transcript --- *///
	log.Debug().Msg("Beginning service extraction from transcript")
	extractedServices, servicesExtractionErr := structOutputs.ServicesExtraction(params.OrganizationID, params.Transcript)
	if servicesExtractionErr != nil {
		log.Error().
			Err(servicesExtractionErr).
			Str("organization_id", params.OrganizationID).
			Msg("Service extraction failed")
		return false, fmt.Errorf("error with service extraction: %w", servicesExtractionErr)
	}

	///* --- Verify Service Uniqueness -> Upload or Update --- *///
	log.Debug().Msg("Beginning to reason on extracted services compared to DB")
	serviceCtx, serviceUpdateAndUploadErr := structOutputs.HandleExtractedServices(extractedServices, params.OrganizationID, params.CallID)
	if serviceUpdateAndUploadErr != nil {
		log.Error().
			Err(serviceUpdateAndUploadErr).
			Msg("Extracted service reasoning and upload failed")
		return false, fmt.Errorf(`An error occurred when doing reasoning and upload on extracted services: %w`, serviceUpdateAndUploadErr)
	}

	///* --- Identify details for triaged analysis --- *///
	log.Debug().Msg("Beginning to identify what details exist for further triaged analysis")
	identifiedDetailTypes, detailIdentificationErr := structOutputs.IdentifyDetailsForTriagedAnalysis(params.Transcript)
	if detailIdentificationErr != nil {
		log.Error().
			Err(detailIdentificationErr).
			Msg("Failed to identify what details exist in the transcript")
		return false, fmt.Errorf(`an error occurred when identifying details that are present in the transcript for further detailed analysis: %w`, detailIdentificationErr)
	}

	///* --- Conduct Triaged Analyses for Details --- *///
	log.Debug().Msg("Starting detail extraction from triaged analysis")
	extractedDetails, detailExtractionErr := structOutputs.HandleTriagedAnalysis(
		params.Transcript,
		identifiedDetailTypes,
		*serviceCtx,
	)
	if detailExtractionErr != nil {
		log.Error().
			Err(detailExtractionErr).
			Msg("Failed to extract details from triaged analysis")
		return false, fmt.Errorf("error extracting details: %w", detailExtractionErr)
	}

	///* --- Validate the Entire Output --- *///
	// log.Debug().
	// 	Interface("extracted_details", extractedDetails).
	// 	Msg("Starting validation of extracted information")
	// validationResult, validatorErr := validation.ValidateExtractedInfo(extractedDetails, *serviceCtx, params.Transcript, params.CallID)
	// if validatorErr != nil {
	// 	log.Error().
	// 		Err(validatorErr).
	// 		Msg("Validation failed for extracted information")
	// 	return false, fmt.Errorf("error when attempting to validate the information extracted from transcript: %w", validatorErr)
	// }

	// if !validationResult {
	// 	log.Error().Msg("Validation failed with unhandled error")
	// 	return false, fmt.Errorf("validation failed with unhandled error")
	// }

	///* --- Store all the details --- *///
	storageFailureErr := structOutputs.StoreDetails(extractedDetails, params.CallID)
	if storageFailureErr != nil {
		log.Error().
			Err(storageFailureErr).
			Msg("Failed to store details from triaged analysis")
		return false, fmt.Errorf("error storing details: %w", storageFailureErr)
	}

	// TODO: update this to log the complete extraction details
	log.Info().
		Str("organization_id", params.OrganizationID).
		Msg("Successfully completed transcript processing")

	return true, nil
}
