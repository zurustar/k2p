package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/oumi/k2p/internal/config"
	"github.com/oumi/k2p/internal/converter"
	"github.com/oumi/k2p/internal/orchestrator"
)

const version = "0.1.0-debug-20251216"

var buildTime = "unknown" // Set via -ldflags during build

func main() {
	exitCode := run(os.Args[1:], os.Stdout, os.Stderr, orchestrator.NewOrchestrator())
	os.Exit(exitCode)
}

func run(args []string, stdout, stderr io.Writer, orch orchestrator.ConversionOrchestrator) int {
	// Define CLI flags using a specific FlagSet
	fs := flag.NewFlagSet("k2p", flag.ContinueOnError)
	fs.SetOutput(stderr)

	var (
		outputDir        = fs.String("output", "", "Output directory (default: current directory)")
		outputShort      = fs.String("o", "", "Output directory (shorthand)")
		quality          = fs.Int("quality", 0, "Screenshot quality 1-100 (default: 95)")
		pageDelay        = fs.Duration("page-delay", 0, "Delay between page turns (default: 500ms)")
		startupDelay     = fs.Duration("startup-delay", 0, "Delay before starting automation (default: 3s)")
		pdfQuality       = fs.String("pdf-quality", "", "PDF quality: low, medium, high (default: high)")
		verbose          = fs.Bool("verbose", false, "Enable verbose logging")
		verboseShort     = fs.Bool("v", false, "Enable verbose logging (shorthand)")
		autoConfirm      = fs.Bool("auto-confirm", false, "Skip confirmation prompts")
		autoConfirmShort = fs.Bool("y", false, "Skip confirmation prompts (shorthand)")
		mode             = fs.String("mode", "generate", "Operation mode: 'detect' (analyze margins) or 'generate' (create PDF)")
		trimTop          = fs.Int("trim-top", 0, "Pixels to trim from top edge (0 = no trim)")
		trimBottom       = fs.Int("trim-bottom", 0, "Pixels to trim from bottom edge (0 = no trim)")
		trimHorizontal   = fs.Int("trim-horizontal", 0, "Pixels to trim from both left and right edges (0 = no trim)")
		inputFile        = fs.String("input", "", "Input PDF file path (required for pdf2md mode)")
		inputShort       = fs.String("i", "", "Input PDF file path (shorthand)")

		pageTurnKey   = fs.String("page-turn-key", "right", "Page turn direction: 'right' or 'left'")
		showVersion   = fs.Bool("version", false, "Show version information")
		showHelp      = fs.Bool("help", false, "Show help message")
		showHelpShort = fs.Bool("h", false, "Show help message (shorthand)")
	)

	// Parse flags
	if err := fs.Parse(args); err != nil {
		// Flag parsing error (already printed by fs)
		return 1 // or 2 per convention? Go flags usually output and return error.
	}

	// Handle version flag
	if *showVersion {
		fmt.Fprintf(stdout, "k2p version %s\n", version)
		fmt.Fprintf(stdout, "Built: %s\n", buildTime)
		return 0
	}

	// Handle help flag
	if *showHelp || *showHelpShort {
		printHelp(stdout)
		return 0
	}

	// Merge shorthand flags
	if *outputShort != "" {
		*outputDir = *outputShort
	}
	if *verboseShort {
		*verbose = true
	}
	if *autoConfirmShort {
		*autoConfirm = true
	}
	if *inputShort != "" {
		*inputFile = *inputShort
	}

	// Build CLI options
	cliOpts := &config.ConversionOptions{
		OutputDir:      *outputDir,
		Verbose:        *verbose,
		AutoConfirm:    *autoConfirm,
		Mode:           *mode,
		TrimTop:        *trimTop,
		TrimBottom:     *trimBottom,
		TrimHorizontal: *trimHorizontal,

		PageTurnKey: *pageTurnKey,
		InputFile:   *inputFile,
	}

	if *quality > 0 {
		cliOpts.ScreenshotQuality = *quality
	}
	if *pageDelay > 0 {
		cliOpts.PageDelay = *pageDelay
	}
	if *startupDelay > 0 {
		cliOpts.StartupDelay = *startupDelay
	}
	if *pdfQuality != "" {
		cliOpts.PDFQuality = *pdfQuality
	}

	// Apply defaults and merge CLI options
	finalOpts := config.ApplyDefaults(cliOpts)

	// Validate final options
	// Validate final options
	if err := finalOpts.Validate(); err != nil {
		fmt.Fprintf(stderr, "Error: %v\n", err)
		return 1
	}

	// Set up signal handling for graceful shutdown
	ctx := context.Background()
	var tempDirToCleanup string
	ctx = orchestrator.SetupSignalHandler(ctx, func() {
		if tempDirToCleanup != "" {
			fmt.Fprintf(stdout, "Cleaning up temporary files...\n")
			// Best effort cleanup
			os.RemoveAll(tempDirToCleanup)
		}
	})

	// Display version before conversion
	if finalOpts.Verbose {
		fmt.Fprintf(stdout, "k2p version: %s\n", version)
		fmt.Fprintf(stdout, "Built: %s\n\n", buildTime)
	}

	// Handle mode-specific actions
	if finalOpts.Mode == "pdf2md" {
		// Output file defaults to input file with .md extension if not specified
		outputPath := finalOpts.OutputDir
		if outputPath == "" {
			// Determine from input path
			// Basic implementation: replace extension
			ext := ".pdf" // simplistic assumption or check
			inputLen := len(finalOpts.InputFile)
			if inputLen > 4 && finalOpts.InputFile[inputLen-4:] == ext {
				outputPath = finalOpts.InputFile[:inputLen-4] + ".md"
			} else {
				outputPath = finalOpts.InputFile + ".md"
			}
		}

		if finalOpts.Verbose {
			fmt.Fprintf(stdout, "Converting PDF to Markdown...\n")
			fmt.Fprintf(stdout, "Input: %s\n", finalOpts.InputFile)
			fmt.Fprintf(stdout, "Output: %s\n", outputPath)
		}

		// Create converter and execute
		conv := converter.NewConverter()
		if err := conv.ConvertPDFToMarkdown(ctx, finalOpts.InputFile, outputPath); err != nil {
			fmt.Fprintf(stderr, "Error converting PDF to Markdown: %v\n", err)
			return 1
		}
		fmt.Fprintf(stdout, "Successfully converted to %s\n", outputPath)
		return 0
	}

	// Run standard conversion (detect or generate)
	result, err := orch.ConvertCurrentBook(ctx, finalOpts)
	if err != nil {
		// Play error sound to alert user (useful when Kindle is fullscreen)
		exec.Command("afplay", "/System/Library/Sounds/Basso.aiff").Start()
		fmt.Fprintf(stderr, "\nError: %v\n", err)
		return 1
	}

	// Success
	if len(result.Warnings) > 0 {
		fmt.Fprintf(stdout, "\nWarnings:\n")
		for _, warning := range result.Warnings {
			fmt.Fprintf(stdout, "  - %s\n", warning)
		}
	}

	// Display build time
	fmt.Fprintf(stdout, "Binary built: %s\n", buildTime)

	return 0
}

