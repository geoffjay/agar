# AGENTS.md

## Build/Lint/Test Commands

```bash
# Build the project
go build ./...

# Run all tests (currently no tests in project)
go test ./...

# Run a specific test file (pattern)
go test -v ./path/to/package -run TestName

# Format code
go fmt ./...

# Check for formatting issues (if golangci-lint is installed)
golangci-lint run

# Install/update dependencies
go mod tidy

# Run example applications
go run examples/input_components.go
```

## Code Style Guidelines

### Imports

- Use standard Go imports
- Group imports in order: standard library, third-party, local packages
- Use aliases only when necessary to avoid conflicts

### Formatting

- Use `go fmt` for all code formatting
- Line length: 100 characters maximum
- Use tabs for indentation (not spaces)

### Types and Naming

- Use camelCase for variables and functions
- Use PascalCase for exported names
- Use descriptive names over abbreviated ones
- Constants should be UPPER_SNAKE_CASE

### Error Handling

- Always handle errors explicitly
- Use early returns for error cases
- Wrap errors with context when propagating

### TUI Components

- All components should implement the tea.Model interface
- Include Init(), Update(), and View() methods
- Provide GetAnswer() and IsDone() methods for data retrieval
- Use lipgloss for styling
- Follow the pattern in existing components (input_text.go, input_yesno.go)

### Dependencies

- This project uses bubbletea for TUI framework
- Uses lipgloss for styling
- All dependencies are managed through go.mod

### Documentation

- Comment all exported functions and types
- Use clear, concise descriptions
- Include examples when helpful
