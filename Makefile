APP_NAME := k2p
GUI_APP_NAME := k2p-gui
APP_ID := com.k2p.app
ICON := $(shell pwd)/cmd/k2p-gui/Icon.png
GUI_SRC_DIR := ./cmd/k2p-gui
CLI_SRC_DIR := ./cmd/k2p
BUILD_DIR := build
FYNE_CMD := $(HOME)/go/bin/fyne

.PHONY: all build build-cli build-gui package clean checks

all: package

checks:
	@if [ ! -f "$(ICON)" ]; then echo "Error: $(ICON) not found. Please ensure icon exists."; exit 1; fi

build: build-cli build-gui

build-cli:
	mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(APP_NAME) $(CLI_SRC_DIR)

build-gui:
	mkdir -p $(BUILD_DIR)
	go build -ldflags "-w -s" -o $(BUILD_DIR)/$(GUI_APP_NAME) $(GUI_SRC_DIR)

package: checks build-gui
	mkdir -p $(BUILD_DIR)
	$(FYNE_CMD) package --os darwin --name "$(GUI_APP_NAME)" --app-id "$(APP_ID)" --src "$(GUI_SRC_DIR)" --icon "$(ICON)" --release
	@if [ -d "$(GUI_APP_NAME).app" ]; then \
		rm -rf $(BUILD_DIR)/$(GUI_APP_NAME).app; \
		mv $(GUI_APP_NAME).app $(BUILD_DIR)/; \
		echo "Package created at $(BUILD_DIR)/$(GUI_APP_NAME).app"; \
	else \
		echo "Error: Failed to create $(GUI_APP_NAME).app"; \
		exit 1; \
	fi

clean:
	rm -rf $(BUILD_DIR)
	rm -rf $(GUI_APP_NAME).app
