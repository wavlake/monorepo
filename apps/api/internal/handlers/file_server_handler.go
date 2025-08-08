package handlers

import (
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/wavlake/monorepo/internal/models"
	"github.com/wavlake/monorepo/internal/services"
)

// FileServerHandler handles file server operations
type FileServerHandler struct {
	fileServerService services.FileServerServiceInterface
	tokenService      services.TokenServiceInterface
}

// NewFileServerHandler creates a new file server handler
func NewFileServerHandler(fileServerService services.FileServerServiceInterface, tokenService services.TokenServiceInterface) *FileServerHandler {
	return &FileServerHandler{
		fileServerService: fileServerService,
		tokenService:      tokenService,
	}
}

// UploadFileRequest represents the request for file upload
type UploadFileRequest struct {
	Path        string `form:"path" binding:"required"`
	ContentType string `form:"content_type"`
}

// UploadFileResponse represents the response for file upload
type UploadFileResponse struct {
	Success bool                 `json:"success"`
	Data    *models.FileMetadata `json:"data,omitempty"`
	Error   string               `json:"error,omitempty"`
}

// UploadFile handles file upload operations (POST/PUT)
func (h *FileServerHandler) UploadFile(c *gin.Context) {
	// Parse form data
	var req UploadFileRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, UploadFileResponse{
			Success: false,
			Error:   "invalid request parameters",
		})
		return
	}

	// Validate token if provided
	token := c.Query("token")
	if token != "" {
		_, err := h.tokenService.ValidateToken(c.Request.Context(), token, req.Path)
		if err != nil {
			c.JSON(http.StatusUnauthorized, UploadFileResponse{
				Success: false,
				Error:   "invalid or expired token",
			})
			return
		}
	} else {
		// Require authentication if no token
		firebaseUID, exists := c.Get("firebase_uid")
		if !exists {
			c.JSON(http.StatusUnauthorized, UploadFileResponse{
				Success: false,
				Error:   "authentication required",
			})
			return
		}
		_ = firebaseUID // Validate user has access
	}

	// Get uploaded file
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, UploadFileResponse{
			Success: false,
			Error:   "no file uploaded",
		})
		return
	}
	defer file.Close()

	// Determine content type
	contentType := req.ContentType
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	// Upload file
	metadata, err := h.fileServerService.UploadFile(c.Request.Context(), req.Path, file, contentType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, UploadFileResponse{
			Success: false,
			Error:   "failed to upload file",
		})
		return
	}

	c.JSON(http.StatusOK, UploadFileResponse{
		Success: true,
		Data:    metadata,
	})
}

// DownloadFileResponse represents the response for file download
type DownloadFileResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// DownloadFile handles file download operations
func (h *FileServerHandler) DownloadFile(c *gin.Context) {
	path := c.Param("path")
	if path == "" {
		c.JSON(http.StatusBadRequest, DownloadFileResponse{
			Success: false,
			Error:   "file path is required",
		})
		return
	}

	// Get file
	reader, err := h.fileServerService.DownloadFile(c.Request.Context(), path)
	if err != nil {
		c.JSON(http.StatusNotFound, DownloadFileResponse{
			Success: false,
			Error:   "file not found",
		})
		return
	}
	defer reader.Close()

	// Get file metadata for content type
	metadata, err := h.fileServerService.GetFileMetadata(c.Request.Context(), path)
	if err == nil {
		c.Header("Content-Type", metadata.ContentType)
		c.Header("Content-Length", string(rune(metadata.Size)))
	}

	// Stream file to response
	_, err = io.Copy(c.Writer, reader)
	if err != nil {
		c.JSON(http.StatusInternalServerError, DownloadFileResponse{
			Success: false,
			Error:   "failed to stream file",
		})
		return
	}
}

// DeleteFileResponse represents the response for file deletion
type DeleteFileResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
	Message string `json:"message,omitempty"`
}

