package services_test

import (
	"context"
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/golang/mock/gomock"

	"github.com/wavlake/monorepo/internal/models"
	"github.com/wavlake/monorepo/tests/mocks"
	"github.com/wavlake/monorepo/tests/testutil"
)

var _ = Describe("PostgresServiceInterface", func() {
	var (
		ctrl                  *gomock.Controller
		mockPostgresService   *mocks.MockPostgresServiceInterface
		ctx                   context.Context
		testFirebaseUID       string
		testArtistID          string
		testAlbumID           string
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockPostgresService = mocks.NewMockPostgresServiceInterface(ctrl)
		ctx = context.Background()
		testFirebaseUID = testutil.TestFirebaseUID
		testArtistID = testutil.TestArtistID
		testAlbumID = testutil.TestAlbumID
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Describe("GetUserByFirebaseUID", func() {
		Context("when user exists and is not locked", func() {
			It("should return the user with all fields populated", func() {
				expectedUser := testutil.ValidLegacyUser()
				
				mockPostgresService.EXPECT().
					GetUserByFirebaseUID(ctx, testFirebaseUID).
					Return(&expectedUser, nil)

				user, err := mockPostgresService.GetUserByFirebaseUID(ctx, testFirebaseUID)
				
				Expect(err).ToNot(HaveOccurred())
				Expect(user).ToNot(BeNil())
				Expect(user.ID).To(Equal(testFirebaseUID))
				Expect(user.Name).To(Equal("Test User"))
				Expect(user.LightningAddress).To(Equal("test@wavlake.com"))
				Expect(user.MSatBalance).To(Equal(int64(1000000)))
				Expect(user.IsLocked).To(BeFalse())
			})

			It("should handle nullable fields correctly", func() {
				expectedUser := testutil.ValidLegacyUser()
				expectedUser.LightningAddress = ""
				expectedUser.ArtworkURL = ""
				expectedUser.MSatBalance = 0
				
				mockPostgresService.EXPECT().
					GetUserByFirebaseUID(ctx, testFirebaseUID).
					Return(&expectedUser, nil)

				user, err := mockPostgresService.GetUserByFirebaseUID(ctx, testFirebaseUID)
				
				Expect(err).ToNot(HaveOccurred())
				Expect(user.LightningAddress).To(BeEmpty())
				Expect(user.ArtworkURL).To(BeEmpty())
				Expect(user.MSatBalance).To(Equal(int64(0)))
			})
		})

		Context("when user does not exist", func() {
			It("should return user not found error", func() {
				expectedError := errors.New("user not found")
				
				mockPostgresService.EXPECT().
					GetUserByFirebaseUID(ctx, testFirebaseUID).
					Return(nil, expectedError)

				user, err := mockPostgresService.GetUserByFirebaseUID(ctx, testFirebaseUID)
				
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("not found"))
				Expect(user).To(BeNil())
			})
		})

		Context("when user is locked", func() {
			It("should not return locked users", func() {
				expectedError := errors.New("user not found")
				
				mockPostgresService.EXPECT().
					GetUserByFirebaseUID(ctx, testFirebaseUID).
					Return(nil, expectedError)

				user, err := mockPostgresService.GetUserByFirebaseUID(ctx, testFirebaseUID)
				
				Expect(err).To(HaveOccurred())
				Expect(user).To(BeNil())
			})
		})

		Context("when database query fails", func() {
			It("should return database error", func() {
				expectedError := errors.New("failed to get user: database connection failed")
				
				mockPostgresService.EXPECT().
					GetUserByFirebaseUID(ctx, testFirebaseUID).
					Return(nil, expectedError)

				user, err := mockPostgresService.GetUserByFirebaseUID(ctx, testFirebaseUID)
				
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to get user"))
				Expect(user).To(BeNil())
			})
		})
	})

	Describe("GetUserTracks", func() {
		Context("when user has tracks", func() {
			It("should return list of tracks ordered by creation date", func() {
				expectedTracks := testutil.ValidLegacyTracksList()
				
				mockPostgresService.EXPECT().
					GetUserTracks(ctx, testFirebaseUID).
					Return(expectedTracks, nil)

				tracks, err := mockPostgresService.GetUserTracks(ctx, testFirebaseUID)
				
				Expect(err).ToNot(HaveOccurred())
				Expect(tracks).To(HaveLen(2))
				Expect(tracks[0].Title).To(Equal("Test Track"))
				Expect(tracks[1].Title).To(Equal("Test Track 2"))
				Expect(tracks[0].Deleted).To(BeFalse())
				Expect(tracks[1].Deleted).To(BeFalse())
			})

			It("should return tracks with all nullable fields handled", func() {
				tracks := testutil.ValidLegacyTracksList()
				tracks[0].RawURL = ""
				tracks[0].PlayCount = 0
				tracks[0].Lyrics = ""
				
				mockPostgresService.EXPECT().
					GetUserTracks(ctx, testFirebaseUID).
					Return(tracks, nil)

				result, err := mockPostgresService.GetUserTracks(ctx, testFirebaseUID)
				
				Expect(err).ToNot(HaveOccurred())
				Expect(result[0].RawURL).To(BeEmpty())
				Expect(result[0].PlayCount).To(Equal(0))
				Expect(result[0].Lyrics).To(BeEmpty())
			})

			It("should only return non-deleted tracks", func() {
				expectedTracks := []models.LegacyTrack{testutil.ValidLegacyTrack()}
				
				mockPostgresService.EXPECT().
					GetUserTracks(ctx, testFirebaseUID).
					Return(expectedTracks, nil)

				tracks, err := mockPostgresService.GetUserTracks(ctx, testFirebaseUID)
				
				Expect(err).ToNot(HaveOccurred())
				Expect(tracks).To(HaveLen(1))
				Expect(tracks[0].Deleted).To(BeFalse())
			})
		})

		Context("when user has no tracks", func() {
			It("should return empty slice", func() {
				mockPostgresService.EXPECT().
					GetUserTracks(ctx, testFirebaseUID).
					Return([]models.LegacyTrack{}, nil)

				tracks, err := mockPostgresService.GetUserTracks(ctx, testFirebaseUID)
				
				Expect(err).ToNot(HaveOccurred())
				Expect(tracks).To(BeEmpty())
			})
		})

		Context("when database query fails", func() {
			It("should return query error", func() {
				expectedError := errors.New("failed to query tracks: connection timeout")
				
				mockPostgresService.EXPECT().
					GetUserTracks(ctx, testFirebaseUID).
					Return(nil, expectedError)

				tracks, err := mockPostgresService.GetUserTracks(ctx, testFirebaseUID)
				
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to query tracks"))
				Expect(tracks).To(BeNil())
			})

			It("should handle row scanning errors", func() {
				expectedError := errors.New("failed to scan track: invalid data type")
				
				mockPostgresService.EXPECT().
					GetUserTracks(ctx, testFirebaseUID).
					Return(nil, expectedError)

				tracks, err := mockPostgresService.GetUserTracks(ctx, testFirebaseUID)
				
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("scan track"))
				Expect(tracks).To(BeNil())
			})

			It("should handle row iteration errors", func() {
				expectedError := errors.New("failed to iterate tracks: cursor error")
				
				mockPostgresService.EXPECT().
					GetUserTracks(ctx, testFirebaseUID).
					Return(nil, expectedError)

				tracks, err := mockPostgresService.GetUserTracks(ctx, testFirebaseUID)
				
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("iterate tracks"))
				Expect(tracks).To(BeNil())
			})
		})
	})

	Describe("GetUserArtists", func() {
		Context("when user has artists", func() {
			It("should return list of artists ordered by creation date", func() {
				expectedArtists := testutil.ValidLegacyArtistsList()
				
				mockPostgresService.EXPECT().
					GetUserArtists(ctx, testFirebaseUID).
					Return(expectedArtists, nil)

				artists, err := mockPostgresService.GetUserArtists(ctx, testFirebaseUID)
				
				Expect(err).ToNot(HaveOccurred())
				Expect(artists).To(HaveLen(2))
				Expect(artists[0].Name).To(Equal("Test Artist"))
				Expect(artists[1].Name).To(Equal("Test Artist 2"))
				Expect(artists[0].UserID).To(Equal(testFirebaseUID))
				Expect(artists[0].Deleted).To(BeFalse())
			})

			It("should handle nullable fields correctly", func() {
				artists := testutil.ValidLegacyArtistsList()
				artists[0].Bio = ""
				artists[0].Twitter = ""
				artists[0].ArtworkURL = ""
				artists[0].MSatTotal = 0
				
				mockPostgresService.EXPECT().
					GetUserArtists(ctx, testFirebaseUID).
					Return(artists, nil)

				result, err := mockPostgresService.GetUserArtists(ctx, testFirebaseUID)
				
				Expect(err).ToNot(HaveOccurred())
				Expect(result[0].Bio).To(BeEmpty())
				Expect(result[0].Twitter).To(BeEmpty())
				Expect(result[0].ArtworkURL).To(BeEmpty())
				Expect(result[0].MSatTotal).To(Equal(int64(0)))
			})
		})

		Context("when user has no artists", func() {
			It("should return empty slice", func() {
				mockPostgresService.EXPECT().
					GetUserArtists(ctx, testFirebaseUID).
					Return([]models.LegacyArtist{}, nil)

				artists, err := mockPostgresService.GetUserArtists(ctx, testFirebaseUID)
				
				Expect(err).ToNot(HaveOccurred())
				Expect(artists).To(BeEmpty())
			})
		})

		Context("when database query fails", func() {
			It("should return query error", func() {
				expectedError := errors.New("failed to query artists: table not found")
				
				mockPostgresService.EXPECT().
					GetUserArtists(ctx, testFirebaseUID).
					Return(nil, expectedError)

				artists, err := mockPostgresService.GetUserArtists(ctx, testFirebaseUID)
				
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to query artists"))
				Expect(artists).To(BeNil())
			})

			It("should handle iteration errors", func() {
				expectedError := errors.New("failed to iterate artists: connection lost")
				
				mockPostgresService.EXPECT().
					GetUserArtists(ctx, testFirebaseUID).
					Return(nil, expectedError)

				artists, err := mockPostgresService.GetUserArtists(ctx, testFirebaseUID)
				
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("iterate artists"))
				Expect(artists).To(BeNil())
			})
		})
	})

	Describe("GetUserAlbums", func() {
		Context("when user has albums", func() {
			It("should return list of albums ordered by creation date", func() {
				expectedAlbums := testutil.ValidLegacyAlbumsList()
				
				mockPostgresService.EXPECT().
					GetUserAlbums(ctx, testFirebaseUID).
					Return(expectedAlbums, nil)

				albums, err := mockPostgresService.GetUserAlbums(ctx, testFirebaseUID)
				
				Expect(err).ToNot(HaveOccurred())
				Expect(albums).To(HaveLen(2))
				Expect(albums[0].Title).To(Equal("Test Album"))
				Expect(albums[1].Title).To(Equal("Test Album 2"))
				Expect(albums[0].ArtistID).To(Equal(testArtistID))
				Expect(albums[0].Deleted).To(BeFalse())
			})

			It("should handle nullable and boolean fields correctly", func() {
				albums := testutil.ValidLegacyAlbumsList()
				albums[0].Description = ""
				albums[0].ArtworkURL = ""
				albums[0].GenreID = 0
				albums[0].IsDraft = true
				albums[0].IsSingle = true
				
				mockPostgresService.EXPECT().
					GetUserAlbums(ctx, testFirebaseUID).
					Return(albums, nil)

				result, err := mockPostgresService.GetUserAlbums(ctx, testFirebaseUID)
				
				Expect(err).ToNot(HaveOccurred())
				Expect(result[0].Description).To(BeEmpty())
				Expect(result[0].ArtworkURL).To(BeEmpty())
				Expect(result[0].GenreID).To(Equal(0))
				Expect(result[0].IsDraft).To(BeTrue())
				Expect(result[0].IsSingle).To(BeTrue())
			})
		})

		Context("when user has no albums", func() {
			It("should return empty slice", func() {
				mockPostgresService.EXPECT().
					GetUserAlbums(ctx, testFirebaseUID).
					Return([]models.LegacyAlbum{}, nil)

				albums, err := mockPostgresService.GetUserAlbums(ctx, testFirebaseUID)
				
				Expect(err).ToNot(HaveOccurred())
				Expect(albums).To(BeEmpty())
			})
		})

		Context("when database query fails", func() {
			It("should return query error", func() {
				expectedError := errors.New("failed to query albums: permission denied")
				
				mockPostgresService.EXPECT().
					GetUserAlbums(ctx, testFirebaseUID).
					Return(nil, expectedError)

				albums, err := mockPostgresService.GetUserAlbums(ctx, testFirebaseUID)
				
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to query albums"))
				Expect(albums).To(BeNil())
			})
		})
	})

	Describe("GetTracksByArtist", func() {
		Context("when artist has tracks", func() {
			It("should return tracks ordered by track order and creation date", func() {
				expectedTracks := testutil.ValidLegacyTracksList()
				
				mockPostgresService.EXPECT().
					GetTracksByArtist(ctx, testArtistID).
					Return(expectedTracks, nil)

				tracks, err := mockPostgresService.GetTracksByArtist(ctx, testArtistID)
				
				Expect(err).ToNot(HaveOccurred())
				Expect(tracks).To(HaveLen(2))
				Expect(tracks[0].ArtistID).To(Equal(testArtistID))
				Expect(tracks[1].ArtistID).To(Equal(testArtistID))
				Expect(tracks[0].Order).To(Equal(1))
				Expect(tracks[1].Order).To(Equal(2))
			})

			It("should only return non-deleted tracks", func() {
				expectedTracks := []models.LegacyTrack{testutil.ValidLegacyTrack()}
				
				mockPostgresService.EXPECT().
					GetTracksByArtist(ctx, testArtistID).
					Return(expectedTracks, nil)

				tracks, err := mockPostgresService.GetTracksByArtist(ctx, testArtistID)
				
				Expect(err).ToNot(HaveOccurred())
				Expect(tracks).To(HaveLen(1))
				Expect(tracks[0].Deleted).To(BeFalse())
			})
		})

		Context("when artist has no tracks", func() {
			It("should return empty slice", func() {
				mockPostgresService.EXPECT().
					GetTracksByArtist(ctx, testArtistID).
					Return([]models.LegacyTrack{}, nil)

				tracks, err := mockPostgresService.GetTracksByArtist(ctx, testArtistID)
				
				Expect(err).ToNot(HaveOccurred())
				Expect(tracks).To(BeEmpty())
			})
		})

		Context("when database query fails", func() {
			It("should return query error", func() {
				expectedError := errors.New("failed to query tracks by artist: invalid artist ID")
				
				mockPostgresService.EXPECT().
					GetTracksByArtist(ctx, testArtistID).
					Return(nil, expectedError)

				tracks, err := mockPostgresService.GetTracksByArtist(ctx, testArtistID)
				
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("tracks by artist"))
				Expect(tracks).To(BeNil())
			})
		})
	})

	Describe("GetTracksByAlbum", func() {
		Context("when album has tracks", func() {
			It("should return tracks ordered by track order and creation date", func() {
				expectedTracks := testutil.ValidLegacyTracksList()
				expectedTracks[0].AlbumID = testAlbumID
				expectedTracks[1].AlbumID = testAlbumID
				
				mockPostgresService.EXPECT().
					GetTracksByAlbum(ctx, testAlbumID).
					Return(expectedTracks, nil)

				tracks, err := mockPostgresService.GetTracksByAlbum(ctx, testAlbumID)
				
				Expect(err).ToNot(HaveOccurred())
				Expect(tracks).To(HaveLen(2))
				Expect(tracks[0].AlbumID).To(Equal(testAlbumID))
				Expect(tracks[1].AlbumID).To(Equal(testAlbumID))
				Expect(tracks[0].Order).To(Equal(1))
				Expect(tracks[1].Order).To(Equal(2))
			})

			It("should handle all track properties correctly", func() {
				tracks := testutil.ValidLegacyTracksList()
				tracks[0].IsProcessing = true
				tracks[0].IsDraft = true
				tracks[0].IsExplicit = true
				tracks[0].CompressorError = true
				
				mockPostgresService.EXPECT().
					GetTracksByAlbum(ctx, testAlbumID).
					Return(tracks, nil)

				result, err := mockPostgresService.GetTracksByAlbum(ctx, testAlbumID)
				
				Expect(err).ToNot(HaveOccurred())
				Expect(result[0].IsProcessing).To(BeTrue())
				Expect(result[0].IsDraft).To(BeTrue())
				Expect(result[0].IsExplicit).To(BeTrue())
				Expect(result[0].CompressorError).To(BeTrue())
			})
		})

		Context("when album has no tracks", func() {
			It("should return empty slice", func() {
				mockPostgresService.EXPECT().
					GetTracksByAlbum(ctx, testAlbumID).
					Return([]models.LegacyTrack{}, nil)

				tracks, err := mockPostgresService.GetTracksByAlbum(ctx, testAlbumID)
				
				Expect(err).ToNot(HaveOccurred())
				Expect(tracks).To(BeEmpty())
			})
		})

		Context("when database query fails", func() {
			It("should return query error", func() {
				expectedError := errors.New("failed to query tracks by album: album not found")
				
				mockPostgresService.EXPECT().
					GetTracksByAlbum(ctx, testAlbumID).
					Return(nil, expectedError)

				tracks, err := mockPostgresService.GetTracksByAlbum(ctx, testAlbumID)
				
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("tracks by album"))
				Expect(tracks).To(BeNil())
			})

			It("should handle scanning errors", func() {
				expectedError := errors.New("failed to scan track: data type mismatch")
				
				mockPostgresService.EXPECT().
					GetTracksByAlbum(ctx, testAlbumID).
					Return(nil, expectedError)

				tracks, err := mockPostgresService.GetTracksByAlbum(ctx, testAlbumID)
				
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("scan track"))
				Expect(tracks).To(BeNil())
			})

			It("should handle iteration errors", func() {
				expectedError := errors.New("failed to iterate tracks: connection interrupted")
				
				mockPostgresService.EXPECT().
					GetTracksByAlbum(ctx, testAlbumID).
					Return(nil, expectedError)

				tracks, err := mockPostgresService.GetTracksByAlbum(ctx, testAlbumID)
				
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("iterate tracks"))
				Expect(tracks).To(BeNil())
			})
		})
	})
})