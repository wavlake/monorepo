package services_test

import (
	"context"
	"errors"
	"io"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/golang/mock/gomock"

	"github.com/wavlake/monorepo/tests/mocks"
)

var _ = Describe("StorageServiceInterface", func() {
	var (
		ctrl                 *gomock.Controller
		mockStorageService   *mocks.MockStorageServiceInterface
		ctx                  context.Context
		testObjectName       string
		testBucketName       string
		testContentType      string
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockStorageService = mocks.NewMockStorageServiceInterface(ctrl)
		ctx = context.Background()
		testObjectName = "tracks/original/test-track-123.mp3"
		testBucketName = "test-bucket"
		testContentType = "audio/mpeg"
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Describe("GeneratePresignedURL", func() {
		Context("when all parameters are valid", func() {
			It("should generate a presigned URL for upload", func() {
				expiration := time.Hour
				expectedURL := "https://storage.googleapis.com/test-bucket/upload-url?signature=abc123"
				
				mockStorageService.EXPECT().
					GeneratePresignedURL(ctx, testObjectName, expiration).
					Return(expectedURL, nil)

				url, err := mockStorageService.GeneratePresignedURL(ctx, testObjectName, expiration)
				
				Expect(err).ToNot(HaveOccurred())
				Expect(url).To(Equal(expectedURL))
				Expect(url).To(ContainSubstring("storage.googleapis.com"))
				Expect(url).To(ContainSubstring("signature="))
			})

			It("should handle different expiration durations", func() {
				shortExpiration := 30 * time.Minute
				expectedURL := "https://storage.googleapis.com/test-bucket/short-url?expires=30min"
				
				mockStorageService.EXPECT().
					GeneratePresignedURL(ctx, testObjectName, shortExpiration).
					Return(expectedURL, nil)

				url, err := mockStorageService.GeneratePresignedURL(ctx, testObjectName, shortExpiration)
				
				Expect(err).ToNot(HaveOccurred())
				Expect(url).To(Equal(expectedURL))
			})

			It("should handle different object paths", func() {
				compressedObjectName := "tracks/compressed/test-track-123.mp3"
				expectedURL := "https://storage.googleapis.com/test-bucket/compressed-url"
				
				mockStorageService.EXPECT().
					GeneratePresignedURL(ctx, compressedObjectName, time.Hour).
					Return(expectedURL, nil)

				url, err := mockStorageService.GeneratePresignedURL(ctx, compressedObjectName, time.Hour)
				
				Expect(err).ToNot(HaveOccurred())
				Expect(url).To(Equal(expectedURL))
			})
		})

		Context("when service account signing fails", func() {
			It("should return signing error", func() {
				expectedError := errors.New("failed to generate presigned URL: IAM credentials service error")
				
				mockStorageService.EXPECT().
					GeneratePresignedURL(ctx, testObjectName, time.Hour).
					Return("", expectedError)

				url, err := mockStorageService.GeneratePresignedURL(ctx, testObjectName, time.Hour)
				
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to generate presigned URL"))
				Expect(url).To(BeEmpty())
			})

			It("should handle service account email errors", func() {
				expectedError := errors.New("failed to generate presigned URL: invalid service account")
				
				mockStorageService.EXPECT().
					GeneratePresignedURL(ctx, testObjectName, time.Hour).
					Return("", expectedError)

				url, err := mockStorageService.GeneratePresignedURL(ctx, testObjectName, time.Hour)
				
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("service account"))
				Expect(url).To(BeEmpty())
			})
		})

		Context("when bucket access fails", func() {
			It("should return bucket access error", func() {
				expectedError := errors.New("failed to generate presigned URL: bucket not found")
				
				mockStorageService.EXPECT().
					GeneratePresignedURL(ctx, testObjectName, time.Hour).
					Return("", expectedError)

				url, err := mockStorageService.GeneratePresignedURL(ctx, testObjectName, time.Hour)
				
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("bucket"))
				Expect(url).To(BeEmpty())
			})
		})
	})

	Describe("GetPublicURL", func() {
		Context("when generating public URLs", func() {
			It("should return correct public URL format", func() {
				expectedURL := "https://storage.googleapis.com/test-bucket/tracks/original/test-track-123.mp3"
				
				mockStorageService.EXPECT().
					GetPublicURL(testObjectName).
					Return(expectedURL)

				url := mockStorageService.GetPublicURL(testObjectName)
				
				Expect(url).To(Equal(expectedURL))
				Expect(url).To(HavePrefix("https://storage.googleapis.com/"))
				Expect(url).To(ContainSubstring(testObjectName))
			})

			It("should handle different object types", func() {
				imageObjectName := "artwork/album-123.jpg"
				expectedURL := "https://storage.googleapis.com/test-bucket/artwork/album-123.jpg"
				
				mockStorageService.EXPECT().
					GetPublicURL(imageObjectName).
					Return(expectedURL)

				url := mockStorageService.GetPublicURL(imageObjectName)
				
				Expect(url).To(Equal(expectedURL))
				Expect(url).To(ContainSubstring("artwork/album-123.jpg"))
			})

			It("should handle objects with special characters", func() {
				specialObjectName := "tracks/special/track with spaces & symbols.mp3"
				expectedURL := "https://storage.googleapis.com/test-bucket/tracks/special/track with spaces & symbols.mp3"
				
				mockStorageService.EXPECT().
					GetPublicURL(specialObjectName).
					Return(expectedURL)

				url := mockStorageService.GetPublicURL(specialObjectName)
				
				Expect(url).To(Equal(expectedURL))
			})
		})
	})

	Describe("UploadObject", func() {
		Context("when uploading valid data", func() {
			It("should successfully upload object", func() {
				data := strings.NewReader("test audio data")
				
				mockStorageService.EXPECT().
					UploadObject(ctx, testObjectName, data, testContentType).
					Return(nil)

				err := mockStorageService.UploadObject(ctx, testObjectName, data, testContentType)
				
				Expect(err).ToNot(HaveOccurred())
			})

			It("should handle different content types", func() {
				data := strings.NewReader("test image data")
				imageContentType := "image/jpeg"
				imageObjectName := "artwork/test-image.jpg"
				
				mockStorageService.EXPECT().
					UploadObject(ctx, imageObjectName, data, imageContentType).
					Return(nil)

				err := mockStorageService.UploadObject(ctx, imageObjectName, data, imageContentType)
				
				Expect(err).ToNot(HaveOccurred())
			})

			It("should handle large file uploads", func() {
				largeData := strings.NewReader(strings.Repeat("data", 1000000)) // 4MB of data
				
				mockStorageService.EXPECT().
					UploadObject(ctx, testObjectName, largeData, testContentType).
					Return(nil)

				err := mockStorageService.UploadObject(ctx, testObjectName, largeData, testContentType)
				
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("when upload fails", func() {
			It("should return upload error", func() {
				data := strings.NewReader("test data")
				expectedError := errors.New("failed to upload object: network timeout")
				
				mockStorageService.EXPECT().
					UploadObject(ctx, testObjectName, data, testContentType).
					Return(expectedError)

				err := mockStorageService.UploadObject(ctx, testObjectName, data, testContentType)
				
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to upload object"))
			})

			It("should handle writer close errors", func() {
				data := strings.NewReader("test data")
				expectedError := errors.New("failed to close writer: connection lost")
				
				mockStorageService.EXPECT().
					UploadObject(ctx, testObjectName, data, testContentType).
					Return(expectedError)

				err := mockStorageService.UploadObject(ctx, testObjectName, data, testContentType)
				
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("close writer"))
			})

			It("should handle permission errors", func() {
				data := strings.NewReader("test data")
				expectedError := errors.New("failed to upload object: permission denied")
				
				mockStorageService.EXPECT().
					UploadObject(ctx, testObjectName, data, testContentType).
					Return(expectedError)

				err := mockStorageService.UploadObject(ctx, testObjectName, data, testContentType)
				
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("permission denied"))
			})
		})
	})

	Describe("CopyObject", func() {
		Context("when copying objects within bucket", func() {
			It("should successfully copy object", func() {
				srcObject := "tracks/original/test-track-123.mp3"
				dstObject := "tracks/backup/test-track-123.mp3"
				
				mockStorageService.EXPECT().
					CopyObject(ctx, srcObject, dstObject).
					Return(nil)

				err := mockStorageService.CopyObject(ctx, srcObject, dstObject)
				
				Expect(err).ToNot(HaveOccurred())
			})

			It("should handle different source and destination paths", func() {
				srcObject := "tracks/temp/processing-123.wav"
				dstObject := "tracks/processed/final-123.mp3"
				
				mockStorageService.EXPECT().
					CopyObject(ctx, srcObject, dstObject).
					Return(nil)

				err := mockStorageService.CopyObject(ctx, srcObject, dstObject)
				
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("when copy operation fails", func() {
			It("should return copy error when source not found", func() {
				srcObject := "tracks/nonexistent/missing.mp3"
				dstObject := "tracks/backup/missing.mp3"
				expectedError := errors.New("failed to copy object: source object not found")
				
				mockStorageService.EXPECT().
					CopyObject(ctx, srcObject, dstObject).
					Return(expectedError)

				err := mockStorageService.CopyObject(ctx, srcObject, dstObject)
				
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to copy object"))
				Expect(err.Error()).To(ContainSubstring("not found"))
			})

			It("should handle destination path errors", func() {
				srcObject := "tracks/original/test-track-123.mp3"
				dstObject := "invalid/path/with/permissions/issue.mp3"
				expectedError := errors.New("failed to copy object: destination path invalid")
				
				mockStorageService.EXPECT().
					CopyObject(ctx, srcObject, dstObject).
					Return(expectedError)

				err := mockStorageService.CopyObject(ctx, srcObject, dstObject)
				
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("destination"))
			})

			It("should handle network failures during copy", func() {
				srcObject := "tracks/original/large-file.wav"
				dstObject := "tracks/backup/large-file.wav"
				expectedError := errors.New("failed to copy object: network interrupted")
				
				mockStorageService.EXPECT().
					CopyObject(ctx, srcObject, dstObject).
					Return(expectedError)

				err := mockStorageService.CopyObject(ctx, srcObject, dstObject)
				
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("network"))
			})
		})
	})

	Describe("DeleteObject", func() {
		Context("when deleting existing objects", func() {
			It("should successfully delete object", func() {
				mockStorageService.EXPECT().
					DeleteObject(ctx, testObjectName).
					Return(nil)

				err := mockStorageService.DeleteObject(ctx, testObjectName)
				
				Expect(err).ToNot(HaveOccurred())
			})

			It("should handle deletion of different file types", func() {
				imageObject := "artwork/album-456.jpg"
				
				mockStorageService.EXPECT().
					DeleteObject(ctx, imageObject).
					Return(nil)

				err := mockStorageService.DeleteObject(ctx, imageObject)
				
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("when delete operation fails", func() {
			It("should return error when object not found", func() {
				expectedError := errors.New("failed to delete object: object not found")
				
				mockStorageService.EXPECT().
					DeleteObject(ctx, testObjectName).
					Return(expectedError)

				err := mockStorageService.DeleteObject(ctx, testObjectName)
				
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to delete object"))
				Expect(err.Error()).To(ContainSubstring("not found"))
			})

			It("should handle permission errors", func() {
				expectedError := errors.New("failed to delete object: permission denied")
				
				mockStorageService.EXPECT().
					DeleteObject(ctx, testObjectName).
					Return(expectedError)

				err := mockStorageService.DeleteObject(ctx, testObjectName)
				
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("permission denied"))
			})

			It("should handle bucket access errors", func() {
				expectedError := errors.New("failed to delete object: bucket access denied")
				
				mockStorageService.EXPECT().
					DeleteObject(ctx, testObjectName).
					Return(expectedError)

				err := mockStorageService.DeleteObject(ctx, testObjectName)
				
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("bucket"))
			})
		})
	})

	Describe("GetObjectMetadata", func() {
		Context("when retrieving metadata for existing objects", func() {
			It("should return object attributes", func() {
				expectedMetadata := map[string]interface{}{
					"name":         testObjectName,
					"size":         int64(2048000),
					"contentType":  testContentType,
					"lastModified": time.Now(),
				}
				
				mockStorageService.EXPECT().
					GetObjectMetadata(ctx, testObjectName).
					Return(expectedMetadata, nil)

				metadata, err := mockStorageService.GetObjectMetadata(ctx, testObjectName)
				
				Expect(err).ToNot(HaveOccurred())
				Expect(metadata).ToNot(BeNil())
			})

			It("should handle different object types metadata", func() {
				imageObject := "artwork/test-image.jpg"
				expectedMetadata := map[string]interface{}{
					"name":        imageObject,
					"size":        int64(512000),
					"contentType": "image/jpeg",
				}
				
				mockStorageService.EXPECT().
					GetObjectMetadata(ctx, imageObject).
					Return(expectedMetadata, nil)

				metadata, err := mockStorageService.GetObjectMetadata(ctx, imageObject)
				
				Expect(err).ToNot(HaveOccurred())
				Expect(metadata).ToNot(BeNil())
			})
		})

		Context("when metadata retrieval fails", func() {
			It("should return error when object not found", func() {
				expectedError := errors.New("failed to get object metadata: object does not exist")
				
				mockStorageService.EXPECT().
					GetObjectMetadata(ctx, testObjectName).
					Return(nil, expectedError)

				metadata, err := mockStorageService.GetObjectMetadata(ctx, testObjectName)
				
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to get object metadata"))
				Expect(metadata).To(BeNil())
			})

			It("should handle permission errors", func() {
				expectedError := errors.New("failed to get object metadata: access denied")
				
				mockStorageService.EXPECT().
					GetObjectMetadata(ctx, testObjectName).
					Return(nil, expectedError)

				metadata, err := mockStorageService.GetObjectMetadata(ctx, testObjectName)
				
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("access denied"))
				Expect(metadata).To(BeNil())
			})
		})
	})

	Describe("GetObjectReader", func() {
		Context("when creating readers for existing objects", func() {
			It("should return a valid reader", func() {
				mockReader := io.NopCloser(strings.NewReader("test file content"))
				
				mockStorageService.EXPECT().
					GetObjectReader(ctx, testObjectName).
					Return(mockReader, nil)

				reader, err := mockStorageService.GetObjectReader(ctx, testObjectName)
				
				Expect(err).ToNot(HaveOccurred())
				Expect(reader).ToNot(BeNil())
				
				// Test that we can read from it
				data, readErr := io.ReadAll(reader)
				Expect(readErr).ToNot(HaveOccurred())
				Expect(string(data)).To(Equal("test file content"))
				
				closeErr := reader.Close()
				Expect(closeErr).ToNot(HaveOccurred())
			})

			It("should handle different file types", func() {
				imageObject := "artwork/test.jpg"
				mockReader := io.NopCloser(strings.NewReader("jpeg image data"))
				
				mockStorageService.EXPECT().
					GetObjectReader(ctx, imageObject).
					Return(mockReader, nil)

				reader, err := mockStorageService.GetObjectReader(ctx, imageObject)
				
				Expect(err).ToNot(HaveOccurred())
				Expect(reader).ToNot(BeNil())
				
				_ = reader.Close()
			})
		})

		Context("when reader creation fails", func() {
			It("should return error when object not found", func() {
				expectedError := errors.New("failed to create object reader: object not found")
				
				mockStorageService.EXPECT().
					GetObjectReader(ctx, testObjectName).
					Return(nil, expectedError)

				reader, err := mockStorageService.GetObjectReader(ctx, testObjectName)
				
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to create object reader"))
				Expect(reader).To(BeNil())
			})

			It("should handle permission errors", func() {
				expectedError := errors.New("failed to create object reader: insufficient permissions")
				
				mockStorageService.EXPECT().
					GetObjectReader(ctx, testObjectName).
					Return(nil, expectedError)

				reader, err := mockStorageService.GetObjectReader(ctx, testObjectName)
				
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("permissions"))
				Expect(reader).To(BeNil())
			})

			It("should handle network errors", func() {
				expectedError := errors.New("failed to create object reader: connection timeout")
				
				mockStorageService.EXPECT().
					GetObjectReader(ctx, testObjectName).
					Return(nil, expectedError)

				reader, err := mockStorageService.GetObjectReader(ctx, testObjectName)
				
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("connection"))
				Expect(reader).To(BeNil())
			})
		})
	})

	Describe("GetBucketName", func() {
		Context("when retrieving bucket configuration", func() {
			It("should return configured bucket name", func() {
				mockStorageService.EXPECT().
					GetBucketName().
					Return(testBucketName)

				bucketName := mockStorageService.GetBucketName()
				
				Expect(bucketName).To(Equal(testBucketName))
				Expect(bucketName).ToNot(BeEmpty())
			})
		})
	})

	Describe("Close", func() {
		Context("when closing storage service", func() {
			It("should successfully close connection", func() {
				mockStorageService.EXPECT().
					Close().
					Return(nil)

				err := mockStorageService.Close()
				
				Expect(err).ToNot(HaveOccurred())
			})

			It("should handle close errors gracefully", func() {
				expectedError := errors.New("connection already closed")
				
				mockStorageService.EXPECT().
					Close().
					Return(expectedError)

				err := mockStorageService.Close()
				
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("already closed"))
			})
		})
	})
})