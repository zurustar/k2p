package focus

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
)

// Activator defines the interface for bringing the app to front
type Activator interface {
	BringToFront() error
}

// DefaultActivator makes its own process frontmost via System Events (macOS)
type DefaultActivator struct {
	pid int
}

// NewActivator creates an activator that targets the current process
func NewActivator() Activator {
	return &DefaultActivator{pid: os.Getpid()}
}

// BringToFront makes the process frontmost.
// Requires the same Accessibility permission already used for page turning.
func (a *DefaultActivator) BringToFront() error {
	cmd := exec.Command("osascript", "-e", activationScript(a.pid))
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to bring window to front: %w, stderr: %s", err, stderr.String())
	}
	return nil
}

func activationScript(pid int) string {
	return fmt.Sprintf(`tell application "System Events" to set frontmost of (first process whose unix id is %d) to true`, pid)
}

// ShouldBringToFront reports whether the window should be brought to front
// after a run. Only Generate and Detect runs qualify, regardless of success.
func ShouldBringToFront(mode string, enabled bool) bool {
	return enabled && (mode == "generate" || mode == "detect")
}

// NoOpActivator is an activator that does nothing (for testing)
type NoOpActivator struct{}

// NewNoOpActivator creates a new no-op activator
func NewNoOpActivator() Activator {
	return &NoOpActivator{}
}

// BringToFront does nothing
func (a *NoOpActivator) BringToFront() error {
	return nil
}
