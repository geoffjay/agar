# Agar CLI

A command-line tool for scaffolding and managing AI applications built with the Agar framework, powered by BAML for intelligent project generation.

## Installation

### From Source (Current)

```bash
git clone https://github.com/geoffjay/agar
cd agar/cmd/agar
go build -o agar .
```

### Via go install (After v0.1.0 Release)

```bash
go install github.com/geoffjay/agar/cmd/agar@latest
```

**Note**: The `replace` directive in `go.mod` is for local development with the workspace. It will be removed before releases to enable `go install` compatibility.

## Usage

```bash
# Launch interactive TUI
agar

# Show help
agar --help

# Initialize a new agar project with AI
agar init my-project

# Initialize with specific requirements
agar init my-agent -r "CLI tool for managing configuration files with TUI"

# Future commands (coming soon)
agar add tool custom-tool
agar list
```

### Requirements

For AI-powered project generation, set your API key:

```bash
export ANTHROPIC_API_KEY=your-key-here
```

## About

The Agar CLI is a separate module from the core Agar library, which means:
- Installing the CLI doesn't add extra dependencies to your library projects
- The CLI can be updated independently from the library
- Library users don't need to install the CLI

## Development

This CLI is part of the Agar workspace but maintains its own `go.mod` file for dependency isolation.

### Building from source

```bash
cd cmd/agar
go build -o agar .
```

### Dependencies

- `github.com/spf13/cobra` - CLI framework
- `github.com/spf13/viper` - Configuration management (optional)

## License

See the main Agar project LICENSE file.
