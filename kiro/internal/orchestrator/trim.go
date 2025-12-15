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
