package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/rriifftt/kindle-to-pdf-go/internal/capturer"
	"github.com/rriifftt/kindle-to-pdf-go/internal/converter"
	"github.com/rriifftt/kindle-to-pdf-go/internal/platform"
)

func main() {
	output := flag.String("output", "output.pdf", "Output PDF filename")
	tempDir := flag.String("temp-dir", "screenshots", "Temporary directory for screenshots")
	pages := flag.Int("pages", 0, "Number of pages to capture (optional, default: infinite until Ctrl+C)")
	direction := flag.String("direction", "ltr", "Page direction: 'ltr' (Left-to-Right) or 'rtl' (Right-to-Left)")
	maxSize := flag.String("max-size", "180MB", "Maximum size of generated PDF (e.g., '1.8MB', '100KB'). Default: 180MB")
	flag.Parse()

	if *direction != "ltr" && *direction != "rtl" {
		fmt.Println("Error: direction must be 'ltr' or 'rtl'")
		os.Exit(1)
	}

	maxSizeBytes, err := converter.ParseSize(*maxSize)
	if err != nil {
		fmt.Printf("Error parsing max-size: %v\n", err)
		os.Exit(1)
	}

	// Setup context for cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle Ctrl+C
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		cancel()
	}()

	// Initialize components
	p := platform.NewMacPlatform()
	cap := capturer.NewCapturer(p)
	conv := converter.NewPdfConverter()

	fmt.Println("=== Kindle for PC to PDF Converter (Go) ===")
	fmt.Println("1. Open Kindle for PC and open your book.")
	fmt.Println("2. Enter Full Screen mode (usually F11).")
	fmt.Println("3. Make sure the mouse cursor is hidden or out of the way.")
	fmt.Println("4. This script will start capturing in 5 seconds after you press Enter.")

	fmt.Print("Press Enter to start the countdown...")
	fmt.Scanln()

	cap.WaitForFocus(5)

	err = cap.CaptureLoop(ctx, *tempDir, *pages, *direction)
	if err != nil {
		fmt.Printf("Error during capture: %v\n", err)
		// We might still want to convert what we have
	}

	// Convert to PDF
	// Check if context was cancelled, if so we still proceed to convert unless critical error
	// The original python script converts after loop finishes (either by limit or keyboard interrupt)

	err = conv.ConvertImagesToPdf(*tempDir, *output, maxSizeBytes)
	if err != nil {
		fmt.Printf("Error converting to PDF: %v\n", err)
	}
}
