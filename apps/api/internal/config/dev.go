package config

import (
	"os"
	"strconv"
	"strings"
)

// DevConfig holds development-specific configuration
type DevConfig struct {
	IsDevelopment     bool
	MockStorage       bool
	MockStoragePath   string
	FileServerURL     string
	LogRequests       bool
	LogResponses      bool
	LogHeaders        bool
	LogRequestBody    bool
	LogResponseBody   bool
	SkipAuth          bool
}

// LoadDevConfig loads development configuration from environment
func LoadDevConfig() DevConfig {
	return DevConfig{
		IsDevelopment:     getBoolEnv("DEVELOPMENT", false),
		MockStorage:       getBoolEnv("MOCK_STORAGE", false),
		MockStoragePath:   getEnv("MOCK_STORAGE_PATH", "./dev-storage"),
		FileServerURL:     getEnv("FILE_SERVER_URL", "http://localhost:8081"),
		LogRequests:       getBoolEnv("LOG_REQUESTS", true),
		LogResponses:      getBoolEnv("LOG_RESPONSES", true),
		LogHeaders:        getBoolEnv("LOG_HEADERS", true),
		LogRequestBody:    getBoolEnv("LOG_REQUEST_BODY", false),
		LogResponseBody:   getBoolEnv("LOG_RESPONSE_BODY", false),
		SkipAuth:          getBoolEnv("SKIP_AUTH", false),
	}
}

// IsFirestoreEmulated checks if we're using the Firestore emulator
func IsFirestoreEmulated() bool {
	return os.Getenv("FIRESTORE_EMULATOR_HOST") != ""
}

// Helper functions
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getBoolEnv(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		val, err := strconv.ParseBool(value)
		if err == nil {
			return val
		}
		// Also handle "yes/no", "on/off", etc.
		value = strings.ToLower(value)
		return value == "yes" || value == "on" || value == "true" || value == "1"
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}