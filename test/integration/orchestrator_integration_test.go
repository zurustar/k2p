package integration

import (
	"context"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"testing"
	"time"

	"github.com/oumi/k2p/internal/config"
	"github.com/oumi/k2p/internal/filemanager"
	"github.com/oumi/k2p/internal/orchestrator"
	"github.com/oumi/k2p/internal/pdf"
)

// Mock Automation for Integration
type MockIntegrationAutomation struct {
	CapturedPages []string
}

func (m *MockIntegrationAutomation) IsKindleInstalled() (bool, error)    { return true, nil }
func (m *MockIntegrationAutomation) IsBookOpen() (bool, error)           { return true, nil }
func (m *MockIntegrationAutomation) IsKindleInForeground() (bool, error) { return true, nil }
func (m *MockIntegrationAutomation) BringKindleToForeground() error      { return nil }
func (m *MockIntegrationAutomation) TurnNextPage(direction string) error { return nil }
func (m *MockIntegrationAutomation) HasMorePages() (bool, error)         { return true, nil }

func TestOrchestratorIntegration_FullWorkflow(t *testing.T) {
	// Setup temporary output directory
	outputDir, err := os.MkdirTemp("", "k2p_integration")
	if err != nil {
		t.Fatalf("Failed to create temp output dir: %v", err)
	}
	defer os.RemoveAll(outputDir)

	// Dependencies
	fm := filemanager.NewFileManager()
	pg := pdf.NewPDFGenerator()
	auto := &MockIntegrationAutomation{}

	// Use special capturer for end detection simulation
	capturer := &MockIntegrationCapturerForEndDetection{}

	// Inject dependencies
	orch := orchestrator.NewOrchestratorWithDeps(auto, fm, pg, capturer)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	opts := &config.ConversionOptions{
		OutputDir:   outputDir,
		AutoConfirm: true,
		Mode:        "generate",
		PageDelay:   10 * time.Millisecond,
		Verbose:     true,
	}

	fmt.Println("Starting integration test conversion...")
	result, err := orch.ConvertCurrentBook(ctx, opts)
	if err != nil {
		t.Fatalf("Conversion failed: %v", err)
	}

	// Verification

	// 1. Check result struct
	if result.PageCount == 0 {
		t.Error("Expected PageCount > 0")
	}
	// Page 1 (Count 5): Valid
	// Page 2 (Count 6): Valid
	// Page 3 (Count 7): Valid
	// Page 4 (Count 8): End
	// Page 5 (Count 9): End
	// Page 6 (Count 10): End
	// Page 7 (Count 11): End
	// Page 8 (Count 12): End --> Detected (last 5 identical)
	//
	// Total screenshots: 4 (detection) + 8 (pages) = 12
	// Screenshots list has: [Det1, Det2, Det3, Det4, P1, P2, P3, P4, P5, P6, P7, P8]
	// Detection (Cover..R3) are added inherently by being files in usage.
	// End detection removes last 5 (P4..P8).
	// Remaining: [Det1..Det4, P1, P2, P3] = 7 pages.

	if result.PageCount != 7 {
		t.Errorf("Expected 7 pages (4 detection + 3 valid content), got %d", result.PageCount)
	}

	// 2. Check output PDF existence
	if _, err := os.Stat(result.OutputPath); os.IsNotExist(err) {
		t.Errorf("Output PDF not found at %s", result.OutputPath)
	}

	// 3. Verify PDF validity (basic)
	content, err := os.ReadFile(result.OutputPath)
	if err == nil {
		if len(content) < 4 || string(content[:4]) != "%PDF" {
			t.Error("Output file is not a valid PDF (header check)")
		}
	}

	fmt.Printf("Integration test success! Generated PDF at %s with %d pages\n", result.OutputPath, result.PageCount)
}

// Helper to generate identical images for end detection
type MockIntegrationCapturerForEndDetection struct {
	Count int
}

func (m *MockIntegrationCapturerForEndDetection) CaptureWithoutActivation(path string) error {
	m.Count++

	img := image.NewRGBA(image.Rect(0, 0, 100, 100))

	var col color.RGBA
	// Distinct images for initial sequence to satisfy detection
	if m.Count <= 4 {
		// Detection images (1-4)
		// 1: 50, 2: 100, 3: 150, 4: 200
		col = color.RGBA{R: uint8(m.Count * 50), G: 100, B: 100, A: 255}
	} else if m.Count <= 7 {
		// Valid pages (5, 6, 7)
		// Make sure these are distinct from Count 4 (R=200) and from each other
		// 5: R=50, G=200
		// 6: R=100, G=200
		// 7: R=150, G=200
		// Use Green channel difference to ensure visual distinction from Detection (Red channel dominant)
		val := uint8((m.Count - 4) * 50)
		col = color.RGBA{R: val, G: 200, B: 100, A: 255}
	} else {
		// End pages (8+) - identical (White)
		col = color.RGBA{R: 255, G: 255, B: 255, A: 255}
	}

	// Add a stripe to ensure structure is different
	// This helps avoiding similarity matches
	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			img.Set(x, y, col)
		}
	}
	// Add a stripe to ensure structure is distinct for valid pages
	// BUT for end pages (8+), it must be IDENTICAL.
	// So cap the stripe length or remove it for end pages.

	stripeLen := m.Count * 5
	if m.Count >= 8 {
		stripeLen = 100 // Constant length for end pages (full width)
	}

	stripeCol := color.RGBA{R: 0, G: 0, B: 0, A: 255}
	for x := 0; x < stripeLen && x < 100; x++ {
		img.Set(x, 50, stripeCol)
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}
func (m *MockIntegrationCapturerForEndDetection) CaptureFrontmostWindow(path string) error {
	return m.CaptureWithoutActivation(path)
}
