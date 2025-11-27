.PHONY: help build run dev test clean templ install deps

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

deps: ## Install dependencies
	@echo "Installing dependencies..."
	@go mod download
	@go install github.com/a-h/templ/cmd/templ@latest
	@go install github.com/air-verse/air@latest

templ: ## Generate templ files
	@echo "Generating templates..."
	@$(shell go env GOPATH)/bin/templ generate

build: templ ## Build the application
	@echo "Building zentube..."
	@go build -o zentube ./cmd/zentube

run: build ## Build and run the application
	@echo "Running zentube..."
	@./zentube

dev: ## Run with hot reload (Air)
	@echo "Starting development server with hot reload..."
	@air -c .air.toml

test: ## Run tests
	@go test -v ./...

test-coverage: ## Run tests with coverage
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -f zentube
	@rm -rf tmp/
	@find . -name "*_templ.go" -type f -delete

fmt: ## Format code
	@go fmt ./...
	@templ fmt .

.DEFAULT_GOAL := help
