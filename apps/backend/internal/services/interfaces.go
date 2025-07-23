package services

//go:generate mockgen -source=interfaces.go -destination=../../tests/mocks/service_mocks.go -package=mocks

import (
	"context"
	"io"
	"time"

	"github.com/wavlake/monorepo/internal/models"
)

// UserServiceInterface defines the interface for user operations
type UserServiceInterface interface {
	LinkPubkeyToUser(ctx context.Context, pubkey, firebaseUID string) error
	UnlinkPubkeyFromUser(ctx context.Context, pubkey, firebaseUID string) error
	GetLinkedPubkeys(ctx context.Context, firebaseUID string) ([]models.NostrAuth, error)
	GetFirebaseUIDByPubkey(ctx context.Context, pubkey string) (string, error)
	GetUserEmail(ctx context.Context, firebaseUID string) (string, error)
}

// PostgresServiceInterface defines the interface for PostgreSQL operations
type PostgresServiceInterface interface {
	GetUserByFirebaseUID(ctx context.Context, firebaseUID string) (*models.LegacyUser, error)
	GetUserTracks(ctx context.Context, firebaseUID string) ([]models.LegacyTrack, error)
	GetUserArtists(ctx context.Context, firebaseUID string) ([]models.LegacyArtist, error)
	GetUserAlbums(ctx context.Context, firebaseUID string) ([]models.LegacyAlbum, error)
	GetTracksByArtist(ctx context.Context, artistID string) ([]models.LegacyTrack, error)
	GetTracksByAlbum(ctx context.Context, albumID string) ([]models.LegacyTrack, error)
}

// StorageServiceInterface defines the interface for storage operations
type StorageServiceInterface interface {
	GeneratePresignedURL(ctx context.Context, objectName string, expiration time.Duration) (string, error)
	GetPublicURL(objectName string) string
	UploadObject(ctx context.Context, objectName string, data io.Reader, contentType string) error
	CopyObject(ctx context.Context, srcObject, dstObject string) error
	DeleteObject(ctx context.Context, objectName string) error
	GetObjectMetadata(ctx context.Context, objectName string) (interface{}, error)
	GetObjectReader(ctx context.Context, objectName string) (io.ReadCloser, error)
	GetBucketName() string
	Close() error
}

// NostrTrackServiceInterface defines the interface for Nostr track operations
type NostrTrackServiceInterface interface {
	CreateTrack(ctx context.Context, pubkey, firebaseUID, extension string) (*models.NostrTrack, error)
	GetTrack(ctx context.Context, trackID string) (*models.NostrTrack, error)
	GetTracksByPubkey(ctx context.Context, pubkey string) ([]*models.NostrTrack, error)
	GetTracksByFirebaseUID(ctx context.Context, firebaseUID string) ([]*models.NostrTrack, error)
	UpdateTrack(ctx context.Context, trackID string, updates map[string]interface{}) error
	MarkTrackAsProcessed(ctx context.Context, trackID string, size int64, duration int) error
	MarkTrackAsCompressed(ctx context.Context, trackID, compressedURL string) error
	DeleteTrack(ctx context.Context, trackID string) error
	HardDeleteTrack(ctx context.Context, trackID string) error
	UpdateCompressionVisibility(ctx context.Context, trackID string, updates []models.VersionUpdate) error
	AddCompressionVersion(ctx context.Context, trackID string, version models.CompressionVersion) error
	SetPendingCompression(ctx context.Context, trackID string, pending bool) error
}

// ProcessingServiceInterface defines the interface for track processing operations
type ProcessingServiceInterface interface {
	ProcessTrack(ctx context.Context, trackID string) error
	ProcessTrackAsync(ctx context.Context, trackID string)
	RequestCompressionVersions(ctx context.Context, trackID string, compressionOptions []models.CompressionOption) error
	ProcessCompressionAsync(ctx context.Context, trackID string, option models.CompressionOption)
	ProcessCompression(ctx context.Context, trackID string, option models.CompressionOption) error
}

// AudioProcessorInterface defines the interface for audio processing operations
type AudioProcessorInterface interface {
	IsFormatSupported(extension string) bool
}

// Ensure services implement their interfaces
var _ UserServiceInterface = (*UserService)(nil)
var _ StorageServiceInterface = (*StorageService)(nil)
var _ PostgresServiceInterface = (*PostgresService)(nil)
var _ NostrTrackServiceInterface = (*NostrTrackService)(nil)
var _ ProcessingServiceInterface = (*ProcessingService)(nil)