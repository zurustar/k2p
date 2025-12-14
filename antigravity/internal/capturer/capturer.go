package capturer

import (
	"context"
	"fmt"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/kbinani/screenshot"
	"github.com/zurustar/k2p/internal/config"
)

// Capture starts the screen capture process
func Capture(ctx context.Context, cfg config.Config) error {
	if err := os.MkdirAll(cfg.TempDir, 0755); err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}

	// Clean up temp dir
	files, err := filepath.Glob(filepath.Join(cfg.TempDir, "*.png"))
	if err == nil {
		for _, f := range files {
			os.Remove(f)
		}
	}

	fmt.Println("Press Ctrl+C to stop early.")

	pageNum := 1
	for {
		select {
		case <-ctx.Done():
			fmt.Println("\nCapture stopped by user (context cancelled).")
			return nil
		default:
		}

		if cfg.PageCount > 0 && pageNum > cfg.PageCount {
			fmt.Printf("Reached limit of %d pages.\n", cfg.PageCount)
			break
		}

		filename := filepath.Join(cfg.TempDir, fmt.Sprintf("page_%04d.png", pageNum))
		fmt.Printf("Capturing page %d...\n", pageNum)

		if err := captureScreen(filename); err != nil {
			fmt.Printf("Failed to capture screen: %v\n", err)
			return err
		}

		if err := nextPage(cfg.Direction); err != nil {
			fmt.Printf("Failed to turn page: %v\n", err)
			return err
		}

		// Wait for animation, but listen for context
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(1500 * time.Millisecond):
		}

		pageNum++
	}

	return nil
}

func captureScreen(filename string) error {
	// Capture primary display (index 0)
	bounds := screenshot.GetDisplayBounds(0)
	img, err := screenshot.CaptureRect(bounds)
	if err != nil {
		return err
	}

	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	return png.Encode(f, img)
}

func nextPage(direction string) error {
	// Use AppleScript to press only the arrow key
	// "key code 123" is Left, "key code 124" is Right
	// Or use "key code" for better reliability than "keystroke"
	// Left Arrow: 123
	// Right Arrow: 124

	keyCode := "124"
	if direction == "rtl" {
		keyCode = "123"
	}

	script := fmt.Sprintf("tell application \"System Events\" to key code %s", keyCode)
	cmd := exec.Command("osascript", "-e", script)
	return cmd.Run()
}
