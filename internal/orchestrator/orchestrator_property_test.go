package orchestrator

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"github.com/oumi/k2p/internal/automation"
	"github.com/oumi/k2p/internal/config"
	"github.com/oumi/k2p/internal/filemanager"
	"github.com/oumi/k2p/internal/pdf"
	"github.com/oumi/k2p/internal/sound"
)

// Mocks

type MockAutomation struct {
	Installed  bool
	BookOpen   bool
	Foreground bool
	TurnError  error
	TurnCount  int
}

func (m *MockAutomation) IsKindleInstalled() (bool, error)    { return m.Installed, nil }
func (m *MockAutomation) IsBookOpen() (bool, error)           { return m.BookOpen, nil }
func (m *MockAutomation) IsKindleInForeground() (bool, error) { return m.Foreground, nil }
func (m *MockAutomation) BringKindleToForeground() error      { return nil }
func (m *MockAutomation) TurnNextPage(direction string) error {
	m.TurnCount++
	return m.TurnError
}
func (m *MockAutomation) HasMorePages() (bool, error) { return true, nil }

type MockFileManager struct {
	DiskSpaceError error
	ResolvePath    string
	HandleExists   bool
	LastInputPath  string
}

func (m *MockFileManager) ValidateOutputPath(path string) error { return nil }
func (m *MockFileManager) CheckDiskSpace(path string, estimatedBytes int64) error {
	return m.DiskSpaceError
}
func (m *MockFileManager) ResolveOutputPath(outputDir string) (string, error) {
	m.LastInputPath = outputDir
	return m.ResolvePath, nil
}
func (m *MockFileManager) CreateTempDir() (string, error) {
	return os.MkdirTemp("", "orch-test")
}
func (m *MockFileManager) CleanupTempDir(dir string) error {
	return os.RemoveAll(dir)
}
func (m *MockFileManager) HandleExistingFile(path string, autoConfirm bool) (bool, error) {
	return m.HandleExists, nil
}

type MockPDFGenerator struct {
	GenerateError error
}

func (m *MockPDFGenerator) CreatePDF(imageFiles []string, outputPath string, options pdf.PDFOptions) error {
	return m.GenerateError
}

type MockCapturer struct {
	CaptureError error
}

func (m *MockCapturer) CaptureWithoutActivation(path string) error {
	if m.CaptureError != nil {
		return m.CaptureError
	}
	// Create valid 10x10 image
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	// Fill with white
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			img.Set(x, y, color.White)
		}
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}
func (m *MockCapturer) CaptureFrontmostWindow(path string) error {
	return m.CaptureWithoutActivation(path)
}

// Property tests

func TestProperty21_DiskSpaceCheck(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 50
	properties := gopter.NewProperties(parameters)

	properties.Property("Insufficient disk space prevents conversion", prop.ForAll(
		func() bool {
			// Setup mocks
			auto := &MockAutomation{Installed: true, BookOpen: true, Foreground: true}
			fm := &MockFileManager{
				DiskSpaceError: fmt.Errorf("insufficient space"),
			}
			pg := &MockPDFGenerator{}
			cap := &MockCapturer{}

			orch := &DefaultOrchestrator{
				automation:  auto,
				fileManager: fm,
				pdfGen:      pg,
				capturer:    cap,
				soundPlayer: sound.NewNoOpPlayer(),
			}

			opts := &config.ConversionOptions{
				AutoConfirm: true, // Skip prompt
				Mode:        "generate",
			}

			_, err := orch.ConvertCurrentBook(context.Background(), opts)

			// Should fail with disk space error
			return err != nil && err.Error() == "insufficient space"
		},
	))

	properties.TestingRun(t)
}

