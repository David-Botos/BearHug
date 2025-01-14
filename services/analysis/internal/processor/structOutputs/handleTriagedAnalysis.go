package structOutputs

import (
	"fmt"
	"sync"
)

// HandleTriagedAnalysis takes the triage results and launches appropriate analysis routines
func HandleTriagedAnalysis(
	transcript string,
	serviceDetailsRes map[string]interface{},
	serviceCtx ServiceContext,
) ([]*DetailAnalysisResult, error) {
	detectedCategories, ok := serviceDetailsRes["detected_categories"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid detected_categories format in response")
	}

	var wg sync.WaitGroup
	results := make([]*DetailAnalysisResult, len(detectedCategories))
	errChan := make(chan error, len(detectedCategories))

	// Launch a goroutine for each detected category
	for i, category := range detectedCategories {
		categoryStr, ok := category.(string)
		if !ok {
			continue
		}

		wg.Add(1)
		go func(index int, cat string) {
			defer wg.Done()

			var result DetailAnalysisResult
			var err error

			switch DetailCategory(cat) {
			case CapacityCategory:
				result, err = analyzeCapacityDetails(transcript, serviceCtx)
			// case SchedulingCategory:
			// 	result, err = analyzeSchedulingDetails(transcript, serviceCtx)
			// case ProgramCategory:
			// 	result, err = analyzeProgramDetails(transcript, serviceCtx)
			// case ReqDocsCategory:
			// 	result, err = analyzeReqDocsDetails(transcript, serviceCtx)
			// case ContactCategory:
			// 	result, err = analyzeContactDetails(transcript, serviceCtx)
			default:
				err = fmt.Errorf("unknown category: %s", cat)
			}

			if err != nil {
				errChan <- fmt.Errorf("error analyzing category %s: %w", cat, err)
				return
			}

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
		return nil, fmt.Errorf("multiple errors occurred: %v", errMsgs)
	}

	// Filter out nil results
	finalResults := make([]*DetailAnalysisResult, 0, len(results))
	for _, result := range results {
		if result != nil {
			finalResults = append(finalResults, result)
		}
	}

	return finalResults, nil
}
