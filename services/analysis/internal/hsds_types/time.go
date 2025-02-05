package hsds_types

import (
	"encoding/json"
	"fmt"
	"reflect"
	"time"
)

// TimeFormats contains all supported time formats across the system
var TimeFormats = []string{
	time.RFC3339,                       // 2006-01-02T15:04:05Z07:00
	time.RFC3339Nano,                   // 2006-01-02T15:04:05.999999999Z07:00
	"2006-01-02T15:04:05.999999Z07:00", // Supabase format
	"2006-01-02 15:04:05.999999Z07:00", // Common database format
	"2006-01-02 15:04:05.999999+00",    // Alternative timezone format
	"2006-01-02T15:04:05.999999",       // Without timezone
	"2006-01-02 15:04:05.999999",       // Without timezone, space separator
	"2006-01-02T15:04:05",              // ISO without timezone
	"2006-01-02 15:04:05",              // Simple datetime
	"2006-01-02",                       // Date only
	"15:04:05",                         // Time only
}

// StandardTimeFields defines common time field names found in HSDS data
var StandardTimeFields = []string{
	"created_at",
	"updated_at",
	"last_modified",
	"assured_date",
	"valid_from",
	"valid_to",
	"dtstart",
	"until",
	"opens_at",
	"closes_at",
}

// ParseTime attempts to parse a timestamp string using all supported formats
func ParseTime(s string) (time.Time, error) {
	for _, layout := range TimeFormats {
		if t, err := time.Parse(layout, s); err == nil {
			return t.UTC(), nil
		}
	}
	return time.Time{}, fmt.Errorf("could not parse time: %s", s)
}

// UnmarshalJSONWithTime unmarshals JSON data into a slice of any type T
// that contains time.Time fields, handling various time formats
func UnmarshalJSONWithTime[T any](data []byte, result *[]T) error {
	var rawData []map[string]interface{}
	if err := json.Unmarshal(data, &rawData); err != nil {
		return fmt.Errorf("unmarshalling raw data: %w", err)
	}

	*result = make([]T, len(rawData))

	for i, raw := range rawData {
		convertTimeFields(raw)

		jsonData, err := json.Marshal(raw)
		if err != nil {
			return fmt.Errorf("marshalling processed data: %w", err)
		}

		if err := json.Unmarshal(jsonData, &(*result)[i]); err != nil {
			return fmt.Errorf("unmarshalling to target type: %w", err)
		}
	}

	return nil
}

// UnmarshalMultipleJSONResponses unmarshals multiple JSON responses into a single
// slice of type T, deduplicating by ID field
func UnmarshalMultipleJSONResponses[T any](responses [][]byte) ([]T, error) {
	uniqueItems := make(map[string]T)

	for _, data := range responses {
		var items []T
		if err := UnmarshalJSONWithTime(data, &items); err != nil {
			return nil, fmt.Errorf("unmarshalling response: %w", err)
		}

		for _, item := range items {
			id := reflect.ValueOf(item).FieldByName("ID").String()
			uniqueItems[id] = item
		}
	}

	result := make([]T, 0, len(uniqueItems))
	for _, item := range uniqueItems {
		result = append(result, item)
	}

	return result, nil
}

// convertTimeFields recursively processes a map and converts any time string fields
// to proper time.Time format
func convertTimeFields(data map[string]interface{}) {
	for key, value := range data {
		switch v := value.(type) {
		case string:
			for _, tf := range StandardTimeFields {
				if key == tf {
					if t, err := ParseTime(v); err == nil {
						data[key] = t
					}
					break
				}
			}
		case map[string]interface{}:
			convertTimeFields(v)
		}
	}
}
