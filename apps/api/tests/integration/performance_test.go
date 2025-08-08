package integration

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/wavlake/monorepo/internal/config"
	"github.com/wavlake/monorepo/internal/handlers"
	"github.com/wavlake/monorepo/internal/services"
	"github.com/wavlake/monorepo/internal/utils"
)

// PerformanceTestSuite tests basic system performance
type PerformanceTestSuite struct {
	suite.Suite
	router         *gin.Engine
	server         *httptest.Server
	userService    services.UserServiceInterface
	storageService services.StorageServiceInterface
	ctx            context.Context
}

// PerformanceMetrics tracks basic performance measurements
type PerformanceMetrics struct {
	TotalRequests       int
	SuccessfulRequests  int
	FailedRequests      int
	AverageResponseTime time.Duration
	MinResponseTime     time.Duration
	MaxResponseTime     time.Duration
	TotalDuration       time.Duration
	RequestsPerSecond   float64
}

// SetupSuite runs once before all tests
func (suite *PerformanceTestSuite) SetupSuite() {
	suite.ctx = context.Background()
	
	// Set test environment
	os.Setenv("DEVELOPMENT", "true")
	os.Setenv("SKIP_AUTH", "true") // Skip auth for performance testing
	os.Setenv("GOOGLE_CLOUD_PROJECT", "test-project")
	
	// Load dev config
	devConfig := config.LoadDevConfig()
	
	// Initialize router with minimal middleware for performance testing
	gin.SetMode(gin.TestMode)
	suite.router = gin.New()
	
	// Add only essential middleware
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
	
	// Set up basic test routes
	suite.setupBasicRoutes(authHandlers, nostrTrackService, processingService, audioProcessor, &devConfig)
	
	// Create test server
	suite.server = httptest.NewServer(suite.router)
}

func (suite *PerformanceTestSuite) setupBasicRoutes(
	authHandlers *handlers.AuthHandlers,
	nostrTrackService *services.NostrTrackService,
	processingService *services.ProcessingService,
	audioProcessor *utils.AudioProcessor,
	devConfig *config.DevConfig,
) {
	// Basic heartbeat endpoint
	suite.router.GET("/heartbeat", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "timestamp": time.Now().Unix()})
	})
	
	// Development status endpoint
	suite.router.GET("/dev/status", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"mode":               "development",
			"mock_storage":       devConfig.MockStorage,
			"firestore_emulator": config.IsFirestoreEmulated(),
		})
	})
	
	// API v1 endpoints for compatibility testing
	v1 := suite.router.Group("/v1")
	{
		v1.GET("/heartbeat", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok", "version": "v1"})
		})
		
		// Auth endpoints (lightweight stubs)
		authGroup := v1.Group("/auth")
		{
			authGroup.GET("/get-linked-pubkeys", func(c *gin.Context) {
				c.JSON(503, gin.H{"error": "Firebase authentication not available in development mode"})
			})
		}
		
		// Tracks endpoints (lightweight stubs)
		tracksGroup := v1.Group("/tracks")
		{
			tracksGroup.GET("/:id", func(c *gin.Context) {
				trackID := c.Param("id")
				c.JSON(http.StatusNotFound, gin.H{"error": "track not found", "track_id": trackID})
			})
		}
	}
}

// TearDownSuite runs once after all tests
func (suite *PerformanceTestSuite) TearDownSuite() {
	if suite.server != nil {
		suite.server.Close()
	}
	if suite.storageService != nil {
		if closer, ok := suite.storageService.(*services.StorageService); ok {
			closer.Close()
		}
	}
}

