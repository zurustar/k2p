package interfaces

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

// FileManager handles all file system operations
type FileManager interface {
	ValidatePath(path string) error
	EnsureOutputDir(dir string) error
	CheckDiskSpace(path string, requiredBytes int64) error
	ResolveOutputPath(input, output string) (string, error)
	DetectFileFormat(filepath string) (*SupportedFormat, error)
	IsDRMProtected(filepath string) (bool, error)
	CheckOutputFileConflict(outputPath string) (bool, error)
	ResolveOutputPathWithConflictHandling(input, output string, handler FileConflictHandler) (string, error)
}

// GetSupportedFormats returns the list of supported Kindle formats
func GetSupportedFormats() []SupportedFormat {
	return []SupportedFormat{
		{
			Extension:   ".azw",
			Description: "Amazon Kindle Format (AZW)",
			MimeType:    "application/vnd.amazon.ebook",
		},
		{
			Extension:   ".azw3",
			Description: "Amazon Kindle Format 8 (AZW3)",
			MimeType:    "application/vnd.amazon.ebook",
		},
		{
			Extension:   ".mobi",
			Description: "Mobipocket eBook (MOBI)",
			MimeType:    "application/x-mobipocket-ebook",
		},
	}
}

// DetectFileFormat detects the format of a Kindle file based on extension and content
func DetectFileFormat(filepath string) (*SupportedFormat, error) {
	if filepath == "" {
		return nil, errors.New("filepath cannot be empty")
	}

	// Check if file exists
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		return nil, fmt.Errorf("file does not exist: %s", filepath)
	}

	// Get file extension
	lastDotIndex := strings.LastIndex(filepath, ".")
	if lastDotIndex == -1 {
		return nil, fmt.Errorf("file has no extension: %s", filepath)
	}
	
	ext := strings.ToLower(filepath[lastDotIndex:])
	
	// Check against supported formats
	supportedFormats := GetSupportedFormats()
	for _, format := range supportedFormats {
		if format.Extension == ext {
			return &format, nil
		}
	}

	return nil, fmt.Errorf("unsupported file format: %s", ext)
}

// IsDRMProtected checks if a file is DRM protected
// This is a simplified implementation - in reality, this would require
// reading file headers and checking for DRM markers
func IsDRMProtected(filepath string) (bool, error) {
	if filepath == "" {
		return false, errors.New("filepath cannot be empty")
	}

	// Check if file exists
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		return false, fmt.Errorf("file does not exist: %s", filepath)
	}

	// Open file and read first few bytes to check for DRM markers
	file, err := os.Open(filepath)
	if err != nil {
		return false, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Read first 1024 bytes to check for DRM indicators
	buffer := make([]byte, 1024)
	n, err := file.Read(buffer)
	if err != nil && n == 0 {
		return false, fmt.Errorf("failed to read file: %w", err)
	}

	content := string(buffer[:n])
	
	// Check for specific DRM indicators in file content
	// These are more specific checks to avoid false positives
	drmIndicators := []string{
		"DRM_PROTECTED", "DRM PROTECTED", 
		"ENCRYPTED_CONTENT", "ENCRYPTED CONTENT",
		"AMAZON_DRM", "KINDLE_DRM",
		"TPZ", "tpz", // Amazon's DRM format
		"ADEPT", "adept", // Adobe DRM
	}

	for _, indicator := range drmIndicators {
		if strings.Contains(content, indicator) {
			return true, nil
		}
	}

	// Additional check: files that are unusually small might be DRM stubs
	fileInfo, err := file.Stat()
	if err == nil && fileInfo.Size() < 1024 {
		// Very small files are suspicious and might be DRM-protected stubs
		return true, nil
	}

	return false, nil
}

