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

// Constants for testing
const (
	TestFirebaseUID = "test-firebase-uid"
	TestPubkey      = "test-pubkey"
	TestPubkey2     = "test-pubkey-2"
	TestEmail       = "test@example.com"
	TestTrackID     = "test-track-123"
)