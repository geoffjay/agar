package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/geoffjay/agar/cmd/agar/baml_client"
	"github.com/geoffjay/agar/cmd/agar/baml_client/types"
	"github.com/geoffjay/agar/tools"
	"github.com/spf13/cobra"
)

// InitCmd creates a new agar init command
func InitCmd() *cobra.Command {
	var requirements string

	cmd := &cobra.Command{
		Use:   "init [project-name]",
		Short: "Initialize a new Agar application",
		Long: `Initialize a new Agar application using AI to generate the project structure.

The AI will analyze your requirements and select appropriate tools and components.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectName := args[0]
			return initProject(projectName, requirements)
		},
	}

	cmd.Flags().StringVarP(&requirements, "requirements", "r", "", "Project requirements description")

	return cmd
}

// initProject initializes a new Agar application
func initProject(projectName, requirements string) error {
	fmt.Printf("Creating new Agar application: %s\n", projectName)

	// Use default requirements if none provided
	if requirements == "" {
		requirements = "A general-purpose AI agent application with file operations and TUI interface"
	}

	fmt.Printf("Requirements: %s\n", requirements)
	fmt.Println("\nGenerating project configuration using AI...")

	// Call BAML function to generate config
	ctx := context.Background()
	config, err := baml_client.CreateAgarApp(ctx, requirements, projectName)
	if err != nil {
		return fmt.Errorf("failed to generate project config: %w", err)
	}

	fmt.Printf("\nGenerated configuration:\n")
	fmt.Printf("  Name: %s\n", config.Name)
	fmt.Printf("  Description: %s\n", config.Description)
	fmt.Printf("  Module: %s\n", config.Module_path)
	fmt.Printf("  Tools: %d selected\n", len(config.Tools))
	fmt.Printf("  Components: %d selected\n", len(config.Components))
	fmt.Printf("  Features: %d planned\n", len(config.Features))

	// Create project directory
	if err = os.MkdirAll(projectName, 0o755); err != nil {
		return fmt.Errorf("failed to create project directory: %w", err)
	}

	fmt.Printf("\nCreating project files...\n")

	// Initialize tool registry for file operations
	registry := tools.NewToolRegistry()
	registry.Register(tools.NewWriteTool())

	writeTool, _ := registry.Get("write")

	// Create go.mod
	goModContent := generateGoMod(config)
	if err = writeFile(writeTool, ctx, filepath.Join(projectName, "go.mod"), goModContent); err != nil {
		return fmt.Errorf("failed to create go.mod: %w", err)
	}
	fmt.Println("  ✓ go.mod")

	// Generate main.go
	fmt.Println("\nGenerating main.go using AI...")
	mainContent, err := baml_client.GenerateMainGo(ctx, config)
	if err != nil {
		return fmt.Errorf("failed to generate main.go: %w", err)
	}

	if err := writeFile(writeTool, ctx, filepath.Join(projectName, "main.go"), mainContent); err != nil {
		return fmt.Errorf("failed to create main.go: %w", err)
	}
	fmt.Println("  ✓ main.go")

	// Create README.md
	readmeContent := generateReadme(config)
	if err := writeFile(writeTool, ctx, filepath.Join(projectName, "README.md"), readmeContent); err != nil {
		return fmt.Errorf("failed to create README.md: %w", err)
	}
	fmt.Println("  ✓ README.md")

	// Create .gitignore
	gitignoreContent := generateGitignore()
	if err := writeFile(writeTool, ctx, filepath.Join(projectName, ".gitignore"), gitignoreContent); err != nil {
		return fmt.Errorf("failed to create .gitignore: %w", err)
	}
	fmt.Println("  ✓ .gitignore")

	fmt.Printf("\n✓ Project '%s' created successfully!\n\n", projectName)
	fmt.Println("Next steps:")
	fmt.Printf("  cd %s\n", projectName)
	fmt.Println("  go mod tidy")
	fmt.Println("  go run main.go")
	fmt.Println()

	return nil
}

// writeFile uses the agar Write tool to create a file
func writeFile(writeTool tools.Tool, ctx context.Context, path, content string) error {
	params := map[string]interface{}{
		"path":    path,
		"content": content,
	}
	paramsJSON, _ := json.Marshal(params)

	_, err := writeTool.Execute(ctx, paramsJSON)
	return err
}

// generateGoMod creates the go.mod content
func generateGoMod(config types.AgarAppConfig) string {
	return fmt.Sprintf(`module %s

go 1.25.2

require (
	github.com/geoffjay/agar v0.0.0
	github.com/charmbracelet/bubbletea v1.3.10
	github.com/charmbracelet/lipgloss v1.1.0
)

replace github.com/geoffjay/agar => ../agar
`, config.Module_path)
}

// generateReadme creates the README.md content
func generateReadme(config types.AgarAppConfig) string {
	var toolsList strings.Builder
	for _, tool := range config.Tools {
		toolsList.WriteString(fmt.Sprintf("- **%s**: %s\n", tool.Name, tool.Purpose))
	}

	return fmt.Sprintf(`# %s

%s

## Features

%s

## Tools

This application uses the following Agar tools:

%s

## Usage

`+"```bash"+`
go run main.go
`+"```"+`

## Dependencies

- [Agar](https://github.com/geoffjay/agar) - AI agent framework
- Bubble Tea - Terminal UI framework

## License

MIT
`, config.Name, config.Description, strings.Join(config.Features, "\n"), toolsList.String())
}

// generateGitignore creates the .gitignore content
func generateGitignore() string {
	return `# Binaries
*.exe
*.exe~
*.dll
*.so
*.dylib

# Test binary
*.test

# Output
*.out

# Go workspace file
go.work

# IDE
.vscode/
.idea/
*.swp
*.swo
*~

# OS
.DS_Store
Thumbs.db
`
}
