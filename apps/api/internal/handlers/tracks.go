package handlers

import (
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/wavlake/monorepo/internal/models"
	"github.com/wavlake/monorepo/internal/services"
)

type TracksHandler struct {
	nostrTrackService services.NostrTrackServiceInterface
	processingService services.ProcessingServiceInterface
	audioProcessor    services.AudioProcessorInterface
}

func NewTracksHandler(nostrTrackService services.NostrTrackServiceInterface, processingService services.ProcessingServiceInterface, audioProcessor services.AudioProcessorInterface) *TracksHandler {
	return &TracksHandler{
		nostrTrackService: nostrTrackService,
		processingService: processingService,
		audioProcessor:    audioProcessor,
	}
}

type CreateNostrTrackRequest struct {
	Extension string `json:"extension" binding:"required"`
}

type CreateTrackResponse struct {
	Success bool               `json:"success"`
	Data    *models.NostrTrack `json:"data,omitempty"`
	Error   string             `json:"error,omitempty"`
	Message string             `json:"message,omitempty"`
}

// CreateTrackNostr creates a new track via NIP-98 authentication
func (h *TracksHandler) CreateTrackNostr(c *gin.Context) {
	var req CreateNostrTrackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, CreateTrackResponse{
			Success: false,
			Error:   "extension field is required",
		})
		return
	}

	// Validate file extension
	if !h.audioProcessor.IsFormatSupported(req.Extension) {
		c.JSON(http.StatusBadRequest, CreateTrackResponse{
			Success: false,
			Error:   "unsupported audio format",
		})
		return
	}

	// Get authenticated user info from NIP-98 middleware context
	pubkey, exists := c.Get("pubkey")
	if !exists {
		c.JSON(http.StatusUnauthorized, CreateTrackResponse{
			Success: false,
			Error:   "authentication required",
		})
		return
	}

	firebaseUID, exists := c.Get("firebase_uid")
	if !exists {
		c.JSON(http.StatusUnauthorized, CreateTrackResponse{
			Success: false,
			Error:   "user account not found",
		})
		return
	}

	pubkeyStr, ok := pubkey.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, CreateTrackResponse{
			Success: false,
			Error:   "invalid pubkey format",
		})
		return
	}

	firebaseUIDStr, ok := firebaseUID.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, CreateTrackResponse{
			Success: false,
			Error:   "invalid user ID format",
		})
		return
	}

	// Create the track
	track, err := h.nostrTrackService.CreateTrack(
		c.Request.Context(),
		pubkeyStr,
		firebaseUIDStr,
		strings.TrimPrefix(req.Extension, "."),
	)
	if err != nil {
		log.Printf("Failed to create track: %v", err)
		c.JSON(http.StatusInternalServerError, CreateTrackResponse{
			Success: false,
			Error:   "failed to create track",
		})
		return
	}

	c.JSON(http.StatusOK, CreateTrackResponse{
		Success: true,
		Data:    track,
	})
}

type GetTracksResponse struct {
	Success bool                 `json:"success"`
	Data    []*models.NostrTrack `json:"data,omitempty"`
	Error   string               `json:"error,omitempty"`
}

// GetMyTracks returns tracks for the authenticated user
func (h *TracksHandler) GetMyTracks(c *gin.Context) {
	// Get authenticated user info from NIP-98 middleware context
	pubkey, exists := c.Get("pubkey")
	if !exists {
		c.JSON(http.StatusUnauthorized, GetTracksResponse{
			Success: false,
			Error:   "authentication required",
		})
		return
	}

	pubkeyStr, ok := pubkey.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, GetTracksResponse{
			Success: false,
			Error:   "invalid pubkey format",
		})
		return
	}

	// Get tracks for this pubkey
	tracks, err := h.nostrTrackService.GetTracksByPubkey(c.Request.Context(), pubkeyStr)
	if err != nil {
		log.Printf("Failed to get tracks for pubkey %s: %v", pubkeyStr, err)
		c.JSON(http.StatusInternalServerError, GetTracksResponse{
			Success: false,
			Error:   "failed to retrieve tracks",
		})
		return
	}

	c.JSON(http.StatusOK, GetTracksResponse{
		Success: true,
		Data:    tracks,
	})
}

// GetTrack returns a specific track by ID
func (h *TracksHandler) GetTrack(c *gin.Context) {
	trackID := c.Param("trackId")
	if trackID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "track ID is required"})
		return
	}

	track, err := h.nostrTrackService.GetTrack(c.Request.Context(), trackID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "track not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": track})
}

// DeleteTrack soft deletes a track
func (h *TracksHandler) DeleteTrack(c *gin.Context) {
	trackID := c.Param("trackId")
	if trackID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "track ID is required"})
		return
	}

	// Get authenticated user info
	pubkey, exists := c.Get("pubkey")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	// Get track to verify ownership
	track, err := h.nostrTrackService.GetTrack(c.Request.Context(), trackID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "track not found"})
		return
	}

	// Verify the user owns this track
	if track.Pubkey != pubkey.(string) {
		c.JSON(http.StatusForbidden, gin.H{"error": "you can only delete your own tracks"})
		return
	}

	// Delete the track
	err = h.nostrTrackService.DeleteTrack(c.Request.Context(), trackID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete track"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "track deleted successfully"})
}