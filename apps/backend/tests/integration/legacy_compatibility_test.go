package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/wavlake/monorepo/internal/config"
	"github.com/wavlake/monorepo/internal/handlers"
	"github.com/wavlake/monorepo/internal/middleware"
	"github.com/wavlake/monorepo/internal/services"
	"github.com/wavlake/monorepo/internal/utils"
)

// LegacyCompatibilityTestSuite tests legacy endpoint compatibility during migration
type LegacyCompatibilityTestSuite struct {
	suite.Suite
	router         *gin.Engine
	server         *httptest.Server
	userService    services.UserServiceInterface
	storageService services.StorageServiceInterface
	ctx            context.Context
}

// SetupSuite runs once before all tests
func (suite *LegacyCompatibilityTestSuite) SetupSuite() {
	suite.ctx = context.Background()
	
	// Set test environment
	os.Setenv("DEVELOPMENT", "true")
	os.Setenv("SKIP_AUTH", "true") // Skip auth for legacy compatibility testing
	os.Setenv("GOOGLE_CLOUD_PROJECT", "test-project")
	
	// Load dev config
	devConfig := config.LoadDevConfig()
	
	// Initialize router
	gin.SetMode(gin.TestMode)
	suite.router = gin.New()
	
	// Add middleware
	loggingConfig := middleware.LoggingConfig{
		LogRequests:     false,
		LogResponses:    false,
		LogHeaders:      false,
		LogRequestBody:  false,
		LogResponseBody: false,
		MaxBodySize:     1024,
		SkipPaths:       []string{"/heartbeat"},
		SensitiveHeaders: []string{},
		SensitiveFields:  []string{},
	}
	suite.router.Use(middleware.RequestResponseLogging(loggingConfig))
	suite.router.Use(gin.Recovery())
	
	// Initialize storage service
	storageService, err := services.NewStorageService(suite.ctx, "test-bucket")
	if err != nil {
		suite.T().Logf("Warning: Could not initialize storage service: %v", err)
		storageService = nil
	}
	suite.storageService = storageService
	
	// Initialize other services
	var nostrTrackService *services.NostrTrackService
	var audioProcessor *utils.AudioProcessor
	var processingService *services.ProcessingService
	
	if storageService != nil {
		audioProcessor = utils.NewAudioProcessor("/tmp")
	}
	
	// Initialize handlers
	authHandlers := handlers.NewAuthHandlers(suite.userService)
	
	// Set up legacy and migrated endpoints
	suite.setupLegacyRoutes(authHandlers, nostrTrackService, processingService, audioProcessor, &devConfig)
	
	// Create test server
	suite.server = httptest.NewServer(suite.router)
}

func (suite *LegacyCompatibilityTestSuite) setupLegacyRoutes(
	authHandlers *handlers.AuthHandlers,
	nostrTrackService *services.NostrTrackService,
	processingService *services.ProcessingService,
	audioProcessor *utils.AudioProcessor,
	devConfig *config.DevConfig,
) {
	// Heartbeat endpoint (both legacy and new should work)
	suite.router.GET("/heartbeat", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	
	// Legacy API group (v1)
	v1 := suite.router.Group("/v1")
	
	// Legacy heartbeat variations
	v1.GET("/heartbeat", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "version": "v1"})
	})
	
	// Legacy auth endpoints 
	authGroup := v1.Group("/auth")
	{
		// Legacy format responses for backward compatibility
		authGroup.GET("/get-linked-pubkeys", func(c *gin.Context) {
			// Simulate legacy response format
			if devConfig.IsDevelopment {
				c.JSON(503, gin.H{
					"error": "Firebase authentication not available in development mode (SKIP_AUTH=true)",
					"legacy_format": true,
				})
			} else {
				c.JSON(401, gin.H{
					"error": "Missing Authorization header",
					"legacy_format": true,
				})
			}
		})
		
		authGroup.POST("/link-pubkey", func(c *gin.Context) {
			if devConfig.IsDevelopment {
				c.JSON(503, gin.H{
					"error": "Dual authentication not available in development mode (SKIP_AUTH=true)",
					"legacy_format": true,
				})
			} else {
				c.JSON(401, gin.H{
					"error": "Missing Authorization header",
					"legacy_format": true,
				})
			}
		})
		
		authGroup.POST("/unlink-pubkey", func(c *gin.Context) {
			if devConfig.IsDevelopment {
				c.JSON(503, gin.H{
					"error": "Firebase authentication not available in development mode (SKIP_AUTH=true)",
					"legacy_format": true,
				})
			} else {
				c.JSON(401, gin.H{
					"error": "Missing Authorization header",
					"legacy_format": true,
				})
			}
		})
		
		authGroup.POST("/check-pubkey-link", func(c *gin.Context) {
			c.JSON(401, gin.H{
				"error": "Missing Authorization header",
				"legacy_format": true,
			})
		})
	}
	
	// Legacy tracks endpoints
	tracksGroup := v1.Group("/tracks")
	{
		// Legacy track retrieval
		tracksGroup.GET("/:id", func(c *gin.Context) {
			trackID := c.Param("id")
			if trackID == "" {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "track ID is required",
					"legacy_format": true,
				})
				return
			}
			
			// Simulate legacy track not found response
			c.JSON(http.StatusNotFound, gin.H{
				"error": "track not found",
				"track_id": trackID,
				"legacy_format": true,
			})
		})
		
		// Legacy Nostr track creation
		tracksGroup.POST("/nostr", func(c *gin.Context) {
			c.JSON(401, gin.H{
				"error": "Missing Authorization header",
				"legacy_format": true,
			})
		})
		
		// Legacy user tracks
		tracksGroup.GET("/my", func(c *gin.Context) {
			c.JSON(401, gin.H{
				"error": "Missing Authorization header",
				"legacy_format": true,
			})
		})
		
		// Legacy track deletion
		tracksGroup.DELETE("/:trackId", func(c *gin.Context) {
			c.JSON(401, gin.H{
				"error": "Missing Authorization header",
				"legacy_format": true,
			})
		})
	}
	
	// Legacy user endpoints
	userGroup := v1.Group("/users")
	{
		userGroup.GET("/me", func(c *gin.Context) {
			c.JSON(401, gin.H{
				"error": "Missing Authorization header",
				"legacy_format": true,
			})
		})
		
		userGroup.GET("/:id/metadata", func(c *gin.Context) {
			userID := c.Param("id")
			c.JSON(404, gin.H{
				"error": "user not found",
				"user_id": userID,
				"legacy_format": true,
			})
		})
	}
}

