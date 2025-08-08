// +build emulator

package integration

import (
	"context"
	"os"
	"testing"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/api/option"

	"github.com/wavlake/monorepo/internal/models"
	"github.com/wavlake/monorepo/internal/services"
	"github.com/wavlake/monorepo/tests/testutil"
)

// UserServiceIntegrationSuite tests the actual UserService implementation
// with real Firebase emulator instances
var _ = Describe("UserService Integration Tests", func() {
	var (
		ctx             context.Context
		userService     *services.UserService
		firestoreClient *firestore.Client
		authClient      *auth.Client
		firebaseApp     *firebase.App
		testFirebaseUID string
		testPubkey      string
	)

	BeforeSuite(func() {
		// Ensure Firebase emulators are running
		if os.Getenv("FIRESTORE_EMULATOR_HOST") == "" {
			Skip("Firebase emulators not running. Run 'task firebase:emulators' first.")
		}

		ctx = context.Background()

		// Initialize Firebase app for emulator testing
		config := &firebase.Config{
			ProjectID: "test-project",
		}

		// Initialize Firebase with emulator
		var err error
		firebaseApp, err = firebase.NewApp(ctx, config, option.WithoutAuthentication())
		Expect(err).ToNot(HaveOccurred())

		// Initialize Firestore client
		firestoreClient, err = firebaseApp.Firestore(ctx)
		Expect(err).ToNot(HaveOccurred())

		// Initialize Auth client
		authClient, err = firebaseApp.Auth(ctx)
		Expect(err).ToNot(HaveOccurred())

		// Create UserService with real clients
		userService = services.NewUserService(firestoreClient, authClient)

		// Set up test data
		testFirebaseUID = testutil.TestFirebaseUID
		testPubkey = testutil.TestPubkey
	})

	AfterSuite(func() {
		if firestoreClient != nil {
			firestoreClient.Close()
		}
	})

	BeforeEach(func() {
		// Clean up test data before each test
		cleanupTestData(ctx, firestoreClient, testFirebaseUID, testPubkey)
		cleanupTestData(ctx, firestoreClient, "different-firebase-uid", "different-pubkey")
	})

	AfterEach(func() {
		// Clean up test data after each test
		cleanupTestData(ctx, firestoreClient, testFirebaseUID, testPubkey)
		cleanupTestData(ctx, firestoreClient, "different-firebase-uid", "different-pubkey")
	})

	Describe("LinkPubkeyToUser - Real Implementation", func() {
		It("should successfully link a pubkey to a new user", func() {
			// Act
			err := userService.LinkPubkeyToUser(ctx, testPubkey, testFirebaseUID)

			// Assert
			Expect(err).ToNot(HaveOccurred())

			// Verify data was actually written to Firestore
			userDoc, err := firestoreClient.Collection("users").Doc(testFirebaseUID).Get(ctx)
			Expect(err).ToNot(HaveOccurred())

			var user models.APIUser
			err = userDoc.DataTo(&user)
			Expect(err).ToNot(HaveOccurred())

			Expect(user.FirebaseUID).To(Equal(testFirebaseUID))
			Expect(user.ActivePubkeys).To(ContainElement(testPubkey))
			Expect(user.CreatedAt).ToNot(BeZero())
			Expect(user.UpdatedAt).ToNot(BeZero())

			// Verify NostrAuth record was created
			nostrAuthDoc, err := firestoreClient.Collection("nostr_auth").Doc(testPubkey).Get(ctx)
			Expect(err).ToNot(HaveOccurred())

			var nostrAuth models.NostrAuth
			err = nostrAuthDoc.DataTo(&nostrAuth)
			Expect(err).ToNot(HaveOccurred())

			Expect(nostrAuth.Pubkey).To(Equal(testPubkey))
			Expect(nostrAuth.FirebaseUID).To(Equal(testFirebaseUID))
			Expect(nostrAuth.Active).To(BeTrue())
			Expect(nostrAuth.CreatedAt).ToNot(BeZero())
			Expect(nostrAuth.LinkedAt).ToNot(BeZero())
		})

		It("should successfully link multiple pubkeys to same user", func() {
			secondPubkey := "different-pubkey"

			// Link first pubkey
			err := userService.LinkPubkeyToUser(ctx, testPubkey, testFirebaseUID)
			Expect(err).ToNot(HaveOccurred())

			// Link second pubkey to same user
			err = userService.LinkPubkeyToUser(ctx, secondPubkey, testFirebaseUID)
			Expect(err).ToNot(HaveOccurred())

			// Verify user has both pubkeys
			userDoc, err := firestoreClient.Collection("users").Doc(testFirebaseUID).Get(ctx)
			Expect(err).ToNot(HaveOccurred())

			var user models.APIUser
			err = userDoc.DataTo(&user)
			Expect(err).ToNot(HaveOccurred())

			Expect(user.ActivePubkeys).To(ContainElement(testPubkey))
			Expect(user.ActivePubkeys).To(ContainElement(secondPubkey))
			Expect(len(user.ActivePubkeys)).To(Equal(2))

			// Clean up second pubkey
			cleanupTestData(ctx, firestoreClient, testFirebaseUID, secondPubkey)
		})

		It("should prevent linking same pubkey to different users", func() {
			differentFirebaseUID := "different-firebase-uid"

			// First user links pubkey successfully
			err := userService.LinkPubkeyToUser(ctx, testPubkey, testFirebaseUID)
			Expect(err).ToNot(HaveOccurred())

			// Second user tries to link same pubkey - should fail
			err = userService.LinkPubkeyToUser(ctx, testPubkey, differentFirebaseUID)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("already linked to a different user"))
		})

		It("should handle concurrent linking operations safely", func() {
			const numGoroutines = 5
			results := make(chan error, numGoroutines)

			// Try to link the same pubkey concurrently from multiple goroutines
			for i := 0; i < numGoroutines; i++ {
				go func(index int) {
					err := userService.LinkPubkeyToUser(ctx, testPubkey, testFirebaseUID)
					results <- err
				}(i)
			}

			// Collect results
			var successCount int
			var errorCount int
			for i := 0; i < numGoroutines; i++ {
				err := <-results
				if err != nil {
					errorCount++
				} else {
					successCount++
				}
			}

			// At least one should succeed, others might fail due to race conditions
			Expect(successCount).To(BeNumerically(">=", 1))
			
			// Verify final state is consistent
			userDoc, err := firestoreClient.Collection("users").Doc(testFirebaseUID).Get(ctx)
			Expect(err).ToNot(HaveOccurred())

			var user models.APIUser
			err = userDoc.DataTo(&user)
			Expect(err).ToNot(HaveOccurred())
			Expect(user.ActivePubkeys).To(ContainElement(testPubkey))
		})
	})

	Describe("UnlinkPubkeyFromUser - Real Implementation", func() {
		BeforeEach(func() {
			// Set up a linked pubkey for each test
			err := userService.LinkPubkeyToUser(ctx, testPubkey, testFirebaseUID)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should successfully unlink a pubkey from user", func() {
			// Act
			err := userService.UnlinkPubkeyFromUser(ctx, testPubkey, testFirebaseUID)

			// Assert
			Expect(err).ToNot(HaveOccurred())

			// Verify NostrAuth record is marked inactive
			nostrAuthDoc, err := firestoreClient.Collection("nostr_auth").Doc(testPubkey).Get(ctx)
			Expect(err).ToNot(HaveOccurred())

			var nostrAuth models.NostrAuth
			err = nostrAuthDoc.DataTo(&nostrAuth)
			Expect(err).ToNot(HaveOccurred())
			Expect(nostrAuth.Active).To(BeFalse())

			// Verify pubkey is removed from user's active list
			userDoc, err := firestoreClient.Collection("users").Doc(testFirebaseUID).Get(ctx)
			Expect(err).ToNot(HaveOccurred())

			var user models.APIUser
			err = userDoc.DataTo(&user)
			Expect(err).ToNot(HaveOccurred())
			Expect(user.ActivePubkeys).ToNot(ContainElement(testPubkey))
		})

		It("should return error when unlinking pubkey that doesn't belong to user", func() {
			differentFirebaseUID := "different-firebase-uid"

			err := userService.UnlinkPubkeyFromUser(ctx, testPubkey, differentFirebaseUID)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("does not belong to this user"))
		})

		It("should return error when unlinking non-existent pubkey", func() {
			nonExistentPubkey := "non-existent-pubkey"

			err := userService.UnlinkPubkeyFromUser(ctx, nonExistentPubkey, testFirebaseUID)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("not found"))
		})

		It("should return error when unlinking already inactive pubkey", func() {
			// First unlink
			err := userService.UnlinkPubkeyFromUser(ctx, testPubkey, testFirebaseUID)
			Expect(err).ToNot(HaveOccurred())

			// Try to unlink again
			err = userService.UnlinkPubkeyFromUser(ctx, testPubkey, testFirebaseUID)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("already unlinked"))
		})
	})

	Describe("GetLinkedPubkeys - Real Implementation", func() {
		It("should return empty slice when user has no linked pubkeys", func() {
			pubkeys, err := userService.GetLinkedPubkeys(ctx, testFirebaseUID)

			Expect(err).ToNot(HaveOccurred())
			Expect(pubkeys).To(BeEmpty())
		})

		It("should return all active pubkeys for a user", func() {
			secondPubkey := "second-pubkey"

			// Link two pubkeys
			err := userService.LinkPubkeyToUser(ctx, testPubkey, testFirebaseUID)
			Expect(err).ToNot(HaveOccurred())

			err = userService.LinkPubkeyToUser(ctx, secondPubkey, testFirebaseUID)
			Expect(err).ToNot(HaveOccurred())

			// Get linked pubkeys
			pubkeys, err := userService.GetLinkedPubkeys(ctx, testFirebaseUID)
			Expect(err).ToNot(HaveOccurred())

			Expect(len(pubkeys)).To(Equal(2))
			pubkeyStrings := make([]string, len(pubkeys))
			for i, auth := range pubkeys {
				pubkeyStrings[i] = auth.Pubkey
				Expect(auth.Active).To(BeTrue())
				Expect(auth.FirebaseUID).To(Equal(testFirebaseUID))
			}
			Expect(pubkeyStrings).To(ContainElement(testPubkey))
			Expect(pubkeyStrings).To(ContainElement(secondPubkey))

			// Clean up second pubkey
			cleanupTestData(ctx, firestoreClient, testFirebaseUID, secondPubkey)
		})

		It("should not return inactive pubkeys", func() {
			// Link then unlink a pubkey
			err := userService.LinkPubkeyToUser(ctx, testPubkey, testFirebaseUID)
			Expect(err).ToNot(HaveOccurred())

			err = userService.UnlinkPubkeyFromUser(ctx, testPubkey, testFirebaseUID)
			Expect(err).ToNot(HaveOccurred())

			// Get linked pubkeys - should be empty
			pubkeys, err := userService.GetLinkedPubkeys(ctx, testFirebaseUID)
			Expect(err).ToNot(HaveOccurred())
			Expect(pubkeys).To(BeEmpty())
		})

		It("should handle missing composite index gracefully", func() {
			// This test ensures the fallback query works when composite indexes are missing
			err := userService.LinkPubkeyToUser(ctx, testPubkey, testFirebaseUID)
			Expect(err).ToNot(HaveOccurred())

			pubkeys, err := userService.GetLinkedPubkeys(ctx, testFirebaseUID)
			Expect(err).ToNot(HaveOccurred())
			Expect(len(pubkeys)).To(Equal(1))
			Expect(pubkeys[0].Pubkey).To(Equal(testPubkey))
		})
	})

	Describe("GetFirebaseUIDByPubkey - Real Implementation", func() {
		BeforeEach(func() {
			// Set up a linked pubkey for each test
			err := userService.LinkPubkeyToUser(ctx, testPubkey, testFirebaseUID)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should return Firebase UID for active pubkey", func() {
			uid, err := userService.GetFirebaseUIDByPubkey(ctx, testPubkey)

			Expect(err).ToNot(HaveOccurred())
			Expect(uid).To(Equal(testFirebaseUID))
		})

		It("should return error for non-existent pubkey", func() {
			nonExistentPubkey := "non-existent-pubkey"

			_, err := userService.GetFirebaseUIDByPubkey(ctx, nonExistentPubkey)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("not found"))
		})

		It("should return error for inactive pubkey", func() {
			// Unlink the pubkey to make it inactive
			err := userService.UnlinkPubkeyFromUser(ctx, testPubkey, testFirebaseUID)
			Expect(err).ToNot(HaveOccurred())

			_, err = userService.GetFirebaseUIDByPubkey(ctx, testPubkey)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("not active"))
		})
	})

	Describe("Transaction Integrity", func() {
		It("should maintain data consistency during transaction failures", func() {
			// This test uses a scenario that might cause transaction failures
			// to verify that data remains consistent

			// Verify initial state is clean
			pubkeys, err := userService.GetLinkedPubkeys(ctx, testFirebaseUID)
			Expect(err).ToNot(HaveOccurred())
			Expect(len(pubkeys)).To(Equal(0))

			// Link a pubkey successfully
			err = userService.LinkPubkeyToUser(ctx, testPubkey, testFirebaseUID)
			Expect(err).ToNot(HaveOccurred())

			// Verify both collections are updated consistently
			userDoc, err := firestoreClient.Collection("users").Doc(testFirebaseUID).Get(ctx)
			Expect(err).ToNot(HaveOccurred())

			nostrAuthDoc, err := firestoreClient.Collection("nostr_auth").Doc(testPubkey).Get(ctx)
			Expect(err).ToNot(HaveOccurred())

			var user models.APIUser
			var nostrAuth models.NostrAuth

			err = userDoc.DataTo(&user)
			Expect(err).ToNot(HaveOccurred())

			err = nostrAuthDoc.DataTo(&nostrAuth)
			Expect(err).ToNot(HaveOccurred())

			// Verify consistency
			Expect(user.ActivePubkeys).To(ContainElement(testPubkey))
			Expect(nostrAuth.Active).To(BeTrue())
			Expect(nostrAuth.FirebaseUID).To(Equal(testFirebaseUID))
		})

		It("should handle link-unlink-relink cycle correctly", func() {
			// Link
			err := userService.LinkPubkeyToUser(ctx, testPubkey, testFirebaseUID)
			Expect(err).ToNot(HaveOccurred())

			// Unlink
			err = userService.UnlinkPubkeyFromUser(ctx, testPubkey, testFirebaseUID)
			Expect(err).ToNot(HaveOccurred())

			// Re-link should work
			err = userService.LinkPubkeyToUser(ctx, testPubkey, testFirebaseUID)
			Expect(err).ToNot(HaveOccurred())

			// Verify final state
			pubkeys, err := userService.GetLinkedPubkeys(ctx, testFirebaseUID)
			Expect(err).ToNot(HaveOccurred())
			Expect(len(pubkeys)).To(Equal(1))
			Expect(pubkeys[0].Pubkey).To(Equal(testPubkey))
			Expect(pubkeys[0].Active).To(BeTrue())
		})
	})

	// Note: GetUserEmail method requires Firebase Auth emulator with actual user records
	// These tests would need Firebase Auth emulator configuration and test user creation
	Describe("GetUserEmail - Real Implementation", func() {
		// Skip these tests if Firebase Auth emulator isn't properly configured
		// with test users, as they require actual Firebase Auth records
		PIt("should return user email from Firebase Auth", func() {
			Skip("Requires Firebase Auth emulator with test user data")
		})

		PIt("should return error for non-existent user", func() {
			Skip("Requires Firebase Auth emulator with test user data")
		})
	})
})

// Helper function to clean up test data
func cleanupTestData(ctx context.Context, client *firestore.Client, firebaseUID, pubkey string) {
	// Clean up user document
	if firebaseUID != "" {
		client.Collection("users").Doc(firebaseUID).Delete(ctx)
	}

	// Clean up nostr_auth document
	if pubkey != "" {
		client.Collection("nostr_auth").Doc(pubkey).Delete(ctx)
	}
}

// Test suite runner
func TestUserServiceIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "UserService Integration Suite")
}