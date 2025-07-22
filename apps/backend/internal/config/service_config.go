package config

import "os"

// ServiceConfig holds service-specific configuration
type ServiceConfig struct {
	// ServiceAccountEmail is used for generating presigned URLs
	ServiceAccountEmail string
}

// NewServiceConfig creates a new service configuration from environment
func NewServiceConfig() *ServiceConfig {
	serviceAccountEmail := os.Getenv("SERVICE_ACCOUNT_EMAIL")
	if serviceAccountEmail == "" {
		// Default for backward compatibility - should be overridden in production
		serviceAccountEmail = "api-service@wavlake-alpha.iam.gserviceaccount.com"
	}
	
	return &ServiceConfig{
		ServiceAccountEmail: serviceAccountEmail,
	}
}