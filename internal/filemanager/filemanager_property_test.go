package filemanager

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// Custom generator for valid filename characters
// Includes spaces, alphanumeric, and common symbols, but excludes path separators and nulls
func filenameGen() gopter.Gen {
	specialChars := []interface{}{' ', '-', '_', '.', '!', '@', '#', '$', '%', '^', '&', '(', ')', '+', '=', '[', ']', '{', '}', ';', '\'', ',', '`', '~'}
	return gen.SliceOf(
		gen.Frequency(
			map[int]gopter.Gen{
				10: gen.AlphaNumChar(),
				1:  gen.OneConstOf(specialChars...),
			},
		),
	).Map(func(chars []rune) string {
		s := string(chars)
		if s == "" || s == "." || s == ".." {
			return "default_filename"
		}
		// Also trim spaces/dots from end if OS might dislike them? macOS is generally okay but for safety.
		// But the property says "Special Character Path Handling".
		return s
	})
}

// Property 8: Path Validation
// For any output path specified, validation must occur before starting conversion.
// Property 12: Special Character Path Handling
// For any valid macOS file path containing spaces or special characters, the tool must handle it correctly.
func TestProperty8_12_PathValidation(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	fm := NewFileManager()

	properties.Property("ValidateOutputPath accepts valid paths with special characters", prop.ForAll(
		func(fileName string) bool {
			// Setup temp dir
			tmpDir, err := os.MkdirTemp("", "fm-prop-test-*")
			if err != nil {
				return false
			}
			defer os.RemoveAll(tmpDir)

			// Construct path
			path := filepath.Join(tmpDir, fileName)

			// Validate
			err = fm.ValidateOutputPath(path)

			// If ValidateOutputPath returns error for a "valid" generated name,
			// log it.
			if err != nil {
				t.Logf("Validation failed for %q: %v", fileName, err)
				return false
			}
			return true
		},
		filenameGen(),
	))

	properties.TestingRun(t)
}
