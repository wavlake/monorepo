// +build emulator

package integration

import (
	"context"
	"io"
	"os"
	"testing"
	"time"

	firebase "firebase.google.com/go/v4"
	"google.golang.org/api/option"

	"github.com/wavlake/monorepo/internal/models"
	"github.com/wavlake/monorepo/internal/services"
	"github.com/wavlake/monorepo/tests/testutil"
)

// Mock storage service for NostrTrackService integration tests
type mockStorageService struct{}

func (m *mockStorageService) GeneratePresignedURL(ctx context.Context, objectName string, expiration time.Duration) (string, error) {
	return "https://mock-storage.example.com/presigned/" + objectName, nil
}

func (m *mockStorageService) GetPublicURL(objectName string) string {
	return "https://mock-storage.example.com/public/" + objectName
}

func (m *mockStorageService) UploadObject(ctx context.Context, objectName string, data io.Reader, contentType string) error {
	return nil
}

func (m *mockStorageService) CopyObject(ctx context.Context, srcObject, dstObject string) error {
	return nil
}

func (m *mockStorageService) DeleteObject(ctx context.Context, objectName string) error {
	return nil
}

func (m *mockStorageService) GetObjectMetadata(ctx context.Context, objectName string) (interface{}, error) {
	return nil, nil
}

func (m *mockStorageService) GetObjectReader(ctx context.Context, objectName string) (io.ReadCloser, error) {
	return nil, nil
}

func (m *mockStorageService) GetBucketName() string {
	return "test-bucket"
}

func (m *mockStorageService) Close() error {
	return nil
}

// Mock path config for NostrTrackService integration tests
type mockPathConfig struct{}

func (m *mockPathConfig) GetOriginalPath(trackID, extension string) string {
	return "original/" + trackID + "." + extension
}

func (m *mockPathConfig) GetCompressedPath(trackID string) string {
	return "compressed/" + trackID + ".mp3"
}

