.PHONY: build test test-lib test-cli test-all clean release release-dry help

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

# Create a new release (specify VERSION=vX.Y.Z or auto-increment)
release:
	@./scripts/release.sh $(VERSION)

# Preview release changes without making them
release-dry:
	@DRY_RUN=1 ./scripts/release.sh $(VERSION)

# Show available targets
help:
	@echo "Agar Makefile Targets:"
	@echo ""
	@echo "Build & Test:"
	@echo "  build         Build the Agar CLI"
	@echo "  test          Run all tests (alias for test-all)"
	@echo "  test-lib      Run library tests only"
	@echo "  test-cli      Run CLI tests only"
	@echo "  test-all      Run all tests (excluding examples)"
	@echo "  clean         Remove build artifacts"
	@echo ""
	@echo "Release:"
	@echo "  release       Create a new release"
	@echo "                Usage: make release VERSION=v1.2.3"
	@echo "                       make release (auto-increment patch)"
	@echo "  release-dry   Preview release without making changes"
	@echo ""
	@echo "Other:"
	@echo "  help          Show this help message"
	@echo ""
