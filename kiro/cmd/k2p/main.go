package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"

	"github.com/oumi/k2p/internal/orchestrator"
	"github.com/oumi/k2p/pkg/config"
)

const version = "0.1.0-debug-20251216"

var buildTime = "unknown" // Set via -ldflags during build

func main() {
	// Define CLI flags
	var (
		outputDir        = flag.String("output", "", "Output directory (default: current directory)")
		outputShort      = flag.String("o", "", "Output directory (shorthand)")
		quality          = flag.Int("quality", 0, "Screenshot quality 1-100 (default: 95)")
		pageDelay        = flag.Duration("page-delay", 0, "Delay between page turns (default: 500ms)")
		startupDelay     = flag.Duration("startup-delay", 0, "Delay before starting automation (default: 3s)")
		pdfQuality       = flag.String("pdf-quality", "", "PDF quality: low, medium, high (default: high)")
		configFile       = flag.String("config", "", "Configuration file path")
		verbose          = flag.Bool("verbose", false, "Enable verbose logging")
		verboseShort     = flag.Bool("v", false, "Enable verbose logging (shorthand)")
		autoConfirm      = flag.Bool("auto-confirm", false, "Skip confirmation prompts")
		autoConfirmShort = flag.Bool("y", false, "Skip confirmation prompts (shorthand)")
		trimBorders      = flag.Bool("trim-borders", false, "Trim black/white borders from screenshots")
		pageTurnKey      = flag.String("page-turn-key", "right", "Page turn direction: 'right' or 'left'")
		showVersion      = flag.Bool("version", false, "Show version information")
		showHelp         = flag.Bool("help", false, "Show help message")
		showHelpShort    = flag.Bool("h", false, "Show help message (shorthand)")
	)

	flag.Parse()

	// Handle version flag
	if *showVersion {
		fmt.Printf("k2p version %s\n", version)
		fmt.Printf("Built: %s\n", buildTime)
		return
	}

	// Handle help flag
	if *showHelp || *showHelpShort {
		printHelp()
		return
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

	// Build CLI options
	cliOpts := &config.ConversionOptions{
		OutputDir:   *outputDir,
		Verbose:     *verbose,
		AutoConfirm: *autoConfirm,
		ConfigFile:  *configFile,
		TrimBorders: *trimBorders, // Opt-in: only trim when explicitly enabled
		PageTurnKey: *pageTurnKey,
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

	// Load config file if specified
	var fileOpts *config.ConversionOptions
	if *configFile != "" {
		cm := config.NewConfigManager()
		var err error
		fileOpts, err = cm.LoadConfig(*configFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading config file: %v\n", err)
			os.Exit(1)
		}
	}

	// Merge options
	cm := config.NewConfigManager()
	finalOpts := cm.MergeOptions(cliOpts, fileOpts)

	// Validate final options
	if finalOpts.ScreenshotQuality < 1 || finalOpts.ScreenshotQuality > 100 {
		fmt.Fprintf(os.Stderr, "Error: screenshot quality must be between 1 and 100\n")
		os.Exit(1)
	}

	validPDFQualities := map[string]bool{"low": true, "medium": true, "high": true}
	if !validPDFQualities[finalOpts.PDFQuality] {
		fmt.Fprintf(os.Stderr, "Error: pdf quality must be 'low', 'medium', or 'high'\n")
		os.Exit(1)
	}

	// Create orchestrator
	orch := orchestrator.NewOrchestrator()

	// Set up signal handling for graceful shutdown
	ctx := context.Background()
	var tempDirToCleanup string
	ctx = orchestrator.SetupSignalHandler(ctx, func() {
		if tempDirToCleanup != "" {
			fmt.Println("Cleaning up temporary files...")
			// Best effort cleanup
			os.RemoveAll(tempDirToCleanup)
		}
	})

	// Display version before conversion
	if finalOpts.Verbose {
		fmt.Printf("k2p version: %s\n", version)
		fmt.Printf("Built: %s\n\n", buildTime)
	}

	// Run conversion
	result, err := orch.ConvertCurrentBook(ctx, finalOpts)
	if err != nil {
		// Play error sound to alert user (useful when Kindle is fullscreen)
		exec.Command("afplay", "/System/Library/Sounds/Basso.aiff").Start()
		fmt.Fprintf(os.Stderr, "\nError: %v\n", err)
		os.Exit(1)
	}

	// Success
	if len(result.Warnings) > 0 {
		fmt.Println("\nWarnings:")
		for _, warning := range result.Warnings {
			fmt.Printf("  - %s\n", warning)
		}
	}

	// Display build time
	fmt.Printf("Binary built: %s\n", buildTime)

	os.Exit(0)
}

func printHelp() {
	fmt.Printf(`k2p - Kindle to PDF Converter v%s

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
    --config FILE             Configuration file path
    --trim-borders            Enable border trimming (removes black/white borders)
                              Note: Most effective for fixed-layout books
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

    # Use configuration file
    k2p --config config.yaml

    # Verbose mode with auto-confirm
    k2p -v -y

    # Enable border trimming
    k2p --trim-borders

CONFIGURATION FILE:
    You can create a YAML configuration file to set default options:

    output_dir: ~/Documents/Kindle-PDFs
    screenshot_quality: 95
    page_delay: 500ms
    startup_delay: 3s
    show_countdown: true
    pdf_quality: high
    trim_borders: false  # Set to true to enable border trimming
    verbose: false
    auto_confirm: false

    See config.example.yaml for a complete example.

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
