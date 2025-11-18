#!/usr/bin/env bash
set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Helper functions
info() {
    echo -e "${GREEN}==>${NC} $*"
}

warn() {
    echo -e "${YELLOW}Warning:${NC} $*"
}

error() {
    echo -e "${RED}Error:${NC} $*" >&2
    exit 1
}

usage() {
    cat <<EOF
Usage: $0 [VERSION]

Create a new release of the Agar library and CLI.

Arguments:
  VERSION    Version to release (e.g., v0.1.0, 0.1.0)
             If not provided, will auto-increment patch version

Examples:
  $0 v1.2.3     # Release version 1.2.3
  $0 1.2.3      # Release version 1.2.3 (v prefix added automatically)
  $0            # Auto-increment patch version from latest tag

Environment Variables:
  SKIP_TESTS    Set to '1' to skip running tests
  DRY_RUN       Set to '1' to preview changes without making them
  GITHUB_TOKEN  Required for creating GitHub releases (uses gh CLI)

EOF
    exit 1
}

# Parse arguments
VERSION="${1:-}"

if [[ "$VERSION" == "-h" || "$VERSION" == "--help" ]]; then
    usage
fi

# Change to repository root
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR/.."

# Check if we're in a git repository
if ! git rev-parse --git-dir > /dev/null 2>&1; then
    error "Not in a git repository"
fi

# Check if working directory is clean
if [[ -n $(git status --porcelain) ]]; then
    error "Working directory is not clean. Please commit or stash your changes."
fi

# Check if gh CLI is installed
if ! command -v gh &> /dev/null; then
    error "GitHub CLI (gh) is required but not installed. Install from https://cli.github.com/"
fi

# Ensure we're on main branch
CURRENT_BRANCH=$(git branch --show-current)
if [[ "$CURRENT_BRANCH" != "main" ]]; then
    warn "You are on branch '$CURRENT_BRANCH', not 'main'"
    read -p "Continue anyway? (y/N) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
fi

# Determine version
if [[ -z "$VERSION" ]]; then
    info "No version specified, auto-incrementing patch version..."

    # Get latest tag
    LATEST_TAG=$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")
    info "Latest tag: $LATEST_TAG"

    # Increment patch version
    if [[ $LATEST_TAG =~ ^v([0-9]+)\.([0-9]+)\.([0-9]+)$ ]]; then
        MAJOR="${BASH_REMATCH[1]}"
        MINOR="${BASH_REMATCH[2]}"
        PATCH="${BASH_REMATCH[3]}"
        NEW_PATCH=$((PATCH + 1))
        VERSION="v${MAJOR}.${MINOR}.${NEW_PATCH}"
    else
        VERSION="v0.1.0"
    fi
    info "Auto-incremented version: $VERSION"
