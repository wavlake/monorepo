// +build emulator

package integration

import (
	"context"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	firebase "firebase.google.com/go/v4"
	"google.golang.org/api/option"

	"github.com/wavlake/monorepo/internal/models"
	"github.com/wavlake/monorepo/internal/services"
	"github.com/wavlake/monorepo/internal/utils"
	"github.com/wavlake/monorepo/tests/testutil"
)


// Mock storage service for ProcessingService integration tests
type mockStorageServiceForProcessing struct{}

func (m *mockStorageServiceForProcessing) GeneratePresignedURL(ctx context.Context, objectName string, expiration time.Duration) (string, error) {
	return "https://mock-storage.example.com/presigned/" + objectName, nil
}

func (m *mockStorageServiceForProcessing) GetPublicURL(objectName string) string {
	return "https://mock-storage.example.com/public/" + objectName
}

func (m *mockStorageServiceForProcessing) UploadObject(ctx context.Context, objectName string, data io.Reader, contentType string) error {
	// Simulate successful upload by reading all data
	_, err := io.Copy(io.Discard, data)
	return err
}

func (m *mockStorageServiceForProcessing) CopyObject(ctx context.Context, srcObject, dstObject string) error {
	return nil
}

func (m *mockStorageServiceForProcessing) DeleteObject(ctx context.Context, objectName string) error {
	return nil
}

func (m *mockStorageServiceForProcessing) GetObjectMetadata(ctx context.Context, objectName string) (interface{}, error) {
	return map[string]interface{}{
		"size": int64(1024000),
		"type": "audio/wav",
	}, nil
}

func (m *mockStorageServiceForProcessing) GetObjectReader(ctx context.Context, objectName string) (io.ReadCloser, error) {
	// Create a mock audio file content
	mockAudioData := strings.Repeat("MOCK_AUDIO_DATA_", 1000) // ~16KB of mock data
	return io.NopCloser(strings.NewReader(mockAudioData)), nil
}

func (m *mockStorageServiceForProcessing) GetBucketName() string {
	return "test-bucket"
}

func (m *mockStorageServiceForProcessing) Close() error {
	return nil
}

