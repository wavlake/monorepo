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

// FileServerService handles file server operations
type FileServerService struct {
	basePath     string
	tokenService TokenServiceInterface
}

// NewFileServerService creates a new file server service
func NewFileServerService(basePath string, tokenService TokenServiceInterface) *FileServerService {
	return &FileServerService{
		basePath:     basePath,
		tokenService: tokenService,
	}
}

// UploadFile uploads a file to the file server
func (s *FileServerService) UploadFile(ctx context.Context, path string, data io.Reader, contentType string) (*models.FileMetadata, error) {
	fullPath := filepath.Join(s.basePath, path)
	
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
		Bucket:      "file-server",
		URL:         "/file/" + path,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}, nil
}

// DownloadFile downloads a file from the file server
func (s *FileServerService) DownloadFile(ctx context.Context, path string) (io.ReadCloser, error) {
	fullPath := filepath.Join(s.basePath, path)
	
	file, err := os.Open(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	
	return file, nil
}

// DeleteFile deletes a file from the file server
func (s *FileServerService) DeleteFile(ctx context.Context, path string) error {
	fullPath := filepath.Join(s.basePath, path)
	
	if err := os.Remove(fullPath); err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	
	return nil
}

// ListFiles lists files with the given prefix
func (s *FileServerService) ListFiles(ctx context.Context, prefix string) ([]string, error) {
	searchPath := filepath.Join(s.basePath, prefix)
	
	var files []string
	err := filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			// Get relative path from base
			relPath, err := filepath.Rel(s.basePath, path)
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

// GetFileMetadata gets metadata for a file
func (s *FileServerService) GetFileMetadata(ctx context.Context, path string) (*models.FileMetadata, error) {
	fullPath := filepath.Join(s.basePath, path)
	
	info, err := os.Stat(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}
	
	return &models.FileMetadata{
		Name:        info.Name(),
		Size:        info.Size(),
		ContentType: "application/octet-stream", // Default, would need mime type detection
		Bucket:      "file-server",
		URL:         "/file/" + path,
		CreatedAt:   info.ModTime(),
		UpdatedAt:   info.ModTime(),
	}, nil
}

// GenerateUploadToken generates a token for file upload
func (s *FileServerService) GenerateUploadToken(ctx context.Context, path, userID string, expiration time.Duration) (*models.FileUploadToken, error) {
	return s.tokenService.GenerateUploadToken(ctx, path, userID, expiration)
}