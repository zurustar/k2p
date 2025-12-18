package capturer

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/rriifftt/kindle-to-pdf-go/internal/platform"
)

type Capturer struct {
	platform platform.Platform
	dir      string
	pages    int // 0 means infinite
}

func NewCapturer(p platform.Platform) *Capturer {
	return &Capturer{
		platform: p,
	}
}

func (c *Capturer) WaitForFocus(seconds int) {
	for i := seconds; i > 0; i-- {
		fmt.Printf("Starting in %d seconds... Please focus the Kindle window!\n", i)
		time.Sleep(1 * time.Second)
	}
	fmt.Println("Started!")
}

// CleanDir removes all files in the directory.
func (c *Capturer) CleanDir(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return os.MkdirAll(dir, 0755)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		path := filepath.Join(dir, entry.Name())
		if err := os.Remove(path); err != nil {
			fmt.Printf("Failed to delete %s: %v\n", path, err)
		}
	}
	return nil
}

func (c *Capturer) CaptureLoop(ctx context.Context, outputDir string, pages int, direction string) error {
	if err := c.CleanDir(outputDir); err != nil {
		return fmt.Errorf("failed to clean temp dir: %w", err)
	}

	pageNum := 1
	fmt.Println("Press Ctrl+C to stop capturing early.")

	var currentDirection string
	if direction != "auto" {
		currentDirection = direction
	}

	for {
		select {
		case <-ctx.Done():
			fmt.Println("\nCapture stopped by user.")
			return nil
		default:
			// Continue
		}

		if pages > 0 && pageNum > pages {
			fmt.Printf("Reached limit of %d pages.\n", pages)
			break
		}

		filename := filepath.Join(outputDir, fmt.Sprintf("page_%04d.png", pageNum))
		fmt.Printf("Capturing page %d...\n", pageNum)

		if err := c.platform.Screenshot(filename); err != nil {
			return fmt.Errorf("failed to take screenshot: %w", err)
		}

		// End of Book Detection: Compare with previous page
		if pageNum > 1 {
			prevFilename := filepath.Join(outputDir, fmt.Sprintf("page_%04d.png", pageNum-1))
			identical, err := ImagesAreIdentical(filename, prevFilename)
			if err != nil {
				fmt.Printf("Warning: failed to compare images: %v\n", err)
			} else if identical {
				fmt.Println("Page content is identical to previous page. Assuming end of book.")
				// Remove the duplicate last page
				if err := os.Remove(filename); err != nil {
					fmt.Printf("Warning: failed to remove duplicate page: %v\n", err)
				}
				break
			}
		}

		// Automatic Direction Detection
		if currentDirection == "" && pageNum == 1 {
			// We need to determine direction.
			// Try Right first (LTR assumption)
			fmt.Println("Auto-detecting direction: trying 'right' (LTR)...")
			if err := c.platform.PressKey("right"); err != nil {
				return fmt.Errorf("failed to turn page: %w", err)
			}
			c.wait(ctx)

			// Capture provisional page 2
			testFilename := filepath.Join(outputDir, fmt.Sprintf("page_%04d.png", pageNum+1))
			if err := c.platform.Screenshot(testFilename); err != nil {
				return fmt.Errorf("failed to take screenshot: %w", err)
			}

			identical, err := ImagesAreIdentical(filename, testFilename) // compare page 1 and test page 2
			if err != nil {
				return fmt.Errorf("failed to compare images during detection: %w", err)
			}

			if !identical {
				currentDirection = "ltr"
				fmt.Println("Direction detected: LTR")
				// We have page 1 and page 2.
				pageNum = 2
				continue // Continue loop, next iter will be page 3 (pageNum increment at bottom)
			} else {
				// Right didn't work. Try Left.
				fmt.Println("Content didn't change. Trying 'left' (RTL)...")
				// Remove the identical test file
				os.Remove(testFilename)

				if err := c.platform.PressKey("left"); err != nil {
					return fmt.Errorf("failed to turn page: %w", err)
				}
				c.wait(ctx)

				if err := c.platform.Screenshot(testFilename); err != nil {
					return fmt.Errorf("failed to take screenshot: %w", err)
				}

				identical, err = ImagesAreIdentical(filename, testFilename)
				if err != nil {
					return fmt.Errorf("failed to compare images during detection: %w", err)
				}

				if !identical {
					currentDirection = "rtl"
					fmt.Println("Direction detected: RTL")
					pageNum = 2
					continue
				} else {
					return fmt.Errorf("could not detect direction: content did not change in either direction")
				}
			}
		}

		keyDirection := "right"
		if currentDirection == "rtl" {
			keyDirection = "left"
		}

		if err := c.platform.PressKey(keyDirection); err != nil {
			return fmt.Errorf("failed to turn page: %w", err)
		}

		c.wait(ctx)

		pageNum++
	}

	return nil
}

func (c *Capturer) wait(ctx context.Context) {
	timer := time.NewTimer(1500 * time.Millisecond)
	defer timer.Stop()
	select {
	case <-ctx.Done():
	case <-timer.C:
	}
}
