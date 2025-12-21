package config

import (
	"fmt"
	"time"
)

// ConversionOptions holds all configuration options for conversion
type ConversionOptions struct {
	// Output directory (empty = current directory)
	OutputDir string

	// Screenshot quality (1-100, default: 95)
	ScreenshotQuality int

	// Delay between page turns (default: 500ms)
	PageDelay time.Duration

	// Delay before starting automation (default: 3s)
	StartupDelay time.Duration

	// Show countdown timer during startup delay
	ShowCountdown bool

	// PDF quality setting (low/medium/high, default: high)
	PDFQuality string

	// Enable verbose logging
	Verbose bool

	// Auto-confirm overwrite without prompting
	AutoConfirm bool

	// Operation mode: "detect" (analyze margins) or "generate" (create PDF)
	// Default: "generate"
	Mode string

	// Custom trim margins in pixels (default: 0 = no trimming)
	// Used when Mode == "generate" and any value is non-zero
	TrimTop        int
	TrimBottom     int
	TrimHorizontal int

	// Page turn key: "right" or "left" (default: "right")
	PageTurnKey string
}

// ApplyDefaults applies default values to any unset options
func ApplyDefaults(opts *ConversionOptions) *ConversionOptions {
	// Create a copy to avoid modifying original if nil (though usually not nil here)
	merged := &ConversionOptions{
		// Set defaults first
		ScreenshotQuality: 95,
		PageDelay:         500 * time.Millisecond,
		StartupDelay:      3 * time.Second,
		ShowCountdown:     true,
		PDFQuality:        "high",
		Verbose:           false,
		AutoConfirm:       false,
		Mode:              "generate",
		TrimTop:           0,
		TrimBottom:        0,
		TrimHorizontal:    0,

		PageTurnKey: "right",
	}

	if opts == nil {
		return merged
	}

	// Override with provided options if set
	if opts.OutputDir != "" {
		merged.OutputDir = opts.OutputDir
	}
	if opts.ScreenshotQuality != 0 {
		merged.ScreenshotQuality = opts.ScreenshotQuality
	}
	if opts.PageDelay != 0 {
		merged.PageDelay = opts.PageDelay
	}
	if opts.StartupDelay != 0 {
		merged.StartupDelay = opts.StartupDelay
	}
	if opts.PDFQuality != "" {
		merged.PDFQuality = opts.PDFQuality
	}

	// For boolean flags (Verbose, AutoConfirm), we only check if true because
	// CLI flags default to false.
	if opts.Verbose {
		merged.Verbose = true
	}
	if opts.AutoConfirm {
		merged.AutoConfirm = true
	}

	// ShowCountdown is not exposed in CLI, so we stick to default (true)
	// unless we decide to expose it later.

	if opts.Mode != "" {
		merged.Mode = opts.Mode
	}
	if opts.TrimTop != 0 {
		merged.TrimTop = opts.TrimTop
	}
	if opts.TrimBottom != 0 {
		merged.TrimBottom = opts.TrimBottom
	}
	if opts.TrimHorizontal != 0 {
		merged.TrimHorizontal = opts.TrimHorizontal
	}

	if opts.PageTurnKey != "" {
		merged.PageTurnKey = opts.PageTurnKey
	}

	return merged
}

// Validate checks if the options are valid
func (o *ConversionOptions) Validate() error {
	if o.ScreenshotQuality < 1 || o.ScreenshotQuality > 100 {
		return fmt.Errorf("screenshot quality must be between 1 and 100")
	}

	validPDFQualities := map[string]bool{"low": true, "medium": true, "high": true}
	if !validPDFQualities[o.PDFQuality] {
		return fmt.Errorf("pdf quality must be 'low', 'medium', or 'high'")
	}

	return nil
}
