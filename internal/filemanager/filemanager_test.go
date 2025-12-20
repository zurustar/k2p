package filemanager

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateOutputPath(t *testing.T) {
	fm := NewFileManager()

	t.Run("empty path", func(t *testing.T) {
		err := fm.ValidateOutputPath("")
		if err == nil {
			t.Error("expected error for empty path")
		}
	})

	t.Run("current directory", func(t *testing.T) {
		cwd, _ := os.Getwd()
		err := fm.ValidateOutputPath(cwd)
		if err != nil {
			t.Errorf("unexpected error for current directory: %v", err)
		}
	})

	t.Run("temp directory", func(t *testing.T) {
		err := fm.ValidateOutputPath(os.TempDir())
		if err != nil {
			t.Errorf("unexpected error for temp directory: %v", err)
		}
	})

	t.Run("non-existent parent", func(t *testing.T) {
		err := fm.ValidateOutputPath("/nonexistent/path/to/file.pdf")
		if err == nil {
			t.Error("expected error for non-existent parent directory")
		}
	})
}

func TestCheckDiskSpace(t *testing.T) {
	fm := NewFileManager()

	t.Run("small file", func(t *testing.T) {
		// Request 1KB, should always succeed
		err := fm.CheckDiskSpace(os.TempDir(), 1024)
		if err != nil {
			t.Errorf("unexpected error for small file: %v", err)
		}
	})

	t.Run("huge file", func(t *testing.T) {
		// Request 1PB, should fail
		err := fm.CheckDiskSpace(os.TempDir(), 1024*1024*1024*1024*1024)
		if err == nil {
			t.Error("expected error for huge file")
		}
	})
}

func TestCreateTempDir(t *testing.T) {
	fm := NewFileManager()

	dir, err := fm.CreateTempDir()
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(dir)

	// Verify directory exists
	info, err := os.Stat(dir)
	if err != nil {
		t.Errorf("temp directory does not exist: %v", err)
	}
	if !info.IsDir() {
		t.Error("created path is not a directory")
	}

	// Verify it's in temp directory
	if !strings.HasPrefix(dir, os.TempDir()) {
		t.Errorf("temp directory not in system temp: %s", dir)
	}
}

func TestCleanupTempDir(t *testing.T) {
	fm := NewFileManager()

	t.Run("cleanup valid temp dir", func(t *testing.T) {
		dir, err := fm.CreateTempDir()
		if err != nil {
			t.Fatalf("failed to create temp dir: %v", err)
		}

		// Create a file in the temp dir
		testFile := filepath.Join(dir, "test.txt")
		if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}

		// Cleanup
		if err := fm.CleanupTempDir(dir); err != nil {
			t.Errorf("failed to cleanup temp dir: %v", err)
		}

		// Verify directory is gone
		if _, err := os.Stat(dir); !os.IsNotExist(err) {
			t.Error("temp directory still exists after cleanup")
		}
	})

	t.Run("refuse to cleanup non-temp dir", func(t *testing.T) {
		err := fm.CleanupTempDir("/usr/local")
		if err == nil {
			t.Error("expected error when trying to cleanup non-temp directory")
		}
	})

	t.Run("empty path", func(t *testing.T) {
		err := fm.CleanupTempDir("")
		if err == nil {
			t.Error("expected error for empty path")
		}
	})
}

func TestResolveOutputPath(t *testing.T) {
	fm := NewFileManager()

	t.Run("empty output dir uses current directory", func(t *testing.T) {
		path, err := fm.ResolveOutputPath("")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if path == "" {
			t.Error("expected non-empty path")
		}
		// Should be in current directory
		cwd, _ := os.Getwd()
		if filepath.Dir(path) != cwd {
			t.Errorf("expected path in current directory, got: %s", path)
		}
	})

	t.Run("valid directory", func(t *testing.T) {
		path, err := fm.ResolveOutputPath(os.TempDir())
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if !strings.HasPrefix(path, os.TempDir()) {
			t.Errorf("expected path in temp directory, got: %s", path)
		}
		if filepath.Ext(path) != ".pdf" {
			t.Errorf("expected .pdf extension, got: %s", path)
		}
	})

	t.Run("non-existent directory", func(t *testing.T) {
		_, err := fm.ResolveOutputPath("/nonexistent/directory")
		if err == nil {
			t.Error("expected error for non-existent directory")
		}
	})
}

func TestHandleExistingFile(t *testing.T) {
	fm := NewFileManager()

	t.Run("non-existent file", func(t *testing.T) {
		proceed, err := fm.HandleExistingFile("/nonexistent/file.pdf", false)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if !proceed {
			t.Error("expected proceed=true for non-existent file")
		}
	})

	t.Run("existing file with auto-confirm", func(t *testing.T) {
		// Create a temp file
		tmpFile, err := os.CreateTemp("", "test-*.pdf")
		if err != nil {
			t.Fatalf("failed to create temp file: %v", err)
		}
		defer os.Remove(tmpFile.Name())
		tmpFile.Close()

		proceed, err := fm.HandleExistingFile(tmpFile.Name(), true)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if !proceed {
			t.Error("expected proceed=true with auto-confirm")
		}
	})
}
