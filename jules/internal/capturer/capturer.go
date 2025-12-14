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

		keyDirection := "right"
		if direction == "rtl" {
			keyDirection = "left"
		}
		if err := c.platform.PressKey(keyDirection); err != nil {
			return fmt.Errorf("failed to turn page: %w", err)
		}

		// Wait for page turn animation
		// Kindle for PC animation can be slow. 1.5s is a safe starting point.
		// Use a ticker or sleep? Sleep is fine as it needs to be sequential.
		// Check context during sleep is better but sleep is short.
		timer := time.NewTimer(1500 * time.Millisecond)
		select {
		case <-ctx.Done():
			timer.Stop()
			fmt.Println("\nCapture stopped by user.")
			return nil
		case <-timer.C:
		}

		pageNum++
	}

	return nil
}
