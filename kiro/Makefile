.PHONY: build test clean install test-unit test-property test-integration run

# Build variables
BINARY_NAME=k2p
BUILD_DIR=build
INSTALL_PATH=/usr/local/bin

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOCLEAN=$(GOCMD) clean
GOMOD=$(GOCMD) mod

# Build the application
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/k2p
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Run all tests
test:
	@echo "Running all tests..."
	$(GOTEST) -v ./...

# Run unit tests only
test-unit:
	@echo "Running unit tests..."
	$(GOTEST) -v -short ./...

# Run property-based tests
test-property:
	@echo "Running property-based tests..."
	$(GOTEST) -v -run Property ./...

# Run integration tests
test-integration:
	@echo "Running integration tests..."
	$(GOTEST) -v -run Integration ./...

# Run with race detection
test-race:
	@echo "Running tests with race detection..."
	$(GOTEST) -race -v ./...

# Clean build artifacts
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	@echo "Clean complete"

# Install the binary
install: build
	@echo "Installing $(BINARY_NAME) to $(INSTALL_PATH)..."
	cp $(BUILD_DIR)/$(BINARY_NAME) $(INSTALL_PATH)/$(BINARY_NAME)
	@echo "Installation complete"

# Uninstall the binary
uninstall:
	@echo "Uninstalling $(BINARY_NAME)..."
	rm -f $(INSTALL_PATH)/$(BINARY_NAME)
	@echo "Uninstall complete"

# Run the application
run: build
	$(BUILD_DIR)/$(BINARY_NAME)

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

# Format code
fmt:
	@echo "Formatting code..."
	$(GOCMD) fmt ./...

# Run linter
lint:
	@echo "Running linter..."
	golangci-lint run

# Show help
help:
	@echo "Available targets:"
	@echo "  build            - Build the application"
	@echo "  test             - Run all tests"
	@echo "  test-unit        - Run unit tests only"
	@echo "  test-property    - Run property-based tests"
	@echo "  test-integration - Run integration tests"
	@echo "  test-race        - Run tests with race detection"
	@echo "  clean            - Clean build artifacts"
	@echo "  install          - Install the binary to $(INSTALL_PATH)"
	@echo "  uninstall        - Uninstall the binary"
	@echo "  run              - Build and run the application"
	@echo "  deps             - Download and tidy dependencies"
	@echo "  fmt              - Format code"
	@echo "  lint             - Run linter"
	@echo "  help             - Show this help message"
