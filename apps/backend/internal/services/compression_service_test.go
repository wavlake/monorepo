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
		ctrl                 *gomock.Controller
		mockStorage          *mocks.MockStorageServiceInterface
		mockNostrTrack       *mocks.MockNostrTrackServiceInterface
		compressionService   services.CompressionServiceInterface
		ctx                  context.Context
		testTrackID          string
		testCompressionOpts  models.CompressionOption
		testCompressionVersion models.CompressionVersion
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockStorage = mocks.NewMockStorageServiceInterface(ctrl)
		mockNostrTrack = mocks.NewMockNostrTrackServiceInterface(ctrl)
		compressionService = services.NewCompressionService(mockStorage, mockNostrTrack)
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
			It("should queue compression job and return job ID", func() {
				expectedJobID := "job-456"
				
				mockNostrTrack.EXPECT().
					GetTrack(ctx, testTrackID).
					Return(&models.NostrTrack{
						ID:          testTrackID,
						OriginalURL: "gs://bucket/original/track-123.wav",
						Extension:   ".wav",
					}, nil)

				mockNostrTrack.EXPECT().
					UpdateCompressionStatus(ctx, testTrackID, true).
					Return(nil)

				jobID, err := compressionService.RequestCompression(ctx, testTrackID, testCompressionOpts)
				
				Expect(err).ToNot(HaveOccurred())
				Expect(jobID).To(Equal(expectedJobID))
			})

			It("should validate compression options before queuing", func() {
				invalidOpts := models.CompressionOption{
					Bitrate: 999, // Invalid bitrate
					Format:  "invalid",
					Quality: "unknown",
				}

				_, err := compressionService.RequestCompression(ctx, testTrackID, invalidOpts)
				
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("invalid compression options"))
			})

			It("should return error when track not found", func() {
				mockNostrTrack.EXPECT().
					GetTrack(ctx, testTrackID).
					Return(nil, errors.New("track not found"))

				_, err := compressionService.RequestCompression(ctx, testTrackID, testCompressionOpts)
				
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("track not found"))
			})

			It("should handle compression already in progress", func() {
				mockNostrTrack.EXPECT().
					GetTrack(ctx, testTrackID).
					Return(&models.NostrTrack{
						ID:                    testTrackID,
						OriginalURL:          "gs://bucket/original/track-123.wav",
						HasPendingCompression: true,
					}, nil)

				_, err := compressionService.RequestCompression(ctx, testTrackID, testCompressionOpts)
				
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("compression already in progress"))
			})
		})

		Context("when compression options validation", func() {
			DescribeTable("should validate compression parameters",
				func(opts models.CompressionOption, expectedValid bool) {
					if expectedValid {
						mockNostrTrack.EXPECT().
							GetTrack(ctx, testTrackID).
							Return(&models.NostrTrack{ID: testTrackID, OriginalURL: "gs://bucket/test.wav"}, nil)
						mockNostrTrack.EXPECT().
							UpdateCompressionStatus(ctx, testTrackID, true).
							Return(nil)
					}

					_, err := compressionService.RequestCompression(ctx, testTrackID, opts)
					
					if expectedValid {
						Expect(err).ToNot(HaveOccurred())
					} else {
						Expect(err).To(HaveOccurred())
					}
				},
				Entry("valid MP3 high quality", models.CompressionOption{Bitrate: 320, Format: "mp3", Quality: "high", SampleRate: 44100}, true),
				Entry("valid AAC medium quality", models.CompressionOption{Bitrate: 256, Format: "aac", Quality: "medium", SampleRate: 48000}, true),
				Entry("valid OGG low quality", models.CompressionOption{Bitrate: 128, Format: "ogg", Quality: "low", SampleRate: 44100}, true),
				Entry("invalid bitrate too low", models.CompressionOption{Bitrate: 64, Format: "mp3", Quality: "medium"}, false),
				Entry("invalid bitrate too high", models.CompressionOption{Bitrate: 500, Format: "mp3", Quality: "high"}, false),
				Entry("invalid format", models.CompressionOption{Bitrate: 256, Format: "wav", Quality: "high"}, false),
				Entry("invalid quality", models.CompressionOption{Bitrate: 256, Format: "mp3", Quality: "ultra"}, false),
				Entry("invalid sample rate", models.CompressionOption{Bitrate: 256, Format: "mp3", Quality: "high", SampleRate: 22000}, false),
			)
		})
	})

	Describe("GetCompressionStatus", func() {
		It("should return compression job status", func() {
			jobID := "job-456"
			expectedStatus := &services.CompressionStatus{
				JobID:     jobID,
				Status:    "processing",
				Progress:  45,
				ETA:       300, // 5 minutes
				CreatedAt: time.Now(),
			}

			status, err := compressionService.GetCompressionStatus(ctx, jobID)
			
			Expect(err).ToNot(HaveOccurred())
			Expect(status).To(Equal(expectedStatus))
		})

		It("should return error for invalid job ID", func() {
			invalidJobID := "invalid-job"

			_, err := compressionService.GetCompressionStatus(ctx, invalidJobID)
			
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("job not found"))
		})

		It("should handle completed compression job", func() {
			jobID := "job-completed"
			expectedStatus := &services.CompressionStatus{
				JobID:     jobID,
				Status:    "completed",
				Progress:  100,
				CreatedAt: time.Now(),
				CompletedAt: &time.Time{},
				Result: &models.CompressionVersion{
					ID:     "version-123",
					URL:    "gs://bucket/compressed/track-123-256.mp3",
					Format: "mp3",
				},
			}

			status, err := compressionService.GetCompressionStatus(ctx, jobID)
			
			Expect(err).ToNot(HaveOccurred())
			Expect(status.Status).To(Equal("completed"))
			Expect(status.Progress).To(Equal(100))
			Expect(status.Result).ToNot(BeNil())
		})

		It("should handle failed compression job", func() {
			jobID := "job-failed"

			status, err := compressionService.GetCompressionStatus(ctx, jobID)
			
			Expect(err).ToNot(HaveOccurred())
			Expect(status.Status).To(Equal("failed"))
			Expect(status.Error).ToNot(BeEmpty())
		})
	})

	Describe("AddCompressionVersion", func() {
		It("should add compression version to track", func() {
			mockNostrTrack.EXPECT().
				GetTrack(ctx, testTrackID).
				Return(&models.NostrTrack{
					ID:                  testTrackID,
					CompressionVersions: []models.CompressionVersion{},
				}, nil)

			mockNostrTrack.EXPECT().
				AddCompressionVersion(ctx, testTrackID, testCompressionVersion).
				Return(nil)

			err := compressionService.AddCompressionVersion(ctx, testTrackID, testCompressionVersion)
			
			Expect(err).ToNot(HaveOccurred())
		})

		It("should prevent duplicate compression versions", func() {
			existingVersion := testCompressionVersion
			existingVersion.ID = "existing-version"

			mockNostrTrack.EXPECT().
				GetTrack(ctx, testTrackID).
				Return(&models.NostrTrack{
					ID:                  testTrackID,
					CompressionVersions: []models.CompressionVersion{existingVersion},
				}, nil)

			// Try to add version with same format/bitrate/quality
			err := compressionService.AddCompressionVersion(ctx, testTrackID, testCompressionVersion)
			
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("compression version already exists"))
		})

		It("should handle storage service errors", func() {
			mockNostrTrack.EXPECT().
				GetTrack(ctx, testTrackID).
				Return(nil, errors.New("storage error"))

			err := compressionService.AddCompressionVersion(ctx, testTrackID, testCompressionVersion)
			
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("storage error"))
		})
	})

	Describe("UpdateVersionVisibility", func() {
		It("should update compression version visibility", func() {
			versionID := "version-123"
			isPublic := false

			mockNostrTrack.EXPECT().
				UpdateCompressionVersionVisibility(ctx, testTrackID, versionID, isPublic).
				Return(nil)

			err := compressionService.UpdateVersionVisibility(ctx, testTrackID, versionID, isPublic)
			
			Expect(err).ToNot(HaveOccurred())
		})

		It("should return error for non-existent version", func() {
			versionID := "non-existent"
			isPublic := true

			mockNostrTrack.EXPECT().
				UpdateCompressionVersionVisibility(ctx, testTrackID, versionID, isPublic).
				Return(errors.New("version not found"))

			err := compressionService.UpdateVersionVisibility(ctx, testTrackID, versionID, isPublic)
			
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("version not found"))
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
		It("should delete compression version and associated file", func() {
			versionID := "version-123"
			versionURL := "gs://bucket/compressed/track-123-256.mp3"

			mockStorage.EXPECT().
				DeleteFile(ctx, versionURL).
				Return(nil)

			mockNostrTrack.EXPECT().
				RemoveCompressionVersion(ctx, testTrackID, versionID).
				Return(nil)

			err := compressionService.DeleteCompressionVersion(ctx, testTrackID, versionID)
			
			Expect(err).ToNot(HaveOccurred())
		})

		It("should handle file deletion errors gracefully", func() {
			versionID := "version-123"

			mockStorage.EXPECT().
				DeleteFile(ctx, gomock.Any()).
				Return(errors.New("file not found"))

			// Should still try to remove from database
			mockNostrTrack.EXPECT().
				RemoveCompressionVersion(ctx, testTrackID, versionID).
				Return(nil)

			err := compressionService.DeleteCompressionVersion(ctx, testTrackID, versionID)
			
			Expect(err).ToNot(HaveOccurred()) // Should not fail if file already gone
		})

		It("should return error when database removal fails", func() {
			versionID := "version-123"

			mockStorage.EXPECT().
				DeleteFile(ctx, gomock.Any()).
				Return(nil)

			mockNostrTrack.EXPECT().
				RemoveCompressionVersion(ctx, testTrackID, versionID).
				Return(errors.New("database error"))

			err := compressionService.DeleteCompressionVersion(ctx, testTrackID, versionID)
			
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("database error"))
		})
	})
})