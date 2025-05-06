package config

import (
	"github.com/caarlos0/env/v9"
)

// Config represents the application configuration
type Config struct {
	Port     string   `env:"PORT" envDefault:"8080"`
	APIKeys  []string `env:"API_KEYS" envSeparator:","`
	LogLevel string   `env:"LOG_LEVEL" envDefault:"info"`
}

// Load loads the configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
