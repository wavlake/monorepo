// +build emulator

package integration

import (
	"context"
	"os"
	"testing"

	firebase "firebase.google.com/go/v4"
	"google.golang.org/api/option"

	"github.com/wavlake/monorepo/internal/models"
	"github.com/wavlake/monorepo/internal/services"
	"github.com/wavlake/monorepo/tests/testutil"
)

// TestUserServiceWithFirebaseEmulators tests the actual UserService implementation
// with real Firebase emulator instances (separate file to avoid BeforeSuite conflicts)
func TestUserServiceWithFirebaseEmulators(t *testing.T) {
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

	// Initialize Auth client
	authClient, err := firebaseApp.Auth(ctx)
	if err != nil {
		t.Fatalf("Failed to initialize Auth client: %v", err)
	}

	// Create UserService with real clients
	userService := services.NewUserService(firestoreClient, authClient)

	// Set up test data
	testFirebaseUID := testutil.TestFirebaseUID
	testPubkey := testutil.TestPubkey

	// Clean up function
	cleanup := func() {
		firestoreClient.Collection("users").Doc(testFirebaseUID).Delete(ctx)
		firestoreClient.Collection("nostr_auth").Doc(testPubkey).Delete(ctx)
	}

	// Clean up before and after
	cleanup()
	defer cleanup()

	t.Run("LinkPubkeyToUser_RealImplementation", func(t *testing.T) {
		// Act
		err := userService.LinkPubkeyToUser(ctx, testPubkey, testFirebaseUID)
		if err != nil {
			t.Fatalf("LinkPubkeyToUser failed: %v", err)
		}

		// Verify data was actually written to Firestore
		userDoc, err := firestoreClient.Collection("users").Doc(testFirebaseUID).Get(ctx)
		if err != nil {
			t.Fatalf("Failed to get user document: %v", err)
		}

		var user models.APIUser
		err = userDoc.DataTo(&user)
		if err != nil {
			t.Fatalf("Failed to parse user data: %v", err)
		}

		// Verify user data
		if user.FirebaseUID != testFirebaseUID {
			t.Errorf("Expected FirebaseUID %s, got %s", testFirebaseUID, user.FirebaseUID)
		}

		found := false
		for _, pubkey := range user.ActivePubkeys {
			if pubkey == testPubkey {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected pubkey %s to be in ActivePubkeys %v", testPubkey, user.ActivePubkeys)
		}

		// Verify NostrAuth record was created
		nostrAuthDoc, err := firestoreClient.Collection("nostr_auth").Doc(testPubkey).Get(ctx)
		if err != nil {
			t.Fatalf("Failed to get nostr_auth document: %v", err)
		}

		var nostrAuth models.NostrAuth
		err = nostrAuthDoc.DataTo(&nostrAuth)
		if err != nil {
			t.Fatalf("Failed to parse nostr_auth data: %v", err)
		}

		if nostrAuth.Pubkey != testPubkey {
			t.Errorf("Expected pubkey %s, got %s", testPubkey, nostrAuth.Pubkey)
		}
		if nostrAuth.FirebaseUID != testFirebaseUID {
			t.Errorf("Expected FirebaseUID %s, got %s", testFirebaseUID, nostrAuth.FirebaseUID)
		}
		if !nostrAuth.Active {
			t.Error("Expected nostr_auth to be active")
		}
	})

	t.Run("GetLinkedPubkeys_RealImplementation", func(t *testing.T) {
		// Setup: Link a pubkey first
		err := userService.LinkPubkeyToUser(ctx, testPubkey, testFirebaseUID)
		if err != nil {
			t.Fatalf("Setup failed: %v", err)
		}

		// Act
		pubkeys, err := userService.GetLinkedPubkeys(ctx, testFirebaseUID)
		if err != nil {
			t.Fatalf("GetLinkedPubkeys failed: %v", err)
		}

		// Assert
		if len(pubkeys) != 1 {
			t.Errorf("Expected 1 pubkey, got %d", len(pubkeys))
		}

		if len(pubkeys) > 0 {
			if pubkeys[0].Pubkey != testPubkey {
				t.Errorf("Expected pubkey %s, got %s", testPubkey, pubkeys[0].Pubkey)
			}
			if pubkeys[0].FirebaseUID != testFirebaseUID {
				t.Errorf("Expected FirebaseUID %s, got %s", testFirebaseUID, pubkeys[0].FirebaseUID)
			}
			if !pubkeys[0].Active {
				t.Error("Expected pubkey to be active")
			}
		}
	})

	t.Run("UnlinkPubkeyFromUser_RealImplementation", func(t *testing.T) {
		// Setup: Link a pubkey first
		err := userService.LinkPubkeyToUser(ctx, testPubkey, testFirebaseUID)
		if err != nil {
			t.Fatalf("Setup failed: %v", err)
		}

		// Act
		err = userService.UnlinkPubkeyFromUser(ctx, testPubkey, testFirebaseUID)
		if err != nil {
			t.Fatalf("UnlinkPubkeyFromUser failed: %v", err)
		}

		// Verify NostrAuth record is marked inactive
		nostrAuthDoc, err := firestoreClient.Collection("nostr_auth").Doc(testPubkey).Get(ctx)
		if err != nil {
			t.Fatalf("Failed to get nostr_auth document: %v", err)
		}

		var nostrAuth models.NostrAuth
		err = nostrAuthDoc.DataTo(&nostrAuth)
		if err != nil {
			t.Fatalf("Failed to parse nostr_auth data: %v", err)
		}

		if nostrAuth.Active {
			t.Error("Expected nostr_auth to be inactive after unlinking")
		}

		// Verify pubkey is removed from user's active list
		userDoc, err := firestoreClient.Collection("users").Doc(testFirebaseUID).Get(ctx)
		if err != nil {
			t.Fatalf("Failed to get user document: %v", err)
		}

		var user models.APIUser
		err = userDoc.DataTo(&user)
		if err != nil {
			t.Fatalf("Failed to parse user data: %v", err)
		}

		for _, pubkey := range user.ActivePubkeys {
			if pubkey == testPubkey {
				t.Errorf("Expected pubkey %s to be removed from ActivePubkeys", testPubkey)
			}
		}
	})

	t.Run("GetFirebaseUIDByPubkey_RealImplementation", func(t *testing.T) {
		// Setup: Link a pubkey first
		err := userService.LinkPubkeyToUser(ctx, testPubkey, testFirebaseUID)
		if err != nil {
			t.Fatalf("Setup failed: %v", err)
		}

		// Act
		uid, err := userService.GetFirebaseUIDByPubkey(ctx, testPubkey)
		if err != nil {
			t.Fatalf("GetFirebaseUIDByPubkey failed: %v", err)
		}

		// Assert
		if uid != testFirebaseUID {
			t.Errorf("Expected Firebase UID %s, got %s", testFirebaseUID, uid)
		}
	})
}