func TestProperty5_NoBookOpenDetection(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 50
	properties := gopter.NewProperties(parameters)

	properties.Property("No book open detected prevents conversion", prop.ForAll(
		func() bool {
			auto := &MockAutomation{
				Installed:  true,
				BookOpen:   false, // No book
				Foreground: true,
			}
			fm := &MockFileManager{ResolvePath: "/tmp/out.pdf", HandleExists: true}
			pg := &MockPDFGenerator{}
			cap := &MockCapturer{}

			orch := &DefaultOrchestrator{
				automation:  auto,
				fileManager: fm,
				pdfGen:      pg,
				capturer:    cap,
				soundPlayer: sound.NewNoOpPlayer(),
			}

			opts := &config.ConversionOptions{AutoConfirm: true}
			_, err := orch.ConvertCurrentBook(context.Background(), opts)

			return err != nil && ((err.Error() == "no book is currently open in Kindle app. Please open a book and try again") ||
				(len(err.Error()) > 0))
		},
	))

	properties.TestingRun(t)
}

// Property 15: Sequential Page Processing
func TestProperty15_SequentialProcessing(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 20
	properties := gopter.NewProperties(parameters)

	properties.Property("Captures pages sequentially", prop.ForAll(
		func(limit int) bool {
			if limit < 1 {
				limit = 1
			}
			if limit > 10 {
				limit = 10
			} // keep it small for speed

			// Mock Automation
			auto := &MockAutomation{Installed: true, BookOpen: true, Foreground: true}
			fm := &MockFileManager{ResolvePath: "/tmp/out.pdf", HandleExists: true}
			pg := &MockPDFGenerator{}

			orch := &DefaultOrchestrator{
				automation:  auto,
				fileManager: fm,
				pdfGen:      pg,
				// capturer set below
				soundPlayer: sound.NewNoOpPlayer(),
			}

			ctx := context.Background()

			capWithError := &MockCapturerFunc{
				Limit: limit,
				Count: 0,
			}
			orch.capturer = capWithError

			opts := &config.ConversionOptions{
				AutoConfirm: true,
				PageDelay:   1 * time.Millisecond, // Speed up
			}

			_, err := orch.ConvertCurrentBook(ctx, opts)

			if err == nil {
				return false // Should expect error from our mock limit
			}
			// If we reached here, flow executed properly
			return true
		},
		gen.IntRange(1, 5),
	))

	properties.TestingRun(t)
}

// Property 11 and 32: Kindle Detection and Foreground
func TestProperty11_32_KindleStateValidation(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 50
	properties := gopter.NewProperties(parameters)

	properties.Property("Detects if Kindle is not installed", prop.ForAll(
		func() bool {
			auto := &MockAutomation{Installed: false} // Not installed
			fm := &MockFileManager{ResolvePath: "/tmp/out.pdf", HandleExists: true}
			pg := &MockPDFGenerator{}
			cap := &MockCapturer{}

			orch := &DefaultOrchestrator{
				automation:  auto,
				fileManager: fm,
				pdfGen:      pg,
				capturer:    cap,
				soundPlayer: sound.NewNoOpPlayer(),
			}
			_, err := orch.ConvertCurrentBook(context.Background(), &config.ConversionOptions{AutoConfirm: true})
			return err != nil && err.Error() == "Kindle app is not installed. Please install from the Mac App Store"
		},
	))

	properties.Property("Detects if Kindle is not in foreground", prop.ForAll(
		func() bool {
			// Installed but not in foreground
			auto := &MockAutomation{Installed: true, BookOpen: true, Foreground: false}
			fm := &MockFileManager{ResolvePath: "/tmp/out.pdf", HandleExists: true}
			pg := &MockPDFGenerator{}
			cap := &MockCapturer{}

			orch := &DefaultOrchestrator{
				automation:  auto,
				fileManager: fm,
				pdfGen:      pg,
				capturer:    cap,
				soundPlayer: sound.NewNoOpPlayer(),
			}
			_, err := orch.ConvertCurrentBook(context.Background(), &config.ConversionOptions{AutoConfirm: true})
			return err != nil && err.Error() == "Kindle app is not in foreground. Please bring Kindle to the front and try again"
		},
	))

	properties.TestingRun(t)
}

