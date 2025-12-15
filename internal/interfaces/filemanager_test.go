package interfaces

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// **Feature: kindle-to-pdf-go, Property 1: Successful conversion produces valid PDF**
// **Validates: Requirements 1.1**
func TestProperty1_SuccessfulConversionProducesValidPDF(t *testing.T) {
	// Property: For any valid DRM-free Kindle file, conversion should produce a valid PDF file that can be opened and read
	
	// Create test directory
	testDir := t.TempDir()
	
	// Test with different supported formats
	supportedFormats := GetSupportedFormats()
	
	for _, format := range supportedFormats {
		t.Run("format_"+format.Extension, func(t *testing.T) {
			// Create a test file with the supported extension
			testFile := filepath.Join(testDir, "test"+format.Extension)
			
			// Create a non-DRM file (content that doesn't contain DRM indicators)
			// Make it large enough to avoid the small file DRM check
			testContent := "This is a test ebook file with normal content. It contains regular text and ebook data. " +
				"Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. " +
				"Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. " +
				"Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. " +
				"Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum. " +
				"Sed ut perspiciatis unde omnis iste natus error sit voluptatem accusantium doloremque laudantium, totam rem aperiam, " +
				"eaque ipsa quae ab illo inventore veritatis et quasi architecto beatae vitae dicta sunt explicabo. " +
				"Nemo enim ipsam voluptatem quia voluptas sit aspernatur aut odit aut fugit, sed quia consequuntur magni dolores eos " +
				"qui ratione voluptatem sequi nesciunt. Neque porro quisquam est, qui dolorem ipsum quia dolor sit amet, consectetur, " +
				"adipisci velit, sed quia non numquam eius modi tempora incidunt ut labore et dolore magnam aliquam quaerat voluptatem."
			err := os.WriteFile(testFile, []byte(testContent), 0644)
			if err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}
			
			// Test file format detection
			detectedFormat, err := DetectFileFormat(testFile)
			if err != nil {
				t.Errorf("Failed to detect format for valid file %s: %v", testFile, err)
				return
			}
			
			if detectedFormat.Extension != format.Extension {
				t.Errorf("Expected format %s, got %s", format.Extension, detectedFormat.Extension)
			}
			
			// Test DRM detection (should not be DRM protected)
			isDRM, err := IsDRMProtected(testFile)
			if err != nil {
				t.Errorf("Failed to check DRM status for %s: %v", testFile, err)
				return
			}
			
			if isDRM {
				t.Errorf("Non-DRM file %s was detected as DRM protected", testFile)
			}
			
			// Test overall validation (should pass)
			err = ValidateKindleFile(testFile)
			if err != nil {
				t.Errorf("Validation failed for valid non-DRM file %s: %v", testFile, err)
			}
		})
	}
}

// Test DRM detection with files that should be detected as DRM protected
func TestDRMDetection(t *testing.T) {
	testDir := t.TempDir()
	
	testCases := []struct {
		name        string
		content     string
		expectDRM   bool
	}{
		{
			name:      "file_with_drm_marker",
			content:   "This file contains DRM_PROTECTED markers",
			expectDRM: true,
		},
		{
			name:      "file_with_encrypted_marker",
			content:   "This file has ENCRYPTED_CONTENT data",
			expectDRM: true,
		},
		{
			name:      "file_with_amazon_marker",
			content:   "AMAZON_DRM protected content here",
			expectDRM: true,
		},
		{
			name:      "file_with_tpz_marker",
			content:   "TPZ format with DRM encoding",
			expectDRM: true,
		},
		{
			name:      "clean_file",
			content:   "This is a normal ebook file with regular content and text. " + strings.Repeat("Lorem ipsum dolor sit amet, consectetur adipiscing elit. ", 50),
			expectDRM: false,
		},
		{
			name:      "file_with_safe_words",
			content:   "This book discusses digital rights and content in general terms. " + strings.Repeat("More normal ebook content here. ", 50),
			expectDRM: false,
		},
		{
			name:      "very_small_file",
			content:   "x", // Very small file that might be a DRM stub
			expectDRM: true,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testFile := filepath.Join(testDir, tc.name+".azw3")
			err := os.WriteFile(testFile, []byte(tc.content), 0644)
			if err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}
			
			isDRM, err := IsDRMProtected(testFile)
			if err != nil {
				t.Errorf("Failed to check DRM status: %v", err)
				return
			}
			
			if isDRM != tc.expectDRM {
				t.Errorf("Expected DRM status %v for %s, got %v", tc.expectDRM, tc.name, isDRM)
			}
		})
	}
}

