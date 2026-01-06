# Certificate Exporter Makefile
# Multi-environment compilation support

# Project name
PROJECT_NAME := cert_exporter
BUILD_DIR := ./bin
PACKAGE_DIR := ./package

GO := go

GOOSs := linux darwin windows
GOARCHs := amd64 arm64

# Default build target (local environment)
build:
	@echo "Building $(PROJECT_NAME) for local environment..."
	@mkdir -p $(BUILD_DIR)
	@if [ "$(GOOS)" = "windows" ]; then \
		$(GO) build -o $(BUILD_DIR)/$(PROJECT_NAME).exe .; \
		echo "Build completed: $(BUILD_DIR)/$(PROJECT_NAME).exe"; \
	else \
		$(GO) build -o $(BUILD_DIR)/$(PROJECT_NAME) .; \
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
				GOOS=$$GOOS GOARCH=$$GOARCH $(GO) build -o $(BUILD_DIR)/$(PROJECT_NAME)-$$GOOS-$$GOARCH/$(PROJECT_NAME).exe .; \
				echo "Build completed: $(BUILD_DIR)/$(PROJECT_NAME)-$$GOOS-$$GOARCH/$(PROJECT_NAME).exe"; \
			else \
				GOOS=$$GOOS GOARCH=$$GOARCH $(GO) build -o $(BUILD_DIR)/$(PROJECT_NAME)-$$GOOS-$$GOARCH/$(PROJECT_NAME) .; \
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



