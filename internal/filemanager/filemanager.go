package filemanager

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

// FileManager handles all file system operations
type FileManager interface {
	// ValidateOutputPath validates the output path and permissions
	ValidateOutputPath(path string) error

	// CheckDiskSpace checks if sufficient disk space is available
	CheckDiskSpace(path string, estimatedBytes int64) error

	// ResolveOutputPath resolves the output file path
	ResolveOutputPath(outputDir string) (string, error)

	// CreateTempDir creates a temporary directory for screenshots
	CreateTempDir() (string, error)

	// CleanupTempDir cleans up temporary files
	CleanupTempDir(dir string) error

	// HandleExistingFile checks if file exists and prompts for overwrite
	HandleExistingFile(path string, autoConfirm bool) (bool, error)
}

// DefaultFileManager is the default implementation of FileManager
type DefaultFileManager struct{}

// NewFileManager creates a new FileManager instance
func NewFileManager() FileManager {
	return &DefaultFileManager{}
}

// ValidateOutputPath validates the output path and permissions
func (fm *DefaultFileManager) ValidateOutputPath(path string) error {
	if path == "" {
		return errors.New("output path cannot be empty")
	}

	// Expand home directory if present
	if len(path) > 0 && path[0] == '~' {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		path = filepath.Join(home, path[1:])
	}

	// Clean the path
	path = filepath.Clean(path)

	// Check if path exists
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Path doesn't exist, check if parent directory exists and is writable
			parent := filepath.Dir(path)
			parentInfo, err := os.Stat(parent)
			if err != nil {
				return fmt.Errorf("parent directory does not exist: %s", parent)
			}
			if !parentInfo.IsDir() {
				return fmt.Errorf("parent path is not a directory: %s", parent)
			}
			// Check write permission on parent
			if err := checkWritePermission(parent); err != nil {
				return fmt.Errorf("no write permission for directory: %s", parent)
			}
			return nil
		}
		return fmt.Errorf("failed to stat path: %w", err)
	}

	// Path exists, check if it's a directory
	if info.IsDir() {
		// Check write permission
		if err := checkWritePermission(path); err != nil {
			return fmt.Errorf("no write permission for directory: %s", path)
		}
		return nil
	}

	// Path exists and is a file, check parent directory write permission
	parent := filepath.Dir(path)
	if err := checkWritePermission(parent); err != nil {
		return fmt.Errorf("no write permission for directory: %s", parent)
	}

	return nil
}

// checkWritePermission checks if the directory has write permission
func checkWritePermission(dir string) error {
	// Try to create a temporary file to test write permission
	testFile := filepath.Join(dir, ".k2p_write_test")
	f, err := os.Create(testFile)
	if err != nil {
		return err
	}
	f.Close()
	os.Remove(testFile)
	return nil
}

// CheckDiskSpace checks if sufficient disk space is available
func (fm *DefaultFileManager) CheckDiskSpace(path string, estimatedBytes int64) error {
	// Expand and clean path
	if len(path) > 0 && path[0] == '~' {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		path = filepath.Join(home, path[1:])
	}
	path = filepath.Clean(path)

	// Get the directory (if path is a file, use its parent)
	dir := path
	if info, err := os.Stat(path); err == nil && !info.IsDir() {
		dir = filepath.Dir(path)
	} else if os.IsNotExist(err) {
		dir = filepath.Dir(path)
	}

	// Get filesystem stats
	var stat syscall.Statfs_t
	if err := syscall.Statfs(dir, &stat); err != nil {
		return fmt.Errorf("failed to get filesystem stats: %w", err)
	}

	// Calculate available space
	availableBytes := stat.Bavail * uint64(stat.Bsize)

	if uint64(estimatedBytes) > availableBytes {
		return fmt.Errorf("insufficient disk space: need %d MB, only %d MB available",
			estimatedBytes/(1024*1024), availableBytes/(1024*1024))
	}

	return nil
}

// ResolveOutputPath resolves the output file path
func (fm *DefaultFileManager) ResolveOutputPath(outputDir string) (string, error) {
	// If no output directory specified, use current directory
	if outputDir == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("failed to get current directory: %w", err)
		}
		outputDir = cwd
	}

	// Expand home directory if present
	if len(outputDir) > 0 && outputDir[0] == '~' {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get home directory: %w", err)
		}
		outputDir = filepath.Join(home, outputDir[1:])
	}

	// Clean the path
	outputDir = filepath.Clean(outputDir)

	// Create directory if it doesn't exist
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %w", err)
	}

	// Verify it's a directory
	info, err := os.Stat(outputDir)
	if err != nil {
		return "", fmt.Errorf("failed to stat output directory: %w", err)
	}

	if !info.IsDir() {
		return "", fmt.Errorf("output path is not a directory: %s", outputDir)
	}

	// Generate output filename with timestamp
	filename := fmt.Sprintf("kindle_book_%s.pdf", generateTimestamp())
	outputPath := filepath.Join(outputDir, filename)

	return outputPath, nil
}

// generateTimestamp generates a timestamp string for filenames
func generateTimestamp() string {
	return time.Now().Format("20060102-150405")
}

// CreateTempDir creates a temporary directory for screenshots
func (fm *DefaultFileManager) CreateTempDir() (string, error) {
	tempDir, err := os.MkdirTemp("", "k2p-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary directory: %w", err)
	}
	return tempDir, nil
}

// CleanupTempDir cleans up temporary files
func (fm *DefaultFileManager) CleanupTempDir(dir string) error {
	if dir == "" {
		return errors.New("temporary directory path cannot be empty")
	}

	// Safety check: ensure it's a temp directory
	// Use explicit path prefix check
	cleanDir := filepath.Clean(dir)
	cleanTemp := filepath.Clean(os.TempDir())

	if !filepath.IsAbs(cleanDir) {
		absDir, err := filepath.Abs(cleanDir)
		if err == nil {
			cleanDir = absDir
		}
	}
	if !filepath.IsAbs(cleanTemp) {
		absTemp, err := filepath.Abs(cleanTemp)
		if err == nil {
			cleanTemp = absTemp
		}
	}

	// Simple prefix check isn't enough (e.g. /tmp/foo vs /tmp), need separator check
	rel, err := filepath.Rel(cleanTemp, cleanDir)
	if err != nil || strings.HasPrefix(rel, "..") {
		return fmt.Errorf("refusing to delete non-temporary directory: %s", dir)
	}

	if err := os.RemoveAll(dir); err != nil {
		return fmt.Errorf("failed to remove temporary directory: %w", err)
	}

	return nil
}

// HandleExistingFile checks if file exists and prompts for overwrite
func (fm *DefaultFileManager) HandleExistingFile(path string, autoConfirm bool) (bool, error) {
	// Check if file exists
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		// File doesn't exist, proceed
		return true, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to check file existence: %w", err)
	}

	// File exists
	if autoConfirm {
		// Auto-confirm overwrite
		return true, nil
	}

	// Prompt user for confirmation
	fmt.Printf("File already exists: %s\n", path)
	fmt.Print("Overwrite? [y/N]: ")

	var response string
	fmt.Scanln(&response)

	if response == "y" || response == "Y" {
		return true, nil
	}

	return false, nil
}
