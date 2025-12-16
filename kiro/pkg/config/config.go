package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// ConversionOptions holds all configuration options for conversion
type ConversionOptions struct {
	// Output directory (empty = current directory)
	OutputDir string `yaml:"output_dir"`

	// Screenshot quality (1-100, default: 95)
	ScreenshotQuality int `yaml:"screenshot_quality"`

	// Delay between page turns (default: 500ms)
	PageDelay time.Duration `yaml:"page_delay"`

	// Delay before starting automation (default: 3s)
	StartupDelay time.Duration `yaml:"startup_delay"`

	// Show countdown timer during startup delay
	ShowCountdown bool `yaml:"show_countdown"`

	// PDF quality setting (low/medium/high, default: high)
	PDFQuality string `yaml:"pdf_quality"`

	// Enable verbose logging
	Verbose bool `yaml:"verbose"`

	// Auto-confirm overwrite without prompting
	AutoConfirm bool `yaml:"auto_confirm"`

	// Trim black/white borders from screenshots
	TrimBorders bool `yaml:"trim_borders"`

	// Page turn key: "right" or "left" (default: "right")
	PageTurnKey string `yaml:"page_turn_key"`

	// Configuration file path
	ConfigFile string `yaml:"-"`
}

// ConfigManager handles configuration loading and management
type ConfigManager interface {
	// LoadConfig loads configuration from file
	LoadConfig(path string) (*ConversionOptions, error)

	// MergeOptions merges CLI flags with config file settings
	MergeOptions(cliOptions, fileOptions *ConversionOptions) *ConversionOptions

	// GetDefaults returns default configuration
	GetDefaults() *ConversionOptions
}

// DefaultConfigManager is the default implementation of ConfigManager
type DefaultConfigManager struct{}

// NewConfigManager creates a new ConfigManager instance
func NewConfigManager() ConfigManager {
	return &DefaultConfigManager{}
}

// GetDefaults returns default configuration
func (cm *DefaultConfigManager) GetDefaults() *ConversionOptions {
	return &ConversionOptions{
		OutputDir:         "",
		ScreenshotQuality: 95,
		PageDelay:         500 * time.Millisecond,
		StartupDelay:      3 * time.Second,
		ShowCountdown:     true,
		PDFQuality:        "high",
		Verbose:           false,
		AutoConfirm:       false,
		TrimBorders:       false,   // Disable by default (opt-in)
		PageTurnKey:       "right", // Default to right arrow
	}
}

// LoadConfig loads configuration from file
func (cm *DefaultConfigManager) LoadConfig(path string) (*ConversionOptions, error) {
	if path == "" {
		return nil, errors.New("config file path cannot be empty")
	}

	// Expand home directory if present
	if len(path) > 0 && path[0] == '~' {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		path = filepath.Join(home, path[1:])
	}

	// Read file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse YAML
	var opts ConversionOptions
	if err := yaml.Unmarshal(data, &opts); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Validate configuration
	if err := cm.validateOptions(&opts); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &opts, nil
}

// validateOptions validates configuration options
func (cm *DefaultConfigManager) validateOptions(opts *ConversionOptions) error {
	// Validate screenshot quality
	if opts.ScreenshotQuality < 1 || opts.ScreenshotQuality > 100 {
		return fmt.Errorf("screenshot quality must be between 1 and 100, got: %d", opts.ScreenshotQuality)
	}

	// Validate PDF quality
	validPDFQualities := map[string]bool{"low": true, "medium": true, "high": true}
	if opts.PDFQuality != "" && !validPDFQualities[opts.PDFQuality] {
		return fmt.Errorf("pdf quality must be 'low', 'medium', or 'high', got: %s", opts.PDFQuality)
	}

	// Validate delays
	if opts.PageDelay < 0 {
		return fmt.Errorf("page delay cannot be negative")
	}
	if opts.StartupDelay < 0 {
		return fmt.Errorf("startup delay cannot be negative")
	}

	return nil
}

// MergeOptions merges CLI flags with config file settings
// CLI flags take precedence over config file settings
func (cm *DefaultConfigManager) MergeOptions(cliOptions, fileOptions *ConversionOptions) *ConversionOptions {
	// Start with defaults
	merged := cm.GetDefaults()

	// Apply file options if provided
	if fileOptions != nil {
		if fileOptions.OutputDir != "" {
			merged.OutputDir = fileOptions.OutputDir
		}
		if fileOptions.ScreenshotQuality != 0 {
			merged.ScreenshotQuality = fileOptions.ScreenshotQuality
		}
		if fileOptions.PageDelay != 0 {
			merged.PageDelay = fileOptions.PageDelay
		}
		if fileOptions.StartupDelay != 0 {
			merged.StartupDelay = fileOptions.StartupDelay
		}
		if fileOptions.PDFQuality != "" {
			merged.PDFQuality = fileOptions.PDFQuality
		}
		merged.ShowCountdown = fileOptions.ShowCountdown
		merged.Verbose = fileOptions.Verbose
		merged.AutoConfirm = fileOptions.AutoConfirm
		merged.TrimBorders = fileOptions.TrimBorders
		if fileOptions.PageTurnKey != "" {
			merged.PageTurnKey = fileOptions.PageTurnKey
		}
	}

	// Apply CLI options (these take precedence)
	if cliOptions != nil {
		if cliOptions.OutputDir != "" {
			merged.OutputDir = cliOptions.OutputDir
		}
		if cliOptions.ScreenshotQuality != 0 {
			merged.ScreenshotQuality = cliOptions.ScreenshotQuality
		}
		if cliOptions.PageDelay != 0 {
			merged.PageDelay = cliOptions.PageDelay
		}
		if cliOptions.StartupDelay != 0 {
			merged.StartupDelay = cliOptions.StartupDelay
		}
		if cliOptions.PDFQuality != "" {
			merged.PDFQuality = cliOptions.PDFQuality
		}
		if cliOptions.Verbose {
			merged.Verbose = true
		}
		if cliOptions.AutoConfirm {
			merged.AutoConfirm = true
		}
		if cliOptions.TrimBorders {
			merged.TrimBorders = true
		}
		if cliOptions.PageTurnKey != "" {
			merged.PageTurnKey = cliOptions.PageTurnKey
		}
		if cliOptions.ConfigFile != "" {
			merged.ConfigFile = cliOptions.ConfigFile
		}
	}

	return merged
}

// ApplyDefaults applies default values to any unset options
func ApplyDefaults(opts *ConversionOptions) *ConversionOptions {
	defaults := &DefaultConfigManager{}
	return defaults.MergeOptions(opts, nil)
}
