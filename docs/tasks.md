# Implementation Tasks

This document tracks the implementation progress of the Kindle to PDF converter.

## Current Status
**Phase**: Phase 8 completed + Image Trimming Feature
**Next Steps**: Ready for real-world testing with Kindle app

## Phase 1: Project Setup and Foundation
- [x] Initialize Go module and project structure
- [x] Set up directory structure (cmd, pkg, internal)
- [x] Configure build system and Makefile
- [x] Set up testing framework (unit, property-based, integration)
- [x] Add basic CLI framework (cobra or flag)
- [x] Implement version and help commands

## Phase 2: File Management Component
- [x] Implement `FileManager` interface
  - [x] `ValidateOutputPath()` - validate paths with special characters
  - [x] `CheckDiskSpace()` - check available disk space
  - [x] `ResolveOutputPath()` - handle default directory and naming
  - [x] `CreateTempDir()` - create temporary directory for screenshots
  - [x] `CleanupTempDir()` - cleanup temporary files
  - [x] `HandleExistingFile()` - prompt for overwrite confirmation
- [x] Write unit tests for file operations
- [ ] Write property-based tests for path handling (Property 12)

## Phase 3: Configuration Management
- [x] Define `ConversionOptions` data model
- [x] Implement `ConfigManager` interface
  - [x] `LoadConfig()` - load from YAML/JSON file
  - [x] `MergeOptions()` - merge CLI flags with config file
  - [x] `GetDefaults()` - provide default configuration
- [x] Implement configuration validation
- [x] Write unit tests for config loading and merging
- [ ] Write property-based tests (Properties 25-29)

## Phase 4: Kindle Automation Service
- [x] Research macOS automation APIs (AppleScript/Accessibility)
- [x] Implement `KindleAutomation` interface
  - [x] `IsKindleInstalled()` - check Kindle app installation
  - [x] `IsBookOpen()` - detect if book is currently open
  - [x] `IsKindleInForeground()` - check app focus
  - [x] `TurnNextPage()` - navigate to next page
  - [x] `HasMorePages()` - detect end of book (placeholder)
  - [ ] `CaptureCurrentPage()` - capture screenshot of current page
- [ ] Implement retry logic for transient failures
- [x] Write unit tests with mocked macOS APIs
- [ ] Manual testing with actual Kindle app

## Phase 5: PDF Generator Service
- [x] Research and select Go PDF library (gofpdf, pdfcpu, etc.)
- [x] Implement `PDFGenerator` interface
  - [x] `CreatePDF()` - combine images into PDF
- [x] Implement quality and compression settings
- [x] Handle large numbers of pages efficiently
- [x] Write unit tests with test images
- [ ] Write property-based test (Property 1)

## Phase 6: Conversion Orchestrator
- [x] Implement `ConversionOrchestrator` interface
- [x] Implement main conversion workflow:
  - [x] Display preparation instructions (Property 30)
  - [x] Wait for user confirmation
  - [x] Apply startup delay with countdown timer (Properties 31, 34)
  - [x] Validate Kindle app state (Properties 5, 11, 32)
  - [x] Check disk space (Property 21)
  - [x] Create temporary directory
  - [x] Execute page capture loop (Property 15)
  - [x] Generate PDF from screenshots
  - [x] Clean up temporary files (Property 23)
  - [x] Display success message (Property 4)
- [ ] Implement error handling for all scenarios
- [x] Implement progress indicators (Property 14)
- [ ] Write integration tests

## Phase 7: CLI Implementation
- [x] Implement command-line argument parsing
  - [x] `--output` / `-o` flag (Properties 2, 3)
  - [x] `--quality` flag for screenshot quality (Property 25)
  - [x] `--page-delay` flag (Property 26)
  - [x] `--startup-delay` flag (Property 31)
  - [x] `--pdf-quality` flag
  - [x] `--config` flag for config file (Property 27)
  - [x] `--verbose` / `-v` flag (Property 24)
  - [x] `--auto-confirm` / `-y` flag
  - [x] `--help` / `-h` flag (Property 7)
  - [x] `--version` flag (Property 10)
- [x] Implement usage display (Property 6)
- [x] Implement input validation (Properties 8, 9)
- [x] Wire up CLI to orchestrator
- [ ] Write CLI unit tests

## Phase 8: Error Handling and Edge Cases
- [x] Implement all error scenarios from design:
  - [x] No Kindle app installed (Property 11)
  - [x] No book open (Property 5)
  - [x] Kindle app not in foreground (Properties 32, 33)
  - [x] Insufficient disk space (Property 21)
  - [x] Screenshot capture failure (Property 20)
  - [x] PDF generation failure
  - [x] User interruption (Ctrl+C) (Property 23)
- [x] Implement cleanup on all error paths (Property 22)
- [x] Implement actionable error messages
- [x] Implement retry logic for transient failures
- [ ] Test all error scenarios

## Additional Features
- [x] Image trimming: Automatic border removal
  - [x] Detect black/white borders from corners
  - [x] Trim uniform colored borders
  - [x] Handle edge cases (no borders, mixed colors)
  - [x] CLI flags: --trim-borders (default), --no-trim-borders
  - [x] Integration with orchestrator workflow
  - [x] Comprehensive unit tests

## Phase 9: Property-Based Testing
- [ ] Implement property tests for all 34 properties:
  - [ ] Properties 1-5: Basic conversion (Req 1)
  - [ ] Properties 6-10: CLI interface (Req 2)
  - [ ] Properties 11-14: macOS integration (Req 3)
  - [ ] Properties 15-19: Sequential processing (Req 4)
  - [ ] Properties 20-24: Error handling (Req 5)
  - [ ] Properties 25-29: Configuration (Req 6)
  - [ ] Properties 30-34: User preparation (Req 7)
- [ ] Ensure minimum 100 iterations per test
- [ ] Tag each test with property reference

## Phase 10: Integration and End-to-End Testing
- [ ] Set up test environment with Kindle app
- [ ] Create test books for conversion
- [ ] Test complete conversion workflow
- [ ] Test with various book sizes and formats
- [ ] Test configuration file scenarios
- [ ] Test error recovery scenarios
- [ ] Performance testing with large books

## Phase 11: Documentation and Polish
- [ ] Write comprehensive README.md
  - [ ] Installation instructions
  - [ ] Usage examples
  - [ ] Configuration guide
  - [ ] Troubleshooting section
- [ ] Write developer documentation
- [ ] Add code comments and documentation
- [ ] Create example configuration files
- [ ] Add build and release scripts

## Phase 12: Release Preparation
- [ ] Final testing on clean macOS system
- [ ] Create release build
- [ ] Write release notes
- [ ] Create installation package
- [ ] Test installation process
- [ ] Prepare distribution

## Notes

### Property References
Each property-based test should reference the corresponding design document property:
- Format: `// Property {number}: {description}`
- Minimum 100 iterations per test
- Tests should be placed close to implementation

### Testing Strategy
- **Unit tests**: Test components in isolation with mocks
- **Property-based tests**: Verify correctness properties with generated inputs
- **Integration tests**: Test component interactions
- **End-to-end tests**: Test complete workflows with real Kindle app

### Development Workflow
1. Implement component interface
2. Write unit tests
3. Write property-based tests
4. Implement component logic
5. Run tests and iterate
6. Integration testing
7. Documentation
