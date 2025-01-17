package structOutputs

import "github.com/david-botos/BearHug/services/analysis/internal/hsds_types"

// DetailCategory represents the high-level categories of related tables
type DetailCategory string

const (
	CapacityCategory   DetailCategory = "CAPACITY"
	SchedulingCategory DetailCategory = "SCHEDULING"
	ProgramCategory    DetailCategory = "PROGRAM"
	ReqDocsCategory    DetailCategory = "REQDOCS"
	ContactCategory    DetailCategory = "CONTACT"
)

// CategoryDescription contains information about what tables and data belong in each category
type CategoryDescription struct {
	Category    DetailCategory
	Tables      []TableName
	Description string
}

// ServiceContext holds both existing and new services for context
type ServiceContext struct {
	ExistingServices []*hsds_types.Service
	NewServices      []*hsds_types.Service
}

// CapacityResult represents the specific data structure for capacity analysis results
type CapacityResult struct {
	Capacities []*hsds_types.ServiceCapacity
	Units      []*hsds_types.Unit
}

// DetailAnalysisResult holds the results of analyzing a specific category of details
type DetailAnalysisResult struct {
	Category DetailCategory

	// Type-specific results
	CapacityData *CapacityResult
	// Add other category-specific fields as they are implemented
	// SchedulingData *SchedulingResult
	// ProgramData    *ProgramResult
	// ReqDocsData    *ReqDocsResult
	// ContactData    *ContactResult
}

// NewCapacityResult creates a new DetailAnalysisResult for capacity data
func NewCapacityResult(capacities []*hsds_types.ServiceCapacity, units []*hsds_types.Unit) *DetailAnalysisResult {
	return &DetailAnalysisResult{
		Category: CapacityCategory,
		CapacityData: &CapacityResult{
			Capacities: capacities,
			Units:      units,
		},
	}
}

// As other categories are implemented, add their corresponding result types and constructor functions:
/*
	type SchedulingResult struct {
		// Add scheduling-specific fields
	}

	func NewSchedulingResult(...) *DetailAnalysisResult {
		return &DetailAnalysisResult{
			Category: SchedulingCategory,
			SchedulingData: &SchedulingResult{...},
		}
	}
*/
