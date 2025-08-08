package models

import "time"

// User represents a user in the system (monorepo format)
type User struct {
	ID          string    `json:"id" firestore:"id"`
	Email       string    `json:"email" firestore:"email"`
	DisplayName string    `json:"displayName" firestore:"displayName"`
	ProfilePic  string    `json:"profilePic" firestore:"profilePic"`
	NostrPubkey string    `json:"nostrPubkey" firestore:"nostrPubkey"`
	CreatedAt   time.Time `json:"createdAt" firestore:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt" firestore:"updatedAt"`
}

// Track represents a music track (monorepo format)
type Track struct {
	ID          string    `json:"id" firestore:"id"`
	Title       string    `json:"title" firestore:"title"`
	Artist      string    `json:"artist" firestore:"artist"`
	Album       string    `json:"album,omitempty" firestore:"album"`
	Duration    int       `json:"duration" firestore:"duration"` // in seconds
	AudioURL    string    `json:"audioUrl" firestore:"audioUrl"`
	ArtworkURL  string    `json:"artworkUrl,omitempty" firestore:"artworkUrl"`
	Genre       string    `json:"genre,omitempty" firestore:"genre"`
	PriceMsat   int64     `json:"priceMsat,omitempty" firestore:"priceMsat"`
	OwnerID     string    `json:"ownerId" firestore:"ownerId"`
	NostrEventID string   `json:"nostrEventId,omitempty" firestore:"nostrEventId"`
	CreatedAt   time.Time `json:"createdAt" firestore:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt" firestore:"updatedAt"`
}

// === API Models (from original API) ===

// APIUser represents a user in the API system 
type APIUser struct {
	FirebaseUID   string    `firestore:"firebase_uid"` // Primary key
	CreatedAt     time.Time `firestore:"created_at"`
	UpdatedAt     time.Time `firestore:"updated_at"`
	ActivePubkeys []string  `firestore:"active_pubkeys"` // Denormalized for quick lookup
}

type NostrAuth struct {
	Pubkey      string    `firestore:"pubkey"`       // Primary key
	FirebaseUID string    `firestore:"firebase_uid"` // Foreign key to User
	Active      bool      `firestore:"active"`
	CreatedAt   time.Time `firestore:"created_at"`
	LastUsedAt  time.Time `firestore:"last_used_at"`
	LinkedAt    time.Time `firestore:"linked_at"` // When linked to Firebase user
}

// LinkedPubkeyInfo represents pubkey information in the response
type LinkedPubkeyInfo struct {
	PubKey     string `json:"pubkey"`
	LinkedAt   string `json:"linked_at"`
	LastUsedAt string `json:"last_used_at,omitempty"`
}

// CompressionOption represents a user's choice for audio compression
type CompressionOption struct {
	Bitrate    int    `json:"bitrate"`               // e.g., 128, 256, 320
	Format     string `json:"format"`                // e.g., "mp3", "aac", "ogg"
	Quality    string `json:"quality"`               // e.g., "low", "medium", "high"
	SampleRate int    `json:"sample_rate,omitempty"` // e.g., 44100, 48000
}

// CompressionVersion represents a generated compressed version
type CompressionVersion struct {
	ID         string            `firestore:"id" json:"id"`                   // Unique ID for this version
	URL        string            `firestore:"url" json:"url"`                 // GCS URL
	Bitrate    int               `firestore:"bitrate" json:"bitrate"`         // Actual bitrate
	Format     string            `firestore:"format" json:"format"`           // File format
	Quality    string            `firestore:"quality" json:"quality"`         // Quality level
	SampleRate int               `firestore:"sample_rate" json:"sample_rate"` // Sample rate
	Size       int64             `firestore:"size" json:"size"`               // File size in bytes
	IsPublic   bool              `firestore:"is_public" json:"is_public"`     // Whether to include in Nostr event
	CreatedAt  time.Time         `firestore:"created_at" json:"created_at"`
	Options    CompressionOption `firestore:"options" json:"options"` // Original compression request
}

