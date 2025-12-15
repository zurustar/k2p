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

		trimmed := TrimBorders(img)
		bounds := trimmed.Bounds()

		// Should trim top border, height should be ~90
		expectedHeight := 90
		tolerance := 5

		if abs32(bounds.Dy(), expectedHeight) > tolerance {
			t.Errorf("expected height ~%d, got %d", expectedHeight, bounds.Dy())
		}

		// Width should remain ~100
		if abs32(bounds.Dx(), 100) > tolerance {
			t.Errorf("expected width ~100, got %d", bounds.Dx())
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
