package integration

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// AuthFlowTestSuite tests authentication flows and middleware
type AuthFlowTestSuite struct {
	suite.Suite
	router           *gin.Engine
	server           *httptest.Server
	ctx              context.Context
	testPubkey       string
	testPrivateKey   string
}

// SetupSuite runs once before all tests
func (suite *AuthFlowTestSuite) SetupSuite() {
	suite.ctx = context.Background()
	
	// Generate test keypair for NIP-98 testing
	suite.generateTestKeypair()
	
	// Set test environment
	os.Setenv("DEVELOPMENT", "true")
	os.Setenv("SKIP_AUTH", "false") // Enable auth for this test suite
	os.Setenv("GOOGLE_CLOUD_PROJECT", "test-project")
	
	// Initialize router
	gin.SetMode(gin.TestMode)
	suite.router = gin.New()
	
	// Add minimal middleware
	suite.router.Use(gin.Recovery())
	
	// Set up authentication test routes
	suite.setupAuthRoutes()
	
	// Create test server
	suite.server = httptest.NewServer(suite.router)
}

func (suite *AuthFlowTestSuite) generateTestKeypair() {
	// Generate a random 32-byte private key for testing
	privateKeyBytes := make([]byte, 32)
	_, err := rand.Read(privateKeyBytes)
	if err != nil {
		suite.T().Fatalf("Failed to generate test private key: %v", err)
	}
	
	suite.testPrivateKey = hex.EncodeToString(privateKeyBytes)
	
	// Generate corresponding public key (simplified for testing)
	// In real implementation, this would use secp256k1
	hash := sha256.Sum256(privateKeyBytes)
	suite.testPubkey = hex.EncodeToString(hash[:])
}

func (suite *AuthFlowTestSuite) setupAuthRoutes() {
	// Test routes for different authentication patterns
	
	// 1. No authentication required
	suite.router.GET("/public", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "public endpoint"})
	})
	
	// 2. Firebase authentication only (simulated)
	suite.router.GET("/firebase-only", suite.simulateFirebaseMiddleware(), func(c *gin.Context) {
		firebaseUID, exists := c.Get("firebase_uid")
		if !exists {
			c.JSON(401, gin.H{"error": "Firebase UID not found"})
			return
		}
		c.JSON(200, gin.H{
			"message":      "firebase authenticated",
			"firebase_uid": firebaseUID,
		})
	})
	
	// 3. NIP-98 authentication only (simulated)
	suite.router.POST("/nip98-only", suite.simulateNIP98Middleware(), func(c *gin.Context) {
		pubkey, exists := c.Get("pubkey")
		if !exists {
			c.JSON(401, gin.H{"error": "NIP-98 pubkey not found"})
			return
		}
		c.JSON(200, gin.H{
			"message": "nip98 authenticated",
			"pubkey":  pubkey,
		})
	})
	
	// 4. Dual authentication (both Firebase and NIP-98)
	suite.router.POST("/dual-auth", 
		suite.simulateFirebaseMiddleware(),
		suite.simulateNIP98Middleware(),
		func(c *gin.Context) {
			firebaseUID, fbExists := c.Get("firebase_uid")
			pubkey, nip98Exists := c.Get("pubkey")
			
			if !fbExists || !nip98Exists {
				c.JSON(401, gin.H{"error": "Both Firebase and NIP-98 authentication required"})
				return
			}
			
			c.JSON(200, gin.H{
				"message":      "dual authenticated",
				"firebase_uid": firebaseUID,
				"pubkey":       pubkey,
			})
		})
	
	// 5. Flexible authentication (either Firebase or NIP-98)
	suite.router.GET("/flexible-auth", suite.simulateFlexibleAuthMiddleware(), func(c *gin.Context) {
		firebaseUID, fbExists := c.Get("firebase_uid")
		pubkey, nip98Exists := c.Get("pubkey")
		
		authType := "none"
		authValue := ""
		
		if fbExists {
			authType = "firebase"
			authValue = firebaseUID.(string)
		} else if nip98Exists {
			authType = "nip98"
			authValue = pubkey.(string)
		}
		
		c.JSON(200, gin.H{
			"message":    "flexible auth success",
			"auth_type":  authType,
			"auth_value": authValue,
		})
	})
}

// Middleware simulators for testing authentication patterns
func (suite *AuthFlowTestSuite) simulateFirebaseMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(401, gin.H{"error": "Missing Authorization header"})
			c.Abort()
			return
		}
		
		// Simulate Firebase token validation
		if authHeader == "Bearer valid-firebase-token" {
			c.Set("firebase_uid", "test-firebase-uid")
			c.Next()
		} else {
			c.JSON(401, gin.H{"error": "Invalid Firebase token"})
			c.Abort()
			return
		}
	}
}