type NostrTrack struct {
	ID                    string               `firestore:"id" json:"id"`                                                         // UUID
	FirebaseUID           string               `firestore:"firebase_uid" json:"firebase_uid"`                                     // User who uploaded
	Pubkey                string               `firestore:"pubkey" json:"pubkey"`                                                 // Nostr pubkey
	OriginalURL           string               `firestore:"original_url" json:"original_url"`                                     // GCS URL for original file
	PresignedURL          string               `firestore:"-" json:"presigned_url,omitempty"`                                     // Temporary upload URL (not stored)
	Extension             string               `firestore:"extension" json:"extension"`                                           // File extension
	Size                  int64                `firestore:"size,omitempty" json:"size,omitempty"`                                 // Original file size in bytes
	Duration              int                  `firestore:"duration,omitempty" json:"duration,omitempty"`                         // Duration in seconds
	IsProcessing          bool                 `firestore:"is_processing" json:"is_processing"`                                   // Processing status
	CompressionVersions   []CompressionVersion `firestore:"compression_versions,omitempty" json:"compression_versions,omitempty"` // All compressed versions
	HasPendingCompression bool                 `firestore:"has_pending_compression" json:"has_pending_compression"`               // Whether compression is queued
	Deleted               bool                 `firestore:"deleted" json:"deleted"`                                               // Soft delete flag
	NostrKind             int                  `firestore:"nostr_kind,omitempty" json:"nostr_kind,omitempty"`                     // Nostr event kind
	NostrDTag             string               `firestore:"nostr_d_tag,omitempty" json:"nostr_d_tag,omitempty"`                   // Nostr d tag
	CreatedAt             time.Time            `firestore:"created_at" json:"created_at"`
	UpdatedAt             time.Time            `firestore:"updated_at" json:"updated_at"`

	// Deprecated fields - kept for backward compatibility
	CompressedURL string `firestore:"compressed_url,omitempty" json:"compressed_url,omitempty"` // Legacy compressed file
	IsCompressed  bool   `firestore:"is_compressed" json:"is_compressed"`                       // Legacy compression status
}

// VersionUpdate represents a request to update compression version visibility
type VersionUpdate struct {
	VersionID string `json:"version_id"`
	IsPublic  bool   `json:"is_public"`
}

// Legacy PostgreSQL Models
// These models map to the legacy catalog API's PostgreSQL database

type LegacyUser struct {
	ID               string    `db:"id" json:"id"`
	Name             string    `db:"name" json:"name"`
	LightningAddress string    `db:"lightning_address" json:"lightning_address"`
	MSatBalance      int64     `db:"msat_balance" json:"msat_balance"`
	AmpMsat          int       `db:"amp_msat" json:"amp_msat"`
	ArtworkURL       string    `db:"artwork_url" json:"artwork_url"`
	ProfileURL       string    `db:"profile_url" json:"profile_url"`
	IsLocked         bool      `db:"is_locked" json:"is_locked"`
	CreatedAt        time.Time `db:"created_at" json:"created_at"`
	UpdatedAt        time.Time `db:"updated_at" json:"updated_at"`
}

type LegacyTrack struct {
	ID              string    `db:"id" json:"id"`
	ArtistID        string    `db:"artist_id" json:"artist_id"`
	AlbumID         string    `db:"album_id" json:"album_id"`
	Title           string    `db:"title" json:"title"`
	Order           int       `db:"order" json:"order"`
	PlayCount       int       `db:"play_count" json:"play_count"`
	MSatTotal       int64     `db:"msat_total" json:"msat_total"`
	LiveURL         string    `db:"live_url" json:"live_url"`
	RawURL          string    `db:"raw_url" json:"raw_url"`
	Size            int       `db:"size" json:"size"`
	Duration        int       `db:"duration" json:"duration"`
	IsProcessing    bool      `db:"is_processing" json:"is_processing"`
	IsDraft         bool      `db:"is_draft" json:"is_draft"`
	IsExplicit      bool      `db:"is_explicit" json:"is_explicit"`
	CompressorError bool      `db:"compressor_error" json:"compressor_error"`
	Deleted         bool      `db:"deleted" json:"deleted"`
	Lyrics          string    `db:"lyrics" json:"lyrics"`
	CreatedAt       time.Time `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time `db:"updated_at" json:"updated_at"`
	PublishedAt     time.Time `db:"published_at" json:"published_at"`
}

type LegacyArtist struct {
	ID         string    `db:"id" json:"id"`
	UserID     string    `db:"user_id" json:"user_id"`
	Name       string    `db:"name" json:"name"`
	ArtworkURL string    `db:"artwork_url" json:"artwork_url"`
	ArtistURL  string    `db:"artist_url" json:"artist_url"`
	Bio        string    `db:"bio" json:"bio"`
	Twitter    string    `db:"twitter" json:"twitter"`
	Instagram  string    `db:"instagram" json:"instagram"`
	Youtube    string    `db:"youtube" json:"youtube"`
	Website    string    `db:"website" json:"website"`
	Npub       string    `db:"npub" json:"npub"`
	Verified   bool      `db:"verified" json:"verified"`
	Deleted    bool      `db:"deleted" json:"deleted"`
	MSatTotal  int64     `db:"msat_total" json:"msat_total"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
	UpdatedAt  time.Time `db:"updated_at" json:"updated_at"`
}

