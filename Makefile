all: deps vet test build

build:
	@mkdir -p bin
	@go build -o bin/newspaper ./cmd

run:
	@echo "Running..."
	@go run ./cmd

vet:
	@echo "Vetting code..."
	@go vet ./...

fmt:
	@echo "Formatting code..."
	@go fmt ./...

test:
	@echo "Running tests..."
	@go test -v ./...

clean:
	@echo "Cleaning build artifacts..."
	@rm -rf bin/newspaper

deps:
	@echo "Installing dependencies..."
	@go mod download
	@go mod tidy

help:
	@echo "Available targets:"
	@echo "  all    - Run vet and build"
	@echo "  build  - Build the application"
	@echo "  run    - Run the application"
	@echo "  vet    - Vet the code"
	@echo "  fmt    - Format the code"
	@echo "  test   - Run tests"
	@echo "  clean  - Clean build artifacts"
	@echo "  deps   - Install dependencies"
	@echo ""
	@echo "  help   - Show this help message"
