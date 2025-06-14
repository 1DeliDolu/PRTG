.PHONY: run run-frontend build clean

# Plugin binary name
BINARY_NAME=prtg-plugin

# Default target
all: build

# Build the plugin
build:
    go build -o dist/$(BINARY_NAME) ./cmd/plugin

# Run the backend
run:
    go run ./cmd/plugin/main.go

# Run the frontend
run-frontend:
    yarn dev

# Clean build artifacts
clean:
    rm -rf dist/
    rm -rf node_modules/

# Install dependencies
deps:
    go mod download
    yarn install