# Makefile for Faktura project

# Add Go and local bin paths to the shell PATH
export PATH := $(PATH):/home/geipman/go/bin:/usr/local/go/bin:$(shell pwd)/bin

.PHONY: all build run test generate css install-tools lint clean

all: build

# Install tools: Tailwind CLI, golangci-lint, templ, air
install-tools:
	@echo "Installing/checking tools..."
	@mkdir -p bin
	@if [ ! -f bin/tailwindcss ]; then \
		echo "Downloading Tailwind CSS CLI v4..."; \
		curl -sLO https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-linux-x64; \
		mv tailwindcss-linux-x64 bin/tailwindcss; \
		chmod +x bin/tailwindcss; \
	else \
		echo "Tailwind CSS CLI already installed."; \
	fi
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		echo "Installing golangci-lint..."; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b /home/geipman/go/bin v1.61.0; \
	else \
		echo "golangci-lint already installed."; \
	fi
	@if ! command -v templ >/dev/null 2>&1; then \
		echo "Installing templ..."; \
		go install github.com/a-h/templ/cmd/templ@latest; \
	else \
		echo "templ already installed."; \
	fi
	@if ! command -v air >/dev/null 2>&1; then \
		echo "Installing air..."; \
		go install github.com/air-verse/air@latest; \
	else \
		echo "air already installed."; \
	fi
	@echo "All tools checked and ready!"

# Generate templ components
generate:
	@echo "Generating templ files..."
	@templ generate

# Compile Tailwind CSS
css:
	@echo "Compiling Tailwind CSS..."
	@mkdir -p internal/server/static/css
	@tailwindcss -i assets/css/input.css -o internal/server/static/css/output.css --minify

# Build production binary
build: install-tools generate css
	@echo "Building binary..."
	@go build -o bin/faktura cmd/faktura/main.go

# Run development server with live reload
run: install-tools
	@echo "Starting development server with air..."
	@air

# Run Go tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Run Go tests with coverage report
test-cover:
	@echo "Running tests with coverage..."
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report written to coverage.html"

# Run linter
lint:
	@echo "Running linter..."
	@golangci-lint run

# Clean temporary files
clean:
	@echo "Cleaning up..."
	@rm -rf bin/faktura
	@rm -rf tmp/
	@rm -f coverage.out coverage.html
	@rm -f faktura.db faktura.db-shm faktura.db-wal
	@find . -name "*_templ.go" -exec rm -f {} \;
