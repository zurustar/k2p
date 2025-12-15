# Makefile for Kindle to PDF Converter

# Variables
BINARY_NAME=kindle-to-pdf
BUILD_DIR=build
CMD_DIR=cmd/kindle-to-pdf

# Default target
.PHONY: all
all: build

# Build the application
.PHONY: build
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) ./$(CMD_DIR)

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	go test -v ./...

# Run tests with coverage
.PHONY: test-coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning..."
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

# Install the application
.PHONY: install
install: build
	@echo "Installing $(BINARY_NAME)..."
	cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/

# Run the application
.PHONY: run
run: build
	./$(BUILD_DIR)/$(BINARY_NAME)

# Format code
.PHONY: fmt
fmt:
	go fmt ./...

# Lint code
.PHONY: lint
lint:
	golangci-lint run

# Tidy dependencies
.PHONY: tidy
tidy:
	go mod tidy

# Help
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build         - Build the application"
	@echo "  test          - Run tests"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  clean         - Clean build artifacts"
	@echo "  install       - Install the application to /usr/local/bin"
	@echo "  run           - Build and run the application"
	@echo "  fmt           - Format code"
	@echo "  lint          - Lint code"
	@echo "  tidy          - Tidy dependencies"
	@echo "  help          - Show this help message"