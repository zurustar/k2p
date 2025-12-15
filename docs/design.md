# Design Document

## Overview

The Kindle to PDF converter is a Go-based command-line application that automates page turning and screenshot capture of the currently open book in the macOS Kindle app, then combines the screenshots into PDF documents. The application provides a user-friendly CLI interface while handling the complexities of page automation, screenshot capture, and PDF generation.

The tool follows a simple pipeline architecture where the user manually opens a book, then the app automatically handles page turning, screenshot capture, and PDF assembly. This design ensures robust error handling while keeping the user workflow simple.

## Architecture

The application follows a layered architecture with clear separation of concerns:

```
┌─────────────────────────────────────┐
│           CLI Layer                 │
│  (Command parsing, user interface)  │
└─────────────────────────────────────┘
                    │
┌─────────────────────────────────────┐
│        Application Layer            │
│   (Business logic, orchestration)   │
└─────────────────────────────────────┘
                    │
┌─────────────────────────────────────┐
│         Service Layer               │
│(Automation, Screenshot, PDF generation)│
└─────────────────────────────────────┘
                    │
┌─────────────────────────────────────┐
│        Infrastructure Layer         │
│ (macOS APIs, Kindle app, File system)│
└─────────────────────────────────────┘
```

## Components and Interfaces

### CLI Component
- **Purpose**: Handle command-line argument parsing and user interaction
- **Key Functions**: 
  - Parse command-line flags and arguments
  - Display help and version information
  - Validate user input and provide feedback
- **Dependencies**: Standard library `flag` package and `cobra` CLI framework

### Converter Service
- **Purpose**: Orchestrate the conversion process
- **Key Functions**:
  - Check if a book is currently open in Kindle app
  - Coordinate page turning and screenshot capture
  - Handle conversion progress and reporting
  - Manage PDF generation from captured screenshots
- **Interface**:
```go
type ConverterService interface {
    ConvertCurrentBook(output string, options ConversionOptions) error
    ValidateKindleAppState() error
    GetBookInfo() (*BookInfo, error)
}
```

### Kindle Automation Interface
- **Purpose**: Manage interaction with the macOS Kindle application
- **Key Functions**:
  - Check Kindle app installation and status
  - Detect if a book is currently open
  - Control page turning navigation
  - Handle app state detection and focus management
- **Interface**:
```go
type KindleAutomation interface {
    IsKindleInstalled() bool
    IsBookOpen() (bool, error)
    IsKindleInForeground() (bool, error)
    BringKindleToForeground() error
    GetBookTitle() (string, error)
    GetCurrentPageNumber() (int, error)
    HasNextPage() (bool, error)
    NextPage() error
    IsLastPage() (bool, error)
}
```

### Screenshot Service
- **Purpose**: Handle screenshot capture and image processing
- **Key Functions**:
  - Capture screenshots of Kindle app content
  - Process and optimize captured images
  - Handle image format conversion
- **Interface**:
```go
type ScreenshotService interface {
    CaptureKindlePage() (*image.Image, error)
    SaveScreenshot(img *image.Image, path string) error
    OptimizeImage(img *image.Image, quality int) (*image.Image, error)
}
```

### PDF Generator
- **Purpose**: Generate PDF documents from captured screenshots
- **Key Functions**:
  - Combine multiple images into a single PDF
  - Apply PDF metadata and properties
  - Optimize PDF file size and quality
- **Interface**:
```go
type PDFGenerator interface {
    CreatePDF(images []string, outputPath string, options PDFOptions) error
    AddMetadata(pdf *PDF, title, author string) error
    OptimizePDF(inputPath, outputPath string) error
}
```

### File Manager
- **Purpose**: Handle all file system operations
- **Key Functions**:
  - Validate file paths and permissions
  - Check disk space availability
  - Create output directories
  - Handle file naming conflicts
  - Manage temporary screenshot files
