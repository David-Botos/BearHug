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
			if pretty, err := json.MarshalIndent(jsonData, "", "  "); err == nil {
				return "\n" + string(pretty)
			}
		}
		// If not valid JSON, return as string
		return string(v)
	case string:
		// Try to unmarshal and pretty print JSON string
		var jsonData interface{}
		if err := json.Unmarshal([]byte(v), &jsonData); err == nil {
			if pretty, err := json.MarshalIndent(jsonData, "", "  "); err == nil {
				return "\n" + string(pretty)
			}
		}
		return v
	default:
		// For other types, try to marshal as JSON
		if pretty, err := json.MarshalIndent(v, "", "  "); err == nil {
			return "\n" + string(pretty)
		}
		return strings.ToUpper(fmt.Sprint(i))
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
				return fmt.Sprintf("| %s |", i)
			},
			FormatFieldName: func(i interface{}) string {
				return fmt.Sprintf("%s:", i)
			},
			FormatFieldValue: prettyPrint,
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
