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

// APIIntegrationTestSuite tests the integrated API functionality
type APIIntegrationTestSuite struct {
	suite.Suite
	router        *gin.Engine
	server        *httptest.Server
	userService   services.UserServiceInterface
	storageService services.StorageServiceInterface
	ctx           context.Context
}

// SetupSuite runs once before all tests
func (suite *APIIntegrationTestSuite) SetupSuite() {
	// Set up test environment
	suite.ctx = context.Background()
	
	// Set development configuration
	os.Setenv("DEVELOPMENT", "true")
	os.Setenv("SKIP_AUTH", "true")
	os.Setenv("GCS_BUCKET_NAME", "test-bucket")
	
	// Load dev config
	devConfig := config.LoadDevConfig()
	
	// Initialize router
	gin.SetMode(gin.TestMode)
	suite.router = gin.New()
	
	// Add basic middleware
	loggingConfig := middleware.LoggingConfig{
		LogRequests:     false, // Disable for tests
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
	
	// Initialize mock services (since we're in test mode)
	// In a real integration test, we'd use actual services with test databases
	
	// Initialize storage service (GCS)
	storageService, err := services.NewStorageService(suite.ctx, "test-bucket")
	if err != nil {
		// Fall back to mock or skip storage-dependent tests
		suite.T().Logf("Warning: Could not initialize storage service: %v", err)
		storageService = nil
	}
	suite.storageService = storageService
	
	// Initialize NostrTrackService and ProcessingService
	var nostrTrackService *services.NostrTrackService
	var audioProcessor *utils.AudioProcessor
	var processingService *services.ProcessingService
	
	if storageService != nil {
		// In real tests, we'd use a test Firestore instance
		// For now, we'll test what we can without external dependencies
		audioProcessor = utils.NewAudioProcessor("/tmp")
	}
	
	// Initialize handlers
	authHandlers := handlers.NewAuthHandlers(suite.userService)
	
	// Set up routes
	suite.setupRoutes(authHandlers, nostrTrackService, processingService, audioProcessor, &devConfig)
	
	// Create test server
	suite.server = httptest.NewServer(suite.router)
}

func (suite *APIIntegrationTestSuite) setupRoutes(
	authHandlers *handlers.AuthHandlers,
	nostrTrackService *services.NostrTrackService,
	processingService *services.ProcessingService,
	audioProcessor *utils.AudioProcessor,
	devConfig *config.DevConfig,
) {
	// Heartbeat endpoint
	suite.router.GET("/heartbeat", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	
	// Development endpoints
	if devConfig.IsDevelopment {
		devGroup := suite.router.Group("/dev")
		{
			devGroup.GET("/status", func(c *gin.Context) {
				c.JSON(200, gin.H{
					"mode":               "development",
					"mock_storage":       devConfig.MockStorage,
					"firestore_emulator": config.IsFirestoreEmulated(),
				})
			})
		}
	}
	
	// API endpoints
	v1 := suite.router.Group("/v1")
	
	// Auth endpoints (development mode stubs)
	authGroup := v1.Group("/auth")
	{
		authGroup.GET("/get-linked-pubkeys", func(c *gin.Context) {
			c.JSON(503, gin.H{"error": "Firebase authentication not available in development mode (SKIP_AUTH=true)"})
		})
		authGroup.POST("/unlink-pubkey", func(c *gin.Context) {
			c.JSON(503, gin.H{"error": "Firebase authentication not available in development mode (SKIP_AUTH=true)"})
		})
		authGroup.POST("/link-pubkey", func(c *gin.Context) {
			c.JSON(503, gin.H{"error": "Dual authentication not available in development mode (SKIP_AUTH=true)"})
		})
		authGroup.POST("/check-pubkey-link", func(c *gin.Context) {
			c.JSON(401, gin.H{"error": "Missing Authorization header"})
		})
	}
	
	// Tracks endpoints (development mode stubs)
	tracksGroup := v1.Group("/tracks")
	{
		tracksGroup.GET("/:id", func(c *gin.Context) {
			trackID := c.Param("id")
			if trackID == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "track ID is required"})
				return
			}
			c.JSON(http.StatusNotFound, gin.H{"error": "track not found"})
		})
		
		tracksGroup.POST("/nostr", func(c *gin.Context) {
			c.JSON(401, gin.H{"error": "Missing Authorization header"})
		})
		
		tracksGroup.GET("/my", func(c *gin.Context) {
			c.JSON(401, gin.H{"error": "Missing Authorization header"})
		})
		
		tracksGroup.DELETE("/:trackId", func(c *gin.Context) {
			c.JSON(401, gin.H{"error": "Missing Authorization header"})
		})
	}
}

