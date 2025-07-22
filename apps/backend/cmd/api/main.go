package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	// Create Gin router
	r := gin.Default()

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"service":   "wavlake-api",
			"version":   "1.0.0",
			"timestamp": "2024-01-01T00:00:00Z",
		})
	})

	// API routes placeholder
	api := r.Group("/api/v1")
	{
		api.GET("/ping", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "pong"})
		})
	}

	// Start server
	port := "8080"
	log.Printf("🎵 Wavlake API server starting on port %s", port)
	log.Fatal(r.Run(":" + port))
}