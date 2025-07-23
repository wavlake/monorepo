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

// Constants for testing
const (
	TestFirebaseUID = "test-firebase-uid"
	TestPubkey      = "test-pubkey"
	TestPubkey2     = "test-pubkey-2"
	TestEmail       = "test@example.com"
)