package handlers

import (
	"database/sql"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/wavlake/monorepo/internal/models"
	"github.com/wavlake/monorepo/internal/services"
)

type LegacyHandler struct {
	postgresService services.PostgresServiceInterface
}

// NewLegacyHandler creates a new legacy handler
func NewLegacyHandler(postgresService services.PostgresServiceInterface) *LegacyHandler {
	return &LegacyHandler{
		postgresService: postgresService,
	}
}

// isDatabaseError checks if the error is a database/SQL error vs user-not-found
func isDatabaseError(err error) bool {
	if err == nil {
		return false
	}

	// If it's sql.ErrNoRows, it's a legitimate "not found" case
	if err == sql.ErrNoRows {
		return false
	}

	errMsg := err.Error()
	// Check for common database/SQL errors
	databaseErrors := []string{
		"relation", "does not exist",
		"syntax error", "column", "unknown",
		"connection", "timeout", "network",
		"permission denied", "access denied",
		"invalid", "constraint",
	}

	for _, dbErr := range databaseErrors {
		if strings.Contains(strings.ToLower(errMsg), dbErr) {
			return true
		}
	}

	return false
}

// UserMetadataResponse represents the complete user metadata response
type UserMetadataResponse struct {
	User    *models.LegacyUser    `json:"user"`
	Artists []models.LegacyArtist `json:"artists"`
	Albums  []models.LegacyAlbum  `json:"albums"`
	Tracks  []models.LegacyTrack  `json:"tracks"`
}

// GetUserMetadata handles GET /v1/legacy/metadata
// Returns all user metadata from the legacy PostgreSQL system
func (h *LegacyHandler) GetUserMetadata(c *gin.Context) {
	firebaseUID := c.GetString("firebase_uid")

	if firebaseUID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Failed to find an associated Firebase UID"})
		return
	}

	ctx := c.Request.Context()

	// Get user data
	user, err := h.postgresService.GetUserByFirebaseUID(ctx, firebaseUID)
	if err != nil {
		// Check if this is a database error vs user not found
		if isDatabaseError(err) {
			log.Printf("PostgreSQL error getting user %s: %v", firebaseUID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}
		// User not found is normal - return empty response
		user = nil
	}

	// Get all related data (empty arrays if user doesn't exist)
	var artists []models.LegacyArtist
	var albums []models.LegacyAlbum
	var tracks []models.LegacyTrack

	if user != nil {
		// Get artists
		if artistsResult, err := h.postgresService.GetUserArtists(ctx, firebaseUID); err == nil {
			artists = artistsResult
		} else if isDatabaseError(err) {
			log.Printf("PostgreSQL error getting artists for %s: %v", firebaseUID, err)
		}

		// Get albums
		if albumsResult, err := h.postgresService.GetUserAlbums(ctx, firebaseUID); err == nil {
			albums = albumsResult
		} else if isDatabaseError(err) {
			log.Printf("PostgreSQL error getting albums for %s: %v", firebaseUID, err)
		}

		// Get tracks
		if tracksResult, err := h.postgresService.GetUserTracks(ctx, firebaseUID); err == nil {
			tracks = tracksResult
		} else if isDatabaseError(err) {
			log.Printf("PostgreSQL error getting tracks for %s: %v", firebaseUID, err)
		}
	}

	// Ensure we always return empty arrays instead of null
	if artists == nil {
		artists = []models.LegacyArtist{}
	}
	if albums == nil {
		albums = []models.LegacyAlbum{}
	}
	if tracks == nil {
		tracks = []models.LegacyTrack{}
	}

	response := UserMetadataResponse{
		User:    user,
		Artists: artists,
		Albums:  albums,
		Tracks:  tracks,
	}

	c.JSON(http.StatusOK, response)
}

// GetUserTracks handles GET /v1/legacy/tracks
func (h *LegacyHandler) GetUserTracks(c *gin.Context) {
	firebaseUID := c.GetString("firebase_uid")
	if firebaseUID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	tracks, err := h.postgresService.GetUserTracks(c.Request.Context(), firebaseUID)
	if err != nil && isDatabaseError(err) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}

	if tracks == nil {
		tracks = []models.LegacyTrack{}
	}

	c.JSON(http.StatusOK, gin.H{"tracks": tracks})
}

// GetUserArtists handles GET /v1/legacy/artists
func (h *LegacyHandler) GetUserArtists(c *gin.Context) {
	firebaseUID := c.GetString("firebase_uid")
	if firebaseUID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	artists, err := h.postgresService.GetUserArtists(c.Request.Context(), firebaseUID)
	if err != nil && isDatabaseError(err) {
		log.Printf("PostgreSQL error getting artists for %s: %v", firebaseUID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}

	if artists == nil {
		artists = []models.LegacyArtist{}
	}

	c.JSON(http.StatusOK, gin.H{"artists": artists})
}

// GetUserAlbums handles GET /v1/legacy/albums
func (h *LegacyHandler) GetUserAlbums(c *gin.Context) {
	firebaseUID := c.GetString("firebase_uid")
	if firebaseUID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	albums, err := h.postgresService.GetUserAlbums(c.Request.Context(), firebaseUID)
	if err != nil && isDatabaseError(err) {
		log.Printf("PostgreSQL error getting albums for %s: %v", firebaseUID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}

	if albums == nil {
		albums = []models.LegacyAlbum{}
	}

	c.JSON(http.StatusOK, gin.H{"albums": albums})
}