package setup

import (
	"context"
	"fmt"
	"log"
	"time"

	"cloud.google.com/go/firestore"
	"firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"google.golang.org/api/option"
)

// TestDatabase provides test database setup and teardown utilities
type TestDatabase struct {
	App       *firebase.App
	Firestore *firestore.Client
	Auth      *auth.Client
	ctx       context.Context
}

// SetupTestDB initializes a test database connection
func SetupTestDB() (*TestDatabase, error) {
	ctx := context.Background()
	
	// For testing, we can use the Firebase emulator or a test project
	// Check if we're running against emulator
	if isEmulatorMode() {
		return setupEmulator(ctx)
	}
	
	// Otherwise, use test service account
	return setupTestProject(ctx)
}

// TeardownTestDB cleans up test database resources
func (tdb *TestDatabase) TeardownTestDB() error {
	if tdb.Firestore != nil {
		return tdb.Firestore.Close()
	}
	return nil
}

// CleanCollection removes all documents from a collection
func (tdb *TestDatabase) CleanCollection(collectionName string) error {
	collection := tdb.Firestore.Collection(collectionName)
	
	// Get all documents
	docs, err := collection.Documents(tdb.ctx).GetAll()
	if err != nil {
		return fmt.Errorf("failed to get documents: %v", err)
	}
	
	// Delete in batches
	batch := tdb.Firestore.Batch()
	batchSize := 0
	
	for _, doc := range docs {
		batch.Delete(doc.Ref)
		batchSize++
		
		// Commit batch when it reaches 500 (Firestore limit)
		if batchSize >= 500 {
			if _, err := batch.Commit(tdb.ctx); err != nil {
				return fmt.Errorf("failed to commit batch: %v", err)
			}
			batch = tdb.Firestore.Batch()
			batchSize = 0
		}
	}
	
	// Commit remaining documents
	if batchSize > 0 {
		if _, err := batch.Commit(tdb.ctx); err != nil {
			return fmt.Errorf("failed to commit final batch: %v", err)
		}
	}
	
	return nil
}

// SeedTestData creates initial test data
func (tdb *TestDatabase) SeedTestData() error {
	// Create test user
	testUser := map[string]interface{}{
		"id":          "test-user-1",
		"email":       "test@wavlake.com",
		"displayName": "Test User",
		"profilePic":  "",
		"nostrPubkey": "test-pubkey-123",
		"createdAt":   time.Now(),
		"updatedAt":   time.Now(),
	}
	
	if _, err := tdb.Firestore.Collection("users").Doc("test-user-1").Set(tdb.ctx, testUser); err != nil {
		return fmt.Errorf("failed to create test user: %v", err)
	}
	
	// Create test track
	testTrack := map[string]interface{}{
		"id":           "test-track-1",
		"title":        "Test Track",
		"artist":       "Test Artist",
		"album":        "Test Album",
		"duration":     180,
		"audioUrl":     "https://example.com/test-track.mp3",
		"artworkUrl":   "https://example.com/test-artwork.jpg",
		"genre":        "Electronic",
		"priceMsat":    1000,
		"ownerId":      "test-user-1",
		"nostrEventId": "",
		"createdAt":    time.Now(),
		"updatedAt":    time.Now(),
	}
	
	if _, err := tdb.Firestore.Collection("tracks").Doc("test-track-1").Set(tdb.ctx, testTrack); err != nil {
		return fmt.Errorf("failed to create test track: %v", err)
	}
	
	return nil
}

// ResetTestDB cleans all collections and reseeds test data
func (tdb *TestDatabase) ResetTestDB() error {
	collections := []string{"users", "tracks", "albums", "playlists", "payments"}
	
	for _, collection := range collections {
		if err := tdb.CleanCollection(collection); err != nil {
			return fmt.Errorf("failed to clean collection %s: %v", collection, err)
		}
	}
	
	return tdb.SeedTestData()
}

// setupEmulator configures connection to Firebase emulator
func setupEmulator(ctx context.Context) (*TestDatabase, error) {
	conf := &firebase.Config{
		ProjectID: "wavlake-test",
	}
	
	app, err := firebase.NewApp(ctx, conf)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Firebase app: %v", err)
	}
	
	firestore, err := app.Firestore(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Firestore client: %v", err)
	}
	
	auth, err := app.Auth(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Auth client: %v", err)
	}
	
	return &TestDatabase{
		App:       app,
		Firestore: firestore,
		Auth:      auth,
		ctx:       ctx,
	}, nil
}

// setupTestProject configures connection to test Firebase project
func setupTestProject(ctx context.Context) (*TestDatabase, error) {
	// Use service account key for test project
	opt := option.WithCredentialsFile("../../config/test-service-account.json")
	
	conf := &firebase.Config{
		ProjectID: "wavlake-test-project",
	}
	
	app, err := firebase.NewApp(ctx, conf, opt)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Firebase app: %v", err)
	}
	
	firestore, err := app.Firestore(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Firestore client: %v", err)
	}
	
	auth, err := app.Auth(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Auth client: %v", err)
	}
	
	return &TestDatabase{
		App:       app,
		Firestore: firestore,
		Auth:      auth,
		ctx:       ctx,
	}, nil
}

// isEmulatorMode checks if we're running against Firebase emulators
func isEmulatorMode() bool {
	// Check common emulator environment variables
	return len([]string{
		// Add emulator detection logic here
		// For now, default to emulator mode for safety
	}) > 0 || true // Default to emulator for safety
}

// CreateTestUser creates a test user with Firebase Auth
func (tdb *TestDatabase) CreateTestUser(email, displayName string) (*auth.UserRecord, error) {
	params := (&auth.UserToCreate{}).
		Email(email).
		DisplayName(displayName).
		EmailVerified(true)
		
	user, err := tdb.Auth.CreateUser(tdb.ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to create test user: %v", err)
	}
	
	return user, nil
}

// DeleteTestUser removes a test user from Firebase Auth
func (tdb *TestDatabase) DeleteTestUser(uid string) error {
	if err := tdb.Auth.DeleteUser(tdb.ctx, uid); err != nil {
		return fmt.Errorf("failed to delete test user: %v", err)
	}
	return nil
}

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}