package services

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/firestore"
	"firebase.google.com/go/v4/auth"
	"github.com/wavlake/monorepo/internal/models"
	"google.golang.org/api/iterator"
)

type UserService struct {
	firestoreClient *firestore.Client
	firebaseAuth    *auth.Client
}

func NewUserService(firestoreClient *firestore.Client, firebaseAuth *auth.Client) *UserService {
	return &UserService{
		firestoreClient: firestoreClient,
		firebaseAuth:    firebaseAuth,
	}
}

// LinkPubkeyToUser links a Nostr pubkey to a Firebase user
func (s *UserService) LinkPubkeyToUser(ctx context.Context, pubkey, firebaseUID string) error {
	now := time.Now()

	// Check if pubkey is already linked to a different user
	existingAuth, err := s.getNostrAuth(ctx, pubkey)
	if err == nil && existingAuth.FirebaseUID != firebaseUID && existingAuth.Active {
		return fmt.Errorf("pubkey is already linked to a different user")
	}

	// Start a transaction
	err = s.firestoreClient.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		// Create or update User record
		userRef := s.firestoreClient.Collection("users").Doc(firebaseUID)
		userDoc, err := tx.Get(userRef)

		var user models.APIUser
		if err != nil {
			// Create new user
			user = models.APIUser{
				FirebaseUID:   firebaseUID,
				CreatedAt:     now,
				UpdatedAt:     now,
				ActivePubkeys: []string{pubkey},
			}
		} else {
			// Update existing user
			if err := userDoc.DataTo(&user); err != nil {
				return fmt.Errorf("failed to parse user data: %w", err)
			}

			// Add pubkey if not already present
			if !contains(user.ActivePubkeys, pubkey) {
				user.ActivePubkeys = append(user.ActivePubkeys, pubkey)
			}
			user.UpdatedAt = now
		}

		if err := tx.Set(userRef, user); err != nil {
			return fmt.Errorf("failed to update user: %w", err)
		}

		// Create or update NostrAuth record
		nostrAuthRef := s.firestoreClient.Collection("nostr_auth").Doc(pubkey)
		nostrAuth := models.NostrAuth{
			Pubkey:      pubkey,
			FirebaseUID: firebaseUID,
			Active:      true,
			CreatedAt:   now,
			LastUsedAt:  now,
			LinkedAt:    now,
		}

		if err := tx.Set(nostrAuthRef, nostrAuth); err != nil {
			return fmt.Errorf("failed to create nostr auth: %w", err)
		}

		return nil
	})

	return err
}

// UnlinkPubkeyFromUser unlinks a pubkey from a Firebase user
func (s *UserService) UnlinkPubkeyFromUser(ctx context.Context, pubkey, firebaseUID string) error {
	// Verify the pubkey belongs to this user
	nostrAuth, err := s.getNostrAuth(ctx, pubkey)
	if err != nil {
		return fmt.Errorf("pubkey not found")
	}

	if nostrAuth.FirebaseUID != firebaseUID {
		return fmt.Errorf("pubkey does not belong to this user")
	}

	if !nostrAuth.Active {
		return fmt.Errorf("pubkey is already unlinked")
	}

	// Start a transaction
	return s.firestoreClient.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		// First, get all documents we need to read
		userRef := s.firestoreClient.Collection("users").Doc(firebaseUID)
		userDoc, err := tx.Get(userRef)
		if err != nil {
			return fmt.Errorf("failed to get user: %w", err)
		}

		var user models.APIUser
		if err := userDoc.DataTo(&user); err != nil {
			return fmt.Errorf("failed to parse user data: %w", err)
		}

		// Now perform all writes
		// Update NostrAuth to inactive
		nostrAuthRef := s.firestoreClient.Collection("nostr_auth").Doc(pubkey)
		updatedNostrAuth := nostrAuth
		updatedNostrAuth.Active = false
		if err := tx.Set(nostrAuthRef, updatedNostrAuth); err != nil {
			return fmt.Errorf("failed to update nostr auth: %w", err)
		}

		// Update User to remove pubkey from active list
		user.ActivePubkeys = removeString(user.ActivePubkeys, pubkey)
		user.UpdatedAt = time.Now()

		if err := tx.Set(userRef, user); err != nil {
			return fmt.Errorf("failed to update user: %w", err)
		}

		return nil
	})
}

// GetLinkedPubkeys returns all active pubkeys for a Firebase user
func (s *UserService) GetLinkedPubkeys(ctx context.Context, firebaseUID string) ([]models.NostrAuth, error) {
	// Try simple query first (without OrderBy) in case indexes are missing
	query := s.firestoreClient.Collection("nostr_auth").
		Where("firebase_uid", "==", firebaseUID).
		Where("active", "==", true)

	// Try with OrderBy first, fall back to simple query if it fails
	orderedQuery := query.OrderBy("linked_at", firestore.Asc)

	iter := orderedQuery.Documents(ctx)
	defer iter.Stop()

	var pubkeys []models.NostrAuth
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			// If the ordered query fails (likely due to missing index), try simple query
			iter.Stop()
			simpleIter := query.Documents(ctx)
			defer simpleIter.Stop()

			for {
				doc, err := simpleIter.Next()
				if err == iterator.Done {
					break
				}
				if err != nil {
					return nil, fmt.Errorf("failed to query pubkeys (both ordered and simple): %w", err)
				}

				var nostrAuth models.NostrAuth
				if err := doc.DataTo(&nostrAuth); err != nil {
					return nil, fmt.Errorf("failed to parse nostr auth: %w", err)
				}

				pubkeys = append(pubkeys, nostrAuth)
			}
			break
		}

		var nostrAuth models.NostrAuth
		if err := doc.DataTo(&nostrAuth); err != nil {
			return nil, fmt.Errorf("failed to parse nostr auth: %w", err)
		}

		pubkeys = append(pubkeys, nostrAuth)
	}

	return pubkeys, nil
}

// GetFirebaseUIDByPubkey returns the Firebase UID for a given pubkey if it's linked and active
func (s *UserService) GetFirebaseUIDByPubkey(ctx context.Context, pubkey string) (string, error) {
	nostrAuth, err := s.getNostrAuth(ctx, pubkey)
	if err != nil {
		return "", fmt.Errorf("pubkey not found: %w", err)
	}

	if !nostrAuth.Active {
		return "", fmt.Errorf("pubkey is not active")
	}

	return nostrAuth.FirebaseUID, nil
}

// getNostrAuth retrieves a NostrAuth record by pubkey
func (s *UserService) getNostrAuth(ctx context.Context, pubkey string) (*models.NostrAuth, error) {
	doc, err := s.firestoreClient.Collection("nostr_auth").Doc(pubkey).Get(ctx)
	if err != nil {
		return nil, err
	}

	var nostrAuth models.NostrAuth
	if err := doc.DataTo(&nostrAuth); err != nil {
		return nil, err
	}

	return &nostrAuth, nil
}

// Helper functions
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func removeString(slice []string, item string) []string {
	var result []string
	for _, s := range slice {
		if s != item {
			result = append(result, s)
		}
	}
	return result
}

// GetUserEmail retrieves the email address for a Firebase user
func (s *UserService) GetUserEmail(ctx context.Context, firebaseUID string) (string, error) {
	user, err := s.firebaseAuth.GetUser(ctx, firebaseUID)
	if err != nil {
		return "", fmt.Errorf("failed to get user from Firebase Auth: %w", err)
	}

	return user.Email, nil
}