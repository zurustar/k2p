package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
)

// Config holds runtime parameters supplied by flags.
type Config struct {
	Pages          int
	OutputDir      string
	FilePrefix     string
	Delay          time.Duration
	ActivateKindle bool
	PrepDelay      time.Duration
}

func main() {
	cfg := parseFlags()

	if runtime.GOOS != "darwin" {
		exitWithError(errors.New("this tool currently targets macOS because it drives the Kindle app and uses screencapture"))
	}

	if cfg.Pages <= 0 {
		exitWithError(fmt.Errorf("invalid pages value %d; must be > 0", cfg.Pages))
	}

	if err := os.MkdirAll(cfg.OutputDir, 0o755); err != nil {
		exitWithError(fmt.Errorf("failed to create output directory: %w", err))
	}

	if cfg.ActivateKindle {
		if err := runAppleScript(`tell application "Kindle" to activate`); err != nil {
			exitWithError(fmt.Errorf("failed to activate Kindle: %w", err))
		}
	}

	time.Sleep(cfg.PrepDelay)

	for i := 1; i <= cfg.Pages; i++ {
		filename := fmt.Sprintf("%s-%04d.png", cfg.FilePrefix, i)
		dest := filepath.Join(cfg.OutputDir, filename)

		if err := captureScreen(dest); err != nil {
			exitWithError(fmt.Errorf("capture failed on page %d: %w", i, err))
		}

		if i == cfg.Pages {
			break
		}

		if err := turnKindlePage(); err != nil {
			exitWithError(fmt.Errorf("failed to advance to page %d: %w", i+1, err))
		}

		time.Sleep(cfg.Delay)
	}

	fmt.Printf("Saved %d screenshots to %s\n", cfg.Pages, cfg.OutputDir)
}

func parseFlags() Config {
	pages := flag.Int("pages", 0, "Number of pages to capture (required)")
	outputDir := flag.String("output", "screens", "Directory to place captured images")
	prefix := flag.String("prefix", "kindle", "Filename prefix for captured pages")
	delay := flag.Duration("delay", 1200*time.Millisecond, "Delay after advancing pages before capture")
	activate := flag.Bool("activate", true, "Bring the Kindle app to the front before starting")
	prepDelay := flag.Duration("prep-delay", 2*time.Second, "Time to wait after activation before the first capture")
	flag.Parse()

	return Config{
		Pages:          *pages,
		OutputDir:      *outputDir,
		FilePrefix:     *prefix,
		Delay:          *delay,
		ActivateKindle: *activate,
		PrepDelay:      *prepDelay,
	}
}

func captureScreen(path string) error {
	cmd := exec.Command("screencapture", "-x", path)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("screencapture: %w (output: %s)", err, string(output))
	}
	return nil
}

func turnKindlePage() error {
	script := `tell application "System Events"
  tell process "Kindle"
    key code 124
  end tell
end tell`
	return runAppleScript(script)
}

func runAppleScript(script string) error {
	cmd := exec.Command("osascript", "-l", "AppleScript", "-e", script)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("osascript: %w (output: %s)", err, string(output))
	}
	return nil
}

func exitWithError(err error) {
	fmt.Fprintln(os.Stderr, "Error:", err)
	os.Exit(1)
}
