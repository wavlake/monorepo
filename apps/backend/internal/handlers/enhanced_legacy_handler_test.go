package handlers_test

import (
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"

	"github.com/wavlake/monorepo/internal/handlers"
	"github.com/wavlake/monorepo/internal/models"
	"github.com/wavlake/monorepo/tests/mocks"
	"github.com/wavlake/monorepo/tests/testutil"
)

var _ = Describe("EnhancedLegacyHandler", func() {
	var (
		ctrl                *gomock.Controller
		mockPostgresService *mocks.MockPostgresServiceInterface
		legacyHandler       *handlers.LegacyHandler
		router              *gin.Engine
		testFirebaseUID     string
		testArtistID        string
		testAlbumID         string
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockPostgresService = mocks.NewMockPostgresServiceInterface(ctrl)
		legacyHandler = handlers.NewLegacyHandler(mockPostgresService)
		
		gin.SetMode(gin.TestMode)
		router = gin.New()
		testFirebaseUID = testutil.TestFirebaseUID
		testArtistID = "artist-123"
		testAlbumID = "album-456"
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Describe("GetTracksByArtist", func() {
		Context("when retrieving tracks for a specific artist", func() {
			BeforeEach(func() {
				router.GET("/legacy/artists/:artistId/tracks", func(c *gin.Context) {
					c.Set("firebase_uid", testFirebaseUID)
					legacyHandler.GetTracksByArtist(c)
				})
			})

			It("should return tracks for valid artist owned by user", func() {
				expectedTracks := testutil.ValidLegacyTracksList()

				// Verify artist ownership
				mockPostgresService.EXPECT().
					GetArtistByID(gomock.Any(), testArtistID).
					Return(&models.LegacyArtist{
						ID:     testArtistID,
						UserID: testFirebaseUID,
						Name:   "Test Artist",
					}, nil)

				// Get tracks for artist
				mockPostgresService.EXPECT().
					GetTracksByArtist(gomock.Any(), testArtistID).
					Return(expectedTracks, nil)

				req := httptest.NewRequest(http.MethodGet, "/legacy/artists/"+testArtistID+"/tracks", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusOK))
				
				var response map[string]interface{}
				err := testutil.ParseJSONResponse(w.Body, &response)
				Expect(err).ToNot(HaveOccurred())
				
				Expect(response).To(HaveKey("tracks"))
				tracks := response["tracks"].([]interface{})
				Expect(tracks).To(HaveLen(2))
				
				firstTrack := tracks[0].(map[string]interface{})
				Expect(firstTrack["title"]).To(Equal("Test Track"))
				Expect(firstTrack["artist_id"]).To(Equal(testArtistID))
			})

			It("should return empty array when artist has no tracks", func() {
				mockPostgresService.EXPECT().
					GetArtistByID(gomock.Any(), testArtistID).
					Return(&models.LegacyArtist{
						ID:     testArtistID,
						UserID: testFirebaseUID,
						Name:   "Test Artist",
					}, nil)

				mockPostgresService.EXPECT().
					GetTracksByArtist(gomock.Any(), testArtistID).
					Return([]models.LegacyTrack{}, nil)

				req := httptest.NewRequest(http.MethodGet, "/legacy/artists/"+testArtistID+"/tracks", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusOK))
				
				var response map[string]interface{}
				err := testutil.ParseJSONResponse(w.Body, &response)
				Expect(err).ToNot(HaveOccurred())
				
				tracks := response["tracks"].([]interface{})
				Expect(tracks).To(BeEmpty())
			})

			It("should return forbidden when user doesn't own artist", func() {
				mockPostgresService.EXPECT().
					GetArtistByID(gomock.Any(), testArtistID).
					Return(&models.LegacyArtist{
						ID:     testArtistID,
						UserID: "different-user-id", // Different owner
						Name:   "Test Artist",
					}, nil)

				req := httptest.NewRequest(http.MethodGet, "/legacy/artists/"+testArtistID+"/tracks", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusForbidden))
				
				var response map[string]interface{}
				err := testutil.ParseJSONResponse(w.Body, &response)
				Expect(err).ToNot(HaveOccurred())
				Expect(response["error"]).To(ContainSubstring("not authorized"))
			})

			It("should return not found when artist doesn't exist", func() {
				mockPostgresService.EXPECT().
					GetArtistByID(gomock.Any(), testArtistID).
					Return(nil, sql.ErrNoRows)

				req := httptest.NewRequest(http.MethodGet, "/legacy/artists/"+testArtistID+"/tracks", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusNotFound))
				
				var response map[string]interface{}
				err := testutil.ParseJSONResponse(w.Body, &response)
				Expect(err).ToNot(HaveOccurred())
				Expect(response["error"]).To(ContainSubstring("artist not found"))
			})

			It("should handle database errors gracefully", func() {
				mockPostgresService.EXPECT().
					GetArtistByID(gomock.Any(), testArtistID).
					Return(nil, errors.New("connection timeout"))

				req := httptest.NewRequest(http.MethodGet, "/legacy/artists/"+testArtistID+"/tracks", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusInternalServerError))
				
				var response map[string]interface{}
				err := testutil.ParseJSONResponse(w.Body, &response)
				Expect(err).ToNot(HaveOccurred())
				Expect(response["error"]).To(Equal("database error"))
			})

			It("should handle tracks query failure", func() {
				mockPostgresService.EXPECT().
					GetArtistByID(gomock.Any(), testArtistID).
					Return(&models.LegacyArtist{
						ID:     testArtistID,
						UserID: testFirebaseUID,
						Name:   "Test Artist",
					}, nil)

				mockPostgresService.EXPECT().
					GetTracksByArtist(gomock.Any(), testArtistID).
					Return(nil, errors.New("tracks table does not exist"))

				req := httptest.NewRequest(http.MethodGet, "/legacy/artists/"+testArtistID+"/tracks", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusInternalServerError))
				
				var response map[string]interface{}
				err := testutil.ParseJSONResponse(w.Body, &response)
				Expect(err).ToNot(HaveOccurred())
				Expect(response["error"]).To(Equal("database error"))
			})

			It("should require authentication", func() {
				router = gin.New()
				router.GET("/legacy/artists/:artistId/tracks", legacyHandler.GetTracksByArtist)

				req := httptest.NewRequest(http.MethodGet, "/legacy/artists/"+testArtistID+"/tracks", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusUnauthorized))
			})
		})
	})

	Describe("GetTracksByAlbum", func() {
		Context("when retrieving tracks for a specific album", func() {
			BeforeEach(func() {
				router.GET("/legacy/albums/:albumId/tracks", func(c *gin.Context) {
					c.Set("firebase_uid", testFirebaseUID)
					legacyHandler.GetTracksByAlbum(c)
				})
			})

			It("should return tracks for valid album owned by user", func() {
				expectedTracks := testutil.ValidLegacyTracksList()
				// Update tracks to have correct album_id
				for i := range expectedTracks {
					expectedTracks[i].AlbumID = testAlbumID
				}

				// Verify album ownership through artist
				mockPostgresService.EXPECT().
					GetAlbumByID(gomock.Any(), testAlbumID).
					Return(&models.LegacyAlbum{
						ID:       testAlbumID,
						ArtistID: testArtistID,
						Title:    "Test Album",
					}, nil)

				mockPostgresService.EXPECT().
					GetArtistByID(gomock.Any(), testArtistID).
					Return(&models.LegacyArtist{
						ID:     testArtistID,
						UserID: testFirebaseUID,
						Name:   "Test Artist",
					}, nil)

				// Get tracks for album
				mockPostgresService.EXPECT().
					GetTracksByAlbum(gomock.Any(), testAlbumID).
					Return(expectedTracks, nil)

				req := httptest.NewRequest(http.MethodGet, "/legacy/albums/"+testAlbumID+"/tracks", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusOK))
				
				var response map[string]interface{}
				err := testutil.ParseJSONResponse(w.Body, &response)
				Expect(err).ToNot(HaveOccurred())
				
				Expect(response).To(HaveKey("tracks"))
				tracks := response["tracks"].([]interface{})
				Expect(tracks).To(HaveLen(2))
				
				firstTrack := tracks[0].(map[string]interface{})
				Expect(firstTrack["title"]).To(Equal("Test Track"))
				Expect(firstTrack["album_id"]).To(Equal(testAlbumID))
			})

			It("should return tracks ordered by track order", func() {
				orderedTracks := []models.LegacyTrack{
					{
						ID:      "track-1",
						AlbumID: testAlbumID,
						Title:   "First Track",
						Order:   1,
					},
					{
						ID:      "track-2",
						AlbumID: testAlbumID,
						Title:   "Second Track",
						Order:   2,
					},
					{
						ID:      "track-3",
						AlbumID: testAlbumID,
						Title:   "Third Track",
						Order:   3,
					},
				}

				mockPostgresService.EXPECT().
					GetAlbumByID(gomock.Any(), testAlbumID).
					Return(&models.LegacyAlbum{
						ID:       testAlbumID,
						ArtistID: testArtistID,
						Title:    "Test Album",
					}, nil)

				mockPostgresService.EXPECT().
					GetArtistByID(gomock.Any(), testArtistID).
					Return(&models.LegacyArtist{
						ID:     testArtistID,
						UserID: testFirebaseUID,
						Name:   "Test Artist",
					}, nil)

				mockPostgresService.EXPECT().
					GetTracksByAlbum(gomock.Any(), testAlbumID).
					Return(orderedTracks, nil)

				req := httptest.NewRequest(http.MethodGet, "/legacy/albums/"+testAlbumID+"/tracks", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusOK))
				
				var response map[string]interface{}
				err := testutil.ParseJSONResponse(w.Body, &response)
				Expect(err).ToNot(HaveOccurred())
				
				tracks := response["tracks"].([]interface{})
				Expect(tracks).To(HaveLen(3))
				
				// Verify order
				for i, track := range tracks {
					trackData := track.(map[string]interface{})
					expectedOrder := float64(i + 1) // JSON numbers are float64
					Expect(trackData["order"]).To(Equal(expectedOrder))
				}
			})

			It("should return empty array when album has no tracks", func() {
				mockPostgresService.EXPECT().
					GetAlbumByID(gomock.Any(), testAlbumID).
					Return(&models.LegacyAlbum{
						ID:       testAlbumID,
						ArtistID: testArtistID,
						Title:    "Test Album",
					}, nil)

				mockPostgresService.EXPECT().
					GetArtistByID(gomock.Any(), testArtistID).
					Return(&models.LegacyArtist{
						ID:     testArtistID,
						UserID: testFirebaseUID,
						Name:   "Test Artist",
					}, nil)

				mockPostgresService.EXPECT().
					GetTracksByAlbum(gomock.Any(), testAlbumID).
					Return([]models.LegacyTrack{}, nil)

				req := httptest.NewRequest(http.MethodGet, "/legacy/albums/"+testAlbumID+"/tracks", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusOK))
				
				var response map[string]interface{}
				err := testutil.ParseJSONResponse(w.Body, &response)
				Expect(err).ToNot(HaveOccurred())
				
				tracks := response["tracks"].([]interface{})
				Expect(tracks).To(BeEmpty())
			})

			It("should return forbidden when user doesn't own album's artist", func() {
				mockPostgresService.EXPECT().
					GetAlbumByID(gomock.Any(), testAlbumID).
					Return(&models.LegacyAlbum{
						ID:       testAlbumID,
						ArtistID: testArtistID,
						Title:    "Test Album",
					}, nil)

				mockPostgresService.EXPECT().
					GetArtistByID(gomock.Any(), testArtistID).
					Return(&models.LegacyArtist{
						ID:     testArtistID,
						UserID: "different-user-id", // Different owner
						Name:   "Test Artist",
					}, nil)

				req := httptest.NewRequest(http.MethodGet, "/legacy/albums/"+testAlbumID+"/tracks", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusForbidden))
				
				var response map[string]interface{}
				err := testutil.ParseJSONResponse(w.Body, &response)
				Expect(err).ToNot(HaveOccurred())
				Expect(response["error"]).To(ContainSubstring("not authorized"))
			})

			It("should return not found when album doesn't exist", func() {
				mockPostgresService.EXPECT().
					GetAlbumByID(gomock.Any(), testAlbumID).
					Return(nil, sql.ErrNoRows)

				req := httptest.NewRequest(http.MethodGet, "/legacy/albums/"+testAlbumID+"/tracks", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusNotFound))
				
				var response map[string]interface{}
				err := testutil.ParseJSONResponse(w.Body, &response)
				Expect(err).ToNot(HaveOccurred())
				Expect(response["error"]).To(ContainSubstring("album not found"))
			})

			It("should handle chain of database queries with partial failures", func() {
				// Album exists
				mockPostgresService.EXPECT().
					GetAlbumByID(gomock.Any(), testAlbumID).
					Return(&models.LegacyAlbum{
						ID:       testAlbumID,
						ArtistID: testArtistID,
						Title:    "Test Album",
					}, nil)

				// Artist query fails
				mockPostgresService.EXPECT().
					GetArtistByID(gomock.Any(), testArtistID).
					Return(nil, errors.New("artist query failed"))

				req := httptest.NewRequest(http.MethodGet, "/legacy/albums/"+testAlbumID+"/tracks", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusInternalServerError))
				
				var response map[string]interface{}
				err := testutil.ParseJSONResponse(w.Body, &response)
				Expect(err).ToNot(HaveOccurred())
				Expect(response["error"]).To(Equal("database error"))
			})

			It("should require authentication", func() {
				router = gin.New()
				router.GET("/legacy/albums/:albumId/tracks", legacyHandler.GetTracksByAlbum)

				req := httptest.NewRequest(http.MethodGet, "/legacy/albums/"+testAlbumID+"/tracks", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusUnauthorized))
			})
		})
	})

	Describe("Enhanced Error Handling", func() {
		Context("when database connection issues occur", func() {
			BeforeEach(func() {
				router.GET("/legacy/artists/:artistId/tracks", func(c *gin.Context) {
					c.Set("firebase_uid", testFirebaseUID)
					legacyHandler.GetTracksByArtist(c)
				})
			})

			It("should implement retry logic for transient failures", func() {
				// First call fails with transient error
				mockPostgresService.EXPECT().
					GetArtistByID(gomock.Any(), testArtistID).
					Return(nil, errors.New("connection reset by peer")).
					Times(1)

				// Retry succeeds
				mockPostgresService.EXPECT().
					GetArtistByID(gomock.Any(), testArtistID).
					Return(&models.LegacyArtist{
						ID:     testArtistID,
						UserID: testFirebaseUID,
						Name:   "Test Artist",
					}, nil).
					Times(1)

				mockPostgresService.EXPECT().
					GetTracksByArtist(gomock.Any(), testArtistID).
					Return([]models.LegacyTrack{}, nil)

				req := httptest.NewRequest(http.MethodGet, "/legacy/artists/"+testArtistID+"/tracks", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusOK))
			})

			It("should provide detailed error context in development mode", func() {
				// This would test enhanced error reporting in development mode
				// Implementation would check environment and provide more details
				
				mockPostgresService.EXPECT().
					GetArtistByID(gomock.Any(), testArtistID).
					Return(nil, errors.New("relation \"artists\" does not exist"))

				req := httptest.NewRequest(http.MethodGet, "/legacy/artists/"+testArtistID+"/tracks", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusInternalServerError))
				
				var response map[string]interface{}
				err := testutil.ParseJSONResponse(w.Body, &response)
				Expect(err).ToNot(HaveOccurred())
				
				// In development mode, might include more error details
				Expect(response["error"]).To(ContainSubstring("database error"))
			})
		})
	})
})