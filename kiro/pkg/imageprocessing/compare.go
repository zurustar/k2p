package imageprocessing

import (
	"image/png"
	"os"
)

// CompareImages compares two images and returns similarity score (0.0 to 1.0)
// Higher score means more similar
func CompareImages(img1Path, img2Path string) (float64, error) {
	// Load first image
	file1, err := os.Open(img1Path)
	if err != nil {
		return 0, err
	}
	defer file1.Close()

	img1, err := png.Decode(file1)
	if err != nil {
		return 0, err
	}

	// Load second image
	file2, err := os.Open(img2Path)
	if err != nil {
		return 0, err
	}
	defer file2.Close()

	img2, err := png.Decode(file2)
	if err != nil {
		return 0, err
	}

	// Check if dimensions match
	bounds1 := img1.Bounds()
	bounds2 := img2.Bounds()

	if bounds1.Dx() != bounds2.Dx() || bounds1.Dy() != bounds2.Dy() {
		return 0, nil
	}

	// Compare pixels
	// Sample every 10th pixel for performance
	matchCount := 0
	totalCount := 0

	for y := bounds1.Min.Y; y < bounds1.Max.Y; y += 10 {
		for x := bounds1.Min.X; x < bounds1.Max.X; x += 10 {
			totalCount++

			r1, g1, b1, _ := img1.At(x, y).RGBA()
			r2, g2, b2, _ := img2.At(x, y).RGBA()

			// Convert to 8-bit
			r1, g1, b1 = r1>>8, g1>>8, b1>>8
			r2, g2, b2 = r2>>8, g2>>8, b2>>8

			// Check if similar (within threshold)
			threshold := uint32(30)
			if absUint32(r1, r2) <= threshold &&
				absUint32(g1, g2) <= threshold &&
				absUint32(b1, b2) <= threshold {
				matchCount++
			}
		}
	}

	// Return similarity score
	similarity := float64(matchCount) / float64(totalCount)
	return similarity, nil
}

// absUint32 returns absolute difference
func absUint32(a, b uint32) uint32 {
	if a > b {
		return a - b
	}
	return b - a
}
