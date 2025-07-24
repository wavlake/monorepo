package services

import (
	"context"
	"fmt"

	"github.com/wavlake/monorepo/internal/models"
)

// CompressionService handles compression version management
type CompressionService struct {
	nostrTrackService NostrTrackServiceInterface
}

// NewCompressionService creates a new compression service
func NewCompressionService(nostrTrackService NostrTrackServiceInterface) *CompressionService {
	return &CompressionService{
		nostrTrackService: nostrTrackService,
	}
}

// RequestCompression requests compression for specific options
func (s *CompressionService) RequestCompression(ctx context.Context, trackID string, options []models.CompressionOption) error {
	// Validate track exists
	track, err := s.nostrTrackService.GetTrack(ctx, trackID)
	if err != nil {
		return fmt.Errorf("track not found: %w", err)
	}

	// Mark as having pending compression
	if err := s.nostrTrackService.SetPendingCompression(ctx, trackID, true); err != nil {
		return fmt.Errorf("failed to set pending compression: %w", err)
	}

	// In a real implementation, this would queue compression jobs
	// For now, just validate the track exists
	_ = track

	return nil
}

// GetCompressionStatus returns the status of compression for a track
func (s *CompressionService) GetCompressionStatus(ctx context.Context, trackID string) (*models.ProcessingStatus, error) {
	track, err := s.nostrTrackService.GetTrack(ctx, trackID)
	if err != nil {
		return nil, fmt.Errorf("track not found: %w", err)
	}

	status := "completed"
	if track.IsProcessing {
		status = "processing"
	} else if track.HasPendingCompression {
		status = "queued"
	}

	return &models.ProcessingStatus{
		TrackID:   trackID,
		Status:    status,
		Progress:  100,
		Message:   "Compression status retrieved",
		StartedAt: track.CreatedAt,
	}, nil
}

// AddCompressionVersion adds a new compression version
func (s *CompressionService) AddCompressionVersion(ctx context.Context, trackID string, version models.CompressionVersion) error {
	return s.nostrTrackService.AddCompressionVersion(ctx, trackID, version)
}

// UpdateVersionVisibility updates the visibility of a compression version
func (s *CompressionService) UpdateVersionVisibility(ctx context.Context, trackID, versionID string, isPublic bool) error {
	updates := []models.VersionUpdate{
		{
			VersionID: versionID,
			IsPublic:  isPublic,
		},
	}
	return s.nostrTrackService.UpdateCompressionVisibility(ctx, trackID, updates)
}

// GetPublicVersions returns only public compression versions
func (s *CompressionService) GetPublicVersions(ctx context.Context, trackID string) ([]models.CompressionVersion, error) {
	track, err := s.nostrTrackService.GetTrack(ctx, trackID)
	if err != nil {
		return nil, fmt.Errorf("track not found: %w", err)
	}

	var publicVersions []models.CompressionVersion
	for _, version := range track.CompressionVersions {
		if version.IsPublic {
			publicVersions = append(publicVersions, version)
		}
	}

	return publicVersions, nil
}

// DeleteCompressionVersion deletes a compression version
func (s *CompressionService) DeleteCompressionVersion(ctx context.Context, trackID, versionID string) error {
	// In a real implementation, this would remove the version from storage and update the track
	// For now, just validate the track exists
	_, err := s.nostrTrackService.GetTrack(ctx, trackID)
	if err != nil {
		return fmt.Errorf("track not found: %w", err)
	}

	// In practice, would modify the CompressionVersions slice
	return nil
}