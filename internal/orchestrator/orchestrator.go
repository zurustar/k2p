package orchestrator

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/oumi/k2p/internal/automation"
	"github.com/oumi/k2p/internal/config"
	"github.com/oumi/k2p/internal/filemanager"
	"github.com/oumi/k2p/internal/imageprocessing"
	"github.com/oumi/k2p/internal/pdf"
	"github.com/oumi/k2p/internal/screenshot"
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
	pageCount, screenshots, margins, allMargins, err := o.capturePages(ctx, tempDir, options)
	if err != nil {
		return nil, fmt.Errorf("failed to capture pages: %w", err)
	}

	result.PageCount = pageCount

	if options.Verbose {
		fmt.Printf("\nCaptured %d pages\n", pageCount)
	}

	// Step 9: Handle mode-specific workflow
	if options.Mode == "detect" {
		// Detection mode: analyze margins and report, no PDF generation
		fmt.Println("\n=== Margin Analysis Complete ===")
		fmt.Printf("Analyzed %d pages\n", pageCount)

		// Show per-page margins if verbose
		if options.Verbose && len(allMargins) > 0 {
			fmt.Println("\nPer-page margin details:")
			for i, m := range allMargins {
				fmt.Printf("  Page %3d: Top=%3d Bottom=%3d Left=%3d Right=%3d\n",
					i+1, m.Top, m.Bottom, m.Left, m.Right)
			}
		}

		fmt.Printf("\nMinimum removable margins (safe for all pages):\n")
		fmt.Printf("  Top:    %d pixels\n", margins.Top)
		fmt.Printf("  Bottom: %d pixels\n", margins.Bottom)
		fmt.Printf("  Left:   %d pixels\n", margins.Left)
		fmt.Printf("  Right:  %d pixels\n", margins.Right)
		fmt.Printf("\nTo generate PDF with these margins, run:\n")
		fmt.Printf("  k2p --mode generate --trim-top %d --trim-bottom %d --trim-left %d --trim-right %d\n",
			margins.Top, margins.Bottom, margins.Left, margins.Right)

		result.Duration = time.Since(startTime)
		fmt.Printf("\nDuration: %s\n", result.Duration.Round(time.Second))

		// Play completion sound
		exec.Command("afplay", "/System/Library/Sounds/Glass.aiff").Start()

		return result, nil
	}

	// Step 10: Apply custom trimming to all screenshots (if specified)
	// This is done AFTER capture to avoid interfering with end-of-book detection
	hasCustomTrim := options.Mode == "generate" &&
		(options.TrimTop != 0 || options.TrimBottom != 0 || options.TrimLeft != 0 || options.TrimRight != 0)

	if hasCustomTrim {
		if options.Verbose {
			fmt.Printf("\nApplying custom trimming to %d pages...\n", len(screenshots))
			fmt.Printf("  Trim margins: Top=%d Bottom=%d Left=%d Right=%d\n",
				options.TrimTop, options.TrimBottom, options.TrimLeft, options.TrimRight)
		}

		trimmedScreenshots := make([]string, 0, len(screenshots))
		for i, screenshot := range screenshots {
			trimmedPath := filepath.Join(tempDir, fmt.Sprintf("page_%04d_trimmed.png", i+1))
			if err := o.trimScreenshotWithCustomMargins(screenshot, trimmedPath,
				options.TrimTop, options.TrimBottom, options.TrimLeft, options.TrimRight, false); err != nil {
				if options.Verbose {
					fmt.Printf("  Warning: Failed to trim page %d, using original: %v\n", i+1, err)
				}
				trimmedScreenshots = append(trimmedScreenshots, screenshot)
			} else {
				trimmedScreenshots = append(trimmedScreenshots, trimmedPath)
				// Remove original to save space
				os.Remove(screenshot)
			}
		}
		screenshots = trimmedScreenshots

		if options.Verbose {
			fmt.Printf("✓ Trimming complete\n")
		}
	}

	// Step 11: Generate PDF (generate mode only)
	fmt.Println("\nGenerating PDF...")
	pdfOpts := pdf.GetQualitySettings(options.PDFQuality)
	if err := o.pdfGen.CreatePDF(screenshots, outputPath, pdfOpts); err != nil {
		return nil, fmt.Errorf("failed to generate PDF: %w", err)
	}

	// Step 11: Get file size
	fileInfo, err := os.Stat(outputPath)
	if err == nil {
		result.FileSize = fileInfo.Size()
	}

	result.Duration = time.Since(startTime)

	// Step 12: Wait for macOS to clear screen recording indicator (blue dot)
	// The optimization removed the 2-second wait that previously allowed this
	time.Sleep(1 * time.Second)

	// Step 13: Play completion sound
	// Use macOS system sound to notify user (helpful when Kindle is in foreground)
	exec.Command("afplay", "/System/Library/Sounds/Glass.aiff").Start()

	// Step 14: Display success message
	fmt.Println("\n=== Conversion Complete ===")
	fmt.Printf("Output: %s\n", outputPath)
	fmt.Printf("Pages: %d\n", len(screenshots)) // Show actual PDF page count
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
// Returns: pageCount, screenshot paths, aggregated margins, all page margins, error
func (o *DefaultOrchestrator) capturePages(ctx context.Context, tempDir string, options *config.ConversionOptions) (int, []string, imageprocessing.TrimMargins, []imageprocessing.TrimMargins, error) {
	var screenshots []string
	var allMargins []imageprocessing.TrimMargins
	pageNum := 1
	maxPages := 1000 // Safety limit
	retryConfig := DefaultRetryConfig()

	// Determine if we should apply custom trimming
	// Allow 0 values - user can trim only specific edges
	// Trimming is enabled if any trim value is non-zero
	hasCustomTrim := options.Mode == "generate" &&
		(options.TrimTop != 0 || options.TrimBottom != 0 || options.TrimLeft != 0 || options.TrimRight != 0)

	// Debug: Show trimming configuration
	if options.Verbose {
		fmt.Printf("\n[DEBUG] Custom trimming configuration:\n")
		fmt.Printf("  Mode:       %s\n", options.Mode)
		fmt.Printf("  TrimTop:    %d\n", options.TrimTop)
		fmt.Printf("  TrimBottom: %d\n", options.TrimBottom)
		fmt.Printf("  TrimLeft:   %d\n", options.TrimLeft)
		fmt.Printf("  TrimRight:  %d\n", options.TrimRight)
		fmt.Printf("  hasCustomTrim: %v\n", hasCustomTrim)
	}

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

			// Add detection images to screenshots (no trimming needed since trimming is opt-in now)
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

	// Activate Kindle once before starting page capture
	// This ensures Kindle is in the foreground and waits for Space switching
	fmt.Println("Activating Kindle app...")
	dummyPath := filepath.Join(tempDir, "activation_check.png")
	if err := o.capturer.CaptureFrontmostWindow(dummyPath); err != nil {
		return 0, nil, imageprocessing.TrimMargins{}, nil, fmt.Errorf("failed to activate Kindle: %w", err)
	}
	// Remove the dummy screenshot
	os.Remove(dummyPath)
	fmt.Println("✓ Kindle is active and ready")

	for pageNum <= maxPages {
		// Check context cancellation
		select {
		case <-ctx.Done():
			aggregatedMargins := imageprocessing.AggregateMinimumMargins(allMargins)
			return pageNum - 1, screenshots, aggregatedMargins, allMargins, ctx.Err()
		default:
		}

		// Display progress
		fmt.Printf("\rCapturing page %d...", pageNum)

		// Capture screenshot with retry (without activation - much faster!)
		screenshotPath := filepath.Join(tempDir, fmt.Sprintf("page_%04d.png", pageNum))
		err := RetryWithBackoff(ctx, retryConfig, func() error {
			return o.capturer.CaptureWithoutActivation(screenshotPath)
		})
		if err != nil {
			// CRITICAL: If we can't capture screenshots, the entire conversion is pointless
			aggregatedMargins := imageprocessing.AggregateMinimumMargins(allMargins)
			return pageNum - 1, screenshots, aggregatedMargins, allMargins, fmt.Errorf("failed to capture page %d: %w", pageNum, err)
		}

		// Calculate margins for this page (for detection mode or analysis)
		margins, err := imageprocessing.CalculateTrimMarginsFromFile(screenshotPath)
		if err != nil && options.Verbose {
			fmt.Printf("\nWarning: Failed to calculate margins for page %d: %v\n", pageNum, err)
		}
		allMargins = append(allMargins, margins)

		// Store screenshot path (trimming will be done in batch before PDF generation)
		screenshots = append(screenshots, screenshotPath)

		// Check for end of book (last 5 pages identical)
		if len(screenshots) >= 5 {
			// Show debug info for end detection only in verbose mode
			if options.Verbose {
				fmt.Printf("\n[DEBUG] End detection check: total screenshots = %d\n", len(screenshots))
				fmt.Printf("[DEBUG] Checking last 5 screenshots (indices %d-%d):\n",
					len(screenshots)-5, len(screenshots)-1)
				for i := len(screenshots) - 5; i < len(screenshots); i++ {
					fmt.Printf("[DEBUG]   [%d] %s\n", i, filepath.Base(screenshots[i]))
				}
			}

			allIdentical := true
			for i := len(screenshots) - 4; i < len(screenshots); i++ {
				similarity, err := imageprocessing.CompareImages(screenshots[i-1], screenshots[i])
				if err != nil {
					if options.Verbose {
						fmt.Printf("\n[DEBUG] Warning: Failed to compare screenshots for end detection: %v\n", err)
					}
					allIdentical = false
					break
				}
				if options.Verbose {
					fmt.Printf("[DEBUG] Compare [%d] %s vs [%d] %s: %.2f%% similarity\n",
						i-1, filepath.Base(screenshots[i-1]),
						i, filepath.Base(screenshots[i]),
						similarity*100)
				}
				if similarity < 0.995 { // 99.5% - end-of-book pages are 100% identical
					allIdentical = false
					break
				}
			}

			if allIdentical {
				// Last 5 pages are identical - we've reached the end
				// These are the rating/review screens, not actual book content
				// Remove the last 5 pages from screenshots AND margins
				fmt.Printf("\n\nReached end of book (last 5 pages are identical)\n")
				fmt.Printf("Removing last 5 pages (rating screens) from PDF and margin analysis\n")
				screenshots = screenshots[:len(screenshots)-5]
				// Also remove from margin analysis to prevent gray backgrounds from affecting detection
				if len(allMargins) >= 5 {
					allMargins = allMargins[:len(allMargins)-5]
				}
				break
			} else if options.Verbose {
				fmt.Printf("[DEBUG] Not all identical, continuing...\n")
			}
		}

		// Turn to next page with retry
		err = RetryWithBackoff(ctx, retryConfig, func() error {
			return o.automation.TurnNextPage(direction)
		})
		if err != nil {
			aggregatedMargins := imageprocessing.AggregateMinimumMargins(allMargins)
			return pageNum, screenshots, aggregatedMargins, allMargins, fmt.Errorf("failed to turn page after retries: %w", err)
		}

		// Wait for page delay
		time.Sleep(options.PageDelay)

		pageNum++
	}

	fmt.Println() // New line after progress

	if pageNum > maxPages {
		aggregatedMargins := imageprocessing.AggregateMinimumMargins(allMargins)
		return pageNum - 1, screenshots, aggregatedMargins, allMargins, fmt.Errorf("reached maximum page limit (%d)", maxPages)
	}

	// Aggregate all margins and return
	aggregatedMargins := imageprocessing.AggregateMinimumMargins(allMargins)
	return pageNum, screenshots, aggregatedMargins, allMargins, nil
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
