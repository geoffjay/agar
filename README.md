# Agar

A comprehensive framework for building AI agent applications with TUI components and tool management.

## Overview

Agar provides two main components:

- **Library** (`tools`, `tui` packages) - Reusable components for building AI applications
- **CLI** (`cmd/agar`) - Scaffolding tool for creating new Agar projects

## Installation

### As a Library

```bash
go get github.com/geoffjay/agar
```

### As a CLI Tool

```bash
go install github.com/geoffjay/agar/cmd/agar@latest
```

## Quick Start

### Using the Library

```go
import (
    "github.com/geoffjay/agar/tools"
    "github.com/geoffjay/agar/tui"
)

// Create a TUI application
app := tui.NewApplication(tui.ApplicationConfig{
    Title:   "MyApp",
    Version: "1.0.0",
})

// Register and use tools
registry := tools.NewToolRegistry()
registry.Register(tools.NewReadTool())
```

### Using the CLI

```bash
# Initialize a new project
agar init my-project

# See available commands
agar --help
```

## Project Structure

This project uses Go Workspaces to maintain clean dependency separation:

- **Root module** (`go.mod`) - Library packages (tools, tui)
- **CLI module** (`cmd/agar/go.mod`) - CLI tool with Cobra
- **Workspace** (`go.work`) - Coordinates both modules

**Benefit**: Library users don't get CLI dependencies, keeping imports lightweight.

## Packages

### Tools (`tools/`)

AI agent tools framework with 11 production-ready tools:

- **File Operations**: Read, Write, Delete, List, Glob
- **Web Access**: Fetch, Download
- **Search**: Search, Grep
- **System**: Shell, TaskList

See [tools documentation](docs/tools.md) for details.

### TUI (`tui/`)

Terminal UI components built on Bubble Tea:

- Application framework with panels, layouts, and footers
- Input components (text, yes/no, options, multi-select)
- Iterative forms for Q&A sessions

## Documentation

- [Tools Framework](docs/tools.md) - AI agent tools
- [CLI README](cmd/agar/README.md) - CLI tool
- [TODO](docs/todo.md) - Future enhancements

## Development

### Local Development

The project uses Go Workspaces. The CLI module has a `replace` directive for local development:

```go
// cmd/agar/go.mod
replace github.com/geoffjay/agar => ../..
```

This allows the CLI to use the local library code during development.

### Release Process

Releases are automated via GitHub Actions:

**Manual Release (Specific Version):**

```bash
git checkout -b release/v1.2.3
# Make changes if needed
git push origin release/v1.2.3
# Create PR to main
# Merge PR → Triggers release with v1.2.3
```

**Automatic Release (Patch Increment):**

```bash
git checkout -b feature/my-feature
# Make changes
git push origin feature/my-feature
# Create PR to main
# Merge PR → Triggers release with auto-incremented version
```

**What Happens Automatically:**

1. Library tagged with vX.Y.Z
2. CLI updated to use new library version
3. CLI tagged with cmd/agar/vX.Y.Z
4. GitHub releases created
5. Replace directive restored for development

## License

[MIT](LICENSE)
