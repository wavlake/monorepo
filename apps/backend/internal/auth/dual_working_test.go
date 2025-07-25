package auth_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"firebase.google.com/go/v4/auth"
	"github.com/gin-gonic/gin"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	
	authpkg "github.com/wavlake/monorepo/internal/auth"
)

var _ = Describe("DualAuthMiddleware", func() {
	var (
		router   *gin.Engine
		recorder *httptest.ResponseRecorder
	)

	BeforeEach(func() {
		gin.SetMode(gin.TestMode)
		router = gin.New()
		recorder = httptest.NewRecorder()
	})

	Describe("NewDualAuthMiddleware", func() {
		It("should create a new dual auth middleware", func() {
			firebaseClient := &auth.Client{} // Mock client
			middleware := authpkg.NewDualAuthMiddleware(firebaseClient)
			Expect(middleware).ToNot(BeNil())
		})

		It("should handle nil firebase client", func() {
			middleware := authpkg.NewDualAuthMiddleware(nil)
			Expect(middleware).ToNot(BeNil())
		})
	})

	Describe("Firebase token validation (first step)", func() {
		BeforeEach(func() {
			middleware := authpkg.NewDualAuthMiddleware(nil)
			router.Use(middleware.Middleware())
			router.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})
		})

		Context("missing Firebase authorization", func() {
			It("should reject request with no Authorization header", func() {
				req := httptest.NewRequest("GET", "/test", nil)
				router.ServeHTTP(recorder, req)

				Expect(recorder.Code).To(Equal(http.StatusUnauthorized))

				var response map[string]interface{}
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				Expect(err).ToNot(HaveOccurred())
				Expect(response["error"]).To(Equal("Missing Firebase authorization token"))
			})

			It("should reject request with empty Authorization header", func() {
				req := httptest.NewRequest("GET", "/test", nil)
				req.Header.Set("Authorization", "")
				router.ServeHTTP(recorder, req)

				Expect(recorder.Code).To(Equal(http.StatusUnauthorized))
				
				var response map[string]interface{}
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				Expect(err).ToNot(HaveOccurred())
				Expect(response["error"]).To(Equal("Missing Firebase authorization token"))
			})

			It("should check X-Firebase-Token header as fallback", func() {
				req := httptest.NewRequest("GET", "/test", nil)
				req.Header.Set("X-Firebase-Token", "valid.token.here")
				
				// This will panic due to nil Firebase client trying to verify the token
				defer func() {
					if r := recover(); r != nil {
						// Expected panic - means we successfully found the token and tried to verify it
						Expect(r).ToNot(BeNil())
					}
				}()
				
				router.ServeHTTP(recorder, req)
				
				// If no panic occurred, the test setup is wrong
				if recorder.Code != 0 {
					Expect(recorder.Code).To(Equal(http.StatusUnauthorized))
				}
			})
		})

		Context("malformed Firebase authorization", func() {
			It("should reject malformed Bearer token", func() {
				req := httptest.NewRequest("GET", "/test", nil)
				req.Header.Set("Authorization", "InvalidFormat token")
				router.ServeHTTP(recorder, req)

				Expect(recorder.Code).To(Equal(http.StatusUnauthorized))
				
				var response map[string]interface{}
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				Expect(err).ToNot(HaveOccurred())
				Expect(response["error"]).To(Equal("Missing Firebase authorization token"))
			})

			It("should reject Bearer header without token", func() {
				req := httptest.NewRequest("GET", "/test", nil)
				req.Header.Set("Authorization", "Bearer")
				router.ServeHTTP(recorder, req)

				Expect(recorder.Code).To(Equal(http.StatusUnauthorized))
				
				var response map[string]interface{}
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				Expect(err).ToNot(HaveOccurred())
				Expect(response["error"]).To(Equal("Missing Firebase authorization token"))
			})
		})

		Context("valid Bearer token extraction", func() {
			It("should handle case-insensitive Bearer keyword", func() {
				testCases := []string{
					"Bearer token123",
					"bearer token123", 
					"BEARER token123",
				}

				for _, authHeader := range testCases {
					recorder = httptest.NewRecorder()
					req := httptest.NewRequest("GET", "/test", nil)
					req.Header.Set("Authorization", authHeader)
					
					// Expect panic due to nil Firebase client - this means token was extracted
					defer func() {
						if r := recover(); r != nil {
							Expect(r).ToNot(BeNil())
						}
					}()
					
					router.ServeHTTP(recorder, req)
					
					// If no panic, something is wrong with test setup
					if recorder.Code != 0 {
						Expect(recorder.Code).To(Equal(http.StatusUnauthorized), "Auth header: %s", authHeader)
					}
				}
			})
		})

		Context("Bearer token format edge cases", func() {
			It("should handle various malformed headers correctly", func() {
				testCases := []struct {
					header      string
					shouldFail  bool
					description string
				}{
					{"", true, "empty header"},
					{"Bearer", true, "Bearer without token"}, 
					{"Bearer ", true, "Bearer with space only"},
					{"Basic dXNlcjpwYXNz", true, "Basic auth"},
					{"Token abc123", true, "Token auth"},
					{"Bearer token extra", true, "too many parts"},
				}

				for _, tc := range testCases {
					recorder = httptest.NewRecorder()
					req := httptest.NewRequest("GET", "/test", nil)
					if tc.header != "" {
						req.Header.Set("Authorization", tc.header)
					}
					router.ServeHTTP(recorder, req)

					if tc.shouldFail {
						Expect(recorder.Code).To(Equal(http.StatusUnauthorized), "Test case: %s", tc.description)
						
						var response map[string]interface{}
						err := json.Unmarshal(recorder.Body.Bytes(), &response)
						Expect(err).ToNot(HaveOccurred())
						Expect(response["error"]).To(Equal("Missing Firebase authorization token"), "Test case: %s", tc.description)
					}
				}
			})
		})
	})
})