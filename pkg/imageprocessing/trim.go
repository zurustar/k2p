package imageprocessing

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
)

// TrimBorders removes uniform colored borders (black or white) from an image
func TrimBorders(img image.Image) image.Image {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	if width == 0 || height == 0 {
		return img
	}

	// Determine border color from corners
	borderColor := detectBorderColor(img)
	if borderColor == nil {
		// No clear border color, return original
		return img
	}

	// Find trim boundaries
	top := findTopBorder(img, borderColor)
	bottom := findBottomBorder(img, borderColor)
	left := findLeftBorder(img, borderColor)
	right := findRightBorder(img, borderColor)

	// Validate boundaries
	if top >= bottom || left >= right {
		// Invalid trim, return original
		return img
	}

	// Create trimmed image
	trimmedBounds := image.Rect(0, 0, right-left, bottom-top)
	trimmed := image.NewRGBA(trimmedBounds)

	for y := top; y < bottom; y++ {
		for x := left; x < right; x++ {
			trimmed.Set(x-left, y-top, img.At(x, y))
		}
	}

	return trimmed
}

// detectBorderColor determines the border color from image corners
func detectBorderColor(img image.Image) color.Color {
	bounds := img.Bounds()

	// Sample corners
	topLeft := img.At(bounds.Min.X, bounds.Min.Y)
	topRight := img.At(bounds.Max.X-1, bounds.Min.Y)
	bottomLeft := img.At(bounds.Min.X, bounds.Max.Y-1)
	bottomRight := img.At(bounds.Max.X-1, bounds.Max.Y-1)

	// Check if corners are similar (black or white)
	if isSimilarColor(topLeft, topRight) &&
		isSimilarColor(topLeft, bottomLeft) &&
		isSimilarColor(topLeft, bottomRight) {

		// Check if it's black or white
		if isBlackish(topLeft) || isWhitish(topLeft) {
			return topLeft
		}
	}

	return nil
}

// findTopBorder finds the top border edge
func findTopBorder(img image.Image, borderColor color.Color) int {
	bounds := img.Bounds()
	threshold := 0.9 // 90% of pixels must match

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		matchCount := 0
		total := 0

		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			total++
			if isSimilarColor(img.At(x, y), borderColor) {
				matchCount++
			}
		}

		if float64(matchCount)/float64(total) < threshold {
			return y
		}
	}

	return bounds.Min.Y
}

// findBottomBorder finds the bottom border edge
func findBottomBorder(img image.Image, borderColor color.Color) int {
	bounds := img.Bounds()
	threshold := 0.9

	for y := bounds.Max.Y - 1; y >= bounds.Min.Y; y-- {
		matchCount := 0
		total := 0

		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			total++
			if isSimilarColor(img.At(x, y), borderColor) {
				matchCount++
			}
		}

		if float64(matchCount)/float64(total) < threshold {
			return y + 1
		}
	}

	return bounds.Max.Y
}

// findLeftBorder finds the left border edge
func findLeftBorder(img image.Image, borderColor color.Color) int {
	bounds := img.Bounds()
	threshold := 0.9

	for x := bounds.Min.X; x < bounds.Max.X; x++ {
		matchCount := 0
		total := 0

		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			total++
			if isSimilarColor(img.At(x, y), borderColor) {
				matchCount++
			}
		}

		if float64(matchCount)/float64(total) < threshold {
			return x
		}
	}

	return bounds.Min.X
}

// findRightBorder finds the right border edge
func findRightBorder(img image.Image, borderColor color.Color) int {
	bounds := img.Bounds()
	threshold := 0.9

	for x := bounds.Max.X - 1; x >= bounds.Min.X; x-- {
		matchCount := 0
		total := 0

		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			total++
			if isSimilarColor(img.At(x, y), borderColor) {
				matchCount++
			}
		}

		if float64(matchCount)/float64(total) < threshold {
			return x + 1
		}
	}

	return bounds.Max.X
}

// isSimilarColor checks if two colors are similar
func isSimilarColor(c1, c2 color.Color) bool {
	r1, g1, b1, _ := c1.RGBA()
	r2, g2, b2, _ := c2.RGBA()

	// Convert to 8-bit values
	r1, g1, b1 = r1>>8, g1>>8, b1>>8
	r2, g2, b2 = r2>>8, g2>>8, b2>>8

	// Allow small difference (threshold)
	threshold := uint32(30)

	return abs(r1, r2) <= threshold &&
		abs(g1, g2) <= threshold &&
		abs(b1, b2) <= threshold
}

// isBlackish checks if a color is close to black
func isBlackish(c color.Color) bool {
	r, g, b, _ := c.RGBA()
	r, g, b = r>>8, g>>8, b>>8

	threshold := uint32(50)
	return r < threshold && g < threshold && b < threshold
}

// isWhitish checks if a color is close to white
func isWhitish(c color.Color) bool {
	r, g, b, _ := c.RGBA()
	r, g, b = r>>8, g>>8, b>>8

	threshold := uint32(200)
	return r > threshold && g > threshold && b > threshold
}

// abs returns absolute difference between two uint32 values
func abs(a, b uint32) uint32 {
	if a > b {
		return a - b
	}
	return b - a
}

// TrimImageFile trims borders from an image file and saves the result
func TrimImageFile(inputPath, outputPath string) error {
	// Open input file
	file, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("failed to open input file: %w", err)
	}
	defer file.Close()

	// Decode image
	img, err := png.Decode(file)
	if err != nil {
		return fmt.Errorf("failed to decode image: %w", err)
	}

	// Trim borders
	trimmed := TrimBorders(img)

	// Save trimmed image
	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	if err := png.Encode(outFile, trimmed); err != nil {
		return fmt.Errorf("failed to encode image: %w", err)
	}

	return nil
}
