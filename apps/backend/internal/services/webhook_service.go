package services

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/wavlake/monorepo/internal/models"
)

// WebhookService handles webhook processing
type WebhookService struct {
	processingService ProcessingServiceInterface
	nostrTrackService NostrTrackServiceInterface
}

// NewWebhookService creates a new webhook service
func NewWebhookService(processingService ProcessingServiceInterface, nostrTrackService NostrTrackServiceInterface) *WebhookService {
	return &WebhookService{
		processingService: processingService,
		nostrTrackService: nostrTrackService,
	}
}

// ProcessCloudFunctionWebhook processes webhooks from Cloud Functions
func (s *WebhookService) ProcessCloudFunctionWebhook(ctx context.Context, payload models.WebhookPayload) error {
	switch payload.EventType {
	case "track.processed":
		// Handle track processing completion
		if trackID, ok := payload.Data["track_id"].(string); ok {
			return s.handleTrackProcessed(ctx, trackID, payload.Data)
		}
		return fmt.Errorf("missing track_id in payload")
		
	case "compression.completed":
		// Handle compression completion
		if trackID, ok := payload.Data["track_id"].(string); ok {
			return s.handleCompressionCompleted(ctx, trackID, payload.Data)
		}
		return fmt.Errorf("missing track_id in payload")
		
	default:
		return fmt.Errorf("unsupported event type: %s", payload.EventType)
	}
}

// ProcessStorageWebhook processes webhooks from storage events
func (s *WebhookService) ProcessStorageWebhook(ctx context.Context, payload models.WebhookPayload) error {
	switch payload.EventType {
	case "object.upload":
		// Handle file upload events
		if objectName, ok := payload.Data["object_name"].(string); ok {
			return s.handleFileUploaded(ctx, objectName, payload.Data)
		}
		return fmt.Errorf("missing object_name in payload")
		
	case "object.delete":
		// Handle file deletion events
		if objectName, ok := payload.Data["object_name"].(string); ok {
			return s.handleFileDeleted(ctx, objectName, payload.Data)
		}
		return fmt.Errorf("missing object_name in payload")
		
	default:
		return fmt.Errorf("unsupported storage event type: %s", payload.EventType)
	}
}

// ProcessNostrRelayWebhook processes webhooks from Nostr relay events
func (s *WebhookService) ProcessNostrRelayWebhook(ctx context.Context, payload models.WebhookPayload) error {
	switch payload.EventType {
	case "event.published":
		// Handle Nostr event publication
		if eventID, ok := payload.Data["event_id"].(string); ok {
			return s.handleNostrEventPublished(ctx, eventID, payload.Data)
		}
		return fmt.Errorf("missing event_id in payload")
		
	case "event.deleted":
		// Handle Nostr event deletion
		if eventID, ok := payload.Data["event_id"].(string); ok {
			return s.handleNostrEventDeleted(ctx, eventID, payload.Data)
		}
		return fmt.Errorf("missing event_id in payload")
		
	default:
		return fmt.Errorf("unsupported Nostr event type: %s", payload.EventType)
	}
}

// GetWebhookStatus returns the status of a webhook
func (s *WebhookService) GetWebhookStatus(ctx context.Context, webhookID string) (*models.ProcessingStatus, error) {
	// In a real implementation, this would query a database for webhook status
	return &models.ProcessingStatus{
		TrackID:     webhookID,
		Status:      "completed",
		Progress:    100,
		Message:     "Webhook processed successfully",
		StartedAt:   time.Now().Add(-5 * time.Minute),
		CompletedAt: time.Now(),
	}, nil
}

// RetryFailedWebhooks retries failed webhooks
func (s *WebhookService) RetryFailedWebhooks(ctx context.Context, maxRetries int) error {
	// In a real implementation, this would:
	// - Query failed webhooks from database
	// - Retry processing up to maxRetries times
	// - Update webhook status
	
	if maxRetries <= 0 {
		return fmt.Errorf("maxRetries must be positive")
	}
	
	return nil
}

// ValidateWebhookSignature validates HMAC signature for webhook security
func (s *WebhookService) ValidateWebhookSignature(payload []byte, signature, secret string) error {
	if secret == "" {
		return nil // Skip validation if no secret configured
	}

	// Calculate expected signature
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	expectedSignature := hex.EncodeToString(mac.Sum(nil))

	// Compare signatures
	if !hmac.Equal([]byte(signature), []byte(expectedSignature)) {
		return fmt.Errorf("invalid webhook signature")
	}

	return nil
}

// Helper methods for handling specific webhook events

func (s *WebhookService) handleTrackProcessed(ctx context.Context, trackID string, data map[string]interface{}) error {
	// Update track as processed
	updates := map[string]interface{}{
		"is_processing": false,
	}
	
	if size, ok := data["size"].(float64); ok {
		updates["size"] = int64(size)
	}
	
	if duration, ok := data["duration"].(float64); ok {
		updates["duration"] = int(duration)
	}
	
	return s.nostrTrackService.UpdateTrack(ctx, trackID, updates)
}

func (s *WebhookService) handleCompressionCompleted(ctx context.Context, trackID string, data map[string]interface{}) error {
	// In a real implementation, this would add the compression version to the track
	_ = trackID
	_ = data
	return nil
}

func (s *WebhookService) handleFileUploaded(ctx context.Context, objectName string, data map[string]interface{}) error {
	// In a real implementation, this might trigger processing
	_ = objectName
	_ = data
	return nil
}

func (s *WebhookService) handleFileDeleted(ctx context.Context, objectName string, data map[string]interface{}) error {
	// In a real implementation, this might clean up related records
	_ = objectName
	_ = data
	return nil
}

func (s *WebhookService) handleNostrEventPublished(ctx context.Context, eventID string, data map[string]interface{}) error {
	// In a real implementation, this might update track status
	_ = eventID
	_ = data
	return nil
}

func (s *WebhookService) handleNostrEventDeleted(ctx context.Context, eventID string, data map[string]interface{}) error {
	// In a real implementation, this might mark track as deleted
	_ = eventID
	_ = data
	return nil
}