package utils_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/wavlake/monorepo/internal/models"
	"github.com/wavlake/monorepo/internal/utils"
)

func TestUtils(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Utils Suite")
}

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
		It("should return list of supported audio formats", func() {
			formats := processor.GetSupportedFormats()
			
			Expect(formats).To(ContainElements("mp3", "wav", "flac", "aac", "ogg", "m4a"))
			Expect(len(formats)).To(BeNumerically(">", 5), "Should support multiple audio formats")
		})
	})

	Describe("IsFormatSupported", func() {
		It("should correctly identify supported formats", func() {
			supportedFormats := []string{"mp3", "wav", "flac", "aac", "ogg", "m4a", "wma", "aiff", "au"}
			
			for _, format := range supportedFormats {
				Expect(processor.IsFormatSupported(format)).To(BeTrue(), "Format %s should be supported", format)
				Expect(processor.IsFormatSupported("."+format)).To(BeTrue(), "Format .%s should be supported", format)
				Expect(processor.IsFormatSupported(strings.ToUpper(format))).To(BeTrue(), "Format %s should be case insensitive", strings.ToUpper(format))
			}
		})

		It("should reject unsupported formats", func() {
			unsupportedFormats := []string{"txt", "jpg", "mp4", "avi", "pdf", "doc", "zip"}
			
			for _, format := range unsupportedFormats {
				Expect(processor.IsFormatSupported(format)).To(BeFalse(), "Format %s should not be supported", format)
				Expect(processor.IsFormatSupported("."+format)).To(BeFalse(), "Format .%s should not be supported", format)
			}
		})

		It("should handle edge cases", func() {
			// Empty string
			Expect(processor.IsFormatSupported("")).To(BeFalse())
			
			// Only dot
			Expect(processor.IsFormatSupported(".")).To(BeFalse())
			
			// Multiple dots
			Expect(processor.IsFormatSupported("..mp3")).To(BeFalse())
		})
	})

	// NOTE: GetQualityLevels, GetBitrateRange, CleanupTempFiles methods don't exist - tests removed
})

