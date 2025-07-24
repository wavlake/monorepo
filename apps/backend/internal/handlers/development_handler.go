package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/wavlake/monorepo/internal/services"
)

// DevelopmentHandler handles development utilities and debugging
type DevelopmentHandler struct {
	developmentService services.DevelopmentServiceInterface
}

// NewDevelopmentHandler creates a new development handler
func NewDevelopmentHandler(developmentService services.DevelopmentServiceInterface) *DevelopmentHandler {
	return &DevelopmentHandler{
		developmentService: developmentService,
	}
}

// DevResponse represents a generic development response
type DevResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
}

// ResetDatabase handles database reset operations
func (h *DevelopmentHandler) ResetDatabase(c *gin.Context) {
	// Only allow in development environment
	if gin.Mode() == gin.ReleaseMode {
		c.JSON(http.StatusForbidden, DevResponse{
			Success: false,
			Error:   "not available in production",
		})
		return
	}

	err := h.developmentService.ResetDatabase(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, DevResponse{
			Success: false,
			Error:   "failed to reset database",
		})
		return
	}

	c.JSON(http.StatusOK, DevResponse{
		Success: true,
		Message: "database reset successfully",
	})
}

// SeedTestData handles test data seeding operations
func (h *DevelopmentHandler) SeedTestData(c *gin.Context) {
	// Only allow in development environment
	if gin.Mode() == gin.ReleaseMode {
		c.JSON(http.StatusForbidden, DevResponse{
			Success: false,
			Error:   "not available in production",
		})
		return
	}

	err := h.developmentService.SeedTestData(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, DevResponse{
			Success: false,
			Error:   "failed to seed test data",
		})
		return
	}

	c.JSON(http.StatusOK, DevResponse{
		Success: true,
		Message: "test data seeded successfully",
	})
}

// GetSystemInfo handles system information retrieval
func (h *DevelopmentHandler) GetSystemInfo(c *gin.Context) {
	info, err := h.developmentService.GetSystemInfo(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, DevResponse{
			Success: false,
			Error:   "failed to get system info",
		})
		return
	}

	c.JSON(http.StatusOK, DevResponse{
		Success: true,
		Data:    info,
	})
}

// ClearCache handles cache clearing operations
func (h *DevelopmentHandler) ClearCache(c *gin.Context) {
	err := h.developmentService.ClearCache(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, DevResponse{
			Success: false,
			Error:   "failed to clear cache",
		})
		return
	}

	c.JSON(http.StatusOK, DevResponse{
		Success: true,
		Message: "cache cleared successfully",
	})
}

// GenerateTestFilesRequest represents the request for test file generation
type GenerateTestFilesRequest struct {
	Count int `json:"count"`
}

// GenerateTestFiles handles test file generation
func (h *DevelopmentHandler) GenerateTestFiles(c *gin.Context) {
	// Only allow in development environment
	if gin.Mode() == gin.ReleaseMode {
		c.JSON(http.StatusForbidden, DevResponse{
			Success: false,
			Error:   "not available in production",
		})
		return
	}

	count := 10 // default
	if countStr := c.Query("count"); countStr != "" {
		if parsed, err := strconv.Atoi(countStr); err == nil && parsed > 0 {
			count = parsed
		}
	}

	files, err := h.developmentService.GenerateTestFiles(c.Request.Context(), count)
	if err != nil {
		c.JSON(http.StatusInternalServerError, DevResponse{
			Success: false,
			Error:   "failed to generate test files",
		})
		return
	}

	c.JSON(http.StatusOK, DevResponse{
		Success: true,
		Data:    files,
		Message: "test files generated successfully",
	})
}

// SimulateLoadRequest represents the request for load simulation
type SimulateLoadRequest struct {
	Duration int `json:"duration"` // duration in seconds
}

// SimulateLoad handles load simulation for testing
func (h *DevelopmentHandler) SimulateLoad(c *gin.Context) {
	// Only allow in development environment
	if gin.Mode() == gin.ReleaseMode {
		c.JSON(http.StatusForbidden, DevResponse{
			Success: false,
			Error:   "not available in production",
		})
		return
	}

	duration := 30 * time.Second // default
	if durationStr := c.Query("duration"); durationStr != "" {
		if parsed, err := strconv.Atoi(durationStr); err == nil && parsed > 0 {
			duration = time.Duration(parsed) * time.Second
		}
	}

	err := h.developmentService.SimulateLoad(c.Request.Context(), duration)
	if err != nil {
		c.JSON(http.StatusInternalServerError, DevResponse{
			Success: false,
			Error:   "failed to simulate load",
		})
		return
	}

	c.JSON(http.StatusOK, DevResponse{
		Success: true,
		Message: "load simulation completed",
	})
}

// GetLogs handles log retrieval for debugging
func (h *DevelopmentHandler) GetLogs(c *gin.Context) {
	level := c.Query("level")
	if level == "" {
		level = "info"
	}

	limit := 100 // default
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	logs, err := h.developmentService.GetLogs(c.Request.Context(), level, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, DevResponse{
			Success: false,
			Error:   "failed to get logs",
		})
		return
	}

	c.JSON(http.StatusOK, DevResponse{
		Success: true,
		Data:    logs,
	})
}