// TestNostrTrackServiceWithFirebaseEmulators tests the actual NostrTrackService implementation
// with real Firebase emulator instances
func TestNostrTrackServiceWithFirebaseEmulators(t *testing.T) {
	// Ensure Firebase emulators are running
	if os.Getenv("FIRESTORE_EMULATOR_HOST") == "" {
		t.Skip("Firebase emulators not running. Run 'export FIRESTORE_EMULATOR_HOST=localhost:8081 && firebase emulators:start --only firestore,auth --project test-project' first.")
	}

	ctx := context.Background()

	// Initialize Firebase app for emulator testing
	config := &firebase.Config{
		ProjectID: "test-project",
	}

	// Initialize Firebase with emulator
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

	// Create mocks for dependencies
	mockStorage := &mockStorageService{}
	mockPaths := &mockPathConfig{}

	// Create NostrTrackService with real Firestore client and mock dependencies
	trackService := services.NewNostrTrackService(firestoreClient, mockStorage, mockPaths)

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

	t.Run("CreateTrack_RealImplementation", func(t *testing.T) {
		// Act
		track, err := trackService.CreateTrack(ctx, testPubkey, testFirebaseUID, testExtension)
		if err != nil {
			t.Fatalf("CreateTrack failed: %v", err)
		}

		// Verify track was created with correct data
		if track.FirebaseUID != testFirebaseUID {
			t.Errorf("Expected FirebaseUID %s, got %s", testFirebaseUID, track.FirebaseUID)
		}
		if track.Pubkey != testPubkey {
			t.Errorf("Expected Pubkey %s, got %s", testPubkey, track.Pubkey)
		}
		if track.Extension != testExtension {
			t.Errorf("Expected Extension %s, got %s", testExtension, track.Extension)
		}
		if !track.IsProcessing {
			t.Error("Expected track to be marked as processing")
		}
		if track.IsCompressed {
			t.Error("Expected track to not be compressed initially")
		}
		if track.Deleted {
			t.Error("Expected track to not be deleted")
		}

		// Verify data was actually written to Firestore
		trackDoc, err := firestoreClient.Collection("nostr_tracks").Doc(track.ID).Get(ctx)
		if err != nil {
			t.Fatalf("Failed to get track document: %v", err)
		}

		var storedTrack models.NostrTrack
		err = trackDoc.DataTo(&storedTrack)
		if err != nil {
			t.Fatalf("Failed to parse track data: %v", err)
		}

		// Verify stored track matches created track
		if storedTrack.ID != track.ID {
			t.Errorf("Expected ID %s, got %s", track.ID, storedTrack.ID)
		}
		if storedTrack.FirebaseUID != testFirebaseUID {
			t.Errorf("Expected FirebaseUID %s, got %s", testFirebaseUID, storedTrack.FirebaseUID)
		}
		if storedTrack.Pubkey != testPubkey {
			t.Errorf("Expected Pubkey %s, got %s", testPubkey, storedTrack.Pubkey)
		}
	})

	t.Run("GetTrack_RealImplementation", func(t *testing.T) {
		// Setup: Create a track first
		createdTrack, err := trackService.CreateTrack(ctx, testPubkey, testFirebaseUID, testExtension)
		if err != nil {
			t.Fatalf("Setup failed: %v", err)
		}

		// Act
		retrievedTrack, err := trackService.GetTrack(ctx, createdTrack.ID)
		if err != nil {
			t.Fatalf("GetTrack failed: %v", err)
		}

		// Assert
		if retrievedTrack.ID != createdTrack.ID {
			t.Errorf("Expected ID %s, got %s", createdTrack.ID, retrievedTrack.ID)
		}
		if retrievedTrack.FirebaseUID != testFirebaseUID {
			t.Errorf("Expected FirebaseUID %s, got %s", testFirebaseUID, retrievedTrack.FirebaseUID)
		}
		if retrievedTrack.Pubkey != testPubkey {
			t.Errorf("Expected Pubkey %s, got %s", testPubkey, retrievedTrack.Pubkey)
		}
		if retrievedTrack.Extension != testExtension {
			t.Errorf("Expected Extension %s, got %s", testExtension, retrievedTrack.Extension)
		}
	})

	t.Run("GetTracksByPubkey_RealImplementation", func(t *testing.T) {
		// Clean up before this test
		cleanup()
		
		// Setup: Create multiple tracks for the same pubkey
		track1, err := trackService.CreateTrack(ctx, testPubkey, testFirebaseUID, "wav")
		if err != nil {
			t.Fatalf("Setup failed creating track1: %v", err)
		}

		track2, err := trackService.CreateTrack(ctx, testPubkey, testFirebaseUID, "mp3")
		if err != nil {
			t.Fatalf("Setup failed creating track2: %v", err)
		}

		// Act
		tracks, err := trackService.GetTracksByPubkey(ctx, testPubkey)
		if err != nil {
			t.Fatalf("GetTracksByPubkey failed: %v", err)
		}

		// Assert
		if len(tracks) != 2 {
			t.Errorf("Expected 2 tracks, got %d", len(tracks))
		}

		// Verify tracks are ordered by created_at desc (newest first)
		trackIDs := make(map[string]bool)
		for _, track := range tracks {
			trackIDs[track.ID] = true
			if track.Pubkey != testPubkey {
				t.Errorf("Expected Pubkey %s, got %s", testPubkey, track.Pubkey)
			}
			if track.FirebaseUID != testFirebaseUID {
				t.Errorf("Expected FirebaseUID %s, got %s", testFirebaseUID, track.FirebaseUID)
			}
			if track.Deleted {
				t.Error("Expected track to not be deleted")
			}
		}

		// Verify both created tracks are present
		if !trackIDs[track1.ID] {
			t.Errorf("Expected track1 ID %s to be in results", track1.ID)
		}
		if !trackIDs[track2.ID] {
			t.Errorf("Expected track2 ID %s to be in results", track2.ID)
		}
	})

	t.Run("UpdateTrack_RealImplementation", func(t *testing.T) {
		// Setup: Create a track first
		track, err := trackService.CreateTrack(ctx, testPubkey, testFirebaseUID, testExtension)
		if err != nil {
			t.Fatalf("Setup failed: %v", err)
		}

		// Act: Update track metadata
		updates := map[string]interface{}{
			"is_processing": false,
			"size":          int64(1024000),
			"duration":      180, // 3 minutes
		}
		err = trackService.UpdateTrack(ctx, track.ID, updates)
		if err != nil {
			t.Fatalf("UpdateTrack failed: %v", err)
		}

		// Verify updates were applied in Firestore
		trackDoc, err := firestoreClient.Collection("nostr_tracks").Doc(track.ID).Get(ctx)
		if err != nil {
			t.Fatalf("Failed to get updated track document: %v", err)
		}

		var updatedTrack models.NostrTrack
		err = trackDoc.DataTo(&updatedTrack)
		if err != nil {
			t.Fatalf("Failed to parse updated track data: %v", err)
		}

		// Assert updates were applied
		if updatedTrack.IsProcessing {
			t.Error("Expected track to not be processing after update")
		}
		if updatedTrack.Size != 1024000 {
			t.Errorf("Expected Size 1024000, got %d", updatedTrack.Size)
		}
		if updatedTrack.Duration != 180 {
			t.Errorf("Expected Duration 180, got %d", updatedTrack.Duration)
		}
		// Verify UpdatedAt was automatically updated
		if !updatedTrack.UpdatedAt.After(updatedTrack.CreatedAt) {
			t.Error("Expected UpdatedAt to be after CreatedAt")
		}
	})

	t.Run("MarkTrackAsProcessed_RealImplementation", func(t *testing.T) {
		// Setup: Create a track first
		track, err := trackService.CreateTrack(ctx, testPubkey, testFirebaseUID, testExtension)
		if err != nil {
			t.Fatalf("Setup failed: %v", err)
		}

		// Act: Mark track as processed
		size := int64(2048000)
		duration := 240 // 4 minutes
		err = trackService.MarkTrackAsProcessed(ctx, track.ID, size, duration)
		if err != nil {
			t.Fatalf("MarkTrackAsProcessed failed: %v", err)
		}

		// Verify updates were applied
		retrievedTrack, err := trackService.GetTrack(ctx, track.ID)
		if err != nil {
			t.Fatalf("Failed to get processed track: %v", err)
		}

		if retrievedTrack.IsProcessing {
			t.Error("Expected track to not be processing after marking as processed")
		}
		if retrievedTrack.Size != size {
			t.Errorf("Expected Size %d, got %d", size, retrievedTrack.Size)
		}
		if retrievedTrack.Duration != duration {
			t.Errorf("Expected Duration %d, got %d", duration, retrievedTrack.Duration)
		}
	})

	t.Run("DeleteTrack_RealImplementation", func(t *testing.T) {
		// Setup: Create a track first
		track, err := trackService.CreateTrack(ctx, testPubkey, testFirebaseUID, testExtension)
		if err != nil {
			t.Fatalf("Setup failed: %v", err)
		}

		// Act: Soft delete the track
		err = trackService.DeleteTrack(ctx, track.ID)
		if err != nil {
			t.Fatalf("DeleteTrack failed: %v", err)
		}

		// Verify track is marked as deleted
		deletedTrack, err := trackService.GetTrack(ctx, track.ID)
		if err != nil {
			t.Fatalf("Failed to get deleted track: %v", err)
		}

		if !deletedTrack.Deleted {
			t.Error("Expected track to be marked as deleted")
		}

		// Verify deleted tracks are not returned in pubkey queries
		tracks, err := trackService.GetTracksByPubkey(ctx, testPubkey)
		if err != nil {
			t.Fatalf("GetTracksByPubkey failed: %v", err)
		}

		for _, returnedTrack := range tracks {
			if returnedTrack.ID == track.ID {
				t.Errorf("Deleted track %s should not appear in pubkey query results", track.ID)
			}
		}
	})

	t.Run("AddCompressionVersion_RealImplementation", func(t *testing.T) {
		// Setup: Create a track first
		track, err := trackService.CreateTrack(ctx, testPubkey, testFirebaseUID, testExtension)
		if err != nil {
			t.Fatalf("Setup failed: %v", err)
		}

		// Act: Add compression version
		version := models.CompressionVersion{
			ID:       "mp3-128",
			URL:      "https://storage.example.com/compressed/track-mp3-128.mp3",
			Bitrate:  128,
			Format:   "mp3",
			IsPublic: true,
		}
		err = trackService.AddCompressionVersion(ctx, track.ID, version)
		if err != nil {
			t.Fatalf("AddCompressionVersion failed: %v", err)
		}

		// Verify compression version was added
		updatedTrack, err := trackService.GetTrack(ctx, track.ID)
		if err != nil {
			t.Fatalf("Failed to get updated track: %v", err)
		}

		if len(updatedTrack.CompressionVersions) != 1 {
			t.Errorf("Expected 1 compression version, got %d", len(updatedTrack.CompressionVersions))
		}

		if len(updatedTrack.CompressionVersions) > 0 {
			addedVersion := updatedTrack.CompressionVersions[0]
			if addedVersion.ID != version.ID {
				t.Errorf("Expected version ID %s, got %s", version.ID, addedVersion.ID)
			}
			if addedVersion.URL != version.URL {
				t.Errorf("Expected URL %s, got %s", version.URL, addedVersion.URL)
			}
			if addedVersion.Bitrate != version.Bitrate {
				t.Errorf("Expected Bitrate %d, got %d", version.Bitrate, addedVersion.Bitrate)
			}
			if addedVersion.Format != version.Format {
				t.Errorf("Expected Format %s, got %s", version.Format, addedVersion.Format)
			}
			if addedVersion.IsPublic != version.IsPublic {
				t.Errorf("Expected IsPublic %t, got %t", version.IsPublic, addedVersion.IsPublic)
			}
		}

		if updatedTrack.HasPendingCompression {
			t.Error("Expected HasPendingCompression to be false after adding compression version")
		}
	})
}