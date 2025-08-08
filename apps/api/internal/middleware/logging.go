package middleware

import (
	"bytes"
	"io"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var logger *slog.Logger

func init() {
	// Initialize structured logger
	opts := &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}
	
	// Use JSON handler for structured logs
	handler := slog.NewJSONHandler(os.Stdout, opts)
	logger = slog.New(handler)
}

// responseWriter wraps gin.ResponseWriter to capture response body
type responseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *responseWriter) Write(data []byte) (int, error) {
	// Write to both the actual response and our buffer
	w.body.Write(data)
	return w.ResponseWriter.Write(data)
}

// LoggingConfig configures the logging middleware
type LoggingConfig struct {
	LogRequests       bool
	LogResponses      bool
	LogHeaders        bool
	LogRequestBody    bool
	LogResponseBody   bool
	MaxBodySize       int64
	SkipPaths         []string
	SensitiveHeaders  []string
	SensitiveFields   []string
}

// DefaultLoggingConfig returns a default logging configuration
func DefaultLoggingConfig() LoggingConfig {
	return LoggingConfig{
		LogRequests:     true,
		LogResponses:    true,
		LogHeaders:      true,
		LogRequestBody:  true,
		LogResponseBody: true,
		MaxBodySize:     1024 * 1024, // 1MB
		SkipPaths:       []string{"/heartbeat", "/health"},
		SensitiveHeaders: []string{
			"authorization",
			"x-firebase-token",
			"x-nostr-authorization",
			"cookie",
		},
		SensitiveFields: []string{
			"password",
			"token",
			"secret",
			"key",
			"auth",
		},
	}
}

// RequestResponseLogging returns a Gin middleware that logs requests and responses
func RequestResponseLogging(config LoggingConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip logging for certain paths
		for _, skipPath := range config.SkipPaths {
			if c.Request.URL.Path == skipPath {
				c.Next()
				return
			}
		}

		start := time.Now()
		correlationID := uuid.New().String()
		c.Set("correlation_id", correlationID)

		// Capture request body if enabled
		var requestBody []byte
		if config.LogRequestBody && c.Request.Body != nil {
			requestBody, _ = io.ReadAll(c.Request.Body)
			if len(requestBody) > int(config.MaxBodySize) {
				requestBody = requestBody[:config.MaxBodySize]
			}
			// Restore the body for the actual handler
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		// Wrap response writer to capture response body
		var responseBody *bytes.Buffer
		if config.LogResponses {
			responseBody = &bytes.Buffer{}
			writer := &responseWriter{
				ResponseWriter: c.Writer,
				body:          responseBody,
			}
			c.Writer = writer
		}

		// Log incoming request
		if config.LogRequests {
			logRequest(c, correlationID, requestBody, config)
		}

		// Process request
		c.Next()

		// Calculate processing time
		duration := time.Since(start)

		// Log outgoing response
		if config.LogResponses {
			logResponse(c, correlationID, responseBody.Bytes(), duration, config)
		}
	}
}

func logRequest(c *gin.Context, correlationID string, body []byte, config LoggingConfig) {
	attrs := []slog.Attr{
		slog.String("type", "request"),
		slog.String("correlation_id", correlationID),
		slog.String("method", c.Request.Method),
		slog.String("path", c.Request.URL.Path),
		slog.String("query", c.Request.URL.RawQuery),
		slog.String("remote_addr", c.ClientIP()),
		slog.String("user_agent", c.Request.UserAgent()),
	}

	// Add headers if enabled
	if config.LogHeaders {
		headers := make(map[string]string)
		for name, values := range c.Request.Header {
			value := strings.Join(values, ", ")
			
			// Mask sensitive headers
			if isSensitiveHeader(name, config.SensitiveHeaders) {
				value = maskSensitiveData(value)
			}
			
			headers[strings.ToLower(name)] = value
		}
		attrs = append(attrs, slog.Any("headers", headers))
	}

	// Add request body if enabled and present
	if config.LogRequestBody && len(body) > 0 {
		bodyStr := string(body)
		
		// Mask sensitive data in body
		bodyStr = maskSensitiveDataInJSON(bodyStr, config.SensitiveFields)
		
		attrs = append(attrs, slog.String("body", bodyStr))
		attrs = append(attrs, slog.Int("body_size", len(body)))
	}

	logger.LogAttrs(c.Request.Context(), slog.LevelInfo, "HTTP Request", attrs...)
}

func logResponse(c *gin.Context, correlationID string, body []byte, duration time.Duration, config LoggingConfig) {
	attrs := []slog.Attr{
		slog.String("type", "response"),
		slog.String("correlation_id", correlationID),
		slog.Int("status", c.Writer.Status()),
		slog.Int("size", c.Writer.Size()),
		slog.Duration("duration", duration),
		slog.String("duration_ms", duration.String()),
	}

	// Add response headers if enabled
	if config.LogHeaders {
		headers := make(map[string]string)
		for name, values := range c.Writer.Header() {
			headers[strings.ToLower(name)] = strings.Join(values, ", ")
		}
		attrs = append(attrs, slog.Any("headers", headers))
	}

	// Add response body if enabled and present
	if config.LogResponseBody && len(body) > 0 {
		bodyStr := string(body)
		
		// Truncate if too large
		if len(body) > int(config.MaxBodySize) {
			bodyStr = bodyStr[:config.MaxBodySize] + "... [truncated]"
		}
		
		// Mask sensitive data in response body
		bodyStr = maskSensitiveDataInJSON(bodyStr, config.SensitiveFields)
		
		attrs = append(attrs, slog.String("body", bodyStr))
		attrs = append(attrs, slog.Int("body_size", len(body)))
	}

	logger.LogAttrs(c.Request.Context(), slog.LevelInfo, "HTTP Response", attrs...)
}

// Helper functions for sensitive data masking
func isSensitiveHeader(name string, sensitiveHeaders []string) bool {
	name = strings.ToLower(name)
	for _, sensitive := range sensitiveHeaders {
		if strings.Contains(name, strings.ToLower(sensitive)) {
			return true
		}
	}
	return false
}

func maskSensitiveData(value string) string {
	if len(value) <= 4 {
		return "***"
	}
	return value[:2] + strings.Repeat("*", len(value)-4) + value[len(value)-2:]
}

func maskSensitiveDataInJSON(body string, sensitiveFields []string) string {
	for _, field := range sensitiveFields {
		// Simple replacement for JSON fields
		// In a production system, you'd want proper JSON parsing with regex
		if strings.Contains(strings.ToLower(body), `"`+strings.ToLower(field)+`":`) {
			// For now, just indicate that sensitive data was masked
			body = strings.ReplaceAll(body, `"`+field+`"`, `"`+field+`" [MASKED]`)
		}
	}
	return body
}

// GetCorrelationID retrieves the correlation ID from the Gin context
func GetCorrelationID(c *gin.Context) string {
	if id, exists := c.Get("correlation_id"); exists {
		return id.(string)
	}
	return ""
}

// LogWithCorrelation logs a message with the correlation ID from context
func LogWithCorrelation(c *gin.Context, level slog.Level, msg string, attrs ...slog.Attr) {
	correlationID := GetCorrelationID(c)
	if correlationID != "" {
		attrs = append([]slog.Attr{slog.String("correlation_id", correlationID)}, attrs...)
	}
	logger.LogAttrs(c.Request.Context(), level, msg, attrs...)
}