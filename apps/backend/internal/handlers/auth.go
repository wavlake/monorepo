package handlers

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/wavlake/monorepo/internal/models"
	"github.com/wavlake/monorepo/internal/services"
)

type AuthHandlers struct {
	userService services.UserServiceInterface
}

func NewAuthHandlers(userService services.UserServiceInterface) *AuthHandlers {
	return &AuthHandlers{
		userService: userService,
	}
}

// LinkPubkeyRequest represents the request body for linking a pubkey
type LinkPubkeyRequest struct {
	PubKey string `json:"pubkey,omitempty"`
}

// LinkPubkeyResponse represents the response for linking a pubkey
type LinkPubkeyResponse struct {
	Success     bool   `json:"success"`
	Message     string `json:"message"`
	FirebaseUID string `json:"firebase_uid"`
	PubKey      string `json:"pubkey"`
	LinkedAt    string `json:"linked_at"`
}

// LinkPubkey handles POST /v1/auth/link-pubkey
// Requires dual authentication (Firebase + NIP-98)
func (h *AuthHandlers) LinkPubkey(c *gin.Context) {
	// Get auth info from context (set by DualAuthMiddleware)
	firebaseUID, exists := c.Get("firebase_uid")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing Firebase authentication"})
		return
	}

	nostrPubkey, exists := c.Get("nostr_pubkey")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing Nostr authentication"})
		return
	}

	pubkey := nostrPubkey.(string)
	uid := firebaseUID.(string)

	log.Printf("Firebase UID: %v", firebaseUID)
	log.Printf("Nostr Pubkey: %v", nostrPubkey)
	log.Printf("Auth header: %v", c.GetHeader("Authorization"))
	log.Printf("Nostr Auth header: %v", c.GetHeader("X-Nostr-Authorization"))

	// Optional: validate request body pubkey matches auth pubkey
	var req LinkPubkeyRequest
	if err := c.ShouldBindJSON(&req); err == nil && req.PubKey != "" {
		if req.PubKey != pubkey {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Request pubkey does not match authenticated pubkey"})
			return
		}
	}

	// Link the pubkey to the Firebase user
	err := h.userService.LinkPubkeyToUser(c.Request.Context(), pubkey, uid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response := LinkPubkeyResponse{
		Success:     true,
		Message:     "Pubkey linked successfully to Firebase account",
		FirebaseUID: uid,
		PubKey:      pubkey,
		LinkedAt:    time.Now().Format(time.RFC3339),
	}

	c.JSON(http.StatusOK, response)
}

// UnlinkPubkeyRequest represents the request body for unlinking a pubkey
type UnlinkPubkeyRequest struct {
	PubKey string `json:"pubkey" binding:"required"`
}

// UnlinkPubkeyResponse represents the response for unlinking a pubkey
type UnlinkPubkeyResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	PubKey  string `json:"pubkey"`
}

// UnlinkPubkey handles POST /v1/auth/unlink-pubkey
// Requires Firebase authentication only
func (h *AuthHandlers) UnlinkPubkey(c *gin.Context) {
	// Get Firebase UID from context (set by FirebaseMiddleware)
	firebaseUID, exists := c.Get("firebase_uid")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing Firebase authentication"})
		return
	}

	var req UnlinkPubkeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	uid := firebaseUID.(string)

	// Unlink the pubkey from the Firebase user
	err := h.userService.UnlinkPubkeyFromUser(c.Request.Context(), req.PubKey, uid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response := UnlinkPubkeyResponse{
		Success: true,
		Message: "Pubkey unlinked successfully from Firebase account",
		PubKey:  req.PubKey,
	}

	c.JSON(http.StatusOK, response)
}

// GetLinkedPubkeysResponse represents the response for getting linked pubkeys
type GetLinkedPubkeysResponse struct {
	Success       bool                      `json:"success"`
	FirebaseUID   string                    `json:"firebase_uid"`
	LinkedPubkeys []models.LinkedPubkeyInfo `json:"linked_pubkeys"`
}