func (suite *AuthFlowTestSuite) simulateNIP98Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		nip98Header := c.GetHeader("X-Nostr-Authorization")
		if nip98Header == "" {
			c.JSON(401, gin.H{"error": "Missing X-Nostr-Authorization header"})
			c.Abort()
			return
		}
		
		// Simulate NIP-98 signature validation
		if nip98Header == "valid-nip98-signature" {
			c.Set("pubkey", suite.testPubkey)
			c.Next()
		} else {
			c.JSON(401, gin.H{"error": "Invalid NIP-98 signature"})
			c.Abort()
			return
		}
	}
}

func (suite *AuthFlowTestSuite) simulateFlexibleAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		nip98Header := c.GetHeader("X-Nostr-Authorization")
		
		// Try Firebase first
		if authHeader == "Bearer valid-firebase-token" {
			c.Set("firebase_uid", "test-firebase-uid")
			c.Next()
			return
		}
		
		// Try NIP-98 second
		if nip98Header == "valid-nip98-signature" {
			c.Set("pubkey", suite.testPubkey)
			c.Next()
			return
		}
		
		// No valid authentication
		c.JSON(401, gin.H{"error": "No valid authentication provided"})
		c.Abort()
	}
}

// TearDownSuite runs once after all tests
func (suite *AuthFlowTestSuite) TearDownSuite() {
	if suite.server != nil {
		suite.server.Close()
	}
}

// Test public endpoint (no auth required)
func (suite *AuthFlowTestSuite) TestPublicEndpoint() {
	resp, err := http.Get(suite.server.URL + "/public")
	assert.NoError(suite.T(), err)
	defer resp.Body.Close()
	
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)
	
	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "public endpoint", response["message"])
}

// Test Firebase authentication only
func (suite *AuthFlowTestSuite) TestFirebaseOnlyAuth() {
	// Test without token
	resp, err := http.Get(suite.server.URL + "/firebase-only")
	assert.NoError(suite.T(), err)
	defer resp.Body.Close()
	assert.Equal(suite.T(), http.StatusUnauthorized, resp.StatusCode)
	
	// Test with valid token
	client := &http.Client{}
	req, err := http.NewRequest("GET", suite.server.URL + "/firebase-only", nil)
	assert.NoError(suite.T(), err)
	req.Header.Set("Authorization", "Bearer valid-firebase-token")
	
	resp, err = client.Do(req)
	assert.NoError(suite.T(), err)
	defer resp.Body.Close()
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)
	
	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "firebase authenticated", response["message"])
	assert.Equal(suite.T(), "test-firebase-uid", response["firebase_uid"])
}

// Test NIP-98 authentication only
func (suite *AuthFlowTestSuite) TestNIP98OnlyAuth() {
	// Test without header
	resp, err := http.Post(suite.server.URL + "/nip98-only", "application/json", nil)
	assert.NoError(suite.T(), err)
	defer resp.Body.Close()
	assert.Equal(suite.T(), http.StatusUnauthorized, resp.StatusCode)
	
	// Test with valid signature
	client := &http.Client{}
	req, err := http.NewRequest("POST", suite.server.URL + "/nip98-only", bytes.NewReader(nil))
	assert.NoError(suite.T(), err)
	req.Header.Set("X-Nostr-Authorization", "valid-nip98-signature")
	req.Header.Set("Content-Type", "application/json")
	
	resp, err = client.Do(req)
	assert.NoError(suite.T(), err)
	defer resp.Body.Close()
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)
	
	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "nip98 authenticated", response["message"])
	assert.Equal(suite.T(), suite.testPubkey, response["pubkey"])
}

// Test dual authentication (both required)
func (suite *AuthFlowTestSuite) TestDualAuth() {
	client := &http.Client{}
	
	// Test with only Firebase token
	req, err := http.NewRequest("POST", suite.server.URL + "/dual-auth", bytes.NewReader(nil))
	assert.NoError(suite.T(), err)
	req.Header.Set("Authorization", "Bearer valid-firebase-token")
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := client.Do(req)
	assert.NoError(suite.T(), err)
	defer resp.Body.Close()
	assert.Equal(suite.T(), http.StatusUnauthorized, resp.StatusCode)
	
	// Test with only NIP-98 signature
	req, err = http.NewRequest("POST", suite.server.URL + "/dual-auth", bytes.NewReader(nil))
	assert.NoError(suite.T(), err)
	req.Header.Set("X-Nostr-Authorization", "valid-nip98-signature")
	req.Header.Set("Content-Type", "application/json")
	
	resp, err = client.Do(req)
	assert.NoError(suite.T(), err)
	defer resp.Body.Close()
	assert.Equal(suite.T(), http.StatusUnauthorized, resp.StatusCode)
	
	// Test with both tokens
	req, err = http.NewRequest("POST", suite.server.URL + "/dual-auth", bytes.NewReader(nil))
	assert.NoError(suite.T(), err)
	req.Header.Set("Authorization", "Bearer valid-firebase-token")
	req.Header.Set("X-Nostr-Authorization", "valid-nip98-signature")
	req.Header.Set("Content-Type", "application/json")
	
	resp, err = client.Do(req)
	assert.NoError(suite.T(), err)
	defer resp.Body.Close()
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)
	
	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "dual authenticated", response["message"])
	assert.Equal(suite.T(), "test-firebase-uid", response["firebase_uid"])
	assert.Equal(suite.T(), suite.testPubkey, response["pubkey"])
}

