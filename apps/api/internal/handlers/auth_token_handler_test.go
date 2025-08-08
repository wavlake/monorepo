package handlers_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
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

var _ = Describe("AuthTokenHandler", func() {
	var (
		ctrl              *gomock.Controller
		mockTokenService  *mocks.MockTokenServiceInterface
		authTokenHandler  *handlers.AuthTokenHandler
		router            *gin.Engine
		testFirebaseUID   string
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockTokenService = mocks.NewMockTokenServiceInterface(ctrl)
		authTokenHandler = handlers.NewAuthTokenHandler(mockTokenService)
		
		gin.SetMode(gin.TestMode)
		router = gin.New()
		testFirebaseUID = testutil.TestFirebaseUID
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Describe("GenerateUploadToken", func() {
		Context("when generating upload tokens", func() {
			BeforeEach(func() {
				router.POST("/auth/upload-token", func(c *gin.Context) {
					c.Set("firebase_uid", testFirebaseUID)
					authTokenHandler.GenerateUploadToken(c)
				})
			})

			It("should generate upload token for authenticated user", func() {
				request := map[string]interface{}{
					"path":       "/uploads/test-file.wav",
					"expiration": 60, // 1 hour in minutes
				}

				expectedToken := &models.FileUploadToken{
					Token:     "upload-token-abc123",
					Path:      "/uploads/test-file.wav",
					UserID:    testFirebaseUID,
					ExpiresAt: time.Now().Add(1 * time.Hour),
					CreatedAt: time.Now(),
				}

				mockTokenService.EXPECT().
					GenerateUploadToken(gomock.Any(), "/uploads/test-file.wav", testFirebaseUID, 1*time.Hour).
					Return(expectedToken, nil)

				jsonPayload, _ := json.Marshal(request)
				req := httptest.NewRequest(http.MethodPost, "/auth/upload-token", bytes.NewBuffer(jsonPayload))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusOK))
				
				var response handlers.TokenResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				Expect(err).ToNot(HaveOccurred())
				
				Expect(response.Success).To(BeTrue())
				Expect(response.Data).ToNot(BeNil())
				Expect(response.Data.Token).To(Equal("upload-token-abc123"))
			})

			It("should use default expiration when not provided", func() {
				request := map[string]interface{}{
					"path": "/uploads/test-file.wav",
				}

				expectedToken := &models.FileUploadToken{
					Token:     "upload-token-default",
					Path:      "/uploads/test-file.wav",
					UserID:    testFirebaseUID,
					ExpiresAt: time.Now().Add(1 * time.Hour),
					CreatedAt: time.Now(),
				}

				mockTokenService.EXPECT().
					GenerateUploadToken(gomock.Any(), "/uploads/test-file.wav", testFirebaseUID, 1*time.Hour).
					Return(expectedToken, nil)

				jsonPayload, _ := json.Marshal(request)
				req := httptest.NewRequest(http.MethodPost, "/auth/upload-token", bytes.NewBuffer(jsonPayload))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusOK))
			})

			It("should require authentication", func() {
				router = gin.New()
				router.POST("/auth/upload-token", authTokenHandler.GenerateUploadToken)

				request := map[string]interface{}{
					"path": "/uploads/test-file.wav",
				}

				jsonPayload, _ := json.Marshal(request)
				req := httptest.NewRequest(http.MethodPost, "/auth/upload-token", bytes.NewBuffer(jsonPayload))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusUnauthorized))
			})

			It("should validate request parameters", func() {
				request := map[string]interface{}{
					"invalid": "request",
				}

				jsonPayload, _ := json.Marshal(request)
				req := httptest.NewRequest(http.MethodPost, "/auth/upload-token", bytes.NewBuffer(jsonPayload))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusBadRequest))
			})

			It("should handle token generation failures", func() {
				request := map[string]interface{}{
					"path": "/uploads/test-file.wav",
				}

				mockTokenService.EXPECT().
					GenerateUploadToken(gomock.Any(), "/uploads/test-file.wav", testFirebaseUID, gomock.Any()).
					Return(nil, errors.New("token generation failed"))

				jsonPayload, _ := json.Marshal(request)
				req := httptest.NewRequest(http.MethodPost, "/auth/upload-token", bytes.NewBuffer(jsonPayload))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusInternalServerError))
			})
		})
	})

	Describe("GenerateDeleteToken", func() {
		Context("when generating delete tokens", func() {
			BeforeEach(func() {
				router.POST("/auth/delete-token", func(c *gin.Context) {
					c.Set("firebase_uid", testFirebaseUID)
					authTokenHandler.GenerateDeleteToken(c)
				})
			})

			It("should generate delete token for authenticated user", func() {
				request := map[string]interface{}{
					"path":       "/uploads/test-file.wav",
					"expiration": 15, // 15 minutes
				}

				expectedToken := &models.FileUploadToken{
					Token:     "delete-token-xyz789",
					Path:      "/uploads/test-file.wav",
					UserID:    testFirebaseUID,
					ExpiresAt: time.Now().Add(15 * time.Minute),
					CreatedAt: time.Now(),
				}

				mockTokenService.EXPECT().
					GenerateDeleteToken(gomock.Any(), "/uploads/test-file.wav", testFirebaseUID, 15*time.Minute).
					Return(expectedToken, nil)

				jsonPayload, _ := json.Marshal(request)
				req := httptest.NewRequest(http.MethodPost, "/auth/delete-token", bytes.NewBuffer(jsonPayload))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusOK))
				
				var response handlers.TokenResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				Expect(err).ToNot(HaveOccurred())
				
				Expect(response.Success).To(BeTrue())
				Expect(response.Data).ToNot(BeNil())
				Expect(response.Data.Token).To(Equal("delete-token-xyz789"))
			})
		})
	})

	Describe("ValidateToken", func() {
		Context("when validating tokens", func() {
			BeforeEach(func() {
				router.POST("/auth/validate-token", authTokenHandler.ValidateToken)
			})

			It("should validate token successfully", func() {
				request := map[string]interface{}{
					"token": "valid-token-123",
					"path":  "/uploads/test-file.wav",
				}

				expectedToken := &models.FileUploadToken{
					Token:     "valid-token-123",
					Path:      "/uploads/test-file.wav",
					UserID:    testFirebaseUID,
					ExpiresAt: time.Now().Add(1 * time.Hour),
					CreatedAt: time.Now(),
				}

				mockTokenService.EXPECT().
					ValidateToken(gomock.Any(), "valid-token-123", "/uploads/test-file.wav").
					Return(expectedToken, nil)

				jsonPayload, _ := json.Marshal(request)
				req := httptest.NewRequest(http.MethodPost, "/auth/validate-token", bytes.NewBuffer(jsonPayload))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusOK))
				
				var response handlers.TokenResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				Expect(err).ToNot(HaveOccurred())
				
				Expect(response.Success).To(BeTrue())
				Expect(response.Data).ToNot(BeNil())
			})

			It("should reject invalid tokens", func() {
				request := map[string]interface{}{
					"token": "invalid-token",
					"path":  "/uploads/test-file.wav",
				}

				mockTokenService.EXPECT().
					ValidateToken(gomock.Any(), "invalid-token", "/uploads/test-file.wav").
					Return(nil, errors.New("invalid token"))

				jsonPayload, _ := json.Marshal(request)
				req := httptest.NewRequest(http.MethodPost, "/auth/validate-token", bytes.NewBuffer(jsonPayload))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusUnauthorized))
			})
		})
	})

	Describe("RevokeToken", func() {
		Context("when revoking tokens", func() {
			BeforeEach(func() {
				router.POST("/auth/revoke-token", func(c *gin.Context) {
					c.Set("firebase_uid", testFirebaseUID)
					authTokenHandler.RevokeToken(c)
				})
			})

			It("should revoke token successfully", func() {
				request := map[string]interface{}{
					"token": "token-to-revoke",
				}

				mockTokenService.EXPECT().
					RevokeToken(gomock.Any(), "token-to-revoke").
					Return(nil)

				jsonPayload, _ := json.Marshal(request)
				req := httptest.NewRequest(http.MethodPost, "/auth/revoke-token", bytes.NewBuffer(jsonPayload))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusOK))
				
				var response handlers.RevokeTokenResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				Expect(err).ToNot(HaveOccurred())
				
				Expect(response.Success).To(BeTrue())
				Expect(response.Message).To(Equal("token revoked successfully"))
			})

			It("should handle revocation failures", func() {
				request := map[string]interface{}{
					"token": "token-to-revoke",
				}

				mockTokenService.EXPECT().
					RevokeToken(gomock.Any(), "token-to-revoke").
					Return(errors.New("revocation failed"))

				jsonPayload, _ := json.Marshal(request)
				req := httptest.NewRequest(http.MethodPost, "/auth/revoke-token", bytes.NewBuffer(jsonPayload))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusInternalServerError))
			})
		})
	})

	Describe("ListActiveTokens", func() {
		Context("when listing active tokens", func() {
			BeforeEach(func() {
				router.GET("/auth/tokens", func(c *gin.Context) {
					c.Set("firebase_uid", testFirebaseUID)
					authTokenHandler.ListActiveTokens(c)
				})
			})

			It("should list all active tokens for user", func() {
				expectedTokens := []models.FileUploadToken{
					{
						Token:     "token-1",
						Path:      "/uploads/file1.wav",
						UserID:    testFirebaseUID,
						ExpiresAt: time.Now().Add(30 * time.Minute),
						CreatedAt: time.Now(),
					},
					{
						Token:     "token-2", 
						Path:      "/uploads/file2.wav",
						UserID:    testFirebaseUID,
						ExpiresAt: time.Now().Add(15 * time.Minute),
						CreatedAt: time.Now(),
					},
				}

				mockTokenService.EXPECT().
					ListActiveTokens(gomock.Any(), testFirebaseUID).
					Return(expectedTokens, nil)

				req := httptest.NewRequest(http.MethodGet, "/auth/tokens", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusOK))
				
				var response handlers.ListActiveTokensResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				Expect(err).ToNot(HaveOccurred())
				
				Expect(response.Success).To(BeTrue())
				Expect(response.Data).To(HaveLen(2))
			})

			It("should return empty list when no active tokens", func() {
				mockTokenService.EXPECT().
					ListActiveTokens(gomock.Any(), testFirebaseUID).
					Return([]models.FileUploadToken{}, nil)

				req := httptest.NewRequest(http.MethodGet, "/auth/tokens", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusOK))
				
				var response handlers.ListActiveTokensResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				Expect(err).ToNot(HaveOccurred())
				
				Expect(response.Success).To(BeTrue())
				Expect(response.Data).To(BeEmpty())
			})

			It("should handle service errors", func() {
				mockTokenService.EXPECT().
					ListActiveTokens(gomock.Any(), testFirebaseUID).
					Return(nil, errors.New("service error"))

				req := httptest.NewRequest(http.MethodGet, "/auth/tokens", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusInternalServerError))
			})

			It("should require authentication", func() {
				router = gin.New()
				router.GET("/auth/tokens", authTokenHandler.ListActiveTokens)

				req := httptest.NewRequest(http.MethodGet, "/auth/tokens", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusUnauthorized))
			})
		})
	})

	Describe("RefreshToken", func() {
		Context("when refreshing tokens", func() {
			BeforeEach(func() {
				router.POST("/auth/refresh-token", func(c *gin.Context) {
					c.Set("firebase_uid", testFirebaseUID)
					authTokenHandler.RefreshToken(c)
				})
			})

			It("should refresh token with new expiration", func() {
				request := map[string]interface{}{
					"token":      "token-to-refresh",
					"expiration": 120, // 2 hours in minutes
				}

				refreshedToken := &models.FileUploadToken{
					Token:     "refreshed-token-def456",
					Path:      "/uploads/test-file.wav",
					UserID:    testFirebaseUID,
					ExpiresAt: time.Now().Add(2 * time.Hour),
					CreatedAt: time.Now(),
				}

				mockTokenService.EXPECT().
					RefreshToken(gomock.Any(), "token-to-refresh", 2*time.Hour).
					Return(refreshedToken, nil)

				jsonPayload, _ := json.Marshal(request)
				req := httptest.NewRequest(http.MethodPost, "/auth/refresh-token", bytes.NewBuffer(jsonPayload))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusOK))
				
				var response handlers.TokenResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				Expect(err).ToNot(HaveOccurred())
				
				Expect(response.Success).To(BeTrue())
				Expect(response.Data).ToNot(BeNil())
				Expect(response.Data.Token).To(Equal("refreshed-token-def456"))
			})

			It("should use default expiration when not provided", func() {
				request := map[string]interface{}{
					"token": "token-to-refresh",
				}

				refreshedToken := &models.FileUploadToken{
					Token:     "refreshed-token-default",
					UserID:    testFirebaseUID,
					ExpiresAt: time.Now().Add(1 * time.Hour),
					CreatedAt: time.Now(),
				}

				mockTokenService.EXPECT().
					RefreshToken(gomock.Any(), "token-to-refresh", 1*time.Hour).
					Return(refreshedToken, nil)

				jsonPayload, _ := json.Marshal(request)
				req := httptest.NewRequest(http.MethodPost, "/auth/refresh-token", bytes.NewBuffer(jsonPayload))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusOK))
			})

			It("should handle refresh failures", func() {
				request := map[string]interface{}{
					"token": "invalid-token",
				}

				mockTokenService.EXPECT().
					RefreshToken(gomock.Any(), "invalid-token", gomock.Any()).
					Return(nil, errors.New("refresh failed"))

				jsonPayload, _ := json.Marshal(request)
				req := httptest.NewRequest(http.MethodPost, "/auth/refresh-token", bytes.NewBuffer(jsonPayload))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusUnauthorized))
			})
		})
	})
})