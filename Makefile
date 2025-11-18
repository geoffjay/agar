.PHONY: build test test-lib test-cli test-all clean help

# Build the Agar CLI
build:
	@echo "Building Agar CLI..."
	@cd cmd/agar && go build -v -o agar .
	@echo "✓ CLI built: cmd/agar/agar"

# Run library tests (excluding examples)
test-lib:
	@echo "Running library tests..."
	@go test ./tools/... ./tui/... -v

# Run CLI tests only
test-cli:
	@echo "Running CLI tests..."
	@cd cmd/agar && go test ./... -v

# Run all tests (excluding examples)
test-all:
	@echo "Running all tests..."
	@go test ./tools/... ./tui/... -v
	@cd cmd/agar && go test ./... -v

# Alias for test-all
test: test-all

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -f cmd/agar/agar
	@rm -f examples/*/[!.]*.go.out examples/*/*_demo
	@echo "✓ Clean complete"

# Show available targets
help:
	@echo "Agar Makefile Targets:"
	@echo ""
	@echo "  build      Build the Agar CLI"
	@echo "  test       Run all tests (alias for test-all)"
	@echo "  test-lib   Run library tests only"
	@echo "  test-cli   Run CLI tests only"
	@echo "  test-all   Run all tests (excluding examples)"
	@echo "  clean      Remove build artifacts"
	@echo "  help       Show this help message"
	@echo ""
