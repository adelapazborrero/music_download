.PHONY: build install clean run test fmt vet help

BINARY_NAME=music-download
BUILD_DIR=bin
INSTALL_PATH=/usr/local/bin

help: ## Show this help message
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

build: ## Build the application
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/music-download
	@echo "✓ Built $(BUILD_DIR)/$(BINARY_NAME)"

install: build ## Install the binary to system path
	@sudo cp $(BUILD_DIR)/$(BINARY_NAME) $(INSTALL_PATH)/$(BINARY_NAME)
	@echo "✓ Installed to $(INSTALL_PATH)/$(BINARY_NAME)"

clean: ## Remove build artifacts
	@rm -rf $(BUILD_DIR)
	@echo "✓ Cleaned build directory"

run: build ## Build and run the application
	@$(BUILD_DIR)/$(BINARY_NAME)

fmt: ## Format Go code
	@go fmt ./...
	@echo "✓ Formatted code"

vet: ## Run Go vet
	@go vet ./...
	@echo "✓ Vet passed"

test: ## Run tests
	@go test ./... -v
	@echo "✓ Tests passed"

deps: ## Download dependencies
	@go mod download
	@go mod tidy
	@echo "✓ Dependencies updated"
