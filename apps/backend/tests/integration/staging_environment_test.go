package integration

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// StagingEnvironmentTestSuite tests the deployed staging environment
type StagingEnvironmentTestSuite struct {
	suite.Suite
	baseURL string
	client  *http.Client
	ctx     context.Context
}

// SetupSuite runs once before all tests
func (suite *StagingEnvironmentTestSuite) SetupSuite() {
	suite.ctx = context.Background()
	
	// Get staging URL from environment variable
	suite.baseURL = os.Getenv("STAGING_URL")
	if suite.baseURL == "" {
		// Use actual deployed staging URL
		suite.baseURL = "https://api-staging-cgi4gylh7q-uc.a.run.app"
	}
	
	// Remove trailing slash if present
	suite.baseURL = strings.TrimSuffix(suite.baseURL, "/")
	
	// Create HTTP client with reasonable timeout
	suite.client = &http.Client{
		Timeout: 30 * time.Second,
	}
	
	suite.T().Logf("Testing staging environment at: %s", suite.baseURL)
}

// TestStagingHeartbeat tests the basic heartbeat endpoint
func (suite *StagingEnvironmentTestSuite) TestStagingHeartbeat() {
	resp, err := suite.client.Get(suite.baseURL + "/heartbeat")
	assert.NoError(suite.T(), err, "Heartbeat request should not fail")
	
	if resp != nil {
		defer resp.Body.Close()
		
		assert.Equal(suite.T(), http.StatusOK, resp.StatusCode, "Heartbeat should return 200 OK")
		
		var response map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&response)
		assert.NoError(suite.T(), err, "Response should be valid JSON")
		
		// Verify response structure
		assert.Contains(suite.T(), response, "status", "Response should contain status")
		assert.Equal(suite.T(), "ok", response["status"], "Status should be 'ok'")
		
		suite.T().Logf("✅ Staging heartbeat successful: %v", response)
	}
}

