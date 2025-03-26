GO_VERSION ?= 1.24.1
APPNAME    ?= nvcat
BUILD_DIR  := ./build

OS := $(shell uname -s | tr '[:upper:]' '[:lower:]')
ARCH := $(shell uname -m)
ifeq ($(ARCH),x86_64)
	ARCH = amd64
endif

ifeq ($(OS),windows)
	EXT = zip
else
	EXT = tar.gz
endif

GO_ARCHIVE = go$(GO_VERSION).$(OS)-$(ARCH).$(EXT)
URL = https://go.dev/dl/$(GO_ARCHIVE)

ifeq ($(OS),windows)
	INSTALL_DIR = $(shell echo $$WINDIR)\System32
	APP_EXE = $(APPNAME).exe
	CP = copy
	GO_BIN = go/bin/go.exe # relative to BUILD_DIR
else
	CP = install -m 0755
	INSTALL_DIR = /usr/local/bin
	APP_EXE = $(APPNAME)
	GO_BIN = go/bin/go # relative to BUILD_DIR
endif

.PHONY: install clean

install:
	@echo "Creating build directory..."
	@mkdir -p $(BUILD_DIR)
	@echo "Downloading Go..."
	@curl -L $(URL) -o $(BUILD_DIR)/$(GO_ARCHIVE)
	@echo "Extracting..."
ifeq ($(OS),windows)
	@cd $(BUILD_DIR) && unzip -q $(GO_ARCHIVE)
else
	@cd $(BUILD_DIR) && tar -xzf $(GO_ARCHIVE)
endif

	@echo "Building $(APP_EXE)..."
	@$(BUILD_DIR)/go/bin/go build -o $(BUILD_DIR)/$(APP_EXE)
	@echo "Installing to $(INSTALL_DIR)..."

ifeq ($(OS),windows)
	@$(CP) $(BUILD_DIR)/$(APP_EXE) "$(INSTALL_DIR)\$(APP_EXE)"
else
	@$(CP) $(BUILD_DIR)/$(APP_EXE) $(INSTALL_DIR)/$(APP_EXE)
endif
	@echo "Cleaning up..."
	@rm -rf $(BUILD_DIR)
	@echo "$(APP_EXE) installed successfully."

clean:
	@rm -rf $(BUILD_DIR)
