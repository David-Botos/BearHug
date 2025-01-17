package logger

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/rs/zerolog"
)

var (
	logger zerolog.Logger
	once   sync.Once
)

// prettyPrint attempts to format an interface as indented JSON
func prettyPrint(i interface{}) string {
	switch v := i.(type) {
	case []byte:
		// Try to unmarshal and pretty print JSON bytes
		var jsonData interface{}
		if err := json.Unmarshal(v, &jsonData); err == nil {
			// Check if it's a complex type that needs pretty printing
			if isComplexType(jsonData) {
				if pretty, err := json.MarshalIndent(jsonData, "", "  "); err == nil {
					return "\n" + string(pretty)
				}
			}
			// For simple types, just return the string representation
			return fmt.Sprint(jsonData)
		}
		// If not valid JSON, return as string
		return string(v)
	case string:
		// Try to unmarshal and pretty print JSON string
		var jsonData interface{}
		if err := json.Unmarshal([]byte(v), &jsonData); err == nil {
			// Check if it's a complex type that needs pretty printing
			if isComplexType(jsonData) {
				if pretty, err := json.MarshalIndent(jsonData, "", "  "); err == nil {
					return "\n" + string(pretty)
				}
			}
			// For simple types, just return the string representation
			return fmt.Sprint(jsonData)
		}
		return v
	default:
		// For other types, check if they need pretty printing
		if isComplexType(v) {
			if pretty, err := json.MarshalIndent(v, "", "  "); err == nil {
				return "\n" + string(pretty)
			}
		}
		return fmt.Sprint(v)
	}
}

// isComplexType checks if the value is a complex type that needs pretty printing
func isComplexType(v interface{}) bool {
	switch v.(type) {
	case map[string]interface{}, []interface{}, map[interface{}]interface{}:
		return true
	default:
		return false
	}
}

// Init initializes the global logger with custom formatting
func Init() {
	once.Do(func() {
		output := zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: "15:04:05",
			FormatLevel: func(i interface{}) string {
				switch i.(string) {
				case "info":
					return "🟢 INFO"
				case "debug":
					return "🔍 DEBUG"
				case "warn":
					return "⚠️ WARN"
				case "error":
					return "❌ ERROR"
				case "fatal":
					return "💀 FATAL"
				default:
					return " " + strings.ToUpper(fmt.Sprint(i))
				}
			},
			FormatMessage: func(i interface{}) string {
				return fmt.Sprintf("| %s |\n", i)
			},
			FormatFieldName: func(i interface{}) string {
				return fmt.Sprintf("%s: ", i)
			},
			FormatFieldValue: func(i interface{}) string {
				return fmt.Sprintf("%s", prettyPrint(i))
			},
		}

		logger = zerolog.New(output).
			With().
			Timestamp().
			Logger()
	})
}

// Get returns the initialized logger instance
func Get() *zerolog.Logger {
	return &logger
}
