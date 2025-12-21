package converter

import (
	"bytes"
	"context"
	"fmt"
	"os"

	"github.com/ledongthuc/pdf"
)

// MarkdownConverter converts PDF files to Markdown
type MarkdownConverter interface {
	// ConvertPDFToMarkdown extracts text from PDF and saves as Markdown
	ConvertPDFToMarkdown(ctx context.Context, inputPDF string, outputMarkdown string) error
}

// DefaultConverter is the default implementation using a pure Go library
type DefaultConverter struct{}

// NewConverter creates a new MarkdownConverter
func NewConverter() MarkdownConverter {
	return &DefaultConverter{}
}

// ConvertPDFToMarkdown implements MarkdownConverter
func (c *DefaultConverter) ConvertPDFToMarkdown(ctx context.Context, inputPDF string, outputMarkdown string) error {
	// 1. Validate input file
	if _, err := os.Stat(inputPDF); err != nil {
		return fmt.Errorf("input file not found: %s", inputPDF)
	}

	// 2. Extract text using ledongthuc/pdf
	f, r, err := pdf.Open(inputPDF)
	if err != nil {
		return fmt.Errorf("failed to open PDF: %w", err)
	}
	defer f.Close()

	var buf bytes.Buffer
	totalPage := r.NumPage()

	for pageIndex := 1; pageIndex <= totalPage; pageIndex++ {
		p := r.Page(pageIndex)
		if p.V.IsNull() {
			continue
		}

		text, err := p.GetPlainText(nil)
		if err != nil {
			// If we fail to get text for a page, strict error might be too harsh,
			// but let's log/buffer it. For now, we continue.
			continue
		}

		buf.WriteString(text)
		buf.WriteString("\n\n") // Separation between pages
	}

	// 3. Write to output file
	if err := os.WriteFile(outputMarkdown, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write Markdown file: %w", err)
	}

	return nil
}
