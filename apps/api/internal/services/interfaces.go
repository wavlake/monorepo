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
	ValidateAudioFile(ctx context.Context, filePath string) error
	ExtractMetadata(ctx context.Context, filePath string) (*models.AudioMetadata, error)
	CompressAudio(ctx context.Context, inputPath, outputPath string, options models.CompressionOption) error
}

// StoragePathConfigInterface defines the interface for storage path operations
type StoragePathConfigInterface interface {
	GetOriginalPath(trackID, extension string) string
	GetCompressedPath(trackID string) string
}

// CompressionServiceInterface defines the interface for compression version management
type CompressionServiceInterface interface {
	RequestCompression(ctx context.Context, trackID string, options []models.CompressionOption) error
	GetCompressionStatus(ctx context.Context, trackID string) (*models.ProcessingStatus, error)
	AddCompressionVersion(ctx context.Context, trackID string, version models.CompressionVersion) error
	UpdateVersionVisibility(ctx context.Context, trackID, versionID string, isPublic bool) error
	GetPublicVersions(ctx context.Context, trackID string) ([]models.CompressionVersion, error)
	DeleteCompressionVersion(ctx context.Context, trackID, versionID string) error
}

// FileServerServiceInterface defines the interface for file server operations
type FileServerServiceInterface interface {
	UploadFile(ctx context.Context, path string, data io.Reader, contentType string) (*models.FileMetadata, error)
	DownloadFile(ctx context.Context, path string) (io.ReadCloser, error)
	DeleteFile(ctx context.Context, path string) error
	ListFiles(ctx context.Context, prefix string) ([]string, error)
	GetFileMetadata(ctx context.Context, path string) (*models.FileMetadata, error)
	GenerateUploadToken(ctx context.Context, path, userID string, expiration time.Duration) (*models.FileUploadToken, error)
}

// MockStorageServiceInterface defines the interface for mock storage operations
type MockStorageServiceInterface interface {
	UploadFile(ctx context.Context, bucket, path string, data io.Reader, contentType string) (*models.FileMetadata, error)
	DownloadFile(ctx context.Context, bucket, path string) (io.ReadCloser, error)
	DeleteFile(ctx context.Context, bucket, path string) error
	ListFiles(ctx context.Context, bucket, prefix string) ([]string, error)
	GetBucketInfo(ctx context.Context, bucket string) (*models.BucketInfo, error)
	CreateBucket(ctx context.Context, bucket, location string) error
	HealthCheck(ctx context.Context) error
}

// DevelopmentServiceInterface defines the interface for development utilities
type DevelopmentServiceInterface interface {
	ResetDatabase(ctx context.Context) error
	SeedTestData(ctx context.Context) error
	GetSystemInfo(ctx context.Context) (*models.SystemInfo, error)
	ClearCache(ctx context.Context) error
	GenerateTestFiles(ctx context.Context, count int) ([]string, error)
	SimulateLoad(ctx context.Context, duration time.Duration) error
	GetLogs(ctx context.Context, level string, limit int) ([]models.LogEntry, error)
}

// TokenServiceInterface defines the interface for token-based authentication
type TokenServiceInterface interface {
	GenerateUploadToken(ctx context.Context, path, userID string, expiration time.Duration) (*models.FileUploadToken, error)
	GenerateDeleteToken(ctx context.Context, path, userID string, expiration time.Duration) (*models.FileUploadToken, error)
	ValidateToken(ctx context.Context, token, path string) (*models.FileUploadToken, error)
	RevokeToken(ctx context.Context, token string) error
	ListActiveTokens(ctx context.Context, userID string) ([]models.FileUploadToken, error)
	RefreshToken(ctx context.Context, token string, expiration time.Duration) (*models.FileUploadToken, error)
}

// WebhookServiceInterface defines the interface for webhook handling
type WebhookServiceInterface interface {
	ProcessCloudFunctionWebhook(ctx context.Context, payload models.WebhookPayload) error
	ProcessStorageWebhook(ctx context.Context, payload models.WebhookPayload) error
	ProcessNostrRelayWebhook(ctx context.Context, payload models.WebhookPayload) error
	GetWebhookStatus(ctx context.Context, webhookID string) (*models.ProcessingStatus, error)
	RetryFailedWebhooks(ctx context.Context, maxRetries int) error
	ValidateWebhookSignature(payload []byte, signature, secret string) error
}

// Ensure services implement their interfaces
var _ UserServiceInterface = (*UserService)(nil)
var _ StorageServiceInterface = (*StorageService)(nil)
var _ PostgresServiceInterface = (*PostgresService)(nil)
var _ NostrTrackServiceInterface = (*NostrTrackService)(nil)
var _ ProcessingServiceInterface = (*ProcessingService)(nil)