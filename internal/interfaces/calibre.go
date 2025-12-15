package interfaces

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

// CalibreInterface manages interaction with Calibre's ebook-convert utility
type CalibreInterface interface {
	IsInstalled() bool
	GetVersion() (string, error)
	Convert(input, output string, options map[string]string) error
}

// CalibreService implements the CalibreInterface
type CalibreService struct {
	executablePath string
}

// NewCalibreService creates a new CalibreService instance
func NewCalibreService() *CalibreService {
	return &CalibreService{
		executablePath: "ebook-convert", // Default path, will be validated
	}
}

// IsInstalled checks if Calibre's ebook-convert utility is available
func (c *CalibreService) IsInstalled() bool {
	_, err := exec.LookPath(c.executablePath)
	return err == nil
}

// GetVersion retrieves the version of the installed Calibre ebook-convert utility
func (c *CalibreService) GetVersion() (string, error) {
	if !c.IsInstalled() {
		return "", fmt.Errorf("ebook-convert is not installed or not found in PATH")
	}

	cmd := exec.Command(c.executablePath, "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get Calibre version: %w", err)
	}

	// Parse version from output like "ebook-convert (calibre 6.29.0)"
	versionRegex := regexp.MustCompile(`calibre\s+(\d+\.\d+\.\d+)`)
	matches := versionRegex.FindStringSubmatch(string(output))
	if len(matches) < 2 {
		return "", fmt.Errorf("could not parse version from output: %s", string(output))
	}

	return matches[1], nil
}

// Convert executes the ebook-convert command with the specified parameters
func (c *CalibreService) Convert(input, output string, options map[string]string) error {
	if !c.IsInstalled() {
		return fmt.Errorf("ebook-convert is not installed or not found in PATH")
	}

	if input == "" {
		return fmt.Errorf("input file path cannot be empty")
	}
	if output == "" {
		return fmt.Errorf("output file path cannot be empty")
	}

	// Build command arguments
	args := []string{input, output}
	
	// Add options as command line arguments
	for key, value := range options {
		if value == "" {
			args = append(args, "--"+key)
		} else {
			args = append(args, "--"+key, value)
		}
	}

	cmd := exec.Command(c.executablePath, args...)
	output_bytes, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("conversion failed: %w\nOutput: %s", err, string(output_bytes))
	}

	return nil
}

// ConvertWithOptions executes the ebook-convert command using ConversionOptions
func (c *CalibreService) ConvertWithOptions(input, output string, opts ConversionOptions) error {
	if !c.IsInstalled() {
		return fmt.Errorf("ebook-convert is not installed or not found in PATH")
	}

	// Validate and set defaults for options
	opts.SetDefaults()
	if err := opts.Validate(); err != nil {
		return fmt.Errorf("invalid conversion options: %w", err)
	}

	// Build Calibre command options from ConversionOptions
	calibreOptions := c.buildCalibreOptions(opts)
	
	return c.Convert(input, output, calibreOptions)
}

// buildCalibreOptions converts ConversionOptions to Calibre command line options
func (c *CalibreService) buildCalibreOptions(opts ConversionOptions) map[string]string {
	options := make(map[string]string)

	// Map quality settings to Calibre options
	switch opts.Quality {
	case "low":
		options["pdf-default-image-dpi"] = "72"
		options["pdf-image-compression"] = "2"
	case "medium":
		options["pdf-default-image-dpi"] = "150"
		options["pdf-image-compression"] = "1"
	case "high":
		options["pdf-default-image-dpi"] = "300"
		options["pdf-image-compression"] = "0"
	case "maximum":
		options["pdf-default-image-dpi"] = "600"
		options["pdf-image-compression"] = "0"
		options["pdf-use-document-margins"] = ""
	}

	// Map page size settings
	if opts.PageSize != "" {
		switch opts.PageSize {
		case "A4":
			options["pdf-page-size"] = "a4"
		case "A3":
			options["pdf-page-size"] = "a3"
		case "A5":
			options["pdf-page-size"] = "a5"
		case "Letter":
			options["pdf-page-size"] = "letter"
		case "Legal":
			options["pdf-page-size"] = "legal"
		case "Tabloid":
			options["pdf-page-size"] = "tabloid"
		}
	}

	// Map orientation settings
	if opts.Orientation == "Landscape" {
		options["pdf-landscape"] = ""
	}

	// Add custom options (these override any defaults)
	for key, value := range opts.CustomOptions {
		options[key] = value
	}

	// Add verbose output if requested
	if opts.Verbose {
		options["verbose"] = ""
	}

	return options
}

// IsCompatibleVersion checks if the installed Calibre version is compatible
func (c *CalibreService) IsCompatibleVersion() (bool, error) {
	version, err := c.GetVersion()
	if err != nil {
		return false, err
	}

	// Parse version components
	parts := strings.Split(version, ".")
	if len(parts) < 2 {
		return false, fmt.Errorf("invalid version format: %s", version)
	}

	// Check if version is >= 5.0 (minimum supported version)
	major := parts[0]
	if major >= "5" {
		return true, nil
	}

	return false, nil
}