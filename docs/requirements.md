# Requirements Document

## Introduction

A native macOS application written in Go that converts Kindle books to PDF format by automating the Kindle app to capture screenshots of each page and combining them into a PDF document. The tool provides a simple graphical interface for users to convert their personal Kindle library books that are accessible through the macOS Kindle application.

## Glossary

- **Kindle_Converter**: The Go-based application that performs the conversion
- **Kindle_App**: The official Amazon Kindle application for macOS
- **Open_Book**: Currently displayed book in the Kindle app (user must open manually)
- **Target_PDF**: Output PDF file generated from captured page screenshots
- **Page_Screenshot**: Individual image capture of a Kindle book page
- **Automation_Engine**: macOS automation system used to turn pages and capture screenshots


## Requirements

### Requirement 1

**User Story:** As a Kindle book owner, I want to convert my currently open Kindle book to PDF format by automating page turns and screenshots, so that I can read them on any device or application that supports PDF.

#### Acceptance Criteria

1. WHEN a user starts the conversion with a book already open in Kindle app, THE Kindle_Converter SHALL wait for user confirmation before beginning the automated process
2. WHEN a user specifies an output directory, THE Kindle_Converter SHALL save the converted PDF to that location
3. WHEN no output directory is specified, THE Kindle_Converter SHALL save the PDF in the current working directory
4. WHEN conversion completes successfully, THE Kindle_Converter SHALL display a success message with the output file path
5. WHEN no book is open in the Kindle app, THE Kindle_Converter SHALL display an informative error message asking the user to open a book first

### Requirement 2

[DELETED] This requirement was for the CLI interface which has been removed.

### Requirement 3

**User Story:** As a Mac user, I want the tool to integrate seamlessly with macOS and the Kindle app, so that I can use it efficiently in my existing workflow.

#### Acceptance Criteria

1. WHEN the tool is installed, THE Kindle_Converter SHALL run natively on macOS and integrate with the installed Kindle app
2. WHEN the Kindle app is not installed, THE Kindle_Converter SHALL detect this and provide installation instructions
3. WHEN file paths contain spaces or special characters, THE Kindle_Converter SHALL handle them correctly on macOS
4. WHEN the tool processes screenshots, THE Kindle_Converter SHALL respect macOS file permissions and ownership
5. WHEN conversion is in progress, THE Kindle_Converter SHALL display progress indicators showing page capture progress

### Requirement 4

**User Story:** As a user with multiple books, I want to convert books one at a time with minimal setup, so that I can process my library efficiently.

#### Acceptance Criteria

1. WHEN a user wants to convert multiple books, THE Kindle_Converter SHALL process one book at a time, requiring the user to manually open each book in Kindle app
2. WHEN a conversion completes, THE Kindle_Converter SHALL wait for the user to open the next book before starting the next conversion
3. WHEN processing multiple books, THE Kindle_Converter SHALL maintain consistent naming conventions for output PDF files
4. WHEN a target PDF file already exists, THE Kindle_Converter SHALL prompt the user for overwrite confirmation or skip the conversion
5. WHEN the user cancels during batch processing, THE Kindle_Converter SHALL complete the current book conversion and then stop
6. WHEN generating filenames automatically, THE Kindle_Converter SHALL use the current timestamp (YYYYMMDD-HHMMSS) at the moment the conversion is initiated to ensure uniqueness and avoid overwrites

### Requirement 5

**User Story:** As a user concerned about process reliability, I want reliable conversion with proper error handling, so that I can trust the conversion process and troubleshoot issues when they occur.

#### Acceptance Criteria

1. WHEN screenshot capture fails during conversion, THE Kindle_Converter SHALL report the specific error and attempt to continue or skip to the next book
2. WHEN insufficient disk space exists, THE Kindle_Converter SHALL detect this condition and inform the user before starting conversion
3. WHEN conversion fails, THE Kindle_Converter SHALL not affect the original book in the Kindle app
4. WHEN the conversion process is interrupted, THE Kindle_Converter SHALL clean up any partial screenshot files and incomplete PDFs
5. WHEN verbose logging is enabled, THE Kindle_Converter SHALL provide detailed conversion progress including page numbers and screenshot status

### Requirement 6

**User Story:** As a developer or advanced user, I want to configure conversion settings, so that I can customize the PDF output quality and capture behavior according to my needs.

#### Acceptance Criteria

