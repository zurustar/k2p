package imageprocessing

import (
	"fmt"
	"image"
	"image/png"
	"os"
)

// Thresholds for "Black-ish" and "White-ish" pixels
const (
	blackThreshold = 60
	whiteThreshold = 195
)

// TrimBorders removes uniform colored borders (black or white) from an image
// Based on improved implementation from gazounomawarinoiranaifuchiwokesu
func TrimBorders(img image.Image) image.Image {
	bounds := findContentBounds(img)
	if bounds.Empty() {
		// Image is completely black or empty, return original
		return img
	}

	// If bounds match original, no trimming needed
	if bounds == img.Bounds() {
		return img
	}

	// Create trimmed image
	trimmedBounds := image.Rect(0, 0, bounds.Dx(), bounds.Dy())
	trimmed := image.NewRGBA(trimmedBounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			trimmed.Set(x-bounds.Min.X, y-bounds.Min.Y, img.At(x, y))
		}
	}

	return trimmed
}

// findContentBounds finds the content area by removing uniform borders
func findContentBounds(img image.Image) image.Rectangle {
	bounds := img.Bounds()
	minX, minY := bounds.Max.X, bounds.Max.Y
	maxX, maxY := bounds.Min.X, bounds.Min.Y

	// Helpers to check color type
	isPixelBlack := func(r8, g8, b8 uint32) bool {
		return r8 <= blackThreshold && g8 <= blackThreshold && b8 <= blackThreshold
	}
	isPixelWhite := func(r8, g8, b8 uint32) bool {
		return r8 >= whiteThreshold && g8 >= whiteThreshold && b8 >= whiteThreshold
	}

	// 1. Determine the target background color (Black or White) based on corners
	corners := []struct{ x, y int }{
		{bounds.Min.X, bounds.Min.Y},
		{bounds.Max.X - 1, bounds.Min.Y},
		{bounds.Min.X, bounds.Max.Y - 1},
		{bounds.Max.X - 1, bounds.Max.Y - 1},
	}

	blackCornerCount := 0
	whiteCornerCount := 0

	for _, p := range corners {
		c := img.At(p.x, p.y)
		r, g, b, _ := c.RGBA()
		r8, g8, b8 := r>>8, g>>8, b>>8

		if isPixelBlack(r8, g8, b8) {
			blackCornerCount++
		} else if isPixelWhite(r8, g8, b8) {
			whiteCornerCount++
		}
	}

	type TargetMode int
	const (
		ModeNone TargetMode = iota
		ModeBlack
		ModeWhite
	)

	var mode TargetMode
	if blackCornerCount > whiteCornerCount {
		mode = ModeBlack
	} else if whiteCornerCount > blackCornerCount {
		mode = ModeWhite
	} else {
		// Tie or neither
		if blackCornerCount > 0 {
			mode = ModeBlack
		} else if whiteCornerCount > 0 {
			mode = ModeWhite
		} else {
			// No detectable background color at corners
			mode = ModeNone
		}
	}

	if mode == ModeNone {
		// No detectable background color, return original bounds
		return bounds
	}

	// Helpers to check row/col uniformity
	// A row is removable if it is MOSTLY (>95%) the Target Color
	const noiseTolerance = 0.95
	const lookaheadGap = 5 // Skip over thin noise lines

	isRowRemovable := func(y int) bool {
		width := bounds.Dx()
		matchCount := 0

		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c := img.At(x, y)
			r, g, b, _ := c.RGBA()
			r8, g8, b8 := r>>8, g>>8, b>>8

			if mode == ModeBlack && isPixelBlack(r8, g8, b8) {
				matchCount++
			} else if mode == ModeWhite && isPixelWhite(r8, g8, b8) {
				matchCount++
			}
		}

		total := float64(width)
		return float64(matchCount)/total >= noiseTolerance
	}

	isColRemovable := func(x int) bool {
		height := bounds.Dy()
		matchCount := 0

		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			c := img.At(x, y)
			r, g, b, _ := c.RGBA()
			r8, g8, b8 := r>>8, g>>8, b>>8

			if mode == ModeBlack && isPixelBlack(r8, g8, b8) {
				matchCount++
			} else if mode == ModeWhite && isPixelWhite(r8, g8, b8) {
				matchCount++
			}
		}

		total := float64(height)
		return float64(matchCount)/total >= noiseTolerance
	}

	// Scan MinY (Top)
	minY = bounds.Min.Y
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		if isRowRemovable(y) {
			minY = y + 1
			continue
		}
		// Lookahead
		allNextRemovable := true
		if y+lookaheadGap >= bounds.Max.Y {
			allNextRemovable = false
		} else {
			for k := 1; k <= lookaheadGap; k++ {
				if !isRowRemovable(y + k) {
					allNextRemovable = false
					break
				}
			}
		}
		if allNextRemovable {
			minY = y + 1
		} else {
			break
		}
	}

	// If whole image is removable, return empty
	if minY >= bounds.Max.Y {
		return image.Rectangle{}
	}

	// Scan MaxY (Bottom)
	maxY = bounds.Max.Y
	for y := bounds.Max.Y - 1; y >= minY; y-- {
		if isRowRemovable(y) {
			maxY = y
			continue
		}
		// Lookahead (Upwards)
		allPriorRemovable := true
		if y-lookaheadGap < minY {
			allPriorRemovable = false
		} else {
			for k := 1; k <= lookaheadGap; k++ {
				if !isRowRemovable(y - k) {
					allPriorRemovable = false
					break
				}
			}
		}
		if allPriorRemovable {
			maxY = y
		} else {
			break
		}
	}

	// Scan MinX (Left)
	minX = bounds.Min.X
	for x := bounds.Min.X; x < bounds.Max.X; x++ {
		if isColRemovable(x) {
			minX = x + 1
			continue
		}
		// Lookahead
		allNextRemovable := true
		if x+lookaheadGap >= bounds.Max.X {
			allNextRemovable = false
		} else {
			for k := 1; k <= lookaheadGap; k++ {
				if !isColRemovable(x + k) {
					allNextRemovable = false
					break
				}
			}
		}
		if allNextRemovable {
			minX = x + 1
		} else {
			break
		}
	}

	// Scan MaxX (Right)
	maxX = bounds.Max.X
	for x := bounds.Max.X - 1; x >= minX; x-- {
		if isColRemovable(x) {
			maxX = x
			continue
		}
		// Lookahead (Leftwards)
		allPriorRemovable := true
		if x-lookaheadGap < minX {
			allPriorRemovable = false
		} else {
			for k := 1; k <= lookaheadGap; k++ {
				if !isColRemovable(x - k) {
					allPriorRemovable = false
					break
				}
			}
		}
		if allPriorRemovable {
			maxX = x
		} else {
			break
		}
	}

	return image.Rect(minX, minY, maxX, maxY)
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

