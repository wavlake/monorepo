package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/wavlake/monorepo/internal/models"
)

// TokenService handles token-based authentication
type TokenService struct {
	// In a real implementation, this would have a database or cache for token storage
	tokens map[string]*models.FileUploadToken
}

// NewTokenService creates a new token service
func NewTokenService() *TokenService {
	return &TokenService{
		tokens: make(map[string]*models.FileUploadToken),
	}
}

// GenerateUploadToken generates a token for file upload
func (s *TokenService) GenerateUploadToken(ctx context.Context, path, userID string, expiration time.Duration) (*models.FileUploadToken, error) {
	// Generate random token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}
	
	token := hex.EncodeToString(tokenBytes)
	expiresAt := time.Now().Add(expiration)
	
	uploadToken := &models.FileUploadToken{
		Token:     token,
		ExpiresAt: expiresAt,
		Path:      path,
		UserID:    userID,
		CreatedAt: time.Now(),
	}
	
	// Store token (in a real implementation, this would be in a database or cache)
	s.tokens[token] = uploadToken
	
	return uploadToken, nil
}

// GenerateDeleteToken generates a token for file deletion
func (s *TokenService) GenerateDeleteToken(ctx context.Context, path, userID string, expiration time.Duration) (*models.FileUploadToken, error) {
	// For simplicity, reuse the same structure as upload token
	return s.GenerateUploadToken(ctx, path, userID, expiration)
}

// ValidateToken validates a token for a specific path
func (s *TokenService) ValidateToken(ctx context.Context, token, path string) (*models.FileUploadToken, error) {
	uploadToken, exists := s.tokens[token]
	if !exists {
		return nil, fmt.Errorf("token not found")
	}
	
	// Check if token is expired
	if time.Now().After(uploadToken.ExpiresAt) {
		// Clean up expired token
		delete(s.tokens, token)
		return nil, fmt.Errorf("token expired")
	}
	
	// Check if path matches
	if uploadToken.Path != path {
		return nil, fmt.Errorf("token path mismatch")
	}
	
	return uploadToken, nil
}

// RevokeToken revokes a token
func (s *TokenService) RevokeToken(ctx context.Context, token string) error {
	if _, exists := s.tokens[token]; !exists {
		return fmt.Errorf("token not found")
	}
	
	delete(s.tokens, token)
	return nil
}

// ListActiveTokens lists all active tokens for a user
func (s *TokenService) ListActiveTokens(ctx context.Context, userID string) ([]models.FileUploadToken, error) {
	var activeTokens []models.FileUploadToken
	now := time.Now()
	
	for token, uploadToken := range s.tokens {
		if uploadToken.UserID == userID && now.Before(uploadToken.ExpiresAt) {
			activeTokens = append(activeTokens, *uploadToken)
		} else if now.After(uploadToken.ExpiresAt) {
			// Clean up expired tokens
			delete(s.tokens, token)
		}
	}
	
	return activeTokens, nil
}

// RefreshToken refreshes a token with new expiration
func (s *TokenService) RefreshToken(ctx context.Context, token string, expiration time.Duration) (*models.FileUploadToken, error) {
	uploadToken, exists := s.tokens[token]
	if !exists {
		return nil, fmt.Errorf("token not found")
	}
	
	// Check if token is not expired
	if time.Now().After(uploadToken.ExpiresAt) {
		delete(s.tokens, token)
		return nil, fmt.Errorf("token expired")
	}
	
	// Update expiration
	uploadToken.ExpiresAt = time.Now().Add(expiration)
	
	return uploadToken, nil
}