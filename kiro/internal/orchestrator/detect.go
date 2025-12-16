package orchestrator

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/oumi/k2p/pkg/config"
	"github.com/oumi/k2p/pkg/imageprocessing"
)

// detectPageTurnDirection tries to auto-detect the correct page turn direction
// by capturing multiple pages and checking if content changes
// Returns: direction string, captured image paths, error
func (o *DefaultOrchestrator) detectPageTurnDirection(ctx context.Context, tempDir string, retryConfig RetryConfig, options *config.ConversionOptions) (string, []string, error) {
	if options.Verbose {
		fmt.Println("Auto-detecting page turn direction...")
	}

	// Create debug directory in project
	debugDir := filepath.Join("debug_samples")
	os.MkdirAll(debugDir, 0755)
	if options.Verbose {
		fmt.Printf("  DEBUG: Screenshots will be saved to: %s\n\n", debugDir)
	}

	// Step 1: Capture cover page (activate Kindle once)
	coverPath := filepath.Join(tempDir, "detect_cover.png")
	coverDebugPath := filepath.Join(debugDir, "detect_cover.png")
	if options.Verbose {
		fmt.Println("  [Cover] Activating Kindle and capturing cover page...")
	}
	err := RetryWithBackoff(ctx, retryConfig, func() error {
		return o.capturer.CaptureFrontmostWindow(coverPath)
	})
	if err != nil {
		return "", nil, fmt.Errorf("failed to capture cover: %w", err)
	}
	// Copy to debug directory
	exec.Command("cp", coverPath, coverDebugPath).Run()
	if options.Verbose {
		fmt.Printf("  [Cover] Saved: %s\n", coverDebugPath)
		fmt.Println("  [Cover] Kindle is now active, using fast capture for detection...")
	}

	// Step 2: Test RIGHT arrow - press 3 times
	if options.Verbose {
		fmt.Println("\n  Testing RIGHT arrow (3 presses)...")
	}
	rightPaths := []string{coverPath} // Start with cover
	for i := 1; i <= 3; i++ {
		// Press right arrow
		if options.Verbose {
			fmt.Printf("  [Right %d] Pressing RIGHT arrow...\n", i)
		}
		err = RetryWithBackoff(ctx, retryConfig, func() error {
			return o.automation.TurnNextPage("right")
		})
		if err != nil {
			return "", nil, fmt.Errorf("failed to press right arrow: %w", err)
		}

		// Wait for page to load (use configured PageDelay)
		time.Sleep(options.PageDelay)

		// Capture screenshot (fast - no activation)
		rightPath := filepath.Join(tempDir, fmt.Sprintf("detect_right_%d.png", i))
		rightDebugPath := filepath.Join(debugDir, fmt.Sprintf("detect_right_%d.png", i))
		if options.Verbose {
			fmt.Printf("  [Right %d] Capturing screenshot...\n", i)
		}
		err = RetryWithBackoff(ctx, retryConfig, func() error {
			return o.capturer.CaptureWithoutActivation(rightPath)
		})
		if err != nil {
			return "", nil, fmt.Errorf("failed to capture right %d: %w", i, err)
		}
		// Copy to debug directory
		exec.Command("cp", rightPath, rightDebugPath).Run()
		if options.Verbose {
			fmt.Printf("  [Right %d] Saved: %s\n", i, rightDebugPath)
		}
		rightPaths = append(rightPaths, rightPath)
	}

	// Check if RIGHT changed pages
	rightChanged := false
	if options.Verbose {
		fmt.Println("\n  Checking if RIGHT arrow changed pages...")
	}
	for i := 1; i < len(rightPaths); i++ {
		similarity, err := imageprocessing.CompareImages(rightPaths[i-1], rightPaths[i])
		if err != nil && options.Verbose {
			fmt.Printf("  Warning: Failed to compare images: %v\n", err)
		}
		if options.Verbose {
			fmt.Printf("  Compare %s vs %s: %.2f%% similarity\n",
				filepath.Base(rightPaths[i-1]),
				filepath.Base(rightPaths[i]),
				similarity*100)
		}
		if similarity < 0.90 {
			rightChanged = true
			if options.Verbose {
				fmt.Println("  → Pages CHANGED!")
			}
			break
		}
	}

	if rightChanged {
		if options.Verbose {
			fmt.Println("\n✓ Direction detected: RIGHT arrow")
			fmt.Println("  Continuing from current page...")
		}
		// Return right direction and the captured images
		return "right", rightPaths, nil
	}

	// Step 3: RIGHT didn't work, test LEFT arrow
	if options.Verbose {
		fmt.Println("\n  RIGHT arrow didn't change pages.")
		fmt.Println("  Testing LEFT arrow (3 presses)...")
	}

	// We're currently at the same position (cover), so start from there
	leftPaths := []string{rightPaths[len(rightPaths)-1]} // Last right capture (should be cover)
	for i := 1; i <= 3; i++ {
		// Press left arrow
		if options.Verbose {
			fmt.Printf("  [Left %d] Pressing LEFT arrow...\n", i)
		}
		err = RetryWithBackoff(ctx, retryConfig, func() error {
			return o.automation.TurnNextPage("left")
		})
		if err != nil {
			return "", nil, fmt.Errorf("failed to press left arrow: %w", err)
		}

		// Wait for page to load (use configured PageDelay)
		time.Sleep(options.PageDelay)

		// Capture screenshot (fast - no activation)
		leftPath := filepath.Join(tempDir, fmt.Sprintf("detect_left_%d.png", i))
		leftDebugPath := filepath.Join(debugDir, fmt.Sprintf("detect_left_%d.png", i))
		if options.Verbose {
			fmt.Printf("  [Left %d] Capturing screenshot...\n", i)
		}
		err = RetryWithBackoff(ctx, retryConfig, func() error {
			return o.capturer.CaptureWithoutActivation(leftPath)
		})
		if err != nil {
			return "", nil, fmt.Errorf("failed to capture left %d: %w", i, err)
		}
		// Copy to debug directory
		exec.Command("cp", leftPath, leftDebugPath).Run()
		if options.Verbose {
			fmt.Printf("  [Left %d] Saved: %s\n", i, leftDebugPath)
		}
		leftPaths = append(leftPaths, leftPath)
	}

	// Check if LEFT changed pages
	leftChanged := false
	if options.Verbose {
		fmt.Println("\n  Checking if LEFT arrow changed pages...")
	}
	for i := 1; i < len(leftPaths); i++ {
		similarity, err := imageprocessing.CompareImages(leftPaths[i-1], leftPaths[i])
		if err != nil && options.Verbose {
			fmt.Printf("  Warning: Failed to compare images: %v\n", err)
		}
		if options.Verbose {
			fmt.Printf("  Compare %s vs %s: %.2f%% similarity\n",
				filepath.Base(leftPaths[i-1]),
				filepath.Base(leftPaths[i]),
				similarity*100)
		}
		if similarity < 0.90 {
			leftChanged = true
			if options.Verbose {
				fmt.Println("  → Pages CHANGED!")
			}
			break
		}
	}

	if leftChanged {
		if options.Verbose {
			fmt.Println("\n✓ Direction detected: LEFT arrow")
			fmt.Println("  Continuing from current page...")
		}
		// Return left direction and the captured images
		return "left", leftPaths, nil
	}

	// Neither direction worked - ERROR
	return "", nil, fmt.Errorf("could not detect page turn direction: neither RIGHT nor LEFT arrow changed pages")
}
