package services_test

import (
	"context"
	"errors"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/golang/mock/gomock"

	"github.com/wavlake/monorepo/internal/models"
	"github.com/wavlake/monorepo/internal/services"
	"github.com/wavlake/monorepo/tests/mocks"
	"github.com/wavlake/monorepo/tests/testutil"
)

var _ = Describe("CompressionService", func() {
	var (
		ctrl               *gomock.Controller
		mockNostrTrack     *mocks.MockNostrTrackServiceInterface
		compressionService services.CompressionServiceInterface
		ctx                context.Context
		testTrackID        string
		testCompressionOpts models.CompressionOption
		testCompressionVersion models.CompressionVersion
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockNostrTrack = mocks.NewMockNostrTrackServiceInterface(ctrl)
		compressionService = services.NewCompressionService(mockNostrTrack)
		ctx = context.Background()
		testTrackID = testutil.TestTrackID
		
		testCompressionOpts = models.CompressionOption{
			Bitrate:    256,
			Format:     "mp3",
			Quality:    "high",
			SampleRate: 44100,
		}
		
		testCompressionVersion = models.CompressionVersion{
			ID:         "version-123",
			URL:        "gs://bucket/compressed/track-123-256.mp3",
			Bitrate:    256,
			Format:     "mp3",
			Quality:    "high",
			SampleRate: 44100,
			Size:       5242880, // 5MB
			IsPublic:   true,
			CreatedAt:  time.Now(),
			Options:    testCompressionOpts,
		}
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Describe("RequestCompression", func() {
		Context("when compression request is valid", func() {
			It("should queue compression job successfully", func() {
				mockNostrTrack.EXPECT().
					GetTrack(ctx, testTrackID).
					Return(&models.NostrTrack{
						ID:          testTrackID,
						OriginalURL: "gs://bucket/original/track-123.wav",
						Extension:   ".wav",
					}, nil)

				mockNostrTrack.EXPECT().
					SetPendingCompression(ctx, testTrackID, true).
					Return(nil)

				err := compressionService.RequestCompression(ctx, testTrackID, []models.CompressionOption{testCompressionOpts})
				
				Expect(err).ToNot(HaveOccurred())
			})

			It("should return error when track not found", func() {
				mockNostrTrack.EXPECT().
					GetTrack(ctx, testTrackID).
					Return(nil, errors.New("track not found"))

				err := compressionService.RequestCompression(ctx, testTrackID, []models.CompressionOption{testCompressionOpts})
				
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("track not found"))
			})
		})
	})

	Describe("GetCompressionStatus", func() {
		It("should return compression status for track", func() {
			track := &models.NostrTrack{
				ID:                    testTrackID,
				IsProcessing:          false,
				HasPendingCompression: false,
				CreatedAt:            time.Now(),
			}

			mockNostrTrack.EXPECT().
				GetTrack(ctx, testTrackID).
				Return(track, nil)

			status, err := compressionService.GetCompressionStatus(ctx, testTrackID)
			
			Expect(err).ToNot(HaveOccurred())
			Expect(status.TrackID).To(Equal(testTrackID))
			Expect(status.Status).To(Equal("completed"))
			Expect(status.Progress).To(Equal(100))
		})

		It("should return processing status when track is processing", func() {
			track := &models.NostrTrack{
				ID:           testTrackID,
				IsProcessing: true,
				CreatedAt:    time.Now(),
			}

			mockNostrTrack.EXPECT().
				GetTrack(ctx, testTrackID).
				Return(track, nil)

			status, err := compressionService.GetCompressionStatus(ctx, testTrackID)
			
			Expect(err).ToNot(HaveOccurred())
			Expect(status.Status).To(Equal("processing"))
		})

		It("should return queued status when has pending compression", func() {
			track := &models.NostrTrack{
				ID:                    testTrackID,
				IsProcessing:          false,
				HasPendingCompression: true,
				CreatedAt:            time.Now(),
			}

			mockNostrTrack.EXPECT().
				GetTrack(ctx, testTrackID).
				Return(track, nil)

			status, err := compressionService.GetCompressionStatus(ctx, testTrackID)
			
			Expect(err).ToNot(HaveOccurred())
			Expect(status.Status).To(Equal("queued"))
		})

		It("should return error for invalid track ID", func() {
			mockNostrTrack.EXPECT().
				GetTrack(ctx, testTrackID).
				Return(nil, errors.New("track not found"))

			_, err := compressionService.GetCompressionStatus(ctx, testTrackID)
			
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("track not found"))
		})
	})

	Describe("AddCompressionVersion", func() {
		It("should add compression version to track", func() {
			mockNostrTrack.EXPECT().
				AddCompressionVersion(ctx, testTrackID, testCompressionVersion).
				Return(nil)

			err := compressionService.AddCompressionVersion(ctx, testTrackID, testCompressionVersion)
			
			Expect(err).ToNot(HaveOccurred())
		})

		It("should handle service errors", func() {
			mockNostrTrack.EXPECT().
				AddCompressionVersion(ctx, testTrackID, testCompressionVersion).
				Return(errors.New("service error"))

			err := compressionService.AddCompressionVersion(ctx, testTrackID, testCompressionVersion)
			
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("service error"))
		})
	})

	Describe("UpdateVersionVisibility", func() {
		It("should update compression version visibility", func() {
			versionID := "version-123"
			isPublic := false

			expectedUpdates := []models.VersionUpdate{
				{
					VersionID: versionID,
					IsPublic:  isPublic,
				},
			}

			mockNostrTrack.EXPECT().
				UpdateCompressionVisibility(ctx, testTrackID, expectedUpdates).
				Return(nil)

			err := compressionService.UpdateVersionVisibility(ctx, testTrackID, versionID, isPublic)
			
			Expect(err).ToNot(HaveOccurred())
		})

		It("should return error for service failures", func() {
			versionID := "version-123"
			isPublic := true

			expectedUpdates := []models.VersionUpdate{
				{
					VersionID: versionID,
					IsPublic:  isPublic,
				},
			}

			mockNostrTrack.EXPECT().
				UpdateCompressionVisibility(ctx, testTrackID, expectedUpdates).
				Return(errors.New("service error"))

			err := compressionService.UpdateVersionVisibility(ctx, testTrackID, versionID, isPublic)
			
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("service error"))
		})
	})

	Describe("GetPublicVersions", func() {
		It("should return only public compression versions", func() {
			publicVersion1 := testCompressionVersion
			publicVersion1.ID = "public-1"
			publicVersion1.IsPublic = true

			publicVersion2 := testCompressionVersion
			publicVersion2.ID = "public-2" 
			publicVersion2.IsPublic = true
			publicVersion2.Format = "aac"

			privateVersion := testCompressionVersion
			privateVersion.ID = "private-1"
			privateVersion.IsPublic = false

			allVersions := []models.CompressionVersion{publicVersion1, privateVersion, publicVersion2}

			mockNostrTrack.EXPECT().
				GetTrack(ctx, testTrackID).
				Return(&models.NostrTrack{
					ID:                  testTrackID,
					CompressionVersions: allVersions,
				}, nil)

			publicVersions, err := compressionService.GetPublicVersions(ctx, testTrackID)
			
			Expect(err).ToNot(HaveOccurred())
			Expect(publicVersions).To(HaveLen(2))
			Expect(publicVersions[0].ID).To(Equal("public-1"))
			Expect(publicVersions[1].ID).To(Equal("public-2"))
			
			for _, version := range publicVersions {
				Expect(version.IsPublic).To(BeTrue())
			}
		})

		It("should return empty array when no public versions exist", func() {
			privateVersion := testCompressionVersion
			privateVersion.IsPublic = false

			mockNostrTrack.EXPECT().
				GetTrack(ctx, testTrackID).
				Return(&models.NostrTrack{
					ID:                  testTrackID,
					CompressionVersions: []models.CompressionVersion{privateVersion},
				}, nil)

			publicVersions, err := compressionService.GetPublicVersions(ctx, testTrackID)
			
			Expect(err).ToNot(HaveOccurred())
			Expect(publicVersions).To(BeEmpty())
		})

		It("should handle track not found", func() {
			mockNostrTrack.EXPECT().
				GetTrack(ctx, testTrackID).
				Return(nil, errors.New("track not found"))

			_, err := compressionService.GetPublicVersions(ctx, testTrackID)
			
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("track not found"))
		})
	})

	Describe("DeleteCompressionVersion", func() {
		It("should delete compression version", func() {
			versionID := "version-123"

			mockNostrTrack.EXPECT().
				GetTrack(ctx, testTrackID).
				Return(&models.NostrTrack{
					ID: testTrackID,
				}, nil)

			err := compressionService.DeleteCompressionVersion(ctx, testTrackID, versionID)
			
			Expect(err).ToNot(HaveOccurred())
		})

		It("should return error when track not found", func() {
			versionID := "version-123"

			mockNostrTrack.EXPECT().
				GetTrack(ctx, testTrackID).
				Return(nil, errors.New("track not found"))

			err := compressionService.DeleteCompressionVersion(ctx, testTrackID, versionID)
			
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("track not found"))
		})
	})
})