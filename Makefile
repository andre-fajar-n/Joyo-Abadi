# Makefile for Joyo Abadi

.PHONY: help test test-unit test-integration test-coverage test-race test-bench clean build run dev

# Default target
help:
	@echo "Available commands:"
	@echo "  make test          - Run all tests"
	@echo "  make test-unit     - Run unit tests only"
	@echo "  make test-integration - Run integration tests only"
	@echo "  make test-coverage - Run tests with coverage report"
	@echo "  make test-race     - Run tests with race detection"
	@echo "  make test-bench    - Run benchmark tests"
	@echo "  make clean         - Clean test artifacts"
	@echo "  make build         - Build the application"
	@echo "  make run           - Run the application"
	@echo "  make dev           - Run in development mode with air"

# Run all tests
test:
	@echo "🧪 Running all tests..."
	@./run-tests.sh

# Run unit tests only
test-unit:
	@echo "🔬 Running unit tests..."
	@go test -v -race ./models/... ./utils/... ./middleware/... ./controllers/...

# Run integration tests only
test-integration:
	@echo "🔗 Running integration tests..."
	@go test -v -race ./integration_test.go

# Run tests with coverage
test-coverage:
	@echo "📊 Running tests with coverage..."
	@go test -v -race -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run tests with race detection
test-race:
	@echo "🏃 Running tests with race detection..."
	@go test -race ./...

# Run benchmark tests
test-bench:
	@echo "⚡ Running benchmark tests..."
	@go test -bench=. -benchmem ./...

# Clean test artifacts
clean:
	@echo "🧹 Cleaning test artifacts..."
	@rm -f coverage*.out coverage.html benchmark_results.txt lint_results.txt
	@go clean -testcache

# Build the application
build:
	@echo "🔨 Building application..."
	@go build -o main .

# Run the application
run: build
	@echo "🚀 Running application..."
	@./main

# Development mode with air (if available)
dev:
	@if command -v air >/dev/null 2>&1; then \
		echo "🔥 Running in development mode with air..."; \
		air; \
	else \
		echo "⚠️  Air not found. Install with: go install github.com/cosmtrek/air@latest"; \
		echo "🚀 Running normally..."; \
		make run; \
	fi

# Install development dependencies
install-dev:
	@echo "📦 Installing development dependencies..."
	@go install github.com/cosmtrek/air@latest
	@go install golang.org/x/lint/golint@latest

# Lint the code
lint:
	@echo "🔍 Running linter..."
	@if command -v golint >/dev/null 2>&1; then \
		golint ./...; \
	else \
		echo "⚠️  golint not found. Install with: make install-dev"; \
	fi
	@go vet ./...

# Format the code
fmt:
	@echo "✨ Formatting code..."
	@go fmt ./...

# Tidy dependencies
tidy:
	@echo "📦 Tidying dependencies..."
	@go mod tidy

# Full check (format, lint, test)
check: fmt lint test
	@echo "✅ All checks completed!"

# Deploy to Railway
deploy:
	@echo "🚂 Deploying to Railway..."
	@./deploy-railway.sh
