package orchestrator

import (
	"github.com/oumi/k2p/internal/imageprocessing"
)

// trimScreenshotWithCustomMargins trims a screenshot using custom pixel margins
func (o *DefaultOrchestrator) trimScreenshotWithCustomMargins(inputPath, outputPath string, top, bottom, left, right int, verbose bool) error {
	return imageprocessing.TrimImageFileWithCustomMargins(inputPath, outputPath, top, bottom, left, right)
}
