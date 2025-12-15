package interfaces

import (
	"testing"
	"testing/quick"
)

// **Feature: kindle-to-pdf-go, Property 17: Quality settings application**
// **Validates: Requirements 6.1**
func TestProperty17_QualitySettingsApplication(t *testing.T) {
	// Property: For any specified PDF quality setting, those settings should be applied during the conversion process
	
	calibreService := NewCalibreService()
	
	// Test with all valid quality settings
	validQualities := []string{"low", "medium", "high", "maximum"}
	
	for _, quality := range validQualities {
		t.Run("quality_"+quality, func(t *testing.T) {
			opts := ConversionOptions{
				Quality:       quality,
				PageSize:      "A4",
				Orientation:   "Portrait",
				CustomOptions: make(map[string]string),
				Verbose:       false,
				Overwrite:     false,
			}
			
			// Build Calibre options from ConversionOptions
			calibreOptions := calibreService.buildCalibreOptions(opts)
			
			// Verify that quality-specific options are set correctly
			switch quality {
			case "low":
				if calibreOptions["pdf-default-image-dpi"] != "72" {
					t.Errorf("Expected DPI 72 for low quality, got %s", calibreOptions["pdf-default-image-dpi"])
				}
				if calibreOptions["pdf-image-compression"] != "2" {
					t.Errorf("Expected compression 2 for low quality, got %s", calibreOptions["pdf-image-compression"])
				}
			case "medium":
				if calibreOptions["pdf-default-image-dpi"] != "150" {
					t.Errorf("Expected DPI 150 for medium quality, got %s", calibreOptions["pdf-default-image-dpi"])
				}
				if calibreOptions["pdf-image-compression"] != "1" {
					t.Errorf("Expected compression 1 for medium quality, got %s", calibreOptions["pdf-image-compression"])
				}
			case "high":
				if calibreOptions["pdf-default-image-dpi"] != "300" {
					t.Errorf("Expected DPI 300 for high quality, got %s", calibreOptions["pdf-default-image-dpi"])
				}
				if calibreOptions["pdf-image-compression"] != "0" {
					t.Errorf("Expected compression 0 for high quality, got %s", calibreOptions["pdf-image-compression"])
				}
			case "maximum":
				if calibreOptions["pdf-default-image-dpi"] != "600" {
					t.Errorf("Expected DPI 600 for maximum quality, got %s", calibreOptions["pdf-default-image-dpi"])
				}
				if calibreOptions["pdf-image-compression"] != "0" {
					t.Errorf("Expected compression 0 for maximum quality, got %s", calibreOptions["pdf-image-compression"])
				}
				if _, exists := calibreOptions["pdf-use-document-margins"]; !exists {
					t.Error("Expected pdf-use-document-margins to be set for maximum quality")
				}
			}
		})
	}
	
	// Property-based test using quick.Check
	f := func(quality string, pageSize string, orientation string, verbose bool) bool {
		// Only test with valid quality values
		validQualities := []string{"low", "medium", "high", "maximum"}
		isValidQuality := false
		for _, vq := range validQualities {
			if quality == vq {
				isValidQuality = true
				break
			}
		}
		if !isValidQuality {
			return true // Skip invalid inputs
		}
		
		opts := ConversionOptions{
			Quality:       quality,
			PageSize:      "A4", // Use fixed valid values for other fields
			Orientation:   "Portrait",
			CustomOptions: make(map[string]string),
			Verbose:       verbose,
			Overwrite:     false,
		}
		
		calibreOptions := calibreService.buildCalibreOptions(opts)
		
		// Verify that DPI setting is always present for any quality
		dpi, exists := calibreOptions["pdf-default-image-dpi"]
		if !exists {
			return false
		}
		
		// Verify DPI values are correct for each quality level
		switch quality {
		case "low":
			return dpi == "72"
		case "medium":
			return dpi == "150"
		case "high":
			return dpi == "300"
		case "maximum":
			return dpi == "600"
		}
		
		return false
	}
	
	if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
		t.Errorf("Property test failed: %v", err)
	}
}
// **Feature: kindle-to-pdf-go, Property 18: Page configuration application**
// **Validates: Requirements 6.2**
func TestProperty18_PageConfigurationApplication(t *testing.T) {
	// Property: For any specified page size or orientation settings, the PDF output should be configured accordingly
	
	calibreService := NewCalibreService()
	
	// Test with all valid page sizes
	validPageSizes := []string{"A4", "A3", "A5", "Letter", "Legal", "Tabloid"}
	
	for _, pageSize := range validPageSizes {
		t.Run("pagesize_"+pageSize, func(t *testing.T) {
			opts := ConversionOptions{
				Quality:       "high",
				PageSize:      pageSize,
				Orientation:   "Portrait",
				CustomOptions: make(map[string]string),
				Verbose:       false,
				Overwrite:     false,
			}
			
			// Build Calibre options from ConversionOptions
			calibreOptions := calibreService.buildCalibreOptions(opts)
			
			// Verify that page size is set correctly
			expectedPageSize := ""
			switch pageSize {
			case "A4":
				expectedPageSize = "a4"
			case "A3":
				expectedPageSize = "a3"
			case "A5":
				expectedPageSize = "a5"
			case "Letter":
				expectedPageSize = "letter"
			case "Legal":
				expectedPageSize = "legal"
			case "Tabloid":
				expectedPageSize = "tabloid"
			}
			
			if calibreOptions["pdf-page-size"] != expectedPageSize {
				t.Errorf("Expected page size %s for %s, got %s", expectedPageSize, pageSize, calibreOptions["pdf-page-size"])
			}
		})
	}
	
	// Test orientation settings
	t.Run("orientation_portrait", func(t *testing.T) {
		opts := ConversionOptions{
			Quality:       "high",
			PageSize:      "A4",
			Orientation:   "Portrait",
			CustomOptions: make(map[string]string),
			Verbose:       false,
			Overwrite:     false,
		}
		
		calibreOptions := calibreService.buildCalibreOptions(opts)
		
		// Portrait should not set the landscape flag
		if _, exists := calibreOptions["pdf-landscape"]; exists {
			t.Error("Portrait orientation should not set pdf-landscape flag")
		}
	})
	
	t.Run("orientation_landscape", func(t *testing.T) {
		opts := ConversionOptions{
			Quality:       "high",
			PageSize:      "A4",
			Orientation:   "Landscape",
			CustomOptions: make(map[string]string),
			Verbose:       false,
			Overwrite:     false,
		}
		
		calibreOptions := calibreService.buildCalibreOptions(opts)
		
		// Landscape should set the landscape flag
		if _, exists := calibreOptions["pdf-landscape"]; !exists {
			t.Error("Landscape orientation should set pdf-landscape flag")
		}
	})
	
	// Property-based test using quick.Check
	f := func(pageSize string, orientation string, quality string) bool {
		// Only test with valid values
		validPageSizes := []string{"A4", "A3", "A5", "Letter", "Legal", "Tabloid"}
		validOrientations := []string{"Portrait", "Landscape"}
		validQualities := []string{"low", "medium", "high", "maximum"}
		
		isValidPageSize := false
		for _, vps := range validPageSizes {
			if pageSize == vps {
				isValidPageSize = true
				break
			}
		}
		
		isValidOrientation := false
		for _, vo := range validOrientations {
			if orientation == vo {
				isValidOrientation = true
				break
			}
		}
		
		isValidQuality := false
		for _, vq := range validQualities {
			if quality == vq {
				isValidQuality = true
				break
			}
		}
		
		if !isValidPageSize || !isValidOrientation || !isValidQuality {
			return true // Skip invalid inputs
		}
		
		opts := ConversionOptions{
			Quality:       quality,
			PageSize:      pageSize,
			Orientation:   orientation,
			CustomOptions: make(map[string]string),
			Verbose:       false,
			Overwrite:     false,
		}
		
		calibreOptions := calibreService.buildCalibreOptions(opts)
		
		// Verify page size is always set for valid page sizes
		if _, exists := calibreOptions["pdf-page-size"]; !exists {
			return false
		}
		
		// Verify landscape flag is set correctly
		_, landscapeExists := calibreOptions["pdf-landscape"]
		if orientation == "Landscape" && !landscapeExists {
			return false
		}
		if orientation == "Portrait" && landscapeExists {
			return false
		}
		
		return true
	}
	
	if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
		t.Errorf("Property test failed: %v", err)
	}
}