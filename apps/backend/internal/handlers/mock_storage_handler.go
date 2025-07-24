package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/wavlake/monorepo/internal/models"
	"github.com/wavlake/monorepo/internal/services"
)

// MockStorageHandler handles mock storage operations for development
type MockStorageHandler struct {
	mockStorageService services.MockStorageServiceInterface
}

// NewMockStorageHandler creates a new mock storage handler
func NewMockStorageHandler(mockStorageService services.MockStorageServiceInterface) *MockStorageHandler {
	return &MockStorageHandler{
		mockStorageService: mockStorageService,
	}
}

// UploadFileRequest represents the request for mock storage file upload
type MockUploadFileRequest struct {
	Bucket      string `form:"bucket" binding:"required"`
	Path        string `form:"path" binding:"required"`
	ContentType string `form:"content_type"`
}

// UploadFileResponse represents the response for mock storage operations
type MockStorageResponse struct {
	Success bool                 `json:"success"`
	Data    *models.FileMetadata `json:"data,omitempty"`
	Error   string               `json:"error,omitempty"`
}

// UploadFile handles mock storage file upload
func (h *MockStorageHandler) UploadFile(c *gin.Context) {
	var req MockUploadFileRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, MockStorageResponse{
			Success: false,
			Error:   "invalid request parameters",
		})
		return
	}

	// Get uploaded file
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, MockStorageResponse{
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

	// Upload to mock storage
	metadata, err := h.mockStorageService.UploadFile(c.Request.Context(), req.Bucket, req.Path, file, contentType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, MockStorageResponse{
			Success: false,
			Error:   "failed to upload file",
		})
		return
	}

	c.JSON(http.StatusOK, MockStorageResponse{
		Success: true,
		Data:    metadata,
	})
}

// DownloadFile handles mock storage file download
func (h *MockStorageHandler) DownloadFile(c *gin.Context) {
	bucket := c.Param("bucket")
	path := c.Param("path")

	if bucket == "" || path == "" {
		c.JSON(http.StatusBadRequest, MockStorageResponse{
			Success: false,
			Error:   "bucket and path are required",
		})
		return
	}

	// Get file from mock storage
	reader, err := h.mockStorageService.DownloadFile(c.Request.Context(), bucket, path)
	if err != nil {
		c.JSON(http.StatusNotFound, MockStorageResponse{
			Success: false,
			Error:   "file not found",
		})
		return
	}
	defer reader.Close()

	// Stream file to response
	c.DataFromReader(http.StatusOK, -1, "application/octet-stream", reader, nil)
}

// DeleteFile handles mock storage file deletion
type MockDeleteResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
	Message string `json:"message,omitempty"`
}

// DeleteFile handles mock storage file deletion
func (h *MockStorageHandler) DeleteFile(c *gin.Context) {
	bucket := c.Param("bucket")
	path := c.Param("path")

	if bucket == "" || path == "" {
		c.JSON(http.StatusBadRequest, MockDeleteResponse{
			Success: false,
			Error:   "bucket and path are required",
		})
		return
	}

	// Delete from mock storage
	err := h.mockStorageService.DeleteFile(c.Request.Context(), bucket, path)
	if err != nil {
		c.JSON(http.StatusInternalServerError, MockDeleteResponse{
			Success: false,
			Error:   "failed to delete file",
		})
		return
	}

	c.JSON(http.StatusOK, MockDeleteResponse{
		Success: true,
		Message: "file deleted successfully",
	})
}

// ListFilesResponse represents the response for mock storage file listing
type MockListFilesResponse struct {
	Success bool     `json:"success"`
	Data    []string `json:"data,omitempty"`
	Error   string   `json:"error,omitempty"`
}

// ListFiles handles mock storage file listing
func (h *MockStorageHandler) ListFiles(c *gin.Context) {
	bucket := c.Query("bucket")
	prefix := c.Query("prefix")

	if bucket == "" {
		c.JSON(http.StatusBadRequest, MockListFilesResponse{
			Success: false,
			Error:   "bucket parameter is required",
		})
		return
	}

	// List files from mock storage
	files, err := h.mockStorageService.ListFiles(c.Request.Context(), bucket, prefix)
	if err != nil {
		c.JSON(http.StatusInternalServerError, MockListFilesResponse{
			Success: false,
			Error:   "failed to list files",
		})
		return
	}

	c.JSON(http.StatusOK, MockListFilesResponse{
		Success: true,
		Data:    files,
	})
}

// GetBucketInfoResponse represents the response for bucket information
type GetBucketInfoResponse struct {
	Success bool                `json:"success"`
	Data    *models.BucketInfo  `json:"data,omitempty"`
	Error   string              `json:"error,omitempty"`
}

// GetBucketInfo handles mock storage bucket information retrieval
func (h *MockStorageHandler) GetBucketInfo(c *gin.Context) {
	bucket := c.Param("bucket")
	if bucket == "" {
		c.JSON(http.StatusBadRequest, GetBucketInfoResponse{
			Success: false,
			Error:   "bucket parameter is required",
		})
		return
	}

	// Get bucket info from mock storage
	info, err := h.mockStorageService.GetBucketInfo(c.Request.Context(), bucket)
	if err != nil {
		c.JSON(http.StatusNotFound, GetBucketInfoResponse{
			Success: false,
			Error:   "bucket not found",
		})
		return
	}

	c.JSON(http.StatusOK, GetBucketInfoResponse{
		Success: true,
		Data:    info,
	})
}

// CreateBucketRequest represents the request for bucket creation
type CreateBucketRequest struct {
	Bucket   string `json:"bucket" binding:"required"`
	Location string `json:"location"`
}

// CreateBucketResponse represents the response for bucket creation
type CreateBucketResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
	Message string `json:"message,omitempty"`
}

// CreateBucket handles mock storage bucket creation
func (h *MockStorageHandler) CreateBucket(c *gin.Context) {
	var req CreateBucketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, CreateBucketResponse{
			Success: false,
			Error:   "invalid request parameters",
		})
		return
	}

	// Set default location
	location := req.Location
	if location == "" {
		location = "us-central1"
	}

	// Create bucket in mock storage
	err := h.mockStorageService.CreateBucket(c.Request.Context(), req.Bucket, location)
	if err != nil {
		c.JSON(http.StatusInternalServerError, CreateBucketResponse{
			Success: false,
			Error:   "failed to create bucket",
		})
		return
	}

	c.JSON(http.StatusOK, CreateBucketResponse{
		Success: true,
		Message: "bucket created successfully",
	})
}

// HealthCheckResponse represents the response for health check
type MockHealthCheckResponse struct {
	Success bool   `json:"success"`
	Status  string `json:"status,omitempty"`
	Error   string `json:"error,omitempty"`
}

// HealthCheck handles mock storage health check
func (h *MockStorageHandler) HealthCheck(c *gin.Context) {
	err := h.mockStorageService.HealthCheck(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, MockHealthCheckResponse{
			Success: false,
			Error:   "mock storage service unavailable",
		})
		return
	}

	c.JSON(http.StatusOK, MockHealthCheckResponse{
		Success: true,
		Status:  "healthy",
	})
}