.PHONY: run worker test fmt vet docker-up docker-down build install help

# Default target
all: fmt vet run

# Run the application
run:
	@echo "Starting server..."
	@go run main.go

# Run in worker mode
worker:
	@echo "Starting worker..."
	@go run main.go --worker

# Run tests
test:
	@echo "Running tests..."
	@go test -v -race -cover ./...

# Run linter
lint:
	@echo "Running linter..."
	@golangci-lint run || true

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# Run go vet
vet:
	@echo "Running go vet..."
	@go vet ./...

# Update dependencies
deps:
	@echo "Updating dependencies..."
	@go mod tidy
	@go mod download

# Build the application
build:
	@echo "Building..."
	@go build -o bin/app main.go

# Install dependencies
install:
	@go mod download

# Clean artifacts
clean:
	@rm -rf bin/
	@rm -f app

# Docker commands
docker-up:
	@echo "Starting Docker containers..."
	@docker-compose up -d

docker-down:
	@echo "Stopping Docker containers..."
	@docker-compose down

docker-logs:
	@docker-compose logs -f

docker-reset:
	@echo "Resetting Docker containers..."
	 @docker-compose down -v
	@docker-compose up -d

# Development utilities
watch:
	@echo "Watching for changes..."
	@go install github.com/cosmtrek/air@latest 2>/dev/null || true
	@air || go run main.go

# Install development tools
dev-tools:
	@echo "Installing development tools..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest || true

# Show this help
help:
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

# Quick setup for new developers
setup: install dev-tools docker-up
	@echo ""
	@echo "Setup complete!"
	@echo "Copy .env.example to .env and configure as needed:"
	@echo "  cp .env.example .env"
	@echo ""
	@echo "Then run:"
	@echo "  make run"