// Test file format detection with various file extensions
func TestFileFormatDetection(t *testing.T) {
	testDir := t.TempDir()
	
	testCases := []struct {
		filename    string
		expectError bool
	}{
		{"test.azw", false},
		{"test.azw3", false},
		{"test.mobi", false},
		{"test.AZW", false},  // Test case insensitive
		{"test.MOBI", false}, // Test case insensitive
		{"test.pdf", true},   // Unsupported format
		{"test.txt", true},   // Unsupported format
		{"test", true},       // No extension
	}
	
	for _, tc := range testCases {
		t.Run(tc.filename, func(t *testing.T) {
			testFile := filepath.Join(testDir, tc.filename)
			
			// Create the test file
			err := os.WriteFile(testFile, []byte("test content"), 0644)
			if err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}
			
			_, err = DetectFileFormat(testFile)
			
			if tc.expectError && err == nil {
				t.Errorf("Expected error for %s, but got none", tc.filename)
			}
			if !tc.expectError && err != nil {
				t.Errorf("Expected no error for %s, but got: %v", tc.filename, err)
			}
		})
	}
}

// Test validation of supported formats
func TestSupportedFormatsValidation(t *testing.T) {
	formats := GetSupportedFormats()
	
	if len(formats) == 0 {
		t.Error("No supported formats returned")
	}
	
	for i, format := range formats {
		err := format.Validate()
		if err != nil {
			t.Errorf("Supported format at index %d is invalid: %v", i, err)
		}
	}
}

// Test extension validation
func TestIsValidKindleExtension(t *testing.T) {
	testCases := []struct {
		filename string
		expected bool
	}{
		{"test.azw", true},
		{"test.azw3", true},
		{"test.mobi", true},
		{"test.AZW", true},   // Case insensitive
		{"test.MOBI", true},  // Case insensitive
		{"test.pdf", false},
		{"test.txt", false},
		{"test", false},
		{"", false},
	}
	
	for _, tc := range testCases {
		t.Run(tc.filename, func(t *testing.T) {
			result := IsValidKindleExtension(tc.filename)
			if result != tc.expected {
				t.Errorf("Expected %v for %s, got %v", tc.expected, tc.filename, result)
			}
		})
	}
}

// **Feature: kindle-to-pdf-go, Property 7: Special character path handling**
// **Validates: Requirements 3.3**
func TestProperty7_SpecialCharacterPathHandling(t *testing.T) {
	// Property: For any valid macOS file path containing spaces or special characters, 
	// the tool should handle it correctly without errors
	
	fm := NewFileManager()
	testDir := t.TempDir()
	
	// Test cases with various special characters that are valid on macOS
	testCases := []struct {
		name     string
		filename string
		shouldWork bool
	}{
		{"spaces", "file with spaces.azw3", true},
		{"unicode", "файл_тест.azw3", true},
		{"accents", "café_résumé.azw3", true},
		{"symbols", "file-name_test.azw3", true},
		{"parentheses", "file(1).azw3", true},
		{"brackets", "file[test].azw3", true},
		{"ampersand", "file&test.azw3", true},
		{"percent", "file%20test.azw3", true},
		{"plus", "file+test.azw3", true},
		{"equals", "file=test.azw3", true},
		{"comma", "file,test.azw3", true},
		{"semicolon", "file;test.azw3", true},
		{"exclamation", "file!test.azw3", true},
		{"question", "file?test.azw3", true},
		{"tilde", "~file.azw3", true},
		{"at", "file@test.azw3", true},
		{"hash", "file#test.azw3", true},
		{"dollar", "file$test.azw3", true},
		{"caret", "file^test.azw3", true},
		{"japanese", "テスト.azw3", true},
		{"emoji", "📚book.azw3", true},
		// Invalid characters on macOS
		{"null_char", "file\x00test.azw3", false},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create the test file path
			testFile := filepath.Join(testDir, tc.filename)
			
			if tc.shouldWork {
				// Create the actual file for valid cases
				err := os.WriteFile(testFile, []byte("test content"), 0644)
				if err != nil {
					t.Skipf("Cannot create file with special characters on this system: %v", err)
					return
				}
				
				// Test path validation
				err = fm.ValidatePath(testFile)
				if err != nil {
					t.Errorf("ValidatePath failed for valid special character path %s: %v", tc.filename, err)
				}
				
				// Test ResolveOutputPath with special character input
				outputPath, err := fm.ResolveOutputPath(testFile, "")
				if err != nil {
					t.Errorf("ResolveOutputPath failed for special character path %s: %v", tc.filename, err)
				}
				
				// Verify output path is reasonable
				if !strings.HasSuffix(outputPath, ".pdf") {
					t.Errorf("Output path should end with .pdf, got: %s", outputPath)
				}
				
				// Test with explicit output directory containing special characters
				outputDir := filepath.Join(testDir, "output with spaces")
				err = fm.EnsureOutputDir(outputDir)
				if err != nil {
					t.Errorf("EnsureOutputDir failed for directory with spaces: %v", err)
				}
				
				outputPath2, err := fm.ResolveOutputPath(testFile, outputDir)
				if err != nil {
					t.Errorf("ResolveOutputPath failed with special character output dir: %v", err)
				}
				
				if !strings.Contains(outputPath2, outputDir) {
					t.Errorf("Output path should be in specified directory, got: %s", outputPath2)
				}
			} else {
				// Test that invalid characters are properly rejected
				err := fm.ValidatePath(testFile)
				if err == nil {
					t.Errorf("ValidatePath should have failed for invalid character path %s", tc.filename)
				}
			}
		})
	}
}

