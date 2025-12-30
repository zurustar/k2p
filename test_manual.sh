#!/bin/bash
# Kindle to PDF Converter - Manual Test Script (GUI)

echo "=================================="
echo "k2p GUI Manual Test Launcher"
echo "=================================="
echo ""

# Check if build exists
if [ ! -f "./build/k2p-gui" ]; then
    echo "Building k2p-gui..."
    make build
    if [ $? -ne 0 ]; then
        echo "❌ Build failed!"
        exit 1
    fi
    echo "✅ Build successful"
    echo ""
fi

echo "🚀 Launching GUI..."
echo "The application will open in a new window."
echo "Please verify functionality manually using the GUI interface."
echo ""

./build/k2p-gui
