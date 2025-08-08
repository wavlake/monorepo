package config_test

import (
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/wavlake/monorepo/internal/config"
)

var _ = Describe("ServiceConfig", func() {
	var originalEnvValue string
	const envKey = "SERVICE_ACCOUNT_EMAIL"

	BeforeEach(func() {
		// Save original environment value
		originalEnvValue = os.Getenv(envKey)
	})

	AfterEach(func() {
		// Restore original environment value
		if originalEnvValue != "" {
			os.Setenv(envKey, originalEnvValue)
		} else {
			os.Unsetenv(envKey)
		}
	})

	Describe("NewServiceConfig", func() {
		Context("when SERVICE_ACCOUNT_EMAIL environment variable is set", func() {
			It("should use the environment variable value", func() {
				testEmail := "test-service@test-project.iam.gserviceaccount.com"
				os.Setenv(envKey, testEmail)

				serviceConfig := config.NewServiceConfig()

				Expect(serviceConfig).ToNot(BeNil())
				Expect(serviceConfig.ServiceAccountEmail).To(Equal(testEmail))
			})

			It("should handle different valid email formats", func() {
				testCases := []string{
					"service@project.iam.gserviceaccount.com",
					"api-service@wavlake-prod.iam.gserviceaccount.com",
					"backend@test-env.iam.gserviceaccount.com",
					"long-service-name@very-long-project-name.iam.gserviceaccount.com",
				}

				for _, testEmail := range testCases {
					os.Setenv(envKey, testEmail)
					serviceConfig := config.NewServiceConfig()
					
					Expect(serviceConfig.ServiceAccountEmail).To(Equal(testEmail), 
						"Should handle email format: %s", testEmail)
				}
			})

			It("should handle email with special characters", func() {
				testEmail := "api-service-v2@wavlake-alpha-123.iam.gserviceaccount.com"
				os.Setenv(envKey, testEmail)

				serviceConfig := config.NewServiceConfig()

				Expect(serviceConfig.ServiceAccountEmail).To(Equal(testEmail))
			})
		})

		Context("when SERVICE_ACCOUNT_EMAIL environment variable is not set", func() {
			It("should use the default value for backward compatibility", func() {
				os.Unsetenv(envKey)

				serviceConfig := config.NewServiceConfig()

				Expect(serviceConfig).ToNot(BeNil())
				Expect(serviceConfig.ServiceAccountEmail).To(Equal("api-service@wavlake-alpha.iam.gserviceaccount.com"))
			})
		})

		Context("when SERVICE_ACCOUNT_EMAIL environment variable is empty", func() {
			It("should use the default value", func() {
				os.Setenv(envKey, "")

				serviceConfig := config.NewServiceConfig()

				Expect(serviceConfig).ToNot(BeNil())
				Expect(serviceConfig.ServiceAccountEmail).To(Equal("api-service@wavlake-alpha.iam.gserviceaccount.com"))
			})
		})

		Context("when SERVICE_ACCOUNT_EMAIL environment variable contains whitespace", func() {
			It("should preserve whitespace in the email", func() {
				testEmail := " service@project.iam.gserviceaccount.com "
				os.Setenv(envKey, testEmail)

				serviceConfig := config.NewServiceConfig()

				Expect(serviceConfig.ServiceAccountEmail).To(Equal(testEmail))
			})
		})

		Context("when validating configuration structure", func() {
			It("should create a valid ServiceConfig struct", func() {
				serviceConfig := config.NewServiceConfig()

				Expect(serviceConfig).ToNot(BeNil())
				Expect(serviceConfig.ServiceAccountEmail).ToNot(BeEmpty())
			})

			It("should have consistent field access", func() {
				testEmail := "test@project.iam.gserviceaccount.com"
				os.Setenv(envKey, testEmail)

				serviceConfig := config.NewServiceConfig()

				// Test multiple access patterns
				email1 := serviceConfig.ServiceAccountEmail
				email2 := serviceConfig.ServiceAccountEmail

				Expect(email1).To(Equal(email2))
				Expect(email1).To(Equal(testEmail))
			})
		})

		Context("when handling edge cases", func() {
			It("should handle very long email addresses", func() {
				longEmail := "very-long-service-account-name-for-testing-purposes@very-long-project-name-for-testing-purposes.iam.gserviceaccount.com"
				os.Setenv(envKey, longEmail)

				serviceConfig := config.NewServiceConfig()

				Expect(serviceConfig.ServiceAccountEmail).To(Equal(longEmail))
				Expect(len(serviceConfig.ServiceAccountEmail)).To(BeNumerically(">", 50))
			})

			It("should handle minimum length email", func() {
				shortEmail := "a@b.iam.gserviceaccount.com"
				os.Setenv(envKey, shortEmail)

				serviceConfig := config.NewServiceConfig()

				Expect(serviceConfig.ServiceAccountEmail).To(Equal(shortEmail))
			})

			It("should handle emails with numeric components", func() {
				numericEmail := "service123@project456.iam.gserviceaccount.com"
				os.Setenv(envKey, numericEmail)

				serviceConfig := config.NewServiceConfig()

				Expect(serviceConfig.ServiceAccountEmail).To(Equal(numericEmail))
			})

			It("should handle mixed case emails", func() {
				mixedCaseEmail := "Service@Project.IAM.gserviceaccount.com"
				os.Setenv(envKey, mixedCaseEmail)

				serviceConfig := config.NewServiceConfig()

				Expect(serviceConfig.ServiceAccountEmail).To(Equal(mixedCaseEmail))
			})
		})

		Context("when validating business rules", func() {
			It("should maintain immutability after creation", func() {
				testEmail := "original@project.iam.gserviceaccount.com"
				os.Setenv(envKey, testEmail)

				serviceConfig := config.NewServiceConfig()
				originalEmail := serviceConfig.ServiceAccountEmail

				// Change environment variable after creation
				os.Setenv(envKey, "changed@project.iam.gserviceaccount.com")

				// Configuration should remain unchanged
				Expect(serviceConfig.ServiceAccountEmail).To(Equal(originalEmail))
				Expect(serviceConfig.ServiceAccountEmail).To(Equal(testEmail))
			})

			It("should create independent configuration instances", func() {
				os.Setenv(envKey, "first@project.iam.gserviceaccount.com")
				config1 := config.NewServiceConfig()

				os.Setenv(envKey, "second@project.iam.gserviceaccount.com")
				config2 := config.NewServiceConfig()

				Expect(config1.ServiceAccountEmail).To(Equal("first@project.iam.gserviceaccount.com"))
				Expect(config2.ServiceAccountEmail).To(Equal("second@project.iam.gserviceaccount.com"))
				Expect(config1.ServiceAccountEmail).ToNot(Equal(config2.ServiceAccountEmail))
			})

			It("should provide consistent behavior across multiple calls", func() {
				testEmail := "consistent@project.iam.gserviceaccount.com"
				os.Setenv(envKey, testEmail)

				configs := make([]*config.ServiceConfig, 5)
				for i := 0; i < 5; i++ {
					configs[i] = config.NewServiceConfig()
				}

				// All configurations should have the same email
				for i, cfg := range configs {
					Expect(cfg.ServiceAccountEmail).To(Equal(testEmail), 
						"Configuration %d should have consistent email", i)
				}
			})
		})
	})
})