// Test flexible authentication (either works)
func (suite *AuthFlowTestSuite) TestFlexibleAuth() {
	client := &http.Client{}
	
	// Test with Firebase token
	req, err := http.NewRequest("GET", suite.server.URL + "/flexible-auth", nil)
	assert.NoError(suite.T(), err)
	req.Header.Set("Authorization", "Bearer valid-firebase-token")
	
	resp, err := client.Do(req)
	assert.NoError(suite.T(), err)
	defer resp.Body.Close()
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)
	
	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "firebase", response["auth_type"])
	assert.Equal(suite.T(), "test-firebase-uid", response["auth_value"])
	
	// Test with NIP-98 signature
	req, err = http.NewRequest("GET", suite.server.URL + "/flexible-auth", nil)
	assert.NoError(suite.T(), err)
	req.Header.Set("X-Nostr-Authorization", "valid-nip98-signature")
	
	resp, err = client.Do(req)
	assert.NoError(suite.T(), err)
	defer resp.Body.Close()
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)
	
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "nip98", response["auth_type"])
	assert.Equal(suite.T(), suite.testPubkey, response["auth_value"])
}

// Test authentication middleware error cases
func (suite *AuthFlowTestSuite) TestAuthErrorCases() {
	client := &http.Client{}
	
	tests := []struct {
		name           string
		endpoint       string
		method         string
		headers        map[string]string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "Firebase - Missing Authorization header",
			endpoint:       "/firebase-only",
			method:         "GET",
			headers:        map[string]string{},
			expectedStatus: 401,
			expectedError:  "Missing Authorization header",
		},
		{
			name:           "Firebase - Invalid token",
			endpoint:       "/firebase-only", 
			method:         "GET",
			headers:        map[string]string{"Authorization": "Bearer invalid-token"},
			expectedStatus: 401,
			expectedError:  "Invalid Firebase token",
		},
		{
			name:           "NIP-98 - Missing header",
			endpoint:       "/nip98-only",
			method:         "POST",
			headers:        map[string]string{},
			expectedStatus: 401,
			expectedError:  "Missing X-Nostr-Authorization header",
		},
		{
			name:           "NIP-98 - Invalid signature",
			endpoint:       "/nip98-only",
			method:         "POST",
			headers:        map[string]string{"X-Nostr-Authorization": "invalid-signature"},
			expectedStatus: 401,
			expectedError:  "Invalid NIP-98 signature",
		},
	}
	
	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			var req *http.Request
			var err error
			
			if tt.method == "GET" {
				req, err = http.NewRequest("GET", suite.server.URL + tt.endpoint, nil)
			} else {
				req, err = http.NewRequest("POST", suite.server.URL + tt.endpoint, bytes.NewReader(nil))
				req.Header.Set("Content-Type", "application/json")
			}
			
			assert.NoError(t, err)
			
			// Set test headers
			for key, value := range tt.headers {
				req.Header.Set(key, value)
			}
			
			resp, err := client.Do(req)
			assert.NoError(t, err)
			defer resp.Body.Close()
			
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
			
			var response map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&response)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedError, response["error"])
		})
	}
}

// Test concurrent authentication requests
func (suite *AuthFlowTestSuite) TestConcurrentAuthRequests() {
	const numRequests = 20
	results := make(chan bool, numRequests)
	
	for i := 0; i < numRequests; i++ {
		go func() {
			client := &http.Client{Timeout: 5 * time.Second}
			req, err := http.NewRequest("GET", suite.server.URL + "/firebase-only", nil)
			if err != nil {
				results <- false
				return
			}
			req.Header.Set("Authorization", "Bearer valid-firebase-token")
			
			resp, err := client.Do(req)
			if err != nil {
				results <- false
				return
			}
			defer resp.Body.Close()
			results <- resp.StatusCode == http.StatusOK
		}()
	}
	
	successCount := 0
	for i := 0; i < numRequests; i++ {
		if <-results {
			successCount++
		}
	}
	
	assert.Equal(suite.T(), numRequests, successCount, "All concurrent auth requests should succeed")
}

// Run the authentication flow test suite
func TestAuthFlowSuite(t *testing.T) {
	suite.Run(t, new(AuthFlowTestSuite))
}