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
					valid := processor.ValidateAudioFile(filename)
					Expect(valid).To(BeTrue(), "Format %s should be supported", format)
				}
			})

			It("should reject unsupported audio formats", func() {
				unsupportedFormats := []string{".txt", ".jpg", ".mp4", ".avi", ".pdf"}
				
				for _, format := range unsupportedFormats {
					filename := "test_file" + format
					valid := processor.ValidateAudioFile(filename)
					Expect(valid).To(BeFalse(), "Format %s should not be supported", format)
				}
			})

			It("should handle case-insensitive file extensions", func() {
				caseVariations := []string{".MP3", ".Mp3", ".WAV", ".Wav", ".AAC", ".aAc"}
				
				for _, format := range caseVariations {
					filename := "test_audio" + format
					valid := processor.ValidateAudioFile(filename)
					Expect(valid).To(BeTrue(), "Format %s should be supported (case insensitive)", format)
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
				
				outputPath, err := processor.CompressAudio(ctx, testAudioFile, options)
				
				Expect(err).ToNot(HaveOccurred())
				Expect(outputPath).ToNot(BeEmpty())
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
				
				outputPath, err := processor.CompressAudio(ctx, testAudioFile, options)
				
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
				
				outputPath, err := processor.CompressAudio(ctx, testAudioFile, options)
				
				Expect(err).ToNot(HaveOccurred())
				Expect(filepath.Ext(outputPath)).To(Equal(".ogg"))
				
				metadata, err := processor.ExtractMetadata(ctx, outputPath)
				Expect(err).ToNot(HaveOccurred())
				Expect(metadata.Format).To(Equal("ogg"))
			})

			DescribeTable("should validate compression options",
				func(options models.CompressionOption, shouldSucceed bool) {
					_, err := processor.CompressAudio(ctx, testAudioFile, options)
					
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
				
				_, err := processor.CompressAudio(ctxWithTimeout, testAudioFile, options)
				
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
				
				_, err := processor.CompressAudio(ctx, nonExistentFile, options)
				
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("input file not found"))
			})
		})
	})

	Describe("ValidateCompressionOptions", func() {
		Context("when validating compression options", func() {
			It("should validate correct bitrate ranges for different formats", func() {
				validOptions := []models.CompressionOption{
					{Bitrate: 128, Format: "mp3", Quality: "medium", SampleRate: 44100},
					{Bitrate: 256, Format: "aac", Quality: "high", SampleRate: 48000},
					{Bitrate: 192, Format: "ogg", Quality: "medium", SampleRate: 44100},
					{Bitrate: 320, Format: "mp3", Quality: "high", SampleRate: 44100},
				}
				
				for _, opts := range validOptions {
					err := processor.ValidateCompressionOptions(opts)
					Expect(err).ToNot(HaveOccurred(), "Options should be valid: %+v", opts)
				}
			})

			It("should reject invalid bitrates", func() {
				invalidBitrates := []int{32, 64, 500, 1000}
				
				for _, bitrate := range invalidBitrates {
					opts := models.CompressionOption{
						Bitrate: bitrate,
						Format:  "mp3",
						Quality: "medium",
					}
					
					err := processor.ValidateCompressionOptions(opts)
					Expect(err).To(HaveOccurred(), "Bitrate %d should be invalid", bitrate)
				}
			})

			It("should reject unsupported formats", func() {
				unsupportedFormats := []string{"wav", "flac", "m4a", "wma", "ape"}
				
				for _, format := range unsupportedFormats {
					opts := models.CompressionOption{
						Bitrate: 256,
						Format:  format,
						Quality: "high",
					}
					
					err := processor.ValidateCompressionOptions(opts)
					Expect(err).To(HaveOccurred(), "Format %s should be unsupported for compression", format)
				}
			})

			It("should reject invalid quality levels", func() {
				invalidQualities := []string{"ultra", "maximum", "poor", "terrible", ""}
				
				for _, quality := range invalidQualities {
					opts := models.CompressionOption{
						Bitrate: 256,
						Format:  "mp3",
						Quality: quality,
					}
					
					err := processor.ValidateCompressionOptions(opts)
					Expect(err).To(HaveOccurred(), "Quality %s should be invalid", quality)
				}
			})

			It("should validate sample rates", func() {
				validSampleRates := []int{44100, 48000, 88200, 96000}
				invalidSampleRates := []int{22050, 32000, 176400, 192000}
				
				for _, sampleRate := range validSampleRates {
					opts := models.CompressionOption{
						Bitrate:    256,
						Format:     "mp3",
						Quality:    "high",
						SampleRate: sampleRate,
					}
					
					err := processor.ValidateCompressionOptions(opts)
					Expect(err).ToNot(HaveOccurred(), "Sample rate %d should be valid", sampleRate)
				}
				
				for _, sampleRate := range invalidSampleRates {
					opts := models.CompressionOption{
						Bitrate:    256,
						Format:     "mp3",
						Quality:    "high",
						SampleRate: sampleRate,
					}
					
					err := processor.ValidateCompressionOptions(opts)
					Expect(err).To(HaveOccurred(), "Sample rate %d should be invalid", sampleRate)
				}
			})
		})
	})

	Describe("GetSupportedFormats", func() {
		It("should return list of supported compression formats", func() {
			formats := processor.GetSupportedFormats()
			
			Expect(formats).To(ContainElements("mp3", "aac", "ogg"))
			Expect(formats).ToNot(ContainElements("wav", "flac", "m4a"))
		})
	})

	Describe("GetQualityLevels", func() {
		It("should return list of supported quality levels", func() {
			qualities := processor.GetQualityLevels()
			
			Expect(qualities).To(ContainElements("low", "medium", "high"))
			Expect(qualities).ToNot(ContainElements("ultra", "maximum", "poor"))
		})
	})

	Describe("GetBitrateRange", func() {
		It("should return valid bitrate range for each format", func() {
			bitrateRanges := map[string][2]int{
				"mp3": {128, 320},
				"aac": {128, 256},
				"ogg": {128, 256},
			}
			
			for format, expectedRange := range bitrateRanges {
				min, max := processor.GetBitrateRange(format)
				Expect(min).To(Equal(expectedRange[0]), "Min bitrate for %s should be %d", format, expectedRange[0])
				Expect(max).To(Equal(expectedRange[1]), "Max bitrate for %s should be %d", format, expectedRange[1])
			}
		})

		It("should return zero for unsupported formats", func() {
			min, max := processor.GetBitrateRange("unsupported")
			Expect(min).To(Equal(0))
			Expect(max).To(Equal(0))
		})
	})

	Describe("CleanupTempFiles", func() {
		It("should clean up temporary files older than specified duration", func() {
			// Create some temporary files
			tempFile1 := filepath.Join(tempDir, "temp1.mp3")
			tempFile2 := filepath.Join(tempDir, "temp2.aac")
			
			err := os.WriteFile(tempFile1, []byte("temp content"), 0644)
			Expect(err).ToNot(HaveOccurred())
			err = os.WriteFile(tempFile2, []byte("temp content"), 0644)
			Expect(err).ToNot(HaveOccurred())
			
			// Change modification time to make them appear old
			oldTime := time.Now().Add(-2 * time.Hour)
			err = os.Chtimes(tempFile1, oldTime, oldTime)
			Expect(err).ToNot(HaveOccurred())
			
			// Clean up files older than 1 hour
			err = processor.CleanupTempFiles(1 * time.Hour)
			Expect(err).ToNot(HaveOccurred())
			
			// tempFile1 should be deleted, tempFile2 should remain
			_, err = os.Stat(tempFile1)
			Expect(os.IsNotExist(err)).To(BeTrue())
			
			_, err = os.Stat(tempFile2)
			Expect(err).ToNot(HaveOccurred())
		})
	})
})