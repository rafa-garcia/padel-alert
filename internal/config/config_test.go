package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfig(t *testing.T) {
	os.Setenv("REDIS_URL", "redis://localhost:6379")
	os.Setenv("API_KEYS", "test-key-1,test-key-2")
	os.Setenv("CHECK_INTERVAL", "300")
	os.Setenv("WORKER_COUNT", "5")

	config, err := Load()
	assert.NoError(t, err)

	assert.Equal(t, "redis://localhost:6379", config.RedisURL)
	assert.Contains(t, config.APIKeys, "test-key-1")
	assert.Contains(t, config.APIKeys, "test-key-2")
	assert.Equal(t, 300, config.CheckInterval)
}

func TestDefaultValues(t *testing.T) {
	os.Unsetenv("REDIS_URL")
	os.Unsetenv("API_KEYS")
	os.Unsetenv("CHECK_INTERVAL")
	os.Unsetenv("WORKER_COUNT")

	config, err := Load()
	assert.NoError(t, err)

	assert.Equal(t, "redis://localhost:6379", config.RedisURL)
	assert.Equal(t, 300, config.CheckInterval)
	assert.Equal(t, 10, config.APIRateLimit)
}