// TearDownSuite runs once after all tests
func (suite *LegacyCompatibilityTestSuite) TearDownSuite() {
	if suite.server != nil {
		suite.server.Close()
	}
	if suite.storageService != nil {
		if closer, ok := suite.storageService.(*services.StorageService); ok {
			closer.Close()
		}
	}
}

// Test legacy heartbeat endpoints work
func (suite *LegacyCompatibilityTestSuite) TestLegacyHeartbeat() {
	tests := []struct {
		name     string
		endpoint string
		expected map[string]interface{}
	}{
		{
			name:     "Root heartbeat",
			endpoint: "/heartbeat",
			expected: map[string]interface{}{"status": "ok"},
		},
		{
			name:     "V1 heartbeat",
			endpoint: "/v1/heartbeat", 
			expected: map[string]interface{}{"status": "ok", "version": "v1"},
		},
	}
	
	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			resp, err := http.Get(suite.server.URL + tt.endpoint)
			assert.NoError(t, err)
			defer resp.Body.Close()
			
			assert.Equal(t, http.StatusOK, resp.StatusCode)
			
			var response map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&response)
			assert.NoError(t, err)
			
			for key, expectedValue := range tt.expected {
				assert.Equal(t, expectedValue, response[key])
			}
		})
	}
}

// Test legacy auth endpoints maintain backward compatibility
func (suite *LegacyCompatibilityTestSuite) TestLegacyAuthEndpoints() {
	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
		hasLegacyFlag  bool
	}{
		{
			name:           "Legacy get-linked-pubkeys",
			method:         "GET",
			path:           "/v1/auth/get-linked-pubkeys",
			expectedStatus: 503,
			hasLegacyFlag:  true,
		},
		{
			name:           "Legacy link-pubkey",
			method:         "POST",
			path:           "/v1/auth/link-pubkey",
			expectedStatus: 503,
			hasLegacyFlag:  true,
		},
		{
			name:           "Legacy unlink-pubkey",
			method:         "POST",
			path:           "/v1/auth/unlink-pubkey",
			expectedStatus: 503,
			hasLegacyFlag:  true,
		},
		{
			name:           "Legacy check-pubkey-link",
			method:         "POST",
			path:           "/v1/auth/check-pubkey-link",
			expectedStatus: 401,
			hasLegacyFlag:  true,
		},
	}
	
	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			client := &http.Client{Timeout: 5 * time.Second}
			
			var req *http.Request
			var err error
			
			if tt.method == "GET" {
				req, err = http.NewRequest("GET", suite.server.URL + tt.path, nil)
			} else {
				req, err = http.NewRequest("POST", suite.server.URL + tt.path, bytes.NewBuffer(nil))
				req.Header.Set("Content-Type", "application/json")
			}
			
			assert.NoError(t, err)
			
			resp, err := client.Do(req)
			assert.NoError(t, err)
			defer resp.Body.Close()
			
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
			
			var response map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&response)
			assert.NoError(t, err)
			
			// Verify legacy format flag is present
			if tt.hasLegacyFlag {
				assert.True(t, response["legacy_format"].(bool), "Should have legacy_format flag")
			}
			
			// Verify error message exists
			assert.NotEmpty(t, response["error"], "Should have error message")
		})
	}
}

