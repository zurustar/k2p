#!/bin/bash
# Kindle to PDF Converter - Manual Test Script

echo "=================================="
echo "k2p Manual Test Script"
echo "=================================="
echo ""

# Check if build exists
if [ ! -f "./build/k2p" ]; then
    echo "Building k2p..."
    make build
    if [ $? -ne 0 ]; then
        echo "❌ Build failed!"
        exit 1
    fi
    echo "✅ Build successful"
    echo ""
fi

echo "📋 Test Checklist:"
echo ""
echo "Before running tests, please:"
echo "  1. ✅ Open Kindle app"
echo "  2. ✅ Sign in to your Amazon account"
echo "  3. ✅ Download a test book (preferably short, 3-5 pages)"
echo "  4. ✅ Open the book to the first page"
echo ""
read -p "Press Enter when ready to start tests..."

# Create test output directory
TEST_DIR="$HOME/Desktop/k2p_test_$(date +%Y%m%d_%H%M%S)"
mkdir -p "$TEST_DIR"
echo ""
echo "📁 Test output directory: $TEST_DIR"
echo ""

# Test 1: Basic conversion with verbose
echo "=================================="
echo "Test 1: Basic Conversion (Verbose)"
echo "=================================="
echo ""
echo "This will convert the currently open book with:"
echo "  - Verbose logging enabled"
echo "  - Border trimming enabled (default)"
echo "  - Default quality settings"
echo ""
read -p "Press Enter to start Test 1..."

./build/k2p --verbose -o "$TEST_DIR/test1_basic"

if [ $? -eq 0 ]; then
    echo "✅ Test 1 completed"
    echo "📄 Check the PDF: $TEST_DIR/test1_basic/"
    ls -lh "$TEST_DIR/test1_basic/"*.pdf 2>/dev/null
else
    echo "❌ Test 1 failed"
fi

echo ""
read -p "Press Enter to continue to Test 2..."

# Test 2: Without border trimming
echo ""
echo "=================================="
echo "Test 2: Without Border Trimming"
echo "=================================="
echo ""
echo "Please:"
echo "  1. Close the current book in Kindle"
echo "  2. Re-open the same book to the first page"
echo ""
read -p "Press Enter when ready for Test 2..."

./build/k2p --no-trim-borders --verbose -o "$TEST_DIR/test2_no_trim"

if [ $? -eq 0 ]; then
    echo "✅ Test 2 completed"
    echo "📄 Check the PDF: $TEST_DIR/test2_no_trim/"
    ls -lh "$TEST_DIR/test2_no_trim/"*.pdf 2>/dev/null
else
    echo "❌ Test 2 failed"
fi

echo ""
echo "=================================="
echo "Test Summary"
echo "=================================="
echo ""
echo "All test outputs are in: $TEST_DIR"
echo ""
echo "Please verify:"
echo "  1. Open both PDFs and compare them"
echo "  2. Check if Test 1 (with trimming) has borders removed"
echo "  3. Check if Test 2 (without trimming) has borders intact"
echo "  4. Verify all pages are captured correctly"
echo ""
echo "To open the test directory:"
echo "  open $TEST_DIR"
echo ""

read -p "Press Enter to open test directory..."
open "$TEST_DIR"

echo ""
echo "✅ Manual testing complete!"
echo ""
echo "Next steps:"
echo "  - Review the generated PDFs"
echo "  - Compare trimmed vs non-trimmed versions"
echo "  - Report any issues found"
echo ""
