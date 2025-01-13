package logger

import (
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
			FormatFieldValue: func(i interface{}) string {
				return strings.ToUpper(fmt.Sprint(i))
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
