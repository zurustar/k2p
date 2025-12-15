package imageprocessing

import (
	"image"
	"image/color"
	"testing"
)

func TestTrimBorders(t *testing.T) {
	t.Run("image with black borders", func(t *testing.T) {
		// Create test image: 100x100 with 10px black border
		img := createTestImageWithBorder(100, 100, 10, color.Black, color.White)

		trimmed := TrimBorders(img)
		bounds := trimmed.Bounds()

		// Should trim to approximately 80x80 (100 - 2*10)
		expectedSize := 80
		tolerance := 5

		if abs32(bounds.Dx(), expectedSize) > tolerance {
			t.Errorf("expected width ~%d, got %d", expectedSize, bounds.Dx())
		}
		if abs32(bounds.Dy(), expectedSize) > tolerance {
			t.Errorf("expected height ~%d, got %d", expectedSize, bounds.Dy())
		}
	})

	t.Run("image with white borders", func(t *testing.T) {
		img := createTestImageWithBorder(100, 100, 10, color.White, color.Black)

		trimmed := TrimBorders(img)
		bounds := trimmed.Bounds()

		expectedSize := 80
		tolerance := 5

		if abs32(bounds.Dx(), expectedSize) > tolerance {
			t.Errorf("expected width ~%d, got %d", expectedSize, bounds.Dx())
		}
	})

	t.Run("image without borders", func(t *testing.T) {
		// Create uniform image (no borders)
		img := image.NewRGBA(image.Rect(0, 0, 100, 100))
		for y := 0; y < 100; y++ {
			for x := 0; x < 100; x++ {
				img.Set(x, y, color.RGBA{128, 128, 128, 255})
			}
		}

		trimmed := TrimBorders(img)
		bounds := trimmed.Bounds()

		// Should return similar size (no significant trim)
		if bounds.Dx() < 90 || bounds.Dy() < 90 {
			t.Error("image without borders was over-trimmed")
		}
	})
}

func TestDetectBorderColor(t *testing.T) {
	t.Run("black corners", func(t *testing.T) {
		img := createTestImageWithBorder(100, 100, 10, color.Black, color.White)
		borderColor := detectBorderColor(img)

		if borderColor == nil {
			t.Error("expected to detect border color")
		}

		if !isBlackish(borderColor) {
			t.Error("expected black border color")
		}
	})

	t.Run("white corners", func(t *testing.T) {
		img := createTestImageWithBorder(100, 100, 10, color.White, color.Black)
		borderColor := detectBorderColor(img)

		if borderColor == nil {
			t.Error("expected to detect border color")
		}

		if !isWhitish(borderColor) {
			t.Error("expected white border color")
		}
	})

	t.Run("mixed corners", func(t *testing.T) {
		// Create image with different colored corners
		img := image.NewRGBA(image.Rect(0, 0, 100, 100))
		// Different colors in corners
		img.Set(0, 0, color.Black)
		img.Set(99, 0, color.White)
		img.Set(0, 99, color.RGBA{128, 128, 128, 255})
		img.Set(99, 99, color.RGBA{64, 64, 64, 255})

		borderColor := detectBorderColor(img)

		if borderColor != nil {
			t.Error("should not detect border color for mixed corners")
		}
	})
}

func TestIsSimilarColor(t *testing.T) {
	tests := []struct {
		name     string
		c1       color.Color
		c2       color.Color
		expected bool
	}{
		{"identical black", color.Black, color.Black, true},
		{"identical white", color.White, color.White, true},
		{"similar grays", color.RGBA{100, 100, 100, 255}, color.RGBA{110, 110, 110, 255}, true},
		{"different colors", color.Black, color.White, false},
		{"slightly different", color.RGBA{100, 100, 100, 255}, color.RGBA{120, 100, 100, 255}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isSimilarColor(tt.c1, tt.c2)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestIsBlackish(t *testing.T) {
	tests := []struct {
		name     string
		c        color.Color
		expected bool
	}{
		{"pure black", color.Black, true},
		{"dark gray", color.RGBA{30, 30, 30, 255}, true},
		{"medium gray", color.RGBA{128, 128, 128, 255}, false},
		{"white", color.White, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isBlackish(tt.c)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestIsWhitish(t *testing.T) {
	tests := []struct {
		name     string
		c        color.Color
		expected bool
	}{
		{"pure white", color.White, true},
		{"light gray", color.RGBA{220, 220, 220, 255}, true},
		{"medium gray", color.RGBA{128, 128, 128, 255}, false},
		{"black", color.Black, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isWhitish(tt.c)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// Helper function to create test image with border
func createTestImageWithBorder(width, height, borderSize int, borderColor, fillColor color.Color) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if x < borderSize || x >= width-borderSize ||
				y < borderSize || y >= height-borderSize {
				img.Set(x, y, borderColor)
			} else {
				img.Set(x, y, fillColor)
			}
		}
	}

	return img
}

// Helper function for absolute difference
func abs32(a, b int) int {
	if a > b {
		return a - b
	}
	return b - a
}
