package converter

import (
	"fmt"
	"image"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-pdf/fpdf"
	"github.com/zurustar/k2p/internal/config"
)

// Convert converts images in tempDir to PDF(s)
func Convert(cfg config.Config) error {
	fmt.Printf("Converting images in %s to %s...\n", cfg.TempDir, cfg.Output)

	files, err := filepath.Glob(filepath.Join(cfg.TempDir, "*.png"))
	if err != nil {
		return err
	}
	if len(files) == 0 {
		return fmt.Errorf("no images found to convert")
	}

	// Simple sort is usually fine if files are named page_0001.png etc.
	// Glob return order is not guaranteed, so we should ensure they are sorted?
	// filepath.Glob usually returns sorted by name, but let's trust naming convention works with default glob sort (often filesystem dependent).
	// Ideally we should sort explicitly, but Glob results are usually sorted on Mac/Linux.
	// For robustness, let's rely on the fact that we named them 0001, 0002...

	if cfg.MaxSize == 0 {
		return createPDF(files, cfg.Output)
	}

	// Split by size
	var currentBatch []string
	var currentSize int64
	partNum := 1

	baseExt := filepath.Ext(cfg.Output)
	baseName := strings.TrimSuffix(cfg.Output, baseExt)

	for _, file := range files {
		info, err := os.Stat(file)
		if err != nil {
			return err
		}
		size := info.Size()

		if len(currentBatch) > 0 && currentSize+size > cfg.MaxSize {
			outputName := fmt.Sprintf("%s_part_%d%s", baseName, partNum, baseExt)
			if err := createPDF(currentBatch, outputName); err != nil {
				return err
			}
			partNum++
			currentBatch = []string{}
			currentSize = 0
		}

		currentBatch = append(currentBatch, file)
		currentSize += size
	}

	if len(currentBatch) > 0 {
		outputName := fmt.Sprintf("%s_part_%d%s", baseName, partNum, baseExt)
		if partNum == 1 {
			outputName = cfg.Output
		}
		if err := createPDF(currentBatch, outputName); err != nil {
			return err
		}
	}

	return nil
}

func createPDF(images []string, output string) error {
	// A4 size: 210 x 297 mm
	// We want to fit images into the page.
	// Or we can set page size to image size.
	// The original tool uses img2pdf which makes PDF page size match image size.
	// We should try to do the same or fit to a standard size.
	// Let's make the PDF page size match the first image size for simplicity, or A4.
	// Original behavior "img2pdf" makes page size = image size.

	// Helper to get image dimensions
	if len(images) == 0 {
		return nil
	}

	// Read first image to determine orientation/size?
	// fpdf defaults to A4.
	// Let's create a new PDF.
	pdf := fpdf.New("P", "mm", "A4", "")

	// We want to change page size dynamically per image to match image size?
	// fpdf allows adding pages with custom format.

	for _, imgPath := range images {
		// Read image config
		f, err := os.Open(imgPath)
		if err != nil {
			return err
		}
		imgCfg, _, err := image.DecodeConfig(f)
		f.Close()
		if err != nil {
			return err
		}

		// Convert pixels to mm (approximate 72 or 96 DPI?)
		// fpdf uses 72 DPI by default for Unit "pt", but we used "mm".
		// 1 inch = 25.4 mm.
		// width_mm = width_px / dpi * 25.4
		// Standard screen DPI is often assumed 72 or 96.
		// fpdf ImageOptions uses internal logic.

		// Detailed approach:
		// pdf.AddPageFormat("P", fpdf.SizeType{Wd: w, Ht: h})
		// We need to convert px to mm.
		// Let's assume 72 DPI for PDF standard.
		w := float64(imgCfg.Width) * 25.4 / 72.0
		h := float64(imgCfg.Height) * 25.4 / 72.0

		pdf.AddPageFormat("P", fpdf.SizeType{Wd: w, Ht: h})

		// Image(src, x, y, w, h, flow, link, linkStr, link)
		pdf.Image(imgPath, 0, 0, w, h, false, "", 0, "")
	}

	fmt.Printf("Successfully created %s\n", output)
	return pdf.OutputFileAndClose(output)
}