// DeleteFile handles file deletion operations
func (h *FileServerHandler) DeleteFile(c *gin.Context) {
	path := c.Param("path")
	if path == "" {
		c.JSON(http.StatusBadRequest, DeleteFileResponse{
			Success: false,
			Error:   "file path is required",
		})
		return
	}

	// Validate token if provided
	token := c.Query("token")
	if token != "" {
		_, err := h.tokenService.ValidateToken(c.Request.Context(), token, path)
		if err != nil {
			c.JSON(http.StatusUnauthorized, DeleteFileResponse{
				Success: false,
				Error:   "invalid or expired token",
			})
			return
		}
	} else {
		// Require authentication if no token
		firebaseUID, exists := c.Get("firebase_uid")
		if !exists {
			c.JSON(http.StatusUnauthorized, DeleteFileResponse{
				Success: false,
				Error:   "authentication required",
			})
			return
		}
		_ = firebaseUID // Validate user has access
	}

	// Delete file
	err := h.fileServerService.DeleteFile(c.Request.Context(), path)
	if err != nil {
		c.JSON(http.StatusInternalServerError, DeleteFileResponse{
			Success: false,
			Error:   "failed to delete file",
		})
		return
	}

	c.JSON(http.StatusOK, DeleteFileResponse{
		Success: true,
		Message: "file deleted successfully",
	})
}

// GetStatusResponse represents the response for status check
type GetStatusResponse struct {
	Success bool              `json:"success"`
	Data    map[string]string `json:"data,omitempty"`
	Error   string            `json:"error,omitempty"`
}

// GetStatus handles health check operations
func (h *FileServerHandler) GetStatus(c *gin.Context) {
	c.JSON(http.StatusOK, GetStatusResponse{
		Success: true,
		Data: map[string]string{
			"status":    "healthy",
			"service":   "file-server",
			"timestamp": time.Now().Format(time.RFC3339),
		},
	})
}

// ListFilesResponse represents the response for file listing
type ListFilesResponse struct {
	Success bool     `json:"success"`
	Data    []string `json:"data,omitempty"`
	Error   string   `json:"error,omitempty"`
}

// ListFiles handles file listing operations
func (h *FileServerHandler) ListFiles(c *gin.Context) {
	prefix := c.Query("prefix")

	// Require authentication
	firebaseUID, exists := c.Get("firebase_uid")
	if !exists {
		c.JSON(http.StatusUnauthorized, ListFilesResponse{
			Success: false,
			Error:   "authentication required",
		})
		return
	}
	_ = firebaseUID // Validate user has access

	// List files
	files, err := h.fileServerService.ListFiles(c.Request.Context(), prefix)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ListFilesResponse{
			Success: false,
			Error:   "failed to list files",
		})
		return
	}

	c.JSON(http.StatusOK, ListFilesResponse{
		Success: true,
		Data:    files,
	})
}

// GenerateUploadTokenRequest represents the request for generating upload token
type GenerateUploadTokenRequest struct {
	Path       string `json:"path" binding:"required"`
	Expiration int    `json:"expiration"` // Expiration in minutes
}

// GenerateUploadTokenResponse represents the response for upload token generation
type GenerateUploadTokenResponse struct {
	Success bool                     `json:"success"`
	Data    *models.FileUploadToken  `json:"data,omitempty"`
	Error   string                   `json:"error,omitempty"`
}

// GenerateUploadToken handles upload token generation
func (h *FileServerHandler) GenerateUploadToken(c *gin.Context) {
	var req GenerateUploadTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, GenerateUploadTokenResponse{
			Success: false,
			Error:   "invalid request parameters",
		})
		return
	}

	// Require authentication
	firebaseUID, exists := c.Get("firebase_uid")
	if !exists {
		c.JSON(http.StatusUnauthorized, GenerateUploadTokenResponse{
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
	token, err := h.fileServerService.GenerateUploadToken(c.Request.Context(), req.Path, userID, expiration)
	if err != nil {
		c.JSON(http.StatusInternalServerError, GenerateUploadTokenResponse{
			Success: false,
			Error:   "failed to generate token",
		})
		return
	}

	c.JSON(http.StatusOK, GenerateUploadTokenResponse{
		Success: true,
		Data:    token,
	})
}