# Makefile for Faktura project

# Add Go and local bin paths to the shell PATH
export PATH := $(PATH):/home/geipman/go/bin:/usr/local/go/bin:$(shell pwd)/bin

.PHONY: all build run test generate css install-tools lint clean migrate install-mustang

all: build

# Run data migration from legacy XML to SQLite
migrate:
	@echo "Running data migration..."
	@go run cmd/migrate/main.go

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
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b /home/geipman/go/bin v1.64.8; \
	else \
		echo "golangci-lint already installed."; \
	fi
	@if ! command -v templ >/dev/null 2>&1; then \
		echo "Installing templ..."; \
		go install github.com/a-h/templ/cmd/templ@v0.3.1020; \
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

# Install Mustang CLI and JRE for integration testing
install-mustang:
	@echo "Installing/checking Mustang CLI and JRE..."
	@mkdir -p bin
	@if [ ! -f bin/mustang-cli.jar ]; then \
		echo "Downloading Mustang CLI JAR..."; \
		curl -sSL -o bin/mustang-cli.jar https://repo1.maven.org/maven2/org/mustangproject/Mustang-CLI/2.24.0/Mustang-CLI-2.24.0.jar; \
	else \
		echo "Mustang CLI JAR already downloaded."; \
	fi
	@if [ ! -d bin/jre ]; then \
		echo "Downloading JRE 21..."; \
		curl -sSL -o bin/jre.tar.gz https://api.adoptium.net/v3/binary/latest/21/ga/linux/x64/jre/hotspot/normal/eclipse; \
		echo "Extracting JRE..."; \
		mkdir -p bin/jre; \
		tar -xzf bin/jre.tar.gz -C bin/jre --strip-components=1; \
		rm bin/jre.tar.gz; \
	else \
		echo "JRE already installed."; \
	fi
	@echo "Creating mustang-cli wrapper script..."
	@echo '#!/bin/bash' > bin/mustang-cli
	@echo 'DIR="$$(cd "$$(dirname "$${BASH_SOURCE[0]}")" && pwd)"' >> bin/mustang-cli
	@echo '"$$DIR/jre/bin/java" -jar "$$DIR/mustang-cli.jar" "$$@"' >> bin/mustang-cli
	@chmod +x bin/mustang-cli
	@echo "Mustang CLI setup complete!"

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
