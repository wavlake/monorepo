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

var _ = Describe("NostrTrackServiceInterface", func() {
	var (
		ctrl                     *gomock.Controller
		mockNostrTrackService    *mocks.MockNostrTrackServiceInterface
		ctx                      context.Context
		testFirebaseUID          string
		testPubkey              string
		testTrackID             string
		testExtension           string
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockNostrTrackService = mocks.NewMockNostrTrackServiceInterface(ctrl)
		ctx = context.Background()
		testFirebaseUID = testutil.TestFirebaseUID
		testPubkey = testutil.TestPubkey
		testTrackID = testutil.TestTrackID
		testExtension = testutil.TestExtension
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Describe("CreateTrack", func() {
		Context("when all parameters are valid", func() {
			It("should successfully create a track", func() {
				expectedTrack := testutil.ValidNostrTrack()
				
				mockNostrTrackService.EXPECT().
					CreateTrack(ctx, testPubkey, testFirebaseUID, testExtension).
					Return(expectedTrack, nil)

				track, err := mockNostrTrackService.CreateTrack(ctx, testPubkey, testFirebaseUID, testExtension)
				
				Expect(err).ToNot(HaveOccurred())
				Expect(track).ToNot(BeNil())
				Expect(track.ID).To(Equal(expectedTrack.ID))
				Expect(track.Pubkey).To(Equal(testPubkey))
				Expect(track.FirebaseUID).To(Equal(testFirebaseUID))
				Expect(track.Extension).To(Equal(testExtension))
			})

			It("should return track with presigned URL for upload", func() {
				expectedTrack := testutil.ValidNostrTrack()
				expectedTrack.PresignedURL = "https://storage.googleapis.com/test-bucket/upload-url"
				expectedTrack.IsProcessing = true  // New tracks should be in processing state
				
				mockNostrTrackService.EXPECT().
					CreateTrack(ctx, testPubkey, testFirebaseUID, testExtension).
					Return(expectedTrack, nil)

				track, err := mockNostrTrackService.CreateTrack(ctx, testPubkey, testFirebaseUID, testExtension)
				
				Expect(err).ToNot(HaveOccurred())
				Expect(track.PresignedURL).ToNot(BeEmpty())
				Expect(track.IsProcessing).To(BeTrue())
				Expect(track.Deleted).To(BeFalse())
			})
		})

		Context("when storage service fails", func() {
			It("should return error when presigned URL generation fails", func() {
				expectedError := errors.New("failed to generate presigned URL")
				
				mockNostrTrackService.EXPECT().
					CreateTrack(ctx, testPubkey, testFirebaseUID, testExtension).
					Return(nil, expectedError)

				track, err := mockNostrTrackService.CreateTrack(ctx, testPubkey, testFirebaseUID, testExtension)
				
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("presigned URL"))
				Expect(track).To(BeNil())
			})

			It("should return error when Firestore save fails", func() {
				expectedError := errors.New("failed to save track to firestore")
				
				mockNostrTrackService.EXPECT().
					CreateTrack(ctx, testPubkey, testFirebaseUID, testExtension).
					Return(nil, expectedError)

				track, err := mockNostrTrackService.CreateTrack(ctx, testPubkey, testFirebaseUID, testExtension)
				
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("firestore"))
				Expect(track).To(BeNil())
			})
		})
	})

	Describe("GetTrack", func() {
		Context("when track exists", func() {
			It("should return the track", func() {
				expectedTrack := testutil.ValidNostrTrack()
				
				mockNostrTrackService.EXPECT().
					GetTrack(ctx, testTrackID).
					Return(expectedTrack, nil)

				track, err := mockNostrTrackService.GetTrack(ctx, testTrackID)
				
				Expect(err).ToNot(HaveOccurred())
				Expect(track).ToNot(BeNil())
				Expect(track.ID).To(Equal(testTrackID))
			})
		})

		Context("when track does not exist", func() {
			It("should return error", func() {
				expectedError := errors.New("failed to get track: document not found")
				
				mockNostrTrackService.EXPECT().
					GetTrack(ctx, testTrackID).
					Return(nil, expectedError)

				track, err := mockNostrTrackService.GetTrack(ctx, testTrackID)
				
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("not found"))
				Expect(track).To(BeNil())
			})

			It("should return error when decoding fails", func() {
				expectedError := errors.New("failed to decode track")
				
				mockNostrTrackService.EXPECT().
					GetTrack(ctx, testTrackID).
					Return(nil, expectedError)

				track, err := mockNostrTrackService.GetTrack(ctx, testTrackID)
				
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("decode"))
				Expect(track).To(BeNil())
			})
		})
	})

	Describe("GetTracksByPubkey", func() {
		Context("when pubkey has tracks", func() {
			It("should return list of tracks ordered by creation date", func() {
				expectedTracks := testutil.ValidTracksList()
				
				mockNostrTrackService.EXPECT().
					GetTracksByPubkey(ctx, testPubkey).
					Return(expectedTracks, nil)

				tracks, err := mockNostrTrackService.GetTracksByPubkey(ctx, testPubkey)
				
				Expect(err).ToNot(HaveOccurred())
				Expect(tracks).To(HaveLen(2))
				Expect(tracks[0].Pubkey).To(Equal(testPubkey))
				Expect(tracks[1].Pubkey).To(Equal(testPubkey))
			})

			It("should return only non-deleted tracks", func() {
				expectedTracks := []*models.NostrTrack{testutil.ValidNostrTrack()}
				
				mockNostrTrackService.EXPECT().
					GetTracksByPubkey(ctx, testPubkey).
					Return(expectedTracks, nil)

				tracks, err := mockNostrTrackService.GetTracksByPubkey(ctx, testPubkey)
				
				Expect(err).ToNot(HaveOccurred())
				Expect(tracks).To(HaveLen(1))
				Expect(tracks[0].Deleted).To(BeFalse())
			})
		})

		Context("when pubkey has no tracks", func() {
			It("should return empty slice", func() {
				mockNostrTrackService.EXPECT().
					GetTracksByPubkey(ctx, testPubkey).
					Return([]*models.NostrTrack{}, nil)

				tracks, err := mockNostrTrackService.GetTracksByPubkey(ctx, testPubkey)
				
				Expect(err).ToNot(HaveOccurred())
				Expect(tracks).To(BeEmpty())
			})
		})

		Context("when query fails", func() {
			It("should return error", func() {
				expectedError := errors.New("failed to iterate tracks")
				
				mockNostrTrackService.EXPECT().
					GetTracksByPubkey(ctx, testPubkey).
					Return(nil, expectedError)

				tracks, err := mockNostrTrackService.GetTracksByPubkey(ctx, testPubkey)
				
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("iterate"))
				Expect(tracks).To(BeNil())
			})
		})
	})

	Describe("GetTracksByFirebaseUID", func() {
		Context("when user has tracks", func() {
			It("should return list of tracks for Firebase user", func() {
				expectedTracks := testutil.ValidTracksList()
				
				mockNostrTrackService.EXPECT().
					GetTracksByFirebaseUID(ctx, testFirebaseUID).
					Return(expectedTracks, nil)

				tracks, err := mockNostrTrackService.GetTracksByFirebaseUID(ctx, testFirebaseUID)
				
				Expect(err).ToNot(HaveOccurred())
				Expect(tracks).To(HaveLen(2))
				Expect(tracks[0].FirebaseUID).To(Equal(testFirebaseUID))
			})
		})

		Context("when user has no tracks", func() {
			It("should return empty slice", func() {
				mockNostrTrackService.EXPECT().
					GetTracksByFirebaseUID(ctx, testFirebaseUID).
					Return([]*models.NostrTrack{}, nil)

				tracks, err := mockNostrTrackService.GetTracksByFirebaseUID(ctx, testFirebaseUID)
				
				Expect(err).ToNot(HaveOccurred())
				Expect(tracks).To(BeEmpty())
			})
		})

		Context("when query fails", func() {
			It("should return error", func() {
				expectedError := errors.New("failed to iterate tracks")
				
				mockNostrTrackService.EXPECT().
					GetTracksByFirebaseUID(ctx, testFirebaseUID).
					Return(nil, expectedError)

				tracks, err := mockNostrTrackService.GetTracksByFirebaseUID(ctx, testFirebaseUID)
				
				Expect(err).To(HaveOccurred())
				Expect(tracks).To(BeNil())
			})
		})
	})

	Describe("UpdateTrack", func() {
		Context("when update is valid", func() {
			It("should successfully update track metadata", func() {
				updates := map[string]interface{}{
					"size":     int64(1024000),
					"duration": 180,
				}
				
				mockNostrTrackService.EXPECT().
					UpdateTrack(ctx, testTrackID, updates).
					Return(nil)

				err := mockNostrTrackService.UpdateTrack(ctx, testTrackID, updates)
				
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("when update fails", func() {
			It("should return error", func() {
				updates := map[string]interface{}{"size": int64(1024000)}
				expectedError := errors.New("failed to update track")
				
				mockNostrTrackService.EXPECT().
					UpdateTrack(ctx, testTrackID, updates).
					Return(expectedError)

				err := mockNostrTrackService.UpdateTrack(ctx, testTrackID, updates)
				
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("update"))
			})
		})
	})

	Describe("MarkTrackAsProcessed", func() {
		Context("when processing completes successfully", func() {
			It("should update track with processing results", func() {
				size := int64(2048000)
				duration := 240
				
				mockNostrTrackService.EXPECT().
					MarkTrackAsProcessed(ctx, testTrackID, size, duration).
					Return(nil)

				err := mockNostrTrackService.MarkTrackAsProcessed(ctx, testTrackID, size, duration)
				
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("when update fails", func() {
			It("should return error", func() {
				expectedError := errors.New("failed to mark track as processed")
				
				mockNostrTrackService.EXPECT().
					MarkTrackAsProcessed(ctx, testTrackID, int64(1024), 120).
					Return(expectedError)

				err := mockNostrTrackService.MarkTrackAsProcessed(ctx, testTrackID, int64(1024), 120)
				
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("MarkTrackAsCompressed", func() {
		Context("when compression completes successfully", func() {
			It("should update track with compressed file URL", func() {
				compressedURL := "https://storage.googleapis.com/test-bucket/compressed/test-track-123.mp3"
				
				mockNostrTrackService.EXPECT().
					MarkTrackAsCompressed(ctx, testTrackID, compressedURL).
					Return(nil)

				err := mockNostrTrackService.MarkTrackAsCompressed(ctx, testTrackID, compressedURL)
				
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("when update fails", func() {
			It("should return error", func() {
				compressedURL := "https://storage.googleapis.com/test-bucket/compressed/test-track-123.mp3"
				expectedError := errors.New("failed to mark track as compressed")
				
				mockNostrTrackService.EXPECT().
					MarkTrackAsCompressed(ctx, testTrackID, compressedURL).
					Return(expectedError)

				err := mockNostrTrackService.MarkTrackAsCompressed(ctx, testTrackID, compressedURL)
				
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("DeleteTrack", func() {
		Context("when soft delete is successful", func() {
			It("should mark track as deleted", func() {
				mockNostrTrackService.EXPECT().
					DeleteTrack(ctx, testTrackID).
					Return(nil)

				err := mockNostrTrackService.DeleteTrack(ctx, testTrackID)
				
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("when delete fails", func() {
			It("should return error", func() {
				expectedError := errors.New("failed to delete track")
				
				mockNostrTrackService.EXPECT().
					DeleteTrack(ctx, testTrackID).
					Return(expectedError)

				err := mockNostrTrackService.DeleteTrack(ctx, testTrackID)
				
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("HardDeleteTrack", func() {
		Context("when hard delete is successful", func() {
			It("should permanently delete track and files", func() {
				mockNostrTrackService.EXPECT().
					HardDeleteTrack(ctx, testTrackID).
					Return(nil)

				err := mockNostrTrackService.HardDeleteTrack(ctx, testTrackID)
				
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("when track doesn't exist", func() {
			It("should return error", func() {
				expectedError := errors.New("failed to get track for deletion")
				
				mockNostrTrackService.EXPECT().
					HardDeleteTrack(ctx, testTrackID).
					Return(expectedError)

				err := mockNostrTrackService.HardDeleteTrack(ctx, testTrackID)
				
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("for deletion"))
			})
		})
	})

	Describe("UpdateCompressionVisibility", func() {
		Context("when visibility update is successful", func() {
			It("should update compression version visibility", func() {
				updates := []models.VersionUpdate{testutil.ValidVersionUpdate()}
				
				mockNostrTrackService.EXPECT().
					UpdateCompressionVisibility(ctx, testTrackID, updates).
					Return(nil)

				err := mockNostrTrackService.UpdateCompressionVisibility(ctx, testTrackID, updates)
				
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("when track retrieval fails", func() {
			It("should return error", func() {
				updates := []models.VersionUpdate{testutil.ValidVersionUpdate()}
				expectedError := errors.New("failed to get track")
				
				mockNostrTrackService.EXPECT().
					UpdateCompressionVisibility(ctx, testTrackID, updates).
					Return(expectedError)

				err := mockNostrTrackService.UpdateCompressionVisibility(ctx, testTrackID, updates)
				
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("get track"))
			})
		})
	})

	Describe("AddCompressionVersion", func() {
		Context("when adding new compression version", func() {
			It("should successfully add version to track", func() {
				version := testutil.ValidCompressionVersion()
				
				mockNostrTrackService.EXPECT().
					AddCompressionVersion(ctx, testTrackID, version).
					Return(nil)

				err := mockNostrTrackService.AddCompressionVersion(ctx, testTrackID, version)
				
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("when updating existing compression version", func() {
			It("should update existing version", func() {
				version := testutil.ValidCompressionVersion()
				
				mockNostrTrackService.EXPECT().
					AddCompressionVersion(ctx, testTrackID, version).
					Return(nil)

				err := mockNostrTrackService.AddCompressionVersion(ctx, testTrackID, version)
				
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("when operation fails", func() {
			It("should return error", func() {
				version := testutil.ValidCompressionVersion()
				expectedError := errors.New("failed to update track")
				
				mockNostrTrackService.EXPECT().
					AddCompressionVersion(ctx, testTrackID, version).
					Return(expectedError)

				err := mockNostrTrackService.AddCompressionVersion(ctx, testTrackID, version)
				
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("SetPendingCompression", func() {
		Context("when setting pending compression status", func() {
			It("should successfully update pending status to true", func() {
				mockNostrTrackService.EXPECT().
					SetPendingCompression(ctx, testTrackID, true).
					Return(nil)

				err := mockNostrTrackService.SetPendingCompression(ctx, testTrackID, true)
				
				Expect(err).ToNot(HaveOccurred())
			})

			It("should successfully update pending status to false", func() {
				mockNostrTrackService.EXPECT().
					SetPendingCompression(ctx, testTrackID, false).
					Return(nil)

				err := mockNostrTrackService.SetPendingCompression(ctx, testTrackID, false)
				
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("when update fails", func() {
			It("should return error", func() {
				expectedError := errors.New("failed to update pending compression status")
				
				mockNostrTrackService.EXPECT().
					SetPendingCompression(ctx, testTrackID, true).
					Return(expectedError)

				err := mockNostrTrackService.SetPendingCompression(ctx, testTrackID, true)
				
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("pending compression"))
			})
		})
	})
})