// TrimMargins represents the removable margins on each edge of an image
type TrimMargins struct {
	Top    int
	Bottom int
	Left   int
	Right  int
}

// CalculateTrimMargins analyzes an image and returns the removable margin size for each edge
func CalculateTrimMargins(img image.Image) TrimMargins {
	bounds := findContentBounds(img)
	originalBounds := img.Bounds()

	return TrimMargins{
		Top:    bounds.Min.Y - originalBounds.Min.Y,
		Bottom: originalBounds.Max.Y - bounds.Max.Y,
		Left:   bounds.Min.X - originalBounds.Min.X,
		Right:  originalBounds.Max.X - bounds.Max.X,
	}
}

// CalculateTrimMarginsFromFile analyzes an image file and returns the removable margins
func CalculateTrimMarginsFromFile(imagePath string) (TrimMargins, error) {
	file, err := os.Open(imagePath)
	if err != nil {
		return TrimMargins{}, fmt.Errorf("failed to open image file: %w", err)
	}
	defer file.Close()

	img, err := png.Decode(file)
	if err != nil {
		return TrimMargins{}, fmt.Errorf("failed to decode image: %w", err)
	}

	return CalculateTrimMargins(img), nil
}

// AggregateMinimumMargins takes margins from multiple pages and returns the minimum safe values
// The minimum values ensure we don't accidentally trim actual book content
func AggregateMinimumMargins(margins []TrimMargins) TrimMargins {
	if len(margins) == 0 {
		return TrimMargins{}
	}

	// Start with the first margin as the minimum
	minMargins := margins[0]

	// Find the minimum for each edge
	for i := 1; i < len(margins); i++ {
		if margins[i].Top < minMargins.Top {
			minMargins.Top = margins[i].Top
		}
		if margins[i].Bottom < minMargins.Bottom {
			minMargins.Bottom = margins[i].Bottom
		}
		if margins[i].Left < minMargins.Left {
			minMargins.Left = margins[i].Left
		}
		if margins[i].Right < minMargins.Right {
			minMargins.Right = margins[i].Right
		}
	}

	return minMargins
}

// TrimWithCustomMargins trims an image using specific pixel margins for each edge
func TrimWithCustomMargins(img image.Image, top, bottom, left, right int) image.Image {
	bounds := img.Bounds()

	// Calculate new bounds
	newMinX := bounds.Min.X + left
	newMinY := bounds.Min.Y + top
	newMaxX := bounds.Max.X - right
	newMaxY := bounds.Max.Y - bottom

	// Validate bounds
	if newMinX >= newMaxX || newMinY >= newMaxY {
		// Invalid trim would result in empty or negative image, return original
		return img
	}

	// Create trimmed image
	trimmedBounds := image.Rect(0, 0, newMaxX-newMinX, newMaxY-newMinY)
	trimmed := image.NewRGBA(trimmedBounds)

	for y := newMinY; y < newMaxY; y++ {
		for x := newMinX; x < newMaxX; x++ {
			trimmed.Set(x-newMinX, y-newMinY, img.At(x, y))
		}
	}

	return trimmed
}

// TrimImageFileWithCustomMargins trims an image file using custom margins and saves the result
func TrimImageFileWithCustomMargins(inputPath, outputPath string, top, bottom, left, right int) error {
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

	// Trim with custom margins
	trimmed := TrimWithCustomMargins(img, top, bottom, left, right)

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
