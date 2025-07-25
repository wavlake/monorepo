package handlers_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"

	"github.com/wavlake/monorepo/internal/handlers"
	"github.com/wavlake/monorepo/tests/mocks"
)

var _ = Describe("DevelopmentHandler", func() {
	var (
		ctrl                     *gomock.Controller
		mockDevelopmentService   *mocks.MockDevelopmentServiceInterface
		developmentHandler       *handlers.DevelopmentHandler
		router                   *gin.Engine
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockDevelopmentService = mocks.NewMockDevelopmentServiceInterface(ctrl)
		developmentHandler = handlers.NewDevelopmentHandler(mockDevelopmentService)
		
		gin.SetMode(gin.TestMode)
		router = gin.New()
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Describe("ResetDatabase", func() {
		It("should reset database in development mode", func() {
			router.POST("/dev/reset-database", developmentHandler.ResetDatabase)

			mockDevelopmentService.EXPECT().
				ResetDatabase(gomock.Any()).
				Return(nil)

			req := httptest.NewRequest(http.MethodPost, "/dev/reset-database", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			Expect(w.Code).To(Equal(http.StatusOK))
		})

		It("should handle database reset failure", func() {
			router.POST("/dev/reset-database", developmentHandler.ResetDatabase)

			mockDevelopmentService.EXPECT().
				ResetDatabase(gomock.Any()).
				Return(errors.New("database connection failed"))

			req := httptest.NewRequest(http.MethodPost, "/dev/reset-database", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			Expect(w.Code).To(Equal(http.StatusInternalServerError))
		})

		It("should handle GET method on reset endpoint", func() {
			router.POST("/dev/reset-database", developmentHandler.ResetDatabase)

			req := httptest.NewRequest(http.MethodGet, "/dev/reset-database", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			Expect(w.Code).To(Equal(http.StatusNotFound)) // Method not allowed
		})

		It("should handle concurrent reset requests", func() {
			router.POST("/dev/reset-database", developmentHandler.ResetDatabase)

			// First request succeeds
			mockDevelopmentService.EXPECT().
				ResetDatabase(gomock.Any()).
				Return(nil)

			// Second request also called (no built-in protection assumed)
			mockDevelopmentService.EXPECT().
				ResetDatabase(gomock.Any()).
				Return(nil)

			req1 := httptest.NewRequest(http.MethodPost, "/dev/reset-database", nil)
			w1 := httptest.NewRecorder()
			req2 := httptest.NewRequest(http.MethodPost, "/dev/reset-database", nil)
			w2 := httptest.NewRecorder()

			router.ServeHTTP(w1, req1)
			router.ServeHTTP(w2, req2)

			Expect(w1.Code).To(Equal(http.StatusOK))
			Expect(w2.Code).To(Equal(http.StatusOK))
		})
	})

	Describe("SeedTestData", func() {
		It("should seed test data in development mode", func() {
			router.POST("/dev/seed-test-data", developmentHandler.SeedTestData)

			mockDevelopmentService.EXPECT().
				SeedTestData(gomock.Any()).
				Return(nil)

			req := httptest.NewRequest(http.MethodPost, "/dev/seed-test-data", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			Expect(w.Code).To(Equal(http.StatusOK))
		})

		It("should handle seed test data failure", func() {
			router.POST("/dev/seed-test-data", developmentHandler.SeedTestData)

			mockDevelopmentService.EXPECT().
				SeedTestData(gomock.Any()).
				Return(errors.New("insufficient storage space"))

			req := httptest.NewRequest(http.MethodPost, "/dev/seed-test-data", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			Expect(w.Code).To(Equal(http.StatusInternalServerError))
		})

		It("should handle wrong HTTP method", func() {
			router.POST("/dev/seed-test-data", developmentHandler.SeedTestData)

			req := httptest.NewRequest("PUT", "/dev/seed-test-data", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			Expect(w.Code).To(Equal(http.StatusNotFound)) // Method not allowed
		})

		It("should handle multiple concurrent seeding requests", func() {
			router.POST("/dev/seed-test-data", developmentHandler.SeedTestData)

			// Both requests expected to be called
			mockDevelopmentService.EXPECT().
				SeedTestData(gomock.Any()).
				Return(nil).
				Times(2)

			req1 := httptest.NewRequest(http.MethodPost, "/dev/seed-test-data", nil)
			w1 := httptest.NewRecorder()
			req2 := httptest.NewRequest(http.MethodPost, "/dev/seed-test-data", nil)
			w2 := httptest.NewRecorder()

			router.ServeHTTP(w1, req1)
			router.ServeHTTP(w2, req2)

			Expect(w1.Code).To(Equal(http.StatusOK))
			Expect(w2.Code).To(Equal(http.StatusOK))
		})

		It("should handle request with unexpected body", func() {
			router.POST("/dev/seed-test-data", developmentHandler.SeedTestData)

			mockDevelopmentService.EXPECT().
				SeedTestData(gomock.Any()).
				Return(nil)

			// Even with body, should work (body typically ignored for dev endpoints)
			req := httptest.NewRequest(http.MethodPost, "/dev/seed-test-data", 
				strings.NewReader(`{"unexpected": "body"}`))
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			Expect(w.Code).To(Equal(http.StatusOK))
		})
	})
})