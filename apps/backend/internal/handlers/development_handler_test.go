package handlers_test

import (
	"net/http"
	"net/http/httptest"

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
	})
})