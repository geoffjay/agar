# Agar CLI

A command-line tool for scaffolding and managing applications built with the Agar framework.

## Installation

```bash
go install github.com/geoffjay/agar/cmd/agar@latest
```

## Usage

```bash
# Show help
agar --help

# Initialize a new agar project (coming soon)
agar init my-project

# Add a new tool to your project (coming soon)
agar add tool custom-tool

# List available components (coming soon)
agar list
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
