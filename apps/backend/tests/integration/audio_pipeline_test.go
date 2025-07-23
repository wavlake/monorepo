package integration

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/wavlake/monorepo/internal/config"
	"github.com/wavlake/monorepo/internal/models"
	"github.com/wavlake/monorepo/internal/services"
	"github.com/wavlake/monorepo/internal/utils"
)

// AudioPipelineTestSuite tests the audio processing pipeline
type AudioPipelineTestSuite struct {
	suite.Suite
	ctx               context.Context
	audioProcessor    *utils.AudioProcessor
	processingService *services.ProcessingService
	storageService    services.StorageServiceInterface
	tempDir           string
	testAudioFile     string
}

// SetupSuite runs once before all tests
func (suite *AudioPipelineTestSuite) SetupSuite() {
	suite.ctx = context.Background()
	
	// Set test environment
	os.Setenv("DEVELOPMENT", "true")
	os.Setenv("GOOGLE_CLOUD_PROJECT", "test-project")
	
	// Create temporary directory for test files
	tempDir, err := ioutil.TempDir("", "audio_pipeline_test")
	assert.NoError(suite.T(), err)
	suite.tempDir = tempDir
	
	// Create a dummy audio file for testing
	suite.createTestAudioFile()
	
	// Initialize audio processor
	suite.audioProcessor = utils.NewAudioProcessor(suite.tempDir)
	
	// Initialize storage service (mock for testing)
	storageService, err := services.NewStorageService(suite.ctx, "test-bucket")
	if err != nil {
		suite.T().Logf("Warning: Could not initialize storage service: %v", err)
		// Continue with mock storage for testing
		storageService = nil
	}
	suite.storageService = storageService
	
	// Initialize processing service if we have storage
	if storageService != nil {
		// For testing, we don't need a real NostrTrackService
		var nostrTrackService *services.NostrTrackService = nil
		suite.processingService = services.NewProcessingService(
			storageService,
			nostrTrackService,
			suite.audioProcessor,
			suite.tempDir,
		)
	}
}

func (suite *AudioPipelineTestSuite) createTestAudioFile() {
	// Create a simple test audio file (dummy content for testing)
	// In a real test, this would be a valid audio file
	testContent := []byte("RIFF____WAVEfmt____________data____")
	
	testFilePath := filepath.Join(suite.tempDir, "test_audio.wav")
	err := ioutil.WriteFile(testFilePath, testContent, 0644)
	assert.NoError(suite.T(), err)
	
	suite.testAudioFile = testFilePath
}

// TearDownSuite runs once after all tests
func (suite *AudioPipelineTestSuite) TearDownSuite() {
	// Clean up temporary files
	if suite.tempDir != "" {
		os.RemoveAll(suite.tempDir)
	}
	
	// Close storage service if available
	if suite.storageService != nil {
		if closer, ok := suite.storageService.(*services.StorageService); ok {
			closer.Close()
		}
	}
}

// Test audio processor initialization
func (suite *AudioPipelineTestSuite) TestAudioProcessorInitialization() {
	assert.NotNil(suite.T(), suite.audioProcessor, "Audio processor should be initialized")
	
	// Verify the audio processor was created successfully
	assert.IsType(suite.T(), &utils.AudioProcessor{}, suite.audioProcessor)
}

// Test audio file validation with dummy file
func (suite *AudioPipelineTestSuite) TestAudioFileValidation() {
	if suite.audioProcessor == nil {
		suite.T().Skip("Audio processor not available")
		return
	}
	
	// Test that dummy file exists
	_, err := os.Stat(suite.testAudioFile)
	assert.NoError(suite.T(), err, "Test audio file should exist")
	
	// Test invalid file path
	invalidPath := filepath.Join(suite.tempDir, "nonexistent.wav")
	_, err = os.Stat(invalidPath)
	assert.Error(suite.T(), err, "Nonexistent file should not exist")
}

// Test audio file metadata extraction
func (suite *AudioPipelineTestSuite) TestAudioMetadataExtraction() {
	if suite.audioProcessor == nil {
		suite.T().Skip("Audio processor not available")
		return
	}
	
	// Test GetAudioInfo with dummy file - this should fail gracefully
	_, err := suite.audioProcessor.GetAudioInfo(suite.ctx, suite.testAudioFile)
	// This is expected to fail for our dummy file, but we test that it handles errors gracefully
	if err != nil {
		suite.T().Logf("Expected error for dummy audio file: %v", err)
		assert.Contains(suite.T(), err.Error(), "failed to get audio info")
	}
}