// GetLinkedPubkeys handles GET /v1/auth/get-linked-pubkeys
// Requires Firebase authentication only
func (h *AuthHandlers) GetLinkedPubkeys(c *gin.Context) {
	// Get Firebase UID from context (set by FirebaseMiddleware)
	firebaseUID, exists := c.Get("firebase_uid")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing Firebase authentication"})
		return
	}

	uid := firebaseUID.(string)

	// Get linked pubkeys for the user
	pubkeys, err := h.userService.GetLinkedPubkeys(c.Request.Context(), uid)
	if err != nil {
		// Log the actual error for debugging
		c.Header("X-Debug-Error", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve linked pubkeys",
			"debug": err.Error(),
		})
		return
	}

	// Convert to response format
	var linkedPubkeys []models.LinkedPubkeyInfo
	for _, p := range pubkeys {
		info := models.LinkedPubkeyInfo{
			PubKey:   p.Pubkey,
			LinkedAt: p.LinkedAt.Format(time.RFC3339),
		}

		if !p.LastUsedAt.IsZero() {
			info.LastUsedAt = p.LastUsedAt.Format(time.RFC3339)
		}

		linkedPubkeys = append(linkedPubkeys, info)
	}

	// Ensure we always return an empty array instead of null
	if linkedPubkeys == nil {
		linkedPubkeys = []models.LinkedPubkeyInfo{}
	}

	response := GetLinkedPubkeysResponse{
		Success:       true,
		FirebaseUID:   uid,
		LinkedPubkeys: linkedPubkeys,
	}

	c.JSON(http.StatusOK, response)
}

// CheckPubkeyLinkRequest represents the request body for checking pubkey link status
type CheckPubkeyLinkRequest struct {
	PubKey string `json:"pubkey" binding:"required"`
}

// CheckPubkeyLinkResponse represents the response for checking pubkey link status
type CheckPubkeyLinkResponse struct {
	Success     bool   `json:"success"`
	IsLinked    bool   `json:"is_linked"`
	FirebaseUID string `json:"firebase_uid,omitempty"`
	PubKey      string `json:"pubkey"`
	Email       string `json:"email,omitempty"`
}

// CheckPubkeyLink handles POST /v1/auth/check-pubkey-link
// Requires NIP-98 authentication - users can only check their own pubkey
func (h *AuthHandlers) CheckPubkeyLink(c *gin.Context) {
	// Get authenticated pubkey from NIP-98 middleware
	authPubkey, exists := c.Get("pubkey")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing Nostr authentication"})
		return
	}

	var req CheckPubkeyLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body - pubkey is required"})
		return
	}

	// Verify that the authenticated pubkey matches the requested pubkey
	if authPubkey.(string) != req.PubKey {
		c.JSON(http.StatusForbidden, gin.H{"error": "You can only check linking status for your own pubkey"})
		return
	}

	// Check if the pubkey is linked to any Firebase account
	firebaseUID, err := h.userService.GetFirebaseUIDByPubkey(c.Request.Context(), req.PubKey)
	if err != nil {
		// If error is "not found", it means pubkey is not linked
		response := CheckPubkeyLinkResponse{
			Success:     true,
			IsLinked:    false,
			FirebaseUID: "",
			PubKey:      req.PubKey,
			Email:       "",
		}
		c.JSON(http.StatusOK, response)
		return
	}

	// Pubkey is linked - get the user's email address
	email, err := h.userService.GetUserEmail(c.Request.Context(), firebaseUID)
	if err != nil {
		// Log the error but continue without email
		log.Printf("Failed to get email for Firebase UID %s: %v", firebaseUID, err)
		email = ""
	}

	response := CheckPubkeyLinkResponse{
		Success:     true,
		IsLinked:    true,
		FirebaseUID: firebaseUID,
		PubKey:      req.PubKey,
		Email:       email,
	}

	c.JSON(http.StatusOK, response)
}