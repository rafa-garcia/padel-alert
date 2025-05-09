package config

import (
	"github.com/caarlos0/env/v9"
)

// Config represents the application configuration
type Config struct {
	// HTTP server settings
	Port     string   `env:"PORT" envDefault:"8080"`
	APIKeys  []string `env:"API_KEYS" envSeparator:","`
	LogLevel string   `env:"LOG_LEVEL" envDefault:"info"`

	// Redis settings
	RedisURL string `env:"REDIS_URL" envDefault:"redis://localhost:6379"`

	// Scheduler configuration
	CheckInterval int `env:"CHECK_INTERVAL" envDefault:"300"` // Seconds between checks

	// API Rate Limiting
	APIRateLimit int `env:"API_RATE_LIMIT" envDefault:"10"` // Requests per minute

	// Email settings
	SMTPServer   string `env:"SMTP_SERVER"`
	SMTPPort     int    `env:"SMTP_PORT" envDefault:"587"`
	SMTPUsername string `env:"SMTP_USERNAME"`
	SMTPPassword string `env:"SMTP_PASSWORD"`
	SMTPSender   string `env:"SMTP_SENDER"`
}

// Load loads the configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
