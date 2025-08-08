package testutil

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// SetupGinTestContext creates a test Gin context with optional authentication values
func SetupGinTestContext(method, path string, body interface{}) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Create request body if provided
	var reqBody []byte
	if body != nil {
		reqBody, _ = json.Marshal(body)
	}

	req, _ := http.NewRequest(method, path, bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	return c, w
}

// SetAuthContext sets authentication values in Gin context for testing
func SetAuthContext(c *gin.Context, firebaseUID, nostrPubkey string) {
	if firebaseUID != "" {
		c.Set("firebase_uid", firebaseUID)
	}
	if nostrPubkey != "" {
		c.Set("nostr_pubkey", nostrPubkey)
		c.Set("pubkey", nostrPubkey) // Some handlers use "pubkey" key
	}
}

// MockController creates a new gomock controller for testing
func MockController(t gomock.TestReporter) *gomock.Controller {
	return gomock.NewController(t)
}

// AssertJSONResponse checks if response contains expected JSON structure
func AssertJSONResponse(w *httptest.ResponseRecorder, expectedStatus int) map[string]interface{} {
	if w.Code != expectedStatus {
		panic("Expected status code does not match")
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		panic("Failed to unmarshal response JSON")
	}

	return response
}

// ContextWithCancel returns a context that can be cancelled for testing timeouts
func ContextWithCancel() (context.Context, context.CancelFunc) {
	return context.WithCancel(context.Background())
}

// ParseJSONResponse parses JSON response from httptest.ResponseRecorder
func ParseJSONResponse(body *bytes.Buffer, target interface{}) error {
	return json.Unmarshal(body.Bytes(), target)
}

// Additional test utilities

// TestContextWithTimeout provides a context with timeout for testing
func TestContextWithTimeout(timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), timeout)
}

// AssertErrorResponse validates that the response contains an error
func AssertErrorResponse(t *testing.T, rec *httptest.ResponseRecorder, expectedStatus int, expectedError string) {
	assert.Equal(t, expectedStatus, rec.Code, "Response status should match")
	
	var response map[string]interface{}
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err, "Response should be valid JSON")
	
	assert.Contains(t, response, "error", "Response should contain error field")
	if expectedError != "" {
		assert.Contains(t, response["error"], expectedError, "Error message should contain expected text")
	}
}

// AssertSuccessResponse validates that the response indicates success
func AssertSuccessResponse(t *testing.T, rec *httptest.ResponseRecorder, expectedStatus int) {
	assert.Equal(t, expectedStatus, rec.Code, "Response status should match")
	
	// Verify response is valid JSON
	var response interface{}
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err, "Response should be valid JSON")
}

// FixedTime returns a fixed time for consistent testing
func FixedTime() time.Time {
	return time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
}

// GenerateTestID generates a test ID with prefix
func GenerateTestID(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
}

// StringPtr returns a pointer to a string
func StringPtr(s string) *string {
	return &s
}

// BoolPtr returns a pointer to a bool
func BoolPtr(b bool) *bool {
	return &b
}

// SkipIfShort skips the test if running in short mode
func SkipIfShort(t *testing.T, reason string) {
	if testing.Short() {
		t.Skipf("Skipping test in short mode: %s", reason)
	}
}

// TempFile creates a temporary file for testing
func TempFile(t *testing.T, content string) string {
	file, err := os.CreateTemp("", "test-*.tmp")
	require.NoError(t, err, "Should create temp file")
	
	defer file.Close()
	
	_, err = file.WriteString(content)
	require.NoError(t, err, "Should write to temp file")
	
	return file.Name()
}