- **Interface**:
```go
type FileManager interface {
    ValidatePath(path string) error
    EnsureOutputDir(dir string) error
    CheckDiskSpace(path string, requiredBytes int64) error
    ResolveOutputPath(bookID, output string) (string, error)
    CreateTempDir() (string, error)
    CleanupTempFiles(dir string) error
}
```

## Data Models

### ConversionOptions
```go
type ConversionOptions struct {
    ScreenshotQuality int               // Screenshot quality (1-100)
    PageDelay         time.Duration     // Delay between page turns
    StartupDelay      time.Duration     // Delay before starting automation
    PDFQuality        string            // PDF compression quality
    CustomOptions     map[string]string // Additional options
    Verbose           bool              // Enable detailed logging
    Overwrite         bool              // Overwrite existing files
    TempDir           string            // Temporary directory for screenshots
    AutoConfirm       bool              // Skip user confirmation prompts
}
```

### ConversionResult
```go
type ConversionResult struct {
    BookTitle     string
    OutputFile    string
    Success       bool
    Error         error
    Duration      time.Duration
    PagesCaptured int
    FileSize      int64
}
```

### BookInfo
```go
type BookInfo struct {
    Title       string
    Author      string
    CurrentPage int
    IsOpen      bool
}
```

### PDFOptions
```go
type PDFOptions struct {
    Title       string
    Author      string
    Subject     string
    Quality     string // low, medium, high
    Compression bool
}
```

### ScreenshotInfo
```go
type ScreenshotInfo struct {
    PageNumber int
    FilePath   string
    Timestamp  time.Time
    Width      int
    Height     int
}
```

## Correctness Properties

*A property is a characteristic or behavior that should hold true across all valid executions of a system-essentially, a formal statement about what the system should do. Properties serve as the bridge between human-readable specifications and machine-verifiable correctness guarantees.*

### Property 1: Successful conversion produces valid PDF
*For any* currently open book in Kindle app, conversion should produce a valid PDF file that can be opened and read
**Validates: Requirements 1.1**

### Property 2: Output directory specification is respected
*For any* valid output directory path, the converted PDF should be saved to that exact location
**Validates: Requirements 1.2**

### Property 3: Default output location consistency
*For any* book when no output directory is specified, the PDF should be created in the current working directory
**Validates: Requirements 1.3**

### Property 4: Success message consistency
*For any* successful conversion, a success message containing the output file path should be displayed
**Validates: Requirements 1.4**

### Property 5: Path validation consistency
*For any* combination of input and output paths, validation should occur before conversion starts and provide consistent results
**Validates: Requirements 2.3**

### Property 6: Invalid argument error handling
*For any* invalid command-line arguments, clear error messages and usage examples should be displayed
**Validates: Requirements 2.4**

### Property 7: Special character path handling
*For any* valid macOS file path containing spaces or special characters, the tool should handle it correctly without errors
**Validates: Requirements 3.3**

### Property 8: Progress indicator display
*For any* conversion process, progress indicators should be displayed during the conversion
**Validates: Requirements 3.5**

### Property 9: Sequential processing completeness
*For any* currently open book, the conversion should process all pages from current position to the end
**Validates: Requirements 4.1**

### Property 10: Batch processing resilience
*For any* batch containing both valid and invalid files, processing should continue for remaining files when one conversion fails
**Validates: Requirements 4.2**

### Property 11: Batch summary reporting
*For any* completed batch processing operation, a summary of successful and failed conversions should be provided
**Validates: Requirements 4.3**

### Property 12: Output naming consistency
*For any* batch of input files, output file names should follow consistent naming conventions
**Validates: Requirements 4.4**

### Property 13: File overwrite handling
*For any* conversion where the target file already exists, the user should be prompted for overwrite confirmation or the file should be skipped
**Validates: Requirements 4.5**

### Property 14: Source file preservation
*For any* conversion failure, the original source file should remain unchanged and unmodified
**Validates: Requirements 5.3**

### Property 15: Cleanup on interruption
*For any* interrupted conversion process, partial output files should be cleaned up automatically
**Validates: Requirements 5.4**

