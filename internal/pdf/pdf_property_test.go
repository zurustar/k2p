package pdf

import (
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// Helper to create a valid dummy image
func createDummyImage(path string, width, height int, format string) error {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	// Fill with some color
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			img.Set(x, y, color.RGBA{uint8(x % 255), uint8(y % 255), 100, 255})
		}
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	if format == "png" {
		return png.Encode(f, img)
	}
	return jpeg.Encode(f, img, nil)
}

// Property 1: Valid PDF Output
// For any currently open book (represented by a sequence of images), if conversion succeeds,
// the output must be a valid PDF file.
func TestProperty1_ValidPDFOutput(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 50 // Reduce count slightly due to IO overhead
	properties := gopter.NewProperties(parameters)

	properties.Property("CreatePDF produces valid PDF from valid images", prop.ForAll(
		func(imageFormats []string) bool {
			// Setup temp environment
			tmpDir, err := os.MkdirTemp("", "pdf-prop-test-*")
			if err != nil {
				return false
			}
			defer os.RemoveAll(tmpDir)

			var imageFiles []string
			for i, fmt := range imageFormats {
				fname := filepath.Join(tmpDir, "page_%d."+fmt)
				// Create image with 1-based index naming but simple name here
				fname = filepath.Join(tmpDir, fmt+string(rune(i+'a'))+"."+fmt) // simple unique name
				if err := createDummyImage(fname, 10, 10, fmt); err != nil {
					return false
				}
				imageFiles = append(imageFiles, fname)
			}

			outputPdf := filepath.Join(tmpDir, "output.pdf")
			generator := NewPDFGenerator()

			// Use default high quality
			opts := GetQualitySettings("high")

			err = generator.CreatePDF(imageFiles, outputPdf, opts)
			if err != nil {
				// Should not fail for valid images
				t.Logf("CreatePDF failed: %v", err)
				return false
			}

			// Verify file exists
			info, err := os.Stat(outputPdf)
			if err != nil || info.Size() == 0 {
				return false
			}

			// verify header
			data, err := os.ReadFile(outputPdf)
			if err != nil {
				return false
			}
			if len(data) < 4 || string(data[:4]) != "%PDF" {
				return false
			}

			return true
		},
		gen.SliceOf(
			gen.OneConstOf("png", "jpg"),
		).SuchThat(func(v interface{}) bool {
			return len(v.([]string)) > 0 // Ensure at least one image
		}),
	))

	properties.TestingRun(t)
}