// Test basic response time for heartbeat endpoint
func (suite *PerformanceTestSuite) TestHeartbeatResponseTime() {
	const numRequests = 50
	var totalDuration time.Duration
	var minDuration = time.Hour
	var maxDuration time.Duration
	successCount := 0
	
	for i := 0; i < numRequests; i++ {
		start := time.Now()
		resp, err := http.Get(suite.server.URL + "/heartbeat")
		duration := time.Since(start)
		
		if err == nil && resp.StatusCode == http.StatusOK {
			successCount++
			resp.Body.Close()
			
			totalDuration += duration
			if duration < minDuration {
				minDuration = duration
			}
			if duration > maxDuration {
				maxDuration = duration
			}
		} else if resp != nil {
			resp.Body.Close()
		}
	}
	
	avgDuration := totalDuration / time.Duration(successCount)
	
	suite.T().Logf("Heartbeat performance:")
	suite.T().Logf("  Requests: %d", numRequests)
	suite.T().Logf("  Successful: %d", successCount)
	suite.T().Logf("  Average: %v", avgDuration)
	suite.T().Logf("  Min: %v", minDuration)
	suite.T().Logf("  Max: %v", maxDuration)
	
	// Basic performance targets
	assert.Greater(suite.T(), successCount, numRequests*95/100, "Should have >95% success rate")
	assert.Less(suite.T(), avgDuration, 20*time.Millisecond, "Average response time should be under 20ms")
	assert.Less(suite.T(), maxDuration, 100*time.Millisecond, "Max response time should be under 100ms")
}

// Test basic concurrent load handling (small scale)
func (suite *PerformanceTestSuite) TestBasicConcurrentLoad() {
	const numWorkers = 10
	const requestsPerWorker = 5
	const totalRequests = numWorkers * requestsPerWorker
	
	var wg sync.WaitGroup
	results := make(chan time.Duration, totalRequests)
	errors := make(chan error, totalRequests)
	
	startTime := time.Now()
	
	// Launch workers
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			
			for j := 0; j < requestsPerWorker; j++ {
				requestStart := time.Now()
				resp, err := http.Get(suite.server.URL + "/heartbeat")
				requestDuration := time.Since(requestStart)
				
				if err != nil {
					errors <- err
					continue
				}
				
				if resp.StatusCode != http.StatusOK {
					errors <- assert.AnError
					resp.Body.Close()
					continue
				}
				
				resp.Body.Close()
				results <- requestDuration
			}
		}(i)
	}
	
	wg.Wait()
	totalDuration := time.Since(startTime)
	close(results)
	close(errors)
	
	// Collect results
	var responseTimes []time.Duration
	for duration := range results {
		responseTimes = append(responseTimes, duration)
	}
	
	var errorCount int
	for range errors {
		errorCount++
	}
	
	// Calculate basic metrics
	metrics := suite.calculateMetrics(responseTimes, totalDuration, errorCount)
	
	suite.T().Logf("Basic concurrent load test results:")
	suite.T().Logf("  Workers: %d", numWorkers)
	suite.T().Logf("  Total requests: %d", totalRequests)
	suite.T().Logf("  Successful requests: %d", metrics.SuccessfulRequests)
	suite.T().Logf("  Failed requests: %d", metrics.FailedRequests)
	suite.T().Logf("  Total duration: %v", metrics.TotalDuration)
	suite.T().Logf("  Average response time: %v", metrics.AverageResponseTime)
	suite.T().Logf("  Requests per second: %.2f", metrics.RequestsPerSecond)
	
	// Basic performance targets
	assert.Greater(suite.T(), metrics.RequestsPerSecond, 100.0, "Should handle at least 100 requests per second")
	assert.Less(suite.T(), float64(metrics.FailedRequests)/float64(totalRequests), 0.05, "Error rate should be under 5%")
	assert.Less(suite.T(), metrics.AverageResponseTime, 50*time.Millisecond, "Average response time should be under 50ms")
}

