package interfaces

import (
	"testing"
	"time"
)

func TestConversionOptionsDefaults(t *testing.T) {
	opts := ConversionOptions{
		Quality:     "high",
		PageSize:    "A4",
		Orientation: "portrait",
		Verbose:     false,
		Overwrite:   false,
	}

	if opts.Quality != "high" {
		t.Errorf("Expected quality 'high', got '%s'", opts.Quality)
	}
	if opts.PageSize != "A4" {
		t.Errorf("Expected page size 'A4', got '%s'", opts.PageSize)
	}
	if opts.Orientation != "portrait" {
		t.Errorf("Expected orientation 'portrait', got '%s'", opts.Orientation)
	}
}

func TestConversionResult(t *testing.T) {
	result := ConversionResult{
		InputFile:  "test.azw3",
		OutputFile: "test.pdf",
		Success:    true,
		Error:      nil,
		Duration:   time.Second * 5,
		FileSize:   1024,
	}

	if !result.Success {
		t.Error("Expected success to be true")
	}
	if result.Error != nil {
		t.Errorf("Expected no error, got %v", result.Error)
	}
	if result.Duration != time.Second*5 {
		t.Errorf("Expected duration 5s, got %v", result.Duration)
	}
}

func TestBatchResult(t *testing.T) {
	batch := BatchResult{
		TotalFiles:      3,
		SuccessfulFiles: 2,
		FailedFiles:     1,
		Results:         make([]ConversionResult, 3),
		TotalDuration:   time.Second * 15,
	}

	if batch.TotalFiles != 3 {
		t.Errorf("Expected 3 total files, got %d", batch.TotalFiles)
	}
	if batch.SuccessfulFiles != 2 {
		t.Errorf("Expected 2 successful files, got %d", batch.SuccessfulFiles)
	}
	if batch.FailedFiles != 1 {
		t.Errorf("Expected 1 failed file, got %d", batch.FailedFiles)
	}
}

// **Feature: kindle-to-pdf-go, Property 20: Invalid configuration fallback**
// **Validates: Requirements 6.4**
func TestProperty20_InvalidConfigurationFallback(t *testing.T) {
	// Property: For any invalid configuration values, default values should be used and the user should be warned
	
	// Test with a controlled set of invalid and valid values
	testValues := []string{
		"", // empty (should be valid - uses defaults)
		"invalid_value", // clearly invalid
		"high", "medium", "low", "maximum", // valid quality values
		"A4", "A3", "A5", "Letter", "Legal", // valid page sizes
		"Portrait", "Landscape", // valid orientations
		"random_invalid", "123", "!@#$%", // more invalid values
	}
	
	for i := 0; i < 100; i++ {
		// Pick random values from our controlled set
		quality := testValues[i%len(testValues)]
		pageSize := testValues[(i+1)%len(testValues)]
		orientation := testValues[(i+2)%len(testValues)]
		
		opts := ConversionOptions{
			Quality:     quality,
			PageSize:    pageSize,
			Orientation: orientation,
		}
		
		// Validate the options
		err := opts.Validate()
		
		// If validation fails, set defaults and validate again
		if err != nil {
			opts.SetDefaults()
			// After setting defaults, validation should pass
			if opts.Validate() != nil {
				t.Errorf("After SetDefaults(), validation should pass but failed for quality=%s, pageSize=%s, orientation=%s", quality, pageSize, orientation)
			}
			if opts.Quality == "" || opts.PageSize == "" || opts.Orientation == "" {
				t.Errorf("SetDefaults() should set non-empty values, got quality=%s, pageSize=%s, orientation=%s", opts.Quality, opts.PageSize, opts.Orientation)
			}
		}
		// If validation passed originally, the values were valid - this is fine
	}
}

// Generator for invalid configuration values to test fallback behavior
func TestInvalidConfigurationFallbackExamples(t *testing.T) {
	testCases := []struct {
		name        string
		opts        ConversionOptions
		expectError bool
	}{
		{
			name: "invalid quality",
			opts: ConversionOptions{
				Quality:     "invalid_quality",
				PageSize:    "A4",
				Orientation: "Portrait",
			},
			expectError: true,
		},
		{
			name: "invalid page size",
			opts: ConversionOptions{
				Quality:     "high",
				PageSize:    "invalid_size",
				Orientation: "Portrait",
			},
			expectError: true,
		},
		{
			name: "invalid orientation",
			opts: ConversionOptions{
				Quality:     "high",
				PageSize:    "A4",
				Orientation: "invalid_orientation",
			},
			expectError: true,
		},
		{
			name: "all valid values",
			opts: ConversionOptions{
				Quality:     "high",
				PageSize:    "A4",
				Orientation: "Portrait",
			},
			expectError: false,
		},
		{
			name: "empty values should be valid (will use defaults)",
			opts: ConversionOptions{
				Quality:     "",
				PageSize:    "",
				Orientation: "",
			},
			expectError: false,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.opts.Validate()
			if tc.expectError && err == nil {
				t.Errorf("Expected validation error for %s, but got none", tc.name)
			}
			if !tc.expectError && err != nil {
				t.Errorf("Expected no validation error for %s, but got: %v", tc.name, err)
			}
			
			// Test that SetDefaults always produces valid configuration
			tc.opts.SetDefaults()
			if err := tc.opts.Validate(); err != nil {
				t.Errorf("After SetDefaults(), validation should pass for %s, but got: %v", tc.name, err)
			}
		})
	}
}