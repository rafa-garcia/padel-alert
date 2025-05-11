package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfig(t *testing.T) {
	if err := os.Setenv("REDIS_URL", "redis://localhost:6379"); err != nil {
		t.Fatalf("Failed to set env var: %v", err)
	}
	if err := os.Setenv("API_KEYS", "test-key-1,test-key-2"); err != nil {
		t.Fatalf("Failed to set env var: %v", err)
	}
	if err := os.Setenv("CHECK_INTERVAL", "300"); err != nil {
		t.Fatalf("Failed to set env var: %v", err)
	}
	if err := os.Setenv("WORKER_COUNT", "5"); err != nil {
		t.Fatalf("Failed to set env var: %v", err)
	}

	config, err := Load()
	assert.NoError(t, err)

	assert.Equal(t, "redis://localhost:6379", config.RedisURL)
	assert.Contains(t, config.APIKeys, "test-key-1")
	assert.Contains(t, config.APIKeys, "test-key-2")
	assert.Equal(t, 300, config.CheckInterval)
}

func TestDefaultValues(t *testing.T) {
	if err := os.Unsetenv("REDIS_URL"); err != nil {
		t.Fatalf("Failed to unset env var: %v", err)
	}
	if err := os.Unsetenv("API_KEYS"); err != nil {
		t.Fatalf("Failed to unset env var: %v", err)
	}
	if err := os.Unsetenv("CHECK_INTERVAL"); err != nil {
		t.Fatalf("Failed to unset env var: %v", err)
	}
	if err := os.Unsetenv("WORKER_COUNT"); err != nil {
		t.Fatalf("Failed to unset env var: %v", err)
	}

	config, err := Load()
	assert.NoError(t, err)

	assert.Equal(t, "redis://localhost:6379", config.RedisURL)
	assert.Equal(t, 300, config.CheckInterval)
	assert.Equal(t, 10, config.APIRateLimit)
}
