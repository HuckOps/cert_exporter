# Certificate Exporter Makefile
# Multi-environment compilation support

# Project name
PROJECT_NAME := cert_exporter
BUILD_DIR := ./bin
PACKAGE_DIR := ./package

GO := go

GOOSs := linux darwin windows
GOARCHs := amd64 arm64

# Version information
VERSION := 0.1.0
REVISION := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || echo "unknown")

# Build flags with version information
BUILD_FLAGS := -ldflags "-X main.version=$(VERSION) -X main.revision=$(REVISION) -X main.buildDate=$(BUILD_DATE)"

# Default build target (local environment)
build:
	@echo "Building $(PROJECT_NAME) for local environment..."
	@mkdir -p $(BUILD_DIR)
	@if [ "$(GOOS)" = "windows" ]; then \
		GOOS=$(GOOS) GOARCH=$(GOARCH) $(GO) build $(BUILD_FLAGS) -o $(BUILD_DIR)/$(PROJECT_NAME).exe .; \
		echo "Build completed: $(BUILD_DIR)/$(PROJECT_NAME).exe"; \
	else \
		GOOS=$(GOOS) GOARCH=$(GOARCH) $(GO) build $(BUILD_FLAGS) -o $(BUILD_DIR)/$(PROJECT_NAME) .; \
		echo "Build completed: $(BUILD_DIR)/$(PROJECT_NAME)"; \
	fi

# Build for all platforms
build-all: 
	@rm -rf $(BUILD_DIR)
	@for GOOS in $(GOOSs); do \
		for GOARCH in $(GOARCHs); do \
			echo "Building $(PROJECT_NAME) for $$GOOS/$$GOARCH..."; \
			mkdir -p $(BUILD_DIR)/$(PROJECT_NAME)-$$GOOS-$$GOARCH/; \
			if [ "$$GOOS" = "windows" ]; then \
				GOOS=$$GOOS GOARCH=$$GOARCH $(GO) build $(BUILD_FLAGS) -o $(BUILD_DIR)/$(PROJECT_NAME)-$$GOOS-$$GOARCH/$(PROJECT_NAME).exe .; \
				echo "Build completed: $(BUILD_DIR)/$(PROJECT_NAME)-$$GOOS-$$GOARCH/$(PROJECT_NAME).exe"; \
			else \
				GOOS=$$GOOS GOARCH=$$GOARCH $(GO) build $(BUILD_FLAGS) -o $(BUILD_DIR)/$(PROJECT_NAME)-$$GOOS-$$GOARCH/$(PROJECT_NAME) .; \
				echo "Build completed: $(BUILD_DIR)/$(PROJECT_NAME)-$$GOOS-$$GOARCH/$(PROJECT_NAME)"; \
			fi; \
			cp config.yaml $(BUILD_DIR)/$(PROJECT_NAME)-$$GOOS-$$GOARCH/; \
			done \
		done

# Package built binaries
package: build-all
	@echo "Packaging all builds..."
	@rm -rf $(PACKAGE_DIR)
	@mkdir -p $(PACKAGE_DIR)
	@for GOOS in $(GOOSs); do \
		for GOARCH in $(GOARCHs); do \
			echo "Packaging $(PROJECT_NAME) for $$GOOS/$$GOARCH..."; \
			tar -czvf $(PACKAGE_DIR)/$(PROJECT_NAME)-$$GOOS-$$GOARCH.tar.gz -C $(BUILD_DIR)/$(PROJECT_NAME)-$$GOOS-$$GOARCH/ .; \
			echo "Package completed: $(PACKAGE_DIR)/$(PROJECT_NAME)-$$GOOS-$$GOARCH.tar.gz"; \
			done \
		done
	@echo "All packages completed: $(PACKAGE_DIR)/"

# Clean build artifacts
clean: 
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@echo "Clean completed"



