package orchestrator

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/oumi/k2p/pkg/automation"
	"github.com/oumi/k2p/pkg/config"
	"github.com/oumi/k2p/pkg/filemanager"
	"github.com/oumi/k2p/pkg/imageprocessing"
	"github.com/oumi/k2p/pkg/pdf"
	"github.com/oumi/k2p/pkg/screenshot"
)

// ConversionResult contains the result of a conversion
type ConversionResult struct {
	// Path to generated PDF
	OutputPath string

	// Number of pages captured
	PageCount int

	// Total conversion duration
	Duration time.Duration

	// Output file size in bytes
	FileSize int64

	// Any warnings encountered
	Warnings []string
}

// ConversionOrchestrator coordinates the entire conversion workflow
type ConversionOrchestrator interface {
	// ConvertCurrentBook converts the currently open book to PDF
	ConvertCurrentBook(ctx context.Context, options *config.ConversionOptions) (*ConversionResult, error)
}

// DefaultOrchestrator is the default implementation
type DefaultOrchestrator struct {
	automation  automation.KindleAutomation
	fileManager filemanager.FileManager
	pdfGen      pdf.PDFGenerator
	capturer    screenshot.Capturer
}

// NewOrchestrator creates a new conversion orchestrator
func NewOrchestrator() ConversionOrchestrator {
	return &DefaultOrchestrator{
		automation:  automation.NewKindleAutomation(),
		fileManager: filemanager.NewFileManager(),
		pdfGen:      pdf.NewPDFGenerator(),
		capturer:    screenshot.NewCapturer(),
	}
}

// ConvertCurrentBook implements the main conversion workflow
func (o *DefaultOrchestrator) ConvertCurrentBook(ctx context.Context, options *config.ConversionOptions) (*ConversionResult, error) {
	startTime := time.Now()
	result := &ConversionResult{
		Warnings: []string{},
	}

	// Step 1: Display preparation instructions
	fmt.Println("=== Kindle to PDF Converter ===")
	fmt.Println("\nPlease ensure:")
	fmt.Println("  1. Kindle app is running")
	fmt.Println("  2. A book is open in Kindle")
	fmt.Println("  3. Kindle app is in the foreground")
	fmt.Println()

	// Step 2: Wait for user confirmation
	if !options.AutoConfirm {
		fmt.Print("Press Enter when ready to begin conversion...")
		fmt.Scanln()
	}

	// Step 3: Apply startup delay with countdown
	if options.StartupDelay > 0 {
		if options.ShowCountdown {
			o.showCountdown(options.StartupDelay)
		} else {
			time.Sleep(options.StartupDelay)
		}
	}

	// Step 4: Validate Kindle app state
	if err := o.validateKindleState(options.Verbose); err != nil {
		return nil, err
	}

	// Step 5: Check disk space
	estimatedSize := int64(100 * 1024 * 1024) // Estimate 100MB for safety
	outputDir := options.OutputDir
	if outputDir == "" {
		var err error
		outputDir, err = os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get current directory: %w", err)
		}
	}

	if err := o.fileManager.CheckDiskSpace(outputDir, estimatedSize); err != nil {
		return nil, err
	}

	// Step 6: Resolve output path
	outputPath, err := o.fileManager.ResolveOutputPath(outputDir)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve output path: %w", err)
	}

	// Check if file exists
	proceed, err := o.fileManager.HandleExistingFile(outputPath, options.AutoConfirm)
	if err != nil {
		return nil, err
	}
	if !proceed {
		return nil, fmt.Errorf("conversion cancelled: file already exists")
	}

	result.OutputPath = outputPath

	// Step 7: Create temporary directory
	tempDir, err := o.fileManager.CreateTempDir()
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary directory: %w", err)
	}
	defer o.fileManager.CleanupTempDir(tempDir)

	// Step 8: Page capture loop
	pageCount, screenshots, err := o.capturePages(ctx, tempDir, options)
	if err != nil {
		return nil, fmt.Errorf("failed to capture pages: %w", err)
	}

	result.PageCount = pageCount

	if options.Verbose {
		fmt.Printf("\nCaptured %d pages\n", pageCount)
	}

	// Step 9: Generate PDF
	fmt.Println("\nGenerating PDF...")
	pdfOpts := pdf.GetQualitySettings(options.PDFQuality)
	if err := o.pdfGen.CreatePDF(screenshots, outputPath, pdfOpts); err != nil {
		return nil, fmt.Errorf("failed to generate PDF: %w", err)
	}

	// Step 10: Get file size
	fileInfo, err := os.Stat(outputPath)
	if err == nil {
		result.FileSize = fileInfo.Size()
	}

	result.Duration = time.Since(startTime)

	// Step 11: Display success message
	fmt.Println("\n=== Conversion Complete ===")
	fmt.Printf("Output: %s\n", outputPath)
	fmt.Printf("Pages: %d\n", pageCount)
	fmt.Printf("Size: %.2f MB\n", float64(result.FileSize)/(1024*1024))
	fmt.Printf("Duration: %s\n", result.Duration.Round(time.Second))

	return result, nil
}