// ValidateKindleFile validates that a file is a supported Kindle format and not DRM protected
func ValidateKindleFile(filepath string) error {
	// Check file format
	format, err := DetectFileFormat(filepath)
	if err != nil {
		return fmt.Errorf("invalid file format: %w", err)
	}

	// Check for DRM protection
	isDRM, err := IsDRMProtected(filepath)
	if err != nil {
		return fmt.Errorf("failed to check DRM status: %w", err)
	}

	if isDRM {
		return fmt.Errorf("file appears to be DRM protected: %s (format: %s)", filepath, format.Description)
	}

	return nil
}

// IsValidKindleExtension checks if the file extension is supported
func IsValidKindleExtension(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	supportedFormats := GetSupportedFormats()
	
	for _, format := range supportedFormats {
		if format.Extension == ext {
			return true
		}
	}
	
	return false
}

// DefaultFileManager implements the FileManager interface
type DefaultFileManager struct{}

// NewFileManager creates a new DefaultFileManager instance
func NewFileManager() FileManager {
	return &DefaultFileManager{}
}

// ValidatePath validates and normalizes a file path
func (fm *DefaultFileManager) ValidatePath(path string) error {
	if path == "" {
		return errors.New("path cannot be empty")
	}

	// Clean and normalize the path
	cleanPath := filepath.Clean(path)
	
	// Convert to absolute path for validation
	absPath, err := filepath.Abs(cleanPath)
	if err != nil {
		return fmt.Errorf("failed to resolve absolute path: %w", err)
	}

	// Check if the path contains invalid characters for macOS
	// macOS doesn't allow certain characters in filenames
	invalidChars := []string{"\x00"} // Null character is not allowed
	for _, char := range invalidChars {
		if strings.Contains(absPath, char) {
			return fmt.Errorf("path contains invalid character: %s", path)
		}
	}

	// Check path length (macOS has a limit of 1024 characters for paths)
	if len(absPath) > 1024 {
		return fmt.Errorf("path too long (max 1024 characters): %s", path)
	}

	// Check if parent directory exists (for file paths)
	parentDir := filepath.Dir(absPath)
	if parentDir != "." && parentDir != "/" {
		if _, err := os.Stat(parentDir); os.IsNotExist(err) {
			return fmt.Errorf("parent directory does not exist: %s", parentDir)
		}
	}

	return nil
}

