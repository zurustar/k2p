package imageprocessing

import (
	"image"
	"image/color"
	"testing"
)

func TestCalculateTrimMargins(t *testing.T) {
	t.Run("image with black borders", func(t *testing.T) {
		// Create test image: 100x100 with 10px black border
		img := createTestImageWithBorder(100, 100, 10, color.Black, color.White)

		margins := CalculateTrimMargins(img)

		// Margins should be 10 on each side
		expectedMargin := 10
		tolerance := 2

		if abs32(margins.Top, expectedMargin) > tolerance {
			t.Errorf("expected top margin ~%d, got %d", expectedMargin, margins.Top)
		}
		if abs32(margins.Bottom, expectedMargin) > tolerance {
			t.Errorf("expected bottom margin ~%d, got %d", expectedMargin, margins.Bottom)
		}
	})

	t.Run("image with white borders", func(t *testing.T) {
		img := createTestImageWithBorder(100, 100, 10, color.White, color.Black)

		margins := CalculateTrimMargins(img)

		expectedMargin := 10
		tolerance := 2

		if abs32(margins.Top, expectedMargin) > tolerance {
			t.Errorf("expected top margin ~%d, got %d", expectedMargin, margins.Top)
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

		margins := CalculateTrimMargins(img)

		// Should find no margins
		if margins.Top > 2 || margins.Bottom > 2 || margins.Left > 2 || margins.Right > 2 {
			t.Errorf("image without borders had incorrect margins: %+v", margins)
		}
	})

	t.Run("image with top-only black border", func(t *testing.T) {
		// Create image with black border only at top (like menu bar)
		img := image.NewRGBA(image.Rect(0, 0, 100, 100))
		for y := 0; y < 100; y++ {
			for x := 0; x < 100; x++ {
				if y < 10 {
					// Top 10 pixels are black
					img.Set(x, y, color.Black)
				} else {
					// Rest is white
					img.Set(x, y, color.White)
				}
			}
		}

		margins := CalculateTrimMargins(img)

		// Should trim top border ~10
		expectedMargin := 10
		tolerance := 2

		if abs32(margins.Top, expectedMargin) > tolerance {
			t.Errorf("expected top margin ~%d, got %d", expectedMargin, margins.Top)
		}

		// Other margins should be ~0
		if margins.Bottom > tolerance {
			t.Errorf("expected bottom margin ~0, got %d", margins.Bottom)
		}
	})
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
