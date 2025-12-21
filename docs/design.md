# Design Document

## Overview

The Kindle to PDF converter is a Go-based command-line application for macOS that automates the conversion of a currently open Kindle book to PDF format. The user manually opens a book in the macOS Kindle app, then runs this tool to automatically capture screenshots of each page and combine them into a single PDF document.

**Key Design Principle**: The tool processes **one book at a time** - the book that is currently open in the Kindle app. The user must manually open each book they want to convert.

## Architecture

The application follows a simple layered architecture:

```
┌─────────────────────────────────────┐
│           CLI Layer                 │
│  (Command parsing, user interface)  │
└─────────────────────────────────────┘
                    │
┌─────────────────────────────────────┐
│      Conversion Orchestrator        │
│   (Workflow coordination, state)    │
└─────────────────────────────────────┘
                    │
        ┌───────────┴───────────┐
        │                       │
┌───────────────┐      ┌────────────────┐
│   Automation  │      │  PDF Generator │
│    Service    │      │    Service     │
└───────────────┘      └────────────────┘
        │                       │
┌───────────────────────────────────────┐
│        Infrastructure Layer           │
│ (macOS APIs, File system, Kindle app) │
└───────────────────────────────────────┘
```

## Components and Interfaces

### CLI Component
**Purpose**: Handle command-line argument parsing and user interaction

**Responsibilities**:
- Parse command-line flags (output path, quality settings, delays, verbose mode, version, help)
- Provide opt-in flags for border trimming and page turn direction overrides
- Display help and version information
- Validate user input before starting conversion
- Show clear error messages for invalid arguments

**Dependencies**: Standard library `flag` package or `cobra` CLI framework

### Conversion Orchestrator
**Purpose**: Coordinate the entire conversion workflow for the currently open book

**Interface**:
```go
type ConversionOrchestrator interface {
    // Convert the currently open book to PDF
    ConvertCurrentBook(ctx context.Context, options ConversionOptions) (*ConversionResult, error)
}
```

**Responsibilities**:
1. Display preparation instructions to user
2. Wait for user confirmation before starting
3. Apply startup delay (with countdown timer if configured)
4. Validate Kindle app state (installed, book open, in foreground)
5. Check disk space availability
6. Create temporary directory for screenshots
7. Coordinate page capture loop (turn page → capture → repeat until end)
8. Generate PDF from captured screenshots
9. Clean up temporary files
10. Display success message with output path
11. Handle errors and cleanup on failure or interruption

### Kindle Automation Service
**Purpose**: Interact with the macOS Kindle application using macOS automation APIs

**Interface**:
```go
type KindleAutomation interface {
    // Check if Kindle app is installed
    IsKindleInstalled() (bool, error)
    
    // Check if a book is currently open
    IsBookOpen() (bool, error)
    
    // Check if Kindle app is in foreground
    IsKindleInForeground() (bool, error)
    
    // Bring Kindle app to foreground
    BringKindleToForeground() error
    
    // Turn to next page
    TurnNextPage() error
}
```

**Implementation Notes**:
- Use macOS AppleScript or Accessibility APIs for automation
- Implement retry logic for transient failures
- Detect end-of-book condition reliably

### PDF Generator Service
**Purpose**: Generate PDF documents from captured page screenshots

**Interface**:
```go
type PDFGenerator interface {
    // Create PDF from a sequence of image files
    CreatePDF(ctx context.Context, imageFiles []string, outputPath string, options PDFOptions) error
}
```

**Responsibilities**:
- Combine multiple images into a single PDF
- Apply quality/compression settings
- Handle large numbers of pages efficiently
- Validate output PDF is readable

**Implementation**: Use Go PDF library (e.g., `gofpdf`, `pdfcpu`)

### File Manager
**Purpose**: Handle all file system operations

**Interface**:
```go
type FileManager interface {
    // Validate output path and permissions
    ValidateOutputPath(path string) error
    
    // Check if sufficient disk space is available
    CheckDiskSpace(path string, estimatedBytes int64) error
    
    // Resolve output file path (handle default directory, existing files)
    ResolveOutputPath(outputDir string) (string, error)
    
    // Create temporary directory for screenshots
    CreateTempDir() (string, error)
    
    // Clean up temporary files
    CleanupTempDir(dir string) error
    
    // Check if file exists and prompt for overwrite
    HandleExistingFile(path string, autoConfirm bool) (bool, error)
}
```

