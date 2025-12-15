# Design Document

## Overview

The Kindle to PDF converter is a Go-based command-line application that leverages Calibre's ebook-convert utility to transform Kindle format files (AZW, AZW3, MOBI) into PDF documents. The application provides a user-friendly CLI interface while handling the complexities of file validation, dependency checking, and conversion orchestration.

The tool follows a pipeline architecture where input validation, dependency verification, and conversion execution are handled as discrete stages. This design ensures robust error handling and allows for future extensibility.

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
│ (File operations, Calibre interface)│
└─────────────────────────────────────┘
                    │
┌─────────────────────────────────────┐
│        Infrastructure Layer         │
│  (File system, external processes)  │
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
  - Validate input files and formats
  - Execute Calibre conversion commands
  - Handle batch processing logic
  - Manage conversion progress and reporting
- **Interface**:
```go
type ConverterService interface {
    ConvertSingle(input, output string, options ConversionOptions) error
    ConvertBatch(inputDir, outputDir string, options ConversionOptions) (*BatchResult, error)
    ValidateFile(filepath string) error
}
```

### File Manager
- **Purpose**: Handle all file system operations
- **Key Functions**:
  - Validate file paths and permissions
  - Check disk space availability
  - Create output directories
  - Handle file naming conflicts
- **Interface**:
```go
type FileManager interface {
    ValidatePath(path string) error
    EnsureOutputDir(dir string) error
    CheckDiskSpace(path string, requiredBytes int64) error
    ResolveOutputPath(input, output string) (string, error)
}
```

### Calibre Interface
- **Purpose**: Manage interaction with Calibre's ebook-convert utility
- **Key Functions**:
  - Check Calibre installation
  - Execute conversion commands
  - Parse Calibre output and errors
- **Interface**:
```go
type CalibreInterface interface {
    IsInstalled() bool
    GetVersion() (string, error)
    Convert(input, output string, options map[string]string) error
}
```

## Data Models

### ConversionOptions
```go
type ConversionOptions struct {
    Quality       string            // PDF quality setting
    PageSize      string            // A4, Letter, etc.
    Orientation   string            // Portrait, Landscape
    CustomOptions map[string]string // Additional Calibre options
    Verbose       bool              // Enable detailed logging
    Overwrite     bool              // Overwrite existing files
}
```

### ConversionResult
```go
type ConversionResult struct {
    InputFile    string
    OutputFile   string
    Success      bool
    Error        error
    Duration     time.Duration
    FileSize     int64
}
```

### BatchResult
```go
type BatchResult struct {
    TotalFiles      int
    SuccessfulFiles int
    FailedFiles     int
    Results         []ConversionResult
    TotalDuration   time.Duration
}
```

### SupportedFormat
```go
type SupportedFormat struct {
    Extension   string
    Description string
    MimeType    string
}
```

## Correctness Properties

*A property is a characteristic or behavior that should hold true across all valid executions of a system-essentially, a formal statement about what the system should do. Properties serve as the bridge between human-readable specifications and machine-verifiable correctness guarantees.*

### Property 1: Successful conversion produces valid PDF
*For any* valid DRM-free Kindle file, conversion should produce a valid PDF file that can be opened and read
**Validates: Requirements 1.1**

### Property 2: Output directory specification is respected
*For any* valid output directory path, the converted PDF should be saved to that exact location
**Validates: Requirements 1.2**

### Property 3: Default output location consistency
*For any* source file when no output directory is specified, the PDF should be created in the same directory as the source file
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

### Property 9: Directory batch processing completeness
*For any* directory containing supported Kindle files, all supported files should be processed during batch conversion
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

### Property 17: Quality settings application
*For any* specified PDF quality setting, those settings should be applied during the conversion process
**Validates: Requirements 6.1**

### Property 18: Page configuration application
*For any* specified page size or orientation settings, the PDF output should be configured accordingly
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

## Error Handling

The application implements comprehensive error handling across multiple layers:

### Input Validation Errors
- Invalid file formats or extensions
- Non-existent input files or directories
- Insufficient file permissions
- DRM-protected content detection

### System Errors
- Missing Calibre dependency
- Insufficient disk space
- File system permission issues
- Process interruption handling

### Conversion Errors
- Calibre execution failures
- Corrupted input files
- Invalid conversion parameters
- Output file creation failures

### Error Recovery Strategies
- Graceful degradation for batch processing
- Automatic cleanup of partial files
- Detailed error reporting with actionable suggestions
- Preservation of original files during failures

## Testing Strategy

### Dual Testing Approach
The testing strategy combines unit testing and property-based testing to ensure comprehensive coverage:

- **Unit tests** verify specific examples, edge cases, and error conditions
- **Property-based tests** verify universal properties that should hold across all inputs
- Together they provide comprehensive coverage: unit tests catch concrete bugs, property tests verify general correctness

### Unit Testing
Unit tests will cover:
- CLI argument parsing with specific flag combinations
- File validation with known good and bad files
- Calibre interface with mocked responses
- Configuration file parsing with sample configurations
- Error message formatting and display

### Property-Based Testing
Property-based testing will use **Rapid** (Go's built-in property testing framework) to verify:
- Conversion properties across randomly generated valid inputs
- Path handling with generated file paths containing various characters
- Batch processing behavior with generated directory structures
- Configuration application with generated settings combinations
- Error handling with generated invalid inputs

**Requirements:**
- Each property-based test must run a minimum of 100 iterations
- Each property-based test must be tagged with a comment referencing the design document property
- Tag format: `**Feature: kindle-to-pdf-go, Property {number}: {property_text}**`
- Each correctness property must be implemented by a single property-based test
- Property-based tests should be placed close to implementation to catch errors early

### Integration Testing
- End-to-end conversion workflows with real Kindle files
- Calibre integration testing with actual ebook-convert utility
- File system integration testing on macOS
- CLI integration testing with various argument combinations

### Test Data Management
- Sample DRM-free Kindle files for testing
- Generated test files with various characteristics
- Mock Calibre responses for unit testing
- Configuration file templates for testing