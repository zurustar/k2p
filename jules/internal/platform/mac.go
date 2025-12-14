package platform

import (
	"fmt"
	"os/exec"
)

// MacPlatform implements the Platform interface for macOS.
type MacPlatform struct{}

// NewMacPlatform creates a new instance of MacPlatform.
func NewMacPlatform() *MacPlatform {
	return &MacPlatform{}
}

// Screenshot captures the screen using the 'screencapture' command.
// It uses '-x' to silence the sound and '-m' to capture the main monitor.
// Note: Depending on the setup, we might want to capture all screens or a specific region.
// The original python script used pyautogui.screenshot() which usually captures the primary screen.
func (p *MacPlatform) Screenshot(filename string) error {
	// -x: do not play sounds
	// -m: only capture the main monitor (prevents capturing all monitors if multiple are connected)
	// -C: capture the cursor as well (optional, usually we want to hide it) -> actually we want to hide it.
	// default is to not capture cursor with -x? No, -C captures it.
	// We'll stick to -x -m.
	cmd := exec.Command("screencapture", "-x", "-m", filename)
	return cmd.Run()
}

// PressKey simulates a key press using AppleScript.
func (p *MacPlatform) PressKey(direction string) error {
	var keyCode int
	switch direction {
	case "left":
		keyCode = 123
	case "right":
		keyCode = 124
	default:
		return fmt.Errorf("unknown direction: %s", direction)
	}

	script := fmt.Sprintf("tell application \"System Events\" to key code %d", keyCode)
	cmd := exec.Command("osascript", "-e", script)
	return cmd.Run()
}
