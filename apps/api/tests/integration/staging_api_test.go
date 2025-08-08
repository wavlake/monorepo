package integration

import (
	"encoding/json"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// StagingAPITestSuite tests the main API functionality against staging
type StagingAPITestSuite struct {
	suite.Suite
	baseURL string
	client  *http.Client
}

// SetupSuite runs once before all tests
func (suite *StagingAPITestSuite) SetupSuite() {
	// Use staging URL by default, allow override
	suite.baseURL = os.Getenv("STAGING_URL")
	if suite.baseURL == "" {
		suite.baseURL = "https://api-staging-cgi4gylh7q-uc.a.run.app"
	}
	
	suite.client = &http.Client{
		Timeout: 10 * time.Second,
	}
}

// TestAPIHeartbeat tests the main API heartbeat
func (suite *StagingAPITestSuite) TestAPIHeartbeat() {
	resp, err := suite.client.Get(suite.baseURL + "/heartbeat")
	assert.NoError(suite.T(), err)
	
	if resp != nil {
		defer resp.Body.Close()
		assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)
		
		var response map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&response)
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), "ok", response["status"])
	}
}

// TestAPIResponseTimes measures API response performance  
func (suite *StagingAPITestSuite) TestAPIResponseTimes() {
	const numRequests = 3
	var totalDuration time.Duration
	successCount := 0
	
	for i := 0; i < numRequests; i++ {
		start := time.Now()
		resp, err := suite.client.Get(suite.baseURL + "/heartbeat")
		duration := time.Since(start)
		
		if err == nil && resp != nil && resp.StatusCode == http.StatusOK {
			successCount++
			totalDuration += duration
			resp.Body.Close()
		}
	}
	
	assert.Greater(suite.T(), successCount, 0, "At least one request should succeed")
	if successCount > 0 {
		avgDuration := totalDuration / time.Duration(successCount)
		suite.T().Logf("Average response time: %v", avgDuration)
		assert.Less(suite.T(), avgDuration, 2*time.Second, "Response time should be reasonable")
	}
}

// TestAPIErrorHandling tests API error responses
func (suite *StagingAPITestSuite) TestAPIErrorHandling() {
	// Test non-existent endpoint
	resp, err := suite.client.Get(suite.baseURL + "/nonexistent")
	assert.NoError(suite.T(), err)
	
	if resp != nil {
		defer resp.Body.Close()
		assert.Equal(suite.T(), http.StatusNotFound, resp.StatusCode)
	}
}

// Run the staging API test suite
func TestStagingAPISuite(t *testing.T) {
	// Skip if no staging URL configured and not in CI
	if os.Getenv("STAGING_URL") == "" && os.Getenv("CI") == "" {
		t.Skip("Skipping staging API tests - set STAGING_URL environment variable to run")
		return
	}
	
	suite.Run(t, new(StagingAPITestSuite))
}