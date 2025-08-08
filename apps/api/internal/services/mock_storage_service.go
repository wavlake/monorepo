package services

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/wavlake/monorepo/internal/models"
)

// MockStorageService provides GCS-compatible mock storage for development
type MockStorageService struct {
	basePath string
}

// NewMockStorageService creates a new mock storage service
func NewMockStorageService(basePath string) *MockStorageService {
	return &MockStorageService{
		basePath: basePath,
	}
}

// UploadFile uploads a file to mock storage
func (s *MockStorageService) UploadFile(ctx context.Context, bucket, path string, data io.Reader, contentType string) (*models.FileMetadata, error) {
	fullPath := filepath.Join(s.basePath, bucket, path)
	
	// Create directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	// Create the file
	file, err := os.Create(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Copy data to file
	size, err := io.Copy(file, data)
	if err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	return &models.FileMetadata{
		Name:        filepath.Base(path),
		Size:        size,
		ContentType: contentType,
		Bucket:      bucket,
		URL:         fmt.Sprintf("/mock-storage/%s/%s", bucket, path),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}, nil
}

// DownloadFile downloads a file from mock storage
func (s *MockStorageService) DownloadFile(ctx context.Context, bucket, path string) (io.ReadCloser, error) {
	fullPath := filepath.Join(s.basePath, bucket, path)
	
	file, err := os.Open(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	
	return file, nil
}

// DeleteFile deletes a file from mock storage
func (s *MockStorageService) DeleteFile(ctx context.Context, bucket, path string) error {
	fullPath := filepath.Join(s.basePath, bucket, path)
	
	if err := os.Remove(fullPath); err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	
	return nil
}

// ListFiles lists files in mock storage with the given prefix
func (s *MockStorageService) ListFiles(ctx context.Context, bucket, prefix string) ([]string, error) {
	searchPath := filepath.Join(s.basePath, bucket, prefix)
	
	var files []string
	err := filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			// Get relative path from bucket root
			bucketRoot := filepath.Join(s.basePath, bucket)
			relPath, err := filepath.Rel(bucketRoot, path)
			if err != nil {
				return err
			}
			files = append(files, relPath)
		}
		return nil
	})
	
	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}
	
	return files, nil
}

// GetBucketInfo returns information about a mock storage bucket
func (s *MockStorageService) GetBucketInfo(ctx context.Context, bucket string) (*models.BucketInfo, error) {
	bucketPath := filepath.Join(s.basePath, bucket)
	
	// Check if bucket exists
	if _, err := os.Stat(bucketPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("bucket not found: %s", bucket)
	}
	
	return &models.BucketInfo{
		Name:         bucket,
		Location:     "mock-region",
		StorageClass: "STANDARD",
		CreatedAt:    time.Now(), // Would need to track actual creation time
	}, nil
}

// CreateBucket creates a new mock storage bucket
func (s *MockStorageService) CreateBucket(ctx context.Context, bucket, location string) error {
	bucketPath := filepath.Join(s.basePath, bucket)
	
	// Check if bucket already exists
	if _, err := os.Stat(bucketPath); err == nil {
		return fmt.Errorf("bucket already exists: %s", bucket)
	}
	
	// Create bucket directory
	if err := os.MkdirAll(bucketPath, 0755); err != nil {
		return fmt.Errorf("failed to create bucket: %w", err)
	}
	
	return nil
}

// HealthCheck checks if the mock storage service is healthy
func (s *MockStorageService) HealthCheck(ctx context.Context) error {
	// Check if base path is accessible
	if _, err := os.Stat(s.basePath); err != nil {
		return fmt.Errorf("mock storage path not accessible: %w", err)
	}
	
	return nil
}