**Responsibilities**:
- Validate file paths (including special characters and spaces)
- Manage temporary screenshot storage
- Handle file naming conflicts
- Ensure proper cleanup on success, failure, or interruption


## Data Models

### ConversionOptions
```go
type ConversionOptions struct {
    // Output directory (empty = current directory)
    OutputDir string
    
    // Screenshot quality (1-100, default: 95)
    ScreenshotQuality int
    
    // Delay between page turns (default: 500ms)
    PageDelay time.Duration
    
    // Delay before starting automation (default: 3s)
    StartupDelay time.Duration
    
    // Show countdown timer during startup delay
    ShowCountdown bool
    
    // PDF quality setting (low/medium/high, default: high)
    PDFQuality string
    
    // Enable verbose logging
    Verbose bool
    
    // Auto-confirm overwrite without prompting
    AutoConfirm bool

    // Operation mode: "detect" (analyze margins) or "generate" (create PDF)
    // Default: "generate"
    Mode string

    // Custom trim margins in pixels (default: 0 = no trimming)
    // Trimming is applied if any value is non-zero
    // 0 means no trimming for that specific edge
    // Example: TrimHorizontal=30, TrimTop=0, TrimBottom=0 trims 30px from both left and right
    TrimTop        int
    TrimBottom     int
    TrimHorizontal int

    // Page turn key: "right" or "left" (auto-detects unless forced to left)

    PageTurnKey string
}

func (o *ConversionOptions) Validate() error {
    // Validates quality range (1-100)
    // Validates PDF quality enum
    // Validates other constraints
}

```

### ConversionResult
```go
type ConversionResult struct {
    // Path to generated PDF
    OutputPath string
    
    // Number of pages captured
    PageCount int
    
    // Total conversion duration
    Duration time.Duration
    
    // Output file size in bytes
    FileSize int64
    
    // Any warnings encountered
    Warnings []string
}
```

### PDFOptions
```go
type PDFOptions struct {
    // Quality setting
    Quality string // "low", "medium", "high"
    
    // Enable compression
    Compression bool
}
```

## Workflow

### Main Conversion Flow

1. **Initialization**
   - Parse CLI arguments
   - Apply defaults
   - Validate options

2. **Pre-flight Checks**
   - Check if Kindle app is installed (error with installation instructions if not)
   - Check if a book is currently open (error with instructions if not)
   - Validate output path and permissions
   - Estimate disk space needed and check availability
   - Resolve output file path and handle existing file conflicts

3. **User Preparation**
   - Display instructions: "Please ensure Kindle app is in foreground and ready"
   - Wait for user confirmation (press Enter to continue)
   - Apply startup delay with countdown timer (if configured)
   - Verify Kindle app is in foreground (bring to front if needed)
   - Auto-detect page turn direction unless user forces left arrow key

4. **Page Capture Loop**
   ```
   Create temporary directory
   direction = detectPageTurnDirection() // uses sample captures unless user forced "left"

   // Activate Kindle once; keep it foregrounded for faster capture
   activateKindleAndDiscardProbeCapture()

   pageNumber = 1

   while pageNumber <= maxPages:
       Display progress: "Capturing page {pageNumber}..."
       screenshot = CaptureWithoutActivationWithRetry()
       Save screenshot to temp directory (trim if enabled)

       if lastFiveScreenshotsAreIdentical():
           removeRatingScreensAndStop()
           break

       TurnNextPage(direction) with retry
       Wait for PageDelay (default 500ms) to let page settle
       pageNumber++
   ```

5. **PDF Generation**
   - Display: "Generating PDF from {pageCount} pages..."
   - Create PDF from all captured screenshots
   - Apply quality and compression settings
   - Save to output path

6. **Cleanup and Completion**
   - Delete temporary screenshot files
   - Wait for macOS screen recording indicator to clear
   - Display success message with output path, page count, and file size
   - Exit with status 0

7. **Error Handling**
   - On any error: log detailed error message
   - Clean up temporary files
   - Preserve Kindle app state
   - Display actionable error message to user
   - Exit with non-zero status

### Error Scenarios

**No Kindle App Installed**
- Error: "Kindle app is not installed. Please install from: [URL]"
- Exit code: 1

**No Book Open**
- Error: "No book is currently open in Kindle app. Please open a book and try again."
- Exit code: 2

