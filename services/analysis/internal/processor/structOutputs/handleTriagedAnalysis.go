package structOutputs

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/david-botos/BearHug/services/analysis/pkg/logger"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Enhanced metrics channel to handle type-specific data
type analysisMetrics struct {
	category   string
	startTime  time.Time
	duration   time.Duration
	resultType string
	details    map[string]interface{}
}

// HandleTriagedAnalysis takes the triage results and launches appropriate analysis routines
func HandleTriagedAnalysis(
	ctx context.Context,
	transcript string,
	identifiedDetails *IdentifiedDetails,
	serviceCtx ServiceContext,
) ([]*DetailAnalysisResult, error) {
	tracer := otel.GetTracerProvider().Tracer("triage-analysis")
	ctx, span := tracer.Start(ctx, "handle_triage_analysis",
		trace.WithAttributes(
			attribute.Int("transcript_length", len(transcript)),
			attribute.Int("identified_categories", len(identifiedDetails.DetectedCategories)),
			attribute.String("categories", strings.Join(identifiedDetails.DetectedCategories, ",")),
			attribute.Int("existing_services", len(serviceCtx.ExistingServices)),
			attribute.Int("new_services", len(serviceCtx.NewServices)),
		),
	)
	defer span.End()

	log := logger.Get()
	log.Info().Msg("Starting triaged analysis processing")

	detectedCategories := identifiedDetails.DetectedCategories
	results := make([]*DetailAnalysisResult, len(detectedCategories))

	var wg sync.WaitGroup
	errChan := make(chan error, len(detectedCategories))

	metricsChan := make(chan analysisMetrics, len(detectedCategories))

	for i, categoryStr := range detectedCategories {
		wg.Add(1)
		go func(i int, categoryStr string) {
			defer wg.Done()

			categoryCtx, categorySpan := tracer.Start(ctx, "analyze_category",
				trace.WithAttributes(
					attribute.String("category_type", categoryStr),
					attribute.Int("category_index", i),
				),
			)
			defer categorySpan.End()

			metrics := analysisMetrics{
				category:  categoryStr,
				startTime: time.Now(),
				details:   make(map[string]interface{}),
			}

			log := log.With().
				Int("category_index", i).
				Str("category", categoryStr).
				Logger()

			log.Debug().Msg("Beginning category analysis")

			var result DetailAnalysisResult
			var err error

			analysisCtx, analysisSpan := tracer.Start(categoryCtx, fmt.Sprintf("%s_analysis", categoryStr))
			defer analysisSpan.End()

			switch DetailCategory(categoryStr) {
			case CapacityCategory:
				result, err = analyzeCapacityCategoryDetails(analysisCtx, transcript, serviceCtx)
				if err == nil && result.CapacityData != nil {
					// Basic count metrics
					metrics.details["capacity_count"] = len(result.CapacityData.Capacities)
					metrics.details["unit_count"] = len(result.CapacityData.Units)

					// Detailed capacity analysis
					var totalCapacity int
					var totalAvailable float64
					var totalMaximum float64
					var capacitiesWithMaximum int
					var capacitiesWithDescription int
					serviceCapacities := make(map[string]int)
					unitUsage := make(map[string]int)

					// Track the newest and oldest capacity updates
					var newestUpdate time.Time
					var oldestUpdate time.Time

					for _, cap := range result.CapacityData.Capacities {
						if cap != nil {
							totalCapacity++
							serviceCapacities[cap.ServiceID]++
							unitUsage[cap.UnitID]++

							// Aggregate numerical data
							totalAvailable += cap.Available
							if cap.Maximum != nil {
								totalMaximum += *cap.Maximum
								capacitiesWithMaximum++
							}
							if cap.Description != nil {
								capacitiesWithDescription++
							}

							// Track update timeline
							if oldestUpdate.IsZero() || cap.Updated.Before(oldestUpdate) {
								oldestUpdate = cap.Updated
							}
							if cap.Updated.After(newestUpdate) {
								newestUpdate = cap.Updated
							}
						}
					}

					// Unit analysis
					var unitsWithScheme int
					var unitsWithIdentifier int
					var unitsWithURI int
					schemeTypes := make(map[string]int)

					for _, unit := range result.CapacityData.Units {
						if unit != nil {
							if unit.Scheme != nil {
								unitsWithScheme++
								schemeTypes[*unit.Scheme]++
							}
							if unit.Identifier != nil {
								unitsWithIdentifier++
							}
							if unit.URI != nil {
								unitsWithURI++
							}
						}
					}

					// Record detailed metrics
					metrics.details["total_capacity_entries"] = totalCapacity
					metrics.details["services_with_capacity"] = len(serviceCapacities)
					metrics.details["total_available"] = totalAvailable
					metrics.details["average_available"] = totalAvailable / float64(totalCapacity)
					if capacitiesWithMaximum > 0 {
						metrics.details["average_maximum"] = totalMaximum / float64(capacitiesWithMaximum)
					}
					metrics.details["capacities_with_maximum"] = capacitiesWithMaximum
					metrics.details["capacities_with_description"] = capacitiesWithDescription
					metrics.details["unique_units_used"] = len(unitUsage)
					metrics.details["units_with_scheme"] = unitsWithScheme
					metrics.details["units_with_identifier"] = unitsWithIdentifier
					metrics.details["units_with_uri"] = unitsWithURI
					metrics.details["scheme_types"] = len(schemeTypes)

					if !oldestUpdate.IsZero() {
						metrics.details["data_age_hours"] = time.Since(oldestUpdate).Hours()
						metrics.details["update_span_hours"] = newestUpdate.Sub(oldestUpdate).Hours()
					}

					// Record span attributes for key metrics
					analysisSpan.SetAttributes(
						attribute.Int("capacity_count", totalCapacity),
						attribute.Int("unit_count", len(result.CapacityData.Units)),
						attribute.Int("services_with_capacity", len(serviceCapacities)),
						attribute.Int("unique_units_used", len(unitUsage)),
						attribute.Int("capacities_with_maximum", capacitiesWithMaximum),
						attribute.Int("capacities_with_description", capacitiesWithDescription),
						attribute.Float64("total_available", totalAvailable),
						attribute.Float64("average_available", totalAvailable/float64(totalCapacity)),
						attribute.Int("units_with_scheme", unitsWithScheme),
						attribute.Int("scheme_types", len(schemeTypes)),
					)

					// Add update timing attributes if available
					if !oldestUpdate.IsZero() {
						analysisSpan.SetAttributes(
							attribute.Float64("data_age_hours", time.Since(oldestUpdate).Hours()),
							attribute.Float64("update_span_hours", newestUpdate.Sub(oldestUpdate).Hours()),
						)
					}

					schemeTypeKeys := make([]string, 0, len(schemeTypes))
					for scheme := range schemeTypes {
						schemeTypeKeys = append(schemeTypeKeys, scheme)
					}

					// Record high cardinality data in events rather than attributes
					analysisSpan.AddEvent("unit_schemes", trace.WithAttributes(
						attribute.String("schemes", strings.Join(schemeTypeKeys, ",")),
					))
				}
				// Prepared for future categories with specific metrics
				// case SchedulingCategory:
				// 	analysisSpan.SetAttributes(attribute.String("status", "not_implemented"))
				// 	return fmt.Errorf("scheduling analysis not implemented")
				// case ProgramCategory:
				// 	analysisSpan.SetAttributes(attribute.String("status", "not_implemented"))
				// 	return fmt.Errorf("program analysis not implemented")
				// case ReqDocsCategory:
				// 	analysisSpan.SetAttributes(attribute.String("status", "not_implemented"))
				// 	return fmt.Errorf("required documents analysis not implemented")
				// case ContactCategory:
				// 	analysisSpan.SetAttributes(attribute.String("status", "not_implemented"))
				// 	return fmt.Errorf("contact analysis not implemented")
			}

			metrics.duration = time.Since(metrics.startTime)

			if err != nil {
				analysisSpan.RecordError(err)
				analysisSpan.SetAttributes(
					attribute.String("status", "failed"),
					attribute.String("error", err.Error()),
				)
				errChan <- fmt.Errorf("error analyzing category %s: %w", categoryStr, err)
				return
			}

			metrics.resultType = string(result.Category)
			results[i] = &result
			metricsChan <- metrics

			logEvent := log.Debug().
				Str("category", categoryStr).
				Dur("duration", metrics.duration)

			for k, v := range metrics.details {
				logEvent = logEvent.Interface(k, v)
			}

			logEvent.Msg("Category analysis completed")
		}(i, categoryStr)
	}

	// Wait for all goroutines to complete
	wg.Wait()
	close(errChan)
	close(metricsChan)

	// Collect any errors
	var errs []error
	for err := range errChan {
		if err != nil {
			errs = append(errs, err)
		}
	}

	// Handle errors if any occurred
	if len(errs) > 0 {
		span.RecordError(fmt.Errorf("multiple errors occurred: %v", errs))
		return nil, fmt.Errorf("multiple errors occurred: %v", errs)
	}

	// Collect and aggregate metrics
	var totalDuration time.Duration
	aggregatedMetrics := make(map[string]interface{})

	for metric := range metricsChan {
		totalDuration += metric.duration

		// Aggregate metrics
		for k, v := range metric.details {
			if count, ok := v.(int); ok {
				if current, exists := aggregatedMetrics[k]; exists {
					aggregatedMetrics[k] = current.(int) + count
				} else {
					aggregatedMetrics[k] = count
				}
			}
		}
	}

	// Filter and process results
	finalResults := make([]*DetailAnalysisResult, 0, len(results))
	for _, result := range results {
		if result != nil {
			finalResults = append(finalResults, result)
		}
	}

	// Record final metrics
	span.SetAttributes(
		attribute.String("status", "success"),
		attribute.Int("successful_analyses", len(finalResults)),
		attribute.Int("total_categories", len(detectedCategories)),
		attribute.Int64("total_duration_ms", totalDuration.Milliseconds()),
		attribute.Float64("avg_duration_ms", float64(totalDuration.Milliseconds())/float64(len(detectedCategories))),
	)

	// Add aggregated metrics to span
	for k, v := range aggregatedMetrics {
		if count, ok := v.(int); ok {
			span.SetAttributes(attribute.Int(k, count))
		}
	}

	log.Info().
		Int("total_results", len(finalResults)).
		Dur("total_duration", totalDuration).
		Interface("metrics", aggregatedMetrics).
		Msg("Triaged analysis completed successfully")

	return finalResults, nil
}