var _ = Describe("StoragePathConfig", func() {
	var config *utils.StoragePathConfig

	BeforeEach(func() {
		config = utils.GetStoragePathConfig()
	})

	Describe("GetStoragePathConfig", func() {
		It("should return a valid configuration with default paths", func() {
			Expect(config).ToNot(BeNil())
			Expect(config.OriginalPrefix).To(Equal("tracks/original"))
			Expect(config.CompressedPrefix).To(Equal("tracks/compressed"))
			Expect(config.UseLegacyPaths).To(BeFalse())
		})
	})

	Describe("GetOriginalPath", func() {
		It("should generate correct path for original files", func() {
			trackID := "track-123"
			extension := "wav"
			
			path := config.GetOriginalPath(trackID, extension)
			
			Expect(path).To(Equal("tracks/original/track-123.wav"))
		})

		It("should handle different extensions", func() {
			trackID := "track-456"
			
			testCases := []struct {
				extension string
				expected  string
			}{
				{"mp3", "tracks/original/track-456.mp3"},
				{"flac", "tracks/original/track-456.flac"},
				{"aac", "tracks/original/track-456.aac"},
				{"m4a", "tracks/original/track-456.m4a"},
			}
			
			for _, tc := range testCases {
				path := config.GetOriginalPath(trackID, tc.extension)
				Expect(path).To(Equal(tc.expected), "Extension %s should generate path %s", tc.extension, tc.expected)
			}
		})

		It("should handle edge cases", func() {
			// Empty track ID
			path := config.GetOriginalPath("", "mp3")
			Expect(path).To(Equal("tracks/original/.mp3"))
			
			// Empty extension
			path = config.GetOriginalPath("track-789", "")
			Expect(path).To(Equal("tracks/original/track-789."))
			
			// Special characters in track ID
			path = config.GetOriginalPath("track_with-special.chars", "wav")
			Expect(path).To(Equal("tracks/original/track_with-special.chars.wav"))
		})
	})

	Describe("GetCompressedPath", func() {
		It("should generate correct path for compressed files", func() {
			trackID := "track-123"
			
			path := config.GetCompressedPath(trackID)
			
			Expect(path).To(Equal("tracks/compressed/track-123.mp3"))
		})

		It("should always use mp3 extension for compressed files", func() {
			trackIDs := []string{"track-001", "track-002", "track-999"}
			
			for _, trackID := range trackIDs {
				path := config.GetCompressedPath(trackID)
				Expect(path).To(HaveSuffix(".mp3"), "Compressed files should always be mp3")
				Expect(path).To(Equal(fmt.Sprintf("tracks/compressed/%s.mp3", trackID)))
			}
		})

		It("should handle edge cases", func() {
			// Empty track ID
			path := config.GetCompressedPath("")
			Expect(path).To(Equal("tracks/compressed/.mp3"))
			
			// Special characters in track ID
			path = config.GetCompressedPath("track_with-special.chars")
			Expect(path).To(Equal("tracks/compressed/track_with-special.chars.mp3"))
		})
	})

	Describe("GetCompressedVersionPath", func() {
		It("should generate correct path for versioned compressed files", func() {
			trackID := "track-123"
			versionID := "v1"
			format := "aac"
			
			path := config.GetCompressedVersionPath(trackID, versionID, format)
			
			Expect(path).To(Equal("tracks/compressed/track-123_v1.aac"))
		})

		It("should handle different formats", func() {
			trackID := "track-456"
			versionID := "hq"
			
			testCases := []struct {
				format   string
				expected string
			}{
				{"mp3", "tracks/compressed/track-456_hq.mp3"},
				{"aac", "tracks/compressed/track-456_hq.aac"},
				{"ogg", "tracks/compressed/track-456_hq.ogg"},
				{"flac", "tracks/compressed/track-456_hq.flac"},
			}
			
			for _, tc := range testCases {
				path := config.GetCompressedVersionPath(trackID, versionID, tc.format)
				Expect(path).To(Equal(tc.expected), "Format %s should generate path %s", tc.format, tc.expected)
			}
		})

		It("should handle edge cases", func() {
			// Empty parameters
			path := config.GetCompressedVersionPath("", "", "")
			Expect(path).To(Equal("tracks/compressed/_."))
			
			// Special characters
			path = config.GetCompressedVersionPath("track-123", "version_2.0", "mp3")
			Expect(path).To(Equal("tracks/compressed/track-123_version_2.0.mp3"))
		})
	})

	Describe("IsOriginalPath", func() {
		It("should correctly identify original paths", func() {
			validPaths := []string{
				"tracks/original/track-123.wav",
				"tracks/original/track-456.mp3",
				"tracks/original/some-long-track-id.flac",
			}
			
			for _, path := range validPaths {
				Expect(config.IsOriginalPath(path)).To(BeTrue(), "Path %s should be identified as original", path)
			}
		})

		It("should reject non-original paths", func() {
			invalidPaths := []string{
				"tracks/compressed/track-123.mp3",
				"tracks/original",                    // Missing filename
				"some/other/path/track-123.wav",
				"tracks/original",                    // Missing trailing slash and filename
				"",
			}
			
			for _, path := range invalidPaths {
				Expect(config.IsOriginalPath(path)).To(BeFalse(), "Path %s should not be identified as original", path)
			}
		})

		It("should handle edge cases", func() {
			// Path exactly matching prefix (no filename)
			Expect(config.IsOriginalPath("tracks/original/")).To(BeFalse())
			
			// Path with prefix but empty filename
			Expect(config.IsOriginalPath("tracks/original/.")).To(BeTrue())
			
			// Path that starts with prefix but is not in directory
			Expect(config.IsOriginalPath("tracks/original")).To(BeFalse())
		})
	})

	Describe("IsCompressedPath", func() {
		It("should correctly identify compressed paths", func() {
			validPaths := []string{
				"tracks/compressed/track-123.mp3",
				"tracks/compressed/track-456_v1.aac",
				"tracks/compressed/some-long-track-id.ogg",
			}
			
			for _, path := range validPaths {
				Expect(config.IsCompressedPath(path)).To(BeTrue(), "Path %s should be identified as compressed", path)
			}
		})

		It("should reject non-compressed paths", func() {
			invalidPaths := []string{
				"tracks/original/track-123.wav",
				"tracks/compressed",                  // Missing filename
				"some/other/path/track-123.mp3",
				"",
			}
			
			for _, path := range invalidPaths {
				Expect(config.IsCompressedPath(path)).To(BeFalse(), "Path %s should not be identified as compressed", path)
			}
		})

		It("should handle edge cases", func() {
			// Path exactly matching prefix (no filename)
			Expect(config.IsCompressedPath("tracks/compressed/")).To(BeFalse())
			
			// Path with prefix but empty filename
			Expect(config.IsCompressedPath("tracks/compressed/.")).To(BeTrue())
			
			// Path that starts with prefix but is not in directory
			Expect(config.IsCompressedPath("tracks/compressed")).To(BeFalse())
		})
	})

	Describe("GetTrackIDFromPath", func() {
		Context("for original paths", func() {
			It("should extract track ID from original file paths", func() {
				testCases := []struct {
					path     string
					expected string
				}{
					{"tracks/original/track-123.wav", "track-123"},
					{"tracks/original/simple.mp3", "simple"},
					{"tracks/original/track_with-special.chars.flac", "track"},  // Stops at first underscore
					{"tracks/original/123456789.aac", "123456789"},
					{"tracks/original/track-with-hyphens.mp3", "track-with-hyphens"},  // No underscores, stops at dot
				}
				
				for _, tc := range testCases {
					trackID := config.GetTrackIDFromPath(tc.path)
					Expect(trackID).To(Equal(tc.expected), "Path %s should extract track ID %s", tc.path, tc.expected)
				}
			})
		})

		Context("for compressed paths", func() {
			It("should extract track ID from compressed file paths", func() {
				testCases := []struct {
					path     string
					expected string
				}{
					{"tracks/compressed/track-123.mp3", "track-123"},
					{"tracks/compressed/simple.aac", "simple"},
					{"tracks/compressed/track-456_v1.ogg", "track-456"},
					{"tracks/compressed/track-789_hq.mp3", "track-789"},
				}
				
				for _, tc := range testCases {
					trackID := config.GetTrackIDFromPath(tc.path)
					Expect(trackID).To(Equal(tc.expected), "Path %s should extract track ID %s", tc.path, tc.expected)
				}
			})

			It("should handle versioned compressed files correctly", func() {
				versionedPaths := []string{
					"tracks/compressed/track-123_v1.mp3",
					"tracks/compressed/track-456_hq.aac",
					"tracks/compressed/track-789_low.ogg",
					"tracks/compressed/long-track-id_version_2.0.flac",
				}
				
				expectedIDs := []string{"track-123", "track-456", "track-789", "long-track-id"}
				
				for i, path := range versionedPaths {
					trackID := config.GetTrackIDFromPath(path)
					Expect(trackID).To(Equal(expectedIDs[i]), "Versioned path %s should extract track ID %s", path, expectedIDs[i])
				}
			})
		})

		Context("for invalid paths", func() {
			It("should return empty string for non-track paths", func() {
				invalidPaths := []string{
					"some/other/path/file.mp3",
					"tracks/other/track-123.wav",
					"tracks/original",
					"tracks/compressed",
					"",
					"invalid-path",
				}
				
				for _, path := range invalidPaths {
					trackID := config.GetTrackIDFromPath(path)
					Expect(trackID).To(Equal(""), "Invalid path %s should return empty track ID", path)
				}
			})
		})

		Context("for edge cases", func() {
			It("should handle files without extensions", func() {
				paths := []string{
					"tracks/original/track-123",
					"tracks/compressed/track-456",
				}
				
				expectedIDs := []string{"track-123", "track-456"}
				
				for i, path := range paths {
					trackID := config.GetTrackIDFromPath(path)
					Expect(trackID).To(Equal(expectedIDs[i]), "Path without extension %s should extract track ID %s", path, expectedIDs[i])
				}
			})

			It("should handle empty filenames gracefully", func() {
				paths := []string{
					"tracks/original/.",
					"tracks/compressed/.",
				}
				
				for _, path := range paths {
					trackID := config.GetTrackIDFromPath(path)
					Expect(trackID).To(Equal(""), "Path with empty filename %s should return empty track ID", path)
				}
			})
		})
	})
})