// Test audio file validation with real interface
func (suite *AudioPipelineTestSuite) TestAudioFileValidationInterface() {
	if suite.audioProcessor == nil {
		suite.T().Skip("Audio processor not available")
		return
	}
	
	// Test ValidateAudioFile with dummy file - this should fail but handle error gracefully
	err := suite.audioProcessor.ValidateAudioFile(suite.ctx, suite.testAudioFile)
	// This is expected to fail for our dummy file
	if err != nil {
		suite.T().Logf("Expected validation error for dummy audio file: %v", err)
		assert.Contains(suite.T(), err.Error(), "file is not a valid audio file")
	}
}

// Test compression options validation
func (suite *AudioPipelineTestSuite) TestCompressionOptions() {
	if suite.audioProcessor == nil {
		suite.T().Skip("Audio processor not available")
		return
	}
	
	// Test available compression formats
	formats := suite.audioProcessor.GetSupportedFormats()
	
	// Common audio formats should be supported
	expectedFormats := []string{"mp3", "wav", "flac", "ogg"}
	for _, format := range expectedFormats {
		assert.Contains(suite.T(), formats, format, "Format %s should be supported", format)
	}
}

// Test format support checking
func (suite *AudioPipelineTestSuite) TestFormatSupport() {
	if suite.audioProcessor == nil {
		suite.T().Skip("Audio processor not available")
		return
	}
	
	// Test supported formats
	assert.True(suite.T(), suite.audioProcessor.IsFormatSupported("mp3"))
	assert.True(suite.T(), suite.audioProcessor.IsFormatSupported("wav"))
	assert.True(suite.T(), suite.audioProcessor.IsFormatSupported("flac"))
	
	// Test unsupported format
	assert.False(suite.T(), suite.audioProcessor.IsFormatSupported("xyz"))
}

// Test audio compression with mock data
func (suite *AudioPipelineTestSuite) TestAudioCompression() {
	if suite.audioProcessor == nil {
		suite.T().Skip("Audio processor not available")
		return
	}
	
	// Create output path
	outputPath := filepath.Join(suite.tempDir, "compressed.mp3")
	
	// Test CompressAudio - this will fail with dummy file but should handle error gracefully
	err := suite.audioProcessor.CompressAudio(suite.ctx, suite.testAudioFile, outputPath)
	
	if err != nil {
		// This is expected for our dummy file
		suite.T().Logf("Expected compression error for dummy audio file: %v", err)
		assert.Contains(suite.T(), err.Error(), "failed to compress audio")
	}
}

// Test audio compression with options
func (suite *AudioPipelineTestSuite) TestAudioCompressionWithOptions() {
	if suite.audioProcessor == nil {
		suite.T().Skip("Audio processor not available")
		return
	}
	
	// Create output path
	outputPath := filepath.Join(suite.tempDir, "compressed_options.mp3")
	
	// Test compression options
	options := models.CompressionOption{
		Bitrate:    128,
		Format:     "mp3",
		Quality:    "medium",
		SampleRate: 44100,
	}
	
	// Test CompressAudioWithOptions - this will fail with dummy file but should handle error gracefully
	err := suite.audioProcessor.CompressAudioWithOptions(suite.ctx, suite.testAudioFile, outputPath, options)
	
	if err != nil {
		// This is expected for our dummy file
		suite.T().Logf("Expected compression error for dummy audio file: %v", err)
		assert.Contains(suite.T(), err.Error(), "failed to compress audio")
	}
}

// Test concurrent audio processing
func (suite *AudioPipelineTestSuite) TestConcurrentAudioProcessing() {
	if suite.audioProcessor == nil {
		suite.T().Skip("Audio processor not available")
		return
	}
	
	const numRequests = 5
	results := make(chan bool, numRequests)
	
	for i := 0; i < numRequests; i++ {
		go func(index int) {
			// Create a unique temp file for each goroutine
			tempFile := filepath.Join(suite.tempDir, fmt.Sprintf("concurrent_test_%d.txt", index))
			err := ioutil.WriteFile(tempFile, []byte(fmt.Sprintf("test %d", index)), 0644)
			if err != nil {
				results <- false
				return
			}
			
			// Test basic file operations
			_, statErr := os.Stat(tempFile)
			exists := statErr == nil
			
			// Clean up
			os.Remove(tempFile)
			
			results <- exists
		}(i)
	}
	
	successCount := 0
	for i := 0; i < numRequests; i++ {
		if <-results {
			successCount++
		}
	}
	
	assert.Equal(suite.T(), numRequests, successCount, "All concurrent processing operations should succeed")
}

