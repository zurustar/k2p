package automation

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"os/exec"
	"strings"
)

// KindleAutomation handles interaction with the macOS Kindle application
type KindleAutomation interface {
	// IsKindleInstalled checks if Kindle app is installed
	IsKindleInstalled() (bool, error)

	// IsBookOpen detects if a book is currently open
	IsBookOpen() (bool, error)

	// IsKindleInForeground checks if Kindle app is in foreground
	IsKindleInForeground() (bool, error)

	// TurnNextPage navigates to next page
	// direction: "right" or "left" for arrow key direction
	TurnNextPage(direction string) error

	// HasMorePages detects if there are more pages (end of book detection)
	HasMorePages() (bool, error)

	// CaptureCurrentPage captures screenshot of current Kindle page
	CaptureCurrentPage() (image.Image, error)
}

// AppleScriptAutomation implements KindleAutomation using AppleScript
type AppleScriptAutomation struct{}

// NewKindleAutomation creates a new KindleAutomation instance
func NewKindleAutomation() KindleAutomation {
	return &AppleScriptAutomation{}
}

// IsKindleInstalled checks if Kindle app is installed
func (a *AppleScriptAutomation) IsKindleInstalled() (bool, error) {
	script := `
tell application "System Events"
	return exists application process "Kindle"
end tell
`
	output, err := runAppleScript(script)
	if err != nil {
		return false, fmt.Errorf("failed to check Kindle installation: %w", err)
	}

	return strings.TrimSpace(output) == "true", nil
}

// IsBookOpen detects if a book is currently open
// This checks if Kindle has a window open
func (a *AppleScriptAutomation) IsBookOpen() (bool, error) {
	script := `
tell application "System Events"
	tell process "Kindle"
		if exists then
			return count of windows > 0
		else
			return false
		end if
	end tell
end tell
`
	output, err := runAppleScript(script)
	if err != nil {
		return false, fmt.Errorf("failed to check if book is open: %w", err)
	}

	return strings.TrimSpace(output) == "true", nil
}

// IsKindleInForeground checks if Kindle app is in foreground
func (a *AppleScriptAutomation) IsKindleInForeground() (bool, error) {
	script := `
tell application "System Events"
	set frontApp to name of first application process whose frontmost is true
	return frontApp is "Kindle"
end tell
`
	output, err := runAppleScript(script)
	if err != nil {
		return false, fmt.Errorf("failed to check if Kindle is in foreground: %w", err)
	}

	return strings.TrimSpace(output) == "true", nil
}

// TurnNextPage navigates to next page by sending arrow key
// direction: "right" for right arrow, "left" for left arrow
func (a *AppleScriptAutomation) TurnNextPage(direction string) error {
	// CRITICAL: Verify Kindle is in foreground before sending keystroke
	// If Kindle lost focus, we MUST NOT send keystrokes to avoid
	// accidentally operating other applications
	inForeground, err := a.IsKindleInForeground()
	if err != nil {
		return fmt.Errorf("failed to check Kindle foreground status: %w", err)
	}
	if !inForeground {
		return fmt.Errorf("Kindle is not in foreground - terminating to prevent accidental operations on other apps")
	}

	// Use key code for arrow keys (without modifiers)
	// Right arrow = 124, Left arrow = 123
	var keyCode string
	if direction == "left" {
		keyCode = "123"
	} else {
		keyCode = "124"
	}

	script := fmt.Sprintf(`
tell application "System Events"
	tell process "Kindle"
		key code %s
	end tell
end tell
`, keyCode)

	_, err = runAppleScript(script)
	if err != nil {
		return fmt.Errorf("failed to turn page: %w", err)
	}

	return nil
}

// HasMorePages is deprecated - end detection is now done via screenshot comparison
// This method is kept for interface compatibility but always returns true
// The orchestrator uses screenshot similarity to detect the end of the book
func (a *AppleScriptAutomation) HasMorePages() (bool, error) {
	return true, nil
}

// CaptureCurrentPage is deprecated - screenshot capture is now handled by screenshot.Capturer
// This method is kept for interface compatibility but returns an error
func (a *AppleScriptAutomation) CaptureCurrentPage() (image.Image, error) {
	return nil, errors.New("use screenshot.Capturer directly for page capture")
}

// runAppleScript executes an AppleScript and returns the output
func runAppleScript(script string) (string, error) {
	cmd := exec.Command("osascript", "-e", script)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("AppleScript error: %w, stderr: %s", err, stderr.String())
	}

	return stdout.String(), nil
}