**Kindle App Not in Foreground**
- Attempt to bring to foreground
- If fails: "Please bring Kindle app to foreground and try again"
- Exit code: 3

**Insufficient Disk Space**
- Error: "Insufficient disk space. Need approximately {X} MB, only {Y} MB available."
- Exit code: 4

**Screenshot Capture Failure**
- Log error with page number
- Attempt to continue with next page (configurable)
- If critical: clean up and exit with error

**PDF Generation Failure**
- Error: "Failed to generate PDF: {reason}"
- Clean up temporary files
- Exit code: 5

**User Interruption (Ctrl+C)**
- Display: "Conversion interrupted by user"
- Clean up temporary files and partial PDF
- Exit code: 130

## Correctness Properties

*Properties are characteristics that should hold true across all valid executions*

### Property 1: Valid PDF Output
**For any** currently open book, if conversion succeeds, the output must be a valid PDF file that can be opened by standard PDF readers.
**Validates**: Requirement 1.1

### Property 2: Output Directory Respected
**For any** valid output directory specified by user, the PDF must be saved to that exact location.
**Validates**: Requirement 1.2

### Property 3: Default Output Location
**For any** conversion when no output directory is specified, the PDF must be created in the current working directory.
**Validates**: Requirement 1.3

### Property 4: Success Message Display
**For any** successful conversion, a success message containing the output file path must be displayed.
**Validates**: Requirement 1.4

### Property 5: No Book Open Detection
**For any** execution when no book is open in Kindle app, an informative error message must be displayed asking user to open a book first.
**Validates**: Requirement 1.5

### Property 6: Usage Display
**For any** execution without arguments, usage instructions and available options must be displayed.
**Validates**: Requirement 2.1

### Property 7: Help Flag Response
**For any** execution with help flag, detailed command documentation must be shown.
**Validates**: Requirement 2.2

### Property 8: Path Validation
**For any** output path specified, validation must occur before starting conversion.
**Validates**: Requirement 2.3

### Property 9: Invalid Argument Handling
**For any** invalid command-line arguments, clear error messages and usage examples must be displayed.
**Validates**: Requirement 2.4

### Property 10: Version Display
**For any** execution with version flag, current version information must be displayed.
**Validates**: Requirement 2.5

### Property 11: Kindle App Detection
**For any** execution when Kindle app is not installed, the tool must detect this and provide installation instructions.
**Validates**: Requirement 3.2

### Property 12: Special Character Path Handling
**For any** valid macOS file path containing spaces or special characters, the tool must handle it correctly.
**Validates**: Requirement 3.3

### Property 13: File Permission Respect
**For any** file operation, macOS file permissions and ownership must be respected.
**Validates**: Requirement 3.4

### Property 14: Progress Indicator Display
**For any** conversion in progress, progress indicators showing page capture progress must be displayed.
**Validates**: Requirement 3.5

### Property 15: Sequential Page Processing
**For any** currently open book, conversion must process all pages sequentially from current position to end.
**Validates**: Requirement 4.1

### Property 16: Manual Book Opening Requirement
**For any** conversion, the user must have manually opened the book in Kindle app before starting.
**Validates**: Requirement 4.2

### Property 17: Output Naming Consistency
**For any** series of conversions, output file names must follow consistent naming conventions.
**Validates**: Requirement 4.3

### Property 18: Existing File Handling
**For any** conversion where target file already exists, user must be prompted for overwrite confirmation or conversion must be skipped.
**Validates**: Requirement 4.4

### Property 19: Cancellation Handling
**For any** user cancellation during conversion, the current book conversion must complete before stopping.
**Validates**: Requirement 4.5

### Property 20: Screenshot Failure Recovery
**For any** screenshot capture failure, the specific error must be reported and conversion must attempt to continue or provide clear guidance.
**Validates**: Requirement 5.1

### Property 21: Disk Space Check
**For any** conversion, insufficient disk space must be detected before starting conversion.
**Validates**: Requirement 5.2

### Property 22: Source Preservation
**For any** conversion failure, the original book in Kindle app must remain unaffected.
**Validates**: Requirement 5.3

### Property 23: Cleanup on Interruption
**For any** interrupted conversion, partial screenshot files and incomplete PDFs must be cleaned up.
**Validates**: Requirement 5.4

