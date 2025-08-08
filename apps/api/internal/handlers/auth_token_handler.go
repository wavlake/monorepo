package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/wavlake/monorepo/internal/models"
	"github.com/wavlake/monorepo/internal/services"
)

// AuthTokenHandler handles token-based authentication operations
type AuthTokenHandler struct {
	tokenService services.TokenServiceInterface
}

// NewAuthTokenHandler creates a new auth token handler
func NewAuthTokenHandler(tokenService services.TokenServiceInterface) *AuthTokenHandler {
	return &AuthTokenHandler{
		tokenService: tokenService,
	}
}

// TokenResponse represents a generic token response
type TokenResponse struct {
	Success bool                    `json:"success"`
	Data    *models.FileUploadToken `json:"data,omitempty"`
	Error   string                  `json:"error,omitempty"`
	Message string                  `json:"message,omitempty"`
}

// GenerateUploadToken handles upload token generation  
func (h *AuthTokenHandler) GenerateUploadToken(c *gin.Context) {
	type generateUploadTokenRequest struct {
		Path       string `json:"path" binding:"required"`
		Expiration int    `json:"expiration"` // Expiration in minutes
	}
	
	var req generateUploadTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, TokenResponse{
			Success: false,
			Error:   "invalid request parameters",
		})
		return
	}

	// Require authentication
	firebaseUID, exists := c.Get("firebase_uid")
	if !exists {
		c.JSON(http.StatusUnauthorized, TokenResponse{
			Success: false,
			Error:   "authentication required",
		})
		return
	}

	userID := firebaseUID.(string)

	// Set default expiration
	expiration := time.Duration(req.Expiration) * time.Minute
	if expiration == 0 {
		expiration = 1 * time.Hour
	}

	// Generate token
	token, err := h.tokenService.GenerateUploadToken(c.Request.Context(), req.Path, userID, expiration)
	if err != nil {
		c.JSON(http.StatusInternalServerError, TokenResponse{
			Success: false,
			Error:   "failed to generate upload token",
		})
		return
	}

	c.JSON(http.StatusOK, TokenResponse{
		Success: true,
		Data:    token,
	})
}

// GenerateDeleteTokenRequest represents the request for delete token generation
type GenerateDeleteTokenRequest struct {
	Path       string `json:"path" binding:"required"`
	Expiration int    `json:"expiration"` // Expiration in minutes
}

// GenerateDeleteToken handles delete token generation
func (h *AuthTokenHandler) GenerateDeleteToken(c *gin.Context) {
	var req GenerateDeleteTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, TokenResponse{
			Success: false,
			Error:   "invalid request parameters",
		})
		return
	}

	// Require authentication
	firebaseUID, exists := c.Get("firebase_uid")
	if !exists {
		c.JSON(http.StatusUnauthorized, TokenResponse{
			Success: false,
			Error:   "authentication required",
		})
		return
	}

	userID := firebaseUID.(string)

	// Set default expiration
	expiration := time.Duration(req.Expiration) * time.Minute
	if expiration == 0 {
		expiration = 1 * time.Hour
	}

	// Generate delete token
	token, err := h.tokenService.GenerateDeleteToken(c.Request.Context(), req.Path, userID, expiration)
	if err != nil {
		c.JSON(http.StatusInternalServerError, TokenResponse{
			Success: false,
			Error:   "failed to generate delete token",
		})
		return
	}

	c.JSON(http.StatusOK, TokenResponse{
		Success: true,
		Data:    token,
	})
}

// ValidateTokenRequest represents the request for token validation
type ValidateTokenRequest struct {
	Token string `json:"token" binding:"required"`
	Path  string `json:"path" binding:"required"`
}

// ValidateToken handles token validation
func (h *AuthTokenHandler) ValidateToken(c *gin.Context) {
	var req ValidateTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, TokenResponse{
			Success: false,
			Error:   "invalid request parameters",
		})
		return
	}

	// Validate token
	token, err := h.tokenService.ValidateToken(c.Request.Context(), req.Token, req.Path)
	if err != nil {
		c.JSON(http.StatusUnauthorized, TokenResponse{
			Success: false,
			Error:   "invalid or expired token",
		})
		return
	}

	c.JSON(http.StatusOK, TokenResponse{
		Success: true,
		Data:    token,
	})
}

// RevokeTokenRequest represents the request for token revocation
type RevokeTokenRequest struct {
	Token string `json:"token" binding:"required"`
}

// RevokeTokenResponse represents the response for token revocation
type RevokeTokenResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
	Message string `json:"message,omitempty"`
}

// RevokeToken handles token revocation
func (h *AuthTokenHandler) RevokeToken(c *gin.Context) {
	var req RevokeTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, RevokeTokenResponse{
			Success: false,
			Error:   "invalid request parameters",
		})
		return
	}

	// Revoke token
	err := h.tokenService.RevokeToken(c.Request.Context(), req.Token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, RevokeTokenResponse{
			Success: false,
			Error:   "failed to revoke token",
		})
		return
	}

	c.JSON(http.StatusOK, RevokeTokenResponse{
		Success: true,
		Message: "token revoked successfully",
	})
}

// ListActiveTokensResponse represents the response for active token listing
type ListActiveTokensResponse struct {
	Success bool                      `json:"success"`
	Data    []models.FileUploadToken  `json:"data,omitempty"`
	Error   string                    `json:"error,omitempty"`
}

// ListActiveTokens handles active token listing
func (h *AuthTokenHandler) ListActiveTokens(c *gin.Context) {
	// Require authentication
	firebaseUID, exists := c.Get("firebase_uid")
	if !exists {
		c.JSON(http.StatusUnauthorized, ListActiveTokensResponse{
			Success: false,
			Error:   "authentication required",
		})
		return
	}

	userID := firebaseUID.(string)

	// Get active tokens
	tokens, err := h.tokenService.ListActiveTokens(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ListActiveTokensResponse{
			Success: false,
			Error:   "failed to list active tokens",
		})
		return
	}

	c.JSON(http.StatusOK, ListActiveTokensResponse{
		Success: true,
		Data:    tokens,
	})
}

// RefreshTokenRequest represents the request for token refresh
type RefreshTokenRequest struct {
	Token      string `json:"token" binding:"required"`
	Expiration int    `json:"expiration"` // New expiration in minutes
}

// RefreshToken handles token refresh
func (h *AuthTokenHandler) RefreshToken(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, TokenResponse{
			Success: false,
			Error:   "invalid request parameters",
		})
		return
	}

	// Set default expiration
	expiration := time.Duration(req.Expiration) * time.Minute
	if expiration == 0 {
		expiration = 1 * time.Hour
	}

	// Refresh token
	refreshedToken, err := h.tokenService.RefreshToken(c.Request.Context(), req.Token, expiration)
	if err != nil {
		c.JSON(http.StatusUnauthorized, TokenResponse{
			Success: false,
			Error:   "failed to refresh token",
		})
		return
	}

	c.JSON(http.StatusOK, TokenResponse{
		Success: true,
		Data:    refreshedToken,
	})
}