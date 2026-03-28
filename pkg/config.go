package pkg

import (
	"os"
	"strconv"
)

// Config holds the application configuration.
type Config struct {
	Port       string
	Timezone   string
	EnableCORS bool
}

// LoadConfig loads configuration from environment variables with sensible defaults.
func LoadConfig() (*Config, error) {
	port := getEnv("PORT", "8081")
	timezone := getEnv("TIMEZONE", "Europe/Paris")
	enableCORS := getEnvBool("ENABLE_CORS", false)

	return &Config{
		Port:       port,
		Timezone:   timezone,
		EnableCORS: enableCORS,
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
