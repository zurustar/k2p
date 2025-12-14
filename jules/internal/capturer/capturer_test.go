package capturer

import (
	"context"
	"os"
	"testing"

	"github.com/rriifftt/kindle-to-pdf-go/internal/platform"
)

func TestCaptureLoop(t *testing.T) {
	mockPlatform := platform.NewMockPlatform()
	c := NewCapturer(mockPlatform)

	tempDir := "test_screenshots"
	defer os.RemoveAll(tempDir)

	// Test 3 pages
	pages := 3
	ctx := context.Background()
	err := c.CaptureLoop(ctx, tempDir, pages, "ltr")
	if err != nil {
		t.Fatalf("CaptureLoop failed: %v", err)
	}

	// Verify screenshots count
	if len(mockPlatform.Screenshots) != pages {
		t.Errorf("Expected %d screenshots, got %d", pages, len(mockPlatform.Screenshots))
	}

	// Verify key presses count
	if len(mockPlatform.KeyPresses) != pages {
		t.Errorf("Expected %d key presses, got %d", pages, len(mockPlatform.KeyPresses))
	}

	// Verify files exist
	entries, err := os.ReadDir(tempDir)
	if err != nil {
		t.Fatalf("Failed to read dir: %v", err)
	}
	if len(entries) != pages {
		t.Errorf("Expected %d files in dir, got %d", pages, len(entries))
	}
}
