package focus

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// Property 35: Bring To Front On Completion
// For any mode and option state, the window is brought to front if and only if
// the option is enabled and the mode is Generate or Detect.
func TestProperty35_ShouldBringToFront(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("brings to front only for enabled generate/detect", prop.ForAll(
		func(mode string, enabled bool) bool {
			expected := enabled && (mode == "generate" || mode == "detect")
			return ShouldBringToFront(mode, enabled) == expected
		},
		gen.OneConstOf("generate", "detect", "pdf2md", "", "unknown"),
		gen.Bool(),
	))

	properties.TestingRun(t)
}

// Property 35: Bring To Front On Completion
// The activation script always targets the given process's PID.
func TestProperty35_ActivationScriptTargetsPID(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("script sets frontmost for the given unix id", prop.ForAll(
		func(pid int) bool {
			script := activationScript(pid)
			return strings.Contains(script, `tell application "System Events"`) &&
				strings.Contains(script, fmt.Sprintf("unix id is %d)", pid)) &&
				strings.Contains(script, "set frontmost")
		},
		gen.IntRange(1, 1<<22),
	))

	properties.TestingRun(t)
}

func TestNewActivatorTargetsOwnProcess(t *testing.T) {
	a, ok := NewActivator().(*DefaultActivator)
	if !ok {
		t.Fatalf("NewActivator should return *DefaultActivator, got %T", NewActivator())
	}
	if a.pid != os.Getpid() {
		t.Errorf("expected pid %d, got %d", os.Getpid(), a.pid)
	}
}

func TestNoOpActivator(t *testing.T) {
	if err := NewNoOpActivator().BringToFront(); err != nil {
		t.Errorf("NoOpActivator.BringToFront should return nil, got %v", err)
	}
}