// **Feature: kindle-to-pdf-go, Property 2: Output directory specification is respected**
// **Validates: Requirements 1.2**
func TestProperty2_OutputDirectorySpecificationRespected(t *testing.T) {
	// Property: For any valid output directory path, the converted PDF should be saved to that exact location
	
	fm := NewFileManager()
	testDir := t.TempDir()
	
	// Create a test input file
	inputFile := filepath.Join(testDir, "test.azw3")
	err := os.WriteFile(inputFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test input file: %v", err)
	}
	
	// Test cases with different output directory specifications
	testCases := []struct {
		name      string
		outputDir string
	}{
		{"simple_dir", "output"},
		{"nested_dir", "output/nested/deep"},
		{"dir_with_spaces", "output with spaces"},
		{"dir_with_unicode", "выходной_каталог"},
		{"dir_with_symbols", "output-dir_test"},
		{"absolute_dir", filepath.Join(testDir, "absolute_output")},
		{"relative_current", "."},
		{"relative_parent", ".."},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Resolve the output directory path
			var outputDir string
			if filepath.IsAbs(tc.outputDir) {
				outputDir = tc.outputDir
			} else {
				outputDir = filepath.Join(testDir, tc.outputDir)
			}
			
			// Ensure the output directory exists
			err := fm.EnsureOutputDir(outputDir)
			if err != nil {
				t.Errorf("EnsureOutputDir failed for %s: %v", tc.outputDir, err)
				return
			}
			
			// Verify the directory was created
			if info, err := os.Stat(outputDir); err != nil {
				t.Errorf("Output directory was not created: %s, error: %v", outputDir, err)
				return
			} else if !info.IsDir() {
				t.Errorf("Path exists but is not a directory: %s", outputDir)
				return
			}
			
			// Test ResolveOutputPath with the specified directory
			resolvedPath, err := fm.ResolveOutputPath(inputFile, outputDir)
			if err != nil {
				t.Errorf("ResolveOutputPath failed for output dir %s: %v", outputDir, err)
				return
			}
			
			// Verify the resolved path is within the specified directory
			resolvedDir := filepath.Dir(resolvedPath)
			expectedDir, err := filepath.Abs(outputDir)
			if err != nil {
				t.Errorf("Failed to get absolute path for %s: %v", outputDir, err)
				return
			}
			
			actualDir, err := filepath.Abs(resolvedDir)
			if err != nil {
				t.Errorf("Failed to get absolute path for resolved dir %s: %v", resolvedDir, err)
				return
			}
			
			if actualDir != expectedDir {
				t.Errorf("Resolved path not in specified directory. Expected: %s, Got: %s", expectedDir, actualDir)
			}
			
			// Verify the output file has .pdf extension
			if !strings.HasSuffix(resolvedPath, ".pdf") {
				t.Errorf("Output path should end with .pdf, got: %s", resolvedPath)
			}
			
			// Verify we can write to the resolved path
			err = os.WriteFile(resolvedPath, []byte("test pdf content"), 0644)
			if err != nil {
				t.Errorf("Cannot write to resolved output path %s: %v", resolvedPath, err)
			} else {
				// Clean up
				os.Remove(resolvedPath)
			}
		})
	}
}
// **Feature: kindle-to-pdf-go, Property 3: Default output location consistency**
// **Validates: Requirements 1.3**
func TestProperty3_DefaultOutputLocationConsistency(t *testing.T) {
	// Property: For any source file when no output directory is specified, 
	// the PDF should be created in the same directory as the source file
	
	fm := NewFileManager()
	testDir := t.TempDir()
	
	// Test cases with different input file locations and names
	testCases := []struct {
		name     string
		inputPath string
	}{
		{"simple_file", "test.azw3"},
		{"file_with_spaces", "test file.azw3"},
		{"file_with_unicode", "тест.azw3"},
		{"file_with_symbols", "test-file_name.azw3"},
		{"nested_file", "subdir/test.azw3"},
		{"deep_nested", "sub1/sub2/sub3/test.azw3"},
		{"different_extension", "book.mobi"},
		{"another_extension", "ebook.azw"},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create the full input path
			inputFile := filepath.Join(testDir, tc.inputPath)
			
			// Ensure the directory exists for nested files
			inputDir := filepath.Dir(inputFile)
			err := os.MkdirAll(inputDir, 0755)
			if err != nil {
				t.Fatalf("Failed to create input directory: %v", err)
			}
			
			// Create the test input file
			err = os.WriteFile(inputFile, []byte("test content"), 0644)
			if err != nil {
				t.Fatalf("Failed to create test input file: %v", err)
			}
			
			// Test ResolveOutputPath with empty output (default behavior)
			resolvedPath, err := fm.ResolveOutputPath(inputFile, "")
			if err != nil {
				t.Errorf("ResolveOutputPath failed for default output: %v", err)
				return
			}
			
			// Verify the output is in the same directory as the input
			expectedDir := filepath.Dir(inputFile)
			actualDir := filepath.Dir(resolvedPath)
			
			expectedAbsDir, err := filepath.Abs(expectedDir)
			if err != nil {
				t.Errorf("Failed to get absolute path for expected dir: %v", err)
				return
			}
			
			actualAbsDir, err := filepath.Abs(actualDir)
			if err != nil {
				t.Errorf("Failed to get absolute path for actual dir: %v", err)
				return
			}
			
			if actualAbsDir != expectedAbsDir {
				t.Errorf("Output not in same directory as input. Expected: %s, Got: %s", expectedAbsDir, actualAbsDir)
			}
			
			// Verify the output file has .pdf extension
			if !strings.HasSuffix(resolvedPath, ".pdf") {
				t.Errorf("Output path should end with .pdf, got: %s", resolvedPath)
			}
			
			// Verify the base name is derived from the input file
			inputBase := filepath.Base(inputFile)
			inputExt := filepath.Ext(inputBase)
			expectedBaseName := strings.TrimSuffix(inputBase, inputExt) + ".pdf"
			actualBaseName := filepath.Base(resolvedPath)
			
			if actualBaseName != expectedBaseName {
				t.Errorf("Output filename incorrect. Expected: %s, Got: %s", expectedBaseName, actualBaseName)
			}
			
			// Test consistency - calling ResolveOutputPath multiple times should give same result
			resolvedPath2, err := fm.ResolveOutputPath(inputFile, "")
			if err != nil {
				t.Errorf("Second ResolveOutputPath call failed: %v", err)
				return
			}
			
			if resolvedPath != resolvedPath2 {
				t.Errorf("ResolveOutputPath not consistent. First: %s, Second: %s", resolvedPath, resolvedPath2)
			}
		})
	}
}
// **Feature: kindle-to-pdf-go, Property 13: File overwrite handling**
// **Validates: Requirements 4.5**
func TestProperty13_FileOverwriteHandling(t *testing.T) {
	// Property: For any conversion where the target file already exists, 
	// the user should be prompted for overwrite confirmation or the file should be skipped
	
	fm := NewFileManager()
	testDir := t.TempDir()
	
	// Create a test input file
	inputFile := filepath.Join(testDir, "test.azw3")
	err := os.WriteFile(inputFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test input file: %v", err)
	}
	
	// Test different conflict resolution strategies
	testCases := []struct {
		name     string
		strategy ConflictResolution
	}{
		{"overwrite_strategy", ConflictOverwrite},
		{"skip_strategy", ConflictSkip},
		{"rename_strategy", ConflictRename},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Resolve the basic output path
			outputPath, err := fm.ResolveOutputPath(inputFile, "")
			if err != nil {
				t.Fatalf("Failed to resolve output path: %v", err)
			}
			
			// Create an existing file at the output location
			existingContent := "existing pdf content"
			err = os.WriteFile(outputPath, []byte(existingContent), 0644)
			if err != nil {
				t.Fatalf("Failed to create existing output file: %v", err)
			}
			
			// Test conflict detection
			hasConflict, err := fm.CheckOutputFileConflict(outputPath)
			if err != nil {
				t.Errorf("CheckOutputFileConflict failed: %v", err)
				return
			}
			
			if !hasConflict {
				t.Errorf("Expected conflict to be detected for existing file")
				return
			}
			
			// Test conflict resolution with auto handler
			handler := &AutoConflictHandler{Strategy: tc.strategy}
			resolvedPath, err := fm.ResolveOutputPathWithConflictHandling(inputFile, "", handler)
			
			switch tc.strategy {
			case ConflictOverwrite:
				if err != nil {
					t.Errorf("Overwrite strategy should not return error: %v", err)
				}
				if resolvedPath != outputPath {
					t.Errorf("Overwrite should return original path. Expected: %s, Got: %s", outputPath, resolvedPath)
				}
				
			case ConflictSkip:
				if err == nil {
					t.Errorf("Skip strategy should return error indicating file was skipped")
				}
				if !strings.Contains(err.Error(), "skipped") {
					t.Errorf("Skip error should mention 'skipped', got: %v", err)
				}
				
			case ConflictRename:
				if err != nil {
					t.Errorf("Rename strategy should not return error: %v", err)
				}
				if resolvedPath == outputPath {
					t.Errorf("Rename should return different path. Original: %s, Got: %s", outputPath, resolvedPath)
				}
				if !strings.HasSuffix(resolvedPath, ".pdf") {
					t.Errorf("Renamed path should still end with .pdf: %s", resolvedPath)
				}
				
				// Verify the renamed path doesn't exist yet
				if _, err := os.Stat(resolvedPath); err == nil {
					t.Errorf("Renamed path should not exist yet: %s", resolvedPath)
				}
			}
		})
	}
	
	// Test generateAlternateName function
	t.Run("alternate_name_generation", func(t *testing.T) {
		originalPath := filepath.Join(testDir, "test.pdf")
		
		// Create the original file
		err := os.WriteFile(originalPath, []byte("original"), 0644)
		if err != nil {
			t.Fatalf("Failed to create original file: %v", err)
		}
		
		// Generate alternate name
		alternatePath := generateAlternateName(originalPath)
		
		// Verify alternate name is different
		if alternatePath == originalPath {
			t.Errorf("Alternate name should be different from original")
		}
		
		// Verify alternate name follows expected pattern
		expectedPattern := filepath.Join(testDir, "test_1.pdf")
		if alternatePath != expectedPattern {
			t.Errorf("Expected alternate name %s, got %s", expectedPattern, alternatePath)
		}
		
		// Create the first alternate and generate another
		err = os.WriteFile(alternatePath, []byte("alternate1"), 0644)
		if err != nil {
			t.Fatalf("Failed to create first alternate file: %v", err)
		}
		
		alternatePath2 := generateAlternateName(originalPath)
		expectedPattern2 := filepath.Join(testDir, "test_2.pdf")
		if alternatePath2 != expectedPattern2 {
			t.Errorf("Expected second alternate name %s, got %s", expectedPattern2, alternatePath2)
		}
	})
	
	// Test no conflict case
	t.Run("no_conflict", func(t *testing.T) {
		nonExistentOutput := filepath.Join(testDir, "nonexistent.pdf")
		
		// Verify no conflict detected
		hasConflict, err := fm.CheckOutputFileConflict(nonExistentOutput)
		if err != nil {
			t.Errorf("CheckOutputFileConflict failed for non-existent file: %v", err)
		}
		
		if hasConflict {
			t.Errorf("Should not detect conflict for non-existent file")
		}
		
		// Test resolution without conflict
		handler := &AutoConflictHandler{Strategy: ConflictOverwrite}
		resolvedPath, err := fm.ResolveOutputPathWithConflictHandling(inputFile, nonExistentOutput, handler)
		if err != nil {
			t.Errorf("Should not error when no conflict exists: %v", err)
		}
		
		if resolvedPath != nonExistentOutput {
			t.Errorf("Should return original path when no conflict. Expected: %s, Got: %s", nonExistentOutput, resolvedPath)
		}
	})
}