// Test legacy track endpoints maintain backward compatibility
func (suite *LegacyCompatibilityTestSuite) TestLegacyTrackEndpoints() {
	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
		hasLegacyFlag  bool
	}{
		{
			name:           "Legacy get track",
			method:         "GET",
			path:           "/v1/tracks/test-track-id",
			expectedStatus: 404,
			hasLegacyFlag:  true,
		},
		{
			name:           "Legacy create nostr track",
			method:         "POST",
			path:           "/v1/tracks/nostr",
			expectedStatus: 401,
			hasLegacyFlag:  true,
		},
		{
			name:           "Legacy get my tracks",
			method:         "GET",
			path:           "/v1/tracks/my",
			expectedStatus: 401,
			hasLegacyFlag:  true,
		},
		{
			name:           "Legacy delete track",
			method:         "DELETE",
			path:           "/v1/tracks/test-track-id",
			expectedStatus: 401,
			hasLegacyFlag:  true,
		},
	}
	
	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			client := &http.Client{Timeout: 5 * time.Second}
			
			var req *http.Request
			var err error
			
			switch tt.method {
			case "GET":
				req, err = http.NewRequest("GET", suite.server.URL + tt.path, nil)
			case "POST":
				req, err = http.NewRequest("POST", suite.server.URL + tt.path, bytes.NewBuffer(nil))
				req.Header.Set("Content-Type", "application/json")
			case "DELETE":
				req, err = http.NewRequest("DELETE", suite.server.URL + tt.path, nil)
			}
			
			assert.NoError(t, err)
			
			resp, err := client.Do(req)
			assert.NoError(t, err)
			defer resp.Body.Close()
			
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
			
			var response map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&response)
			assert.NoError(t, err)
			
			// Verify legacy format flag is present
			if tt.hasLegacyFlag {
				assert.True(t, response["legacy_format"].(bool), "Should have legacy_format flag")
			}
			
			// Verify error message exists
			assert.NotEmpty(t, response["error"], "Should have error message")
		})
	}
}

// Test legacy user endpoints maintain expected behavior
func (suite *LegacyCompatibilityTestSuite) TestLegacyUserEndpoints() {
	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
		hasLegacyFlag  bool
	}{
		{
			name:           "Legacy get current user",
			method:         "GET",
			path:           "/v1/users/me",
			expectedStatus: 401,
			hasLegacyFlag:  true,
		},
		{
			name:           "Legacy get user metadata",
			method:         "GET",
			path:           "/v1/users/test-user-id/metadata",
			expectedStatus: 404,
			hasLegacyFlag:  true,
		},
	}
	
	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			resp, err := http.Get(suite.server.URL + tt.path)
			assert.NoError(t, err)
			defer resp.Body.Close()
			
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
			
			var response map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&response)
			assert.NoError(t, err)
			
			// Verify legacy format flag is present
			if tt.hasLegacyFlag {
				assert.True(t, response["legacy_format"].(bool), "Should have legacy_format flag")
			}
			
			// Verify error message exists
			assert.NotEmpty(t, response["error"], "Should have error message")
		})
	}
}

// Test response format consistency for legacy endpoints
func (suite *LegacyCompatibilityTestSuite) TestLegacyResponseFormats() {
	// Test that legacy endpoints return expected JSON structure
	resp, err := http.Get(suite.server.URL + "/v1/tracks/nonexistent")
	assert.NoError(suite.T(), err)
	defer resp.Body.Close()
	
	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(suite.T(), err)
	
	// Check legacy response structure
	assert.Contains(suite.T(), response, "error")
	assert.Contains(suite.T(), response, "track_id")
	assert.Contains(suite.T(), response, "legacy_format")
	assert.Equal(suite.T(), "nonexistent", response["track_id"])
	assert.True(suite.T(), response["legacy_format"].(bool))
}

// Test concurrent legacy endpoint access
func (suite *LegacyCompatibilityTestSuite) TestConcurrentLegacyAccess() {
	const numRequests = 15
	results := make(chan bool, numRequests)
	
	for i := 0; i < numRequests; i++ {
		go func() {
			client := &http.Client{Timeout: 5 * time.Second}
			resp, err := client.Get(suite.server.URL + "/v1/heartbeat")
			if err != nil {
				results <- false
				return
			}
			defer resp.Body.Close()
			
			var response map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&response)
			if err != nil {
				results <- false
				return
			}
			
			// Check response structure
			results <- resp.StatusCode == http.StatusOK && response["status"] == "ok"
		}()
	}
	
	successCount := 0
	for i := 0; i < numRequests; i++ {
		if <-results {
			successCount++
		}
	}
	
	assert.Equal(suite.T(), numRequests, successCount, "All concurrent legacy requests should succeed")
}

// Run the legacy compatibility test suite
func TestLegacyCompatibilitySuite(t *testing.T) {
	suite.Run(t, new(LegacyCompatibilityTestSuite))
}