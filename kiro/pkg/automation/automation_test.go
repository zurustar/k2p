package automation

import (
	"testing"
)

// Note: These tests require Kindle app to be installed and may require manual setup
// They are marked as integration tests and can be skipped with -short flag

func TestIsKindleInstalled(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	automation := NewKindleAutomation()
	installed, err := automation.IsKindleInstalled()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// This test will pass or fail depending on whether Kindle is installed
	t.Logf("Kindle installed: %v", installed)
}

func TestIsBookOpen(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	automation := NewKindleAutomation()

	// First check if Kindle is installed
	installed, err := automation.IsKindleInstalled()
	if err != nil {
		t.Fatalf("failed to check installation: %v", err)
	}
	if !installed {
		t.Skip("Kindle not installed, skipping test")
	}

	bookOpen, err := automation.IsBookOpen()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	t.Logf("Book open: %v", bookOpen)
}

func TestIsKindleInForeground(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	automation := NewKindleAutomation()

	inForeground, err := automation.IsKindleInForeground()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	t.Logf("Kindle in foreground: %v", inForeground)
}

func TestTurnNextPage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	automation := NewKindleAutomation()

	// Check prerequisites
	installed, _ := automation.IsKindleInstalled()
	if !installed {
		t.Skip("Kindle not installed")
	}

	bookOpen, _ := automation.IsBookOpen()
	if !bookOpen {
		t.Skip("No book open")
	}

	inForeground, _ := automation.IsKindleInForeground()
	if !inForeground {
		t.Skip("Kindle not in foreground")
	}

	// Attempt to turn page
	err := automation.TurnNextPage("right")
	if err != nil {
		t.Errorf("failed to turn page: %v", err)
	}
}

func TestHasMorePages(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	automation := NewKindleAutomation()

	hasMore, err := automation.HasMorePages()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Current implementation always returns true
	if !hasMore {
		t.Error("expected hasMore to be true")
	}
}

// Unit tests for AppleScript execution

func TestRunAppleScript(t *testing.T) {
	// Test simple AppleScript
	script := `return "hello"`
	output, err := runAppleScript(script)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if output != "hello\n" {
		t.Errorf("expected 'hello\\n', got %q", output)
	}
}

func TestRunAppleScriptError(t *testing.T) {
	// Test invalid AppleScript
	script := `this is not valid applescript`
	_, err := runAppleScript(script)

	if err == nil {
		t.Error("expected error for invalid AppleScript")
	}
}
