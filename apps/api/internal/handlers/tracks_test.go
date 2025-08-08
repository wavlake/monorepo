package handlers_test

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/golang/mock/gomock"

	"github.com/wavlake/monorepo/internal/handlers"
	"github.com/wavlake/monorepo/internal/models"
	"github.com/wavlake/monorepo/tests/mocks"
	"github.com/wavlake/monorepo/tests/testutil"
)

var _ = Describe("TracksHandler", func() {
	var (
		ctrl                  *gomock.Controller
		mockNostrTrackService *mocks.MockNostrTrackServiceInterface
		mockProcessingService *mocks.MockProcessingServiceInterface
		mockAudioProcessor    *mocks.MockAudioProcessorInterface
		tracksHandler         *handlers.TracksHandler
		testFirebaseUID       string
		testPubkey           string
		testTrackID          string
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockNostrTrackService = mocks.NewMockNostrTrackServiceInterface(ctrl)
		mockProcessingService = mocks.NewMockProcessingServiceInterface(ctrl)
		mockAudioProcessor = mocks.NewMockAudioProcessorInterface(ctrl)
		tracksHandler = handlers.NewTracksHandler(mockNostrTrackService, mockProcessingService, mockAudioProcessor)
		testFirebaseUID = testutil.TestFirebaseUID
		testPubkey = testutil.TestPubkey
		testTrackID = testutil.TestTrackID
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Describe("CreateTrackNostr", func() {
		Context("when all required authentication is present", func() {
			It("should successfully create a track", func() {
				c, w := testutil.SetupGinTestContext("POST", "/v1/tracks", 
					testutil.ValidCreateTrackRequest())
				testutil.SetAuthContext(c, testFirebaseUID, testPubkey)

				expectedTrack := testutil.ValidNostrTrack()
				
				mockAudioProcessor.EXPECT().
					IsFormatSupported("mp3").
					Return(true)

				mockNostrTrackService.EXPECT().
					CreateTrack(c.Request.Context(), testPubkey, testFirebaseUID, "mp3").
					Return(expectedTrack, nil)

				tracksHandler.CreateTrackNostr(c)

				Expect(w.Code).To(Equal(http.StatusOK))
				response := testutil.AssertJSONResponse(w, http.StatusOK)
				Expect(response["success"]).To(BeTrue())
				
				data, ok := response["data"].(map[string]interface{})
				Expect(ok).To(BeTrue())
				Expect(data["id"]).To(Equal(expectedTrack.ID))
				Expect(data["pubkey"]).To(Equal(expectedTrack.Pubkey))
				Expect(data["extension"]).To(Equal(expectedTrack.Extension))
			})

			It("should handle extension with leading dot", func() {
				c, w := testutil.SetupGinTestContext("POST", "/v1/tracks", map[string]interface{}{
					"extension": ".wav",
				})
				testutil.SetAuthContext(c, testFirebaseUID, testPubkey)

				expectedTrack := testutil.ValidNostrTrack()
				expectedTrack.Extension = "wav"
				
				mockAudioProcessor.EXPECT().
					IsFormatSupported(".wav").
					Return(true)

				mockNostrTrackService.EXPECT().
					CreateTrack(c.Request.Context(), testPubkey, testFirebaseUID, "wav").
					Return(expectedTrack, nil)

				tracksHandler.CreateTrackNostr(c)

				Expect(w.Code).To(Equal(http.StatusOK))
				response := testutil.AssertJSONResponse(w, http.StatusOK)
				Expect(response["success"]).To(BeTrue())
			})
		})

		Context("when request body is invalid", func() {
			It("should return bad request for missing extension", func() {
				c, w := testutil.SetupGinTestContext("POST", "/v1/tracks", map[string]interface{}{
					// Missing required extension field
				})
				testutil.SetAuthContext(c, testFirebaseUID, testPubkey)

				tracksHandler.CreateTrackNostr(c)

				Expect(w.Code).To(Equal(http.StatusBadRequest))
				response := testutil.AssertJSONResponse(w, http.StatusBadRequest)
				Expect(response["success"]).To(BeFalse())
				Expect(response["error"]).To(Equal("extension field is required"))
			})

			It("should return bad request for unsupported format", func() {
				c, w := testutil.SetupGinTestContext("POST", "/v1/tracks", map[string]interface{}{
					"extension": "txt",
				})
				testutil.SetAuthContext(c, testFirebaseUID, testPubkey)

				mockAudioProcessor.EXPECT().
					IsFormatSupported("txt").
					Return(false)

				tracksHandler.CreateTrackNostr(c)

				Expect(w.Code).To(Equal(http.StatusBadRequest))
				response := testutil.AssertJSONResponse(w, http.StatusBadRequest)
				Expect(response["success"]).To(BeFalse())
				Expect(response["error"]).To(Equal("unsupported audio format"))
			})

			It("should handle malformed JSON gracefully", func() {
				c, w := testutil.SetupGinTestContext("POST", "/v1/tracks", "{malformed json")
				testutil.SetAuthContext(c, testFirebaseUID, testPubkey)

				tracksHandler.CreateTrackNostr(c)

				Expect(w.Code).To(Equal(http.StatusBadRequest))
				response := testutil.AssertJSONResponse(w, http.StatusBadRequest)
				Expect(response["success"]).To(BeFalse())
				Expect(response["error"]).To(Equal("extension field is required"))
			})

			It("should handle empty extension string", func() {
				c, w := testutil.SetupGinTestContext("POST", "/v1/tracks", map[string]interface{}{
					"extension": "",
				})
				testutil.SetAuthContext(c, testFirebaseUID, testPubkey)

				tracksHandler.CreateTrackNostr(c)

				Expect(w.Code).To(Equal(http.StatusBadRequest))
				response := testutil.AssertJSONResponse(w, http.StatusBadRequest)
				Expect(response["success"]).To(BeFalse())
				Expect(response["error"]).To(Equal("extension field is required"))
			})

			It("should handle very long extension string", func() {
				c, w := testutil.SetupGinTestContext("POST", "/v1/tracks", map[string]interface{}{
					"extension": strings.Repeat("a", 1000), // Very long extension
				})
				testutil.SetAuthContext(c, testFirebaseUID, testPubkey)

				mockAudioProcessor.EXPECT().
					IsFormatSupported(strings.Repeat("a", 1000)).
					Return(false)

				tracksHandler.CreateTrackNostr(c)

				Expect(w.Code).To(Equal(http.StatusBadRequest))
				response := testutil.AssertJSONResponse(w, http.StatusBadRequest)
				Expect(response["success"]).To(BeFalse())
				Expect(response["error"]).To(Equal("unsupported audio format"))
			})

			It("should handle case-sensitive extension validation", func() {
				c, w := testutil.SetupGinTestContext("POST", "/v1/tracks", map[string]interface{}{
					"extension": "MP3", // Uppercase version
				})
				testutil.SetAuthContext(c, testFirebaseUID, testPubkey)

				mockAudioProcessor.EXPECT().
					IsFormatSupported("MP3").
					Return(false) // Assuming case-sensitive validation

				tracksHandler.CreateTrackNostr(c)

				Expect(w.Code).To(Equal(http.StatusBadRequest))
				response := testutil.AssertJSONResponse(w, http.StatusBadRequest)
				Expect(response["success"]).To(BeFalse())
				Expect(response["error"]).To(Equal("unsupported audio format"))
			})
		})

		Context("when Nostr authentication is missing", func() {
			It("should return unauthorized error", func() {
				c, w := testutil.SetupGinTestContext("POST", "/v1/tracks", 
					testutil.ValidCreateTrackRequest())
				testutil.SetAuthContext(c, testFirebaseUID, "") // Missing Nostr pubkey

				mockAudioProcessor.EXPECT().
					IsFormatSupported("mp3").
					Return(true)

				tracksHandler.CreateTrackNostr(c)

				Expect(w.Code).To(Equal(http.StatusUnauthorized))
				response := testutil.AssertJSONResponse(w, http.StatusUnauthorized)
				Expect(response["success"]).To(BeFalse())
				Expect(response["error"]).To(Equal("authentication required"))
			})
		})

		Context("when Firebase authentication is missing", func() {
			It("should return unauthorized error", func() {
				c, w := testutil.SetupGinTestContext("POST", "/v1/tracks", 
					testutil.ValidCreateTrackRequest())
				testutil.SetAuthContext(c, "", testPubkey) // Missing Firebase UID

				mockAudioProcessor.EXPECT().
					IsFormatSupported("mp3").
					Return(true)

				tracksHandler.CreateTrackNostr(c)

				Expect(w.Code).To(Equal(http.StatusUnauthorized))
				response := testutil.AssertJSONResponse(w, http.StatusUnauthorized)
				Expect(response["success"]).To(BeFalse())
				Expect(response["error"]).To(Equal("user account not found"))
			})
		})

		Context("when authentication context has invalid types", func() {
			It("should return internal server error for invalid pubkey type", func() {
				c, w := testutil.SetupGinTestContext("POST", "/v1/tracks", 
					testutil.ValidCreateTrackRequest())
				c.Set("pubkey", 123) // Invalid type
				c.Set("firebase_uid", testFirebaseUID)

				mockAudioProcessor.EXPECT().
					IsFormatSupported("mp3").
					Return(true)

				tracksHandler.CreateTrackNostr(c)

				Expect(w.Code).To(Equal(http.StatusInternalServerError))
				response := testutil.AssertJSONResponse(w, http.StatusInternalServerError)
				Expect(response["success"]).To(BeFalse())
				Expect(response["error"]).To(Equal("invalid pubkey format"))
			})

			It("should return internal server error for invalid firebase_uid type", func() {
				c, w := testutil.SetupGinTestContext("POST", "/v1/tracks", 
					testutil.ValidCreateTrackRequest())
				c.Set("pubkey", testPubkey)
				c.Set("firebase_uid", 123) // Invalid type

				mockAudioProcessor.EXPECT().
					IsFormatSupported("mp3").
					Return(true)

				tracksHandler.CreateTrackNostr(c)

				Expect(w.Code).To(Equal(http.StatusInternalServerError))
				response := testutil.AssertJSONResponse(w, http.StatusInternalServerError)
				Expect(response["success"]).To(BeFalse())
				Expect(response["error"]).To(Equal("invalid user ID format"))
			})
		})

		Context("when service returns error", func() {
			It("should return internal server error", func() {
				c, w := testutil.SetupGinTestContext("POST", "/v1/tracks", 
					testutil.ValidCreateTrackRequest())
				testutil.SetAuthContext(c, testFirebaseUID, testPubkey)

				expectedError := errors.New("database connection failed")
				
				mockAudioProcessor.EXPECT().
					IsFormatSupported("mp3").
					Return(true)

				mockNostrTrackService.EXPECT().
					CreateTrack(c.Request.Context(), testPubkey, testFirebaseUID, "mp3").
					Return(nil, expectedError)

				tracksHandler.CreateTrackNostr(c)

				Expect(w.Code).To(Equal(http.StatusInternalServerError))
				response := testutil.AssertJSONResponse(w, http.StatusInternalServerError)
				Expect(response["success"]).To(BeFalse())
				Expect(response["error"]).To(Equal("failed to create track"))
			})
		})
	})

	Describe("GetMyTracks", func() {
		Context("when Nostr authentication is present", func() {
			It("should return user's tracks", func() {
				c, w := testutil.SetupGinTestContext("GET", "/v1/tracks/my", nil)
				testutil.SetAuthContext(c, "", testPubkey) // Only Nostr auth required

				expectedTracks := testutil.ValidTracksList()
				
				mockNostrTrackService.EXPECT().
					GetTracksByPubkey(c.Request.Context(), testPubkey).
					Return(expectedTracks, nil)

				tracksHandler.GetMyTracks(c)

				Expect(w.Code).To(Equal(http.StatusOK))
				response := testutil.AssertJSONResponse(w, http.StatusOK)
				Expect(response["success"]).To(BeTrue())
				
				data, ok := response["data"].([]interface{})
				Expect(ok).To(BeTrue())
				Expect(data).To(HaveLen(2))
			})

			It("should return empty array when user has no tracks", func() {
				c, w := testutil.SetupGinTestContext("GET", "/v1/tracks/my", nil)
				testutil.SetAuthContext(c, "", testPubkey)

				mockNostrTrackService.EXPECT().
					GetTracksByPubkey(c.Request.Context(), testPubkey).
					Return([]*models.NostrTrack{}, nil)

				tracksHandler.GetMyTracks(c)

				Expect(w.Code).To(Equal(http.StatusOK))
				response := testutil.AssertJSONResponse(w, http.StatusOK)
				Expect(response["success"]).To(BeTrue())
				
				// Handle the case where empty slice may be serialized as null
				data := response["data"]
				if data == nil {
					// This is acceptable for empty results
					Expect(data).To(BeNil())
				} else {
					// If it's an array, it should be empty
					dataArray, ok := data.([]interface{})
					Expect(ok).To(BeTrue())
					Expect(dataArray).To(HaveLen(0))
				}
			})
		})

		Context("when Nostr authentication is missing", func() {
			It("should return unauthorized error", func() {
				c, w := testutil.SetupGinTestContext("GET", "/v1/tracks/my", nil)
				// No auth context set

				tracksHandler.GetMyTracks(c)

				Expect(w.Code).To(Equal(http.StatusUnauthorized))
				response := testutil.AssertJSONResponse(w, http.StatusUnauthorized)
				Expect(response["success"]).To(BeFalse())
				Expect(response["error"]).To(Equal("authentication required"))
			})
		})

		Context("when authentication context has invalid type", func() {
			It("should return internal server error for invalid pubkey type", func() {
				c, w := testutil.SetupGinTestContext("GET", "/v1/tracks/my", nil)
				c.Set("pubkey", 123) // Invalid type

				tracksHandler.GetMyTracks(c)

				Expect(w.Code).To(Equal(http.StatusInternalServerError))
				response := testutil.AssertJSONResponse(w, http.StatusInternalServerError)
				Expect(response["success"]).To(BeFalse())
				Expect(response["error"]).To(Equal("invalid pubkey format"))
			})
		})

		Context("when service returns error", func() {
			It("should return internal server error", func() {
				c, w := testutil.SetupGinTestContext("GET", "/v1/tracks/my", nil)
				testutil.SetAuthContext(c, "", testPubkey)

				expectedError := errors.New("database query failed")
				
				mockNostrTrackService.EXPECT().
					GetTracksByPubkey(c.Request.Context(), testPubkey).
					Return(nil, expectedError)

				tracksHandler.GetMyTracks(c)

				Expect(w.Code).To(Equal(http.StatusInternalServerError))
				response := testutil.AssertJSONResponse(w, http.StatusInternalServerError)
				Expect(response["success"]).To(BeFalse())
				Expect(response["error"]).To(Equal("failed to retrieve tracks"))
			})
		})
	})

	Describe("GetTrack", func() {
		Context("when track ID is provided", func() {
			It("should return the track", func() {
				c, w := testutil.SetupGinTestContext("GET", "/v1/tracks/:trackId", nil)
				c.Params = []gin.Param{{Key: "trackId", Value: testTrackID}}

				expectedTrack := testutil.ValidNostrTrack()
				
				mockNostrTrackService.EXPECT().
					GetTrack(c.Request.Context(), testTrackID).
					Return(expectedTrack, nil)

				tracksHandler.GetTrack(c)

				Expect(w.Code).To(Equal(http.StatusOK))
				response := testutil.AssertJSONResponse(w, http.StatusOK)
				Expect(response["success"]).To(BeTrue())
				
				data, ok := response["data"].(map[string]interface{})
				Expect(ok).To(BeTrue())
				Expect(data["id"]).To(Equal(expectedTrack.ID))
			})
		})

		Context("when track ID is missing", func() {
			It("should return bad request error", func() {
				c, w := testutil.SetupGinTestContext("GET", "/v1/tracks/:trackId", nil)
				c.Params = []gin.Param{{Key: "trackId", Value: ""}}

				tracksHandler.GetTrack(c)

				Expect(w.Code).To(Equal(http.StatusBadRequest))
				response := testutil.AssertJSONResponse(w, http.StatusBadRequest)
				Expect(response["error"]).To(Equal("track ID is required"))
			})
		})

		Context("when track is not found", func() {
			It("should return not found error", func() {
				c, w := testutil.SetupGinTestContext("GET", "/v1/tracks/:trackId", nil)
				c.Params = []gin.Param{{Key: "trackId", Value: testTrackID}}

				expectedError := errors.New("track not found")
				
				mockNostrTrackService.EXPECT().
					GetTrack(c.Request.Context(), testTrackID).
					Return(nil, expectedError)

				tracksHandler.GetTrack(c)

				Expect(w.Code).To(Equal(http.StatusNotFound))
				response := testutil.AssertJSONResponse(w, http.StatusNotFound)
				Expect(response["error"]).To(Equal("track not found"))
			})
		})
	})

	Describe("DeleteTrack", func() {
		Context("when user owns the track", func() {
			It("should successfully delete the track", func() {
				c, w := testutil.SetupGinTestContext("DELETE", "/v1/tracks/:trackId", nil)
				c.Params = []gin.Param{{Key: "trackId", Value: testTrackID}}
				testutil.SetAuthContext(c, "", testPubkey)

				expectedTrack := testutil.ValidNostrTrack()
				
				mockNostrTrackService.EXPECT().
					GetTrack(c.Request.Context(), testTrackID).
					Return(expectedTrack, nil)

				mockNostrTrackService.EXPECT().
					DeleteTrack(c.Request.Context(), testTrackID).
					Return(nil)

				tracksHandler.DeleteTrack(c)

				Expect(w.Code).To(Equal(http.StatusOK))
				response := testutil.AssertJSONResponse(w, http.StatusOK)
				Expect(response["success"]).To(BeTrue())
				Expect(response["message"]).To(Equal("track deleted successfully"))
			})
		})

		Context("when track ID is missing", func() {
			It("should return bad request error", func() {
				c, w := testutil.SetupGinTestContext("DELETE", "/v1/tracks/:trackId", nil)
				c.Params = []gin.Param{{Key: "trackId", Value: ""}}
				testutil.SetAuthContext(c, "", testPubkey)

				tracksHandler.DeleteTrack(c)

				Expect(w.Code).To(Equal(http.StatusBadRequest))
				response := testutil.AssertJSONResponse(w, http.StatusBadRequest)
				Expect(response["error"]).To(Equal("track ID is required"))
			})
		})

		Context("when Nostr authentication is missing", func() {
			It("should return unauthorized error", func() {
				c, w := testutil.SetupGinTestContext("DELETE", "/v1/tracks/:trackId", nil)
				c.Params = []gin.Param{{Key: "trackId", Value: testTrackID}}
				// No auth context set

				tracksHandler.DeleteTrack(c)

				Expect(w.Code).To(Equal(http.StatusUnauthorized))
				response := testutil.AssertJSONResponse(w, http.StatusUnauthorized)
				Expect(response["error"]).To(Equal("authentication required"))
			})
		})

		Context("when track is not found", func() {
			It("should return not found error", func() {
				c, w := testutil.SetupGinTestContext("DELETE", "/v1/tracks/:trackId", nil)
				c.Params = []gin.Param{{Key: "trackId", Value: testTrackID}}
				testutil.SetAuthContext(c, "", testPubkey)

				expectedError := errors.New("track not found")
				
				mockNostrTrackService.EXPECT().
					GetTrack(c.Request.Context(), testTrackID).
					Return(nil, expectedError)

				tracksHandler.DeleteTrack(c)

				Expect(w.Code).To(Equal(http.StatusNotFound))
				response := testutil.AssertJSONResponse(w, http.StatusNotFound)
				Expect(response["error"]).To(Equal("track not found"))
			})
		})

		Context("when user does not own the track", func() {
			It("should return forbidden error", func() {
				c, w := testutil.SetupGinTestContext("DELETE", "/v1/tracks/:trackId", nil)
				c.Params = []gin.Param{{Key: "trackId", Value: testTrackID}}
				testutil.SetAuthContext(c, "", testPubkey)

				differentTrack := testutil.ValidNostrTrack()
				differentTrack.Pubkey = "different-pubkey"
				
				mockNostrTrackService.EXPECT().
					GetTrack(c.Request.Context(), testTrackID).
					Return(differentTrack, nil)

				tracksHandler.DeleteTrack(c)

				Expect(w.Code).To(Equal(http.StatusForbidden))
				response := testutil.AssertJSONResponse(w, http.StatusForbidden)
				Expect(response["error"]).To(Equal("you can only delete your own tracks"))
			})
		})

		Context("when delete operation fails", func() {
			It("should return internal server error", func() {
				c, w := testutil.SetupGinTestContext("DELETE", "/v1/tracks/:trackId", nil)
				c.Params = []gin.Param{{Key: "trackId", Value: testTrackID}}
				testutil.SetAuthContext(c, "", testPubkey)

				expectedTrack := testutil.ValidNostrTrack()
				expectedError := errors.New("delete operation failed")
				
				mockNostrTrackService.EXPECT().
					GetTrack(c.Request.Context(), testTrackID).
					Return(expectedTrack, nil)

				mockNostrTrackService.EXPECT().
					DeleteTrack(c.Request.Context(), testTrackID).
					Return(expectedError)

				tracksHandler.DeleteTrack(c)

				Expect(w.Code).To(Equal(http.StatusInternalServerError))
				response := testutil.AssertJSONResponse(w, http.StatusInternalServerError)
				Expect(response["error"]).To(Equal("failed to delete track"))
			})
		})
	})
})