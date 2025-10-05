# Makefile for arrest-go project
# Supports both root module and gin submodule

.PHONY: help test test-verbose test-race test-coverage build clean lint fmt vet mod-tidy mod-verify install-tools bench examples

# Default target
help: ## Show this help message
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Test targets
test: ## Run tests in all modules
	@echo "Running tests in root module..."
	go test ./...
	@echo "Running tests in gin module..."
	cd gin && go test ./...

test-verbose: ## Run tests with verbose output in all modules
	@echo "Running verbose tests in root module..."
	go test -v ./...
	@echo "Running verbose tests in gin module..."
	cd gin && go test -v ./...

test-race: ## Run tests with race detection in all modules
	@echo "Running race tests in root module..."
	go test -race ./...
	@echo "Running race tests in gin module..."
	cd gin && go test -race ./...

test-coverage: ## Run tests with coverage in all modules
	@echo "Running coverage tests in root module..."
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out
	@echo "Running coverage tests in gin module..."
	cd gin && go test -coverprofile=coverage.out ./...
	cd gin && go tool cover -func=coverage.out

bench: ## Run benchmarks in all modules
	@echo "Running benchmarks in root module..."
	go test -bench=. ./...
	@echo "Running benchmarks in gin module..."
	cd gin && go test -bench=. ./...

# Build targets
build: ## Build all packages
	@echo "Building root module..."
	go build ./...
	@echo "Building gin module..."
	cd gin && go build ./...

clean: ## Clean build artifacts and test caches
	@echo "Cleaning root module..."
	go clean -cache -testcache -modcache
	rm -f coverage.out
	@echo "Cleaning gin module..."
	cd gin && go clean -cache -testcache
	cd gin && rm -f coverage.out

# Code quality targets
lint: ## Run linters on all modules
	@echo "Running linters on root module..."
	golangci-lint run
	@echo "Running linters on gin module..."
	cd gin && golangci-lint run

fmt: ## Format code in all modules
	@echo "Formatting root module..."
	go fmt ./...
	@echo "Formatting gin module..."
	cd gin && go fmt ./...

vet: ## Run go vet on all modules
	@echo "Running go vet on root module..."
	go vet ./...
	@echo "Running go vet on gin module..."
	cd gin && go vet ./...

# Module management targets
mod-tidy: ## Run go mod tidy in all modules
	@echo "Running go mod tidy in root module..."
	go mod tidy
	@echo "Running go mod tidy in gin module..."
	cd gin && go mod tidy

mod-verify: ## Verify modules in all modules
	@echo "Verifying root module..."
	go mod verify
	@echo "Verifying gin module..."
	cd gin && go mod verify

mod-download: ## Download dependencies for all modules
	@echo "Downloading dependencies for root module..."
	go mod download
	@echo "Downloading dependencies for gin module..."
	cd gin && go mod download

# Development tools
install-tools: ## Install development tools
	@echo "Installing golangci-lint..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "Installing goimports..."
	go install golang.org/x/tools/cmd/goimports@latest

# Example targets
examples: ## Run all examples
	@echo "Running main petstore example..."
	go run examples/petstore.go
	@echo "Running gin handler example..."
	cd gin && go run examples/petstore/handler/petstore.go
	@echo "Running gin call example..."
	cd gin && go run examples/petstore/call/petstore.go

# Composite targets
check: fmt vet lint ## Run all code quality checks
	@echo "All code quality checks completed!"

test-all: test test-race test-coverage ## Run all types of tests
	@echo "All tests completed!"

ci: check test-all ## Run full CI pipeline locally
	@echo "CI pipeline completed!"

# Development workflow
dev-setup: install-tools mod-download ## Setup development environment
	@echo "Development environment setup completed!"

quick-check: fmt vet test ## Quick development check (format, vet, test)
	@echo "Quick check completed!"

# PR management
retidy-prs: ## Run retidy-pr on all PRs with failed tests
	@echo "Processing PRs with failed tests..."
	./scripts/retidy-prs