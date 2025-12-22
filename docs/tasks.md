# Implementation Tasks

This document tracks the implementation progress of the Kindle to PDF converter.

## Current Status
**Phase**: Phase 8 completed + Enhanced Trimming Feature (Margin Detection & Custom Trimming)
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

## Bug Fixes
- [x] Verbose logging: Fix debug messages appearing without --verbose flag
  - [x] Wrap end-of-book detection debug output in verbose checks
  - [x] Ensure similarity comparison logs only show with --verbose
  - [x] Keep important user-facing messages (end-of-book detection result) always visible


## Maintenance
- [x] Investigate and remove unused code
  - [x] Run static analysis (Completed via manual check)
  - [x] Check for unused exported functions (Completed via manual check)
  - [x] Check for unused constants (Completed via manual check)
  - [x] Remove `pkg/automation` unused methods
  - [x] Remove `internal/orchestrator` unused trim functionality
  - [x] Remove `pkg/imageprocessing` unused trim logic
  - [x] Clean up `config.example.yaml`
  - [x] Update `test_manual.sh`

## Refactoring
- [x] Remove configuration file support
  - [x] Remove `config.example.yaml`
  - [x] Update `cmd/k2p/main.go` (remove flag)
  - [x] Update `pkg/config` (remove loader)
  - [x] Update documentation (README)
- [x] Move `pkg/*` to `internal/*`
  - [x] Move directories
  - [x] Update imports
  - [x] Update documentation
- [x] Decouple option validation from CLI
  - [x] Move validation logic to `config.ConversionOptions.Validate()`
  - [x] Update CLI to use `Validate()`
- [x] Merge trim-left/right into trim-horizontal
  - [x] Update config/CLI
  - [x] Update documentation







## Additional Features
- [x] Image trimming: Automatic border removal
  - [x] Detect black/white borders from corners
  - [x] Trim uniform colored borders independently per edge
  - [x] Handle edge cases (no borders, mixed colors, top-only borders)
  - [x] 95% noise tolerance with lookahead gap skipping
  - [x] CLI flag: --trim-borders (opt-in)
  - [x] Integration with orchestrator workflow (uses originals when trimming is disabled)
  - [x] Detection images remain untrimmed for reliability
  - [x] Comprehensive unit tests
  - [x] Based on gazounomawarinoiranaifuchiwokesu implementation

- [x] Enhanced trimming: Margin detection and custom trimming
  - [x] Two-mode operation: detect and generate
  - [x] Margin detection mode: analyze all pages without PDF output
  - [x] Calculate minimum safe margins across all pages
  - [x] Custom trim margins: specify top, bottom, left, right pixels
  - [x] Replace --trim-borders with --mode and --trim-* flags
  - [x] Update ConversionOptions with Mode and trim margin fields
  - [x] Update orchestrator to support both modes
  - [x] Add margin analysis functions to imageprocessing package
  - [x] Update CLI help text with trimming workflow documentation
  - [x] Update design.md and tasks.md

- [x] Screenshot capture improvements
  - [x] Fullscreen Kindle support
  - [x] Screen recording permission handling
  - [x] Space switching for fullscreen apps
  - [x] Kindleを一度だけアクティブ化し、その後は再アクティベーションなしでCaptureWithoutActivationを使用
  - [x] ページめくり後にPageDelayで待機してから次のキャプチャを実施

- [x] End-of-book detection
  - [x] Detect 5 consecutive identical pages
  - [x] Remove rating screen pages from PDF
  - [x] 95% similarity threshold

- [x] PDF generation improvements
  - [x] Use actual image dimensions (no distortion)
  - [x] Include detection images in PDF
  - [x] Per-page size adjustment

- [x] Documentation updates
  - [x] Add screen recording permission to README
  - [x] Update MANUAL_TESTING.md
  - [x] Create comprehensive walkthrough

## Phase 9: Property-Based Testing
- [ ] Implement property tests for all 34 properties:
  - [x] Properties 1-5: Basic conversion (Req 1)
  - [ ] Properties 6-10: CLI interface (Req 2)
  - [ ] Properties 11-14: macOS integration (Req 3)
  - [ ] Properties 15-19: Sequential processing (Req 4)
  - [ ] Properties 20-24: Error handling (Req 5)
  - [ ] Properties 25-29: Configuration (Req 6)
  - [ ] Properties 30-34: User preparation (Req 7)
- [ ] Ensure minimum 100 iterations per test
- [ ] Tag each test with property reference

## Phase 10: Integration and End-to-End Testing
- [x] Create `test/integration` directory
- [x] Implement Orchestrator integration tests (Mock automation, real FS/PDF)
- [x] Update `MANUAL_TESTING.md` for E2E scenarios
- [x] Enhance `test_manual.sh` for guided E2E testing
- [x] Run and verify tests

## Phase 11: Documentation and Polish
## Phase 11: Documentation and Polish
- [x] Update README.md with Full Reference (Installation, Usage, Troubleshooting)
- [x] Ensure all major functions have GoDoc comments
- [x] Create `CONTRIBUTING.md` for developers
- [x] Finalize `MANUAL_TESTING.md` (already updated in Phase 10)
- [x] Verify `make build` and `make clean` work flawlessly

## Phase 12: Release Preparation
- [ ] Final testing on clean macOS system
- [ ] Create release build
- [ ] Write release notes
- [ ] Create installation package
- [ ] Test installation process
- [ ] Prepare distribution
 
 ## Phase 13: PDF to Markdown Conversion
 - [x] Create `internal/converter` package
   - [x] Define `MarkdownConverter` interface
   - [x] Implement Go wrapper using `ledongthuc/pdf`
     - [x] Import `github.com/ledongthuc/pdf`
     - [x] Iterate over pages and extract text
     - [x] Write to Markdown file
 - [x] Update CLI
   - [x] Add `--mode pdf2md`
   - [x] Add `--input` / `-i` flag
   - [x] Wire up command logic in main.go
 - [x] Integration test
   - [x] Test with a sample text-based PDF

## Phase 14: Web-Based GUI Implementation
- [x] Create `cmd/k2p-gui` directory structure
- [x] Create `cmd/k2p-gui/assets` for static files
- [x] Implement HTML/JS frontend
  - [x] Design form with all configuration options
  - [x] Implement log streaming console
- [x] Implement Go Web Server
  - [x] Serve static files via `embed`
  - [x] Implement API endpoints for conversion
  - [x] Implement WebSocket/SSE for log streaming
  - [x] Auto-open default browser on startup
- [x] Integrate with Orchestrator
  - [x] Hook up `ConversionOptions`
  - [x] Redirect logs to frontend

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
