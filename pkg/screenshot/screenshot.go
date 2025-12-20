package screenshot

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// Capturer handles screenshot capture operations
type Capturer interface {
	// CaptureFrontmostWindow captures a screenshot of the frontmost window
	// This method activates Kindle and waits for it to come to front
	CaptureFrontmostWindow(outputPath string) error

	// CaptureWithoutActivation captures a screenshot without activating Kindle
	// Returns error if Kindle is not already in the foreground
	CaptureWithoutActivation(outputPath string) error
}

// MacOSCapturer implements screenshot capture for macOS
type MacOSCapturer struct{}

// NewCapturer creates a new screenshot capturer
func NewCapturer() Capturer {
	return &MacOSCapturer{}
}

// CaptureFrontmostWindow captures a screenshot of the Kindle window
// Since Kindle should be in fullscreen mode, we activate it and capture the frontmost window
func (c *MacOSCapturer) CaptureFrontmostWindow(outputPath string) error {
	// Activate Kindle to bring it to front
	// Note: Application name is "Amazon Kindle" but process name is "Kindle"
	activateScript := `
tell application "Amazon Kindle"
	activate
end tell
`
	activateCmd := exec.Command("osascript", "-e", activateScript)
	var activateStderr bytes.Buffer
	activateCmd.Stderr = &activateStderr
	if err := activateCmd.Run(); err != nil {
		return fmt.Errorf("failed to activate Kindle: %w, stderr: %s", err, activateStderr.String())
	}

	// Wait longer for Kindle to come to front and for Space to switch
	// Fullscreen apps are in separate Spaces, so we need time for the switch
	time.Sleep(2 * time.Second)

	// Verify Kindle is in foreground
	checkScript := `
tell application "System Events"
	set frontApp to name of first application process whose frontmost is true
	return frontApp is "Kindle"
end tell
`
	checkCmd := exec.Command("osascript", "-e", checkScript)
	output, err := checkCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to verify Kindle is frontmost: %w", err)
	}

	if strings.TrimSpace(string(output)) != "true" {
		return fmt.Errorf("Kindle is not in foreground after activation")
	}

	// Capture entire screen
	// With screen recording permission, this captures the active Space (Kindle fullscreen)
	// -x: disable sound
	captureCmd := exec.Command("screencapture", "-x", outputPath)
	if err := captureCmd.Run(); err != nil {
		return fmt.Errorf("failed to capture screenshot: %w", err)
	}

	return nil
}

// CaptureWithoutActivation captures a screenshot without activating Kindle
// This is much faster than CaptureFrontmostWindow as it skips activation and waiting
// Returns error if Kindle is not in the foreground
func (c *MacOSCapturer) CaptureWithoutActivation(outputPath string) error {
	// Verify Kindle is in foreground (fail fast if not)
	checkScript := `
tell application "System Events"
	set frontApp to name of first application process whose frontmost is true
	return frontApp is "Kindle"
end tell
`
	checkCmd := exec.Command("osascript", "-e", checkScript)
	output, err := checkCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to verify Kindle is frontmost: %w", err)
	}

	if strings.TrimSpace(string(output)) != "true" {
		return fmt.Errorf("Kindle is not in foreground. Please keep Kindle active during conversion")
	}

	// Capture entire screen
	// With screen recording permission, this captures the active Space (Kindle fullscreen)
	// -x: disable sound
	captureCmd := exec.Command("screencapture", "-x", outputPath)
	if err := captureCmd.Run(); err != nil {
		return fmt.Errorf("failed to capture screenshot: %w", err)
	}

	return nil
}
