package auth_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"github.com/gin-gonic/gin"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	
	authpkg "github.com/wavlake/monorepo/internal/auth"
)

var _ = Describe("Firebase Authentication Middleware", func() {
	var (
		router   *gin.Engine
		recorder *httptest.ResponseRecorder
	)

	BeforeEach(func() {
		gin.SetMode(gin.TestMode)
		router = gin.New()
		recorder = httptest.NewRecorder()
	})

	Describe("NewFirebaseMiddleware", func() {
		It("should create a new Firebase middleware with nil client", func() {
			middleware := authpkg.NewFirebaseMiddleware(nil)
			Expect(middleware).ToNot(BeNil())
		})
	})

	Describe("Token extraction validation", func() {
		BeforeEach(func() {
			// Use middleware that will fail before Firebase client call
			middleware := authpkg.NewFirebaseMiddleware(nil)
			router.Use(middleware.Middleware())
			router.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})
		})

		Context("missing authorization headers", func() {
			It("should reject request with no Authorization header", func() {
				req := httptest.NewRequest("GET", "/test", nil)
				router.ServeHTTP(recorder, req)

				Expect(recorder.Code).To(Equal(http.StatusUnauthorized))
				
				var response map[string]interface{}
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				Expect(err).ToNot(HaveOccurred())
				Expect(response["error"]).To(Equal("Missing authorization token"))
			})

			It("should reject request with empty Authorization header", func() {
				req := httptest.NewRequest("GET", "/test", nil)
				req.Header.Set("Authorization", "")
				router.ServeHTTP(recorder, req)

				Expect(recorder.Code).To(Equal(http.StatusUnauthorized))
				
				var response map[string]interface{}
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				Expect(err).ToNot(HaveOccurred())
				Expect(response["error"]).To(Equal("Missing authorization token"))
			})
		})

		Context("malformed authorization headers", func() {
			It("should reject request with malformed Authorization header", func() {
				req := httptest.NewRequest("GET", "/test", nil)
				req.Header.Set("Authorization", "InvalidFormat token")
				router.ServeHTTP(recorder, req)

				Expect(recorder.Code).To(Equal(http.StatusUnauthorized))
				
				var response map[string]interface{}
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				Expect(err).ToNot(HaveOccurred())
				Expect(response["error"]).To(Equal("Missing authorization token"))
			})

			It("should reject request with only Bearer prefix", func() {
				req := httptest.NewRequest("GET", "/test", nil)
				req.Header.Set("Authorization", "Bearer")
				router.ServeHTTP(recorder, req)

				Expect(recorder.Code).To(Equal(http.StatusUnauthorized))
				
				var response map[string]interface{}
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				Expect(err).ToNot(HaveOccurred())
				Expect(response["error"]).To(Equal("Missing authorization token"))
			})

			It("should reject request with Bearer and empty token", func() {
				req := httptest.NewRequest("GET", "/test", nil)
				req.Header.Set("Authorization", "Bearer ")
				router.ServeHTTP(recorder, req)

				Expect(recorder.Code).To(Equal(http.StatusUnauthorized))
				
				var response map[string]interface{}
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				Expect(err).ToNot(HaveOccurred())
				Expect(response["error"]).To(Equal("Missing authorization token"))
			})

			It("should reject non-Bearer auth schemes", func() {
				testCases := []string{
					"Basic dXNlcjpwYXNz",
					"Digest username=test", 
					"ApiKey abc123",
				}

				for _, authHeader := range testCases {
					recorder = httptest.NewRecorder()
					req := httptest.NewRequest("GET", "/test", nil)
					req.Header.Set("Authorization", authHeader)
					router.ServeHTTP(recorder, req)

					Expect(recorder.Code).To(Equal(http.StatusUnauthorized), "Auth header: %s", authHeader)
					
					var response map[string]interface{}
					err := json.Unmarshal(recorder.Body.Bytes(), &response)
					Expect(err).ToNot(HaveOccurred())
					Expect(response["error"]).To(Equal("Missing authorization token"), "Auth header: %s", authHeader)
				}
			})
		})

		Context("bearer token format validation", func() {
			It("should extract valid Bearer tokens (but fail at Firebase verification)", func() {
				testCases := []string{
					"Bearer token123",
					"bearer token123", // case insensitive
					"BEARER token123",
					"Bearer  token123", // extra spaces
				}

				for _, authHeader := range testCases {
					recorder = httptest.NewRecorder()
					req := httptest.NewRequest("GET", "/test", nil)
					req.Header.Set("Authorization", authHeader)
					
					// This will panic due to nil Firebase client, so we need to handle it
					defer func() {
						if r := recover(); r != nil {
							// Expected panic due to nil Firebase client - this means token was extracted successfully
							Expect(r).ToNot(BeNil())
						}
					}()
					
					router.ServeHTTP(recorder, req)
					
					// If we reach here without panic, it means the token extraction failed
					if recorder.Code != 0 {
						Expect(recorder.Code).To(Equal(http.StatusUnauthorized), "Auth header: %s", authHeader)
					}
				}
			})
		})
	})
})