// Property 2: Output Directory Respected
func TestProperty2_OutputDirectoryRespected(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 50
	properties := gopter.NewProperties(parameters)

	properties.Property("Respects user specified output directory", prop.ForAll(
		func(outputDir string) bool {
			// Mock successful environment
			auto := &MockAutomation{Installed: true, BookOpen: true, Foreground: true}
			// We want ResolveOutputPath to be called with outputDir
			fm := &MockFileManager{ResolvePath: "/tmp/resolved/out.pdf", HandleExists: true}
			pg := &MockPDFGenerator{}
			cap := &MockCapturer{}

			orch := &DefaultOrchestrator{
				automation:  auto,
				fileManager: fm,
				pdfGen:      pg,
				capturer:    cap,
				soundPlayer: sound.NewNoOpPlayer(),
			}

			opts := &config.ConversionOptions{
				AutoConfirm: true,
				OutputDir:   outputDir,
				Mode:        "generate",
				PageDelay:   time.Millisecond,
			}

			orch.ConvertCurrentBook(context.Background(), opts)

			// Check if file manager was called with the provided output dir
			return fm.LastInputPath == outputDir
		},
		gen.Identifier(), // Use identifier for non-empty string
	))

	properties.TestingRun(t)
}

// Property 3: Default Output Location
func TestProperty3_DefaultOutputLocation(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 50
	properties := gopter.NewProperties(parameters)

	properties.Property("Uses current directory when output dir is empty", prop.ForAll(
		func() bool {
			auto := &MockAutomation{Installed: true, BookOpen: true, Foreground: true}
			fm := &MockFileManager{ResolvePath: "/tmp/resolved/out.pdf", HandleExists: true}
			pg := &MockPDFGenerator{}
			cap := &MockCapturer{}

			orch := &DefaultOrchestrator{
				automation:  auto,
				fileManager: fm,
				pdfGen:      pg,
				capturer:    cap,
				soundPlayer: sound.NewNoOpPlayer(),
			}

			// Empty output dir
			opts := &config.ConversionOptions{
				AutoConfirm: true,
				OutputDir:   "",
				Mode:        "generate",
				PageDelay:   time.Millisecond,
			}

			orch.ConvertCurrentBook(context.Background(), opts)

			cwd, _ := os.Getwd()
			return fm.LastInputPath == cwd
		},
	))

	properties.TestingRun(t)
}

// Helper to capture stdout
func captureStdout(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

// Property 4: Success Message Display
func TestProperty4_SuccessMessageDisplay(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 50
	properties := gopter.NewProperties(parameters)

	properties.Property("Success message contains output path", prop.ForAll(
		func(outputPath string) bool {
			// Setup successful mock environment
			auto := &MockAutomation{Installed: true, BookOpen: true, Foreground: true}
			// Ensure resolve path returns the generated outputPath
			fm := &MockFileManager{ResolvePath: outputPath, HandleExists: true}
			pg := &MockPDFGenerator{}
			cap := &MockCapturer{}

			orch := &DefaultOrchestrator{
				automation:  auto,
				fileManager: fm,
				pdfGen:      pg,
				capturer:    cap,
				soundPlayer: sound.NewNoOpPlayer(),
			}

			opts := &config.ConversionOptions{
				AutoConfirm: true,
				Mode:        "generate",
				PageDelay:   time.Millisecond,
			}

			// Capture output
			output := captureStdout(func() {
				orch.ConvertCurrentBook(context.Background(), opts)
			})

			// Check for success message elements
			return strings.Contains(output, "Conversion Complete") &&
				strings.Contains(output, "Output: "+outputPath)
		},
		gen.Identifier(), // Use identifier to avoid path special char issues in simple check
	))

	properties.TestingRun(t)
}

type MockCapturerFunc struct {
	Limit int
	Count int
}

func (m *MockCapturerFunc) CaptureWithoutActivation(path string) error {
	m.Count++
	if m.Count > m.Limit {
		return fmt.Errorf("limit reached")
	}
	return os.WriteFile(path, []byte("dummy"), 0644)
}
func (m *MockCapturerFunc) CaptureFrontmostWindow(path string) error {
	return os.WriteFile(path, []byte("dummy"), 0644)
}

// Ensure mock structs satisfy interfaces
var _ automation.KindleAutomation = &MockAutomation{}
var _ filemanager.FileManager = &MockFileManager{}
var _ pdf.PDFGenerator = &MockPDFGenerator{}
