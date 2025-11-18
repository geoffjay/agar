# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Agar is a Go framework for building AI agent applications with TUI (Terminal User Interface) components and tool management. The project is structured as:

- **Library packages** (`tools/`, `tui/`, `commands/`) - Reusable components for building AI applications
- **CLI tool** (`cmd/agar/`) - Scaffolding tool for creating new Agar projects

## Development Commands

### Building

```bash
# Build the CLI tool
make build
# Output: cmd/agar/agar

# Build library packages only
go build -v ./tools/... ./tui/...

# Build CLI directly
cd cmd/agar && go build -v -o agar .
```

### Testing

```bash
# Run all tests
make test

# Run library tests only (tools, tui, commands packages)
make test-lib
go test ./tools/... ./tui/... -v

# Run CLI tests only
make test-cli
cd cmd/agar && go test ./... -v

# Run a single test
go test -run TestToolRegistry_Register ./tools/... -v

# Run tests with race detection and coverage
go test ./tools/... ./tui/... -v -race -coverprofile=coverage.out -covermode=atomic
go tool cover -func=coverage.out
```

### Linting

```bash
# Run golangci-lint (if installed)
golangci-lint run --timeout=5m
```

### Cleanup

```bash
# Remove build artifacts
make clean
```

## Go Workspaces Architecture

**Critical:** This project uses Go Workspaces with **two separate modules**:

1. **Root module** (`./go.mod`) - Library packages: `tools/`, `tui/`, `commands/`

   - Minimal dependencies (Bubble Tea, Lipgloss only)
   - Used by library consumers

2. **CLI module** (`cmd/agar/go.mod`) - CLI tool
   - Additional dependencies: Cobra, BAML
   - Contains `replace github.com/geoffjay/agar => ../..` for local development

**Workspace file** (`go.work`):

```go
use (
    .
    ./cmd/agar
)
```

### When Adding Dependencies

- **Library code** (tools/tui/commands): Add to root `go.mod` - keep lightweight
- **CLI code** (cmd/agar): Add to `cmd/agar/go.mod`
- After changes, run: `go work sync && go mod tidy`

### Building Considerations

When building CLI code that imports the library:

- Use workspace context: `go build -v -o agar .` (from `cmd/agar/`)
- Do NOT use `GOWORK=off` during local development - it breaks the replace directive

## High-Level Architecture

### Tools Package (`tools/`)

**Registry-based plugin system** for AI agent tools:

- **Core interface**: `tools/tool.go` - defines `Tool` interface with `Name()`, `Description()`, `Execute()`, `Validate()`, `Schema()`
- **Registry**: `tools/registry.go` - thread-safe tool registration and lookup (uses RWMutex)
- **Built-in tools** (11 total): Read, Write, Delete, List, Glob, Fetch, Download, Search, Grep, Shell, TaskList

**Pattern**: Each tool is a standalone type implementing the `Tool` interface. Registration is dynamic via `tools.NewToolRegistry()`.

Example:

```go
registry := tools.NewToolRegistry()
registry.Register(tools.NewReadTool())
tool, _ := registry.Get("read")
result, _ := tool.Execute(ctx, jsonParams)
```

### TUI Package (`tui/`)

**Bubble Tea-based UI framework** following Model/Update/View pattern:

- **Application** (`tui/application.go`) - Main container implementing `tea.Model`

  - Manages content display, scrolling, command system integration
  - Methods: `AddLine()`, `AddLines()`, `Clear()`, `SetContent()`
  - Auto-scrolls to bottom when content exceeds viewport

- **Panel** (`tui/panel.go`) - Configurable content area with borders/margins/padding

  - Border styles: NoBorder, SingleBorder, DoubleBorder, RoundedBorder
  - Selective borders: BorderTop, BorderBottom, BorderLeft, BorderRight

- **PromptModel** (`tui/input_prompt.go`) - Input component with command autocompletion

  - Single/multi-line modes, history navigation
  - Emits `PromptSubmitMsg` on Enter
  - Shows slash command dropdown when typing "/"

- **Layout** (`tui/layout.go`) - Container for vertical/horizontal component arrangement

- **Input components**: InputText, InputYesNo, InputOptions, InputMultiSelect, IterativeForm

**Message Flow**: User input → KeyMsg → PromptModel → PromptSubmitMsg → Application → Command execution → View update

### Commands Package (`commands/`)

