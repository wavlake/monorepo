package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"

	"github.com/wavlake/monorepo/internal/handlers"
	"github.com/wavlake/monorepo/internal/models"
	"github.com/wavlake/monorepo/tests/mocks"
	"github.com/wavlake/monorepo/tests/testutil"
)

var _ = Describe("DevelopmentHandler", func() {
	var (
		ctrl                     *gomock.Controller
		mockNostrTrackService    *mocks.MockNostrTrackServiceInterface
		mockStorageService       *mocks.MockStorageServiceInterface
		mockCompressionService   *mocks.MockCompressionServiceInterface
		mockPostgresService      *mocks.MockPostgresServiceInterface
		developmentHandler       *handlers.DevelopmentHandler
		router                   *gin.Engine
		tempDir                  string
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockNostrTrackService = mocks.NewMockNostrTrackServiceInterface(ctrl)
		mockStorageService = mocks.NewMockStorageServiceInterface(ctrl)
		mockCompressionService = mocks.NewMockCompressionServiceInterface(ctrl)
		mockPostgresService = mocks.NewMockPostgresServiceInterface(ctrl)
		developmentHandler = handlers.NewDevelopmentHandler(
			mockNostrTrackService,
			mockStorageService,
			mockCompressionService,
			mockPostgresService,
		)
		
		gin.SetMode(gin.TestMode)
		router = gin.New()
		
		var err error
		tempDir, err = os.MkdirTemp("", "development_test")
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		ctrl.Finish()
		os.RemoveAll(tempDir)
	})

	Describe("ResetDatabase", func() {
		Context("when resetting development database", func() {
			BeforeEach(func() {
				router.POST("/dev/reset-db", developmentHandler.ResetDatabase)
			})

			It("should reset database with confirmation", func() {
				resetRequest := models.DatabaseResetRequest{
					Confirm:    true,
					PreserveUsers: false,
				}

				mockPostgresService.EXPECT().
					ResetDatabase(gomock.Any(), false).
					Return(nil)

				mockNostrTrackService.EXPECT().
					ResetCollection(gomock.Any()).
					Return(nil)

				jsonPayload, _ := json.Marshal(resetRequest)
				req := httptest.NewRequest(http.MethodPost, "/dev/reset-db", bytes.NewBuffer(jsonPayload))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusOK))
				
				var response map[string]interface{}
				err := testutil.ParseJSONResponse(w.Body, &response)
				Expect(err).ToNot(HaveOccurred())
				Expect(response["status"]).To(Equal("database_reset"))
				Expect(response["timestamp"]).ToNot(BeEmpty())
			})

			It("should preserve users when requested", func() {
				resetRequest := models.DatabaseResetRequest{
					Confirm:       true,
					PreserveUsers: true,
				}

				mockPostgresService.EXPECT().
					ResetDatabase(gomock.Any(), true). // Preserve users
					Return(nil)

				mockNostrTrackService.EXPECT().
					ResetCollection(gomock.Any()).
					Return(nil)

				jsonPayload, _ := json.Marshal(resetRequest)
				req := httptest.NewRequest(http.MethodPost, "/dev/reset-db", bytes.NewBuffer(jsonPayload))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusOK))
			})

			It("should require confirmation", func() {
				resetRequest := models.DatabaseResetRequest{
					Confirm: false,
				}

				jsonPayload, _ := json.Marshal(resetRequest)
				req := httptest.NewRequest(http.MethodPost, "/dev/reset-db", bytes.NewBuffer(jsonPayload))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusBadRequest))
				
				var response map[string]interface{}
				err := testutil.ParseJSONResponse(w.Body, &response)
				Expect(err).ToNot(HaveOccurred())
				Expect(response["error"]).To(ContainSubstring("confirmation required"))
			})

			It("should only work in development environment", func() {
				// Set production environment
				os.Setenv("ENVIRONMENT", "production")
				defer os.Unsetenv("ENVIRONMENT")

				resetRequest := models.DatabaseResetRequest{
					Confirm: true,
				}

				jsonPayload, _ := json.Marshal(resetRequest)
				req := httptest.NewRequest(http.MethodPost, "/dev/reset-db", bytes.NewBuffer(jsonPayload))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusForbidden))
				
				var response map[string]interface{}
				err := testutil.ParseJSONResponse(w.Body, &response)
				Expect(err).ToNot(HaveOccurred())
				Expect(response["error"]).To(ContainSubstring("not available in production"))
			})

			It("should handle database reset errors", func() {
				resetRequest := models.DatabaseResetRequest{
					Confirm: true,
				}

				mockPostgresService.EXPECT().
					ResetDatabase(gomock.Any(), false).
					Return(handlers.ErrDatabaseResetFailure)

				jsonPayload, _ := json.Marshal(resetRequest)
				req := httptest.NewRequest(http.MethodPost, "/dev/reset-db", bytes.NewBuffer(jsonPayload))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusInternalServerError))
				
				var response map[string]interface{}
				err := testutil.ParseJSONResponse(w.Body, &response)
				Expect(err).ToNot(HaveOccurred())
				Expect(response["error"]).To(ContainSubstring("database reset failed"))
			})
		})
	})

	Describe("SeedTestData", func() {
		Context("when seeding development test data", func() {
			BeforeEach(func() {
				router.POST("/dev/seed-data", developmentHandler.SeedTestData)
			})

			It("should seed comprehensive test data", func() {
				seedRequest := models.SeedDataRequest{
					TrackCount:  10,
					ArtistCount: 5,
					AlbumCount:  3,
					UserCount:   2,
				}

				// Expect database seeding
				mockPostgresService.EXPECT().
					SeedTestData(gomock.Any(), seedRequest).
					Return(&models.SeedDataResult{
						TracksCreated:  10,
						ArtistsCreated: 5,
						AlbumsCreated:  3,
						UsersCreated:   2,
					}, nil)

				// Expect Nostr data creation
				mockNostrTrackService.EXPECT().
					CreateTestTracks(gomock.Any(), 10).
					Return([]string{"track-1", "track-2", "track-3"}, nil)

				jsonPayload, _ := json.Marshal(seedRequest)
				req := httptest.NewRequest(http.MethodPost, "/dev/seed-data", bytes.NewBuffer(jsonPayload))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusOK))
				
				var response map[string]interface{}
				err := testutil.ParseJSONResponse(w.Body, &response)
				Expect(err).ToNot(HaveOccurred())
				
				Expect(response["status"]).To(Equal("test_data_seeded"))
				Expect(response["tracks_created"]).To(Equal(10.0))
				Expect(response["artists_created"]).To(Equal(5.0))
				Expect(response["nostr_tracks"]).To(HaveLen(3))
			})

			It("should validate seed data limits", func() {
				seedRequest := models.SeedDataRequest{
					TrackCount:  1000, // Too many
					ArtistCount: 500,  // Too many
				}

				jsonPayload, _ := json.Marshal(seedRequest)
				req := httptest.NewRequest(http.MethodPost, "/dev/seed-data", bytes.NewBuffer(jsonPayload))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusBadRequest))
				
				var response map[string]interface{}
				err := testutil.ParseJSONResponse(w.Body, &response)
				Expect(err).ToNot(HaveOccurred())
				Expect(response["error"]).To(ContainSubstring("exceeds maximum"))
			})

			It("should handle seeding errors gracefully", func() {
				seedRequest := models.SeedDataRequest{
					TrackCount: 5,
				}

				mockPostgresService.EXPECT().
					SeedTestData(gomock.Any(), seedRequest).
					Return(nil, handlers.ErrDatabaseSeedFailure)

				jsonPayload, _ := json.Marshal(seedRequest)
				req := httptest.NewRequest(http.MethodPost, "/dev/seed-data", bytes.NewBuffer(jsonPayload))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusInternalServerError))
			})
		})
	})

	Describe("GetSystemInfo", func() {
		Context("when retrieving system information", func() {
			BeforeEach(func() {
				router.GET("/dev/system-info", developmentHandler.GetSystemInfo)
			})

			It("should return comprehensive system information", func() {
				// Mock storage stats
				mockStorageService.EXPECT().
					GetStorageStats(gomock.Any()).
					Return(&models.StorageStats{
						TotalFiles:      150,
						TotalSize:       1024 * 1024 * 512, // 512MB
						AvailableSpace:  1024 * 1024 * 1024 * 10, // 10GB
						UploadsToday:    25,
						DownloadsToday:  85,
					}, nil)

				// Mock database stats
				mockPostgresService.EXPECT().
					GetDatabaseStats(gomock.Any()).
					Return(&models.DatabaseStats{
						TrackCount:     100,
						ArtistCount:    25,
						AlbumCount:     15,
						UserCount:     10,
						TotalSize:     1024 * 1024 * 50, // 50MB
						ConnectionCount: 5,
					}, nil)

				// Mock Nostr stats
				mockNostrTrackService.EXPECT().
					GetCollectionStats(gomock.Any()).
					Return(&models.NostrStats{
						TrackCount:       95,
						EventCount:       500,
						RelayCount:       3,
						LastSyncTime:     time.Now().Add(-5 * time.Minute),
						PendingUploads:   2,
					}, nil)

				req := httptest.NewRequest(http.MethodGet, "/dev/system-info", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusOK))
				
				var response map[string]interface{}
				err := testutil.ParseJSONResponse(w.Body, &response)
				Expect(err).ToNot(HaveOccurred())
				
				Expect(response["environment"]).ToNot(BeEmpty())
				Expect(response["version"]).ToNot(BeEmpty())
				Expect(response["uptime"]).ToNot(BeEmpty())
				
				storage := response["storage"].(map[string]interface{})
				Expect(storage["total_files"]).To(Equal(150.0))
				
				database := response["database"].(map[string]interface{})
				Expect(database["track_count"]).To(Equal(100.0))
				
				nostr := response["nostr"].(map[string]interface{})
				Expect(nostr["track_count"]).To(Equal(95.0))
			})

			It("should handle partial service failures gracefully", func() {
				// Storage succeeds
				mockStorageService.EXPECT().
					GetStorageStats(gomock.Any()).
					Return(&models.StorageStats{TotalFiles: 50}, nil)

				// Database fails
				mockPostgresService.EXPECT().
					GetDatabaseStats(gomock.Any()).
					Return(nil, handlers.ErrDatabaseConnectionFailure)

				// Nostr succeeds
				mockNostrTrackService.EXPECT().
					GetCollectionStats(gomock.Any()).
					Return(&models.NostrStats{TrackCount: 45}, nil)

				req := httptest.NewRequest(http.MethodGet, "/dev/system-info", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusOK))
				
				var response map[string]interface{}
				err := testutil.ParseJSONResponse(w.Body, &response)
				Expect(err).ToNot(HaveOccurred())
				
				// Should still return storage and nostr data
				Expect(response["storage"]).ToNot(BeNil())
				Expect(response["nostr"]).ToNot(BeNil())
				
				// Database should show error
				database := response["database"].(map[string]interface{})
				Expect(database["error"]).ToNot(BeEmpty())
			})
		})
	})

	Describe("ClearCache", func() {
		Context("when clearing system caches", func() {
			BeforeEach(func() {
				router.POST("/dev/clear-cache", developmentHandler.ClearCache)
			})

			It("should clear all caches successfully", func() {
				cacheRequest := models.ClearCacheRequest{
					CacheTypes: []string{"compression", "storage", "nostr"},
				}

				mockCompressionService.EXPECT().
					ClearCache(gomock.Any()).
					Return(nil)

				mockStorageService.EXPECT().
					ClearCache(gomock.Any()).
					Return(nil)

				mockNostrTrackService.EXPECT().
					ClearCache(gomock.Any()).
					Return(nil)

				jsonPayload, _ := json.Marshal(cacheRequest)
				req := httptest.NewRequest(http.MethodPost, "/dev/clear-cache", bytes.NewBuffer(jsonPayload))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusOK))
				
				var response map[string]interface{}
				err := testutil.ParseJSONResponse(w.Body, &response)
				Expect(err).ToNot(HaveOccurred())
				
				Expect(response["status"]).To(Equal("caches_cleared"))
				clearedCaches := response["cleared_caches"].([]interface{})
				Expect(clearedCaches).To(ContainElements("compression", "storage", "nostr"))
			})

			It("should clear specific cache types", func() {
				cacheRequest := models.ClearCacheRequest{
					CacheTypes: []string{"compression"},
				}

				mockCompressionService.EXPECT().
					ClearCache(gomock.Any()).
					Return(nil)

				jsonPayload, _ := json.Marshal(cacheRequest)
				req := httptest.NewRequest(http.MethodPost, "/dev/clear-cache", bytes.NewBuffer(jsonPayload))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusOK))
				
				var response map[string]interface{}
				err := testutil.ParseJSONResponse(w.Body, &response)
				Expect(err).ToNot(HaveOccurred())
				
				clearedCaches := response["cleared_caches"].([]interface{})
				Expect(clearedCaches).To(ContainElement("compression"))
				Expect(clearedCaches).ToNot(ContainElements("storage", "nostr"))
			})

			It("should handle invalid cache types", func() {
				cacheRequest := models.ClearCacheRequest{
					CacheTypes: []string{"invalid-cache"},
				}

				jsonPayload, _ := json.Marshal(cacheRequest)
				req := httptest.NewRequest(http.MethodPost, "/dev/clear-cache", bytes.NewBuffer(jsonPayload))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusBadRequest))
				
				var response map[string]interface{}
				err := testutil.ParseJSONResponse(w.Body, &response)
				Expect(err).ToNot(HaveOccurred())
				Expect(response["error"]).To(ContainSubstring("invalid cache type"))
			})
		})
	})

	Describe("GenerateTestFiles", func() {
		Context("when generating test audio files", func() {
			BeforeEach(func() {
				router.POST("/dev/generate-test-files", developmentHandler.GenerateTestFiles)
			})

			It("should generate test audio files", func() {
				fileRequest := models.TestFileRequest{
					Count:    3,
					Format:   "wav",
					Duration: 30, // 30 seconds
					SampleRate: 44100,
				}

				expectedFiles := []string{
					filepath.Join(tempDir, "test-audio-1.wav"),
					filepath.Join(tempDir, "test-audio-2.wav"),
					filepath.Join(tempDir, "test-audio-3.wav"),
				}

				mockStorageService.EXPECT().
					GenerateTestFiles(gomock.Any(), fileRequest).
					Return(expectedFiles, nil)

				jsonPayload, _ := json.Marshal(fileRequest)
				req := httptest.NewRequest(http.MethodPost, "/dev/generate-test-files", bytes.NewBuffer(jsonPayload))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusOK))
				
				var response map[string]interface{}
				err := testutil.ParseJSONResponse(w.Body, &response)
				Expect(err).ToNot(HaveOccurred())
				
				Expect(response["status"]).To(Equal("test_files_generated"))
				files := response["files"].([]interface{})
				Expect(files).To(HaveLen(3))
			})

			It("should validate file generation parameters", func() {
				fileRequest := models.TestFileRequest{
					Count:    100, // Too many
					Duration: 600, // Too long
				}

				jsonPayload, _ := json.Marshal(fileRequest)
				req := httptest.NewRequest(http.MethodPost, "/dev/generate-test-files", bytes.NewBuffer(jsonPayload))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusBadRequest))
				
				var response map[string]interface{}
				err := testutil.ParseJSONResponse(w.Body, &response)
				Expect(err).ToNot(HaveOccurred())
				Expect(response["error"]).To(ContainSubstring("exceeds limits"))
			})
		})
	})

	Describe("SimulateLoad", func() {
		Context("when simulating system load", func() {
			BeforeEach(func() {
				router.POST("/dev/simulate-load", developmentHandler.SimulateLoad)
			})

			It("should simulate upload and compression load", func() {
				loadRequest := models.LoadTestRequest{
					Type:        "upload",
					Concurrent:  5,
					Duration:    30, // 30 seconds
					FileSize:    1024 * 1024, // 1MB
				}

				expectedResults := &models.LoadTestResults{
					RequestsCompleted: 150,
					SuccessRate:      95.5,
					AverageLatency:   250, // 250ms
					ErrorCount:       7,
					Duration:         30,
				}

				mockStorageService.EXPECT().
					SimulateLoad(gomock.Any(), loadRequest).
					Return(expectedResults, nil)

				jsonPayload, _ := json.Marshal(loadRequest)
				req := httptest.NewRequest(http.MethodPost, "/dev/simulate-load", bytes.NewBuffer(jsonPayload))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusOK))
				
				var response map[string]interface{}
				err := testutil.ParseJSONResponse(w.Body, &response)
				Expect(err).ToNot(HaveOccurred())
				
				Expect(response["status"]).To(Equal("load_test_completed"))
				Expect(response["requests_completed"]).To(Equal(150.0))
				Expect(response["success_rate"]).To(Equal(95.5))
				Expect(response["average_latency"]).To(Equal(250.0))
			})

			It("should validate load test parameters", func() {
				loadRequest := models.LoadTestRequest{
					Type:       "upload",
					Concurrent: 100, // Too high
					Duration:   600, // Too long
				}

				jsonPayload, _ := json.Marshal(loadRequest)
				req := httptest.NewRequest(http.MethodPost, "/dev/simulate-load", bytes.NewBuffer(jsonPayload))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusBadRequest))
				
				var response map[string]interface{}
				err := testutil.ParseJSONResponse(w.Body, &response)
				Expect(err).ToNot(HaveOccurred())
				Expect(response["error"]).To(ContainSubstring("exceeds safety limits"))
			})
		})
	})

	Describe("GetLogs", func() {
		Context("when retrieving system logs", func() {
			BeforeEach(func() {
				router.GET("/dev/logs", developmentHandler.GetLogs)
			})

			It("should return recent log entries", func() {
				expectedLogs := []models.LogEntry{
					{
						Timestamp: time.Now().Add(-5 * time.Minute),
						Level:     "INFO",
						Message:   "Track uploaded successfully",
						Component: "upload_handler",
						TrackID:   "track-123",
					},
					{
						Timestamp: time.Now().Add(-3 * time.Minute),
						Level:     "ERROR",
						Message:   "Compression failed",
						Component: "compression_service",
						Error:     "ffmpeg timeout",
					},
				}

				mockStorageService.EXPECT().
					GetRecentLogs(gomock.Any(), 100, "").
					Return(expectedLogs, nil)

				req := httptest.NewRequest(http.MethodGet, "/dev/logs?limit=100", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusOK))
				
				var response map[string]interface{}
				err := testutil.ParseJSONResponse(w.Body, &response)
				Expect(err).ToNot(HaveOccurred())
				
				logs := response["logs"].([]interface{})
				Expect(logs).To(HaveLen(2))
				
				firstLog := logs[0].(map[string]interface{})
				Expect(firstLog["level"]).To(Equal("INFO"))
				Expect(firstLog["component"]).To(Equal("upload_handler"))
			})

			It("should support log level filtering", func() {
				req := httptest.NewRequest(http.MethodGet, "/dev/logs?level=ERROR&limit=50", nil)
				
				mockStorageService.EXPECT().
					GetRecentLogs(gomock.Any(), 50, "ERROR").
					Return([]models.LogEntry{}, nil)

				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusOK))
			})
		})
	})
})