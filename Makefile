# Makefile for PromViz

.PHONY: build test test-unit test-integration test-coverage bench clean lint fmt vet

# Build variables
BINARY_NAME=promviz
GO_FILES=$(shell find . -name "*.go" -not -path "./vendor/*")

# Build the application
build:
	go build -o $(BINARY_NAME) main.go

# Install dependencies
deps:
	go mod download
	go mod tidy

# Run all tests
test: test-unit test-integration

# Run unit tests only
test-unit:
	go test -v ./internal/...

# Run integration tests (requires build tags)
test-integration:
	go test -v -tags=integration .

# Run tests with coverage
test-coverage:
	go test -v -coverprofile=coverage.out ./internal/...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run benchmarks
bench:
	go test -bench=. -benchmem ./internal/...
	go test -bench=. -benchmem -tags=integration .

# Clean build artifacts
clean:
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html

# Lint the code
lint:
	golangci-lint run

# Format the code
fmt:
	go fmt ./...

# Vet the code
vet:
	go vet ./...

# Check code quality
check: fmt vet lint test-unit

# Install development tools
dev-tools:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run the application with default config
run:
	./$(BINARY_NAME)

# Run the application with Prometheus config
run-prometheus:
	./$(BINARY_NAME) --config queries.yaml

# Run the application with InfluxDB config
run-influxdb:
	./$(BINARY_NAME) --config queries-influxdb.yaml

# Help
help:
	@echo "Available targets:"
	@echo "  build           Build the application"
	@echo "  deps            Install dependencies"
	@echo "  test            Run all tests"
	@echo "  test-unit       Run unit tests only"
	@echo "  test-integration Run integration tests"
	@echo "  test-coverage   Run tests with coverage report"
	@echo "  bench           Run benchmarks"
	@echo "  clean           Clean build artifacts"
	@echo "  lint            Lint the code"
	@echo "  fmt             Format the code"
	@echo "  vet             Vet the code"
	@echo "  check           Run fmt, vet, lint, and unit tests"
	@echo "  dev-tools       Install development tools"
	@echo "  run             Run with default config"
	@echo "  run-prometheus  Run with Prometheus config"
	@echo "  run-influxdb    Run with InfluxDB config"
	@echo "  help            Show this help message"
