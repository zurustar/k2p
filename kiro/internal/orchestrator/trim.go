package orchestrator

import (
	"fmt"

	"github.com/oumi/k2p/pkg/imageprocessing"
)

// trimScreenshot trims borders from a screenshot
func (o *DefaultOrchestrator) trimScreenshot(inputPath, outputPath string, verbose bool) error {
	if verbose {
		fmt.Print(" [trimming]")
	}

	return imageprocessing.TrimImageFile(inputPath, outputPath)
}

// trimScreenshotWithCustomMargins trims a screenshot using custom pixel margins
func (o *DefaultOrchestrator) trimScreenshotWithCustomMargins(inputPath, outputPath string, top, bottom, left, right int, verbose bool) error {
	return imageprocessing.TrimImageFileWithCustomMargins(inputPath, outputPath, top, bottom, left, right)
}
