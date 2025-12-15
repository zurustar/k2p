package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/zurustar/k2p/internal/capturer"
	"github.com/zurustar/k2p/internal/config"
	"github.com/zurustar/k2p/internal/converter"
)

func parseSize(s string) int64 {
	s = strings.ToUpper(s)
	multiplier := int64(1)
	if strings.HasSuffix(s, "KB") {
		multiplier = 1024
		s = strings.TrimSuffix(s, "KB")
	} else if strings.HasSuffix(s, "MB") {
		multiplier = 1024 * 1024
		s = strings.TrimSuffix(s, "MB")
	} else if strings.HasSuffix(s, "GB") {
		multiplier = 1024 * 1024 * 1024
		s = strings.TrimSuffix(s, "GB")
	}

	val, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return int64(val * float64(multiplier))
}

func main() {
	output := flag.String("output", "output.pdf", "Output PDF filename")
	tempDir := flag.String("temp-dir", "screenshots", "Temporary directory for screenshots")
	pages := flag.Int("pages", 0, "Number of pages to capture (optional, default: infinite until Ctrl+C)")
	direction := flag.String("direction", "ltr", "Page direction: 'ltr' or 'rtl'")
	maxSize := flag.String("max-size", "180MB", "Maximum size of generated PDF")

	flag.Parse()

	cfg := config.Config{
		Output:     *output,
		TempDir:    *tempDir,
		PageCount:  *pages,
		Direction:  *direction,
		MaxSizeStr: *maxSize,
		MaxSize:    parseSize(*maxSize),
	}

	fmt.Println("=== Kindle to PDF Converter (Go) ===")
	fmt.Println("1. Open Kindle for PC/Mac and open your book.")
	fmt.Println("2. Enter Full Screen mode.")
	fmt.Println("3. Make sure the mouse cursor is hidden.")
	fmt.Println("4. This script will start capturing in 5 seconds after you press Enter.")

	fmt.Print("Press Enter to start the countdown...")
	fmt.Scanln()

	fmt.Println("Starting in 5 seconds... Please focus the Kindle window!")
	time.Sleep(5 * time.Second)

	// Handle Ctrl+C gracefully
	ctx, cancel := context.WithCancel(context.Background())
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("\nSignal received, stopping capture...")
		cancel()
	}()

	// Capturer needs to be interruptible.
	// Since `capturer.Capture` is blocking, we need to modify it to accept a context or channel?
	// Or just run it. If user presses Ctrl+C, the default behavior terminates.
	// To convert AFTER Ctrl+C, we need to catch it.
	// But `capturer.Capture` is busy loop.
	// Let's modify `capturer.Capture` to be smarter later if needed.
	// For now, let's assume the user will let it finish if `--pages` is set,
	// or we accept that Ctrl+C kills it immediately without conversion.
	// Wait, original tool feature: "Press Ctrl+C to stop capturing early" -> proceeds to conversion.
	// So we DO need to handle this.

	// We'll update capturer code slightly to handle this in next step if verification fails,
	// but for now let's implement the basic main.
	// Actually, let's just run Capture. If it returns, Convert.
	// If user sends SIGINT, the program exits.
	// To fix this, we'd need to run Capture in a goroutine or handle signals in Capture.

	err := capturer.Capture(ctx, cfg)
	if err != nil {
		fmt.Printf("Capture finished with error: %v\n", err)
	}

	// Always attempt conversion if we have images
	// (Converter will handle "no images" case gracefully)
	err = converter.Convert(cfg)
	if err != nil {
		fmt.Printf("Conversion failed: %v\n", err)
	}
}
