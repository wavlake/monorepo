package testutil

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
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
func MockController(t interface {
	Helper()
	Errorf(format string, args ...interface{})
	FailNow()
}) *gomock.Controller {
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