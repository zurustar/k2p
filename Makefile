APP_NAME := k2p-gui
APP_ID := com.k2p.app
ICON := $(shell pwd)/cmd/k2p-gui/Icon.png
GUI_SRC_DIR := ./cmd/k2p-gui
CMD_DIR := cmd
BUILD_DIR := build
FYNE_CMD := $(HOME)/go/bin/fyne

.PHONY: all build build-gui package clean checks run

all: package

checks:
	@if [ ! -f "$(ICON)" ]; then echo "Error: $(ICON) not found. Please ensure icon exists."; exit 1; fi

build: build-gui

build-gui:
	@echo "Building GUI..."
	@mkdir -p $(BUILD_DIR)
	@go build -ldflags "-w -s" -o $(BUILD_DIR)/$(APP_NAME) $(GUI_SRC_DIR)

package: checks build-gui
	mkdir -p $(BUILD_DIR)
	$(FYNE_CMD) package --os darwin --name "$(APP_NAME)" --app-id "$(APP_ID)" --src "$(GUI_SRC_DIR)" --icon "$(ICON)" --release
	@if [ -d "$(APP_NAME).app" ]; then \
		rm -rf $(BUILD_DIR)/$(APP_NAME).app; \
		mv $(APP_NAME).app $(BUILD_DIR)/; \
		echo "Package created at $(BUILD_DIR)/$(APP_NAME).app"; \
	else \
		echo "Error: Failed to create $(APP_NAME).app"; \
		exit 1; \
	fi

clean:
	rm -rf $(BUILD_DIR)
	rm -rf $(APP_NAME).app
