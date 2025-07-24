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
	"github.com/wavlake/monorepo/tests/mocks"
	"github.com/wavlake/monorepo/tests/testutil"
)

var _ = Describe("AuthTokenHandler", func() {
	var (
		ctrl                *gomock.Controller
		mockTokenService    *mocks.MockTokenServiceInterface
		mockNostrService    *mocks.MockNostrTrackServiceInterface
		authTokenHandler    *handlers.AuthTokenHandler
		router              *gin.Engine
		testFirebaseUID     string
		testPubkey          string
		testTrackID         string
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockTokenService = mocks.NewMockTokenServiceInterface(ctrl)
		mockNostrService = mocks.NewMockNostrTrackServiceInterface(ctrl)
		authTokenHandler = handlers.NewAuthTokenHandler(mockTokenService, mockNostrService)
		
		gin.SetMode(gin.TestMode)
		router = gin.New()
		testFirebaseUID = testutil.TestFirebaseUID
		testPubkey = "test-pubkey-123"
		testTrackID = testutil.TestTrackID
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
				tokenRequest := models.UploadTokenRequest{
					TrackID:   testTrackID,
					ExpiresIn: 3600, // 1 hour
				}
				
				expectedToken := "upload-token-abc123"
				
				mockTokenService.EXPECT().
					GenerateUploadToken(testFirebaseUID, testTrackID, time.Duration(3600)*time.Second).
					Return(expectedToken, nil)

				jsonPayload, _ := json.Marshal(tokenRequest)
				req := httptest.NewRequest(http.MethodPost, "/auth/upload-token", bytes.NewBuffer(jsonPayload))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusOK))
				
				var response map[string]interface{}
				err := testutil.ParseJSONResponse(w.Body, &response)
				Expect(err).ToNot(HaveOccurred())
				
				Expect(response["token"]).To(Equal(expectedToken))
				Expect(response["track_id"]).To(Equal(testTrackID))
				Expect(response["expires_in"]).To(Equal(3600.0))
				Expect(response["token_type"]).To(Equal("upload"))
			})

			It("should validate expiration time limits", func() {
				tokenRequest := models.UploadTokenRequest{
					TrackID:   testTrackID,
					ExpiresIn: 86400 * 7, // 7 days - too long
				}

				jsonPayload, _ := json.Marshal(tokenRequest)
				req := httptest.NewRequest(http.MethodPost, "/auth/upload-token", bytes.NewBuffer(jsonPayload))
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
				router.POST("/auth/upload-token", authTokenHandler.GenerateUploadToken)

				tokenRequest := models.UploadTokenRequest{
					TrackID:   testTrackID,
					ExpiresIn: 3600,
				}

				jsonPayload, _ := json.Marshal(tokenRequest)
				req := httptest.NewRequest(http.MethodPost, "/auth/upload-token", bytes.NewBuffer(jsonPayload))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusUnauthorized))
			})

			It("should validate track ID format", func() {
				tokenRequest := models.UploadTokenRequest{
					TrackID:   "invalid-track-id", // Invalid format
					ExpiresIn: 3600,
				}

				jsonPayload, _ := json.Marshal(tokenRequest)
				req := httptest.NewRequest(http.MethodPost, "/auth/upload-token", bytes.NewBuffer(jsonPayload))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusBadRequest))
				
				var response map[string]interface{}
				err := testutil.ParseJSONResponse(w.Body, &response)
				Expect(err).ToNot(HaveOccurred())
				Expect(response["error"]).To(ContainSubstring("invalid track ID"))
			})

			It("should handle token generation failures", func() {
				tokenRequest := models.UploadTokenRequest{
					TrackID:   testTrackID,
					ExpiresIn: 3600,
				}
				
				mockTokenService.EXPECT().
					GenerateUploadToken(testFirebaseUID, testTrackID, time.Duration(3600)*time.Second).
					Return("", handlers.ErrTokenGenerationFailure)

				jsonPayload, _ := json.Marshal(tokenRequest)
				req := httptest.NewRequest(http.MethodPost, "/auth/upload-token", bytes.NewBuffer(jsonPayload))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusInternalServerError))
				
				var response map[string]interface{}
				err := testutil.ParseJSONResponse(w.Body, &response)
				Expect(err).ToNot(HaveOccurred())
				Expect(response["error"]).To(ContainSubstring("token generation failed"))
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

			It("should generate delete token for file owner", func() {
				deleteRequest := models.DeleteTokenRequest{
					FilePath:  "/uploads/track-123.wav",
					ExpiresIn: 300, // 5 minutes
				}
				
				expectedToken := "delete-token-xyz789"
				
				mockTokenService.EXPECT().
					GenerateDeleteToken(testFirebaseUID, deleteRequest.FilePath, time.Duration(300)*time.Second).
					Return(expectedToken, nil)

				jsonPayload, _ := json.Marshal(deleteRequest)
				req := httptest.NewRequest(http.MethodPost, "/auth/delete-token", bytes.NewBuffer(jsonPayload))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusOK))
				
				var response map[string]interface{}
				err := testutil.ParseJSONResponse(w.Body, &response)
				Expect(err).ToNot(HaveOccurred())
				
				Expect(response["token"]).To(Equal(expectedToken))
				Expect(response["file_path"]).To(Equal(deleteRequest.FilePath))
				Expect(response["expires_in"]).To(Equal(300.0))
				Expect(response["token_type"]).To(Equal("delete"))
			})

			It("should validate file path", func() {
				deleteRequest := models.DeleteTokenRequest{
					FilePath:  "../../../etc/passwd", // Path traversal attempt
					ExpiresIn: 300,
				}

				jsonPayload, _ := json.Marshal(deleteRequest)
				req := httptest.NewRequest(http.MethodPost, "/auth/delete-token", bytes.NewBuffer(jsonPayload))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusBadRequest))
				
				var response map[string]interface{}
				err := testutil.ParseJSONResponse(w.Body, &response)
				Expect(err).ToNot(HaveOccurred())
				Expect(response["error"]).To(ContainSubstring("invalid file path"))
			})

			It("should enforce shorter expiration for delete tokens", func() {
				deleteRequest := models.DeleteTokenRequest{
					FilePath:  "/uploads/track-123.wav",
					ExpiresIn: 86400, // 24 hours - too long for delete
				}

				jsonPayload, _ := json.Marshal(deleteRequest)
				req := httptest.NewRequest(http.MethodPost, "/auth/delete-token", bytes.NewBuffer(jsonPayload))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusBadRequest))
				
				var response map[string]interface{}
				err := testutil.ParseJSONResponse(w.Body, &response)
				Expect(err).ToNot(HaveOccurred())
				Expect(response["error"]).To(ContainSubstring("expiration too long"))
			})
		})
	})

	Describe("ValidateToken", func() {
		Context("when validating tokens", func() {
			BeforeEach(func() {
				router.POST("/auth/validate-token", authTokenHandler.ValidateToken)
			})

			It("should validate upload token successfully", func() {
				validationRequest := models.TokenValidationRequest{
					Token:     "upload-token-abc123",
					TokenType: "upload",
				}
				
				expectedClaims := &models.UploadTokenClaims{
					UserID:    testFirebaseUID,
					TrackID:   testTrackID,
					ExpiresAt: time.Now().Add(1 * time.Hour),
				}
				
				mockTokenService.EXPECT().
					ValidateUploadToken(validationRequest.Token).
					Return(expectedClaims, nil)

				jsonPayload, _ := json.Marshal(validationRequest)
				req := httptest.NewRequest(http.MethodPost, "/auth/validate-token", bytes.NewBuffer(jsonPayload))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusOK))
				
				var response map[string]interface{}
				err := testutil.ParseJSONResponse(w.Body, &response)
				Expect(err).ToNot(HaveOccurred())
				
				Expect(response["valid"]).To(Equal(true))
				Expect(response["token_type"]).To(Equal("upload"))
				Expect(response["user_id"]).To(Equal(testFirebaseUID))
				Expect(response["track_id"]).To(Equal(testTrackID))
			})

			It("should validate delete token successfully", func() {
				validationRequest := models.TokenValidationRequest{
					Token:     "delete-token-xyz789",
					TokenType: "delete",
				}
				
				expectedClaims := &models.DeleteTokenClaims{
					UserID:    testFirebaseUID,
					FilePath:  "/uploads/track-123.wav",
					ExpiresAt: time.Now().Add(5 * time.Minute),
				}
				
				mockTokenService.EXPECT().
					ValidateDeleteToken(validationRequest.Token).
					Return(expectedClaims, nil)

				jsonPayload, _ := json.Marshal(validationRequest)
				req := httptest.NewRequest(http.MethodPost, "/auth/validate-token", bytes.NewBuffer(jsonPayload))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusOK))
				
				var response map[string]interface{}
				err := testutil.ParseJSONResponse(w.Body, &response)
				Expect(err).ToNot(HaveOccurred())
				
				Expect(response["valid"]).To(Equal(true))
				Expect(response["token_type"]).To(Equal("delete"))
				Expect(response["user_id"]).To(Equal(testFirebaseUID))
				Expect(response["file_path"]).To(Equal("/uploads/track-123.wav"))
			})

			It("should reject invalid tokens", func() {
				validationRequest := models.TokenValidationRequest{
					Token:     "invalid-token",
					TokenType: "upload",
				}
				
				mockTokenService.EXPECT().
					ValidateUploadToken(validationRequest.Token).
					Return(nil, handlers.ErrInvalidToken)

				jsonPayload, _ := json.Marshal(validationRequest)
				req := httptest.NewRequest(http.MethodPost, "/auth/validate-token", bytes.NewBuffer(jsonPayload))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusOK))
				
				var response map[string]interface{}
				err := testutil.ParseJSONResponse(w.Body, &response)
				Expect(err).ToNot(HaveOccurred())
				
				Expect(response["valid"]).To(Equal(false))
				Expect(response["error"]).To(ContainSubstring("invalid token"))
			})

			It("should reject expired tokens", func() {
				validationRequest := models.TokenValidationRequest{
					Token:     "expired-token-abc123",
					TokenType: "upload",
				}
				
				expiredClaims := &models.UploadTokenClaims{
					UserID:    testFirebaseUID,
					TrackID:   testTrackID,
					ExpiresAt: time.Now().Add(-1 * time.Hour), // Expired
				}
				
				mockTokenService.EXPECT().
					ValidateUploadToken(validationRequest.Token).
					Return(expiredClaims, nil)

				jsonPayload, _ := json.Marshal(validationRequest)
				req := httptest.NewRequest(http.MethodPost, "/auth/validate-token", bytes.NewBuffer(jsonPayload))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusOK))
				
				var response map[string]interface{}
				err := testutil.ParseJSONResponse(w.Body, &response)
				Expect(err).ToNot(HaveOccurred())
				
				Expect(response["valid"]).To(Equal(false))
				Expect(response["error"]).To(ContainSubstring("token expired"))
			})

			It("should reject unknown token types", func() {
				validationRequest := models.TokenValidationRequest{
					Token:     "some-token",
					TokenType: "unknown",
				}

				jsonPayload, _ := json.Marshal(validationRequest)
				req := httptest.NewRequest(http.MethodPost, "/auth/validate-token", bytes.NewBuffer(jsonPayload))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusBadRequest))
				
				var response map[string]interface{}
				err := testutil.ParseJSONResponse(w.Body, &response)
				Expect(err).ToNot(HaveOccurred())
				Expect(response["error"]).To(ContainSubstring("unsupported token type"))
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

			It("should revoke upload token", func() {
				revokeRequest := models.TokenRevokeRequest{
					Token:     "upload-token-abc123",
					TokenType: "upload",
				}
				
				mockTokenService.EXPECT().
					RevokeUploadToken(revokeRequest.Token, testFirebaseUID).
					Return(nil)

				jsonPayload, _ := json.Marshal(revokeRequest)
				req := httptest.NewRequest(http.MethodPost, "/auth/revoke-token", bytes.NewBuffer(jsonPayload))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusOK))
				
				var response map[string]interface{}
				err := testutil.ParseJSONResponse(w.Body, &response)
				Expect(err).ToNot(HaveOccurred())
				
				Expect(response["status"]).To(Equal("revoked"))
				Expect(response["token_type"]).To(Equal("upload"))
			})

			It("should revoke delete token", func() {
				revokeRequest := models.TokenRevokeRequest{
					Token:     "delete-token-xyz789",
					TokenType: "delete",
				}
				
				mockTokenService.EXPECT().
					RevokeDeleteToken(revokeRequest.Token, testFirebaseUID).
					Return(nil)

				jsonPayload, _ := json.Marshal(revokeRequest)
				req := httptest.NewRequest(http.MethodPost, "/auth/revoke-token", bytes.NewBuffer(jsonPayload))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusOK))
				
				var response map[string]interface{}
				err := testutil.ParseJSONResponse(w.Body, &response)
				Expect(err).ToNot(HaveOccurred())
				
				Expect(response["status"]).To(Equal("revoked"))
				Expect(response["token_type"]).To(Equal("delete"))
			})

			It("should reject revocation from non-owner", func() {
				revokeRequest := models.TokenRevokeRequest{
					Token:     "upload-token-abc123",
					TokenType: "upload",
				}
				
				mockTokenService.EXPECT().
					RevokeUploadToken(revokeRequest.Token, testFirebaseUID).
					Return(handlers.ErrUnauthorizedRevocation)

				jsonPayload, _ := json.Marshal(revokeRequest)
				req := httptest.NewRequest(http.MethodPost, "/auth/revoke-token", bytes.NewBuffer(jsonPayload))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusForbidden))
				
				var response map[string]interface{}
				err := testutil.ParseJSONResponse(w.Body, &response)
				Expect(err).ToNot(HaveOccurred())
				Expect(response["error"]).To(ContainSubstring("not authorized"))
			})

			It("should handle non-existent tokens gracefully", func() {
				revokeRequest := models.TokenRevokeRequest{
					Token:     "non-existent-token",
					TokenType: "upload",
				}
				
				mockTokenService.EXPECT().
					RevokeUploadToken(revokeRequest.Token, testFirebaseUID).
					Return(handlers.ErrTokenNotFound)

				jsonPayload, _ := json.Marshal(revokeRequest)
				req := httptest.NewRequest(http.MethodPost, "/auth/revoke-token", bytes.NewBuffer(jsonPayload))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusNotFound))
				
				var response map[string]interface{}
				err := testutil.ParseJSONResponse(w.Body, &response)
				Expect(err).ToNot(HaveOccurred())
				Expect(response["error"]).To(ContainSubstring("token not found"))
			})
		})
	})

	Describe("ListActiveTokens", func() {
		Context("when listing user's active tokens", func() {
			BeforeEach(func() {
				router.GET("/auth/tokens", func(c *gin.Context) {
					c.Set("firebase_uid", testFirebaseUID)
					authTokenHandler.ListActiveTokens(c)
				})
			})

			It("should list all active tokens for user", func() {
				expectedTokens := []models.TokenInfo{
					{
						ID:        "token-1",
						Type:      "upload",
						TrackID:   testTrackID,
						CreatedAt: time.Now().Add(-30 * time.Minute),
						ExpiresAt: time.Now().Add(30 * time.Minute),
						LastUsed:  time.Now().Add(-10 * time.Minute),
					},
					{
						ID:        "token-2",
						Type:      "delete",
						FilePath:  "/uploads/track-123.wav",
						CreatedAt: time.Now().Add(-5 * time.Minute),
						ExpiresAt: time.Now().Add(5 * time.Minute),
					},
				}
				
				mockTokenService.EXPECT().
					ListActiveTokens(testFirebaseUID).
					Return(expectedTokens, nil)

				req := httptest.NewRequest(http.MethodGet, "/auth/tokens", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusOK))
				
				var response map[string]interface{}
				err := testutil.ParseJSONResponse(w.Body, &response)
				Expect(err).ToNot(HaveOccurred())
				
				tokens := response["tokens"].([]interface{})
				Expect(tokens).To(HaveLen(2))
				
				uploadToken := tokens[0].(map[string]interface{})
				Expect(uploadToken["type"]).To(Equal("upload"))
				Expect(uploadToken["track_id"]).To(Equal(testTrackID))
				
				deleteToken := tokens[1].(map[string]interface{})
				Expect(deleteToken["type"]).To(Equal("delete"))
				Expect(deleteToken["file_path"]).To(Equal("/uploads/track-123.wav"))
			})

			It("should support token type filtering", func() {
				uploadTokens := []models.TokenInfo{
					{
						ID:      "token-1",
						Type:    "upload",
						TrackID: testTrackID,
					},
				}
				
				mockTokenService.EXPECT().
					ListActiveTokensByType(testFirebaseUID, "upload").
					Return(uploadTokens, nil)

				req := httptest.NewRequest(http.MethodGet, "/auth/tokens?type=upload", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusOK))
				
				var response map[string]interface{}
				err := testutil.ParseJSONResponse(w.Body, &response)
				Expect(err).ToNot(HaveOccurred())
				
				tokens := response["tokens"].([]interface{})
				Expect(tokens).To(HaveLen(1))
			})

			It("should return empty list when no active tokens", func() {
				mockTokenService.EXPECT().
					ListActiveTokens(testFirebaseUID).
					Return([]models.TokenInfo{}, nil)

				req := httptest.NewRequest(http.MethodGet, "/auth/tokens", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusOK))
				
				var response map[string]interface{}
				err := testutil.ParseJSONResponse(w.Body, &response)
				Expect(err).ToNot(HaveOccurred())
				
				tokens := response["tokens"].([]interface{})
				Expect(tokens).To(BeEmpty())
			})

			It("should handle service errors gracefully", func() {
				mockTokenService.EXPECT().
					ListActiveTokens(testFirebaseUID).
					Return(nil, handlers.ErrTokenServiceFailure)

				req := httptest.NewRequest(http.MethodGet, "/auth/tokens", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusInternalServerError))
				
				var response map[string]interface{}
				err := testutil.ParseJSONResponse(w.Body, &response)
				Expect(err).ToNot(HaveOccurred())
				Expect(response["error"]).To(ContainSubstring("token service error"))
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

			It("should refresh upload token with new expiration", func() {
				refreshRequest := models.TokenRefreshRequest{
					Token:     "upload-token-abc123",
					TokenType: "upload",
					ExpiresIn: 7200, // 2 hours
				}
				
				newToken := "upload-token-def456"
				
				mockTokenService.EXPECT().
					RefreshUploadToken(refreshRequest.Token, testFirebaseUID, time.Duration(7200)*time.Second).
					Return(newToken, nil)

				jsonPayload, _ := json.Marshal(refreshRequest)
				req := httptest.NewRequest(http.MethodPost, "/auth/refresh-token", bytes.NewBuffer(jsonPayload))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusOK))
				
				var response map[string]interface{}
				err := testutil.ParseJSONResponse(w.Body, &response)
				Expect(err).ToNot(HaveOccurred())
				
				Expect(response["token"]).To(Equal(newToken))
				Expect(response["token_type"]).To(Equal("upload"))
				Expect(response["expires_in"]).To(Equal(7200.0))
			})

			It("should reject refresh for expired tokens", func() {
				refreshRequest := models.TokenRefreshRequest{
					Token:     "expired-token",
					TokenType: "upload",
					ExpiresIn: 3600,
				}
				
				mockTokenService.EXPECT().
					RefreshUploadToken(refreshRequest.Token, testFirebaseUID, time.Duration(3600)*time.Second).
					Return("", handlers.ErrTokenExpired)

				jsonPayload, _ := json.Marshal(refreshRequest)
				req := httptest.NewRequest(http.MethodPost, "/auth/refresh-token", bytes.NewBuffer(jsonPayload))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusUnauthorized))
				
				var response map[string]interface{}
				err := testutil.ParseJSONResponse(w.Body, &response)
				Expect(err).ToNot(HaveOccurred())
				Expect(response["error"]).To(ContainSubstring("token expired"))
			})

			It("should validate refresh parameters", func() {
				refreshRequest := models.TokenRefreshRequest{
					Token:     "upload-token-abc123",
					TokenType: "upload",
					ExpiresIn: 86400 * 7, // Too long
				}

				jsonPayload, _ := json.Marshal(refreshRequest)
				req := httptest.NewRequest(http.MethodPost, "/auth/refresh-token", bytes.NewBuffer(jsonPayload))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusBadRequest))
				
				var response map[string]interface{}
				err := testutil.ParseJSONResponse(w.Body, &response)
				Expect(err).ToNot(HaveOccurred())
				Expect(response["error"]).To(ContainSubstring("expiration too long"))
			})
		})
	})
})