func printHelp(w io.Writer) {
	fmt.Fprintf(w, `k2p - Kindle to PDF Converter v%s

USAGE:
    k2p [OPTIONS]

DESCRIPTION:
    Converts the currently open Kindle book to PDF format by automating
    page turns and screenshot capture on macOS.

    IMPORTANT: Before running k2p:
      1. Open a book in the Kindle app
      2. Ensure the Kindle app window is visible
      3. Be ready to bring Kindle to the foreground when prompted

OPTIONS:
    -o, --output DIR          Output directory (default: current directory)
    --quality NUM             Screenshot quality 1-100 (default: 95)
    --page-delay DURATION     Delay between page turns (default: 500ms)
    --startup-delay DURATION  Delay before starting automation (default: 3s)
    --pdf-quality LEVEL       PDF quality: low, medium, high (default: high)
    --mode MODE               Operation mode (default: generate)
                              - detect: Analyze margins, no PDF output
                              - generate: Create PDF with optional trimming
                              - pdf2md: Convert existing PDF to Markdown (requires --input)
    --input, -i FILE          Input PDF file path (required for pdf2md mode)
    --trim-top PIXELS         Pixels to trim from top (0 = no trim, default: 0)
    --trim-bottom PIXELS      Pixels to trim from bottom (0 = no trim, default: 0)
    --trim-horizontal PIXELS  Pixels to trim from left and right (0 = no trim, default: 0)

    -v, --verbose             Enable verbose logging
    -y, --auto-confirm        Skip confirmation prompts
    --version                 Show version information
    -h, --help                Show this help message

EXAMPLES:
    # Convert currently open book to current directory
    k2p

    # Specify output directory
    k2p --output ~/Documents/MyBooks

    # High quality conversion with custom delays
    k2p --quality 100 --page-delay 1s --pdf-quality high

    # Verbose mode with auto-confirm
    k2p -v -y

    # Two-step trimming workflow:
    # Step 1: Detect optimal margins (no PDF output)
    k2p --mode detect -v

    # Step 2: Generate PDF with detected margins
    # Step 2: Generate PDF with detected margins
    k2p --mode generate --trim-top 50 --trim-bottom 50 --trim-horizontal 30


TRIMMING WORKFLOW:
    Kindle books have consistent app margins, but pages may have varying
    content margins. Use the two-step workflow for optimal results:

    1. Detection Mode (--mode detect):
       - Captures all pages and analyzes margins
       - Reports minimum safe trim values for each edge
       - Does NOT generate a PDF

    2. Generation Mode (--mode generate):
       - Creates PDF with optional custom trimming
       - Specify trim values for edges you want to trim (0 = no trim)
       - Example: --trim-horizontal 30 (trims 30px from both left/right)

    3. Markdown Conversion (--mode pdf2md):
       - Extracts text from an existing PDF (with embedded text)
       - Uses macOS native text extraction (like Preview)
       - Example: k2p --mode pdf2md --input book.pdf



TROUBLESHOOTING:
    "Kindle app is not installed"
      → Install Kindle from the Mac App Store

    "No book is currently open"
      → Open a book in Kindle before running k2p

    "Kindle app is not in foreground"
      → Bring the Kindle window to the front

    "Insufficient disk space"
      → Free up disk space (needs ~1-2 MB per page)
`, version)
}
