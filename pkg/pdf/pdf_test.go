package pdf

import (
	"os"
	"testing"
)

func TestCreatePDF(t *testing.T) {
	generator := NewPDFGenerator()

	t.Run("no images", func(t *testing.T) {
		err := generator.CreatePDF([]string{}, "output.pdf", PDFOptions{})
		if err == nil {
			t.Error("expected error for empty image list")
		}
	})

	t.Run("non-existent image", func(t *testing.T) {
		err := generator.CreatePDF([]string{"/nonexistent/image.jpg"}, "output.pdf", PDFOptions{})
		if err == nil {
			t.Error("expected error for non-existent image")
		}
	})

	t.Run("create PDF from test images", func(t *testing.T) {
		// Create temporary test images
		tmpDir, err := os.MkdirTemp("", "pdf-test-*")
		if err != nil {
			t.Fatalf("failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tmpDir)

		// For this test to work, we'd need actual image files
		// Skipping actual PDF creation test for now
		t.Skip("Requires actual image files")
	})
}

func TestGetQualitySettings(t *testing.T) {
	tests := []struct {
		quality      string
		wantCompress bool
	}{
		{"low", true},
		{"medium", true},
		{"high", false},
		{"unknown", false}, // defaults to high
	}

	for _, tt := range tests {
		t.Run(tt.quality, func(t *testing.T) {
			opts := GetQualitySettings(tt.quality)
			if opts.Compression != tt.wantCompress {
				t.Errorf("quality %s: expected compression=%v, got %v",
					tt.quality, tt.wantCompress, opts.Compression)
			}
		})
	}
}

func TestPDFOptionsValidation(t *testing.T) {
	t.Run("valid quality levels", func(t *testing.T) {
		qualities := []string{"low", "medium", "high"}
		for _, q := range qualities {
			opts := GetQualitySettings(q)
			if opts.Quality != q {
				t.Errorf("expected quality %s, got %s", q, opts.Quality)
			}
		}
	})

	t.Run("default quality", func(t *testing.T) {
		opts := GetQualitySettings("")
		if opts.Quality != "high" {
			t.Errorf("expected default quality 'high', got %s", opts.Quality)
		}
	})
}

// Integration test for PDF creation
func TestCreatePDFIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// This test would require actual image files
	// For now, we'll create a placeholder test
	t.Skip("Integration test requires actual image files")
}