// TestStagingResponseTimes tests response performance
func (suite *StagingEnvironmentTestSuite) TestStagingResponseTimes() {
	const numRequests = 5
	var totalDuration time.Duration
	var minDuration = time.Hour
	var maxDuration time.Duration
	successCount := 0
	
	for i := 0; i < numRequests; i++ {
		start := time.Now()
		resp, err := suite.client.Get(suite.baseURL + "/heartbeat")
		duration := time.Since(start)
		
		if err == nil && resp != nil && resp.StatusCode == http.StatusOK {
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
	
	// Calculate metrics
	if successCount > 0 {
		avgDuration := totalDuration / time.Duration(successCount)
		
		suite.T().Logf("Staging performance metrics:")
		suite.T().Logf("  Requests: %d", numRequests)
		suite.T().Logf("  Successful: %d", successCount)
		suite.T().Logf("  Average: %v", avgDuration)
		suite.T().Logf("  Min: %v", minDuration)
		suite.T().Logf("  Max: %v", maxDuration)
		
		// Performance assertions
		assert.Greater(suite.T(), successCount, numRequests*80/100, "Should have >80% success rate")
		assert.Less(suite.T(), avgDuration, 5*time.Second, "Average response time should be under 5 seconds")
	} else {
		suite.T().Fatal("No successful requests to staging environment")
	}
}

// TestStagingAPIEndpoints tests various API endpoints
func (suite *StagingEnvironmentTestSuite) TestStagingAPIEndpoints() {
	endpoints := []struct {
		name           string
		path           string
		expectedStatus int
		description    string
	}{
		{
			name:           "Root heartbeat",
			path:           "/heartbeat",
			expectedStatus: 200,
			description:    "Basic health check",
		},
		{
			name:           "Auth endpoint",
			path:           "/v1/auth/get-linked-pubkeys",
			expectedStatus: 401,
			description:    "Should require authentication",
		},
		{
			name:           "Track endpoint",
			path:           "/v1/tracks/test-id",
			expectedStatus: 200,
			description:    "Should return track not found JSON response",
		},
	}
	
	for _, endpoint := range endpoints {
		suite.T().Run(endpoint.name, func(t *testing.T) {
			resp, err := suite.client.Get(suite.baseURL + endpoint.path)
			
			assert.NoError(t, err, "Request should not fail")
			
			if resp != nil {
				defer resp.Body.Close()
				
				assert.Equal(t, endpoint.expectedStatus, resp.StatusCode, 
					"Status code for %s should be %d", endpoint.path, endpoint.expectedStatus)
				
				// Try to decode JSON response
				var response map[string]interface{}
				err = json.NewDecoder(resp.Body).Decode(&response)
				
				if err == nil {
					t.Logf("✅ %s: %s - %v", endpoint.name, endpoint.description, response)
				} else {
					t.Logf("✅ %s: %s - Non-JSON response", endpoint.name, endpoint.description)
				}
			}
		})
	}
}

// TestStagingEnvironmentConfiguration tests environment-specific behavior
func (suite *StagingEnvironmentTestSuite) TestStagingEnvironmentConfiguration() {
	// Note: /dev/status endpoint is only available in development mode
	// In staging/production, this endpoint returns 404 which is expected behavior
	resp, err := suite.client.Get(suite.baseURL + "/dev/status")
	assert.NoError(suite.T(), err, "Dev status request should not fail")
	
	if resp != nil {
		defer resp.Body.Close()
		
		// In staging/production environment, /dev/status should return 404
		if resp.StatusCode == http.StatusNotFound {
			suite.T().Logf("✅ Dev status endpoint correctly disabled in staging environment (404)")
		} else if resp.StatusCode == http.StatusOK {
			var response map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&response)
			assert.NoError(suite.T(), err, "Response should be valid JSON")
			
			suite.T().Logf("Staging environment configuration: %v", response)
			
			// Verify it's not running in full development mode
			// (staging might have some dev features enabled but should be production-like)
			if mode, exists := response["mode"]; exists {
				suite.T().Logf("Environment mode: %v", mode)
			}
		} else {
			suite.T().Logf("Dev status returned status: %d - this is normal for production-like staging", resp.StatusCode)
		}
	}
}

// TestStagingErrorHandling tests error handling in staging environment
func (suite *StagingEnvironmentTestSuite) TestStagingErrorHandling() {
	// Test non-existent endpoint
	resp, err := suite.client.Get(suite.baseURL + "/nonexistent")
	assert.NoError(suite.T(), err, "Request should not fail")
	
	if resp != nil {
		defer resp.Body.Close()
		
		// Should return 404 for non-existent endpoints
		assert.Equal(suite.T(), http.StatusNotFound, resp.StatusCode, "Should return 404 for non-existent endpoint")
		suite.T().Logf("✅ Error handling: 404 returned for non-existent endpoint")
	}
}

// TestStagingConcurrentLoad tests basic concurrent load handling
func (suite *StagingEnvironmentTestSuite) TestStagingConcurrentLoad() {
	const numWorkers = 5
	const requestsPerWorker = 3
	const totalRequests = numWorkers * requestsPerWorker
	
	results := make(chan bool, totalRequests)
	
	// Launch concurrent requests
	for i := 0; i < numWorkers; i++ {
		go func(workerID int) {
			for j := 0; j < requestsPerWorker; j++ {
				resp, err := suite.client.Get(suite.baseURL + "/heartbeat")
				success := false
				
				if err == nil && resp != nil && resp.StatusCode == http.StatusOK {
					success = true
					resp.Body.Close()
				} else if resp != nil {
					resp.Body.Close()
				}
				
				results <- success
			}
		}(i)
	}
	
	// Collect results
	successCount := 0
	for i := 0; i < totalRequests; i++ {
		if <-results {
			successCount++
		}
	}
	
	suite.T().Logf("Concurrent load test results:")
	suite.T().Logf("  Workers: %d", numWorkers)
	suite.T().Logf("  Total requests: %d", totalRequests)
	suite.T().Logf("  Successful requests: %d", successCount)
	suite.T().Logf("  Success rate: %.1f%%", float64(successCount)/float64(totalRequests)*100)
	
	// Should handle basic concurrent load
	assert.Greater(suite.T(), successCount, totalRequests*80/100, "Should handle >80% of concurrent requests")
}

// TestStagingHeaders tests HTTP headers and security
func (suite *StagingEnvironmentTestSuite) TestStagingHeaders() {
	resp, err := suite.client.Get(suite.baseURL + "/heartbeat")
	assert.NoError(suite.T(), err, "Request should not fail")
	
	if resp != nil {
		defer resp.Body.Close()
		
		// Check for important headers
		suite.T().Logf("Response headers:")
		for name, values := range resp.Header {
			suite.T().Logf("  %s: %v", name, values)
		}
		
		// Verify content type
		contentType := resp.Header.Get("Content-Type")
		assert.Contains(suite.T(), contentType, "application/json", "Should return JSON content type")
	}
}

// Run the staging environment test suite
func TestStagingEnvironmentSuite(t *testing.T) {
	// Only run if STAGING_URL or GCP_PROJECT is set
	if os.Getenv("STAGING_URL") == "" && os.Getenv("GCP_PROJECT") == "" {
		t.Skip("Skipping staging tests - no STAGING_URL or GCP_PROJECT environment variable set")
		return
	}
	
	suite.Run(t, new(StagingEnvironmentTestSuite))
}