package testutil

import (
	"time"

	"github.com/wavlake/monorepo/internal/models"
)

// User fixtures for testing

// ValidAPIUser returns a valid APIUser for testing
func ValidAPIUser() models.APIUser {
	return models.APIUser{
		FirebaseUID:   "test-firebase-uid",
		ActivePubkeys: []string{"test-pubkey-1", "test-pubkey-2"},
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
}

// ValidNostrAuth returns a valid NostrAuth for testing
func ValidNostrAuth() models.NostrAuth {
	return models.NostrAuth{
		Pubkey:      "test-pubkey",
		FirebaseUID: "test-firebase-uid",
		Active:      true,
		LinkedAt:    time.Now(),
		LastUsedAt:  time.Now(),
		CreatedAt:   time.Now(),
	}
}

// InactiveNostrAuth returns an inactive NostrAuth for testing
func InactiveNostrAuth() models.NostrAuth {
	auth := ValidNostrAuth()
	auth.Active = false
	return auth
}

// ValidLinkedPubkeyInfo returns a valid LinkedPubkeyInfo for testing
func ValidLinkedPubkeyInfo() models.LinkedPubkeyInfo {
	return models.LinkedPubkeyInfo{
		PubKey:     "test-pubkey",
		LinkedAt:   time.Now().Format(time.RFC3339),
		LastUsedAt: time.Now().Format(time.RFC3339),
	}
}

// HTTP request fixtures

// ValidLinkPubkeyRequest returns a valid link pubkey request body
func ValidLinkPubkeyRequest() map[string]interface{} {
	return map[string]interface{}{
		"pubkey": "test-pubkey",
	}
}

// ValidUnlinkPubkeyRequest returns a valid unlink pubkey request body
func ValidUnlinkPubkeyRequest() map[string]interface{} {
	return map[string]interface{}{
		"pubkey": "test-pubkey",
	}
}

// ValidCheckPubkeyLinkRequest returns a valid check pubkey link request body
func ValidCheckPubkeyLinkRequest() map[string]interface{} {
	return map[string]interface{}{
		"pubkey": "test-pubkey",
	}
}

// Track-related fixtures

// ValidNostrTrack returns a valid NostrTrack for testing
func ValidNostrTrack() *models.NostrTrack {
	return &models.NostrTrack{
		ID:                    "test-track-123",
		FirebaseUID:           TestFirebaseUID,
		Pubkey:               TestPubkey,
		Extension:            "mp3",
		OriginalURL:          "https://storage.googleapis.com/test-bucket/uploads/test-track-123.mp3",
		PresignedURL:         "",
		Size:                 0,
		Duration:             0,
		IsProcessing:         false,
		CompressionVersions:  nil,
		HasPendingCompression: false,
		Deleted:              false,
		NostrKind:            0,
		NostrDTag:            "",
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
		CompressedURL:        "",
		IsCompressed:         false,
	}
}

// ValidCreateTrackRequest returns a valid create track request body
func ValidCreateTrackRequest() map[string]interface{} {
	return map[string]interface{}{
		"extension": "mp3",
	}
}

// ValidTracksList returns a list of valid NostrTracks for testing
func ValidTracksList() []*models.NostrTrack {
	track1 := ValidNostrTrack()
	track2 := ValidNostrTrack()
	track2.ID = "test-track-456"
	track2.Extension = "wav"
	
	return []*models.NostrTrack{track1, track2}
}

// ValidCompressionVersion returns a valid CompressionVersion for testing
func ValidCompressionVersion() models.CompressionVersion {
	return models.CompressionVersion{
		ID:         "v1-128k",
		URL:        "https://storage.googleapis.com/test-bucket/compressed/test-track-123/128k.mp3",
		Bitrate:    128,
		Format:     "mp3",
		Quality:    "standard",
		SampleRate: 44100,
		Size:       2048000,
		IsPublic:   true,
		CreatedAt:  time.Now(),
		Options:    models.CompressionOption{
			Bitrate:    128,
			Format:     "mp3",
			SampleRate: 44100,
		},
	}
}

// ValidVersionUpdate returns a valid VersionUpdate for testing
func ValidVersionUpdate() models.VersionUpdate {
	return models.VersionUpdate{
		VersionID: "v1-128k",
		IsPublic:  false,
	}
}

// Legacy model fixtures

// ValidLegacyUser returns a valid LegacyUser for testing
func ValidLegacyUser() models.LegacyUser {
	return models.LegacyUser{
		ID:               TestFirebaseUID,
		Name:             "Test User",
		LightningAddress: "test@wavlake.com",
		MSatBalance:      1000000,
		AmpMsat:          1000,
		ArtworkURL:       "https://example.com/avatar.jpg",
		ProfileURL:       "https://example.com/profile",
		IsLocked:         false,
		CreatedAt:        time.Now().Add(-24 * time.Hour),
		UpdatedAt:        time.Now(),
	}
}

// ValidLegacyTrack returns a valid LegacyTrack for testing
func ValidLegacyTrack() models.LegacyTrack {
	return models.LegacyTrack{
		ID:              "track-123",
		ArtistID:        "artist-123",
		AlbumID:         "album-123",
		Title:           "Test Track",
		Order:           1,
		PlayCount:       100,
		MSatTotal:       50000,
		LiveURL:         "https://example.com/track.mp3",
		RawURL:          "https://example.com/raw.wav",
		Size:            2048000,
		Duration:        180,
		IsProcessing:    false,
		IsDraft:         false,
		IsExplicit:      false,
		CompressorError: false,
		Deleted:         false,
		Lyrics:          "Test lyrics",
		CreatedAt:       time.Now().Add(-24 * time.Hour),
		UpdatedAt:       time.Now(),
		PublishedAt:     time.Now().Add(-12 * time.Hour),
	}
}

// ValidLegacyArtist returns a valid LegacyArtist for testing
func ValidLegacyArtist() models.LegacyArtist {
	return models.LegacyArtist{
		ID:         "artist-123",
		UserID:     TestFirebaseUID,
		Name:       "Test Artist",
		ArtworkURL: "https://example.com/artist.jpg",
		ArtistURL:  "test-artist",
		Bio:        "Test artist bio",
		Twitter:    "@testartist",
		Instagram:  "@testartist",
		Youtube:    "testartist",
		Website:    "https://testartist.com",
		Npub:       "npub1test...",
		Verified:   false,
		Deleted:    false,
		MSatTotal:  75000,
		CreatedAt:  time.Now().Add(-30 * 24 * time.Hour),
		UpdatedAt:  time.Now(),
	}
}

// ValidLegacyAlbum returns a valid LegacyAlbum for testing
func ValidLegacyAlbum() models.LegacyAlbum {
	return models.LegacyAlbum{
		ID:              "album-123",
		ArtistID:        "artist-123",
		Title:           "Test Album",
		ArtworkURL:      "https://example.com/album.jpg",
		Description:     "Test album description",
		GenreID:         1,
		SubgenreID:      2,
		IsDraft:         false,
		IsSingle:        false,
		Deleted:         false,
		MSatTotal:       150000,
		IsFeedPublished: true,
		PublishedAt:     time.Now().Add(-7 * 24 * time.Hour),
		CreatedAt:       time.Now().Add(-30 * 24 * time.Hour),
		UpdatedAt:       time.Now(),
	}
}

// ValidLegacyTracksList returns a list of valid LegacyTracks for testing
func ValidLegacyTracksList() []models.LegacyTrack {
	track1 := ValidLegacyTrack()
	track2 := ValidLegacyTrack()
	track2.ID = "track-456"
	track2.Title = "Test Track 2"
	track2.Order = 2
	
	return []models.LegacyTrack{track1, track2}
}

// ValidLegacyArtistsList returns a list of valid LegacyArtists for testing
func ValidLegacyArtistsList() []models.LegacyArtist {
	artist1 := ValidLegacyArtist()
	artist2 := ValidLegacyArtist()
	artist2.ID = "artist-456"
	artist2.Name = "Test Artist 2"
	artist2.ArtistURL = "test-artist-2"
	
	return []models.LegacyArtist{artist1, artist2}
}

// ValidLegacyAlbumsList returns a list of valid LegacyAlbums for testing
func ValidLegacyAlbumsList() []models.LegacyAlbum {
	album1 := ValidLegacyAlbum()
	album2 := ValidLegacyAlbum()
	album2.ID = "album-456"
	album2.Title = "Test Album 2"
	
	return []models.LegacyAlbum{album1, album2}
}

// Constants for testing
const (
	TestFirebaseUID = "test-firebase-uid"
	TestPubkey      = "test-pubkey"
	TestPubkey2     = "test-pubkey-2"
	TestEmail       = "test@example.com"
	TestTrackID     = "test-track-123"
	TestExtension   = "mp3"
	TestArtistID    = "artist-123"
	TestAlbumID     = "album-123"
)

// Additional fixtures for comprehensive testing

// ValidUser returns a valid User for testing
func ValidUser() models.User {
	return models.User{
		ID:          "user-123",
		Email:       TestEmail,
		DisplayName: "Test User",
		ProfilePic:  "https://example.com/profile.jpg",
		NostrPubkey: TestPubkey,
		CreatedAt:   time.Now().Add(-24 * time.Hour),
		UpdatedAt:   time.Now(),
	}
}

// ValidTrack returns a valid Track for testing
func ValidTrack() models.Track {
	return models.Track{
		ID:           TestTrackID,
		Title:        "Test Track",
		Artist:       "Test Artist",
		Album:        "Test Album",
		Duration:     180,
		AudioURL:     "https://storage.googleapis.com/test-bucket/track.mp3",
		ArtworkURL:   "https://example.com/artwork.jpg",
		Genre:        "Rock",
		PriceMsat:    1000,
		OwnerID:      "user-123",
		NostrEventID: "nostr-event-123",
		CreatedAt:    time.Now().Add(-24 * time.Hour),
		UpdatedAt:    time.Now(),
	}
}

// Processing and File fixtures

// ValidFileUploadToken returns a valid FileUploadToken for testing
func ValidFileUploadToken() models.FileUploadToken {
	return models.FileUploadToken{
		Token:     "upload-token-abc123",
		ExpiresAt: time.Now().Add(1 * time.Hour),
		Path:      "/uploads/user-123/track.mp3",
		UserID:    "user-123",
		CreatedAt: time.Now(),
	}
}

// ValidFileMetadata returns a valid FileMetadata for testing
func ValidFileMetadata() models.FileMetadata {
	return models.FileMetadata{
		Name:        "track.mp3",
		Size:        5242880,
		ContentType: "audio/mpeg",
		Bucket:      "test-bucket",
		URL:         "https://storage.googleapis.com/test-bucket/track.mp3",
		Metadata: map[string]string{
			"user_id":  "user-123",
			"track_id": TestTrackID,
		},
		CreatedAt: time.Now().Add(-1 * time.Hour),
		UpdatedAt: time.Now(),
	}
}

// ValidProcessingStatus returns a valid ProcessingStatus for testing
func ValidProcessingStatus() models.ProcessingStatus {
	return models.ProcessingStatus{
		TrackID:     TestTrackID,
		Status:      "completed",
		Progress:    100,
		Message:     "Processing completed successfully",
		StartedAt:   time.Now().Add(-10 * time.Minute),
		CompletedAt: time.Now().Add(-2 * time.Minute),
	}
}

// ValidAudioMetadata returns a valid AudioMetadata for testing
func ValidAudioMetadata() models.AudioMetadata {
	return models.AudioMetadata{
		Duration:    180,
		Bitrate:     320,
		SampleRate:  44100,
		Channels:    2,
		Format:      "mp3",
		Title:       "Test Track",
		Artist:      "Test Artist",
		Album:       "Test Album",
		Genre:       "Rock",
		Year:        2024,
		TrackNumber: 1,
		Tags: map[string]string{
			"encoder": "LAME",
			"comment": "Test track",
		},
	}
}

// ValidSystemInfo returns a valid SystemInfo for testing
func ValidSystemInfo() models.SystemInfo {
	return models.SystemInfo{
		Version:     "1.0.0-test",
		Environment: "test",
		Uptime:      "1h 30m",
		Memory: map[string]string{
			"total": "8GB",
			"used":  "2GB",
		},
		Database: map[string]string{
			"status": "healthy",
		},
		Storage: map[string]string{
			"status": "healthy",
			"bucket": "test-bucket",
		},
		Services: map[string]string{
			"firebase": "connected",
		},
	}
}

// ValidWebhookPayload returns a valid WebhookPayload for testing
func ValidWebhookPayload() models.WebhookPayload {
	return models.WebhookPayload{
		Type:      "storage",
		Source:    "cloud_storage",
		EventType: "object_uploaded",
		Data: map[string]interface{}{
			"bucket":      "test-bucket",
			"object_name": "uploads/track.mp3",
		},
		Timestamp: time.Now(),
		Signature: "test-signature",
	}
}

// Error case fixtures

// ExpiredFileUploadToken returns an expired FileUploadToken for testing
func ExpiredFileUploadToken() models.FileUploadToken {
	token := ValidFileUploadToken()
	token.ExpiresAt = time.Now().Add(-1 * time.Hour)
	return token
}

// FailedProcessingStatus returns a failed ProcessingStatus for testing
func FailedProcessingStatus() models.ProcessingStatus {
	return models.ProcessingStatus{
		TrackID:   TestTrackID,
		Status:    "failed",
		Progress:  50,
		Message:   "Processing failed",
		Error:     "invalid audio format",
		StartedAt: time.Now().Add(-10 * time.Minute),
	}
}