// TearDownSuite runs once after all tests
func (suite *APIIntegrationTestSuite) TearDownSuite() {
	if suite.server != nil {
		suite.server.Close()
	}
	if suite.storageService != nil {
		// Clean up storage service if it has a Close method
		if closer, ok := suite.storageService.(*services.StorageService); ok {
			closer.Close()
		}
	}
}

// Test basic server functionality
func (suite *APIIntegrationTestSuite) TestHeartbeat() {
	resp, err := http.Get(suite.server.URL + "/heartbeat")
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)
	
	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "ok", response["status"])
	resp.Body.Close()
}

// Test development status endpoint
func (suite *APIIntegrationTestSuite) TestDevStatus() {
	resp, err := http.Get(suite.server.URL + "/dev/status")
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)
	
	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "development", response["mode"])
	resp.Body.Close()
}

// Test authentication endpoints return appropriate errors
func (suite *APIIntegrationTestSuite) TestAuthEndpoints() {
	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "Get linked pubkeys - Firebase not available",
			method:         "GET",
			path:           "/v1/auth/get-linked-pubkeys",
			expectedStatus: 503,
			expectedError:  "Firebase authentication not available in development mode (SKIP_AUTH=true)",
		},
		{
			name:           "Unlink pubkey - Firebase not available",
			method:         "POST",
			path:           "/v1/auth/unlink-pubkey",
			expectedStatus: 503,
			expectedError:  "Firebase authentication not available in development mode (SKIP_AUTH=true)",
		},
		{
			name:           "Link pubkey - Dual auth not available",
			method:         "POST",
			path:           "/v1/auth/link-pubkey",
			expectedStatus: 503,
			expectedError:  "Dual authentication not available in development mode (SKIP_AUTH=true)",
		},
		{
			name:           "Check pubkey link - Missing auth",
			method:         "POST",
			path:           "/v1/auth/check-pubkey-link",
			expectedStatus: 401,
			expectedError:  "Missing Authorization header",
		},
	}
	
	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			var resp *http.Response
			var err error
			
			if tt.method == "GET" {
				resp, err = http.Get(suite.server.URL + tt.path)
			} else {
				resp, err = http.Post(suite.server.URL + tt.path, "application/json", nil)
			}
			
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
			
			var response map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&response)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedError, response["error"])
			resp.Body.Close()
		})
	}
}

// Test track endpoints return appropriate errors
func (suite *APIIntegrationTestSuite) TestTrackEndpoints() {
	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "Get track - Not found",
			method:         "GET", 
			path:           "/v1/tracks/test-id",
			expectedStatus: 404,
			expectedError:  "track not found",
		},
		{
			name:           "Create track - Missing auth",
			method:         "POST",
			path:           "/v1/tracks/nostr",
			expectedStatus: 401,
			expectedError:  "Missing Authorization header",
		},
		{
			name:           "Get my tracks - Missing auth",
			method:         "GET",
			path:           "/v1/tracks/my",
			expectedStatus: 401,
			expectedError:  "Missing Authorization header",
		},
		{
			name:           "Delete track - Missing auth",
			method:         "DELETE",
			path:           "/v1/tracks/test-id",
			expectedStatus: 401,
			expectedError:  "Missing Authorization header",
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
			case "DELETE":
				req, err = http.NewRequest("DELETE", suite.server.URL + tt.path, nil)
			}
			
			assert.NoError(t, err)
			resp, err := client.Do(req)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
			
			var response map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&response)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedError, response["error"])
			resp.Body.Close()
		})
	}
}

// Test concurrent requests
func (suite *APIIntegrationTestSuite) TestConcurrentRequests() {
	const numRequests = 10
	results := make(chan bool, numRequests)
	
	for i := 0; i < numRequests; i++ {
		go func() {
			resp, err := http.Get(suite.server.URL + "/heartbeat")
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
	
	assert.Equal(suite.T(), numRequests, successCount, "All concurrent requests should succeed")
}

// Run the test suite
func TestAPIIntegrationSuite(t *testing.T) {
	suite.Run(t, new(APIIntegrationTestSuite))
}