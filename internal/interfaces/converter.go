package interfaces

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// ConversionOptions holds configuration for PDF conversion
type ConversionOptions struct {
	Quality       string            // PDF quality setting
	PageSize      string            // A4, Letter, etc.
	Orientation   string            // Portrait, Landscape
	CustomOptions map[string]string // Additional Calibre options
	Verbose       bool              // Enable detailed logging
	Overwrite     bool              // Overwrite existing files
}

// Validate checks if the ConversionOptions are valid
func (co *ConversionOptions) Validate() error {
	// Validate quality settings
	validQualities := []string{"", "low", "medium", "high", "maximum"}
	if !contains(validQualities, co.Quality) {
		return fmt.Errorf("invalid quality setting '%s', must be one of: %s", co.Quality, strings.Join(validQualities[1:], ", "))
	}

	// Validate page size
	validPageSizes := []string{"", "A4", "A3", "A5", "Letter", "Legal", "Tabloid"}
	if !contains(validPageSizes, co.PageSize) {
		return fmt.Errorf("invalid page size '%s', must be one of: %s", co.PageSize, strings.Join(validPageSizes[1:], ", "))
	}

	// Validate orientation
	validOrientations := []string{"", "Portrait", "Landscape"}
	if !contains(validOrientations, co.Orientation) {
		return fmt.Errorf("invalid orientation '%s', must be one of: %s", co.Orientation, strings.Join(validOrientations[1:], ", "))
	}

	return nil
}

// SetDefaults sets default values for empty or invalid fields
func (co *ConversionOptions) SetDefaults() {
	validQualities := []string{"", "low", "medium", "high", "maximum"}
	if !contains(validQualities, co.Quality) {
		co.Quality = "high"
	} else if co.Quality == "" {
		co.Quality = "high"
	}

	validPageSizes := []string{"", "A4", "A3", "A5", "Letter", "Legal", "Tabloid"}
	if !contains(validPageSizes, co.PageSize) {
		co.PageSize = "A4"
	} else if co.PageSize == "" {
		co.PageSize = "A4"
	}

	validOrientations := []string{"", "Portrait", "Landscape"}
	if !contains(validOrientations, co.Orientation) {
		co.Orientation = "Portrait"
	} else if co.Orientation == "" {
		co.Orientation = "Portrait"
	}

	if co.CustomOptions == nil {
		co.CustomOptions = make(map[string]string)
	}
}

// ConversionResult represents the result of a single conversion
type ConversionResult struct {
	InputFile    string
	OutputFile   string
	Success      bool
	Error        error
	Duration     time.Duration
	FileSize     int64
}

// Validate checks if the ConversionResult is valid
func (cr *ConversionResult) Validate() error {
	if cr.InputFile == "" {
		return errors.New("input file cannot be empty")
	}
	if cr.OutputFile == "" {
		return errors.New("output file cannot be empty")
	}
	if cr.Success && cr.Error != nil {
		return errors.New("successful conversion cannot have an error")
	}
	if !cr.Success && cr.Error == nil {
		return errors.New("failed conversion must have an error")
	}
	if cr.Duration < 0 {
		return errors.New("duration cannot be negative")
	}
	if cr.FileSize < 0 {
		return errors.New("file size cannot be negative")
	}
	return nil
}

// BatchResult represents the result of batch conversion
type BatchResult struct {
	TotalFiles      int
	SuccessfulFiles int
	FailedFiles     int
	Results         []ConversionResult
	TotalDuration   time.Duration
}

// Validate checks if the BatchResult is valid
func (br *BatchResult) Validate() error {
	if br.TotalFiles < 0 {
		return errors.New("total files cannot be negative")
	}
	if br.SuccessfulFiles < 0 {
		return errors.New("successful files cannot be negative")
	}
	if br.FailedFiles < 0 {
		return errors.New("failed files cannot be negative")
	}
	if br.SuccessfulFiles+br.FailedFiles != br.TotalFiles {
		return errors.New("successful files + failed files must equal total files")
	}
	if len(br.Results) != br.TotalFiles {
		return errors.New("number of results must match total files")
	}
	if br.TotalDuration < 0 {
		return errors.New("total duration cannot be negative")
	}

	// Validate each result
	for i, result := range br.Results {
		if err := result.Validate(); err != nil {
			return fmt.Errorf("invalid result at index %d: %w", i, err)
		}
	}

	return nil
}

// SupportedFormat represents a supported file format
type SupportedFormat struct {
	Extension   string
	Description string
	MimeType    string
}

// Validate checks if the SupportedFormat is valid
func (sf *SupportedFormat) Validate() error {
	if sf.Extension == "" {
		return errors.New("extension cannot be empty")
	}
	if !strings.HasPrefix(sf.Extension, ".") {
		return errors.New("extension must start with a dot")
	}
	if sf.Description == "" {
		return errors.New("description cannot be empty")
	}
	if sf.MimeType == "" {
		return errors.New("mime type cannot be empty")
	}
	return nil
}

// ConverterService orchestrates the conversion process
type ConverterService interface {
	ConvertSingle(input, output string, options ConversionOptions) error
	ConvertBatch(inputDir, outputDir string, options ConversionOptions) (*BatchResult, error)
	ValidateFile(filepath string) error
}

// contains checks if a slice contains a specific string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}