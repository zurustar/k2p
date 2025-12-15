# Kindle to PDF Converter

A command-line tool written in Go that converts Kindle books (AZW, AZW3, MOBI formats) to PDF format on macOS systems.

## Project Structure

```
kindle-to-pdf-go/
├── cmd/
│   └── kindle-to-pdf/          # Main application entry point
│       └── main.go
├── internal/
│   ├── cli/                    # Command-line interface
│   │   ├── root.go            # Root command and CLI setup
│   │   └── convert.go         # Convert command implementation
│   ├── interfaces/            # Core interfaces and data models
│   │   ├── converter.go       # ConverterService interface
│   │   ├── filemanager.go     # FileManager interface
│   │   └── calibre.go         # CalibreInterface
│   └── testing/               # Testing utilities
│       └── helpers.go         # Property-based testing helpers
├── .kiro/
│   └── specs/                 # Specification documents
│       └── kindle-to-pdf-go/
│           ├── requirements.md
│           ├── design.md
│           └── tasks.md
├── go.mod                     # Go module definition
├── Makefile                   # Build and development tasks
└── README.md                  # This file
```

## Prerequisites

- Go 1.19 or later
- Calibre (for ebook conversion)

## Building

```bash
# Build the application
make build

# Run tests
make test

# Install to /usr/local/bin
make install
```

## Usage

```bash
# Convert a single file
kindle-to-pdf convert book.azw3

# Convert with custom output directory
kindle-to-pdf convert book.azw3 --output /path/to/output

# Convert all files in a directory
kindle-to-pdf convert /path/to/kindle/books

# Show help
kindle-to-pdf --help

# Show version
kindle-to-pdf --version
```

## Development

This project follows the specification-driven development methodology. See the `.kiro/specs/` directory for detailed requirements, design, and implementation tasks.

## Testing

The project uses both unit testing and property-based testing:

- Unit tests verify specific examples and edge cases
- Property-based tests verify universal properties across many inputs
- Minimum 100 iterations for property-based tests

Run tests with:
```bash
make test
```