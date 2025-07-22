package main

import (
	"crypto/md5"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type FileServer struct {
	storagePath string
}

func NewFileServer(storagePath string) *FileServer {
	return &FileServer{storagePath: storagePath}
}

// Token validation - simple mock implementation
func validateToken(token string, expires string, path string) bool {
	if token == "" || expires == "" {
		return false
	}

	expiresInt, err := strconv.ParseInt(expires, 10, 64)
	if err != nil {
		return false
	}

	if time.Now().Unix() > expiresInt {
		return false
	}

	// Simple token validation - in a real implementation, this would be signed
	expectedToken := fmt.Sprintf("%x", md5.Sum([]byte(path+expires)))
	return token == expectedToken[:16] // Use first 16 chars
}

func (fs *FileServer) handleUpload(c *gin.Context) {
	// Validate query parameters
	token := c.Query("token")
	expires := c.Query("expires")
	path := c.Query("path")

	if !validateToken(token, expires, path) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
		return
	}

	// Get uploaded file
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}
	defer file.Close()

	// Create full path
	fullPath := filepath.Join(fs.storagePath, path)
	
	// Ensure directory exists
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Printf("Failed to create directory %s: %v", dir, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create directory"})
		return
	}

	// Create destination file
	dst, err := os.Create(fullPath)
	if err != nil {
		log.Printf("Failed to create file %s: %v", fullPath, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create file"})
		return
	}
	defer dst.Close()

	// Copy file contents
	size, err := io.Copy(dst, file)
	if err != nil {
		log.Printf("Failed to copy file: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to copy file"})
		return
	}

	log.Printf("File uploaded successfully: %s (%d bytes)", fullPath, size)

	// Return response similar to GCS
	c.JSON(http.StatusOK, gin.H{
		"name":   path,
		"size":   size,
		"bucket": "mock-bucket",
		"contentType": header.Header.Get("Content-Type"),
	})
}

func (fs *FileServer) handleDownload(c *gin.Context) {
	// Get file path from URL
	filePath := c.Param("filepath")
	fullPath := filepath.Join(fs.storagePath, filePath)

	// Check if file exists
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	// Serve file
	c.File(fullPath)
}

func (fs *FileServer) handleList(c *gin.Context) {
	// List files in storage directory
	prefix := c.Query("prefix")
	searchPath := filepath.Join(fs.storagePath, prefix)
	
	var files []string
	err := filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			// Get relative path from storage root
			relPath, err := filepath.Rel(fs.storagePath, path)
			if err != nil {
				return err
			}
			files = append(files, relPath)
		}
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list files"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"files": files})
}

func (fs *FileServer) handleDelete(c *gin.Context) {
	filePath := c.Param("filepath")
	fullPath := filepath.Join(fs.storagePath, filePath)

	if err := os.Remove(fullPath); err != nil {
		if os.IsNotExist(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete file"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "File deleted successfully"})
}

func (fs *FileServer) handleStatus(c *gin.Context) {
	// Health check endpoint
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
		"storage_path": fs.storagePath,
		"timestamp": time.Now().Unix(),
	})
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	storagePath := os.Getenv("STORAGE_PATH")
	if storagePath == "" {
		storagePath = "./storage"
	}

	// Create storage directory if it doesn't exist
	if err := os.MkdirAll(storagePath, 0755); err != nil {
		log.Fatalf("Failed to create storage directory: %v", err)
	}

	// Set Gin to release mode in production
	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.ReleaseMode)
	}

	fs := NewFileServer(storagePath)
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Upload endpoint (presigned URL destination)
	router.POST("/upload", fs.handleUpload)
	router.PUT("/upload", fs.handleUpload) // Support both POST and PUT

	// Download endpoint
	router.GET("/file/*filepath", fs.handleDownload)

	// Admin endpoints
	router.GET("/status", fs.handleStatus)
	router.GET("/list", fs.handleList)
	router.DELETE("/file/*filepath", fs.handleDelete)

	log.Printf("File server starting on port %s", port)
	log.Printf("Storage path: %s", storagePath)
	
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start file server: %v", err)
	}
}