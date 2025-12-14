package converter

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func createDummyImage(filename string, width, height int) error {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	// Fill with some color
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{R: uint8(x % 255), G: uint8(y % 255), B: 100, A: 255})
		}
	}
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

func TestConvertImagesToPdf(t *testing.T) {
	tempDir := "test_images"
	os.MkdirAll(tempDir, 0755)
	defer os.RemoveAll(tempDir)

	// Create 3 dummy images
	for i := 1; i <= 3; i++ {
		filename := filepath.Join(tempDir, "page_000"+string(rune('0'+i))+".png") // simple naming
		if i == 1 {
			filename = filepath.Join(tempDir, "page_0001.png")
		} else if i == 2 {
			filename = filepath.Join(tempDir, "page_0002.png")
		} else {
			filename = filepath.Join(tempDir, "page_0003.png")
		}

		if err := createDummyImage(filename, 100, 100); err != nil {
			t.Fatalf("Failed to create dummy image: %v", err)
		}
	}

	conv := NewPdfConverter()
	outputPdf := "test_output.pdf"
	defer os.Remove(outputPdf)

	// Test no max size
	err := conv.ConvertImagesToPdf(tempDir, outputPdf, 0)
	if err != nil {
		t.Fatalf("ConvertImagesToPdf failed: %v", err)
	}

	if _, err := os.Stat(outputPdf); os.IsNotExist(err) {
		t.Errorf("PDF file was not created")
	}

	// Test with max size (force split)
	// Make max size very small to force split
	outputPdfSplit := "test_output_split.pdf"
	// clean up potential parts
	defer func() {
		files, _ := filepath.Glob("test_output_split*.pdf")
		for _, f := range files {
			os.Remove(f)
		}
	}()

	// 100x100 png is roughly some KB. Set limit to 1 byte to force split every page?
	// Or just small enough.
	// Actually, the logic sums file sizes.
	// Let's get file size of one image.
	info, _ := os.Stat(filepath.Join(tempDir, "page_0001.png"))
	singleSize := info.Size()

	// Limit to size of 1.5 images -> should split.
	// 3 images. 1st fits. 2nd: 1+1 = 2 > 1.5 -> split (write 1st). Reset batch. 2nd added.
	// 3rd: 1+1 = 2 > 1.5 -> split (write 2nd). Reset batch. 3rd added.
	// Write 3rd.
	// Expect 3 files.

	err = conv.ConvertImagesToPdf(tempDir, outputPdfSplit, int64(float64(singleSize)*1.5))
	if err != nil {
		t.Fatalf("ConvertImagesToPdf split failed: %v", err)
	}

	// Check for parts
	// part 1: test_output_split_part_1.pdf
	// part 2: test_output_split_part_2.pdf
	// part 3: test_output_split_part_3.pdf

	if _, err := os.Stat("test_output_split_part_1.pdf"); os.IsNotExist(err) {
		t.Errorf("Part 1 not found")
	}
	if _, err := os.Stat("test_output_split_part_2.pdf"); os.IsNotExist(err) {
		t.Errorf("Part 2 not found")
	}
	if _, err := os.Stat("test_output_split_part_3.pdf"); os.IsNotExist(err) {
		t.Errorf("Part 3 not found")
	}
}
