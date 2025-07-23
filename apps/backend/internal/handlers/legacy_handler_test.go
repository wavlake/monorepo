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

var _ = Describe("LegacyHandler", func() {
	var (
		ctrl                *gomock.Controller
		mockPostgresService *mocks.MockPostgresServiceInterface
		legacyHandler       *handlers.LegacyHandler
		router              *gin.Engine
		testFirebaseUID     string
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockPostgresService = mocks.NewMockPostgresServiceInterface(ctrl)
		legacyHandler = handlers.NewLegacyHandler(mockPostgresService)

		gin.SetMode(gin.TestMode)
		router = gin.New()
		testFirebaseUID = testutil.TestFirebaseUID
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Describe("GetUserMetadata", func() {
		var (
			expectedUser    *models.LegacyUser
			expectedArtists []models.LegacyArtist
			expectedAlbums  []models.LegacyAlbum
			expectedTracks  []models.LegacyTrack
		)

		BeforeEach(func() {
			user := testutil.ValidLegacyUser()
			expectedUser = &user
			expectedArtists = testutil.ValidLegacyArtistsList()
			expectedAlbums = testutil.ValidLegacyAlbumsList()
			expectedTracks = testutil.ValidLegacyTracksList()
		})

		Context("when user is authenticated", func() {
			BeforeEach(func() {
				router.GET("/legacy/metadata", func(c *gin.Context) {
					c.Set("firebase_uid", testFirebaseUID)
					legacyHandler.GetUserMetadata(c)
				})
			})

			It("should return complete user metadata when user exists", func() {
				mockPostgresService.EXPECT().
					GetUserByFirebaseUID(gomock.Any(), testFirebaseUID).
					Return(expectedUser, nil)

				mockPostgresService.EXPECT().
					GetUserArtists(gomock.Any(), testFirebaseUID).
					Return(expectedArtists, nil)

				mockPostgresService.EXPECT().
					GetUserAlbums(gomock.Any(), testFirebaseUID).
					Return(expectedAlbums, nil)

				mockPostgresService.EXPECT().
					GetUserTracks(gomock.Any(), testFirebaseUID).
					Return(expectedTracks, nil)

				req := httptest.NewRequest(http.MethodGet, "/legacy/metadata", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusOK))

				var response map[string]interface{}
				err := testutil.ParseJSONResponse(w.Body, &response)
				Expect(err).ToNot(HaveOccurred())

				Expect(response).To(HaveKey("user"))
				Expect(response).To(HaveKey("artists"))
				Expect(response).To(HaveKey("albums"))
				Expect(response).To(HaveKey("tracks"))

				// Verify user data structure
				user := response["user"].(map[string]interface{})
				Expect(user["id"]).To(Equal(expectedUser.ID))
				Expect(user["name"]).To(Equal(expectedUser.Name))

				// Verify arrays are populated
				artists := response["artists"].([]interface{})
				albums := response["albums"].([]interface{})
				tracks := response["tracks"].([]interface{})

				Expect(artists).To(HaveLen(2))
				Expect(albums).To(HaveLen(2))
				Expect(tracks).To(HaveLen(2))
			})

			It("should return empty metadata when user does not exist", func() {
				mockPostgresService.EXPECT().
					GetUserByFirebaseUID(gomock.Any(), testFirebaseUID).
					Return(nil, sql.ErrNoRows)

				req := httptest.NewRequest(http.MethodGet, "/legacy/metadata", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusOK))

				var response map[string]interface{}
				err := testutil.ParseJSONResponse(w.Body, &response)
				Expect(err).ToNot(HaveOccurred())

				Expect(response["user"]).To(BeNil())

				// Verify empty arrays are returned, not null
				artists := response["artists"].([]interface{})
				albums := response["albums"].([]interface{})
				tracks := response["tracks"].([]interface{})

				Expect(artists).To(BeEmpty())
				Expect(albums).To(BeEmpty())
				Expect(tracks).To(BeEmpty())
			})

			It("should handle partial data retrieval when some queries fail", func() {
				mockPostgresService.EXPECT().
					GetUserByFirebaseUID(gomock.Any(), testFirebaseUID).
					Return(expectedUser, nil)

				mockPostgresService.EXPECT().
					GetUserArtists(gomock.Any(), testFirebaseUID).
					Return(expectedArtists, nil)

				// Albums query fails with non-database error (user not found)
				mockPostgresService.EXPECT().
					GetUserAlbums(gomock.Any(), testFirebaseUID).
					Return(nil, sql.ErrNoRows)

				mockPostgresService.EXPECT().
					GetUserTracks(gomock.Any(), testFirebaseUID).
					Return(expectedTracks, nil)

				req := httptest.NewRequest(http.MethodGet, "/legacy/metadata", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusOK))

				var response map[string]interface{}
				err := testutil.ParseJSONResponse(w.Body, &response)
				Expect(err).ToNot(HaveOccurred())

				// User and other data should be present
				Expect(response["user"]).ToNot(BeNil())

				artists := response["artists"].([]interface{})
				albums := response["albums"].([]interface{})
				tracks := response["tracks"].([]interface{})

				Expect(artists).To(HaveLen(2))
				Expect(albums).To(BeEmpty()) // Failed query returns empty array
				Expect(tracks).To(HaveLen(2))
			})

			It("should return database error when user query fails with database error", func() {
				mockPostgresService.EXPECT().
					GetUserByFirebaseUID(gomock.Any(), testFirebaseUID).
					Return(nil, errors.New("connection timeout"))

				req := httptest.NewRequest(http.MethodGet, "/legacy/metadata", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusInternalServerError))

				var response map[string]interface{}
				err := testutil.ParseJSONResponse(w.Body, &response)
				Expect(err).ToNot(HaveOccurred())

				Expect(response).To(HaveKey("error"))
				Expect(response["error"]).To(Equal("Database error"))
			})

			It("should handle database errors gracefully in related data queries", func() {
				mockPostgresService.EXPECT().
					GetUserByFirebaseUID(gomock.Any(), testFirebaseUID).
					Return(expectedUser, nil)

				// Artists query fails with database error
				mockPostgresService.EXPECT().
					GetUserArtists(gomock.Any(), testFirebaseUID).
					Return(nil, errors.New("relation artists does not exist"))

				mockPostgresService.EXPECT().
					GetUserAlbums(gomock.Any(), testFirebaseUID).
					Return(expectedAlbums, nil)

				mockPostgresService.EXPECT().
					GetUserTracks(gomock.Any(), testFirebaseUID).
					Return(expectedTracks, nil)

				req := httptest.NewRequest(http.MethodGet, "/legacy/metadata", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusOK))

				var response map[string]interface{}
				err := testutil.ParseJSONResponse(w.Body, &response)
				Expect(err).ToNot(HaveOccurred())

				// Should continue with other data despite database error
				Expect(response["user"]).ToNot(BeNil())

				artists := response["artists"].([]interface{})
				albums := response["albums"].([]interface{})
				tracks := response["tracks"].([]interface{})

				Expect(artists).To(BeEmpty()) // Database error results in empty array
				Expect(albums).To(HaveLen(2))
				Expect(tracks).To(HaveLen(2))
			})

			It("should handle network timeout errors in all queries", func() {
				mockPostgresService.EXPECT().
					GetUserByFirebaseUID(gomock.Any(), testFirebaseUID).
					Return(expectedUser, nil)

				mockPostgresService.EXPECT().
					GetUserArtists(gomock.Any(), testFirebaseUID).
					Return(nil, errors.New("network timeout"))

				mockPostgresService.EXPECT().
					GetUserAlbums(gomock.Any(), testFirebaseUID).
					Return(nil, errors.New("connection lost"))

				mockPostgresService.EXPECT().
					GetUserTracks(gomock.Any(), testFirebaseUID).
					Return(nil, errors.New("permission denied"))

				req := httptest.NewRequest(http.MethodGet, "/legacy/metadata", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusOK))

				var response map[string]interface{}
				err := testutil.ParseJSONResponse(w.Body, &response)
				Expect(err).ToNot(HaveOccurred())

				// Should handle all database errors gracefully
				Expect(response["user"]).ToNot(BeNil())

				artists := response["artists"].([]interface{})
				albums := response["albums"].([]interface{})
				tracks := response["tracks"].([]interface{})

				Expect(artists).To(BeEmpty())
				Expect(albums).To(BeEmpty())
				Expect(tracks).To(BeEmpty())
			})
		})

		Context("when user is not authenticated", func() {
			BeforeEach(func() {
				router.GET("/legacy/metadata", legacyHandler.GetUserMetadata)
			})

			It("should return unauthorized when firebase_uid is missing", func() {
				req := httptest.NewRequest(http.MethodGet, "/legacy/metadata", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusUnauthorized))

				var response map[string]interface{}
				err := testutil.ParseJSONResponse(w.Body, &response)
				Expect(err).ToNot(HaveOccurred())

				Expect(response).To(HaveKey("error"))
				Expect(response["error"]).To(Equal("Failed to find an associated Firebase UID"))
			})

			It("should return unauthorized when firebase_uid is empty string", func() {
				router = gin.New()
				router.GET("/legacy/metadata", func(c *gin.Context) {
					c.Set("firebase_uid", "")
					legacyHandler.GetUserMetadata(c)
				})

				req := httptest.NewRequest(http.MethodGet, "/legacy/metadata", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusUnauthorized))
			})
		})
	})

	Describe("GetUserTracks", func() {
		var expectedTracks []models.LegacyTrack

		BeforeEach(func() {
			expectedTracks = testutil.ValidLegacyTracksList()
		})

		Context("when user is authenticated", func() {
			BeforeEach(func() {
				router.GET("/legacy/tracks", func(c *gin.Context) {
					c.Set("firebase_uid", testFirebaseUID)
					legacyHandler.GetUserTracks(c)
				})
			})

			It("should return user tracks when they exist", func() {
				mockPostgresService.EXPECT().
					GetUserTracks(gomock.Any(), testFirebaseUID).
					Return(expectedTracks, nil)

				req := httptest.NewRequest(http.MethodGet, "/legacy/tracks", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusOK))

				var response map[string]interface{}
				err := testutil.ParseJSONResponse(w.Body, &response)
				Expect(err).ToNot(HaveOccurred())

				Expect(response).To(HaveKey("tracks"))
				tracks := response["tracks"].([]interface{})
				Expect(tracks).To(HaveLen(2))

				// Verify track data structure
				firstTrack := tracks[0].(map[string]interface{})
				Expect(firstTrack["title"]).To(Equal("Test Track"))
				Expect(firstTrack["id"]).To(Equal("track-123"))
			})

			It("should return empty array when user has no tracks", func() {
				mockPostgresService.EXPECT().
					GetUserTracks(gomock.Any(), testFirebaseUID).
					Return([]models.LegacyTrack{}, nil)

				req := httptest.NewRequest(http.MethodGet, "/legacy/tracks", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusOK))

				var response map[string]interface{}
				err := testutil.ParseJSONResponse(w.Body, &response)
				Expect(err).ToNot(HaveOccurred())

				Expect(response).To(HaveKey("tracks"))
				tracks := response["tracks"].([]interface{})
				Expect(tracks).To(BeEmpty())
			})

			It("should return empty array when user not found (non-database error)", func() {
				mockPostgresService.EXPECT().
					GetUserTracks(gomock.Any(), testFirebaseUID).
					Return(nil, sql.ErrNoRows)

				req := httptest.NewRequest(http.MethodGet, "/legacy/tracks", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusOK))

				var response map[string]interface{}
				err := testutil.ParseJSONResponse(w.Body, &response)
				Expect(err).ToNot(HaveOccurred())

				Expect(response).To(HaveKey("tracks"))
				tracks := response["tracks"].([]interface{})
				Expect(tracks).To(BeEmpty())
			})

			It("should return database error when query fails with database error", func() {
				mockPostgresService.EXPECT().
					GetUserTracks(gomock.Any(), testFirebaseUID).
					Return(nil, errors.New("relation tracks does not exist"))

				req := httptest.NewRequest(http.MethodGet, "/legacy/tracks", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusInternalServerError))

				var response map[string]interface{}
				err := testutil.ParseJSONResponse(w.Body, &response)
				Expect(err).ToNot(HaveOccurred())

				Expect(response).To(HaveKey("error"))
				Expect(response["error"]).To(Equal("database error"))
			})

			It("should handle connection timeout errors", func() {
				mockPostgresService.EXPECT().
					GetUserTracks(gomock.Any(), testFirebaseUID).
					Return(nil, errors.New("connection timeout"))

				req := httptest.NewRequest(http.MethodGet, "/legacy/tracks", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusInternalServerError))

				var response map[string]interface{}
				err := testutil.ParseJSONResponse(w.Body, &response)
				Expect(err).ToNot(HaveOccurred())

				Expect(response["error"]).To(Equal("database error"))
			})

			It("should handle network errors gracefully", func() {
				mockPostgresService.EXPECT().
					GetUserTracks(gomock.Any(), testFirebaseUID).
					Return(nil, errors.New("network unreachable"))

				req := httptest.NewRequest(http.MethodGet, "/legacy/tracks", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusInternalServerError))
			})

			It("should handle permission denied errors", func() {
				mockPostgresService.EXPECT().
					GetUserTracks(gomock.Any(), testFirebaseUID).
					Return(nil, errors.New("permission denied on table tracks"))

				req := httptest.NewRequest(http.MethodGet, "/legacy/tracks", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusInternalServerError))
			})
		})

		Context("when user is not authenticated", func() {
			BeforeEach(func() {
				router.GET("/legacy/tracks", legacyHandler.GetUserTracks)
			})

			It("should return unauthorized when firebase_uid is missing", func() {
				req := httptest.NewRequest(http.MethodGet, "/legacy/tracks", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusUnauthorized))

				var response map[string]interface{}
				err := testutil.ParseJSONResponse(w.Body, &response)
				Expect(err).ToNot(HaveOccurred())

				Expect(response).To(HaveKey("error"))
				Expect(response["error"]).To(Equal("authentication required"))
			})

			It("should return unauthorized when firebase_uid is empty", func() {
				router = gin.New()
				router.GET("/legacy/tracks", func(c *gin.Context) {
					c.Set("firebase_uid", "")
					legacyHandler.GetUserTracks(c)
				})

				req := httptest.NewRequest(http.MethodGet, "/legacy/tracks", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusUnauthorized))
			})
		})
	})

	Describe("GetUserArtists", func() {
		var expectedArtists []models.LegacyArtist

		BeforeEach(func() {
			expectedArtists = testutil.ValidLegacyArtistsList()
		})

		Context("when user is authenticated", func() {
			BeforeEach(func() {
				router.GET("/legacy/artists", func(c *gin.Context) {
					c.Set("firebase_uid", testFirebaseUID)
					legacyHandler.GetUserArtists(c)
				})
			})

			It("should return user artists when they exist", func() {
				mockPostgresService.EXPECT().
					GetUserArtists(gomock.Any(), testFirebaseUID).
					Return(expectedArtists, nil)

				req := httptest.NewRequest(http.MethodGet, "/legacy/artists", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusOK))

				var response map[string]interface{}
				err := testutil.ParseJSONResponse(w.Body, &response)
				Expect(err).ToNot(HaveOccurred())

				Expect(response).To(HaveKey("artists"))
				artists := response["artists"].([]interface{})
				Expect(artists).To(HaveLen(2))

				// Verify artist data structure
				firstArtist := artists[0].(map[string]interface{})
				Expect(firstArtist["name"]).To(Equal("Test Artist"))
				Expect(firstArtist["id"]).To(Equal("artist-123"))
			})

			It("should return empty array when user has no artists", func() {
				mockPostgresService.EXPECT().
					GetUserArtists(gomock.Any(), testFirebaseUID).
					Return([]models.LegacyArtist{}, nil)

				req := httptest.NewRequest(http.MethodGet, "/legacy/artists", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusOK))

				var response map[string]interface{}
				err := testutil.ParseJSONResponse(w.Body, &response)
				Expect(err).ToNot(HaveOccurred())

				Expect(response).To(HaveKey("artists"))
				artists := response["artists"].([]interface{})
				Expect(artists).To(BeEmpty())
			})

			It("should return empty array when user not found (non-database error)", func() {
				mockPostgresService.EXPECT().
					GetUserArtists(gomock.Any(), testFirebaseUID).
					Return(nil, sql.ErrNoRows)

				req := httptest.NewRequest(http.MethodGet, "/legacy/artists", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusOK))

				var response map[string]interface{}
				err := testutil.ParseJSONResponse(w.Body, &response)
				Expect(err).ToNot(HaveOccurred())

				Expect(response).To(HaveKey("artists"))
				artists := response["artists"].([]interface{})
				Expect(artists).To(BeEmpty())
			})

			It("should return database error when query fails with database error", func() {
				mockPostgresService.EXPECT().
					GetUserArtists(gomock.Any(), testFirebaseUID).
					Return(nil, errors.New("relation artists does not exist"))

				req := httptest.NewRequest(http.MethodGet, "/legacy/artists", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusInternalServerError))

				var response map[string]interface{}
				err := testutil.ParseJSONResponse(w.Body, &response)
				Expect(err).ToNot(HaveOccurred())

				Expect(response).To(HaveKey("error"))
				Expect(response["error"]).To(Equal("database error"))
			})

			It("should handle connection timeout errors", func() {
				mockPostgresService.EXPECT().
					GetUserArtists(gomock.Any(), testFirebaseUID).
					Return(nil, errors.New("connection timeout"))

				req := httptest.NewRequest(http.MethodGet, "/legacy/artists", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusInternalServerError))

				var response map[string]interface{}
				err := testutil.ParseJSONResponse(w.Body, &response)
				Expect(err).ToNot(HaveOccurred())

				Expect(response["error"]).To(Equal("database error"))
			})

			It("should handle network errors gracefully", func() {
				mockPostgresService.EXPECT().
					GetUserArtists(gomock.Any(), testFirebaseUID).
					Return(nil, errors.New("network unreachable"))

				req := httptest.NewRequest(http.MethodGet, "/legacy/artists", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusInternalServerError))
			})

			It("should handle permission denied errors", func() {
				mockPostgresService.EXPECT().
					GetUserArtists(gomock.Any(), testFirebaseUID).
					Return(nil, errors.New("permission denied on table artists"))

				req := httptest.NewRequest(http.MethodGet, "/legacy/artists", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusInternalServerError))
			})
		})

		Context("when user is not authenticated", func() {
			BeforeEach(func() {
				router.GET("/legacy/artists", legacyHandler.GetUserArtists)
			})

			It("should return unauthorized when firebase_uid is missing", func() {
				req := httptest.NewRequest(http.MethodGet, "/legacy/artists", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusUnauthorized))

				var response map[string]interface{}
				err := testutil.ParseJSONResponse(w.Body, &response)
				Expect(err).ToNot(HaveOccurred())

				Expect(response).To(HaveKey("error"))
				Expect(response["error"]).To(Equal("authentication required"))
			})

			It("should return unauthorized when firebase_uid is empty", func() {
				router = gin.New()
				router.GET("/legacy/artists", func(c *gin.Context) {
					c.Set("firebase_uid", "")
					legacyHandler.GetUserArtists(c)
				})

				req := httptest.NewRequest(http.MethodGet, "/legacy/artists", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusUnauthorized))
			})
		})
	})

	Describe("GetUserAlbums", func() {
		var expectedAlbums []models.LegacyAlbum

		BeforeEach(func() {
			expectedAlbums = testutil.ValidLegacyAlbumsList()
		})

		Context("when user is authenticated", func() {
			BeforeEach(func() {
				router.GET("/legacy/albums", func(c *gin.Context) {
					c.Set("firebase_uid", testFirebaseUID)
					legacyHandler.GetUserAlbums(c)
				})
			})

			It("should return user albums when they exist", func() {
				mockPostgresService.EXPECT().
					GetUserAlbums(gomock.Any(), testFirebaseUID).
					Return(expectedAlbums, nil)

				req := httptest.NewRequest(http.MethodGet, "/legacy/albums", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusOK))

				var response map[string]interface{}
				err := testutil.ParseJSONResponse(w.Body, &response)
				Expect(err).ToNot(HaveOccurred())

				Expect(response).To(HaveKey("albums"))
				albums := response["albums"].([]interface{})
				Expect(albums).To(HaveLen(2))

				// Verify album data structure
				firstAlbum := albums[0].(map[string]interface{})
				Expect(firstAlbum["title"]).To(Equal("Test Album"))
				Expect(firstAlbum["id"]).To(Equal("album-123"))
			})

			It("should return empty array when user has no albums", func() {
				mockPostgresService.EXPECT().
					GetUserAlbums(gomock.Any(), testFirebaseUID).
					Return([]models.LegacyAlbum{}, nil)

				req := httptest.NewRequest(http.MethodGet, "/legacy/albums", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusOK))

				var response map[string]interface{}
				err := testutil.ParseJSONResponse(w.Body, &response)
				Expect(err).ToNot(HaveOccurred())

				Expect(response).To(HaveKey("albums"))
				albums := response["albums"].([]interface{})
				Expect(albums).To(BeEmpty())
			})

			It("should return empty array when user not found (non-database error)", func() {
				mockPostgresService.EXPECT().
					GetUserAlbums(gomock.Any(), testFirebaseUID).
					Return(nil, sql.ErrNoRows)

				req := httptest.NewRequest(http.MethodGet, "/legacy/albums", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusOK))

				var response map[string]interface{}
				err := testutil.ParseJSONResponse(w.Body, &response)
				Expect(err).ToNot(HaveOccurred())

				Expect(response).To(HaveKey("albums"))
				albums := response["albums"].([]interface{})
				Expect(albums).To(BeEmpty())
			})

			It("should return database error when query fails with database error", func() {
				mockPostgresService.EXPECT().
					GetUserAlbums(gomock.Any(), testFirebaseUID).
					Return(nil, errors.New("relation albums does not exist"))

				req := httptest.NewRequest(http.MethodGet, "/legacy/albums", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusInternalServerError))

				var response map[string]interface{}
				err := testutil.ParseJSONResponse(w.Body, &response)
				Expect(err).ToNot(HaveOccurred())

				Expect(response).To(HaveKey("error"))
				Expect(response["error"]).To(Equal("database error"))
			})

			It("should handle connection timeout errors", func() {
				mockPostgresService.EXPECT().
					GetUserAlbums(gomock.Any(), testFirebaseUID).
					Return(nil, errors.New("connection timeout"))

				req := httptest.NewRequest(http.MethodGet, "/legacy/albums", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusInternalServerError))

				var response map[string]interface{}
				err := testutil.ParseJSONResponse(w.Body, &response)
				Expect(err).ToNot(HaveOccurred())

				Expect(response["error"]).To(Equal("database error"))
			})

			It("should handle network errors gracefully", func() {
				mockPostgresService.EXPECT().
					GetUserAlbums(gomock.Any(), testFirebaseUID).
					Return(nil, errors.New("network unreachable"))

				req := httptest.NewRequest(http.MethodGet, "/legacy/albums", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusInternalServerError))
			})

			It("should handle permission denied errors", func() {
				mockPostgresService.EXPECT().
					GetUserAlbums(gomock.Any(), testFirebaseUID).
					Return(nil, errors.New("permission denied on table albums"))

				req := httptest.NewRequest(http.MethodGet, "/legacy/albums", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusInternalServerError))
			})
		})

		Context("when user is not authenticated", func() {
			BeforeEach(func() {
				router.GET("/legacy/albums", legacyHandler.GetUserAlbums)
			})

			It("should return unauthorized when firebase_uid is missing", func() {
				req := httptest.NewRequest(http.MethodGet, "/legacy/albums", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusUnauthorized))

				var response map[string]interface{}
				err := testutil.ParseJSONResponse(w.Body, &response)
				Expect(err).ToNot(HaveOccurred())

				Expect(response).To(HaveKey("error"))
				Expect(response["error"]).To(Equal("authentication required"))
			})

			It("should return unauthorized when firebase_uid is empty", func() {
				router = gin.New()
				router.GET("/legacy/albums", func(c *gin.Context) {
					c.Set("firebase_uid", "")
					legacyHandler.GetUserAlbums(c)
				})

				req := httptest.NewRequest(http.MethodGet, "/legacy/albums", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusUnauthorized))
			})
		})
	})
})