1. WHEN a user specifies screenshot quality settings, THE Kindle_Converter SHALL capture pages at the specified resolution and quality
2. WHEN a user requests specific page delay settings, THE Kindle_Converter SHALL wait the specified time between page turns
3. WHEN conversion settings are provided via configuration file, THE Kindle_Converter SHALL read and apply those settings
4. WHEN invalid configuration values are provided, THE Kindle_Converter SHALL use default values and warn the user
5. WHEN no configuration is specified, THE Kindle_Converter SHALL use sensible default settings (quality 100) for high-quality screenshot capture and PDF generation

### Requirement 7

**User Story:** As a user who needs to prepare the Kindle app, I want clear instructions and adequate time to set up the app before automation begins, so that the conversion process works reliably.

#### Acceptance Criteria

1. WHEN the conversion process starts, THE Kindle_Converter SHALL display instructions asking the user to ensure the Kindle app is in the foreground and ready
2. WHEN the user confirms readiness, THE Kindle_Converter SHALL wait a configurable delay period before beginning automation to allow final preparations
3. WHEN the Kindle app is not in the foreground during automation, THE Kindle_Converter SHALL attempt to bring it to the front or display an error message
4. WHEN automation fails due to app focus issues, THE Kindle_Converter SHALL provide clear instructions for the user to manually bring the Kindle app to the foreground
5. WHEN a startup delay is configured, THE Kindle_Converter SHALL display a countdown timer showing the remaining preparation time

### Requirement 8

**User Story:** As a user converting Kindle books with varying margin designs, I want to analyze the optimal trim margins for a book and then apply those margins consistently, so that I can remove Kindle app margins without accidentally trimming book content.

#### Acceptance Criteria

1. WHEN a user runs the tool in margin detection mode, THE Kindle_Converter SHALL capture all pages and analyze the trim margins without generating a PDF
2. WHEN margin detection completes, THE Kindle_Converter SHALL report the minimum trim values (in pixels) for top, bottom, left, and right edges across all pages, along with a recommended symmetric 'horizontal' value
3. WHEN a user runs the tool in PDF generation mode with custom trim margins, THE Kindle_Converter SHALL apply the specified pixel values to trim each page before PDF generation
4. WHEN custom trim margins are specified, THE Kindle_Converter SHALL require three margin values (top, bottom, horizontal) to be provided
5. WHEN margin detection mode is active, THE Kindle_Converter SHALL NOT generate a PDF output file
6. WHEN analyzing margins, THE Kindle_Converter SHALL identify the minimum removable margin for each edge to avoid trimming actual book content that may appear on some pages
7. WHEN trim margins are applied during PDF generation, THE Kindle_Converter SHALL use the same trimming algorithm for consistency with margin detection results

### Requirement 10

**User Story:** As a developer, I want the system to be testable without side effects (like sounds), so that I can run tests silently and efficiently.

#### Acceptance Criteria

1. WHEN running automated tests, THE Kindle_Converter SHALL NOT play any sounds
2. WHEN running the application normally, THE Kindle_Converter SHALL play completion sounds on success
3. WHEN running the application normally, THE Kindle_Converter SHALL play error sounds on failure
4. WHEN the application is packaged, THE Kindle_Converter SHALL be named `k2p-gui.app`.


**User Story:** As a user who prefers graphical interfaces, I want a native macOS application window, so that I can configure and run conversions without using the command line.

#### Acceptance Criteria

1. WHEN the application is launched via the GUI binary, THE Kindle_Converter SHALL open a native graphical window (not a terminal)
2. WHEN using the GUI, THE Kindle_Converter SHALL provide visual controls for all major configuration options (Mode, Output Dir, Quality, Delays, etc.)
3. WHEN selecting file or directory paths, THE Kindle_Converter SHALL open standard macOS native file picker dialogs
4. WHEN a conversion is running, THE Kindle_Converter SHALL display logs and status within the GUI window
5. WHEN the GUI window is closed, THE Kindle_Converter SHALL terminate the application process cleanly
6. WHEN running in GUI mode, THE Kindle_Converter SHALL support all underlying conversion features (Generate, Detect, PDF2MD) available in the CLI
7. WHEN the GUI application starts, THE Kindle_Converter SHALL pre-fill the output directory field with the user's Desktop path by default
8. WHEN margin detection completes in the GUI, THE Kindle_Converter SHALL display the results (margins and recommendations) in a dedicated, easy-to-read area separate from the scrolling logs. The text color SHALL be high-contrast against the background for readability.
```