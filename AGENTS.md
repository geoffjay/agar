# AGENTS.md

## Project Overview
This is an AI agent library providing TUI components and tools for building AI applications. The codebase consists of:
- 7,320+ lines of Go code
- Comprehensive tool framework with 17 tool implementations
- 79+ test functions ensuring reliability
- TUI components built with Bubble Tea framework

## Build/Lint/Test Commands

```bash
# Build the project
go build ./...

# Run all tests
# Note: The project now has extensive test coverage with 79+ test functions
go test ./...

# Run a specific test file (pattern)
go test -v ./path/to/package -run TestName

# Run tests for specific tool
go test -v ./tools -run TestGlob
go test -v ./tools -run TestTaskList
go test -v ./tools -run TestRegistry

# Format code
go fmt ./...

# Check for formatting issues (if golangci-lint is installed)
golangci-lint run

# Install/update dependencies
go mod tidy

# Run example applications
go run examples/application.go
go run examples/input_components.go

# Tool usage examples in code
go run examples/border.go
go run examples/footer.go
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
- Examples available in: /Users/geoff/Projects/agar/tui/
  - application.go - Main application components
  - footer.go - Footer functionality
  - input_text.go, input_yesno.go - Input components
  - iterative_form.go - Form handling
  - layout.go, panel.go, style.go - UI layout and styling

### AI Tools Framework

The project includes a comprehensive tools framework for AI agents:

**Tool Interface:** All tools implement the `Tool` interface in `/Users/geoff/Projects/agar/tools/tool.go`:
- Name() - Tool name
- Description() - What the tool does
- Execute() - Run the tool with parameters
- Validate() - Validate tool parameters
- Schema() - JSON schema for parameters

**Tool Registry:** Central management in `/Users/geoff/Projects/agar/tools/registry.go`
- Register/UnRegister tools
- Thread-safe operations
- Tool discovery and validation

**Available Tools:** (17 implementations in /Users/geoff/Projects/agar/tools/)
- **File Operations:** read, write, delete, list, glob
- **System:** shell execution, task management
- **Search:** file pattern matching

**Tool Examples:**
```go
// Create registry
registry := tools.NewToolRegistry()

// Register tools
registry.Register(tools.NewReadTool())
registry.Register(tools.NewGlobTool())
registry.Register(tools.NewTaskListTool())

// Execute tool
result, err := registry.Execute("read", []byte(`{"path": "file.txt"}`))
```

### Dependencies

- This project uses bubbletea for TUI framework
- Uses lipgloss for styling
- All dependencies are managed through go.mod

### Documentation

- Comment all exported functions and types
- Use clear, concise descriptions
- Include examples when helpful
- Refer to `/Users/geoff/Projects/agar/docs/` for detailed framework documentation:
  - `tools.md` - Comprehensive tools framework documentation
  - `tui.md` - TUI components and patterns

## Development Guidelines

### Getting Started
1. Review existing tool implementations in `/Users/geoff/Projects/agar/tools/`
2. Follow TUI patterns in `/Users/geoff/Projects/agar/tui/`
3. Run tests to verify your changes: `go test ./tools -v`
4. Check project structure with: `go run examples/application.go`

### Adding New Tools
1. Implement the `Tool` interface
2. Add comprehensive tests
3. Register in the tool registry
4. Update documentation in `docs/tools.md`

### Common Patterns
- Error wrapper functions for consistent error handling
- Context-aware execution for cancellation support
- JSON schema validation for tool parameters
- Thread-safe operations in shared components
