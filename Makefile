GO_VERSION   ?= 1.24.1
APPNAME    ?= nvcat

OS := $(shell uname -s | tr '[:upper:]' '[:lower:]')

ARCH := $(shell uname -m)
ifeq ($(ARCH),x86_64)
	ARCH = amd64
endif

GO_TARBALL = go$(GO_VERSION).$(OS)-$(ARCH).tar.gz
URL = https://go.dev/dl/$(GO_TARBALL)

.PHONY: install clean

install:
	@echo "Downloading Go"
	@curl -L $(URL) -o $(GO_TARBALL)
	@echo "Extracting..."
	@tar -xzf $(GO_TARBALL)
	@./go/bin/go build -o $(APPNAME)
	@echo "Installing to /usr/local/bin/ ..."
	@install -m 0755 $(APPNAME) /usr/local/bin/$(APPNAME)
	@echo "Cleaning up..."
	@rm -f $(APPNAME).tar.gz $(APPNAME)
	@echo "$(APPNAME) installed successfully."

clean:
	@rm -f $(APPNAME).tar.gz $(APPNAME)
