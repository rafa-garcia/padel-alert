package logger

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
)

// Logger is the global logger
var Logger zerolog.Logger

// Init initializes the logger with the specified level
func Init(level string) {
	// Parse log level
	logLevel, err := zerolog.ParseLevel(level)
	if err != nil {
		logLevel = zerolog.InfoLevel
	}

	// Set global level
	zerolog.SetGlobalLevel(logLevel)

	// Configure console writer with color and timestamp formatting
	output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
	output.FormatLevel = func(i interface{}) string {
		return fmt.Sprintf("| %-6s|", i)
	}
	output.FormatMessage = func(i interface{}) string {
		return fmt.Sprintf(" %s |", i)
	}
	output.FormatFieldName = func(i interface{}) string {
		return fmt.Sprintf(" %s:", i)
	}
	output.FormatFieldValue = func(i interface{}) string {
		return fmt.Sprintf("%s", i)
	}

	// Initialize the logger
	Logger = zerolog.New(output).With().Timestamp().Caller().Logger()

	Logger.Info().Str("level", logLevel.String()).Msg("Logger initialized")
}

// Output returns the logger output for use with other libraries
func Output() io.Writer {
	return Logger
}

// Debug logs a debug message
func Debug(msg string, args ...interface{}) {
	Logger.Debug().Fields(argsToMap(args...)).Msg(msg)
}

// Info logs an info message
func Info(msg string, args ...interface{}) {
	Logger.Info().Fields(argsToMap(args...)).Msg(msg)
}

// Warn logs a warning message
func Warn(msg string, args ...interface{}) {
	Logger.Warn().Fields(argsToMap(args...)).Msg(msg)
}

// Error logs an error message
func Error(msg string, err error, args ...interface{}) {
	logEvent := Logger.Error().Fields(argsToMap(args...))
	if err != nil {
		logEvent = logEvent.Err(err)
	}
	logEvent.Msg(msg)
}

// Fatal logs a fatal message and exits
func Fatal(msg string, err error, args ...interface{}) {
	logEvent := Logger.Fatal().Fields(argsToMap(args...))
	if err != nil {
		logEvent = logEvent.Err(err)
	}
	logEvent.Msg(msg)
}

// argsToMap converts key/value pairs to a map for structured logging
// It expects args to be in pairs: key1, value1, key2, value2, ...
func argsToMap(args ...interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for i := 0; i < len(args); i += 2 {
		if i+1 < len(args) {
			if key, ok := args[i].(string); ok {
				result[key] = args[i+1]
			}
		}
	}
	return result
}
