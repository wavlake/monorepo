package handlers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/wavlake/monorepo/internal/models"
	"github.com/wavlake/monorepo/internal/services"
)

// WebhookHandler handles webhook operations from Cloud Functions
type WebhookHandler struct {
	webhookService services.WebhookServiceInterface
}

// NewWebhookHandler creates a new webhook handler
func NewWebhookHandler(webhookService services.WebhookServiceInterface) *WebhookHandler {
	return &WebhookHandler{
		webhookService: webhookService,
	}
}

// WebhookResponse represents a generic webhook response
type WebhookResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
}

// CloudFunctionWebhook handles webhooks from Cloud Functions
func (h *WebhookHandler) CloudFunctionWebhook(c *gin.Context) {
	// Read the request body
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, WebhookResponse{
			Success: false,
			Error:   "failed to read request body",
		})
		return
	}

	// Validate HMAC signature if present
	signature := c.GetHeader("X-Webhook-Signature")
	if signature != "" {
		secret := c.GetHeader("X-Webhook-Secret")
		if err := h.validateSignature(body, signature, secret); err != nil {
			c.JSON(http.StatusUnauthorized, WebhookResponse{
				Success: false,
				Error:   "invalid webhook signature",
			})
			return
		}
	}

	// Parse the webhook payload
	var payload models.WebhookPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, WebhookResponse{
			Success: false,
			Error:   "invalid webhook payload",
		})
		return
	}

	// Process the webhook
	err = h.webhookService.ProcessCloudFunctionWebhook(c.Request.Context(), payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, WebhookResponse{
			Success: false,
			Error:   "failed to process webhook",
		})
		return
	}

	c.JSON(http.StatusOK, WebhookResponse{
		Success: true,
		Message: "webhook processed successfully",
	})
}

// StorageWebhook handles webhooks from storage events
func (h *WebhookHandler) StorageWebhook(c *gin.Context) {
	// Read the request body
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, WebhookResponse{
			Success: false,
			Error:   "failed to read request body",
		})
		return
	}

	// Validate HMAC signature if present
	signature := c.GetHeader("X-Webhook-Signature")
	if signature != "" {
		secret := c.GetHeader("X-Webhook-Secret")
		if err := h.validateSignature(body, signature, secret); err != nil {
			c.JSON(http.StatusUnauthorized, WebhookResponse{
				Success: false,
				Error:   "invalid webhook signature",
			})
			return
		}
	}

	// Parse the webhook payload
	var payload models.WebhookPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, WebhookResponse{
			Success: false,
			Error:   "invalid webhook payload",
		})
		return
	}

	// Process the storage webhook
	err = h.webhookService.ProcessStorageWebhook(c.Request.Context(), payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, WebhookResponse{
			Success: false,
			Error:   "failed to process storage webhook",
		})
		return
	}

	c.JSON(http.StatusOK, WebhookResponse{
		Success: true,
		Message: "storage webhook processed successfully",
	})
}

// NostrRelayWebhook handles webhooks from Nostr relay events
func (h *WebhookHandler) NostrRelayWebhook(c *gin.Context) {
	// Read the request body
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, WebhookResponse{
			Success: false,
			Error:   "failed to read request body",
		})
		return
	}

	// Validate HMAC signature if present
	signature := c.GetHeader("X-Webhook-Signature")
	if signature != "" {
		secret := c.GetHeader("X-Webhook-Secret")
		if err := h.validateSignature(body, signature, secret); err != nil {
			c.JSON(http.StatusUnauthorized, WebhookResponse{
				Success: false,
				Error:   "invalid webhook signature",
			})
			return
		}
	}

	// Parse the webhook payload
	var payload models.WebhookPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, WebhookResponse{
			Success: false,
			Error:   "invalid webhook payload",
		})
		return
	}

	// Process the Nostr relay webhook
	err = h.webhookService.ProcessNostrRelayWebhook(c.Request.Context(), payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, WebhookResponse{
			Success: false,
			Error:   "failed to process Nostr relay webhook",
		})
		return
	}

	c.JSON(http.StatusOK, WebhookResponse{
		Success: true,
		Message: "Nostr relay webhook processed successfully",
	})
}

// WebhookStatus handles webhook status queries
func (h *WebhookHandler) WebhookStatus(c *gin.Context) {
	webhookID := c.Param("id")
	if webhookID == "" {
		c.JSON(http.StatusBadRequest, WebhookResponse{
			Success: false,
			Error:   "webhook ID is required",
		})
		return
	}

	// Get webhook status
	status, err := h.webhookService.GetWebhookStatus(c.Request.Context(), webhookID)
	if err != nil {
		c.JSON(http.StatusNotFound, WebhookResponse{
			Success: false,
			Error:   "webhook status not found",
		})
		return
	}

	c.JSON(http.StatusOK, WebhookResponse{
		Success: true,
		Data:    status,
	})
}

// RetryFailedWebhooks handles retry of failed webhooks
func (h *WebhookHandler) RetryFailedWebhooks(c *gin.Context) {
	maxRetries := 3 // default
	if maxRetriesStr := c.Query("max_retries"); maxRetriesStr != "" {
		if parsed, err := strconv.Atoi(maxRetriesStr); err == nil && parsed > 0 {
			maxRetries = parsed
		}
	}

	// Retry failed webhooks
	err := h.webhookService.RetryFailedWebhooks(c.Request.Context(), maxRetries)
	if err != nil {
		c.JSON(http.StatusInternalServerError, WebhookResponse{
			Success: false,
			Error:   "failed to retry webhooks",
		})
		return
	}

	c.JSON(http.StatusOK, WebhookResponse{
		Success: true,
		Message: "failed webhooks retry initiated",
	})
}

// validateSignature validates HMAC signature for webhook security
func (h *WebhookHandler) validateSignature(payload []byte, signature, secret string) error {
	if secret == "" {
		return nil // Skip validation if no secret configured
	}

	// Calculate expected signature
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	expectedSignature := hex.EncodeToString(mac.Sum(nil))

	// Compare signatures
	if !hmac.Equal([]byte(signature), []byte(expectedSignature)) {
		return ErrInvalidSignature
	}

	return nil
}