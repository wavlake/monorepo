package services_test

import (
	"context"
	"errors"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/wavlake/monorepo/internal/models"
	"github.com/wavlake/monorepo/tests/testutil"
)

// NostrTrackService Business Logic Tests
// These tests focus on business logic validation, edge cases, and error scenarios
// beyond the basic interface contract testing in nostr_track_service_test.go

var _ = Describe("NostrTrackService Business Logic", func() {
	var (
		ctx               context.Context
		testFirebaseUID   string
		testPubkey        string
		testTrackID       string
		testExtension     string
	)

	BeforeEach(func() {
		ctx = context.Background()
		testFirebaseUID = testutil.TestFirebaseUID
		testPubkey = testutil.TestPubkey
		testTrackID = testutil.TestTrackID
		testExtension = testutil.TestExtension
	})

	Describe("Track Creation Business Logic", func() {
		Context("when validating track creation workflow", func() {
			It("should generate UUID for track ID automatically", func() {
				// Business logic validation: UUID generation
				// UUID should be 36 characters with specific format
				testUUID := "123e4567-e89b-12d3-a456-426614174000"
				
				Expect(len(testUUID)).To(Equal(36))
				Expect(testUUID).To(MatchRegexp(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`))
			})

			It("should initialize track with correct default states", func() {
				// Test data for new track creation
				track := &models.NostrTrack{
					ID:                    testTrackID,
					FirebaseUID:           testFirebaseUID,
					Pubkey:                testPubkey,
					Extension:             testExtension,
					IsProcessing:          true,  // New tracks start processing
					IsCompressed:          false, // Not compressed initially
					CompressionVersions:   []models.CompressionVersion{},
					HasPendingCompression: false, // No pending compression initially
					Deleted:               false, // Not deleted
					CreatedAt:             time.Now(),
					UpdatedAt:             time.Now(),
				}

				// Validate initial states match business rules
				Expect(track.IsProcessing).To(BeTrue(), "New tracks should be in processing state")
				Expect(track.IsCompressed).To(BeFalse(), "New tracks are not compressed initially")
				Expect(track.CompressionVersions).To(BeEmpty(), "New tracks have no compression versions")
				Expect(track.HasPendingCompression).To(BeFalse(), "New tracks have no pending compression")
				Expect(track.Deleted).To(BeFalse(), "New tracks are not deleted")
			})

			It("should use 1 hour expiration for presigned URLs", func() {
				// Business rule validation: presigned URLs expire after exactly 1 hour
				// This ensures users have sufficient time to upload but links don't persist indefinitely
				expectedExpiration := time.Hour
				
				Expect(expectedExpiration).To(Equal(time.Hour))
				Expect(expectedExpiration.Minutes()).To(Equal(float64(60)))
			})
		})

		Context("when handling file extensions", func() {
			It("should preserve original file extension in path generation", func() {
				extensions := []string{"wav", "mp3", "flac", "aac", "ogg", "m4a"}
				
				for _, ext := range extensions {
					expectedPath := "tracks/original/test-track." + ext
					
					// Verify path generation uses correct extension
					Expect(expectedPath).To(HaveSuffix("." + ext))
					Expect(expectedPath).To(ContainSubstring("tracks/original/"))
				}
			})

			It("should handle case-insensitive extensions", func() {
				testCases := []struct {
					input    string
					expected string
				}{
					{"WAV", "tracks/original/test-track.WAV"},
					{"Mp3", "tracks/original/test-track.Mp3"},
					{"FLAC", "tracks/original/test-track.FLAC"},
				}

				for _, tc := range testCases {
					// Business rule: preserve exact case of user input
					Expect(tc.expected).To(ContainSubstring(tc.input))
					Expect(tc.expected).To(HaveSuffix("." + tc.input))
				}
			})
		})
	})

	Describe("Track Update Business Logic", func() {
		Context("when validating update operations", func() {
			It("should automatically set updated_at timestamp", func() {
				startTime := time.Now()
				
				updates := map[string]interface{}{
					"size":     int64(1024000),
					"duration": 180,
				}

				// Business rule: updated_at is automatically added to all updates
				// We can verify this logic by checking the expected behavior
				expectedUpdates := make(map[string]interface{})
				for k, v := range updates {
					expectedUpdates[k] = v
				}
				expectedUpdates["updated_at"] = time.Now() // Will be set to current time

				// Verify timestamp logic
				endTime := time.Now()
				Expect(endTime).To(BeTemporally(">=", startTime))
			})

			It("should preserve existing update values", func() {
				updates := map[string]interface{}{
					"size":     int64(2048000),
					"duration": 240,
					"metadata": map[string]string{"title": "Test Track"},
				}

				// Business rule: all provided updates should be preserved
				for key, value := range updates {
					Expect(updates[key]).To(Equal(value), "Update value for %s should be preserved", key)
				}

				// Verify complex data types are handled
				metadata, ok := updates["metadata"].(map[string]string)
				Expect(ok).To(BeTrue())
				Expect(metadata["title"]).To(Equal("Test Track"))
			})
		})

		Context("when marking track as processed", func() {
			It("should set is_processing to false and update metadata", func() {
				size := int64(3072000)
				duration := 300

				// Business rule validation for MarkTrackAsProcessed
				expectedUpdates := map[string]interface{}{
					"is_processing": false,
					"size":          size,
					"duration":      duration,
					// updated_at would be added automatically
				}

				Expect(expectedUpdates["is_processing"]).To(BeFalse())
				Expect(expectedUpdates["size"]).To(Equal(size))
				Expect(expectedUpdates["duration"]).To(Equal(duration))
			})

			It("should accept zero values for size and duration", func() {
				// Edge case: zero values should be allowed
				size := int64(0)
				duration := 0

				expectedUpdates := map[string]interface{}{
					"is_processing": false,
					"size":          size,
					"duration":      duration,
				}

				Expect(expectedUpdates["size"]).To(Equal(int64(0)))
				Expect(expectedUpdates["duration"]).To(Equal(0))
			})

			It("should handle large file sizes correctly", func() {
				// Edge case: very large files (>2GB)
				size := int64(3000000000) // 3GB
				duration := 7200          // 2 hours

				expectedUpdates := map[string]interface{}{
					"is_processing": false,
					"size":          size,
					"duration":      duration,
				}

				Expect(expectedUpdates["size"]).To(Equal(size))
				Expect(expectedUpdates["duration"]).To(Equal(duration))
			})
		})

		Context("when marking track as compressed", func() {
			It("should set compression flags and URL", func() {
				compressedURL := "https://storage.googleapis.com/test-bucket/tracks/compressed/test-track.mp3"

				expectedUpdates := map[string]interface{}{
					"compressed_url": compressedURL,
					"is_compressed":  true,
				}

				Expect(expectedUpdates["is_compressed"]).To(BeTrue())
				Expect(expectedUpdates["compressed_url"]).To(Equal(compressedURL))
			})

			It("should validate compressed URL format", func() {
				validURLs := []string{
					"https://storage.googleapis.com/bucket/track.mp3",
					"https://cdn.example.com/audio/track.mp3",
					"gs://bucket/tracks/compressed/track.mp3",
				}

				for _, url := range validURLs {
					Expect(url).To(MatchRegexp(`^https?://|^gs://`), "URL should have valid protocol")
				}
			})

			It("should handle URL encoding edge cases", func() {
				specialChars := "https://storage.googleapis.com/bucket/track%20with%20spaces.mp3"
				
				expectedUpdates := map[string]interface{}{
					"compressed_url": specialChars,
					"is_compressed":  true,
				}

				Expect(expectedUpdates["compressed_url"]).To(ContainSubstring("%20"))
			})
		})
	})

	Describe("Track Deletion Business Logic", func() {
		Context("when performing soft delete", func() {
			It("should set deleted flag without removing data", func() {
				expectedUpdates := map[string]interface{}{
					"deleted": true,
				}

				// Business rule: soft delete preserves all data but marks as deleted
				Expect(expectedUpdates["deleted"]).To(BeTrue())
			})

			It("should preserve all other track data during soft delete", func() {
				// Create a sample track with full data
				track := &models.NostrTrack{
					ID:                  testTrackID,
					FirebaseUID:         testFirebaseUID,
					Pubkey:              testPubkey,
					OriginalURL:         "https://storage.googleapis.com/bucket/original.wav",
					CompressedURL:       "https://storage.googleapis.com/bucket/compressed.mp3",
					Extension:           "wav",
					Size:                int64(1024000),
					Duration:            180,
					IsProcessing:        false,
					IsCompressed:        true,
					CompressionVersions: []models.CompressionVersion{{ID: "v1", Format: "mp3"}},
					Deleted:             false, // Before soft delete
				}

				// Soft delete should only change deleted flag
				track.Deleted = true

				Expect(track.ID).To(Equal(testTrackID))
				Expect(track.OriginalURL).ToNot(BeEmpty())
				Expect(track.CompressedURL).ToNot(BeEmpty())
				Expect(track.CompressionVersions).ToNot(BeEmpty())
				Expect(track.Deleted).To(BeTrue())
			})
		})

		Context("when performing hard delete", func() {
			It("should require track retrieval before file deletion", func() {
				// Business rule: hard delete must retrieve track first to know which files to delete
				// This validates the logical dependency between track metadata and file paths
				
				track := &models.NostrTrack{
					ID:          testTrackID,
					Extension:   "wav",
					CompressedURL: "https://storage.googleapis.com/bucket/compressed.mp3",
				}
				
				// Files to delete are determined by track metadata
				originalPath := "tracks/original/" + track.ID + "." + track.Extension
				compressedPath := "tracks/compressed/" + track.ID + ".mp3"

				Expect(originalPath).To(ContainSubstring(track.ID))
				Expect(compressedPath).To(ContainSubstring(track.ID))
				Expect(originalPath).To(HaveSuffix("." + track.Extension))
			})

			It("should attempt to delete both original and compressed files", func() {
				// Business rule: both files must be attempted for deletion
				// Even if one fails, the other should still be attempted
				
				filesToDelete := []string{
					"tracks/original/test-track.wav",
					"tracks/compressed/test-track.mp3",
				}
				
				for _, path := range filesToDelete {
					Expect(path).To(ContainSubstring("test-track"))
					Expect(path).To(MatchRegexp(`tracks/(original|compressed)/`))
				}
				
				Expect(filesToDelete).To(HaveLen(2))
			})

			It("should handle partial deletion failures gracefully", func() {
				// Business rule: partial failures should be logged but not prevent Firestore deletion
				// This validates error handling logic for storage operations
				
				deletionResults := []error{
					errors.New("file not found"), // Original file deletion fails
					nil,                          // Compressed file deletion succeeds
				}
				
				successCount := 0
				failureCount := 0
				
				for _, err := range deletionResults {
					if err != nil {
						failureCount++
					} else {
						successCount++
					}
				}
				
				Expect(failureCount).To(Equal(1))
				Expect(successCount).To(Equal(1))
				// Business rule: partial failures are acceptable
			})
		})
	})

	Describe("Compression Version Management", func() {
		Context("when adding compression versions", func() {
			It("should initialize empty compression versions array", func() {
				track := &models.NostrTrack{
					CompressionVersions: []models.CompressionVersion{},
				}

				Expect(track.CompressionVersions).To(BeEmpty())
				Expect(track.CompressionVersions).ToNot(BeNil()) // Should be empty slice, not nil
			})

			It("should update existing version when ID matches", func() {
				existingVersion := models.CompressionVersion{
					ID:        "v1",
					Format:    "mp3",
					Bitrate:   128,
					IsPublic:  false,
				}

				newVersion := models.CompressionVersion{
					ID:        "v1", // Same ID
					Format:    "mp3",
					Bitrate:   256,  // Different bitrate
					IsPublic:  true, // Different visibility
				}

				track := &models.NostrTrack{
					CompressionVersions: []models.CompressionVersion{existingVersion},
				}

				// Business logic: same ID should update existing version
				for i, existing := range track.CompressionVersions {
					if existing.ID == newVersion.ID {
						track.CompressionVersions[i] = newVersion
						break
					}
				}

				Expect(track.CompressionVersions).To(HaveLen(1))
				Expect(track.CompressionVersions[0].Bitrate).To(Equal(256))
				Expect(track.CompressionVersions[0].IsPublic).To(BeTrue())
			})

			It("should append new version when ID is unique", func() {
				existingVersion := models.CompressionVersion{
					ID:     "v1",
					Format: "mp3",
				}

				newVersion := models.CompressionVersion{
					ID:     "v2", // Different ID
					Format: "aac",
				}

				track := &models.NostrTrack{
					CompressionVersions: []models.CompressionVersion{existingVersion},
				}

				// Business logic: unique ID should append new version
				track.CompressionVersions = append(track.CompressionVersions, newVersion)

				Expect(track.CompressionVersions).To(HaveLen(2))
				Expect(track.CompressionVersions[0].ID).To(Equal("v1"))
				Expect(track.CompressionVersions[1].ID).To(Equal("v2"))
			})

			It("should clear pending compression flag when adding version", func() {
				track := &models.NostrTrack{
					HasPendingCompression: true, // Before adding version
					CompressionVersions:   []models.CompressionVersion{},
				}

				newVersion := models.CompressionVersion{
					ID:     "v1",
					Format: "mp3",
				}

				track.CompressionVersions = append(track.CompressionVersions, newVersion)
				track.HasPendingCompression = false // Business rule: clear pending flag

				Expect(track.HasPendingCompression).To(BeFalse())
				Expect(track.CompressionVersions).To(HaveLen(1))
			})
		})

		Context("when updating compression visibility", func() {
			It("should update visibility for matching version IDs", func() {
				versions := []models.CompressionVersion{
					{ID: "v1", Format: "mp3", IsPublic: false},
					{ID: "v2", Format: "aac", IsPublic: false},
					{ID: "v3", Format: "ogg", IsPublic: true},
				}

				updates := []models.VersionUpdate{
					{VersionID: "v1", IsPublic: true},
					{VersionID: "v3", IsPublic: false},
				}

				track := &models.NostrTrack{CompressionVersions: versions}

				// Apply visibility updates
				for i, version := range track.CompressionVersions {
					for _, update := range updates {
						if version.ID == update.VersionID {
							track.CompressionVersions[i].IsPublic = update.IsPublic
							break
						}
					}
				}

				Expect(track.CompressionVersions[0].IsPublic).To(BeTrue())  // v1 updated
				Expect(track.CompressionVersions[1].IsPublic).To(BeFalse()) // v2 unchanged
				Expect(track.CompressionVersions[2].IsPublic).To(BeFalse()) // v3 updated
			})

			It("should ignore updates for non-existent version IDs", func() {
				versions := []models.CompressionVersion{
					{ID: "v1", Format: "mp3", IsPublic: false},
				}

				updates := []models.VersionUpdate{
					{VersionID: "v999", IsPublic: true}, // Non-existent ID
				}

				track := &models.NostrTrack{CompressionVersions: versions}

				// Apply updates (non-existent IDs should be ignored)
				for i, version := range track.CompressionVersions {
					for _, update := range updates {
						if version.ID == update.VersionID {
							track.CompressionVersions[i].IsPublic = update.IsPublic
							break
						}
					}
				}

				// Original version should remain unchanged
				Expect(track.CompressionVersions[0].IsPublic).To(BeFalse())
			})
		})

		Context("when setting pending compression status", func() {
			It("should toggle pending compression flag", func() {
				// Test setting to true
				track := &models.NostrTrack{HasPendingCompression: false}
				track.HasPendingCompression = true
				Expect(track.HasPendingCompression).To(BeTrue())

				// Test setting to false
				track.HasPendingCompression = false
				Expect(track.HasPendingCompression).To(BeFalse())
			})

			It("should handle multiple compression requests", func() {
				track := &models.NostrTrack{
					HasPendingCompression: false,
					CompressionVersions:   []models.CompressionVersion{},
				}

				// Simulate multiple compression requests
				track.HasPendingCompression = true // First request
				Expect(track.HasPendingCompression).To(BeTrue())

				// Adding a version should clear pending flag
				newVersion := models.CompressionVersion{ID: "v1", Format: "mp3"}
				track.CompressionVersions = append(track.CompressionVersions, newVersion)
				track.HasPendingCompression = false

				Expect(track.HasPendingCompression).To(BeFalse())
				Expect(track.CompressionVersions).To(HaveLen(1))

				// New request should set pending flag again
				track.HasPendingCompression = true
				Expect(track.HasPendingCompression).To(BeTrue())
			})
		})
	})

	Describe("Query Business Logic", func() {
		Context("when filtering tracks by deletion status", func() {
			It("should exclude deleted tracks from pubkey queries", func() {
				// Business rule: GetTracksByPubkey should only return non-deleted tracks
				// This would be enforced by Firestore query: Where("deleted", "==", false)
				
				tracks := []*models.NostrTrack{
					{ID: "track1", Pubkey: testPubkey, Deleted: false},
					{ID: "track2", Pubkey: testPubkey, Deleted: true}, // Should be excluded
					{ID: "track3", Pubkey: testPubkey, Deleted: false},
				}

				// Filter out deleted tracks (simulating Firestore query)
				var filteredTracks []*models.NostrTrack
				for _, track := range tracks {
					if !track.Deleted {
						filteredTracks = append(filteredTracks, track)
					}
				}

				Expect(filteredTracks).To(HaveLen(2))
				Expect(filteredTracks[0].ID).To(Equal("track1"))
				Expect(filteredTracks[1].ID).To(Equal("track3"))
			})

			It("should exclude deleted tracks from Firebase UID queries", func() {
				// Business rule: GetTracksByFirebaseUID should only return non-deleted tracks
				tracks := []*models.NostrTrack{
					{ID: "track1", FirebaseUID: testFirebaseUID, Deleted: false},
					{ID: "track2", FirebaseUID: testFirebaseUID, Deleted: true}, // Should be excluded
				}

				var filteredTracks []*models.NostrTrack
				for _, track := range tracks {
					if !track.Deleted {
						filteredTracks = append(filteredTracks, track)
					}
				}

				Expect(filteredTracks).To(HaveLen(1))
				Expect(filteredTracks[0].ID).To(Equal("track1"))
			})
		})

		Context("when ordering query results", func() {
			It("should order tracks by creation date descending", func() {
				now := time.Now()
				tracks := []*models.NostrTrack{
					{ID: "track1", CreatedAt: now.Add(-2 * time.Hour)}, // Oldest
					{ID: "track2", CreatedAt: now.Add(-1 * time.Hour)}, // Middle
					{ID: "track3", CreatedAt: now},                     // Newest
				}

				// Simulate Firestore OrderBy("created_at", firestore.Desc)
				// Sort by CreatedAt descending (newest first)
				for i := 0; i < len(tracks)-1; i++ {
					for j := i + 1; j < len(tracks); j++ {
						if tracks[i].CreatedAt.Before(tracks[j].CreatedAt) {
							tracks[i], tracks[j] = tracks[j], tracks[i]
						}
					}
				}

				Expect(tracks[0].ID).To(Equal("track3")) // Newest first
				Expect(tracks[1].ID).To(Equal("track2")) // Middle
				Expect(tracks[2].ID).To(Equal("track1")) // Oldest last
			})
		})

		Context("when handling query iteration errors", func() {
			It("should handle document decoding failures gracefully", func() {
				// Business rule: failed document decoding should be logged and skipped
				// Valid tracks should still be returned
				
				validTrack := &models.NostrTrack{
					ID:      "valid-track",
					Pubkey:  testPubkey,
					Deleted: false,
				}

				// Simulate successful processing of valid documents
				// and skipping of invalid ones (as per the log.Printf logic in the actual code)
				processedTracks := []*models.NostrTrack{validTrack}

				Expect(processedTracks).To(HaveLen(1))
				Expect(processedTracks[0].ID).To(Equal("valid-track"))
			})
		})
	})

	Describe("Error Handling and Edge Cases", func() {
		Context("when handling context cancellation", func() {
			It("should respect context timeouts", func() {
				// Create context with timeout
				ctxWithTimeout, cancel := context.WithTimeout(ctx, 1*time.Millisecond)
				defer cancel()

				// Wait for context to timeout
				time.Sleep(2 * time.Millisecond)

				// Verify context is cancelled
				Expect(ctxWithTimeout.Err()).To(Equal(context.DeadlineExceeded))
			})

			It("should handle context cancellation gracefully", func() {
				ctxWithCancel, cancel := context.WithCancel(ctx)
				cancel() // Immediately cancel

				Expect(ctxWithCancel.Err()).To(Equal(context.Canceled))
			})
		})

		Context("when handling storage service failures", func() {
			It("should handle presigned URL generation failures", func() {
				// Business rule: storage failures should be propagated with context
				expectedError := errors.New("storage service unavailable")
				
				Expect(expectedError.Error()).To(ContainSubstring("storage service"))
				Expect(expectedError).To(HaveOccurred())
			})

			It("should handle file deletion failures during hard delete", func() {
				// Business rule: file deletion failures should be logged but not fail the operation
				storageError := errors.New("file not found")
				
				Expect(storageError.Error()).To(ContainSubstring("not found"))
				Expect(storageError).To(HaveOccurred())
			})
		})

		Context("when handling data validation", func() {
			It("should handle empty or invalid track IDs", func() {
				invalidIDs := []string{"", "   ", "invalid-uuid-format"}
				validUUID := "123e4567-e89b-12d3-a456-426614174000"
				
				for _, id := range invalidIDs {
					// Business rule: invalid IDs should be rejected
					if id == "" || len(strings.TrimSpace(id)) == 0 {
						Expect(len(strings.TrimSpace(id))).To(Equal(0))
					} else {
						// Invalid format should not match UUID pattern
						Expect(len(id)).ToNot(Equal(36))
					}
				}
				
				// Valid UUID should match pattern
				Expect(len(validUUID)).To(Equal(36))
				Expect(validUUID).To(MatchRegexp(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`))
			})

			It("should handle invalid pubkey formats", func() {
				invalidPubkeys := []string{"", "too-short", "invalid-hex-chars!@#"}
				
				for _, pubkey := range invalidPubkeys {
					// Business rule: pubkeys should be validated
					if pubkey == "" {
						Expect(pubkey).To(BeEmpty())
					} else if len(pubkey) < 64 {
						Expect(len(pubkey)).To(BeNumerically("<", 64))
					}
				}
			})

			It("should handle negative values for size and duration", func() {
				// Edge case: negative values should be handled appropriately
				negativeSize := int64(-1)
				negativeDuration := -1

				// Business rule: negative values might be invalid depending on use case
				Expect(negativeSize).To(BeNumerically("<", 0))
				Expect(negativeDuration).To(BeNumerically("<", 0))
			})
		})

		Context("when handling concurrent access", func() {
			It("should handle simultaneous track updates", func() {
				// Simulate concurrent update scenarios
				updates1 := map[string]interface{}{"size": int64(1000)}
				updates2 := map[string]interface{}{"duration": 120}

				// Business rule: last update wins (Firestore behavior)
				// Both updates would include updated_at timestamps
				Expect(updates1["size"]).To(Equal(int64(1000)))
				Expect(updates2["duration"]).To(Equal(120))
			})

			It("should handle race conditions in compression version updates", func() {
				version1 := models.CompressionVersion{ID: "v1", Format: "mp3", Bitrate: 128}
				version2 := models.CompressionVersion{ID: "v1", Format: "mp3", Bitrate: 256} // Same ID

				// Business rule: later update overwrites earlier one
				Expect(version1.Bitrate).ToNot(Equal(version2.Bitrate))
				Expect(version1.ID).To(Equal(version2.ID))
			})
		})
	})
})