type LegacyAlbum struct {
	ID              string    `db:"id" json:"id"`
	ArtistID        string    `db:"artist_id" json:"artist_id"`
	Title           string    `db:"title" json:"title"`
	ArtworkURL      string    `db:"artwork_url" json:"artwork_url"`
	Description     string    `db:"description" json:"description"`
	GenreID         int       `db:"genre_id" json:"genre_id"`
	SubgenreID      int       `db:"subgenre_id" json:"subgenre_id"`
	IsDraft         bool      `db:"is_draft" json:"is_draft"`
	IsSingle        bool      `db:"is_single" json:"is_single"`
	Deleted         bool      `db:"deleted" json:"deleted"`
	MSatTotal       int64     `db:"msat_total" json:"msat_total"`
	IsFeedPublished bool      `db:"is_feed_published" json:"is_feed_published"`
	PublishedAt     time.Time `db:"published_at" json:"published_at"`
	CreatedAt       time.Time `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time `db:"updated_at" json:"updated_at"`
}

// === Phase 2 Models ===

// FileUploadToken represents a token for file upload authentication
type FileUploadToken struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	Path      string    `json:"path"`
	UserID    string    `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
}

// FileMetadata represents metadata about a file
type FileMetadata struct {
	Name        string            `json:"name"`
	Size        int64             `json:"size"`
	ContentType string            `json:"content_type"`
	Bucket      string            `json:"bucket"`
	URL         string            `json:"url,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// BucketInfo represents information about a storage bucket
type BucketInfo struct {
	Name         string    `json:"name"`
	Location     string    `json:"location"`
	StorageClass string    `json:"storage_class"`
	CreatedAt    time.Time `json:"created_at"`
}

// SystemInfo represents system diagnostic information
type SystemInfo struct {
	Version     string            `json:"version"`
	Environment string            `json:"environment"`
	Uptime      string            `json:"uptime"`
	Memory      map[string]string `json:"memory"`
	Database    map[string]string `json:"database"`
	Storage     map[string]string `json:"storage"`
	Services    map[string]string `json:"services"`
}

// WebhookPayload represents a webhook payload from Cloud Functions
type WebhookPayload struct {
	Type      string                 `json:"type"`        // "storage", "nostr_relay", "cloud_function"
	Source    string                 `json:"source"`      // Source of the webhook
	EventType string                 `json:"event_type"`  // Type of event
	Data      map[string]interface{} `json:"data"`        // Event data
	Timestamp time.Time              `json:"timestamp"`
	Signature string                 `json:"signature,omitempty"` // HMAC signature for validation
}

// LogEntry represents a log entry for debugging
type LogEntry struct {
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	Service   string                 `json:"service"`
}

// ProcessingStatus represents the status of track processing
type ProcessingStatus struct {
	TrackID     string    `json:"track_id"`
	Status      string    `json:"status"` // "queued", "processing", "completed", "failed"
	Progress    int       `json:"progress"` // 0-100
	Message     string    `json:"message,omitempty"`
	Error       string    `json:"error,omitempty"`
	StartedAt   time.Time `json:"started_at,omitempty"`
	CompletedAt time.Time `json:"completed_at,omitempty"`
}

// AudioMetadata represents metadata extracted from audio files
type AudioMetadata struct {
	Duration    int               `json:"duration"`     // Duration in seconds
	Bitrate     int               `json:"bitrate"`      // Bitrate in kbps
	SampleRate  int               `json:"sample_rate"`  // Sample rate in Hz
	Channels    int               `json:"channels"`     // Number of audio channels
	Format      string            `json:"format"`       // Audio format (mp3, wav, etc.)
	Title       string            `json:"title,omitempty"`
	Artist      string            `json:"artist,omitempty"`
	Album       string            `json:"album,omitempty"`
	Genre       string            `json:"genre,omitempty"`
	Year        int               `json:"year,omitempty"`
	TrackNumber int               `json:"track_number,omitempty"`
	Tags        map[string]string `json:"tags,omitempty"` // Additional metadata tags
}