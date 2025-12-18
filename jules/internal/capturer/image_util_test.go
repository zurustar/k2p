package capturer

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"testing"
)

func createTestImage(path string, color color.RGBA) error {
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			img.Set(x, y, color)
		}
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

func TestImagesAreIdentical(t *testing.T) {
	tmpDir := t.TempDir()
	img1Path := tmpDir + "/img1.png"
	img2Path := tmpDir + "/img2.png"
	img3Path := tmpDir + "/img3.png"

	red := color.RGBA{255, 0, 0, 255}
	blue := color.RGBA{0, 0, 255, 255}

	createTestImage(img1Path, red)
	createTestImage(img2Path, red)
	createTestImage(img3Path, blue)

	tests := []struct {
		name     string
		path1    string
		path2    string
		expected bool
	}{
		{"Identical", img1Path, img2Path, true},
		{"Different", img1Path, img3Path, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ImagesAreIdentical(tt.path1, tt.path2)
			if err != nil {
				t.Fatalf("ImagesAreIdentical() error = %v", err)
			}
			if got != tt.expected {
				t.Errorf("ImagesAreIdentical() = %v, want %v", got, tt.expected)
			}
		})
	}
}
