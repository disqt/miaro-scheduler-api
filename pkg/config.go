package pkg

import (
	"os"
	"strconv"
	"time"
)

// Config holds the application configuration.
type Config struct {
	Port          string
	Timezone      string
	ScheduleStart time.Time
	EnableCORS    bool
}

// LoadConfig loads configuration from environment variables with sensible defaults.
func LoadConfig() (*Config, error) {
	port := getEnv("PORT", "8081")
	timezone := getEnv("TIMEZONE", "Europe/Paris")
	enableCORS := getEnvBool("ENABLE_CORS", false)

	// Load timezone
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return nil, err
	}

	// Schedule start date (August 31, 2024 at midnight in the configured timezone)
	scheduleStart := time.Date(2024, time.August, 31, 0, 0, 0, 0, loc)

	return &Config{
		Port:          port,
		Timezone:      timezone,
		ScheduleStart: scheduleStart,
		EnableCORS:    enableCORS,
	}, nil
}

// getEnv retrieves an environment variable or returns a default value.
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvBool retrieves a boolean environment variable or returns a default value.
func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if b, err := strconv.ParseBool(value); err == nil {
			return b
		}
	}
	return defaultValue
}
