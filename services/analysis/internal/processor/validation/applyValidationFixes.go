package validation

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/david-botos/BearHug/services/analysis/internal/hsds_types"
	"github.com/david-botos/BearHug/services/analysis/internal/processor/structOutputs"
)

// applyFixes applies the fixes suggested by the inference to the data structures
func applyFixes(
	fixes FixOutput,
	affectedServices map[string]*hsds_types.Service,
	affectedCapacities map[string]*hsds_types.ServiceCapacity,
	details []*structOutputs.DetailAnalysisResult,
	serviceCtx structOutputs.ServiceContext,
) ([]*structOutputs.DetailAnalysisResult, structOutputs.ServiceContext, error) {
	// Create copies of our input structures to modify
	newDetails := make([]*structOutputs.DetailAnalysisResult, len(details))
	copy(newDetails, details)

	newServiceCtx := structOutputs.ServiceContext{
		ExistingServices: make([]*hsds_types.Service, len(serviceCtx.ExistingServices)),
		NewServices:      make([]*hsds_types.Service, len(serviceCtx.NewServices)),
	}
	copy(newServiceCtx.ExistingServices, serviceCtx.ExistingServices)
	copy(newServiceCtx.NewServices, serviceCtx.NewServices)

	// Process each fix
	for _, fix := range fixes.Fixes {
		switch fix.Action {
		case "REMOVE":
			if err := applyRemoveFix(fix, &newDetails, &newServiceCtx); err != nil {
				return nil, structOutputs.ServiceContext{}, fmt.Errorf("error applying remove fix: %w", err)
			}

		case "MODIFY":
			if err := applyModifyFix(fix, affectedServices, affectedCapacities, &newDetails, &newServiceCtx); err != nil {
				return nil, structOutputs.ServiceContext{}, fmt.Errorf("error applying modify fix: %w", err)
			}

		case "MERGE":
			if err := applyMergeFix(fix, affectedServices, affectedCapacities, &newDetails, &newServiceCtx); err != nil {
				return nil, structOutputs.ServiceContext{}, fmt.Errorf("error applying merge fix: %w", err)
			}
		}
	}

	return newDetails, newServiceCtx, nil
}

// applyRemoveFix handles removal of hallucinated information
func applyRemoveFix(
	fix Fix,
	details *[]*structOutputs.DetailAnalysisResult,
	serviceCtx *structOutputs.ServiceContext,
) error {
	switch fix.IssueType {
	case "HALLUCINATION":
		// Remove the specified objects by ID
		for _, id := range fix.ObjectIDs {
			// Remove from details if it's a capacity
			for _, detail := range *details {
				if detail.CapacityData != nil {
					// Filter out the capacity
					newCapacities := make([]*hsds_types.ServiceCapacity, 0)
					for _, cap := range detail.CapacityData.Capacities {
						if cap.ID != id {
							newCapacities = append(newCapacities, cap)
						}
					}
					detail.CapacityData.Capacities = newCapacities
				}
			}

			// Remove from services if it's a service
			newServices := make([]*hsds_types.Service, 0)
			for _, svc := range serviceCtx.NewServices {
				if svc.ID != id {
					newServices = append(newServices, svc)
				}
			}
			serviceCtx.NewServices = newServices
		}
	}
	return nil
}

// applyModifyFix handles modification of incorrect information
func applyModifyFix(
	fix Fix,
	affectedServices map[string]*hsds_types.Service,
	affectedCapacities map[string]*hsds_types.ServiceCapacity,
	details *[]*structOutputs.DetailAnalysisResult,
	serviceCtx *structOutputs.ServiceContext,
) error {
	if fix.Modification == nil {
		return fmt.Errorf("modification details missing for MODIFY action")
	}

	for _, id := range fix.ObjectIDs {
		switch fix.IssueType {
		case "HALLUCINATION":
			// Modify service field
			if service, exists := affectedServices[id]; exists {
				if err := setServiceField(service, fix.Modification.Field, fix.Modification.NewValue); err != nil {
					return fmt.Errorf("error modifying service field: %w", err)
				}
			}

			// Modify capacity field
			if capacity, exists := affectedCapacities[id]; exists {
				if err := setCapacityField(capacity, fix.Modification.Field, fix.Modification.NewValue); err != nil {
					return fmt.Errorf("error modifying capacity field: %w", err)
				}
			}
		}
	}
	return nil
}

// applyMergeFix handles merging of duplicate records
func applyMergeFix(
	fix Fix,
	affectedServices map[string]*hsds_types.Service,
	affectedCapacities map[string]*hsds_types.ServiceCapacity,
	details *[]*structOutputs.DetailAnalysisResult,
	serviceCtx *structOutputs.ServiceContext,
) error {
	if fix.KeepID == "" {
		return fmt.Errorf("keep_id missing for MERGE action")
	}

	switch fix.IssueType {
	case "DUPLICATE":
		// Apply field resolutions to the kept record
		if service, exists := affectedServices[fix.KeepID]; exists {
			for _, resolution := range fix.FieldResolutions {
				if err := setServiceField(service, resolution.Field, resolution.Value); err != nil {
					return fmt.Errorf("error applying field resolution: %w", err)
				}
			}
		}

		// Remove other duplicates
		for _, id := range fix.ObjectIDs {
			if id == fix.KeepID {
				continue
			}

			// Remove from new services
			newServices := make([]*hsds_types.Service, 0)
			for _, svc := range serviceCtx.NewServices {
				if svc.ID != id {
					newServices = append(newServices, svc)
				}
			}
			serviceCtx.NewServices = newServices

			// Update any references in capacities
			for _, detail := range *details {
				if detail.CapacityData != nil {
					for _, cap := range detail.CapacityData.Capacities {
						if cap.ServiceID == id {
							cap.ServiceID = fix.KeepID
						}
					}
				}
			}
		}
	}
	return nil
}

// setServiceField sets a field on a Service object using reflection
func setServiceField(service *hsds_types.Service, fieldName, value string) error {
	v := reflect.ValueOf(service).Elem()
	field := v.FieldByName(fieldName)

	if !field.IsValid() {
		return fmt.Errorf("invalid field name: %s", fieldName)
	}

	switch field.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Float64:
		f, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("invalid float value: %s", value)
		}
		field.SetFloat(f)
	// Add other types as needed
	default:
		return fmt.Errorf("unsupported field type: %v", field.Kind())
	}

	return nil
}

// setCapacityField sets a field on a ServiceCapacity object using reflection
func setCapacityField(capacity *hsds_types.ServiceCapacity, fieldName, value string) error {
	v := reflect.ValueOf(capacity).Elem()
	field := v.FieldByName(fieldName)

	if !field.IsValid() {
		return fmt.Errorf("invalid field name: %s", fieldName)
	}

	switch field.Kind() {
	case reflect.Float64:
		f, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("invalid float value: %s", value)
		}
		field.SetFloat(f)
	case reflect.String:
		field.SetString(value)
	// Add other types as needed
	default:
		return fmt.Errorf("unsupported field type: %v", field.Kind())
	}

	return nil
}