### Property 16: Verbose logging consistency
*For any* conversion when verbose logging is enabled, detailed progress and debugging information should be provided
**Validates: Requirements 5.5**

### Property 17: Screenshot quality settings application
*For any* specified screenshot quality setting, those settings should be applied during the page capture process
**Validates: Requirements 6.1**

### Property 18: Page delay configuration application
*For any* specified page delay setting, the system should wait the specified time between page turns during capture
**Validates: Requirements 6.2**

### Property 19: Configuration file processing
*For any* valid configuration file, the settings should be read and applied during conversion
**Validates: Requirements 6.3**

### Property 20: Invalid configuration fallback
*For any* invalid configuration values, default values should be used and the user should be warned
**Validates: Requirements 6.4**

### Property 21: Default configuration consistency
*For any* conversion when no configuration is specified, sensible default settings for high-quality PDF output should be used
**Validates: Requirements 6.5**

### Property 22: User preparation time consistency
*For any* conversion process, adequate preparation time should be provided for the user to set up the Kindle app
**Validates: Requirements 7.2**

### Property 23: App focus validation
*For any* automation attempt, the system should verify that the Kindle app is in the foreground before proceeding
**Validates: Requirements 7.3**

## Error Handling

The application implements comprehensive error handling across multiple layers:

### Input Validation Errors
- No book currently open in Kindle app
- Insufficient file permissions for output directory
- Invalid configuration parameters
- Invalid output path specifications

### System Errors
- Missing Kindle app installation
- Kindle app not responding or crashed
- Insufficient disk space for screenshots and PDF
- macOS automation permission issues
- Process interruption handling

### Automation Errors
- Kindle app automation failures
- Screenshot capture failures
- Page navigation errors
- App state synchronization issues
- App focus and foreground detection failures
- User interaction timeout errors

### Conversion Errors
- PDF generation failures
- Image processing errors
- Invalid screenshot data
- Output file creation failures

### Error Recovery Strategies
- Graceful degradation for batch processing
- Automatic cleanup of temporary screenshot files
- Retry mechanisms for transient automation failures
- Detailed error reporting with actionable suggestions
- Preservation of Kindle app state during failures

## Testing Strategy

### Dual Testing Approach
The testing strategy combines unit testing and property-based testing to ensure comprehensive coverage:

- **Unit tests** verify specific examples, edge cases, and error conditions
- **Property-based tests** verify universal properties that should hold across all inputs
- Together they provide comprehensive coverage: unit tests catch concrete bugs, property tests verify general correctness

### Unit Testing
Unit tests will cover:
- CLI argument parsing with specific flag combinations
- Book identifier validation with various formats
- Kindle automation interface with mocked responses
- Screenshot service with mock image data
- PDF generation with test images
- Configuration file parsing with sample configurations
- Error message formatting and display

### Property-Based Testing
Property-based testing will use **Rapid** (Go's built-in property testing framework) to verify:
- Conversion properties across randomly generated valid book identifiers
- Path handling with generated file paths containing various characters
- Batch processing behavior with generated book lists
- Configuration application with generated settings combinations
- Error handling with generated invalid inputs

**Requirements:**
- Each property-based test must run a minimum of 100 iterations
- Each property-based test must be tagged with a comment referencing the design document property
- Tag format: `**Feature: kindle-to-pdf-go, Property {number}: {property_text}**`
- Each correctness property must be implemented by a single property-based test
- Property-based tests should be placed close to implementation to catch errors early

### Integration Testing
- End-to-end conversion workflows with real Kindle books
- Kindle app automation testing with actual app instances
- Screenshot capture testing on macOS
- PDF generation testing with real image sequences
- CLI integration testing with various argument combinations

### Test Data Management
- Sample Kindle books available in test library
- Generated test images simulating screenshots
- Mock Kindle app responses for unit testing
- Configuration file templates for testing
- Temporary directory management for test artifacts