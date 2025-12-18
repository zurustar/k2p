package capturer

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"os"
)

// ImagesAreIdentical compares two images at the given paths.
// It returns true if they are identical (pixel-wise), false otherwise.
func ImagesAreIdentical(path1, path2 string) (bool, error) {
	img1, err := loadImage(path1)
	if err != nil {
		return false, fmt.Errorf("failed to load %s: %w", path1, err)
	}

	img2, err := loadImage(path2)
	if err != nil {
		return false, fmt.Errorf("failed to load %s: %w", path2, err)
	}

	return compareImages(img1, img2), nil
}

func loadImage(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	img, err := png.Decode(f)
	if err != nil {
		return nil, err
	}
	return img, nil
}

func compareImages(img1, img2 image.Image) bool {
	b1 := img1.Bounds()
	b2 := img2.Bounds()

	if b1 != b2 {
		return false
	}

	// Compare raw bytes if possible for speed?
	// No, image.Image hides underlying buffer.
	// We can cast to specific types (NRGBA etc) but that's brittle.
	// A simple pixel loop is robust enough for screenshots (usually not huge resolution, and done once per page).
	// Optimization: Check center pixel, corners, then full scan.

	width := b1.Max.X
	height := b1.Max.Y

	// Quick check: Center
	cx, cy := width/2, height/2
	r1, g1, b1_val, a1 := img1.At(cx, cy).RGBA()
	r2, g2, b2_val, a2 := img2.At(cx, cy).RGBA()
	if r1 != r2 || g1 != g2 || b1_val != b2_val || a1 != a2 {
		return false
	}

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			r1, g1, b1_val, a1 := img1.At(x, y).RGBA()
			r2, g2, b2_val, a2 := img2.At(x, y).RGBA()

			if r1 != r2 || g1 != g2 || b1_val != b2_val || a1 != a2 {
				return false
			}
		}
	}

	return true
}

// ImagesAreIdenticalByBytes compares two files by content.
// This is faster but might fail if metadata changes (e.g. timestamp in PNG chunk).
// Screenshots usually have new timestamps. So pixel comparison is safer.
func ImagesAreIdenticalByBytes(path1, path2 string) (bool, error) {
	b1, err := os.ReadFile(path1)
	if err != nil {
		return false, err
	}
	b2, err := os.ReadFile(path2)
	if err != nil {
		return false, err
	}
	return bytes.Equal(b1, b2), nil
}