**Slash command system** for interactive applications:

- **Command interface** (`commands/command.go`) - defines `Command` interface

  - Methods: `Name()`, `Description()`, `Usage()`, `Execute()`, `Aliases()`
  - Uses `ApplicationState` interface for loose coupling

- **Manager** (`commands/manager.go`) - Facade combining Registry, Handler, Loader

  - `Initialize()` - loads built-in and file-based commands
  - `Handle()` - parses and executes commands
  - `GetCompletions()` - for autocomplete

- **Registry** (`commands/registry.go`) - Thread-safe command registration with alias support

- **Built-in commands** (`commands/builtin.go`): exit, help, export, import, clear

- **File-based commands** (`commands/loader.go`) - Load from YAML/JSON files
  - Default paths: `$HOME/.agar/commands`, `./commands`, `.agar/commands`

**Command Execution Flow**: Input → Handler.Handle() → Registry.Get() → Command.Execute() → State modification

### CLI Integration (`cmd/agar/`)

- **Entry point**: `main.go` → `cmd/root.go` (Cobra-based)
- **TUI runner**: `internal/app/tui.go` - Composes Application + PromptModel
- **Commands**: Uses `commands.Manager` for slash command support

When launched without subcommands, runs interactive TUI with prompt and command system enabled.

## Key Design Patterns

1. **Registry Pattern** - Used for tools (`tools/registry.go`) and commands (`commands/registry.go`)

   - Thread-safe with RWMutex
   - Dynamic registration enables plugin architecture

2. **Interface Segregation** - Loose coupling via interfaces

   - `Tool` interface in tools package
   - `Command` and `ApplicationState` interfaces in commands package

3. **Bubble Tea Model/Update/View** - All TUI components follow this pattern

   - Immutable-style updates (return new model + cmd)
   - Message-based event handling

4. **Facade Pattern** - `commands.Manager` provides simplified API over Registry/Handler/Loader

5. **Composite Pattern** - Layout composability via `LayoutComponent` interface

## Testing Guidelines

- Write tests in `*_test.go` files alongside implementation
- Use table-driven tests for multiple scenarios
- Mock external dependencies (filesystem, network)
- Run tests with `-race` flag to detect race conditions
- Aim for high coverage on core packages (tools, tui, commands)

## File-Based Commands

Applications can load custom commands from YAML/JSON files. Example:

```yaml
# .agar/commands/custom.yaml
name: greet
description: Greet the user
usage: "/greet [name]"
aliases:
  - hello
  - hi
```

Commands are automatically discovered from configured paths.

## Release Process

The project supports both manual and automated releases:

### Manual Release (Recommended)

Use the release script for manual control over releases:

```bash
# Release a specific version
make release VERSION=v1.2.3

# Auto-increment patch version from latest tag
make release

# Preview changes without making them
make release-dry VERSION=v1.2.3

# Skip tests during release (not recommended)
SKIP_TESTS=1 make release VERSION=v1.2.3
```

The release script (`scripts/release.sh`):
1. Validates version format and checks working directory is clean
2. Runs tests to ensure code quality
3. Tags and releases library with `vX.Y.Z`
4. Updates CLI `go.mod` (removes `replace`, adds library version)
5. Tags and releases CLI with `cmd/agar/vX.Y.Z`
6. Restores `replace` directive for development
7. Creates GitHub releases for both library and CLI

**Requirements**:
- Clean git working directory (no uncommitted changes)
- GitHub CLI (`gh`) installed and authenticated
- On `main` branch (or will warn)

### Automated Release via GitHub Actions

**Via Release Branch**:

```bash
git checkout -b release/v1.2.3
git push origin release/v1.2.3
# Create PR to main → merge triggers release with v1.2.3
```

**Auto-increment**:

```bash
# Any PR merged to main without release/ prefix auto-increments patch version
```

The automated workflow mirrors the manual script process.

## Important Considerations

- **Dependency management**: Keep library packages lightweight - only add dependencies to root `go.mod` when absolutely necessary
- **Workspace awareness**: Always work within workspace context (`go.work` active) during development
- **Thread safety**: Registries use RWMutex - maintain this pattern when adding concurrent access
- **Bubble Tea patterns**: All TUI components must implement `tea.Model` interface correctly
- **Interface contracts**: Commands and tools must properly implement their interfaces
- **Message passing**: Use Bubble Tea's message system for component communication - avoid direct coupling