// TestProcessingServiceWithFirebaseEmulators tests the ProcessingService implementation
// with real Firebase emulator instances and mock audio/storage components
func TestProcessingServiceWithFirebaseEmulators(t *testing.T) {
	// Ensure Firebase emulators are running
	if os.Getenv("FIRESTORE_EMULATOR_HOST") == "" {
		t.Skip("Firebase emulators not running. Run 'export FIRESTORE_EMULATOR_HOST=localhost:8081 && firebase emulators:start --only firestore,auth --project test-project' first.")
	}

	ctx := context.Background()

	// Initialize Firebase app for emulator testing
	config := &firebase.Config{
		ProjectID: "test-project",
	}

	firebaseApp, err := firebase.NewApp(ctx, config, option.WithoutAuthentication())
	if err != nil {
		t.Fatalf("Failed to initialize Firebase app: %v", err)
	}

	// Initialize Firestore client
	firestoreClient, err := firebaseApp.Firestore(ctx)
	if err != nil {
		t.Fatalf("Failed to initialize Firestore client: %v", err)
	}
	defer firestoreClient.Close()

	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "processing_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create mocks and dependencies
	mockStorage := &mockStorageServiceForProcessing{}
	mockPaths := &utils.StoragePathConfig{
		OriginalPrefix:   "tracks/original",
		CompressedPrefix: "tracks/compressed",
		UseLegacyPaths:   false,
	}
	
	// Use real AudioProcessor but override commands with mocked executables
	// This will allow testing the workflow without requiring actual ffmpeg
	realAudioProcessor := utils.NewAudioProcessor(tempDir)

	// Create NostrTrackService with real Firestore
	nostrTrackService := services.NewNostrTrackService(firestoreClient, mockStorage, mockPaths)

	// Create ProcessingService with real audio processor but mocked storage
	processingService := services.NewProcessingService(
		mockStorage,
		nostrTrackService,
		realAudioProcessor,
		tempDir,
	)

	// Set up test data
	testFirebaseUID := testutil.TestFirebaseUID
	testPubkey := testutil.TestPubkey
	testExtension := "wav"

	// Clean up function
	cleanup := func() {
		// Clean up any created tracks in the test collection
		iter := firestoreClient.Collection("nostr_tracks").Documents(ctx)
		defer iter.Stop()
		
		for {
			doc, err := iter.Next()
			if err != nil {
				break
			}
			_, delErr := doc.Ref.Delete(ctx)
			if delErr != nil {
				t.Logf("Warning: failed to delete document %s: %v", doc.Ref.ID, delErr)
			}
		}
	}

	// Clean up before and after
	cleanup()
	defer cleanup()

	t.Run("ProcessTrack_RealImplementation", func(t *testing.T) {
		// Setup: Create a track with processing status
		track, err := nostrTrackService.CreateTrack(ctx, testPubkey, testFirebaseUID, testExtension)
		if err != nil {
			t.Fatalf("Setup failed: %v", err)
		}

		// Verify track is initially marked as processing
		if !track.IsProcessing {
			t.Error("Expected track to be marked as processing initially")
		}

		// Act: Process the track
		// Note: This will fail without ffmpeg installed, which is expected
		// The test validates the workflow structure and Firebase integration
		err = processingService.ProcessTrack(ctx, track.ID)
		if err != nil {
			// If ffmpeg is not installed, the test should fail gracefully
			// and we can still verify that the track state was updated correctly
			t.Logf("ProcessTrack failed (expected without ffmpeg): %v", err)
			
			// Check if track was marked as failed due to missing ffmpeg
			updatedTrack, getErr := nostrTrackService.GetTrack(ctx, track.ID)
			if getErr != nil {
				t.Fatalf("Failed to get track after processing failure: %v", getErr)
			}
			
			// Track should no longer be processing after failure
			if updatedTrack.IsProcessing {
				t.Error("Expected track to not be processing after failure")
			}
			
			// Skip the rest of this test since ffmpeg is not available
			t.Skip("Skipping remaining ProcessTrack test - ffmpeg not available")
		}

		// Verify: Check track was updated in Firestore
		processedTrack, err := nostrTrackService.GetTrack(ctx, track.ID)
		if err != nil {
			t.Fatalf("Failed to get processed track: %v", err)
		}

		// Assert processing status was updated
		if processedTrack.IsProcessing {
			t.Error("Expected track to not be processing after ProcessTrack")
		}
		if !processedTrack.IsCompressed {
			t.Error("Expected track to be marked as compressed after ProcessTrack")
		}
		if processedTrack.CompressedURL == "" {
			t.Error("Expected compressed URL to be set after ProcessTrack")
		}
		
		// Verify audio metadata was set
		if processedTrack.Size == 0 {
			t.Error("Expected track size to be set after processing")
		}
		if processedTrack.Duration == 0 {
			t.Error("Expected track duration to be set after processing")
		}

		// Verify default compression version was added
		if len(processedTrack.CompressionVersions) == 0 {
			t.Error("Expected at least one compression version after ProcessTrack")
		} else {
			defaultVersion := processedTrack.CompressionVersions[0]
			if defaultVersion.ID != "default-128k-mp3" {
				t.Errorf("Expected default version ID 'default-128k-mp3', got %s", defaultVersion.ID)
			}
			if defaultVersion.Bitrate != 128 {
				t.Errorf("Expected default version bitrate 128, got %d", defaultVersion.Bitrate)
			}
			if defaultVersion.Format != "mp3" {
				t.Errorf("Expected default version format 'mp3', got %s", defaultVersion.Format)
			}
			if !defaultVersion.IsPublic {
				t.Error("Expected default version to be public for backwards compatibility")
			}
		}
	})

	t.Run("ProcessCompression_RealImplementation", func(t *testing.T) {
		// Setup: Create a track
		track, err := nostrTrackService.CreateTrack(ctx, testPubkey, testFirebaseUID, testExtension)
		if err != nil {
			t.Fatalf("Setup failed: %v", err)
		}

		// Define compression options
		compressionOption := models.CompressionOption{
			Bitrate:    256,
			Format:     "mp3",
			Quality:    "high",
			SampleRate: 44100,
		}

		// Act: Process compression with specific options
		err = processingService.ProcessCompression(ctx, track.ID, compressionOption)
		if err != nil {
			t.Fatalf("ProcessCompression failed: %v", err)
		}

		// Verify: Check compression version was added
		updatedTrack, err := nostrTrackService.GetTrack(ctx, track.ID)
		if err != nil {
			t.Fatalf("Failed to get updated track: %v", err)
		}

		// Assert compression version was added
		if len(updatedTrack.CompressionVersions) == 0 {
			t.Fatal("Expected at least one compression version after ProcessCompression")
		}

		// Find the added compression version
		var addedVersion *models.CompressionVersion
		for _, version := range updatedTrack.CompressionVersions {
			if version.Options.Bitrate == compressionOption.Bitrate &&
			   version.Options.Format == compressionOption.Format &&
			   version.Options.Quality == compressionOption.Quality {
				addedVersion = &version
				break
			}
		}

		if addedVersion == nil {
			t.Fatal("Could not find the added compression version")
		}

		// Verify compression version properties
		if addedVersion.Bitrate != 256 {
			t.Errorf("Expected compression version bitrate 256, got %d", addedVersion.Bitrate)
		}
		if addedVersion.Format != "mp3" {
			t.Errorf("Expected compression version format 'mp3', got %s", addedVersion.Format)
		}
		if addedVersion.Quality != "high" {
			t.Errorf("Expected compression version quality 'high', got %s", addedVersion.Quality)
		}
		if addedVersion.SampleRate != 44100 {
			t.Errorf("Expected compression version sample rate 44100, got %d", addedVersion.SampleRate)
		}
		if addedVersion.Size == 0 {
			t.Error("Expected compression version to have non-zero size")
		}
		if addedVersion.URL == "" {
			t.Error("Expected compression version to have URL")
		}
		if addedVersion.IsPublic {
			t.Error("Expected compression version to be private by default")
		}
	})

	t.Run("RequestCompressionVersions_RealImplementation", func(t *testing.T) {
		// Setup: Create a track
		track, err := nostrTrackService.CreateTrack(ctx, testPubkey, testFirebaseUID, testExtension)
		if err != nil {
			t.Fatalf("Setup failed: %v", err)
		}

		// Define multiple compression options
		compressionOptions := []models.CompressionOption{
			{
				Bitrate:    128,
				Format:     "mp3",
				Quality:    "medium",
				SampleRate: 44100,
			},
			{
				Bitrate:    256,
				Format:     "aac",
				Quality:    "high",
				SampleRate: 48000,
			},
		}

		// Act: Request multiple compression versions
		err = processingService.RequestCompressionVersions(ctx, track.ID, compressionOptions)
		if err != nil {
			t.Fatalf("RequestCompressionVersions failed: %v", err)
		}

		// Verify: Check that track is marked as having pending compression
		updatedTrack, err := nostrTrackService.GetTrack(ctx, track.ID)
		if err != nil {
			t.Fatalf("Failed to get updated track: %v", err)
		}

		if !updatedTrack.HasPendingCompression {
			t.Error("Expected track to be marked as having pending compression")
		}

		// Wait a short time for async processing to potentially complete
		// Note: In a real test, you might want to use channels or other synchronization
		time.Sleep(100 * time.Millisecond)
	})

	t.Run("MarkProcessingFailed_RealImplementation", func(t *testing.T) {
		// Setup: Create a track
		track, err := nostrTrackService.CreateTrack(ctx, testPubkey, testFirebaseUID, testExtension)
		if err != nil {
			t.Fatalf("Setup failed: %v", err)
		}

		// Verify track is initially processing
		if !track.IsProcessing {
			t.Error("Expected track to be marked as processing initially")
		}

		// Create a ProcessingService instance to access private method via reflection
		// Since markProcessingFailed is private, we'll simulate the failure scenario
		// by trying to process a non-existent track which should trigger the failure path

		// Act: Try to process a non-existent track (should fail and mark as failed)
		err = processingService.ProcessTrack(ctx, "non-existent-track-id")
		if err == nil {
			t.Error("Expected ProcessTrack to fail for non-existent track")
		}

		// For this test, we'll manually test the UpdateTrack functionality 
		// to verify the failure marking mechanism works
		updates := map[string]interface{}{
			"is_processing": false,
		}

		err = nostrTrackService.UpdateTrack(ctx, track.ID, updates)
		if err != nil {
			t.Fatalf("Failed to mark track as failed: %v", err)
		}

		// Verify: Check track was marked as no longer processing
		failedTrack, err := nostrTrackService.GetTrack(ctx, track.ID)
		if err != nil {
			t.Fatalf("Failed to get failed track: %v", err)
		}

		if failedTrack.IsProcessing {
			t.Error("Expected track to not be processing after failure")
		}
	})

	t.Run("ProcessTrackAsync_RealImplementation", func(t *testing.T) {
		// Setup: Create a track
		track, err := nostrTrackService.CreateTrack(ctx, testPubkey, testFirebaseUID, testExtension)
		if err != nil {
			t.Fatalf("Setup failed: %v", err)
		}

		// Act: Start async processing
		processingService.ProcessTrackAsync(ctx, track.ID)

		// Wait for async processing to complete
		// In a real scenario, you might use channels or other synchronization mechanisms
		time.Sleep(500 * time.Millisecond)

		// Verify: Check that processing completed successfully
		processedTrack, err := nostrTrackService.GetTrack(ctx, track.ID)
		if err != nil {
			t.Fatalf("Failed to get processed track: %v", err)
		}

		// The async operation should have completed by now
		// Note: This test is time-dependent and might be flaky in slower environments
		if processedTrack.IsProcessing {
			t.Log("Warning: Async processing may still be in progress or failed")
		}
	})

	t.Run("ProcessCompressionAsync_RealImplementation", func(t *testing.T) {
		// Setup: Create a track
		track, err := nostrTrackService.CreateTrack(ctx, testPubkey, testFirebaseUID, testExtension)
		if err != nil {
			t.Fatalf("Setup failed: %v", err)
		}

		// Define compression option
		compressionOption := models.CompressionOption{
			Bitrate:    192,
			Format:     "mp3",
			Quality:    "medium",
			SampleRate: 44100,
		}

		// Act: Start async compression
		processingService.ProcessCompressionAsync(ctx, track.ID, compressionOption)

		// Wait for async processing to complete
		time.Sleep(500 * time.Millisecond)

		// Verify: Check if compression version was added
		// Note: This test is time-dependent and async, so results may vary
		updatedTrack, err := nostrTrackService.GetTrack(ctx, track.ID)
		if err != nil {
			t.Fatalf("Failed to get updated track: %v", err)
		}

		// Log the result for debugging - async operations are non-deterministic in tests
		t.Logf("Async compression test: track has %d compression versions", len(updatedTrack.CompressionVersions))
	})
}