// EnsureOutputDir creates the output directory if it doesn't exist and checks permissions
func (fm *DefaultFileManager) EnsureOutputDir(dir string) error {
	if dir == "" {
		return errors.New("directory path cannot be empty")
	}

	// Clean and normalize the path
	cleanDir := filepath.Clean(dir)
	
	// Convert to absolute path
	absDir, err := filepath.Abs(cleanDir)
	if err != nil {
		return fmt.Errorf("failed to resolve absolute directory path: %w", err)
	}

	// Check if directory already exists
	if info, err := os.Stat(absDir); err == nil {
		// Directory exists, check if it's actually a directory
		if !info.IsDir() {
			return fmt.Errorf("path exists but is not a directory: %s", absDir)
		}
		
		// Check write permissions
		return fm.checkWritePermissions(absDir)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("failed to check directory status: %w", err)
	}

	// Directory doesn't exist, create it
	err = os.MkdirAll(absDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Verify the directory was created and is writable
	return fm.checkWritePermissions(absDir)
}

// checkWritePermissions checks if we have write permissions to a directory
func (fm *DefaultFileManager) checkWritePermissions(dir string) error {
	// Try to create a temporary file to test write permissions
	tempFile := filepath.Join(dir, ".kindle-converter-test")
	file, err := os.Create(tempFile)
	if err != nil {
		return fmt.Errorf("no write permission to directory: %s", dir)
	}
	file.Close()
	
	// Clean up the test file
	os.Remove(tempFile)
	return nil
}

// CheckDiskSpace checks if there's enough disk space available at the given path
func (fm *DefaultFileManager) CheckDiskSpace(path string, requiredBytes int64) error {
	if path == "" {
		return errors.New("path cannot be empty")
	}
	
	if requiredBytes < 0 {
		return errors.New("required bytes cannot be negative")
	}

	// Clean and normalize the path
	cleanPath := filepath.Clean(path)
	
	// Convert to absolute path
	absPath, err := filepath.Abs(cleanPath)
	if err != nil {
		return fmt.Errorf("failed to resolve absolute path: %w", err)
	}

	// Find the existing directory in the path hierarchy
	checkPath := absPath
	for {
		if info, err := os.Stat(checkPath); err == nil && info.IsDir() {
			break
		}
		parent := filepath.Dir(checkPath)
		if parent == checkPath {
			// Reached root without finding existing directory
			return fmt.Errorf("no existing directory found in path hierarchy: %s", absPath)
		}
		checkPath = parent
	}

	// Get disk usage statistics
	var stat syscall.Statfs_t
	err = syscall.Statfs(checkPath, &stat)
	if err != nil {
		return fmt.Errorf("failed to get disk space information: %w", err)
	}

	// Calculate available space
	availableBytes := int64(stat.Bavail) * int64(stat.Bsize)
	
	if availableBytes < requiredBytes {
		return fmt.Errorf("insufficient disk space: need %d bytes, have %d bytes available", requiredBytes, availableBytes)
	}

	return nil
}

// ResolveOutputPath resolves the output path based on input and output parameters
func (fm *DefaultFileManager) ResolveOutputPath(input, output string) (string, error) {
	if input == "" {
		return "", errors.New("input path cannot be empty")
	}

	// Clean and normalize input path
	cleanInput := filepath.Clean(input)
	absInput, err := filepath.Abs(cleanInput)
	if err != nil {
		return "", fmt.Errorf("failed to resolve absolute input path: %w", err)
	}

	// If no output specified, use same directory as input
	if output == "" {
		inputDir := filepath.Dir(absInput)
		inputBase := filepath.Base(absInput)
		
		// Remove extension and add .pdf
		inputExt := filepath.Ext(inputBase)
		baseName := strings.TrimSuffix(inputBase, inputExt)
		outputFile := baseName + ".pdf"
		
		return filepath.Join(inputDir, outputFile), nil
	}

	// Clean and normalize output path
	cleanOutput := filepath.Clean(output)
	absOutput, err := filepath.Abs(cleanOutput)
	if err != nil {
		return "", fmt.Errorf("failed to resolve absolute output path: %w", err)
	}

	// Check if output is a directory or file
	if info, err := os.Stat(absOutput); err == nil && info.IsDir() {
		// Output is an existing directory, generate filename
		inputBase := filepath.Base(absInput)
		inputExt := filepath.Ext(inputBase)
		baseName := strings.TrimSuffix(inputBase, inputExt)
		outputFile := baseName + ".pdf"
		
		return filepath.Join(absOutput, outputFile), nil
	}

	// Output is a file path (existing or not)
	// Ensure it has .pdf extension
	if !strings.HasSuffix(strings.ToLower(absOutput), ".pdf") {
		absOutput += ".pdf"
	}

	return absOutput, nil
}

// DetectFileFormat detects the format of a Kindle file
func (fm *DefaultFileManager) DetectFileFormat(filepath string) (*SupportedFormat, error) {
	return DetectFileFormat(filepath)
}

// IsDRMProtected checks if a file is DRM protected
func (fm *DefaultFileManager) IsDRMProtected(filepath string) (bool, error) {
	return IsDRMProtected(filepath)
}

// ConflictResolution represents how to handle file conflicts
type ConflictResolution int

const (
	ConflictAsk ConflictResolution = iota
	ConflictOverwrite
	ConflictSkip
	ConflictRename
)

// FileConflictHandler handles file conflicts when output files already exist
type FileConflictHandler interface {
	HandleConflict(outputPath string) (ConflictResolution, string, error)
}

// InteractiveConflictHandler prompts the user for conflict resolution
type InteractiveConflictHandler struct{}

// HandleConflict prompts the user to decide how to handle a file conflict
func (h *InteractiveConflictHandler) HandleConflict(outputPath string) (ConflictResolution, string, error) {
	fmt.Printf("File already exists: %s\n", outputPath)
	fmt.Print("Choose action: (o)verwrite, (s)kip, (r)ename, (q)uit: ")
	
	var response string
	_, err := fmt.Scanln(&response)
	if err != nil {
		return ConflictSkip, "", fmt.Errorf("failed to read user input: %w", err)
	}
	
	switch strings.ToLower(strings.TrimSpace(response)) {
	case "o", "overwrite":
		return ConflictOverwrite, outputPath, nil
	case "s", "skip":
		return ConflictSkip, "", nil
	case "r", "rename":
		newPath := generateAlternateName(outputPath)
		return ConflictRename, newPath, nil
	case "q", "quit":
		return ConflictSkip, "", fmt.Errorf("user requested to quit")
	default:
		fmt.Println("Invalid choice, skipping file...")
		return ConflictSkip, "", nil
	}
}

// AutoConflictHandler automatically resolves conflicts based on predefined strategy
type AutoConflictHandler struct {
	Strategy ConflictResolution
}

// HandleConflict automatically resolves conflicts based on the configured strategy
func (h *AutoConflictHandler) HandleConflict(outputPath string) (ConflictResolution, string, error) {
	switch h.Strategy {
	case ConflictOverwrite:
		return ConflictOverwrite, outputPath, nil
	case ConflictSkip:
		return ConflictSkip, "", nil
	case ConflictRename:
		newPath := generateAlternateName(outputPath)
		return ConflictRename, newPath, nil
	default:
		return ConflictSkip, "", nil
	}
}

// generateAlternateName creates an alternate filename when the original exists
func generateAlternateName(originalPath string) string {
	dir := filepath.Dir(originalPath)
	base := filepath.Base(originalPath)
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)
	
	counter := 1
	for {
		newName := fmt.Sprintf("%s_%d%s", name, counter, ext)
		newPath := filepath.Join(dir, newName)
		
		if _, err := os.Stat(newPath); os.IsNotExist(err) {
			return newPath
		}
		counter++
		
		// Prevent infinite loop
		if counter > 1000 {
			return filepath.Join(dir, fmt.Sprintf("%s_%d%s", name, counter, ext))
		}
	}
}

