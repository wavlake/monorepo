package handlers

import (
	"errors"
	
	"github.com/wavlake/monorepo/internal/models"
)

// Common errors
var (
	ErrInvalidSignature = errors.New("invalid webhook signature")
	ErrUnauthorized     = errors.New("unauthorized")
	ErrNotFound         = errors.New("not found")
	ErrInvalidRequest   = errors.New("invalid request")
)

// APIResponse represents a standard API response
type APIResponse struct {
	Success   bool        `json:"success"`
	Message   string      `json:"message,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	Error     string      `json:"error,omitempty"`
	Timestamp string      `json:"timestamp"`
}

// UserResponse represents a user response
type UserResponse struct {
	User *models.User `json:"user"`
}

// TracksResponse represents multiple tracks response
type TracksResponse struct {
	Tracks []models.Track `json:"tracks"`
	Total  int            `json:"total"`
}

// CreateTrackRequest represents a request to create a track
type CreateTrackRequest struct {
	Title      string `json:"title" binding:"required"`
	Artist     string `json:"artist" binding:"required"`
	Album      string `json:"album"`
	Duration   int    `json:"duration" binding:"required"`
	AudioURL   string `json:"audioUrl" binding:"required"`
	ArtworkURL string `json:"artworkUrl"`
	Genre      string `json:"genre"`
	PriceMsat  int64  `json:"priceMsat"`
}