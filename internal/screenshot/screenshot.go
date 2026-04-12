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

// getKindleWindowBounds returns the Kindle window bounds as "x,y,width,height"
// in global screen coordinates, which works correctly on multi-display setups.
func getKindleWindowBounds() (string, error) {
	script := `
tell application "System Events"
	tell process "Kindle"
		if (count windows) = 0 then
			return "NO_WINDOW"
		end if
		set theWindow to window 1
		set pos to position of theWindow
		set sz to size of theWindow
		set x to (item 1 of pos) as integer as text
		set y to (item 2 of pos) as integer as text
		set w to (item 1 of sz) as integer as text
		set h to (item 2 of sz) as integer as text
		return x & "," & y & "," & w & "," & h
	end tell
end tell
`
	cmd := exec.Command("osascript", "-e", script)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get Kindle window bounds: %w, stderr: %s", err, stderr.String())
	}
	boundsStr := strings.TrimSpace(string(output))
	if boundsStr == "NO_WINDOW" || boundsStr == "" {
		return "", fmt.Errorf("Kindle window not found. Please ensure Kindle is open")
	}
	return boundsStr, nil
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

	// Get Kindle window bounds and capture by region.
	// screencapture -R captures a specific rectangle in global screen coordinates,
	// so this works correctly whether Kindle is on the primary or a secondary display.
	boundsStr, err := getKindleWindowBounds()
	if err != nil {
		// Fall back to full-screen capture (works for single-display / fullscreen Space setups)
		captureCmd := exec.Command("screencapture", "-x", outputPath)
		if err2 := captureCmd.Run(); err2 != nil {
			return fmt.Errorf("failed to capture screenshot: %w", err2)
		}
		return nil
	}

	captureCmd := exec.Command("screencapture", "-x", "-R", boundsStr, outputPath)
	if err := captureCmd.Run(); err != nil {
		return fmt.Errorf("failed to capture screenshot: %w", err)
	}

	return nil
}

// CaptureWithoutActivation captures a screenshot of the Kindle window by its
// screen coordinates. This avoids activating Kindle (which would steal focus
// from the GUI) and works correctly on multi-display setups where
// screencapture -x would otherwise capture the wrong display.
func (c *MacOSCapturer) CaptureWithoutActivation(outputPath string) error {
	// Get Kindle window bounds without activating it.
	// screencapture -R uses global screen coordinates, so this works on any display.
	boundsStr, err := getKindleWindowBounds()
	if err != nil {
		return fmt.Errorf("Kindle is not available: %w", err)
	}

	captureCmd := exec.Command("screencapture", "-x", "-R", boundsStr, outputPath)
	if err := captureCmd.Run(); err != nil {
		return fmt.Errorf("failed to capture Kindle window: %w", err)
	}

	return nil
}
