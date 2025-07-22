package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go/v4"
	firebase_auth "firebase.google.com/go/v4/auth"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq" // PostgreSQL driver
	"github.com/wavlake/monorepo/internal/auth"
	"github.com/wavlake/monorepo/internal/config"
	// "github.com/wavlake/monorepo/internal/handlers" // PLACEHOLDER FOR PHASE 2
	"github.com/wavlake/monorepo/internal/middleware"
	"github.com/wavlake/monorepo/internal/services"
	"github.com/wavlake/monorepo/internal/utils"
	"google.golang.org/api/option"
)

// getEnvAsInt returns an environment variable as an integer with a default value
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func main() {
	// Load development configuration
	devConfig := config.LoadDevConfig()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	if projectID == "" {
		if devConfig.IsDevelopment {
			projectID = "wavlake-dev"
		} else {
			log.Println("Warning: GOOGLE_CLOUD_PROJECT environment variable not set")
			projectID = "default-project"
		}
	}

	// Storage configuration - GCS only
	bucketName := os.Getenv("GCS_BUCKET_NAME")
	if bucketName == "" {
		log.Println("Warning: GCS_BUCKET_NAME environment variable not set")
		bucketName = "default-bucket"
	}

	tempDir := os.Getenv("TEMP_DIR")
	if tempDir == "" {
		tempDir = "/tmp"
	}

	ctx := context.Background()

	// Initialize Firebase
	var firebaseApp *firebase.App
	var firebaseAuth *firebase_auth.Client
	var err error

	// Skip Firebase initialization in development mode if SKIP_AUTH is enabled
	if devConfig.SkipAuth && devConfig.IsDevelopment {
		log.Println("⚠️  Skipping Firebase initialization (SKIP_AUTH=true in development mode)")
		log.Println("   Firebase-dependent endpoints will not work!")
	} else {
		// Configure Firebase for emulator if environment variable is set
		var firebaseConfig *firebase.Config
		if emulatorHost := os.Getenv("FIREBASE_AUTH_EMULATOR_HOST"); emulatorHost != "" {
			log.Printf("🔧 Using Firebase Auth emulator at: %s", emulatorHost)
			// For Firebase Auth emulator, we still need to initialize normally
			// The SDK will automatically use the emulator based on the environment variable
		}

		// Try to use service account key if available, otherwise use default credentials
		if keyPath := os.Getenv("FIREBASE_SERVICE_ACCOUNT_KEY"); keyPath != "" {
			opt := option.WithCredentialsFile(keyPath)
			firebaseApp, err = firebase.NewApp(ctx, firebaseConfig, opt)
		} else {
			firebaseApp, err = firebase.NewApp(ctx, firebaseConfig)
		}

		if err != nil {
			if devConfig.IsDevelopment {
				log.Printf("⚠️  Failed to initialize Firebase in development mode: %v", err)
				log.Println("   You can set SKIP_AUTH=true to bypass Firebase initialization")
				log.Println("   Or set up Firebase Auth emulator with FIREBASE_AUTH_EMULATOR_HOST")
			} else {
				log.Fatalf("Failed to initialize Firebase: %v", err)
			}
		} else {
			// Initialize Firebase Auth client
			firebaseAuth, err = firebaseApp.Auth(ctx)
			if err != nil {
				if devConfig.IsDevelopment {
					log.Printf("⚠️  Failed to initialize Firebase Auth in development mode: %v", err)
					log.Println("   Firebase-dependent endpoints will not work!")
				} else {
					log.Fatalf("Failed to initialize Firebase Auth: %v", err)
				}
			} else {
				if os.Getenv("FIREBASE_AUTH_EMULATOR_HOST") != "" {
					log.Println("✅ Firebase Auth initialized with emulator")
				} else {
					log.Println("✅ Firebase Auth initialized")
				}
			}
		}
	}

	// Initialize Firestore client
	firestoreClient, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Failed to initialize Firestore: %v", err)
	}
	defer firestoreClient.Close()

	// Initialize PostgreSQL connection (optional) - PLACEHOLDER FOR PHASE 2
	// var postgresService services.PostgresServiceInterface
	// pgConnStr := os.Getenv("PROD_POSTGRES_CONNECTION_STRING_RO")
	// if pgConnStr != "" {
	//     maxOpenConns := getEnvAsInt("POSTGRES_MAX_CONNECTIONS", 10)
	//     maxIdleConns := getEnvAsInt("POSTGRES_MAX_IDLE_CONNECTIONS", 5)
	//
	//     db, err := sql.Open("postgres", pgConnStr)
	//     if err != nil {
	//         log.Fatalf("Failed to open PostgreSQL connection: %v", err)
	//     }
	//     defer db.Close()
	//
	//     // Configure connection pool
	//     db.SetMaxOpenConns(maxOpenConns)
	//     db.SetMaxIdleConns(maxIdleConns)
	//     db.SetConnMaxLifetime(time.Hour)
	//
	//     // Test connection
	//     if err := db.PingContext(ctx); err != nil {
	//         log.Printf("PostgreSQL connection test failed: %v", err)
	//     } else {
	//         postgresService = services.NewPostgresService(db)
	//         log.Println("PostgreSQL connection established successfully")
	//     }
	// } else {
	//     log.Println("PostgreSQL connection string not provided, skipping PostgreSQL setup")
	// }

	// Initialize services - PLACEHOLDER FOR PHASE 2
	// var userService services.UserServiceInterface
	// if firebaseAuth != nil {
	//     userService = services.NewUserService(firestoreClient, firebaseAuth)
	// } else {
	//     // For development without Firebase, we'll need a mock user service
	//     // This would need to be implemented in services if needed
	//     log.Println("⚠️  UserService requires Firebase Auth - some features will not work")
	//     userService = services.NewUserService(firestoreClient, nil) // This might need adjustment based on your service implementation
	// }

	// Initialize storage service (GCS or mock) - PLACEHOLDER FOR PHASE 2
	// var storageService services.StorageServiceInterface
	// if devConfig.MockStorage {
	//     log.Printf("Initializing mock storage service with path: %s", devConfig.MockStoragePath)
	//     storageService, err = services.NewMockStorageService(bucketName, devConfig.MockStoragePath, devConfig.FileServerURL)
	//     if err != nil {
	//         log.Fatalf("Failed to initialize mock storage service: %v", err)
	//     }
	// } else {
	//     log.Printf("Initializing GCS storage service with bucket: %s", bucketName)
	//     realStorageService, err := services.NewStorageService(ctx, bucketName)
	//     if err != nil {
	//         log.Fatalf("Failed to initialize GCS storage service: %v", err)
	//     }
	//     defer realStorageService.Close()
	//     storageService = realStorageService
	// }

	// PLACEHOLDER FOR PHASE 2 - Services will be migrated when packages exist
	// nostrTrackService := services.NewNostrTrackService(firestoreClient, storageService)
	// audioProcessor := utils.NewAudioProcessor(tempDir)
	// processingService := services.NewProcessingService(storageService, nostrTrackService, audioProcessor, tempDir)

	// Initialize middleware - PLACEHOLDER FOR PHASE 2
	// var firebaseMiddleware *auth.FirebaseMiddleware
	// var dualAuthMiddleware *auth.DualAuthMiddleware
	// var flexibleAuthMiddleware *auth.FlexibleAuthMiddleware

	// if firebaseAuth != nil {
	//     firebaseMiddleware = auth.NewFirebaseMiddleware(firebaseAuth)
	//     dualAuthMiddleware = auth.NewDualAuthMiddleware(firebaseAuth)
	//     flexibleAuthMiddleware = auth.NewFlexibleAuthMiddleware(firebaseAuth, firestoreClient)
	// } else if devConfig.IsDevelopment {
	//     log.Println("⚠️  Firebase middleware not initialized - Firebase-dependent endpoints will be disabled")
	// }

	// firebaseLinkGuard := auth.NewFirebaseLinkGuard(firestoreClient)
	// nip98Middleware, err := auth.NewNIP98Middleware(ctx, projectID)
	// if err != nil {
	//     log.Fatalf("Failed to create NIP-98 middleware: %v", err)
	// }

	// Initialize handlers - PLACEHOLDER FOR PHASE 2
	// These will be implemented in Phase 2: Core Migration when handlers package is migrated
	// authHandlers := handlers.NewAuthHandlers(userService)
	// tracksHandler := handlers.NewTracksHandler(nostrTrackService, processingService, audioProcessor)

	// Initialize legacy handler if PostgreSQL is available - PLACEHOLDER FOR PHASE 2
	// var legacyHandler *handlers.LegacyHandler
	// if postgresService != nil {
	//     legacyHandler = handlers.NewLegacyHandler(postgresService)
	// }

	// Set up Gin router
	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Add logging middleware if enabled
	if devConfig.LogRequests || devConfig.LogResponses {
		loggingConfig := middleware.LoggingConfig{
			LogRequests:     devConfig.LogRequests,
			LogResponses:    devConfig.LogResponses,
			LogHeaders:      devConfig.LogHeaders,
			LogRequestBody:  devConfig.LogRequestBody,
			LogResponseBody: devConfig.LogResponseBody,
			MaxBodySize:     1024 * 1024, // 1MB
			SkipPaths:       []string{"/heartbeat"},
			SensitiveHeaders: []string{
				"authorization",
				"x-firebase-token",
				"x-nostr-authorization",
				"cookie",
			},
			SensitiveFields: []string{
				"password",
				"token",
				"secret",
				"key",
				"auth",
			},
		}
		router.Use(middleware.RequestResponseLogging(loggingConfig))
	} else {
		router.Use(gin.Logger())
	}

	router.Use(gin.Recovery())

	// Configure CORS
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = []string{
		"http://localhost:8080",                           // Development
		"http://localhost:3000",                           // Alternative dev port
		"http://localhost:8083",                           // Another dev port
		"https://wavlake.com",                             // Production
		"https://*.wavlake.com",                           // Subdomains
		"https://web-wavlake.vercel.app",                  // Vercel main deployment
		"https://web-git-auth-updates-wavlake.vercel.app", // Vercel auth-updates branch
		"https://*.vercel.app",                            // All Vercel preview deployments
	}

	// In development, allow all origins for LAN access
	if devConfig.IsDevelopment {
		corsConfig.AllowOrigins = []string{"*"}
		corsConfig.AllowCredentials = false // Required when using "*"
	} else {
		corsConfig.AllowCredentials = true
	}
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	corsConfig.AllowHeaders = []string{
		"Origin",
		"Content-Type",
		"Accept",
		"Authorization",
		"X-Nostr-Authorization",
		"X-Requested-With",
		"x-firebase-token",
		"X-Firebase-Token",
	}
	router.Use(cors.New(corsConfig))

	// Heartbeat endpoint (no auth required)
	router.GET("/heartbeat", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	// Development-only endpoints
	if devConfig.IsDevelopment {
		devGroup := router.Group("/dev")
		{
			// Status endpoint showing development configuration
			devGroup.GET("/status", func(c *gin.Context) {
				c.JSON(200, gin.H{
					"mode":               "development",
					"mock_storage":       devConfig.MockStorage,
					"storage_path":       devConfig.MockStoragePath,
					"file_server_url":    devConfig.FileServerURL,
					"firestore_emulator": config.IsFirestoreEmulated(),
					"logging": gin.H{
						"requests":      devConfig.LogRequests,
						"responses":     devConfig.LogResponses,
						"headers":       devConfig.LogHeaders,
						"request_body":  devConfig.LogRequestBody,
						"response_body": devConfig.LogResponseBody,
					},
				})
			})

			// List files in mock storage
			devGroup.GET("/storage/list", func(c *gin.Context) {
				if !devConfig.MockStorage {
					c.JSON(400, gin.H{"error": "Mock storage not enabled"})
					return
				}

				// This would be implemented in the mock storage service
				c.JSON(200, gin.H{
					"message": "File listing endpoint - implement in mock storage service",
					"path":    devConfig.MockStoragePath,
				})
			})

			// Clear mock storage
			devGroup.DELETE("/storage/clear", func(c *gin.Context) {
				if !devConfig.MockStorage {
					c.JSON(400, gin.H{"error": "Mock storage not enabled"})
					return
				}

				c.JSON(200, gin.H{
					"message": "Storage clearing endpoint - implement in mock storage service",
				})
			})
		}
	}

	// API Endpoints - PLACEHOLDER FOR PHASE 2
	// All API endpoints will be added in Phase 2 when handlers and middleware are migrated
	// v1 := router.Group("/v1")
	
	// PLACEHOLDER FOR PHASE 2: Auth endpoints will be added when handlers are migrated
	// authGroup := v1.Group("/auth")
	// {
	//     // Firebase auth only endpoints (only register if Firebase is available)
	//     if firebaseMiddleware != nil {
	//         authGroup.GET("/get-linked-pubkeys", firebaseMiddleware.Middleware(), authHandlers.GetLinkedPubkeys)
	//         authGroup.POST("/unlink-pubkey", firebaseMiddleware.Middleware(), authHandlers.UnlinkPubkey)
	//     } else if devConfig.IsDevelopment {
	//         // Add stub endpoints that return appropriate development errors
	//         authGroup.GET("/get-linked-pubkeys", func(c *gin.Context) {
	//             c.JSON(503, gin.H{"error": "Firebase authentication not available in development mode (SKIP_AUTH=true)"})
	//         })
	//         authGroup.POST("/unlink-pubkey", func(c *gin.Context) {
	//             c.JSON(503, gin.H{"error": "Firebase authentication not available in development mode (SKIP_AUTH=true)"})
	//         })
	//     }
	//
	//     // Dual auth required endpoint (only register if Firebase is available)
	//     if dualAuthMiddleware != nil {
	//         authGroup.POST("/link-pubkey", dualAuthMiddleware.Middleware(), authHandlers.LinkPubkey)
	//     } else if devConfig.IsDevelopment {
	//         authGroup.POST("/link-pubkey", func(c *gin.Context) {
	//             c.JSON(503, gin.H{"error": "Dual authentication not available in development mode (SKIP_AUTH=true)"})
	//         })
	//     }
	//
	//     // NIP-98 signature validation only endpoint (no database lookup required)
	//     authGroup.POST("/check-pubkey-link", gin.WrapH(nip98Middleware.SignatureValidationMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	//         c, _ := gin.CreateTestContext(w)
	//         c.Request = r
	//         if pubkey := r.Context().Value("pubkey"); pubkey != nil {
	//             c.Set("pubkey", pubkey)
	//         }
	//         authHandlers.CheckPubkeyLink(c)
	//     }))))
	// }

	// PLACEHOLDER FOR PHASE 2: Protected endpoints will be added when middleware is migrated
	// protectedGroup := v1.Group("/protected")
	// protectedGroup.Use(gin.WrapH(nip98Middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	//     // Convert back to Gin context
	//     c, _ := gin.CreateTestContext(w)
	//     c.Request = r
	//     c.Next()
	// }))))
	// {
	//     // Add NIP-98 protected endpoints here in the future
	// }

	// PLACEHOLDER FOR PHASE 2: Tracks endpoints will be added when handlers are migrated
	// tracksGroup := v1.Group("/tracks")
	// {
	//     // Public endpoints
	//     tracksGroup.GET("/:id", tracksHandler.GetTrack)
	//
	//     // Webhook endpoint for processing notifications
	//     tracksGroup.POST("/webhook/process", tracksHandler.ProcessTrackWebhook)
	//
	//     // All other track endpoints require handlers and middleware...
	// }

	// PLACEHOLDER FOR PHASE 2: Legacy endpoints will be added when handlers are migrated
	// if legacyHandler != nil && flexibleAuthMiddleware != nil {
	//     legacyGroup := v1.Group("/legacy")
	//     {
	//         legacyGroup.GET("/metadata", flexibleAuthMiddleware.Middleware(), legacyHandler.GetUserMetadata)
	//         legacyGroup.GET("/tracks", flexibleAuthMiddleware.Middleware(), legacyHandler.GetUserTracks)
	//         legacyGroup.GET("/artists", flexibleAuthMiddleware.Middleware(), legacyHandler.GetUserArtists)
	//         legacyGroup.GET("/albums", flexibleAuthMiddleware.Middleware(), legacyHandler.GetUserAlbums)
	//         legacyGroup.GET("/artists/:artist_id/tracks", flexibleAuthMiddleware.Middleware(), legacyHandler.GetTracksByArtist)
	//         legacyGroup.GET("/albums/:album_id/tracks", flexibleAuthMiddleware.Middleware(), legacyHandler.GetTracksByAlbum)
	//     }
	// } else if legacyHandler != nil && devConfig.IsDevelopment {
	//     log.Println("⚠️  Legacy endpoints not registered - FlexibleAuthMiddleware requires Firebase Auth")
	// }

	// Start server
	log.Printf("Starting server on port %s", port)
	if devConfig.IsDevelopment {
		log.Printf("🚀 Running in DEVELOPMENT mode")
		log.Printf("  📁 Storage: %s (path: %s)",
			map[bool]string{true: "Mock", false: "GCS"}[devConfig.MockStorage],
			devConfig.MockStoragePath)
		log.Printf("  📄 Logging: requests=%t responses=%t", devConfig.LogRequests, devConfig.LogResponses)
		log.Printf("  🔧 Firestore: %s",
			map[bool]string{true: "Emulator", false: "Production"}[config.IsFirestoreEmulated()])
		if devConfig.MockStorage {
			log.Printf("  📡 File server: %s", devConfig.FileServerURL)
		}
	} else {
		log.Printf("🏭 Running in PRODUCTION mode")
	}

	go func() {
		if err := router.Run(":" + port); err != nil {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	log.Println("Server shutdown complete")
}
