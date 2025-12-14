package converter

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/go-pdf/fpdf"
)

type PdfConverter struct{}

func NewPdfConverter() *PdfConverter {
	return &PdfConverter{}
}

// ParseSize parses a size string like "1.8MB" into bytes.
func ParseSize(sizeStr string) (int64, error) {
	sizeStr = strings.ToUpper(strings.TrimSpace(sizeStr))
	units := map[string]int64{
		"KB": 1024,
		"MB": 1024 * 1024,
		"GB": 1024 * 1024 * 1024,
	}

	for unit, factor := range units {
		if strings.HasSuffix(sizeStr, unit) {
			valStr := strings.TrimSuffix(sizeStr, unit)
			val, err := strconv.ParseFloat(valStr, 64)
			if err != nil {
				return 0, err
			}
			return int64(val * float64(factor)), nil
		}
	}

	// Try simple int
	val, err := strconv.ParseInt(sizeStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid size format: %s", sizeStr)
	}
	return val, nil
}

func (c *PdfConverter) ConvertImagesToPdf(imageDir, outputFilename string, maxSize int64) error {
	fmt.Printf("Converting images in %s to %s...\n", imageDir, outputFilename)

	entries, err := os.ReadDir(imageDir)
	if err != nil {
		return err
	}

	var images []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".png") {
			images = append(images, filepath.Join(imageDir, entry.Name()))
		}
	}

	sort.Strings(images)

	if len(images) == 0 {
		fmt.Println("No images found to convert.")
		return nil
	}

	if maxSize <= 0 {
		return c.writeBatch(images, outputFilename)
	}

	var currentBatch []string
	var currentSize int64
	partNum := 1

	baseName := strings.TrimSuffix(outputFilename, filepath.Ext(outputFilename))
	ext := filepath.Ext(outputFilename)
	if ext == "" {
		ext = ".pdf"
	}

	for _, imgPath := range images {
		info, err := os.Stat(imgPath)
		if err != nil {
			return err
		}
		imgSize := info.Size()

		if len(currentBatch) > 0 && currentSize+imgSize > maxSize {
			partName := fmt.Sprintf("%s_part_%d%s", baseName, partNum, ext)
			if err := c.writeBatch(currentBatch, partName); err != nil {
				return err
			}
			partNum++
			currentBatch = []string{}
			currentSize = 0
		}

		currentBatch = append(currentBatch, imgPath)
		currentSize += imgSize
	}

	if len(currentBatch) > 0 {
		partName := outputFilename
		if partNum > 1 {
			partName = fmt.Sprintf("%s_part_%d%s", baseName, partNum, ext)
		}
		if err := c.writeBatch(currentBatch, partName); err != nil {
			return err
		}
	}

	return nil
}

func (c *PdfConverter) writeBatch(images []string, outputFilename string) error {
	if len(images) == 0 {
		return nil
	}

	// Initialize PDF with "pt" unit to simplify pixel-to-point mapping.
	// We'll set the page size for each image individually.
	pdf := fpdf.New("P", "pt", "Letter", "")

	for _, imgPath := range images {
        opt := fpdf.ImageOptions{
			ReadDpi: true,
		}
        info := pdf.RegisterImageOptions(imgPath, opt)
        if info == nil {
            fmt.Printf("Warning: failed to load image %s\n", imgPath)
            continue
        }

        // info.Width and Height are float64.
        w := info.Width()
        h := info.Height()

        // Add a page with the size of the image
        pdf.AddPageFormat("P", fpdf.SizeType{Wd: w, Ht: h})

        // Place image at 0,0 with full width and height
        pdf.ImageOptions(imgPath, 0, 0, w, h, false, opt, 0, "")
    }

	err := pdf.OutputFileAndClose(outputFilename)
	if err != nil {
		return fmt.Errorf("error saving PDF %s: %v", outputFilename, err)
	}
	fmt.Printf("Successfully created %s\n", outputFilename)
	return nil
}