else
    # Ensure version has v prefix
    if [[ ! "$VERSION" =~ ^v ]]; then
        VERSION="v$VERSION"
    fi

    # Validate version format
    if [[ ! "$VERSION" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
        error "Invalid version format: $VERSION (expected vX.Y.Z)"
    fi
fi

CLI_TAG="cmd/agar/$VERSION"

info "Release version: $VERSION"
info "CLI tag: $CLI_TAG"

# Check if version already exists
if git rev-parse "$VERSION" >/dev/null 2>&1; then
    error "Version $VERSION already exists"
fi

if git rev-parse "$CLI_TAG" >/dev/null 2>&1; then
    error "CLI tag $CLI_TAG already exists"
fi

# Dry run mode
if [[ "${DRY_RUN:-0}" == "1" ]]; then
    info "DRY RUN MODE - No changes will be made"
fi

# Run tests
if [[ "${SKIP_TESTS:-0}" != "1" ]]; then
    info "Running tests..."
    go work sync
    go test ./tools/... ./tui/... ./commands/... -v
    info "Tests passed ✓"
else
    warn "Skipping tests (SKIP_TESTS=1)"
fi

# Confirm before proceeding
echo
info "Ready to release:"
echo "  - Library: $VERSION"
echo "  - CLI: $CLI_TAG"
echo
read -p "Continue with release? (y/N) " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    info "Release cancelled"
    exit 0
fi

# Function to run commands (supports dry run)
run_cmd() {
    if [[ "${DRY_RUN:-0}" == "1" ]]; then
        echo "[DRY RUN] $*"
    else
        "$@"
    fi
}

# Release Library
info "Releasing library $VERSION..."

run_cmd git tag -a "$VERSION" -m "Release $VERSION"
run_cmd git push origin "$VERSION"

if [[ "${DRY_RUN:-0}" != "1" ]]; then
    gh release create "$VERSION" \
        --title "Agar $VERSION" \
        --notes "Release $VERSION

## Installation

\`\`\`bash
go get github.com/geoffjay/agar@$VERSION
\`\`\`

## What's Changed

See the [full changelog](https://github.com/geoffjay/agar/compare/$VERSION...main) for details." \
        --draft=false

    info "Library released ✓"
else
    info "[DRY RUN] Would create GitHub release for $VERSION"
fi

# Update CLI for release
info "Updating CLI to use library $VERSION..."

cd cmd/agar

# Remove replace directive
run_cmd go mod edit -dropreplace=github.com/geoffjay/agar

# Update to use new library version
run_cmd go mod edit -require=github.com/geoffjay/agar@$VERSION

# Tidy dependencies
if [[ "${DRY_RUN:-0}" != "1" ]]; then
    GOWORK=off go mod tidy
fi

# Verify it builds
info "Verifying CLI builds..."
if [[ "${DRY_RUN:-0}" != "1" ]]; then
    GOWORK=off go build -v -o agar .
    info "CLI build successful ✓"
else
    info "[DRY RUN] Would verify CLI build"
fi

# Commit CLI changes
info "Committing CLI changes..."
run_cmd git add go.mod go.sum
run_cmd git commit -m "chore(cli): update to agar $VERSION for release" || true
run_cmd git push origin "$CURRENT_BRANCH"

cd ../..

# Release CLI
info "Releasing CLI $CLI_TAG..."

run_cmd git tag -a "$CLI_TAG" -m "Release CLI $VERSION"
run_cmd git push origin "$CLI_TAG"

# Build CLI binary for release
cd cmd/agar
if [[ "${DRY_RUN:-0}" != "1" ]]; then
    GOWORK=off go build -v -ldflags="-s -w" -o agar .

    gh release create "$CLI_TAG" \
        --title "Agar CLI $VERSION" \
        --notes "Agar CLI $VERSION

## Installation

\`\`\`bash
go install github.com/geoffjay/agar/cmd/agar@$VERSION
\`\`\`

## Changes

- Uses agar library $VERSION
- AI-powered project scaffolding with BAML
- Interactive TUI interface

## Requirements

Set \`ANTHROPIC_API_KEY\` environment variable for AI features." \
        --draft=false \
        agar

    info "CLI released ✓"
else
    info "[DRY RUN] Would create GitHub release for $CLI_TAG"
fi

cd ../..

# Restore replace directive
info "Restoring replace directive for development..."

cd cmd/agar

run_cmd go mod edit -replace=github.com/geoffjay/agar=../..

if [[ "${DRY_RUN:-0}" != "1" ]]; then
    # Verify workspace still works
    cd ../..
    go work sync
    cd cmd/agar
fi

run_cmd git add go.mod
run_cmd git commit -m "chore(cli): restore replace directive for development"
run_cmd git push origin "$CURRENT_BRANCH"

cd ../..

# Success!
echo
info "Release complete! 🎉"
echo
echo "Library: https://github.com/geoffjay/agar/releases/tag/$VERSION"
echo "CLI: https://github.com/geoffjay/agar/releases/tag/$CLI_TAG"
echo
info "Users can now install with:"
echo "  go get github.com/geoffjay/agar@$VERSION"
echo "  go install github.com/geoffjay/agar/cmd/agar@$VERSION"
