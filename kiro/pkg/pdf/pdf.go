package pdf

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jung-kurt/gofpdf"
)

// PDFGenerator handles PDF generation from images
type PDFGenerator interface {
	// CreatePDF creates a PDF from a sequence of image files
	CreatePDF(imageFiles []string, outputPath string, options PDFOptions) error
}

// PDFOptions contains options for PDF generation
type PDFOptions struct {
	// Quality setting: "low", "medium", "high"
	Quality string

	// Enable compression
	Compression bool
}

// DefaultPDFGenerator is the default implementation using gofpdf
type DefaultPDFGenerator struct{}

// NewPDFGenerator creates a new PDFGenerator instance
func NewPDFGenerator() PDFGenerator {
	return &DefaultPDFGenerator{}
}

// CreatePDF creates a PDF from a sequence of image files
func (g *DefaultPDFGenerator) CreatePDF(imageFiles []string, outputPath string, options PDFOptions) error {
	if len(imageFiles) == 0 {
		return fmt.Errorf("no images provided")
	}

	// Validate all image files exist
	for _, imgPath := range imageFiles {
		if _, err := os.Stat(imgPath); err != nil {
			return fmt.Errorf("image file not found: %s", imgPath)
		}
	}

	// Create PDF without specifying page size (we'll set it per page)
	pdf := gofpdf.New("P", "pt", "", "")

	// Set compression based on options
	if options.Compression {
		pdf.SetCompression(true)
	}

	// Add each image as a page
	for _, imgPath := range imageFiles {
		// Get image type from extension
		ext := filepath.Ext(imgPath)
		var imgType string
		switch ext {
		case ".jpg", ".jpeg":
			imgType = "JPEG"
		case ".png":
			imgType = "PNG"
		default:
			return fmt.Errorf("unsupported image format: %s", ext)
		}

		// Register image to get dimensions
		opts := gofpdf.ImageOptions{
			ImageType: imgType,
			ReadDpi:   true,
		}

		info := pdf.RegisterImageOptions(imgPath, opts)
		if pdf.Error() != nil {
			return fmt.Errorf("failed to register image %s: %w", imgPath, pdf.Error())
		}

		// Get image dimensions in points
		imgWidth := info.Width()
		imgHeight := info.Height()

		// Add page with image dimensions
		pdf.AddPageFormat("P", gofpdf.SizeType{Wd: imgWidth, Ht: imgHeight})

		// Add image to fill the page exactly
		pdf.ImageOptions(imgPath, 0, 0, imgWidth, imgHeight, false, opts, 0, "")
	}

	// Output PDF
	if err := pdf.OutputFileAndClose(outputPath); err != nil {
		return fmt.Errorf("failed to create PDF: %w", err)
	}

	return nil
}

// GetQualitySettings returns compression settings based on quality level
func GetQualitySettings(quality string) PDFOptions {
	switch quality {
	case "low":
		return PDFOptions{
			Quality:     "low",
			Compression: true,
		}
	case "medium":
		return PDFOptions{
			Quality:     "medium",
			Compression: true,
		}
	case "high":
		return PDFOptions{
			Quality:     "high",
			Compression: false,
		}
	default:
		// Default to high quality
		return PDFOptions{
			Quality:     "high",
			Compression: false,
		}
	}
}