### Property 24: Verbose Logging
**For any** conversion when verbose logging is enabled, detailed progress including page numbers and screenshot status must be provided.
**Validates**: Requirement 5.5

### Property 25: Screenshot Quality Application
**For any** specified screenshot quality setting, pages must be captured at that quality level.
**Validates**: Requirement 6.1

### Property 26: Page Delay Application
**For any** specified page delay setting, the system must wait that duration between page turns.
**Validates**: Requirement 6.2


### Property 28: Invalid Configuration Fallback
**For any** invalid configuration values, default values must be used and user must be warned.
**Validates**: Requirement 6.4

### Property 29: Default Configuration
**For any** conversion when no configuration is specified, sensible default settings for high-quality output must be used.
**Validates**: Requirement 6.5

### Property 30: Preparation Instructions Display
**For any** conversion start, instructions must be displayed asking user to ensure Kindle app is in foreground and ready.
**Validates**: Requirement 7.1

### Property 31: Preparation Delay
**For any** conversion after user confirmation, a configurable delay must be applied before beginning automation.
**Validates**: Requirement 7.2

### Property 32: Foreground Verification
**For any** automation attempt, the system must verify Kindle app is in foreground or attempt to bring it forward.
**Validates**: Requirement 7.3

### Property 33: Focus Failure Guidance
**For any** automation failure due to app focus issues, clear instructions must be provided for user to manually bring Kindle app to foreground.
**Validates**: Requirement 7.4

### Property 34: Countdown Timer Display
**For any** conversion with configured startup delay, a countdown timer showing remaining preparation time must be displayed.
**Validates**: Requirement 7.5

## Testing Strategy

### Unit Testing
Test individual components in isolation with mocked dependencies:

- **CLI Component**: Argument parsing, flag validation, help/version display
- **File Manager**: Path validation, disk space checking, temp directory management

- **PDF Generator**: PDF creation from test images, quality settings application

### Property-Based Testing
Use Go's `testing/quick` or `rapid` framework to verify correctness properties:

- **Requirements**: Minimum 100 iterations per property test
- **Tag Format**: `// Property {number}: {property description}`
- **Coverage**: Each of the 34 correctness properties must have a corresponding property-based test
- **Placement**: Tests should be placed close to implementation

Example:
```go
// Property 3: Default Output Location
// For any conversion when no output directory is specified,
// the PDF must be created in the current working directory
func TestProperty3_DefaultOutputLocation(t *testing.T) {
    rapid.Check(t, func(t *rapid.T) {
        // Test implementation
    })
}
```

### Integration Testing
Test component interactions with real or realistic mocks:

- **Conversion Orchestrator**: Full workflow with mocked Kindle automation
- **Kindle Automation**: Test with actual Kindle app (manual or CI with GUI)
- **End-to-End**: Complete conversion with test book

### Manual Testing
Document manual test procedures for:

- Installing and running on clean macOS system
- Converting actual Kindle books
- Verifying PDF quality and readability
- Testing error scenarios (no app, no book, etc.)

### Test Data

- Test images simulating Kindle page screenshots
- Expected PDF outputs for comparison

## Error Handling Strategy

### Error Categories

1. **User Input Errors** (Exit codes 1-10)
   - Invalid arguments
   - Invalid configuration
   - Invalid output path

2. **Environment Errors** (Exit codes 11-20)
   - Kindle app not installed
   - No book open
   - Insufficient permissions
   - Insufficient disk space

3. **Runtime Errors** (Exit codes 21-30)
   - Kindle app automation failures
   - Screenshot capture failures
   - PDF generation failures

4. **System Errors** (Exit codes 31-40)
   - File system errors
   - macOS API errors

### Error Recovery

- **Transient failures**: Retry with exponential backoff (e.g., screenshot capture)
- **Permanent failures**: Fail fast with clear error message
- **All failures**: Clean up temporary files before exit
- **Interruptions**: Handle gracefully, clean up, preserve Kindle app state

### Logging

- **Normal mode**: Show progress and errors only
- **Verbose mode**: Show detailed step-by-step progress, API calls, file operations
- **Error messages**: Always actionable with clear next steps for user


## Future Considerations

The following are **not** in scope for initial implementation but may be considered for future versions:

- Automatic book title detection and metadata extraction
- Resume interrupted conversions
- Parallel processing of multiple books
- OCR for searchable PDFs
- Custom page ranges
- Automatic cropping of screenshots
- Cloud storage integration