// validateKindleState validates that Kindle is ready for conversion
func (o *DefaultOrchestrator) validateKindleState(verbose bool) error {
	if verbose {
		fmt.Println("Checking Kindle app state...")
	}

	// Check if Kindle is installed
	installed, err := o.automation.IsKindleInstalled()
	if err != nil {
		return fmt.Errorf("failed to check Kindle installation: %w", err)
	}
	if !installed {
		return fmt.Errorf("Kindle app is not installed. Please install from the Mac App Store")
	}

	// Check if book is open
	bookOpen, err := o.automation.IsBookOpen()
	if err != nil {
		return fmt.Errorf("failed to check if book is open: %w", err)
	}
	if !bookOpen {
		return fmt.Errorf("no book is currently open in Kindle app. Please open a book and try again")
	}

	// Check if Kindle is in foreground
	inForeground, err := o.automation.IsKindleInForeground()
	if err != nil {
		return fmt.Errorf("failed to check if Kindle is in foreground: %w", err)
	}
	if !inForeground {
		return fmt.Errorf("Kindle app is not in foreground. Please bring Kindle to the front and try again")
	}

	if verbose {
		fmt.Println("✓ Kindle app is ready")
	}

	return nil
}

// capturePages captures all pages from the current book
func (o *DefaultOrchestrator) capturePages(ctx context.Context, tempDir string, options *config.ConversionOptions) (int, []string, error) {
	var screenshots []string
	pageNum := 1
	maxPages := 1000 // Safety limit
	retryConfig := DefaultRetryConfig()

	// Auto-detect page turn direction (unless explicitly set to "left")
	direction := options.PageTurnKey
	if direction != "left" {
		// Try to auto-detect
		if options.Verbose {
			fmt.Println("\nAuto-detecting page turn direction...")
		}

		detectedDirection, detectionImages, err := o.detectPageTurnDirection(ctx, tempDir, retryConfig, options)
		if err == nil && detectedDirection != "" {
			direction = detectedDirection
			// Add detection images to screenshots (they are the first pages of the book)
			screenshots = append(screenshots, detectionImages...)
		} else {
			direction = "right" // fallback to default
			if options.Verbose {
				fmt.Println("Using default direction: right")
			}
		}
	} else if options.Verbose {
		fmt.Println("\nUsing configured direction: left")
	}

	fmt.Println("\nCapturing pages...")

	for pageNum <= maxPages {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return pageNum - 1, screenshots, ctx.Err()
		default:
		}

		// Display progress
		fmt.Printf("\rCapturing page %d...", pageNum)

		// Capture screenshot with retry (captures frontmost window)
		screenshotPath := filepath.Join(tempDir, fmt.Sprintf("page_%04d.png", pageNum))
		err := RetryWithBackoff(ctx, retryConfig, func() error {
			return o.capturer.CaptureFrontmostWindow(screenshotPath)
		})
		if err != nil {
			// CRITICAL: If we can't capture screenshots, the entire conversion is pointless
			return pageNum - 1, screenshots, fmt.Errorf("failed to capture page %d: %w", pageNum, err)
		}

		// Trim borders if enabled
		if options.TrimBorders {
			trimmedPath := filepath.Join(tempDir, fmt.Sprintf("page_%04d_trimmed.png", pageNum))
			if err := o.trimScreenshot(screenshotPath, trimmedPath, options.Verbose); err != nil {
				if options.Verbose {
					fmt.Printf("\nWarning: Failed to trim page %d: %v\n", pageNum, err)
				}
				// Use original screenshot if trimming fails
				screenshots = append(screenshots, screenshotPath)
			} else {
				// Use trimmed screenshot
				screenshots = append(screenshots, trimmedPath)
				// Remove original to save space
				os.Remove(screenshotPath)
			}
		} else {
			screenshots = append(screenshots, screenshotPath)
		}

		// Check if there are more pages by comparing with previous screenshots
		// If the last 5 pages are all identical, we've reached the end
		// (This handles books with blank pages at the end)
		if len(screenshots) >= 5 {
			allIdentical := true
			for i := len(screenshots) - 4; i < len(screenshots); i++ {
				similarity, err := imageprocessing.CompareImages(screenshots[i-1], screenshots[i])
				if err != nil && options.Verbose {
					fmt.Printf("\nWarning: Failed to compare screenshots for end detection: %v\n", err)
					allIdentical = false
					break
				}
				if similarity < 0.95 {
					allIdentical = false
					break
				}
			}

			if allIdentical {
				// Last 5 pages are identical - we've reached the end
				// These are the rating/review screens, not actual book content
				// Remove the last 5 pages from screenshots
				if options.Verbose {
					fmt.Printf("\n\nReached end of book (last 5 pages are identical)\n")
					fmt.Printf("Removing last 5 pages (rating screens) from PDF\n")
				}
				screenshots = screenshots[:len(screenshots)-5]
				break
			}
		}

		// Turn to next page with retry
		err = RetryWithBackoff(ctx, retryConfig, func() error {
			return o.automation.TurnNextPage(direction)
		})
		if err != nil {
			return pageNum, screenshots, fmt.Errorf("failed to turn page after retries: %w", err)
		}

		// Wait for page delay
		time.Sleep(options.PageDelay)

		pageNum++
	}

	fmt.Println() // New line after progress

	if pageNum > maxPages {
		return pageNum - 1, screenshots, fmt.Errorf("reached maximum page limit (%d)", maxPages)
	}

	return pageNum, screenshots, nil
}

// showCountdown displays a countdown timer
func (o *DefaultOrchestrator) showCountdown(duration time.Duration) {
	fmt.Printf("Starting in ")
	seconds := int(duration.Seconds())
	for i := seconds; i > 0; i-- {
		fmt.Printf("%d...", i)
		time.Sleep(time.Second)
	}
	fmt.Println("Go!")
}
