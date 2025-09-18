.PHONY: build install test clean run dev deps

# Variables
BINARY_NAME=terminal-history-navigator
BUILD_DIR=bin
INSTALL_PATH=/usr/local/bin

# Main commands
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build -ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY_NAME) main.go

install: build
	@echo "Installing $(BINARY_NAME) to $(INSTALL_PATH)..."
	sudo cp $(BUILD_DIR)/$(BINARY_NAME) $(INSTALL_PATH)/$(BINARY_NAME)
	@echo "Installation complete. Run '$(BINARY_NAME)' to start."

run: build
	@echo "Running $(BINARY_NAME)..."
	./$(BUILD_DIR)/$(BINARY_NAME)

dev:
	@echo "Running in development mode..."
	go run main.go

test:
	@echo "Running tests..."
	go test -v ./...

test-coverage:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

deps:
	@echo "Installing dependencies..."
	go mod tidy
	go mod download

clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

uninstall:
	@echo "Removing $(BINARY_NAME) from $(INSTALL_PATH)..."
	sudo rm -f $(INSTALL_PATH)/$(BINARY_NAME)

# Development commands
fmt:
	@echo "Formatting code..."
	go fmt ./...

lint:
	@echo "Running linter..."
	golangci-lint run

# Create configuration files
setup-config:
	@echo "Setting up configuration directory..."
	@mkdir -p ~/.config/history-nav
	@if [ ! -f ~/.config/history-nav/config.yaml ]; then \
		cp configs/config.example.yaml ~/.config/history-nav/config.yaml; \
		echo "Created config.yaml"; \
	fi
	@if [ ! -f ~/.config/history-nav/templates.yaml ]; then \
		cp configs/templates.example.yaml ~/.config/history-nav/templates.yaml; \
		echo "Created templates.yaml"; \
	fi

# Full installation with setup
full-install: deps build install setup-config
	@echo "Full installation complete!"
	@echo "Add 'alias h=history-navigator' to your shell config file"