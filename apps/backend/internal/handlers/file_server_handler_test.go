package handlers_test

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
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

var _ = Describe("FileServerHandler", func() {
	var (
		ctrl                  *gomock.Controller
		mockStorageService    *mocks.MockStorageServiceInterface
		mockTokenService      *mocks.MockTokenServiceInterface
		fileServerHandler     *handlers.FileServerHandler
		router                *gin.Engine
		testToken             string
		testFileContent       []byte
		tempDir               string
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockStorageService = mocks.NewMockStorageServiceInterface(ctrl)
		mockTokenService = mocks.NewMockTokenServiceInterface(ctrl)
		fileServerHandler = handlers.NewFileServerHandler(mockStorageService, mockTokenService)
		
		gin.SetMode(gin.TestMode)
		router = gin.New()
		testToken = "valid-upload-token-123"
		testFileContent = []byte("test audio file content")
		
		var err error
		tempDir, err = os.MkdirTemp("", "file_server_test")
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		ctrl.Finish()
		os.RemoveAll(tempDir)
	})

	Describe("UploadFile", func() {
		Context("when uploading files via POST", func() {
			BeforeEach(func() {
				router.POST("/upload", fileServerHandler.UploadFile)
			})

			It("should accept valid file upload with authentication token", func() {
				expectedURL := "gs://bucket/uploads/test-file-123.wav"
				
				mockTokenService.EXPECT().
					ValidateUploadToken(testToken).
					Return(&models.UploadTokenClaims{
						UserID:    "user-123",
						TrackID:   "track-456",
						ExpiresAt: time.Now().Add(1 * time.Hour),
					}, nil)

				mockStorageService.EXPECT().
					UploadFile(gomock.Any(), gomock.Any(), "test-audio.wav", testFileContent).
					Return(expectedURL, nil)

				// Create multipart form request
				body := &bytes.Buffer{}
				writer := multipart.NewWriter(body)
				part, err := writer.CreateFormFile("file", "test-audio.wav")
				Expect(err).ToNot(HaveOccurred())
				_, err = part.Write(testFileContent)
				Expect(err).ToNot(HaveOccurred())
				writer.Close()

				req := httptest.NewRequest(http.MethodPost, "/upload", body)
				req.Header.Set("Content-Type", writer.FormDataContentType())
				req.Header.Set("Authorization", "Bearer "+testToken)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusOK))
				
				var response map[string]interface{}
				err = testutil.ParseJSONResponse(w.Body, &response)
				Expect(err).ToNot(HaveOccurred())
				
				Expect(response["status"]).To(Equal("uploaded"))
				Expect(response["url"]).To(Equal(expectedURL))
				Expect(response["filename"]).To(Equal("test-audio.wav"))
				Expect(response["size"]).To(Equal(float64(len(testFileContent))))
			})

			It("should reject upload without authorization token", func() {
				body := &bytes.Buffer{}
				writer := multipart.NewWriter(body)
				part, err := writer.CreateFormFile("file", "test-audio.wav")
				Expect(err).ToNot(HaveOccurred())
				_, err = part.Write(testFileContent)
				Expect(err).ToNot(HaveOccurred())
				writer.Close()

				req := httptest.NewRequest(http.MethodPost, "/upload", body)
				req.Header.Set("Content-Type", writer.FormDataContentType())
				// Missing Authorization header
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusUnauthorized))
				
				var response map[string]interface{}
				err = testutil.ParseJSONResponse(w.Body, &response)
				Expect(err).ToNot(HaveOccurred())
				Expect(response["error"]).To(ContainSubstring("missing authorization"))
			})

			It("should reject upload with invalid token", func() {
				invalidToken := "invalid-token"
				
				mockTokenService.EXPECT().
					ValidateUploadToken(invalidToken).
					Return(nil, handlers.ErrInvalidToken)

				body := &bytes.Buffer{}
				writer := multipart.NewWriter(body)
				part, err := writer.CreateFormFile("file", "test-audio.wav")
				Expect(err).ToNot(HaveOccurred())
				_, err = part.Write(testFileContent)
				Expect(err).ToNot(HaveOccurred())
				writer.Close()

				req := httptest.NewRequest(http.MethodPost, "/upload", body)
				req.Header.Set("Content-Type", writer.FormDataContentType())
				req.Header.Set("Authorization", "Bearer "+invalidToken)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusUnauthorized))
			})

			It("should reject upload with expired token", func() {
				expiredClaims := &models.UploadTokenClaims{
					UserID:    "user-123",
					TrackID:   "track-456",
					ExpiresAt: time.Now().Add(-1 * time.Hour), // Expired
				}
				
				mockTokenService.EXPECT().
					ValidateUploadToken(testToken).
					Return(expiredClaims, nil)

				body := &bytes.Buffer{}
				writer := multipart.NewWriter(body)
				part, err := writer.CreateFormFile("file", "test-audio.wav")
				Expect(err).ToNot(HaveOccurred())
				_, err = part.Write(testFileContent)
				Expect(err).ToNot(HaveOccurred())
				writer.Close()

				req := httptest.NewRequest(http.MethodPost, "/upload", body)
				req.Header.Set("Content-Type", writer.FormDataContentType())
				req.Header.Set("Authorization", "Bearer "+testToken)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusUnauthorized))
				
				var response map[string]interface{}
				err = testutil.ParseJSONResponse(w.Body, &response)
				Expect(err).ToNot(HaveOccurred())
				Expect(response["error"]).To(ContainSubstring("token expired"))
			})

			It("should validate file format", func() {
				mockTokenService.EXPECT().
					ValidateUploadToken(testToken).
					Return(&models.UploadTokenClaims{
						UserID:    "user-123",
						TrackID:   "track-456",
						ExpiresAt: time.Now().Add(1 * time.Hour),
					}, nil)

				// Try to upload unsupported file type
				body := &bytes.Buffer{}
				writer := multipart.NewWriter(body)
				part, err := writer.CreateFormFile("file", "document.pdf")
				Expect(err).ToNot(HaveOccurred())
				_, err = part.Write([]byte("PDF content"))
				Expect(err).ToNot(HaveOccurred())
				writer.Close()

				req := httptest.NewRequest(http.MethodPost, "/upload", body)
				req.Header.Set("Content-Type", writer.FormDataContentType())
				req.Header.Set("Authorization", "Bearer "+testToken)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusBadRequest))
				
				var response map[string]interface{}
				err = testutil.ParseJSONResponse(w.Body, &response)
				Expect(err).ToNot(HaveOccurred())
				Expect(response["error"]).To(ContainSubstring("unsupported file format"))
			})

			It("should enforce file size limits", func() {
				mockTokenService.EXPECT().
					ValidateUploadToken(testToken).
					Return(&models.UploadTokenClaims{
						UserID:    "user-123",
						TrackID:   "track-456",
						ExpiresAt: time.Now().Add(1 * time.Hour),
					}, nil)

				// Create large file content (exceed limit)
				largeContent := make([]byte, 100*1024*1024) // 100MB

				body := &bytes.Buffer{}
				writer := multipart.NewWriter(body)
				part, err := writer.CreateFormFile("file", "large-audio.wav")
				Expect(err).ToNot(HaveOccurred())
				_, err = part.Write(largeContent)
				Expect(err).ToNot(HaveOccurred())
				writer.Close()

				req := httptest.NewRequest(http.MethodPost, "/upload", body)
				req.Header.Set("Content-Type", writer.FormDataContentType())
				req.Header.Set("Authorization", "Bearer "+testToken)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusRequestEntityTooLarge))
				
				var response map[string]interface{}
				err = testutil.ParseJSONResponse(w.Body, &response)
				Expect(err).ToNot(HaveOccurred())
				Expect(response["error"]).To(ContainSubstring("file too large"))
			})

			It("should handle storage service errors", func() {
				mockTokenService.EXPECT().
					ValidateUploadToken(testToken).
					Return(&models.UploadTokenClaims{
						UserID:    "user-123",
						TrackID:   "track-456",
						ExpiresAt: time.Now().Add(1 * time.Hour),
					}, nil)

				mockStorageService.EXPECT().
					UploadFile(gomock.Any(), gomock.Any(), "test-audio.wav", testFileContent).
					Return("", handlers.ErrStorageFailure)

				body := &bytes.Buffer{}
				writer := multipart.NewWriter(body)
				part, err := writer.CreateFormFile("file", "test-audio.wav")
				Expect(err).ToNot(HaveOccurred())
				_, err = part.Write(testFileContent)
				Expect(err).ToNot(HaveOccurred())
				writer.Close()

				req := httptest.NewRequest(http.MethodPost, "/upload", body)
				req.Header.Set("Content-Type", writer.FormDataContentType())
				req.Header.Set("Authorization", "Bearer "+testToken)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusInternalServerError))
				
				var response map[string]interface{}
				err = testutil.ParseJSONResponse(w.Body, &response)
				Expect(err).ToNot(HaveOccurred())
				Expect(response["error"]).To(ContainSubstring("storage error"))
			})
		})

		Context("when uploading files via PUT", func() {
			BeforeEach(func() {
				router.PUT("/upload", fileServerHandler.UploadFile)
			})

			It("should accept PUT uploads with same validation", func() {
				expectedURL := "gs://bucket/uploads/test-file-456.mp3"
				
				mockTokenService.EXPECT().
					ValidateUploadToken(testToken).
					Return(&models.UploadTokenClaims{
						UserID:    "user-123",
						TrackID:   "track-456",
						ExpiresAt: time.Now().Add(1 * time.Hour),
					}, nil)

				mockStorageService.EXPECT().
					UploadFile(gomock.Any(), gomock.Any(), "audio.mp3", testFileContent).
					Return(expectedURL, nil)

				req := httptest.NewRequest(http.MethodPut, "/upload", bytes.NewReader(testFileContent))
				req.Header.Set("Content-Type", "audio/mpeg")
				req.Header.Set("Authorization", "Bearer "+testToken)
				req.Header.Set("X-Filename", "audio.mp3")
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusOK))
			})
		})
	})

	Describe("DownloadFile", func() {
		Context("when downloading files", func() {
			BeforeEach(func() {
				router.GET("/file/*filepath", fileServerHandler.DownloadFile)
			})

			It("should serve file content with correct headers", func() {
				filePath := "/uploads/test-file.mp3"
				
				mockStorageService.EXPECT().
					GetFile(gomock.Any(), filePath).
					Return(testFileContent, "audio/mpeg", nil)

				req := httptest.NewRequest(http.MethodGet, "/file"+filePath, nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusOK))
				Expect(w.Header().Get("Content-Type")).To(Equal("audio/mpeg"))
				Expect(w.Header().Get("Content-Disposition")).To(ContainSubstring("filename=\"test-file.mp3\""))
				Expect(w.Body.Bytes()).To(Equal(testFileContent))
			})

			It("should return 404 for non-existent files", func() {
				filePath := "/uploads/non-existent.mp3"
				
				mockStorageService.EXPECT().
					GetFile(gomock.Any(), filePath).
					Return(nil, "", handlers.ErrFileNotFound)

				req := httptest.NewRequest(http.MethodGet, "/file"+filePath, nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusNotFound))
				
				var response map[string]interface{}
				err := testutil.ParseJSONResponse(w.Body, &response)
				Expect(err).ToNot(HaveOccurred())
				Expect(response["error"]).To(Equal("file not found"))
			})

			It("should handle storage service errors", func() {
				filePath := "/uploads/error-file.mp3"
				
				mockStorageService.EXPECT().
					GetFile(gomock.Any(), filePath).
					Return(nil, "", handlers.ErrStorageFailure)

				req := httptest.NewRequest(http.MethodGet, "/file"+filePath, nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusInternalServerError))
			})

			It("should support range requests for large files", func() {
				filePath := "/uploads/large-file.mp3"
				largeContent := make([]byte, 1024*1024) // 1MB
				
				mockStorageService.EXPECT().
					GetFileRange(gomock.Any(), filePath, int64(0), int64(1023)).
					Return(largeContent[:1024], "audio/mpeg", nil)

				req := httptest.NewRequest(http.MethodGet, "/file"+filePath, nil)
				req.Header.Set("Range", "bytes=0-1023")
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusPartialContent))
				Expect(w.Header().Get("Content-Range")).To(ContainSubstring("bytes 0-1023"))
				Expect(w.Header().Get("Accept-Ranges")).To(Equal("bytes"))
			})
		})
	})

	Describe("DeleteFile", func() {
		Context("when deleting files", func() {
			BeforeEach(func() {
				router.DELETE("/file/*filepath", fileServerHandler.DeleteFile)
			})

			It("should delete file with valid authentication", func() {
				filePath := "/uploads/test-file.mp3"
				
				mockTokenService.EXPECT().
					ValidateDeleteToken(testToken).
					Return(&models.DeleteTokenClaims{
						UserID:    "user-123",
						FilePath:  filePath,
						ExpiresAt: time.Now().Add(1 * time.Hour),
					}, nil)

				mockStorageService.EXPECT().
					DeleteFile(gomock.Any(), filePath).
					Return(nil)

				req := httptest.NewRequest(http.MethodDelete, "/file"+filePath, nil)
				req.Header.Set("Authorization", "Bearer "+testToken)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusOK))
				
				var response map[string]interface{}
				err := testutil.ParseJSONResponse(w.Body, &response)
				Expect(err).ToNot(HaveOccurred())
				Expect(response["status"]).To(Equal("deleted"))
			})

			It("should reject deletion without authorization", func() {
				filePath := "/uploads/test-file.mp3"

				req := httptest.NewRequest(http.MethodDelete, "/file"+filePath, nil)
				// Missing Authorization header
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusUnauthorized))
			})

			It("should validate file path permissions", func() {
				filePath := "/uploads/other-user-file.mp3"
				
				mockTokenService.EXPECT().
					ValidateDeleteToken(testToken).
					Return(&models.DeleteTokenClaims{
						UserID:    "user-123",
						FilePath:  "/uploads/different-file.mp3", // Token for different file
						ExpiresAt: time.Now().Add(1 * time.Hour),
					}, nil)

				req := httptest.NewRequest(http.MethodDelete, "/file"+filePath, nil)
				req.Header.Set("Authorization", "Bearer "+testToken)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusForbidden))
				
				var response map[string]interface{}
				err := testutil.ParseJSONResponse(w.Body, &response)
				Expect(err).ToNot(HaveOccurred())
				Expect(response["error"]).To(ContainSubstring("not authorized"))
			})
		})
	})

	Describe("GetStatus", func() {
		Context("when checking server status", func() {
			BeforeEach(func() {
				router.GET("/status", fileServerHandler.GetStatus)
			})

			It("should return server status and statistics", func() {
				mockStorageService.EXPECT().
					GetStorageStats(gomock.Any()).
					Return(&models.StorageStats{
						TotalFiles:      150,
						TotalSize:       1024*1024*512, // 512MB
						AvailableSpace:  1024*1024*1024*10, // 10GB
						UploadsToday:    25,
						DownloadsToday:  85,
					}, nil)

				req := httptest.NewRequest(http.MethodGet, "/status", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusOK))
				
				var response map[string]interface{}
				err := testutil.ParseJSONResponse(w.Body, &response)
				Expect(err).ToNot(HaveOccurred())
				
				Expect(response["status"]).To(Equal("healthy"))
				Expect(response["uptime"]).ToNot(BeEmpty())
				Expect(response["storage"]).ToNot(BeNil())
				
				storage := response["storage"].(map[string]interface{})
				Expect(storage["total_files"]).To(Equal(150.0))
				Expect(storage["uploads_today"]).To(Equal(25.0))
			})

			It("should handle storage stats errors gracefully", func() {
				mockStorageService.EXPECT().
					GetStorageStats(gomock.Any()).
					Return(nil, handlers.ErrStorageFailure)

				req := httptest.NewRequest(http.MethodGet, "/status", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusOK))
				
				var response map[string]interface{}
				err := testutil.ParseJSONResponse(w.Body, &response)
				Expect(err).ToNot(HaveOccurred())
				
				Expect(response["status"]).To(Equal("degraded"))
				Expect(response["storage_error"]).ToNot(BeNil())
			})
		})
	})

	Describe("ListFiles", func() {
		Context("when listing files", func() {
			BeforeEach(func() {
				router.GET("/list", fileServerHandler.ListFiles)
			})

			It("should list files with pagination", func() {
				expectedFiles := []models.FileInfo{
					{
						Name:     "track1.mp3",
						Path:     "/uploads/track1.mp3",
						Size:     1024*512, // 512KB
						Modified: time.Now().Add(-1 * time.Hour),
						Type:     "audio/mpeg",
					},
					{
						Name:     "track2.wav",
						Path:     "/uploads/track2.wav",
						Size:     1024*1024*5, // 5MB
						Modified: time.Now().Add(-2 * time.Hour),
						Type:     "audio/wav",
					},
				}
				
				mockStorageService.EXPECT().
					ListFiles(gomock.Any(), "/uploads", 50, 0).
					Return(expectedFiles, 2, nil)

				req := httptest.NewRequest(http.MethodGet, "/list?path=/uploads&limit=50&offset=0", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusOK))
				
				var response map[string]interface{}
				err := testutil.ParseJSONResponse(w.Body, &response)
				Expect(err).ToNot(HaveOccurred())
				
				files := response["files"].([]interface{})
				Expect(files).To(HaveLen(2))
				
				firstFile := files[0].(map[string]interface{})
				Expect(firstFile["name"]).To(Equal("track1.mp3"))
				Expect(firstFile["type"]).To(Equal("audio/mpeg"))
				
				Expect(response["total"]).To(Equal(2.0))
				Expect(response["offset"]).To(Equal(0.0))
				Expect(response["limit"]).To(Equal(50.0))
			})

			It("should handle invalid pagination parameters", func() {
				req := httptest.NewRequest(http.MethodGet, "/list?limit=invalid&offset=negative", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusBadRequest))
				
				var response map[string]interface{}
				err := testutil.ParseJSONResponse(w.Body, &response)
				Expect(err).ToNot(HaveOccurred())
				Expect(response["error"]).To(ContainSubstring("invalid pagination"))
			})

			It("should enforce maximum limit", func() {
				req := httptest.NewRequest(http.MethodGet, "/list?limit=1000", nil) // Too high
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusBadRequest))
				
				var response map[string]interface{}
				err := testutil.ParseJSONResponse(w.Body, &response)
				Expect(err).ToNot(HaveOccurred())
				Expect(response["error"]).To(ContainSubstring("limit too high"))
			})
		})
	})

	Describe("GenerateUploadToken", func() {
		Context("when generating upload tokens", func() {
			BeforeEach(func() {
				router.POST("/token/upload", func(c *gin.Context) {
					c.Set("firebase_uid", "user-123")
					fileServerHandler.GenerateUploadToken(c)
				})
			})

			It("should generate valid upload token for authenticated user", func() {
				tokenRequest := models.UploadTokenRequest{
					TrackID:   "track-456",
					ExpiresIn: 3600, // 1 hour
				}
				
				expectedToken := "generated-upload-token-789"
				
				mockTokenService.EXPECT().
					GenerateUploadToken("user-123", "track-456", time.Duration(3600)*time.Second).
					Return(expectedToken, nil)

				jsonPayload, _ := json.Marshal(tokenRequest)
				req := httptest.NewRequest(http.MethodPost, "/token/upload", bytes.NewBuffer(jsonPayload))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusOK))
				
				var response map[string]interface{}
				err := testutil.ParseJSONResponse(w.Body, &response)
				Expect(err).ToNot(HaveOccurred())
				
				Expect(response["token"]).To(Equal(expectedToken))
				Expect(response["expires_in"]).To(Equal(3600.0))
				Expect(response["track_id"]).To(Equal("track-456"))
			})

			It("should validate expiration time limits", func() {
				tokenRequest := models.UploadTokenRequest{
					TrackID:   "track-456",
					ExpiresIn: 86400 * 7, // 7 days - too long
				}

				jsonPayload, _ := json.Marshal(tokenRequest)
				req := httptest.NewRequest(http.MethodPost, "/token/upload", bytes.NewBuffer(jsonPayload))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusBadRequest))
				
				var response map[string]interface{}
				err := testutil.ParseJSONResponse(w.Body, &response)
				Expect(err).ToNot(HaveOccurred())
				Expect(response["error"]).To(ContainSubstring("expiration too long"))
			})

			It("should require authentication", func() {
				router = gin.New()
				router.POST("/token/upload", fileServerHandler.GenerateUploadToken)

				tokenRequest := models.UploadTokenRequest{
					TrackID:   "track-456",
					ExpiresIn: 3600,
				}

				jsonPayload, _ := json.Marshal(tokenRequest)
				req := httptest.NewRequest(http.MethodPost, "/token/upload", bytes.NewBuffer(jsonPayload))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusUnauthorized))
			})
		})
	})
})