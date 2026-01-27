.PHONY: build run test lint fmt clean help

# Build variables
BINARY_NAME=daily-dash
BINARY_PATH=bin/$(BINARY_NAME)
CMD_PATH=./cmd/daily-dash
GO=go
GOFLAGS=-v

# Default target
all: build

## build: Build the application
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p bin
	@$(GO) build $(GOFLAGS) -o $(BINARY_PATH) $(CMD_PATH)
	@echo "Build complete: $(BINARY_PATH)"

## run: Run the application
run: build
	@$(BINARY_PATH)

## test: Run tests
test:
	@echo "Running tests..."
	@$(GO) test -v -race -coverprofile=coverage.out ./...

## test-coverage: Run tests with coverage report
test-coverage: test
	@$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

## lint: Run linter
lint:
	@echo "Running linter..."
	@which golangci-lint > /dev/null || (echo "golangci-lint not installed. Run: curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b \$$(go env GOPATH)/bin" && exit 1)
	@golangci-lint run ./...

## fmt: Format code
fmt:
	@echo "Formatting code..."
	@$(GO) fmt ./...
	@echo "Code formatted"

## clean: Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html
	@echo "Clean complete"

## tidy: Tidy go modules
tidy:
	@echo "Tidying go modules..."
	@$(GO) mod tidy
	@echo "Modules tidied"

## help: Show this help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@sed -n 's/^##//p' Makefile | column -t -s ':' | sed -e 's/^/ /'
