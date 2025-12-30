package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"

	"github.com/oumi/k2p/internal/config"
	"github.com/oumi/k2p/internal/converter"
	"github.com/oumi/k2p/internal/orchestrator"
)

func main() {
	a := app.New()
	w := a.NewWindow("k2p - Kindle to PDF")
	w.Resize(fyne.NewSize(600, 700))

	// UI Components references (for binding)
	var (
		outputDir    *widget.Entry
		inputFile    *widget.Entry
		pageTurnKey  *widget.Select
		quality      *widget.Entry
		pdfQuality   *widget.Select
		pageDelay    *widget.Entry
		startupDelay *widget.Entry
		trimH        *widget.Entry
		trimTop      *widget.Entry
		trimBottom   *widget.Entry
		verbose      *widget.Check
		autoConfirm  *widget.Check
		logArea      *widget.Entry
		startBtn     *widget.Button
		statusLabel  *widget.Label
	)

	// Get defaults
	defaults := config.ApplyDefaults(nil)

	// --- 1. Settings Form Components ---

	// Output
	outputDir = widget.NewEntry()
	// Set default to Desktop
	if home, err := os.UserHomeDir(); err == nil {
		outputDir.SetText(filepath.Join(home, "Desktop"))
	} else {
		outputDir.SetPlaceHolder("Current Directory")
	}

	outputDirBtn := widget.NewButton("Browse", func() {
		dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
			if uri != nil {
				outputDir.SetText(uri.Path())
			}
		}, w)
	})

	// Input (for pdf2md)
	inputFile = widget.NewEntry()
	inputFile.SetPlaceHolder("/path/to/book.pdf")
	inputFileBtn := widget.NewButton("Browse", func() {
		fd := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if reader != nil {
				inputFile.SetText(reader.URI().Path())
			}
		}, w)
		fd.SetFilter(storage.NewExtensionFileFilter([]string{".pdf"}))
		fd.Show()
	})

	// Options
	pageTurnKey = widget.NewSelect([]string{"Auto (Right/Left)", "Right", "Left"}, nil)
	pageTurnKey.SetSelected("Auto (Right/Left)")

	quality = widget.NewEntry()
	quality.SetText(strconv.Itoa(defaults.ScreenshotQuality))

	pdfQuality = widget.NewSelect([]string{"High", "Medium", "Low"}, nil)
	// Capitalize default for UI match (internal is "high", UI is "High")
	// Or just default to High since we know it
	pdfQuality.SetSelected("High")

	pageDelay = widget.NewEntry()
	// Convert duration to int ms
	pageDelay.SetText(strconv.Itoa(int(defaults.PageDelay.Milliseconds())))

	startupDelay = widget.NewEntry()
	// Convert duration to int seconds
	startupDelay.SetText(strconv.Itoa(int(defaults.StartupDelay.Seconds())))

	// Trimming
	trimH = widget.NewEntry()
	trimH.SetText("0")
	trimTop = widget.NewEntry()
	trimTop.SetText("0")
	trimBottom = widget.NewEntry()
	trimBottom.SetText("0")

	// Flags
	verbose = widget.NewCheck("Verbose Logging", nil)
	autoConfirm = widget.NewCheck("Auto Confirm", nil)

	// --- 2. Layouts ---

	// Helper to create form rows
	formRow := func(label string, inputs ...fyne.CanvasObject) *fyne.Container {
		var content fyne.CanvasObject
		if len(inputs) == 1 {
			content = inputs[0]
		} else {
			// Use Grid to share space equally, or HBox if we want them packed but typically we want expansion
			// GridWithColumns is good for "Delays: [ 500 ] [ 3 ]"
			content = container.NewGridWithColumns(len(inputs), inputs...)
		}
		return container.New(layout.NewFormLayout(), widget.NewLabel(label), content)
	}

	// Tab 1: Generate PDF
	tabGenerate := container.NewVBox(
		widget.NewLabelWithStyle("Generate PDF from Kindle", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		formRow("Output Dir:", outputDir, outputDirBtn),
		widget.NewSeparator(),
		widget.NewLabel("Trimming (Pixels):"),
		formRow("Horizontal:", trimH),
		formRow("Top / Bottom:", trimTop, trimBottom),
		widget.NewSeparator(),
		widget.NewLabel("Settings:"),
		formRow("Page Turn:", pageTurnKey),
		formRow("Qual (1-100):", quality),
		formRow("PDF Qual:", pdfQuality),
		formRow("Delays (ms/s):", pageDelay, startupDelay),
		container.NewHBox(verbose, autoConfirm),
	)

	// Tab 2: Detect Margins
	tabDetect := container.NewVBox(
		widget.NewLabelWithStyle("Detect Optimal Margins", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel("Analyzes pages without generating PDF."),
		widget.NewSeparator(),
		formRow("Page Turn:", pageTurnKey),
		formRow("Delays (ms/s):", pageDelay, startupDelay),
		container.NewHBox(verbose, autoConfirm),
		widget.NewSeparator(),
		widget.NewLabelWithStyle("Result:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
	)

	resultLabel := widget.NewLabel("")
	resultLabel.TextStyle = fyne.TextStyle{Monospace: true}
	resultLabel.Wrapping = fyne.TextWrapWord

	resultScroll := container.NewVScroll(resultLabel)
	resultScroll.SetMinSize(fyne.NewSize(0, 150))

	tabDetect.Add(resultScroll)

	// Tab 3: PDF to Markdown
	tabPdf2Md := container.NewVBox(
		widget.NewLabelWithStyle("PDF to Markdown", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		formRow("Input PDF:", inputFile, inputFileBtn),
		formRow("Output Dir:", outputDir, outputDirBtn), // Reuse output dir
	)

	tabs := container.NewAppTabs(
		container.NewTabItem("Generate", container.NewPadded(tabGenerate)),
		container.NewTabItem("Detect", container.NewPadded(tabDetect)),
		container.NewTabItem("PDF2MD", container.NewPadded(tabPdf2Md)),
	)

	// --- 3. Logs & Actions ---
	logArea = widget.NewMultiLineEntry()
	logArea.TextStyle = fyne.TextStyle{Monospace: true}
	logArea.Disable() // Read-only
	logScroll := container.NewScroll(logArea)
	logScroll.SetMinSize(fyne.NewSize(0, 200))

	statusLabel = widget.NewLabel("Ready")
	statusLabel.Alignment = fyne.TextAlignCenter

	startBtn = widget.NewButton("Start Conversion", nil) // Handler attached below
	startBtn.Importance = widget.HighImportance

	// Better layout: Top part is tabs, Bottom is logs.
	// We want logs to expand.
	split := container.NewVSplit(
		tabs,
		container.NewBorder(
			container.NewVBox(startBtn, statusLabel, widget.NewLabel("Logs:")),
			nil, nil, nil,
			logScroll,
		),
	)
	split.SetOffset(0.6) // 60% for tabs

	w.SetContent(split)

	// --- 4. Logic ---

	// Log Writer
	logWriter := &uiWriter{entry: logArea}

	startBtn.OnTapped = func() {
		startBtn.Disable()
		statusLabel.SetText("Running...")
		logArea.SetText("") // Clear logs

		// Collect Config
		mode := "generate"
		if tabs.Selected().Text == "Detect" {
			mode = "detect"
		} else if tabs.Selected().Text == "PDF2MD" {
			mode = "pdf2md"
		}

		// Helper to parse int
		parseInt := func(e *widget.Entry) int {
			val, _ := strconv.Atoi(e.Text)
			return val
		}

		// Helper for page turn
		ptKey := "right"
		if pageTurnKey.Selected == "Left" {
			ptKey = "left"
		}
		// "Auto" -> "right" (orchestrator handles auto-detection logic if configured)
		// Wait, Orchestrator expects "right" or "left".
		// If "Auto" is selected, we should verify what orchestrator expects.
		// Current logic in Orchestrator: if options.PageTurnKey != "left", it attempts auto-detect.
		// So passing "right" (default) allows auto-detect.

		opts := &config.ConversionOptions{
			OutputDir:         outputDir.Text,
			Mode:              mode,
			InputFile:         inputFile.Text,
			PageTurnKey:       ptKey,
			ScreenshotQuality: parseInt(quality),
			PDFQuality:        strings.ToLower(pdfQuality.Selected),
			PageDelay:         time.Duration(parseInt(pageDelay)) * time.Millisecond,
			StartupDelay:      time.Duration(parseInt(startupDelay)) * time.Second,
			TrimHorizontal:    parseInt(trimH),
			TrimTop:           parseInt(trimTop),
			TrimBottom:        parseInt(trimBottom),
			Verbose:           verbose.Checked,
			AutoConfirm:       autoConfirm.Checked,
		}

		finalOpts := config.ApplyDefaults(opts)

		// Run in Goroutine
		go func() {
			defer func() {
				startBtn.Enable()
				statusLabel.SetText("Done")
			}()

			ctx := context.Background()

			// Capture stdout/stderr
			// Creating a pipe to capture fmt.Println from orchestrator
			// Fyne doesn't support easy redirection of os.Stdout globally for the app
			// without pipes.

			// We will just write a wrapper that logs to both our UI and Stdout
			// BUT, Orchestrator uses fmt.Printf directly.
			// We MUST hijack os.Stdout/Stderr to see orchestrator logs.
			pr, pw, _ := os.Pipe()
			oldStdout := os.Stdout
			oldStderr := os.Stderr
			os.Stdout = pw
			os.Stderr = pw

			outC := make(chan string)

			// Copy pipe to UI
			go func() {
				buf := make([]byte, 1024)
				for {
					n, err := pr.Read(buf)
					if n > 0 {
						chunk := string(buf[:n])
						logWriter.Write([]byte(chunk)) // Append to UI
						// Also write to real stdout for debugging
						oldStdout.Write(buf[:n])
					}
					if err != nil {
						break
					}
				}
				close(outC)
			}()

			var err error
			var result *orchestrator.ConversionResult
			if finalOpts.Mode == "pdf2md" {
				fmt.Printf("Converting PDF to Markdown...\nInput: %s\n", finalOpts.InputFile)
				outputPath := finalOpts.OutputDir
				if outputPath == "" {
					outputPath = finalOpts.InputFile + ".md"
				}
				conv := converter.NewConverter()
				err = conv.ConvertPDFToMarkdown(ctx, finalOpts.InputFile, outputPath)
			} else {
				orch := orchestrator.NewOrchestrator()
				result, err = orch.ConvertCurrentBook(ctx, finalOpts)
			}

			// Update result display if in detect mode
			if err == nil && finalOpts.Mode == "detect" && result != nil && result.DetectedMargins != nil {
				margins := result.DetectedMargins
				maxH := margins.Left
				if margins.Right > maxH {
					maxH = margins.Right
				}

				resText := fmt.Sprintf(
					"Top:    %d\nBottom: %d\nLeft:   %d\nRight:  %d\n\n"+
						"Recommended Settings:\n"+
						"Trim Top:        %d\n"+
						"Trim Bottom:     %d\n"+
						"Trim Horizontal: %d (Max of Left/Right)",
					margins.Top, margins.Bottom, margins.Left, margins.Right,
					margins.Top, margins.Bottom, maxH,
				)
				resultLabel.SetText(resText)
			} else if finalOpts.Mode == "detect" {
				// Clear on failure or if no result
				resultLabel.SetText("")
			}

			// Restore stdout
			pw.Close()
			os.Stdout = oldStdout
			os.Stderr = oldStderr
			<-outC // wait for copier

			if err != nil {
				dialog.ShowError(err, w)
				statusLabel.SetText("Failed")
			} else {
				dialog.ShowInformation("Success", "Conversion Completed Successfully!", w)
			}
		}()
	}

	w.ShowAndRun()
}

// uiWriter implements io.Writer and appends to a MultiLineEntry
type uiWriter struct {
	entry *widget.Entry
}

func (w *uiWriter) Write(p []byte) (n int, err error) {
	// Must run on main thread
	// But we are in a goroutine context usually
	// Fyne is thread-safe mostly for SetText? No, usually safer to use Append or binding.
	// entry.Append is not a method of MultiLineEntry in older fyne?
	// Checking v2... SetText is available.
	// For large logs, appending string is expensive.
	// But for this use case, it's fine.

	text := string(p)
	// Remove ANSI codes if any (Orchestrator might use colors?)
	// For now, raw.

	// We should probably limit the log size
	current := w.entry.Text
	if len(current) > 100000 {
		current = current[50000:] // Truncate old logs
	}

	w.entry.SetText(current + text)
	// Auto-scroll happens if we refresh/cursor?
	// MultiLineEntry auto-scrolls to cursor usually.
	w.entry.CursorRow = len(strings.Split(w.entry.Text, "\n"))
	w.entry.Refresh()

	return len(p), nil
}
