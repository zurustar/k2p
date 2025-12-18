# Design Documentation

## Architecture
The application follows a clean architecture with:
- `cmd/k2p`: Entry point and CLI handling.
- `internal/capturer`: Core logic for capturing and page navigation.
- `internal/converter`: PDF generation logic.
- `internal/platform`: OS-specific interactions (screenshots, key presses).

## Detailed Design

### Image Comparison
- Location: `internal/capturer/image_util.go`
- Functionality: Load two images and compare them to determine if the page has changed.
- Strategy: Pixel-by-pixel comparison.
  - To handle potential noise (e.g., system clock), we might consider:
    - Downsampling before comparison.
    - Comparing only the central region.
    - Using a perceptual hash (too complex for now?).
  - **Decision:** Start with a simple byte comparison of the file content if the screenshot tool is reliable. However, decoding and comparing pixels is safer against metadata changes. We will decode PNGs and compare bounds and pixel data.

### Direction Detection
- Location: `Capturer.DetectDirection`
- Flow:
  1. Take `snapshot_A`.
  2. Press `KeyRight`. Wait.
  3. Take `snapshot_B`.
  4. if `Compare(snapshot_A, snapshot_B) == DIFFERENT`: return `LTR`.
  5. Press `KeyLeft` (to go back). Wait.
  6. Press `KeyLeft`. Wait.
  7. Take `snapshot_C`.
  8. if `Compare(snapshot_A, snapshot_C) == DIFFERENT`: return `RTL`.
  9. Default/Error: Prompt user or fail.

### End-of-Book Detection
- Location: `Capturer.CaptureLoop`
- Flow:
  - Loop:
    1. Capture `current_page.png`.
    2. If `previous_page.png` exists:
       - if `Compare(current, previous) == SAME`:
         - Log "End of book detected".
         - Break loop.
    3. Turn Page.
    4. Wait.

### Dependencies
- Standard library `image/png` for decoding.
- Standard library `image` for bounds/pixels.
