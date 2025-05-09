package logger

import (
	"bytes"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestLoggerInitialization(t *testing.T) {
	Init("debug")
	assert.NotNil(t, Logger)
}

func TestInfoLogging(t *testing.T) {
	var buf bytes.Buffer
	originalLogger := Logger
	defer func() { Logger = originalLogger }()

	output := zerolog.ConsoleWriter{Out: &buf, NoColor: true}
	Logger = zerolog.New(output).With().Timestamp().Logger()

	Info("Test message", "key", "value")

	logOutput := buf.String()
	assert.Contains(t, logOutput, "Test message")
	assert.Contains(t, logOutput, "key")
	assert.Contains(t, logOutput, "value")
}

func TestErrorLogging(t *testing.T) {
	var buf bytes.Buffer
	originalLogger := Logger
	defer func() { Logger = originalLogger }()

	output := zerolog.ConsoleWriter{Out: &buf, NoColor: true}
	Logger = zerolog.New(output).With().Timestamp().Logger()

	Error("Error message", nil, "key", "value")

	logOutput := buf.String()
	assert.Contains(t, logOutput, "Error message")
	assert.Contains(t, logOutput, "key")
	assert.Contains(t, logOutput, "value")
}

func TestArgsToMap(t *testing.T) {
	result := argsToMap("key1", "value1", "key2", 123)

	assert.Equal(t, "value1", result["key1"])
	assert.Equal(t, 123, result["key2"])
}