// Test API endpoint response consistency
func (suite *PerformanceTestSuite) TestAPIEndpointPerformance() {
	endpoints := []struct {
		name string
		path string
		expectedStatus int
	}{
		{"Root heartbeat", "/heartbeat", 200},
		{"V1 heartbeat", "/v1/heartbeat", 200},
		{"Dev status", "/dev/status", 200},
		{"Auth endpoint", "/v1/auth/get-linked-pubkeys", 503},
		{"Track endpoint", "/v1/tracks/test-id", 404},
	}
	
	for _, endpoint := range endpoints {
		suite.T().Run(endpoint.name, func(t *testing.T) {
			const numRequests = 10
			var responseTimes []time.Duration
			successCount := 0
			
			for i := 0; i < numRequests; i++ {
				start := time.Now()
				resp, err := http.Get(suite.server.URL + endpoint.path)
				duration := time.Since(start)
				
				if err == nil && resp.StatusCode == endpoint.expectedStatus {
					successCount++
					responseTimes = append(responseTimes, duration)
				}
				
				if resp != nil {
					resp.Body.Close()
				}
			}
			
			if len(responseTimes) > 0 {
				var totalTime time.Duration
				for _, rt := range responseTimes {
					totalTime += rt
				}
				avgTime := totalTime / time.Duration(len(responseTimes))
				
				t.Logf("  Endpoint: %s", endpoint.path)
				t.Logf("  Successful: %d/%d", successCount, numRequests)
				t.Logf("  Average response time: %v", avgTime)
				
				assert.Greater(t, successCount, numRequests*80/100, "Should have >80% success rate")
				assert.Less(t, avgTime, 30*time.Millisecond, "Average response time should be under 30ms")
			}
		})
	}
}

// Test response format consistency under basic load
func (suite *PerformanceTestSuite) TestResponseFormatConsistency() {
	const numRequests = 20
	var responseTimes []time.Duration
	
	for i := 0; i < numRequests; i++ {
		start := time.Now()
		resp, err := http.Get(suite.server.URL + "/heartbeat")
		duration := time.Since(start)
		
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)
		
		var response map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&response)
		assert.NoError(suite.T(), err)
		resp.Body.Close()
		
		// Verify response structure
		assert.Contains(suite.T(), response, "status")
		assert.Contains(suite.T(), response, "timestamp")
		assert.Equal(suite.T(), "ok", response["status"])
		
		responseTimes = append(responseTimes, duration)
	}
	
	// Calculate metrics
	var totalTime time.Duration
	for _, rt := range responseTimes {
		totalTime += rt
	}
	avgTime := totalTime / time.Duration(len(responseTimes))
	
	suite.T().Logf("Response format consistency test:")
	suite.T().Logf("  Requests: %d", numRequests)
	suite.T().Logf("  Average response time: %v", avgTime)
	
	// All responses should be consistent and fast
	assert.Less(suite.T(), avgTime, 25*time.Millisecond, "Consistent responses should be fast")
}

// Helper function to calculate basic performance metrics
func (suite *PerformanceTestSuite) calculateMetrics(responseTimes []time.Duration, totalDuration time.Duration, errorCount int) PerformanceMetrics {
	if len(responseTimes) == 0 {
		return PerformanceMetrics{
			FailedRequests: errorCount,
			TotalDuration: totalDuration,
			RequestsPerSecond: 0,
		}
	}
	
	var totalResponseTime time.Duration
	minTime := responseTimes[0]
	maxTime := responseTimes[0]
	
	for _, responseTime := range responseTimes {
		totalResponseTime += responseTime
		if responseTime < minTime {
			minTime = responseTime
		}
		if responseTime > maxTime {
			maxTime = responseTime
		}
	}
	
	avgResponseTime := totalResponseTime / time.Duration(len(responseTimes))
	requestsPerSecond := float64(len(responseTimes)) / totalDuration.Seconds()
	
	return PerformanceMetrics{
		TotalRequests:      len(responseTimes) + errorCount,
		SuccessfulRequests: len(responseTimes),
		FailedRequests:     errorCount,
		AverageResponseTime: avgResponseTime,
		MinResponseTime:    minTime,
		MaxResponseTime:    maxTime,
		TotalDuration:      totalDuration,
		RequestsPerSecond:  requestsPerSecond,
	}
}

// Run the performance test suite
func TestPerformanceSuite(t *testing.T) {
	suite.Run(t, new(PerformanceTestSuite))
}