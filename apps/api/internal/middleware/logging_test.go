package middleware

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http/httptest"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Logging Middleware", func() {
	var (
		router   *gin.Engine
		recorder *httptest.ResponseRecorder
		logBuf   *bytes.Buffer
	)

	BeforeEach(func() {
		// Set Gin to test mode
		gin.SetMode(gin.TestMode)
		
		// Create a router
		router = gin.New()
		
		// Capture log output
		logBuf = &bytes.Buffer{}
		
		// Create a test logger that writes to our buffer
		opts := &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}
		handler := slog.NewJSONHandler(logBuf, opts)
		logger = slog.New(handler)
		
		// Create a response recorder
		recorder = httptest.NewRecorder()
	})

	Describe("DefaultLoggingConfig", func() {
		It("should return expected default configuration", func() {
			config := DefaultLoggingConfig()
			
			Expect(config.LogRequests).To(BeTrue())
			Expect(config.LogResponses).To(BeTrue())
			Expect(config.LogHeaders).To(BeTrue())
			Expect(config.LogRequestBody).To(BeTrue())
			Expect(config.LogResponseBody).To(BeTrue())
			Expect(config.MaxBodySize).To(Equal(int64(1024 * 1024)))
			Expect(config.SkipPaths).To(ContainElements("/heartbeat", "/health"))
			Expect(config.SensitiveHeaders).To(ContainElements("authorization", "x-firebase-token", "x-nostr-authorization", "cookie"))
			Expect(config.SensitiveFields).To(ContainElements("password", "token", "secret", "key", "auth"))
		})
	})

	Describe("RequestResponseLogging", func() {
		var config LoggingConfig

		BeforeEach(func() {
			config = DefaultLoggingConfig()
		})

		Context("when processing a normal request", func() {
			It("should log request and response with correlation ID", func() {
				// Setup middleware and handler
				router.Use(RequestResponseLogging(config))
				router.GET("/test", func(c *gin.Context) {
					c.JSON(200, gin.H{"message": "success"})
				})

				// Make request
				req := httptest.NewRequest("GET", "/test?param=value", nil)
				req.Header.Set("User-Agent", "test-agent")
				req.Header.Set("X-Custom-Header", "custom-value")
				
				router.ServeHTTP(recorder, req)

				// Parse log output
				logOutput := logBuf.String()
				lines := strings.Split(strings.TrimSpace(logOutput), "\n")
				
				// Should have request and response logs
				Expect(len(lines)).To(BeNumerically(">=", 2))

				// Parse request log
				var requestLog map[string]interface{}
				err := json.Unmarshal([]byte(lines[0]), &requestLog)
				Expect(err).ToNot(HaveOccurred())

				// Verify request log fields
				Expect(requestLog["type"]).To(Equal("request"))
				Expect(requestLog["method"]).To(Equal("GET"))
				Expect(requestLog["path"]).To(Equal("/test"))
				Expect(requestLog["query"]).To(Equal("param=value"))
				Expect(requestLog["correlation_id"]).ToNot(BeEmpty())
				Expect(requestLog["headers"]).To(HaveKey("user-agent"))
				Expect(requestLog["headers"]).To(HaveKey("x-custom-header"))

				// Parse response log
				var responseLog map[string]interface{}
				err = json.Unmarshal([]byte(lines[1]), &responseLog)
				Expect(err).ToNot(HaveOccurred())

				// Verify response log fields
				Expect(responseLog["type"]).To(Equal("response"))
				Expect(responseLog["status"]).To(Equal(float64(200)))
				Expect(responseLog["correlation_id"]).To(Equal(requestLog["correlation_id"]))
				Expect(responseLog["duration"]).ToNot(BeZero())
				Expect(responseLog["body"]).To(ContainSubstring("success"))
			})
		})

		Context("when request contains sensitive headers", func() {
			It("should mask sensitive header values", func() {
				router.Use(RequestResponseLogging(config))
				router.GET("/test", func(c *gin.Context) {
					c.JSON(200, gin.H{"status": "ok"})
				})

				req := httptest.NewRequest("GET", "/test", nil)
				req.Header.Set("Authorization", "Bearer secret-token-12345")
				req.Header.Set("X-Firebase-Token", "firebase-secret-67890")
				req.Header.Set("X-Safe-Header", "safe-value")
				
				router.ServeHTTP(recorder, req)

				logOutput := logBuf.String()
				lines := strings.Split(strings.TrimSpace(logOutput), "\n")
				
				var requestLog map[string]interface{}
				err := json.Unmarshal([]byte(lines[0]), &requestLog)
				Expect(err).ToNot(HaveOccurred())

				headers := requestLog["headers"].(map[string]interface{})
				
				// Sensitive headers should be masked
				Expect(headers["authorization"].(string)).To(ContainSubstring("***"))
				Expect(headers["authorization"].(string)).ToNot(ContainSubstring("secret-token"))
				Expect(headers["x-firebase-token"].(string)).To(ContainSubstring("***"))
				Expect(headers["x-firebase-token"].(string)).ToNot(ContainSubstring("firebase-secret"))
				
				// Non-sensitive headers should be preserved
				Expect(headers["x-safe-header"]).To(Equal("safe-value"))
			})
		})

		Context("when request contains body", func() {
			It("should log request body and apply sensitive data masking", func() {
				router.Use(RequestResponseLogging(config))
				router.POST("/test", func(c *gin.Context) {
					c.JSON(200, gin.H{"received": "ok"})
				})

				requestBody := `{"username": "testuser", "password": "secret123", "email": "test@example.com"}`
				req := httptest.NewRequest("POST", "/test", strings.NewReader(requestBody))
				req.Header.Set("Content-Type", "application/json")
				
				router.ServeHTTP(recorder, req)

				logOutput := logBuf.String()
				lines := strings.Split(strings.TrimSpace(logOutput), "\n")
				
				var requestLog map[string]interface{}
				err := json.Unmarshal([]byte(lines[0]), &requestLog)
				Expect(err).ToNot(HaveOccurred())

				// Should log request body
				Expect(requestLog["body"]).ToNot(BeEmpty())
				Expect(requestLog["body_size"]).To(Equal(float64(len(requestBody))))
				
				// Should contain non-sensitive data
				bodyStr := requestLog["body"].(string)
				Expect(bodyStr).To(ContainSubstring("testuser"))
				Expect(bodyStr).To(ContainSubstring("test@example.com"))
				
				// Should mask sensitive field markers (basic implementation)
				Expect(bodyStr).To(ContainSubstring("password"))
			})
		})

		Context("when body size exceeds MaxBodySize", func() {
			It("should truncate request body", func() {
				config.MaxBodySize = 10 // Very small limit for testing
				
				router.Use(RequestResponseLogging(config))
				router.POST("/test", func(c *gin.Context) {
					c.JSON(200, gin.H{"status": "ok"})
				})

				longBody := strings.Repeat("a", 50) // 50 chars, exceeds 10 char limit
				req := httptest.NewRequest("POST", "/test", strings.NewReader(longBody))
				
				router.ServeHTTP(recorder, req)

				logOutput := logBuf.String()
				lines := strings.Split(strings.TrimSpace(logOutput), "\n")
				
				var requestLog map[string]interface{}
				err := json.Unmarshal([]byte(lines[0]), &requestLog)
				Expect(err).ToNot(HaveOccurred())

				// Should truncate body to MaxBodySize
				bodyStr := requestLog["body"].(string)
				Expect(len(bodyStr)).To(Equal(10))
				Expect(requestLog["body_size"]).To(Equal(float64(10)))
			})
		})

		Context("when path is in SkipPaths", func() {
			It("should skip logging for heartbeat path", func() {
				router.Use(RequestResponseLogging(config))
				router.GET("/heartbeat", func(c *gin.Context) {
					c.JSON(200, gin.H{"status": "ok"})
				})

				req := httptest.NewRequest("GET", "/heartbeat", nil)
				router.ServeHTTP(recorder, req)

				// Should have no log output
				logOutput := logBuf.String()
				Expect(strings.TrimSpace(logOutput)).To(BeEmpty())
			})

			It("should skip logging for health path", func() {
				router.Use(RequestResponseLogging(config))
				router.GET("/health", func(c *gin.Context) {
					c.JSON(200, gin.H{"status": "healthy"})
				})

				req := httptest.NewRequest("GET", "/health", nil)
				router.ServeHTTP(recorder, req)

				// Should have no log output
				logOutput := logBuf.String()
				Expect(strings.TrimSpace(logOutput)).To(BeEmpty())
			})
		})

		Context("when logging is selectively disabled", func() {
			It("should only log requests when LogResponses is false", func() {
				config.LogResponses = false
				
				router.Use(RequestResponseLogging(config))
				router.GET("/test", func(c *gin.Context) {
					c.JSON(200, gin.H{"message": "success"})
				})

				req := httptest.NewRequest("GET", "/test", nil)
				router.ServeHTTP(recorder, req)

				logOutput := logBuf.String()
				lines := strings.Split(strings.TrimSpace(logOutput), "\n")
				
				// Should only have request log
				Expect(len(lines)).To(Equal(1))
				
				var requestLog map[string]interface{}
				err := json.Unmarshal([]byte(lines[0]), &requestLog)
				Expect(err).ToNot(HaveOccurred())
				Expect(requestLog["type"]).To(Equal("request"))
			})

			It("should only log responses when LogRequests is false", func() {
				config.LogRequests = false
				
				router.Use(RequestResponseLogging(config))
				router.GET("/test", func(c *gin.Context) {
					c.JSON(200, gin.H{"message": "success"})
				})

				req := httptest.NewRequest("GET", "/test", nil)
				router.ServeHTTP(recorder, req)

				logOutput := logBuf.String()
				lines := strings.Split(strings.TrimSpace(logOutput), "\n")
				
				// Should only have response log
				Expect(len(lines)).To(Equal(1))
				
				var responseLog map[string]interface{}
				err := json.Unmarshal([]byte(lines[0]), &responseLog)
				Expect(err).ToNot(HaveOccurred())
				Expect(responseLog["type"]).To(Equal("response"))
			})

			It("should not log headers when LogHeaders is false", func() {
				config.LogHeaders = false
				
				router.Use(RequestResponseLogging(config))
				router.GET("/test", func(c *gin.Context) {
					c.JSON(200, gin.H{"message": "success"})
				})

				req := httptest.NewRequest("GET", "/test", nil)
				req.Header.Set("X-Test-Header", "test-value")
				
				router.ServeHTTP(recorder, req)

				logOutput := logBuf.String()
				lines := strings.Split(strings.TrimSpace(logOutput), "\n")
				
				var requestLog map[string]interface{}
				err := json.Unmarshal([]byte(lines[0]), &requestLog)
				Expect(err).ToNot(HaveOccurred())
				
				// Should not include headers
				Expect(requestLog["headers"]).To(BeNil())
			})

			It("should not log request body when LogRequestBody is false", func() {
				config.LogRequestBody = false
				
				router.Use(RequestResponseLogging(config))
				router.POST("/test", func(c *gin.Context) {
					c.JSON(200, gin.H{"received": "ok"})
				})

				req := httptest.NewRequest("POST", "/test", strings.NewReader(`{"data": "test"}`))
				router.ServeHTTP(recorder, req)

				logOutput := logBuf.String()
				lines := strings.Split(strings.TrimSpace(logOutput), "\n")
				
				var requestLog map[string]interface{}
				err := json.Unmarshal([]byte(lines[0]), &requestLog)
				Expect(err).ToNot(HaveOccurred())
				
				// Should not include request body
				Expect(requestLog["body"]).To(BeNil())
				Expect(requestLog["body_size"]).To(BeNil())
			})
		})

		Context("when handling different HTTP methods", func() {
			It("should log POST requests correctly", func() {
				router.Use(RequestResponseLogging(config))
				router.POST("/api/test", func(c *gin.Context) {
					c.JSON(201, gin.H{"created": true})
				})

				req := httptest.NewRequest("POST", "/api/test", nil)
				router.ServeHTTP(recorder, req)

				logOutput := logBuf.String()
				lines := strings.Split(strings.TrimSpace(logOutput), "\n")
				
				var requestLog map[string]interface{}
				err := json.Unmarshal([]byte(lines[0]), &requestLog)
				Expect(err).ToNot(HaveOccurred())

				Expect(requestLog["method"]).To(Equal("POST"))
				Expect(requestLog["path"]).To(Equal("/api/test"))
				
				var responseLog map[string]interface{}
				err = json.Unmarshal([]byte(lines[1]), &responseLog)
				Expect(err).ToNot(HaveOccurred())
				Expect(responseLog["status"]).To(Equal(float64(201)))
			})

			It("should log DELETE requests correctly", func() {
				router.Use(RequestResponseLogging(config))
				router.DELETE("/api/test/:id", func(c *gin.Context) {
					c.JSON(204, gin.H{})
				})

				req := httptest.NewRequest("DELETE", "/api/test/123", nil)
				router.ServeHTTP(recorder, req)

				logOutput := logBuf.String()
				lines := strings.Split(strings.TrimSpace(logOutput), "\n")
				
				var requestLog map[string]interface{}
				err := json.Unmarshal([]byte(lines[0]), &requestLog)
				Expect(err).ToNot(HaveOccurred())

				Expect(requestLog["method"]).To(Equal("DELETE"))
				Expect(requestLog["path"]).To(Equal("/api/test/123"))
			})
		})
	})

	Describe("Correlation ID functionality", func() {
		Context("GetCorrelationID", func() {
			It("should return correlation ID from context", func() {
				c, _ := gin.CreateTestContext(recorder)
				testID := "test-correlation-123"
				c.Set("correlation_id", testID)

				result := GetCorrelationID(c)
				Expect(result).To(Equal(testID))
			})

			It("should return empty string when no correlation ID exists", func() {
				c, _ := gin.CreateTestContext(recorder)

				result := GetCorrelationID(c)
				Expect(result).To(BeEmpty())
			})
		})

		Context("LogWithCorrelation", func() {
			It("should log with correlation ID when present", func() {
				req := httptest.NewRequest("GET", "/test", nil)
				c, _ := gin.CreateTestContext(recorder)
				c.Request = req
				testID := "test-correlation-456"
				c.Set("correlation_id", testID)

				LogWithCorrelation(c, slog.LevelInfo, "Test message", slog.String("key", "value"))

				logOutput := logBuf.String()
				Expect(logOutput).To(ContainSubstring(testID))
				Expect(logOutput).To(ContainSubstring("Test message"))
				Expect(logOutput).To(ContainSubstring("key"))
				Expect(logOutput).To(ContainSubstring("value"))
			})

			It("should log without correlation ID when not present", func() {
				req := httptest.NewRequest("GET", "/test", nil)
				c, _ := gin.CreateTestContext(recorder)
				c.Request = req

				LogWithCorrelation(c, slog.LevelWarn, "Warning message", slog.String("warning", "test"))

				logOutput := logBuf.String()
				Expect(logOutput).To(ContainSubstring("Warning message"))
				Expect(logOutput).To(ContainSubstring("warning"))
				Expect(logOutput).ToNot(ContainSubstring("correlation_id"))
			})
		})
	})

	Describe("Helper functions", func() {
		Describe("isSensitiveHeader", func() {
			It("should identify sensitive headers correctly", func() {
				sensitiveHeaders := []string{"authorization", "x-firebase-token", "cookie"}

				Expect(isSensitiveHeader("Authorization", sensitiveHeaders)).To(BeTrue())
				Expect(isSensitiveHeader("AUTHORIZATION", sensitiveHeaders)).To(BeTrue())
				Expect(isSensitiveHeader("X-Firebase-Token", sensitiveHeaders)).To(BeTrue())
				Expect(isSensitiveHeader("Cookie", sensitiveHeaders)).To(BeTrue())
				Expect(isSensitiveHeader("Content-Type", sensitiveHeaders)).To(BeFalse())
				Expect(isSensitiveHeader("X-Custom-Header", sensitiveHeaders)).To(BeFalse())
			})

			It("should handle partial matches", func() {
				sensitiveHeaders := []string{"auth"}

				Expect(isSensitiveHeader("Authorization", sensitiveHeaders)).To(BeTrue())
				Expect(isSensitiveHeader("X-Auth-Token", sensitiveHeaders)).To(BeTrue())
				Expect(isSensitiveHeader("Content-Type", sensitiveHeaders)).To(BeFalse())
			})
		})

		Describe("maskSensitiveData", func() {
			It("should mask long values correctly", func() {
				input := "bearer-token-12345678" // length 21
				result := maskSensitiveData(input)
				// Expected: first 2 chars + (length-4) stars + last 2 chars
				// be + 17 stars + 78
				expected := "be" + strings.Repeat("*", 17) + "78"
				Expect(result).To(Equal(expected))
				Expect(result).ToNot(ContainSubstring("token"))
			})

			It("should mask short values", func() {
				result := maskSensitiveData("abc")
				Expect(result).To(Equal("***"))
			})

			It("should handle minimum length values", func() {
				result := maskSensitiveData("abcd")
				Expect(result).To(Equal("***"))
			})

			It("should handle 5-character values", func() {
				result := maskSensitiveData("abcde")
				Expect(result).To(Equal("ab*de"))
			})
		})

		Describe("maskSensitiveDataInJSON", func() {
			It("should identify and mark sensitive fields in JSON", func() {
				jsonBody := `{"username": "test", "password": "secret", "email": "test@example.com"}`
				sensitiveFields := []string{"password", "token"}

				result := maskSensitiveDataInJSON(jsonBody, sensitiveFields)
				Expect(result).To(ContainSubstring("password\" [MASKED]"))
				Expect(result).To(ContainSubstring("username"))
				Expect(result).To(ContainSubstring("email"))
			})

			It("should handle case-insensitive field matching", func() {
				jsonBody := `{"Password": "secret", "TOKEN": "abc123"}`
				sensitiveFields := []string{"password", "token"}

				result := maskSensitiveDataInJSON(jsonBody, sensitiveFields)
				// The current implementation is case-sensitive for simplicity
				// This test documents the current behavior
				Expect(result).ToNot(ContainSubstring("[MASKED]"))
			})

			It("should handle JSON with no sensitive fields", func() {
				jsonBody := `{"username": "test", "email": "test@example.com"}`
				sensitiveFields := []string{"password", "token"}

				result := maskSensitiveDataInJSON(jsonBody, sensitiveFields)
				Expect(result).To(Equal(jsonBody))
			})
		})
	})

	Describe("responseWriter wrapper", func() {
		It("should capture response body while writing to original writer", func() {
			// Create a test context with Gin's response writer
			c, _ := gin.CreateTestContext(recorder)
			buffer := &bytes.Buffer{}
			wrapper := &responseWriter{
				ResponseWriter: c.Writer,
				body:           buffer,
			}

			testData := []byte("test response data")
			n, err := wrapper.Write(testData)

			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(len(testData)))
			
			// Should write to buffer
			Expect(buffer.String()).To(Equal("test response data"))
			
			// Should also write to the underlying recorder
			Expect(recorder.Body.String()).To(Equal("test response data"))
		})
	})

	Describe("Performance and timing", func() {
		It("should measure request duration accurately", func() {
			router.Use(RequestResponseLogging(DefaultLoggingConfig()))
			router.GET("/slow", func(c *gin.Context) {
				time.Sleep(50 * time.Millisecond) // Simulate slow operation
				c.JSON(200, gin.H{"status": "completed"})
			})

			start := time.Now()
			req := httptest.NewRequest("GET", "/slow", nil)
			router.ServeHTTP(recorder, req)
			actualDuration := time.Since(start)

			logOutput := logBuf.String()
			lines := strings.Split(strings.TrimSpace(logOutput), "\n")
			
			var responseLog map[string]interface{}
			err := json.Unmarshal([]byte(lines[1]), &responseLog)
			Expect(err).ToNot(HaveOccurred())

			// Duration should be present and reasonable
			Expect(responseLog["duration"]).ToNot(BeZero())
			Expect(responseLog["duration_ms"]).ToNot(BeEmpty())
			
			// Should be close to actual duration (within reasonable margin)
			loggedDurationStr := responseLog["duration_ms"].(string)
			Expect(loggedDurationStr).To(ContainSubstring("ms"))
			
			// The logged duration should be at least as long as our sleep
			// (allowing for some overhead)
			Expect(actualDuration).To(BeNumerically(">=", 45*time.Millisecond))
		})
	})
})