package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/geoffjay/agar/commands"
)

// InitCommand handles project initialization within the TUI
type InitCommand struct{}

func NewInitCommand() *InitCommand {
	return &InitCommand{}
}

func (c *InitCommand) Name() string {
	return "init"
}

func (c *InitCommand) Description() string {
	return "Initialize a new Agar project"
}

func (c *InitCommand) Usage() string {
	return "/init <project-name>"
}

func (c *InitCommand) Aliases() []string {
	return []string{}
}

func (c *InitCommand) Execute(ctx context.Context, args []string, state commands.ApplicationState) error {
	if len(args) == 0 {
		state.AddLine("Error: Project name required")
		state.AddLine("Usage: /init <project-name>")
		return nil
	}

	projectName := args[0]

	// Validate project name
	if strings.TrimSpace(projectName) == "" {
		state.AddLine("Error: Project name cannot be empty")
		return nil
	}

	// Check if directory already exists
	if _, err := os.Stat(projectName); err == nil {
		state.AddLine(fmt.Sprintf("Error: Directory '%s' already exists", projectName))
		return nil
	}

	// Create project directory
	if err := os.MkdirAll(projectName, 0755); err != nil {
		state.AddLine(fmt.Sprintf("Error creating directory: %v", err))
		return nil
	}

	// Create basic project structure
	dirs := []string{
		filepath.Join(projectName, "tools"),
		filepath.Join(projectName, "commands"),
		filepath.Join(projectName, "baml_src"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			state.AddLine(fmt.Sprintf("Error creating directory %s: %v", dir, err))
			return nil
		}
	}

	// Create main.go
	mainContent := fmt.Sprintf(`package main

import (
	"fmt"
	"os"

	"github.com/geoffjay/agar/tui"
	"github.com/geoffjay/agar/tools"
)

func main() {
	// Create tool registry
	toolRegistry := tools.NewToolRegistry()

	// Register built-in tools
	toolRegistry.Register(tools.NewReadTool())
	toolRegistry.Register(tools.NewWriteTool())
	toolRegistry.Register(tools.NewListTool())
	toolRegistry.Register(tools.NewShellTool())
	// Add more tools as needed...

	// Get current directory
	cwd, err := os.Getwd()
	if err != nil {
		cwd = "/"
	}

	// Create TUI application
	app := tui.NewApplication(tui.ApplicationConfig{
		Title:          "%s",
		Version:        "0.1.0",
		Mode:           "INTERACTIVE",
		Directory:      cwd,
		PanelMargin:    0,
		PanelPadding:   1,
		BorderStyle:    tui.NoBorder,
		EnableCommands: true,
		ToolRegistry:   toolRegistry,
	})

	// Add welcome message
	app.AddLine("")
	app.AddLine("Welcome to %s!")
	app.AddLine("")
	app.AddLine("Type /help to see available commands")
	app.AddLine("Type /tools to see available tools")
	app.AddLine("")

	// Run the application
	// TODO: Implement your application logic here
	fmt.Println("Application initialized. Run with: go run main.go")
}
`, projectName, projectName)

	mainPath := filepath.Join(projectName, "main.go")
	if err := os.WriteFile(mainPath, []byte(mainContent), 0644); err != nil {
		state.AddLine(fmt.Sprintf("Error creating main.go: %v", err))
		return nil
	}

	// Create go.mod
	goModContent := fmt.Sprintf(`module %s

go 1.21

require (
	github.com/geoffjay/agar v0.1.0
)
`, projectName)

	goModPath := filepath.Join(projectName, "go.mod")
	if err := os.WriteFile(goModPath, []byte(goModContent), 0644); err != nil {
		state.AddLine(fmt.Sprintf("Error creating go.mod: %v", err))
		return nil
	}

	// Create README.md
	readmeContent := fmt.Sprintf(`# %s

An Agar-based AI agent application.

## Setup

1. Install dependencies:
   `+"```"+`
   go mod tidy
   `+"```"+`

2. Run the application:
   `+"```"+`
   go run main.go
   `+"```"+`

## Features

- Interactive TUI interface
- Built-in tools for file operations, shell commands, and more
- Slash command system
- Extensible with custom tools and commands

## Environment Variables

- `+"`ANTHROPIC_API_KEY`"+` - Required for AI features (if using BAML)

`, projectName)

	readmePath := filepath.Join(projectName, "README.md")
	if err := os.WriteFile(readmePath, []byte(readmeContent), 0644); err != nil {
		state.AddLine(fmt.Sprintf("Error creating README.md: %v", err))
		return nil
	}

	state.AddLine(fmt.Sprintf("✓ Project '%s' initialized successfully!", projectName))
	state.AddLine("")
	state.AddLine("Next steps:")
	state.AddLine(fmt.Sprintf("  cd %s", projectName))
	state.AddLine("  go mod tidy")
	state.AddLine("  go run main.go")
	state.AddLine("")
	state.AddLine("Note: Set ANTHROPIC_API_KEY environment variable for AI features")

	return nil
}
