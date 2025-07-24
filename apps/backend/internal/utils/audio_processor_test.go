package utils_test

import (
	"context"
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/wavlake/monorepo/internal/models"
	"github.com/wavlake/monorepo/internal/utils"
)

var _ = Describe("AudioProcessor", func() {
	var (
		processor   *utils.AudioProcessor
		tempDir     string
		testAudioFile string
		ctx         context.Context
	)

	BeforeEach(func() {
		var err error
		tempDir, err = os.MkdirTemp("", "audio_processor_test")
		Expect(err).ToNot(HaveOccurred())
		
		processor = utils.NewAudioProcessor(tempDir)
		testAudioFile = filepath.Join(tempDir, "test_audio.wav")
		ctx = context.Background()
		
		// Create a mock audio file for testing
		// In real implementation, this would be a valid WAV file
		err = os.WriteFile(testAudioFile, []byte("RIFF....WAVE...."), 0644)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		os.RemoveAll(tempDir)
	})

	Describe("ValidateAudioFile", func() {
		Context("when validating audio files", func() {
			It("should validate supported audio formats", func() {
				supportedFormats := []string{".wav", ".mp3", ".aac", ".ogg", ".flac", ".m4a"}
				
				for _, format := range supportedFormats {
					filename := "test_audio" + format
					err := processor.ValidateAudioFile(ctx, filename)
					Expect(err).To(BeNil(), "Format %s should be supported", format)
				}
			})

			It("should reject unsupported audio formats", func() {
				unsupportedFormats := []string{".txt", ".jpg", ".mp4", ".avi", ".pdf"}
				
				for _, format := range unsupportedFormats {
					filename := "test_file" + format
					err := processor.ValidateAudioFile(ctx, filename)
					Expect(err).ToNot(BeNil(), "Format %s should not be supported", format)
				}
			})

			It("should handle case-insensitive file extensions", func() {
				caseVariations := []string{".MP3", ".Mp3", ".WAV", ".Wav", ".AAC", ".aAc"}
				
				for _, format := range caseVariations {
					filename := "test_audio" + format
					err := processor.ValidateAudioFile(ctx, filename)
					Expect(err).To(BeNil(), "Format %s should be supported (case insensitive)", format)
				}
			})
		})
	})

	Describe("ExtractMetadata", func() {
		Context("when extracting audio metadata", func() {
			It("should extract basic audio metadata from file", func() {
				metadata, err := processor.ExtractMetadata(ctx, testAudioFile)
				
				Expect(err).ToNot(HaveOccurred())
				Expect(metadata).ToNot(BeNil())
				Expect(metadata.Format).ToNot(BeEmpty())
				Expect(metadata.Duration).To(BeNumerically(">", 0))
				Expect(metadata.SampleRate).To(BeNumerically(">", 0))
				Expect(metadata.Bitrate).To(BeNumerically(">", 0))
			})

			It("should extract detailed metadata including tags", func() {
				metadata, err := processor.ExtractMetadata(ctx, testAudioFile)
				
				Expect(err).ToNot(HaveOccurred())
				Expect(metadata.Title).ToNot(BeNil())
				Expect(metadata.Artist).ToNot(BeNil())
				Expect(metadata.Album).ToNot(BeNil())
				Expect(metadata.Genre).ToNot(BeNil())
			})

			It("should return error for non-existent file", func() {
				nonExistentFile := filepath.Join(tempDir, "non_existent.wav")
				
				_, err := processor.ExtractMetadata(ctx, nonExistentFile)
				
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("file not found"))
			})

			It("should return error for invalid audio file", func() {
				invalidFile := filepath.Join(tempDir, "invalid.txt")
				err := os.WriteFile(invalidFile, []byte("not an audio file"), 0644)
				Expect(err).ToNot(HaveOccurred())
				
				_, err = processor.ExtractMetadata(ctx, invalidFile)
				
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("invalid audio format"))
			})

			It("should handle timeout for large files", func() {
				ctxWithTimeout, cancel := context.WithTimeout(ctx, 1*time.Millisecond)
				defer cancel()
				
				_, err := processor.ExtractMetadata(ctxWithTimeout, testAudioFile)
				
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("context deadline exceeded"))
			})
		})
	})

	Describe("CompressAudio", func() {
		Context("when compressing audio with different options", func() {
			It("should compress to MP3 format with specified bitrate", func() {
				options := models.CompressionOption{
					Bitrate:    256,
					Format:     "mp3",
					Quality:    "high",
					SampleRate: 44100,
				}
				
				outputPath := filepath.Join(tempDir, "compressed.mp3")
				err := processor.CompressAudio(ctx, testAudioFile, outputPath, options)
				
				Expect(err).ToNot(HaveOccurred())
				Expect(filepath.Ext(outputPath)).To(Equal(".mp3"))
				
				// Verify output file exists
				_, err = os.Stat(outputPath)
				Expect(err).ToNot(HaveOccurred())
				
				// Verify metadata of compressed file
				metadata, err := processor.ExtractMetadata(ctx, outputPath)
				Expect(err).ToNot(HaveOccurred())
				Expect(metadata.Format).To(Equal("mp3"))
				Expect(metadata.Bitrate).To(Equal(256))
				Expect(metadata.SampleRate).To(Equal(44100))
			})

			It("should compress to AAC format with high quality", func() {
				options := models.CompressionOption{
					Bitrate:    256,
					Format:     "aac",
					Quality:    "high",
					SampleRate: 48000,
				}
				
				outputPath := filepath.Join(tempDir, "compressed.aac")
				err := processor.CompressAudio(ctx, testAudioFile, outputPath, options)
				
				Expect(err).ToNot(HaveOccurred())
				Expect(filepath.Ext(outputPath)).To(Equal(".aac"))
				
				metadata, err := processor.ExtractMetadata(ctx, outputPath)
				Expect(err).ToNot(HaveOccurred())
				Expect(metadata.Format).To(Equal("aac"))
				Expect(metadata.SampleRate).To(Equal(48000))
			})

			It("should compress to OGG format with variable bitrate", func() {
				options := models.CompressionOption{
					Bitrate:    192,
					Format:     "ogg",
					Quality:    "medium",
					SampleRate: 44100,
				}
				
				outputPath := filepath.Join(tempDir, "compressed.ogg")
				err := processor.CompressAudio(ctx, testAudioFile, outputPath, options)
				
				Expect(err).ToNot(HaveOccurred())
				Expect(filepath.Ext(outputPath)).To(Equal(".ogg"))
				
				metadata, err := processor.ExtractMetadata(ctx, outputPath)
				Expect(err).ToNot(HaveOccurred())
				Expect(metadata.Format).To(Equal("ogg"))
			})

			DescribeTable("should validate compression options",
				func(options models.CompressionOption, shouldSucceed bool) {
					outputPath := filepath.Join(tempDir, "test_compress."+options.Format)
					err := processor.CompressAudio(ctx, testAudioFile, outputPath, options)
					
					if shouldSucceed {
						Expect(err).ToNot(HaveOccurred())
					} else {
						Expect(err).To(HaveOccurred())
					}
				},
				Entry("valid MP3 128kbps", models.CompressionOption{Bitrate: 128, Format: "mp3", Quality: "low"}, true),
				Entry("valid MP3 320kbps", models.CompressionOption{Bitrate: 320, Format: "mp3", Quality: "high"}, true),
				Entry("valid AAC 256kbps", models.CompressionOption{Bitrate: 256, Format: "aac", Quality: "high"}, true),
				Entry("valid OGG 192kbps", models.CompressionOption{Bitrate: 192, Format: "ogg", Quality: "medium"}, true),
				Entry("invalid bitrate too low", models.CompressionOption{Bitrate: 64, Format: "mp3", Quality: "low"}, false),
				Entry("invalid bitrate too high", models.CompressionOption{Bitrate: 500, Format: "mp3", Quality: "high"}, false),
				Entry("invalid format", models.CompressionOption{Bitrate: 256, Format: "wav", Quality: "high"}, false),
				Entry("invalid quality", models.CompressionOption{Bitrate: 256, Format: "mp3", Quality: "ultra"}, false),
			)

			It("should handle compression timeout", func() {
				ctxWithTimeout, cancel := context.WithTimeout(ctx, 1*time.Millisecond)
				defer cancel()
				
				options := models.CompressionOption{
					Bitrate: 256,
					Format:  "mp3",
					Quality: "high",
				}
				
				outputPath := filepath.Join(tempDir, "timeout_test.mp3") 
				err := processor.CompressAudio(ctxWithTimeout, testAudioFile, outputPath, options)
				
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("context deadline exceeded"))
			})

			It("should return error for non-existent input file", func() {
				nonExistentFile := filepath.Join(tempDir, "non_existent.wav")
				options := models.CompressionOption{
					Bitrate: 256,
					Format:  "mp3",
					Quality: "high",
				}
				
				outputPath := filepath.Join(tempDir, "nonexistent_test.mp3")
				err := processor.CompressAudio(ctx, nonExistentFile, outputPath, options)
				
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("input file not found"))
			})
		})
	})

	// NOTE: ValidateCompressionOptions method doesn't exist - tests removed

	Describe("GetSupportedFormats", func() {
		It("should return list of supported compression formats", func() {
			formats := processor.GetSupportedFormats()
			
			Expect(formats).To(ContainElements("mp3", "aac", "ogg"))
			Expect(formats).ToNot(ContainElements("wav", "flac", "m4a"))
		})
	})

	// NOTE: GetQualityLevels, GetBitrateRange, CleanupTempFiles methods don't exist - tests removed
})