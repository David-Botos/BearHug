package triage

import (
	"fmt"
	"sync"

	"github.com/david-botos/BearHug/services/analysis/internal/hsds_types"
)

// DetailAnalysisResult holds the results of analyzing a specific category of details
type DetailAnalysisResult struct {
	Category DetailCategory
	Error    error
	Data     interface{} // Placeholder for category-specific structured data
}

// analyzeCapacityDetails processes service capacity and unit information
func analyzeCapacityDetails(transcript string, serviceCtx ServiceContext) (*DetailAnalysisResult, error) {
	// TODO: Implement capacity and unit analysis
	return nil, nil
}

// analyzeSchedulingDetails processes service schedule information
func analyzeSchedulingDetails(transcript string, serviceCtx ServiceContext) (*DetailAnalysisResult, error) {
	// TODO: Implement schedule analysis
	return nil, nil
}

// analyzeProgramDetails processes program hierarchy information
func analyzeProgramDetails(transcript string, serviceCtx ServiceContext) (*DetailAnalysisResult, error) {
	// TODO: Implement program analysis
	return nil, nil
}

// analyzeReqDocsDetails processes required document information
func analyzeReqDocsDetails(transcript string, serviceCtx ServiceContext) (*DetailAnalysisResult, error) {
	// TODO: Implement required documents analysis
	return nil, nil
}

// analyzeContactDetails processes contact and phone information
func analyzeContactDetails(transcript string, serviceCtx ServiceContext) (*DetailAnalysisResult, error) {
	// TODO: Implement contact and phone analysis
	return nil, nil
}

// ServiceContext holds both existing and new services for context
type ServiceContext struct {
	ExistingServices []hsds_types.Service
	NewServices      []*hsds_types.Service
}

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

	// Launch a goroutine for each detected category
	for i, category := range detectedCategories {
		categoryStr, ok := category.(string)
		if !ok {
			continue
		}

		wg.Add(1)
		go func(index int, cat string) {
			defer wg.Done()

			var result *DetailAnalysisResult
			var err error

			// Switch on the category to call the appropriate analysis function
			switch DetailCategory(cat) {
			case CapacityCategory:
				result, err = analyzeCapacityDetails(transcript, serviceCtx)
			case SchedulingCategory:
				result, err = analyzeSchedulingDetails(transcript, serviceCtx)
			case ProgramCategory:
				result, err = analyzeProgramDetails(transcript, serviceCtx)
			case ReqDocsCategory:
				result, err = analyzeReqDocsDetails(transcript, serviceCtx)
			case ContactCategory:
				result, err = analyzeContactDetails(transcript, serviceCtx)
			default:
				err = fmt.Errorf("unknown category: %s", cat)
			}

			if err != nil {
				results[index] = &DetailAnalysisResult{
					Category: DetailCategory(cat),
					Error:    err,
				}
				return
			}

			results[index] = result
		}(i, categoryStr)
	}

	// Wait for all goroutines to complete
	wg.Wait()

	// Filter out nil results
	finalResults := make([]*DetailAnalysisResult, 0, len(results))
	for _, result := range results {
		if result != nil {
			finalResults = append(finalResults, result)
		}
	}

	return finalResults, nil
}
