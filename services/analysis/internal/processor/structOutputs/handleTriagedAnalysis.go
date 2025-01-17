package structOutputs

import (
	"fmt"
	"sync"

	"github.com/david-botos/BearHug/services/analysis/pkg/logger"
)

// HandleTriagedAnalysis takes the triage results and launches appropriate analysis routines
func HandleTriagedAnalysis(
	transcript string,
	identifiedDetails *IdentifiedDetails,
	serviceCtx ServiceContext,
) ([]*DetailAnalysisResult, error) {
	log := logger.Get()

	log.Info().Msg("Entered HandleTriagedAnalysis")

	// Log input data
	log.Debug().
		Interface("transcript", transcript).
		Int("num_identified_detail_categories", len(identifiedDetails.DetectedCategories)).
		Int("existing_services_count", len(serviceCtx.ExistingServices)).
		Int("new_services_count", len(serviceCtx.NewServices)).
		Msg("Input data state")

	log.Debug().
		Str("transcript_length", fmt.Sprint(len(transcript))).
		Msg("Starting triage analysis")

	if identifiedDetails == nil || len(identifiedDetails.DetectedCategories) == 0 {
		log.Error().Msg("No categories detected in response")
		return nil, fmt.Errorf("no categories detected in response")
	}

	detectedCategories := identifiedDetails.DetectedCategories
	log.Info().
		Interface("categories", detectedCategories).
		Msg("Processing detected categories")

	var wg sync.WaitGroup
	results := make([]*DetailAnalysisResult, len(detectedCategories))
	errChan := make(chan error, len(detectedCategories))

	// Launch a goroutine for each detected category
	for i, categoryStr := range detectedCategories {
		wg.Add(1)
		go func(index int, cat string) {
			defer wg.Done()

			log := log.With().
				Int("category_index", index).
				Str("category", cat).
				Logger()

			log.Debug().Msg("Starting category analysis")

			var result DetailAnalysisResult
			var err error

			switch DetailCategory(cat) {
			case CapacityCategory:
				result, err = analyzeCapacityCategoryDetails(transcript, serviceCtx)
			// case SchedulingCategory:
			//     result, err = analyzeSchedulingDetails(transcript, serviceCtx)
			// case ProgramCategory:
			//     result, err = analyzeProgramDetails(transcript, serviceCtx)
			// case ReqDocsCategory:
			//     result, err = analyzeReqDocsDetails(transcript, serviceCtx)
			// case ContactCategory:
			//     result, err = analyzeContactDetails(transcript, serviceCtx)
			default:
				err = fmt.Errorf("unknown category: %s", cat)
			}

			if err != nil {
				wrappedErr := fmt.Errorf("error analyzing category %s: %w", cat, err)
				log.Error().Err(wrappedErr).Msg("Category analysis failed")
				errChan <- wrappedErr
				return
			}

			log.Debug().Msg("Category analysis completed successfully")
			results[index] = &result
		}(i, categoryStr)
	}

	// Wait for all goroutines to complete
	wg.Wait()
	close(errChan)

	// Check for any errors
	var errs []error
	for err := range errChan {
		errs = append(errs, err)
	}
	if len(errs) > 0 {
		// Combine all errors into a single error message
		errMsgs := make([]string, len(errs))
		for i, err := range errs {
			errMsgs[i] = err.Error()
		}
		log.Error().
			Interface("errors", errMsgs).
			Msg("Multiple errors occurred during analysis")
		return nil, fmt.Errorf("multiple errors occurred: %v", errMsgs)
	}

	// Filter out nil results
	finalResults := make([]*DetailAnalysisResult, 0, len(results))
	for _, result := range results {
		if result != nil {
			finalResults = append(finalResults, result)
		}
	}

	log.Info().
		Int("total_results", len(finalResults)).
		Msg("Triage analysis completed successfully")

	return finalResults, nil
}
