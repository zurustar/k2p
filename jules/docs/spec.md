# Requirements Specification

## Overview
The Kindle-to-PDF tool automates the process of capturing pages from the Kindle app on macOS and converting them into a single PDF file.

## Features

### 1. Page Capture & Navigation
- The tool must capture the screen.
- It must simulate page turning key presses (Right arrow for LTR, Left arrow for RTL).
- It must wait for the page turn animation to complete.

### 2. Automatic Direction Detection
- The tool should automatically determine the page turning direction (Left-to-Right or Right-to-Left).
- **Mechanism:**
  - Capture the initial page.
  - Attempt to turn the page in one direction (e.g., Right).
  - Capture the screen again.
  - Compare the two images.
  - If the images are different, the direction is correct.
  - If identical, attempt to turn the page in the opposite direction and verify.
  - If neither works, it might be the end of the book or an error.

### 3. End-of-Book Detection
- The tool must automatically detect when the end of the book is reached.
- **Mechanism:**
  - Compare the currently captured page with the previously captured page.
  - If the content is identical (meaning the page turn didn't change the screen), it indicates the end of the book or a "Review this book" screen where page turning is disabled.
  - The loop should terminate upon detection.

### 4. PDF Generation
- Combine all captured images into a PDF.
- Support splitting by size (existing feature).

## Constraints
- macOS only (relies on `screencapture` and `osascript`).
- Requires Kindle for Mac to be in Full Screen mode (recommended).
