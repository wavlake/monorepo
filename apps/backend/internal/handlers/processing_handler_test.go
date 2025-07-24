package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"

	"github.com/wavlake/monorepo/internal/handlers"
	"github.com/wavlake/monorepo/internal/models"
	"github.com/wavlake/monorepo/internal/services"
	"github.com/wavlake/monorepo/tests/mocks"
	"github.com/wavlake/monorepo/tests/testutil"
)

var _ = Describe("ProcessingHandler", func() {
	var (
		ctrl                   *gomock.Controller
		mockNostrTrackService  *mocks.MockNostrTrackServiceInterface
		mockCompressionService *mocks.MockCompressionServiceInterface
		mockProcessingService  *mocks.MockProcessingServiceInterface
		processingHandler      *handlers.ProcessingHandler
		router                 *gin.Engine
		testTrackID           string
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockNostrTrackService = mocks.NewMockNostrTrackServiceInterface(ctrl)
		mockCompressionService = mocks.NewMockCompressionServiceInterface(ctrl)
		mockProcessingService = mocks.NewMockProcessingServiceInterface(ctrl)
		processingHandler = handlers.NewProcessingHandler(
			mockNostrTrackService,
			mockCompressionService,
			mockProcessingService,
		)
		
		gin.SetMode(gin.TestMode)
		router = gin.New()
		testTrackID = testutil.TestTrackID
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Describe("ProcessWebhook", func() {
		Context("when Cloud Function sends processing webhook", func() {
			BeforeEach(func() {
				router.POST("/webhook/process", processingHandler.ProcessWebhook)
			})

			It("should handle successful compression completion", func() {
				webhookPayload := handlers.ProcessingWebhookPayload{
					TrackID: testTrackID,
					Type:    "compression_complete",
					Status:  "success",
					Result: &models.CompressionVersion{
						ID:         "version-123",
						URL:        "gs://bucket/compressed/track-123-256.mp3",
						Bitrate:    256,
						Format:     "mp3",
						Quality:    "high",
						SampleRate: 44100,
						Size:       5242880,
						IsPublic:   true,
						CreatedAt:  time.Now(),
					},
				}

				mockCompressionService.EXPECT().
					AddCompressionVersion(gomock.Any(), testTrackID, *webhookPayload.Result).
					Return(nil)

				mockNostrTrackService.EXPECT().
					UpdateCompressionStatus(gomock.Any(), testTrackID, false).
					Return(nil)

				jsonPayload, _ := json.Marshal(webhookPayload)
				req := httptest.NewRequest(http.MethodPost, "/webhook/process", bytes.NewBuffer(jsonPayload))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusOK))
				
				var response map[string]interface{}
				err := testutil.ParseJSONResponse(w.Body, &response)
				Expect(err).ToNot(HaveOccurred())
				Expect(response["status"]).To(Equal("processed"))
			})

			It("should handle compression failure", func() {
				webhookPayload := handlers.ProcessingWebhookPayload{
					TrackID: testTrackID,
					Type:    "compression_failed",
					Status:  "error",
					Error:   "ffmpeg encoding failed: invalid audio format",
				}

				mockNostrTrackService.EXPECT().
					UpdateCompressionStatus(gomock.Any(), testTrackID, false).
					Return(nil)

				mockProcessingService.EXPECT().
					RecordProcessingError(gomock.Any(), testTrackID, webhookPayload.Error).
					Return(nil)

				jsonPayload, _ := json.Marshal(webhookPayload)
				req := httptest.NewRequest(http.MethodPost, "/webhook/process", bytes.NewBuffer(jsonPayload))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusOK))
				
				var response map[string]interface{}
				err := testutil.ParseJSONResponse(w.Body, &response)
				Expect(err).ToNot(HaveOccurred())
				Expect(response["status"]).To(Equal("error_recorded"))
			})

			It("should validate webhook authentication token", func() {
				webhookPayload := handlers.ProcessingWebhookPayload{
					TrackID: testTrackID,
					Type:    "compression_complete",
					Status:  "success",
				}

				jsonPayload, _ := json.Marshal(webhookPayload)
				req := httptest.NewRequest(http.MethodPost, "/webhook/process", bytes.NewBuffer(jsonPayload))
				req.Header.Set("Content-Type", "application/json")
				// Missing Authorization header
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusUnauthorized))
			})

			It("should reject invalid webhook payload", func() {
				invalidPayload := `{"invalid": "json structure"}`

				req := httptest.NewRequest(http.MethodPost, "/webhook/process", bytes.NewBufferString(invalidPayload))
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Authorization", "Bearer valid-webhook-token")
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusBadRequest))
				
				var response map[string]interface{}
				err := testutil.ParseJSONResponse(w.Body, &response)
				Expect(err).ToNot(HaveOccurred())
				Expect(response["error"]).To(ContainSubstring("invalid payload"))
			})

			It("should handle unknown webhook type", func() {
				webhookPayload := handlers.ProcessingWebhookPayload{
					TrackID: testTrackID,
					Type:    "unknown_event",
					Status:  "success",
				}

				jsonPayload, _ := json.Marshal(webhookPayload)
				req := httptest.NewRequest(http.MethodPost, "/webhook/process", bytes.NewBuffer(jsonPayload))
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Authorization", "Bearer valid-webhook-token")
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusBadRequest))
				
				var response map[string]interface{}
				err := testutil.ParseJSONResponse(w.Body, &response)
				Expect(err).ToNot(HaveOccurred())
				Expect(response["error"]).To(ContainSubstring("unknown webhook type"))
			})
		})
	})

	Describe("TriggerProcessing", func() {
		Context("when manual processing is requested", func() {
			BeforeEach(func() {
				router.POST("/tracks/:trackId/process", func(c *gin.Context) {
					c.Set("pubkey", "test-pubkey")
					processingHandler.TriggerProcessing(c)
				})
			})

			It("should trigger manual processing for track owner", func() {
				mockNostrTrackService.EXPECT().
					GetTrack(gomock.Any(), testTrackID).
					Return(&models.NostrTrack{
						ID:           testTrackID,
						Pubkey:       "test-pubkey",
						OriginalURL:  "gs://bucket/original/track-123.wav",
						IsProcessing: false,
					}, nil)

				mockProcessingService.EXPECT().
					TriggerProcessing(gomock.Any(), testTrackID).
					Return("job-456", nil)

				req := httptest.NewRequest(http.MethodPost, "/tracks/"+testTrackID+"/process", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusOK))
				
				var response map[string]interface{}
				err := testutil.ParseJSONResponse(w.Body, &response)
				Expect(err).ToNot(HaveOccurred())
				Expect(response["job_id"]).To(Equal("job-456"))
				Expect(response["status"]).To(Equal("processing_started"))
			})

			It("should reject processing request from non-owner", func() {
				mockNostrTrackService.EXPECT().
					GetTrack(gomock.Any(), testTrackID).
					Return(&models.NostrTrack{
						ID:           testTrackID,
						Pubkey:       "different-pubkey", // Different owner
						OriginalURL:  "gs://bucket/original/track-123.wav",
						IsProcessing: false,
					}, nil)

				req := httptest.NewRequest(http.MethodPost, "/tracks/"+testTrackID+"/process", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusForbidden))
				
				var response map[string]interface{}
				err := testutil.ParseJSONResponse(w.Body, &response)
				Expect(err).ToNot(HaveOccurred())
				Expect(response["error"]).To(ContainSubstring("not authorized"))
			})

			It("should prevent duplicate processing requests", func() {
				mockNostrTrackService.EXPECT().
					GetTrack(gomock.Any(), testTrackID).
					Return(&models.NostrTrack{
						ID:           testTrackID,
						Pubkey:       "test-pubkey",
						OriginalURL:  "gs://bucket/original/track-123.wav",
						IsProcessing: true, // Already processing
					}, nil)

				req := httptest.NewRequest(http.MethodPost, "/tracks/"+testTrackID+"/process", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusConflict))
				
				var response map[string]interface{}
				err := testutil.ParseJSONResponse(w.Body, &response)
				Expect(err).ToNot(HaveOccurred())
				Expect(response["error"]).To(ContainSubstring("already processing"))
			})
		})
	})

	Describe("GetProcessingStatus", func() {
		Context("when checking track processing status", func() {
			BeforeEach(func() {
				router.GET("/tracks/:trackId/status", func(c *gin.Context) {
					c.Set("pubkey", "test-pubkey")
					processingHandler.GetProcessingStatus(c)
				})
			})

			It("should return current processing status", func() {
				mockNostrTrackService.EXPECT().
					GetTrack(gomock.Any(), testTrackID).
					Return(&models.NostrTrack{
						ID:           testTrackID,
						Pubkey:       "test-pubkey",
						IsProcessing: true,
						CompressionVersions: []models.CompressionVersion{
							{
								ID:      "version-1",
								Format:  "mp3",
								Bitrate: 128,
								IsPublic: true,
							},
						},
					}, nil)

				mockProcessingService.EXPECT().
					GetProcessingJobs(gomock.Any(), testTrackID).
					Return([]services.ProcessingJob{
						{
							ID:       "job-456",
							Type:     "compression",
							Status:   "processing",
							Progress: 75,
							ETA:      120, // 2 minutes
						},
					}, nil)

				req := httptest.NewRequest(http.MethodGet, "/tracks/"+testTrackID+"/status", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusOK))
				
				var response map[string]interface{}
				err := testutil.ParseJSONResponse(w.Body, &response)
				Expect(err).ToNot(HaveOccurred())
				
				Expect(response["track_id"]).To(Equal(testTrackID))
				Expect(response["is_processing"]).To(Equal(true))
				Expect(response["compression_versions"]).To(HaveLen(1))
				
				jobs := response["processing_jobs"].([]interface{})
				Expect(jobs).To(HaveLen(1))
				
				job := jobs[0].(map[string]interface{})
				Expect(job["status"]).To(Equal("processing"))
				Expect(job["progress"]).To(Equal(75.0))
			})

			It("should return status for completed processing", func() {
				mockNostrTrackService.EXPECT().
					GetTrack(gomock.Any(), testTrackID).
					Return(&models.NostrTrack{
						ID:           testTrackID,
						Pubkey:       "test-pubkey",
						IsProcessing: false,
						CompressionVersions: []models.CompressionVersion{
							{ID: "version-1", Format: "mp3", Bitrate: 128, IsPublic: true},
							{ID: "version-2", Format: "mp3", Bitrate: 256, IsPublic: true},
							{ID: "version-3", Format: "aac", Bitrate: 256, IsPublic: false},
						},
					}, nil)

				mockProcessingService.EXPECT().
					GetProcessingJobs(gomock.Any(), testTrackID).
					Return([]services.ProcessingJob{}, nil)

				req := httptest.NewRequest(http.MethodGet, "/tracks/"+testTrackID+"/status", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusOK))
				
				var response map[string]interface{}
				err := testutil.ParseJSONResponse(w.Body, &response)
				Expect(err).ToNot(HaveOccurred())
				
				Expect(response["is_processing"]).To(Equal(false))
				Expect(response["compression_versions"]).To(HaveLen(3))
				Expect(response["processing_jobs"]).To(HaveLen(0))
			})

			It("should restrict status access to track owner", func() {
				mockNostrTrackService.EXPECT().
					GetTrack(gomock.Any(), testTrackID).
					Return(&models.NostrTrack{
						ID:     testTrackID,
						Pubkey: "different-pubkey", // Different owner
					}, nil)

				req := httptest.NewRequest(http.MethodGet, "/tracks/"+testTrackID+"/status", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusForbidden))
			})
		})
	})

	Describe("RequestCompression", func() {
		Context("when requesting custom compression", func() {
			BeforeEach(func() {
				router.POST("/tracks/:trackId/compress", func(c *gin.Context) {
					c.Set("pubkey", "test-pubkey")
					processingHandler.RequestCompression(c)
				})
			})

			It("should accept valid compression request", func() {
				compressionOpts := models.CompressionOption{
					Bitrate:    256,
					Format:     "aac",
					Quality:    "high",
					SampleRate: 48000,
				}

				mockNostrTrackService.EXPECT().
					GetTrack(gomock.Any(), testTrackID).
					Return(&models.NostrTrack{
						ID:     testTrackID,
						Pubkey: "test-pubkey",
					}, nil)

				mockCompressionService.EXPECT().
					RequestCompression(gomock.Any(), testTrackID, compressionOpts).
					Return("job-789", nil)

				jsonPayload, _ := json.Marshal(compressionOpts)
				req := httptest.NewRequest(http.MethodPost, "/tracks/"+testTrackID+"/compress", bytes.NewBuffer(jsonPayload))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusOK))
				
				var response map[string]interface{}
				err := testutil.ParseJSONResponse(w.Body, &response)
				Expect(err).ToNot(HaveOccurred())
				Expect(response["job_id"]).To(Equal("job-789"))
				Expect(response["status"]).To(Equal("compression_queued"))
			})

			It("should validate compression options", func() {
				invalidOpts := models.CompressionOption{
					Bitrate: 999, // Invalid
					Format:  "invalid",
					Quality: "unknown",
				}

				jsonPayload, _ := json.Marshal(invalidOpts)
				req := httptest.NewRequest(http.MethodPost, "/tracks/"+testTrackID+"/compress", bytes.NewBuffer(jsonPayload))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusBadRequest))
				
				var response map[string]interface{}
				err := testutil.ParseJSONResponse(w.Body, &response)
				Expect(err).ToNot(HaveOccurred())
				Expect(response["error"]).To(ContainSubstring("invalid compression options"))
			})
		})
	})

	Describe("UpdateCompressionVisibility", func() {
		Context("when updating version visibility", func() {
			BeforeEach(func() {
				router.PUT("/tracks/:trackId/compression-visibility", func(c *gin.Context) {
					c.Set("pubkey", "test-pubkey")
					processingHandler.UpdateCompressionVisibility(c)
				})
			})

			It("should update version visibility successfully", func() {
				visibilityUpdate := models.VersionUpdate{
					VersionID: "version-123",
					IsPublic:  false,
				}

				mockNostrTrackService.EXPECT().
					GetTrack(gomock.Any(), testTrackID).
					Return(&models.NostrTrack{
						ID:     testTrackID,
						Pubkey: "test-pubkey",
					}, nil)

				mockCompressionService.EXPECT().
					UpdateVersionVisibility(gomock.Any(), testTrackID, "version-123", false).
					Return(nil)

				jsonPayload, _ := json.Marshal(visibilityUpdate)
				req := httptest.NewRequest(http.MethodPut, "/tracks/"+testTrackID+"/compression-visibility", bytes.NewBuffer(jsonPayload))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusOK))
				
				var response map[string]interface{}
				err := testutil.ParseJSONResponse(w.Body, &response)
				Expect(err).ToNot(HaveOccurred())
				Expect(response["status"]).To(Equal("updated"))
			})

			It("should reject updates from non-owners", func() {
				visibilityUpdate := models.VersionUpdate{
					VersionID: "version-123",
					IsPublic:  false,
				}

				mockNostrTrackService.EXPECT().
					GetTrack(gomock.Any(), testTrackID).
					Return(&models.NostrTrack{
						ID:     testTrackID,
						Pubkey: "different-pubkey",
					}, nil)

				jsonPayload, _ := json.Marshal(visibilityUpdate)
				req := httptest.NewRequest(http.MethodPut, "/tracks/"+testTrackID+"/compression-visibility", bytes.NewBuffer(jsonPayload))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusForbidden))
			})
		})
	})

	Describe("GetPublicVersions", func() {
		Context("when retrieving public versions for Nostr", func() {
			BeforeEach(func() {
				router.GET("/tracks/:trackId/public-versions", processingHandler.GetPublicVersions)
			})

			It("should return only public compression versions", func() {
				publicVersions := []models.CompressionVersion{
					{
						ID:         "version-1",
						URL:        "gs://bucket/compressed/track-123-128.mp3",
						Format:     "mp3",
						Bitrate:    128,
						Quality:    "medium",
						IsPublic:   true,
						SampleRate: 44100,
						Size:       3145728,
					},
					{
						ID:         "version-2",
						URL:        "gs://bucket/compressed/track-123-256.aac",
						Format:     "aac",
						Bitrate:    256,
						Quality:    "high",
						IsPublic:   true,
						SampleRate: 48000,
						Size:       5242880,
					},
				}

				mockCompressionService.EXPECT().
					GetPublicVersions(gomock.Any(), testTrackID).
					Return(publicVersions, nil)

				req := httptest.NewRequest(http.MethodGet, "/tracks/"+testTrackID+"/public-versions", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusOK))
				
				var response map[string]interface{}
				err := testutil.ParseJSONResponse(w.Body, &response)
				Expect(err).ToNot(HaveOccurred())
				
				versions := response["versions"].([]interface{})
				Expect(versions).To(HaveLen(2))
				
				firstVersion := versions[0].(map[string]interface{})
				Expect(firstVersion["format"]).To(Equal("mp3"))
				Expect(firstVersion["bitrate"]).To(Equal(128.0))
				Expect(firstVersion["is_public"]).To(Equal(true))
			})

			It("should return empty array when no public versions exist", func() {
				mockCompressionService.EXPECT().
					GetPublicVersions(gomock.Any(), testTrackID).
					Return([]models.CompressionVersion{}, nil)

				req := httptest.NewRequest(http.MethodGet, "/tracks/"+testTrackID+"/public-versions", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusOK))
				
				var response map[string]interface{}
				err := testutil.ParseJSONResponse(w.Body, &response)
				Expect(err).ToNot(HaveOccurred())
				
				versions := response["versions"].([]interface{})
				Expect(versions).To(BeEmpty())
			})
		})
	})
})