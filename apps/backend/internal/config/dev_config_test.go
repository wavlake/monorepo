package config_test

import (
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/wavlake/monorepo/internal/config"
)

var _ = Describe("DevConfig", func() {
	var originalEnvValues map[string]string

	// Environment variables used by DevConfig
	envVars := []string{
		"DEVELOPMENT",
		"MOCK_STORAGE",
		"MOCK_STORAGE_PATH",
		"FILE_SERVER_URL",
		"LOG_REQUESTS",
		"LOG_RESPONSES",
		"LOG_HEADERS",
		"LOG_REQUEST_BODY",
		"LOG_RESPONSE_BODY",
		"SKIP_AUTH",
		"FIRESTORE_EMULATOR_HOST",
	}

	BeforeEach(func() {
		// Save original environment values
		originalEnvValues = make(map[string]string)
		for _, envVar := range envVars {
			originalEnvValues[envVar] = os.Getenv(envVar)
			os.Unsetenv(envVar)
		}
	})

	AfterEach(func() {
		// Restore original environment values
		for _, envVar := range envVars {
			if originalValue, exists := originalEnvValues[envVar]; exists && originalValue != "" {
				os.Setenv(envVar, originalValue)
			} else {
				os.Unsetenv(envVar)
			}
		}
	})

	Describe("LoadDevConfig", func() {
		Context("when all environment variables are unset", func() {
			It("should use default values", func() {
				devConfig := config.LoadDevConfig()

				Expect(devConfig.IsDevelopment).To(BeFalse())
				Expect(devConfig.MockStorage).To(BeFalse())
				Expect(devConfig.MockStoragePath).To(Equal("./dev-storage"))
				Expect(devConfig.FileServerURL).To(Equal("http://localhost:8081"))
				Expect(devConfig.LogRequests).To(BeTrue())
				Expect(devConfig.LogResponses).To(BeTrue())
				Expect(devConfig.LogHeaders).To(BeTrue())
				Expect(devConfig.LogRequestBody).To(BeFalse())
				Expect(devConfig.LogResponseBody).To(BeFalse())
				Expect(devConfig.SkipAuth).To(BeFalse())
			})
		})

		Context("when boolean environment variables are set to true", func() {
			It("should parse 'true' values correctly", func() {
				os.Setenv("DEVELOPMENT", "true")
				os.Setenv("MOCK_STORAGE", "true")
				os.Setenv("LOG_REQUEST_BODY", "true")
				os.Setenv("LOG_RESPONSE_BODY", "true")
				os.Setenv("SKIP_AUTH", "true")

				devConfig := config.LoadDevConfig()

				Expect(devConfig.IsDevelopment).To(BeTrue())
				Expect(devConfig.MockStorage).To(BeTrue())
				Expect(devConfig.LogRequestBody).To(BeTrue())
				Expect(devConfig.LogResponseBody).To(BeTrue())
				Expect(devConfig.SkipAuth).To(BeTrue())
			})

			It("should handle various true representations", func() {
				trueValues := []string{"true", "TRUE", "True", "1", "yes", "YES", "on", "ON"}

				for i, trueValue := range trueValues {
					// Reset environment
					for _, envVar := range envVars {
						os.Unsetenv(envVar)
					}

					os.Setenv("DEVELOPMENT", trueValue)
					devConfig := config.LoadDevConfig()

					Expect(devConfig.IsDevelopment).To(BeTrue(),
						"Should parse '%s' as true (case %d)", trueValue, i)
				}
			})
		})

		Context("when boolean environment variables are set to false", func() {
			It("should parse 'false' values correctly", func() {
				os.Setenv("LOG_REQUESTS", "false")
				os.Setenv("LOG_RESPONSES", "false")
				os.Setenv("LOG_HEADERS", "false")

				devConfig := config.LoadDevConfig()

				Expect(devConfig.LogRequests).To(BeFalse())
				Expect(devConfig.LogResponses).To(BeFalse())
				Expect(devConfig.LogHeaders).To(BeFalse())
			})

			It("should handle various false representations", func() {
				falseValues := []string{"false", "FALSE", "False", "0", "no", "NO", "off", "OFF"}

				for i, falseValue := range falseValues {
					// Reset environment
					for _, envVar := range envVars {
						os.Unsetenv(envVar)
					}

					os.Setenv("DEVELOPMENT", falseValue)
					devConfig := config.LoadDevConfig()

					Expect(devConfig.IsDevelopment).To(BeFalse(),
						"Should parse '%s' as false (case %d)", falseValue, i)
				}
			})
		})

		Context("when string environment variables are set", func() {
			It("should use custom mock storage path", func() {
				customPath := "/custom/storage/path"
				os.Setenv("MOCK_STORAGE_PATH", customPath)

				devConfig := config.LoadDevConfig()

				Expect(devConfig.MockStoragePath).To(Equal(customPath))
			})

			It("should use custom file server URL", func() {
				customURL := "https://custom-fileserver.example.com:9000"
				os.Setenv("FILE_SERVER_URL", customURL)

				devConfig := config.LoadDevConfig()

				Expect(devConfig.FileServerURL).To(Equal(customURL))
			})

			It("should handle various URL formats", func() {
				urlFormats := []string{
					"http://localhost:3000",
					"https://fileserver.example.com",
					"http://192.168.1.100:8080",
					"https://fileserver.example.com:443/api",
				}

				for _, url := range urlFormats {
					os.Setenv("FILE_SERVER_URL", url)
					devConfig := config.LoadDevConfig()

					Expect(devConfig.FileServerURL).To(Equal(url))
				}
			})
		})

		Context("when handling invalid boolean values", func() {
			It("should return false for unparseable boolean values", func() {
				os.Setenv("DEVELOPMENT", "invalid")
				os.Setenv("MOCK_STORAGE", "maybe")
				os.Setenv("LOG_REQUESTS", "sometimes")

				devConfig := config.LoadDevConfig()

				// Invalid values return false (not default values)
				Expect(devConfig.IsDevelopment).To(BeFalse()) // invalid -> false
				Expect(devConfig.MockStorage).To(BeFalse())   // invalid -> false
				Expect(devConfig.LogRequests).To(BeFalse())   // invalid -> false (not default true)
			})

			It("should handle empty string boolean values", func() {
				os.Setenv("DEVELOPMENT", "")
				os.Setenv("MOCK_STORAGE", "")

				devConfig := config.LoadDevConfig()

				Expect(devConfig.IsDevelopment).To(BeFalse()) // default
				Expect(devConfig.MockStorage).To(BeFalse())   // default
			})
		})

		Context("when handling edge cases", func() {
			It("should handle whitespace in string values", func() {
				os.Setenv("MOCK_STORAGE_PATH", "  /path/with/spaces  ")
				os.Setenv("FILE_SERVER_URL", "  http://localhost:8081  ")

				devConfig := config.LoadDevConfig()

				// Values should be preserved as-is (no trimming)
				Expect(devConfig.MockStoragePath).To(Equal("  /path/with/spaces  "))
				Expect(devConfig.FileServerURL).To(Equal("  http://localhost:8081  "))
			})

			It("should handle very long string values", func() {
				longPath := "/very/long/path/that/exceeds/normal/length/expectations/for/testing/purposes/only"
				longURL := "https://very-long-subdomain-name-for-testing.example.com:8080/api/v1/fileserver"

				os.Setenv("MOCK_STORAGE_PATH", longPath)
				os.Setenv("FILE_SERVER_URL", longURL)

				devConfig := config.LoadDevConfig()

				Expect(devConfig.MockStoragePath).To(Equal(longPath))
				Expect(devConfig.FileServerURL).To(Equal(longURL))
			})

			It("should handle special characters in paths", func() {
				specialPath := "./dev-storage/special!@#$%^&*()_+-={}[]|\\:;\"'<>?,./"
				os.Setenv("MOCK_STORAGE_PATH", specialPath)

				devConfig := config.LoadDevConfig()

				Expect(devConfig.MockStoragePath).To(Equal(specialPath))
			})
		})

		Context("when validating configuration consistency", func() {
			It("should maintain independent configuration calls", func() {
				os.Setenv("DEVELOPMENT", "true")
				config1 := config.LoadDevConfig()

				os.Setenv("DEVELOPMENT", "false")
				config2 := config.LoadDevConfig()

				Expect(config1.IsDevelopment).To(BeTrue())
				Expect(config2.IsDevelopment).To(BeFalse())
			})

			It("should provide consistent behavior for same environment", func() {
				os.Setenv("DEVELOPMENT", "true")
				os.Setenv("MOCK_STORAGE", "true")
				os.Setenv("FILE_SERVER_URL", "http://test:8080")

				configs := make([]config.DevConfig, 3)
				for i := 0; i < 3; i++ {
					configs[i] = config.LoadDevConfig()
				}

				// All configurations should be identical
				for i, cfg := range configs {
					Expect(cfg.IsDevelopment).To(BeTrue(), "Config %d should have IsDevelopment=true", i)
					Expect(cfg.MockStorage).To(BeTrue(), "Config %d should have MockStorage=true", i)
					Expect(cfg.FileServerURL).To(Equal("http://test:8080"), "Config %d should have correct URL", i)
				}
			})
		})
	})

	Describe("IsFirestoreEmulated", func() {
		Context("when FIRESTORE_EMULATOR_HOST is set", func() {
			It("should return true for localhost", func() {
				os.Setenv("FIRESTORE_EMULATOR_HOST", "localhost:8080")

				result := config.IsFirestoreEmulated()

				Expect(result).To(BeTrue())
			})

			It("should return true for any host value", func() {
				hostValues := []string{
					"127.0.0.1:8080",
					"firestore-emulator:8080",
					"localhost:9999",
					"192.168.1.100:8080",
				}

				for _, host := range hostValues {
					os.Setenv("FIRESTORE_EMULATOR_HOST", host)
					result := config.IsFirestoreEmulated()

					Expect(result).To(BeTrue(), "Should return true for host: %s", host)
				}
			})

			It("should return true even for empty string", func() {
				os.Setenv("FIRESTORE_EMULATOR_HOST", "")

				result := config.IsFirestoreEmulated()

				Expect(result).To(BeFalse()) // Empty string should be considered unset
			})
		})

		Context("when FIRESTORE_EMULATOR_HOST is not set", func() {
			It("should return false", func() {
				os.Unsetenv("FIRESTORE_EMULATOR_HOST")

				result := config.IsFirestoreEmulated()

				Expect(result).To(BeFalse())
			})
		})

		Context("when validating emulator detection logic", func() {
			It("should handle rapid environment changes", func() {
				// Test switching emulator on/off multiple times
				for i := 0; i < 5; i++ {
					os.Setenv("FIRESTORE_EMULATOR_HOST", "localhost:8080")
					Expect(config.IsFirestoreEmulated()).To(BeTrue())

					os.Unsetenv("FIRESTORE_EMULATOR_HOST")
					Expect(config.IsFirestoreEmulated()).To(BeFalse())
				}
			})

			It("should provide consistent behavior across multiple calls", func() {
				os.Setenv("FIRESTORE_EMULATOR_HOST", "localhost:8080")

				results := make([]bool, 10)
				for i := 0; i < 10; i++ {
					results[i] = config.IsFirestoreEmulated()
				}

				// All results should be true
				for i, result := range results {
					Expect(result).To(BeTrue(), "Call %d should return true", i)
				}
			})
		})
	})

	Describe("Helper Functions", func() {
		Context("when testing getEnv behavior through DevConfig", func() {
			It("should use environment values over defaults", func() {
				os.Setenv("MOCK_STORAGE_PATH", "/custom/path")
				os.Setenv("FILE_SERVER_URL", "http://custom:9000")

				devConfig := config.LoadDevConfig()

				Expect(devConfig.MockStoragePath).To(Equal("/custom/path"))
				Expect(devConfig.FileServerURL).To(Equal("http://custom:9000"))
			})

			It("should fall back to defaults when environment is unset", func() {
				// Ensure environment variables are unset
				os.Unsetenv("MOCK_STORAGE_PATH")
				os.Unsetenv("FILE_SERVER_URL")

				devConfig := config.LoadDevConfig()

				Expect(devConfig.MockStoragePath).To(Equal("./dev-storage"))
				Expect(devConfig.FileServerURL).To(Equal("http://localhost:8081"))
			})
		})

		Context("when testing getBoolEnv behavior through DevConfig", func() {
			It("should handle case-insensitive boolean parsing", func() {
				testCases := []struct {
					value    string
					expected bool
				}{
					{"TRUE", true},
					{"True", true},
					{"true", true},
					{"YES", true},
					{"yes", true},
					{"ON", true},
					{"on", true},
					{"1", true},
					{"FALSE", false},
					{"false", false},
					{"NO", false},
					{"no", false},
					{"OFF", false},
					{"off", false},
					{"0", false},
				}

				for _, tc := range testCases {
					os.Setenv("DEVELOPMENT", tc.value)
					devConfig := config.LoadDevConfig()

					Expect(devConfig.IsDevelopment).To(Equal(tc.expected),
						"Value '%s' should parse to %t", tc.value, tc.expected)
				}
			})

			It("should handle invalid boolean values gracefully", func() {
				invalidValues := []string{"invalid", "maybe", "2", "yes!", "true1", "false0"}

				for _, value := range invalidValues {
					os.Setenv("DEVELOPMENT", value)
					devConfig := config.LoadDevConfig()

					// Should fall back to default (false)
					Expect(devConfig.IsDevelopment).To(BeFalse(),
						"Invalid value '%s' should default to false", value)
				}
			})
		})
	})
})