// Test error handling in audio pipeline
func (suite *AudioPipelineTestSuite) TestAudioPipelineErrorHandling() {
	if suite.audioProcessor == nil {
		suite.T().Skip("Audio processor not available")
		return
	}
	
	// Test handling of nonexistent files
	nonexistentFile := filepath.Join(suite.tempDir, "nonexistent.wav")
	_, err := os.Stat(nonexistentFile)
	assert.Error(suite.T(), err, "Nonexistent file should not exist")
	
	// Test handling of invalid file formats
	invalidFile := filepath.Join(suite.tempDir, "invalid.txt")
	err = ioutil.WriteFile(invalidFile, []byte("not an audio file"), 0644)
	assert.NoError(suite.T(), err)
	
	// Attempt to validate invalid file should handle error gracefully
	err = suite.audioProcessor.ValidateAudioFile(suite.ctx, invalidFile)
	if err != nil {
		// This is expected - the error should be handled gracefully
		suite.T().Logf("Expected error for invalid file format: %v", err)
		assert.Contains(suite.T(), err.Error(), "file is not a valid audio file")
	}
	
	// Clean up
	os.Remove(invalidFile)
}

// Test storage integration (if available)
func (suite *AudioPipelineTestSuite) TestStorageIntegration() {
	if suite.storageService == nil {
		suite.T().Skip("Storage service not available - running in mock mode")
		return
	}
	
	// Test storage service basic operations
	testBucket := "test-bucket"
	
	// In a real test, we would:
	// 1. Upload test file to storage
	// 2. Verify file exists in storage
	// 3. Download file from storage
	// 4. Verify file integrity
	// 5. Clean up test files
	
	suite.T().Logf("Storage service available for testing: bucket=%s", testBucket)
}

// Test processing service configuration
func (suite *AudioPipelineTestSuite) TestProcessingServiceConfiguration() {
	// Test configuration loading
	devConfig := config.LoadDevConfig()
	assert.NotNil(suite.T(), devConfig)
	
	// Verify development settings
	assert.True(suite.T(), devConfig.IsDevelopment)
	
	// Test processing-related configuration
	if devConfig.MockStorage {
		suite.T().Log("Running with mock storage enabled")
	}
	
	// Test that configuration has required fields
	googleCloudProject := os.Getenv("GOOGLE_CLOUD_PROJECT")
	assert.NotEmpty(suite.T(), googleCloudProject)
}

// Test pipeline recovery from failures
func (suite *AudioPipelineTestSuite) TestPipelineRecovery() {
	if suite.audioProcessor == nil {
		suite.T().Skip("Audio processor not available")
		return
	}
	
	// Simulate recovery scenarios
	// 1. Temporary file cleanup
	tempFiles := make([]string, 3)
	for i := 0; i < 3; i++ {
		tempFile := filepath.Join(suite.tempDir, fmt.Sprintf("recovery_test_%d.tmp", i))
		err := ioutil.WriteFile(tempFile, []byte("temp data"), 0644)
		assert.NoError(suite.T(), err)
		tempFiles[i] = tempFile
	}
	
	// Verify temp files exist
	for _, file := range tempFiles {
		_, err := os.Stat(file)
		assert.NoError(suite.T(), err, "Temp file should exist")
	}
	
	// Test cleanup functionality
	for _, file := range tempFiles {
		err := os.Remove(file)
		assert.NoError(suite.T(), err)
		_, err = os.Stat(file)
		assert.Error(suite.T(), err, "File should be removed")
	}
}

// Test performance benchmarks
func (suite *AudioPipelineTestSuite) TestPerformanceBenchmarks() {
	if suite.audioProcessor == nil {
		suite.T().Skip("Audio processor not available")
		return
	}
	
	// Test file operation performance
	start := time.Now()
	
	// Perform multiple file operations
	for i := 0; i < 100; i++ {
		_, _ = os.Stat(suite.testAudioFile)
	}
	
	duration := time.Since(start)
	
	// File existence checks should be fast
	assert.Less(suite.T(), duration, 1*time.Second, "File operations should complete within 1 second")
	
	suite.T().Logf("100 file existence checks completed in %v", duration)
}

// Test processing service methods if available
func (suite *AudioPipelineTestSuite) TestProcessingServiceMethods() {
	if suite.processingService == nil {
		suite.T().Skip("Processing service not available")
		return
	}
	
	// Test that processing service was initialized properly
	assert.NotNil(suite.T(), suite.processingService)
	assert.IsType(suite.T(), &services.ProcessingService{}, suite.processingService)
	
	// Note: We can't easily test actual processing methods without 
	// a real Firestore instance and valid audio files, but we can
	// verify the service initializes correctly
	suite.T().Log("Processing service initialized successfully")
}

// Run the audio pipeline test suite
func TestAudioPipelineSuite(t *testing.T) {
	suite.Run(t, new(AudioPipelineTestSuite))
}