// ResolveOutputPathWithConflictHandling resolves output path and handles conflicts
func (fm *DefaultFileManager) ResolveOutputPathWithConflictHandling(input, output string, handler FileConflictHandler) (string, error) {
	// First resolve the basic output path
	outputPath, err := fm.ResolveOutputPath(input, output)
	if err != nil {
		return "", err
	}
	
	// Check if file already exists
	if _, err := os.Stat(outputPath); err == nil {
		// File exists, handle conflict
		if handler != nil {
			resolution, newPath, err := handler.HandleConflict(outputPath)
			if err != nil {
				return "", err
			}
			
			switch resolution {
			case ConflictOverwrite:
				return outputPath, nil
			case ConflictSkip:
				return "", fmt.Errorf("file skipped due to conflict: %s", outputPath)
			case ConflictRename:
				return newPath, nil
			default:
				return "", fmt.Errorf("unknown conflict resolution")
			}
		} else {
			// No handler provided, default to error
			return "", fmt.Errorf("output file already exists: %s", outputPath)
		}
	}
	
	return outputPath, nil
}

// CheckOutputFileConflict checks if an output file would conflict with existing files
func (fm *DefaultFileManager) CheckOutputFileConflict(outputPath string) (bool, error) {
	if outputPath == "" {
		return false, errors.New("output path cannot be empty")
	}
	
	_, err := os.Stat(outputPath)
	if err == nil {
		return true, nil // File exists, conflict detected
	}
	
	if os.IsNotExist(err) {
		return false, nil // File doesn't exist, no conflict
	}
	
	return false, fmt.Errorf("failed to check file existence: %w", err)
}