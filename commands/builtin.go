package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// RegisterBuiltinCommands registers all built-in commands with the registry
func RegisterBuiltinCommands(registry *Registry) error {
	builtins := []Command{
		NewExitCommand(),
		NewHelpCommand(registry),
		NewExportCommand(),
		NewImportCommand(),
		NewClearCommand(),
	}

	for _, cmd := range builtins {
		if err := registry.Register(cmd); err != nil {
			return fmt.Errorf("failed to register builtin command %q: %w", cmd.Name(), err)
		}
	}

	return nil
}

// ExitCommand exits the application
type ExitCommand struct{}

func NewExitCommand() *ExitCommand {
	return &ExitCommand{}
}

func (c *ExitCommand) Name() string {
	return "exit"
}

func (c *ExitCommand) Description() string {
	return "Exit the application"
}

func (c *ExitCommand) Usage() string {
	return "/exit"
}

func (c *ExitCommand) Aliases() []string {
	return []string{"quit", "q"}
}

func (c *ExitCommand) Execute(ctx context.Context, args []string, state ApplicationState) error {
	state.Exit()
	return nil
}

// HelpCommand displays all available commands
type HelpCommand struct {
	registry *Registry
}

func NewHelpCommand(registry *Registry) *HelpCommand {
	return &HelpCommand{registry: registry}
}

func (c *HelpCommand) Name() string {
	return "help"
}

func (c *HelpCommand) Description() string {
	return "List all available slash commands"
}

func (c *HelpCommand) Usage() string {
	return "/help [command]"
}

func (c *HelpCommand) Aliases() []string {
	return []string{"h", "?"}
}

func (c *HelpCommand) Execute(ctx context.Context, args []string, state ApplicationState) error {
	// If a specific command is requested, show detailed help
	if len(args) > 0 {
		cmdName := args[0]
		cmd, err := c.registry.Get(cmdName)
		if err != nil {
			state.AddLine(fmt.Sprintf("Command %q not found", cmdName))
			return nil
		}

		state.AddLine(fmt.Sprintf("Command: %s", cmd.Name()))
		state.AddLine(fmt.Sprintf("Description: %s", cmd.Description()))
		state.AddLine(fmt.Sprintf("Usage: %s", cmd.Usage()))
		if len(cmd.Aliases()) > 0 {
			state.AddLine(fmt.Sprintf("Aliases: %s", strings.Join(cmd.Aliases(), ", ")))
		}
		return nil
	}

	// Show all commands
	state.AddLine("Available Commands:")
	state.AddLine("")

	commands := c.registry.List()
	if len(commands) == 0 {
		state.AddLine("No commands registered")
		return nil
	}

	// Find max name length for formatting
	maxLen := 0
	for _, cmd := range commands {
		if len(cmd.Name()) > maxLen {
			maxLen = len(cmd.Name())
		}
	}

	// Display each command
	for _, cmd := range commands {
		padding := strings.Repeat(" ", maxLen-len(cmd.Name())+2)
		line := fmt.Sprintf("  /%s%s- %s", cmd.Name(), padding, cmd.Description())
		state.AddLine(line)

		// Show aliases if any
		if len(cmd.Aliases()) > 0 {
			aliasLine := fmt.Sprintf("    %sAliases: %s",
				strings.Repeat(" ", maxLen),
				strings.Join(cmd.Aliases(), ", "))
			state.AddLine(aliasLine)
		}
	}

	state.AddLine("")
	state.AddLine("Type '/help <command>' for more information about a specific command")

	return nil
}

// ExportCommand exports the application state
type ExportCommand struct{}

func NewExportCommand() *ExportCommand {
	return &ExportCommand{}
}

func (c *ExportCommand) Name() string {
	return "export"
}

func (c *ExportCommand) Description() string {
	return "Export the current state of the application"
}

func (c *ExportCommand) Usage() string {
	return "/export [filename]"
}

func (c *ExportCommand) Aliases() []string {
	return []string{"save"}
}

func (c *ExportCommand) Execute(ctx context.Context, args []string, state ApplicationState) error {
	// Determine filename
	filename := "agar-export.json"
	if len(args) > 0 {
		filename = args[0]
		if !strings.HasSuffix(filename, ".json") {
			filename += ".json"
		}
	}

	// Build export data
	exportData := map[string]interface{}{
		"content":  state.GetContent(),
		"mode":     state.GetMode(),
		"metadata": state.GetMetadata(),
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(exportData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal export data: %w", err)
	}

	// Write to file
	if err := writeFile(filename, data); err != nil {
		return fmt.Errorf("failed to write export file: %w", err)
	}

	state.AddLine(fmt.Sprintf("✓ Exported state to %s", filename))
	return nil
}

// ImportCommand imports application state from a file
type ImportCommand struct{}

func NewImportCommand() *ImportCommand {
	return &ImportCommand{}
}

func (c *ImportCommand) Name() string {
	return "import"
}

func (c *ImportCommand) Description() string {
	return "Import the state of the application from a file"
}

func (c *ImportCommand) Usage() string {
	return "/import <filename>"
}

func (c *ImportCommand) Aliases() []string {
	return []string{"load"}
}

func (c *ImportCommand) Execute(ctx context.Context, args []string, state ApplicationState) error {
	if len(args) == 0 {
		return fmt.Errorf("filename required: %s", c.Usage())
	}

	filename := args[0]
	if !strings.HasSuffix(filename, ".json") {
		filename += ".json"
	}

	// Read file
	data, err := readFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read import file: %w", err)
	}

	// Unmarshal JSON
	var importData map[string]interface{}
	if err := json.Unmarshal(data, &importData); err != nil {
		return fmt.Errorf("failed to parse import file: %w", err)
	}

	// Restore content
	if content, ok := importData["content"].([]interface{}); ok {
		lines := make([]string, len(content))
		for i, line := range content {
			if str, ok := line.(string); ok {
				lines[i] = str
			}
		}
		state.SetContent(lines)
	}

	// Restore mode
	if mode, ok := importData["mode"].(string); ok {
		state.SetMode(mode)
	}

	// Restore metadata
	if metadata, ok := importData["metadata"].(map[string]interface{}); ok {
		for k, v := range metadata {
			state.SetMetadata(k, v)
		}
	}

	state.AddLine(fmt.Sprintf("✓ Imported state from %s", filename))
	return nil
}

// ClearCommand clears the application content
type ClearCommand struct{}

func NewClearCommand() *ClearCommand {
	return &ClearCommand{}
}

func (c *ClearCommand) Name() string {
	return "clear"
}

func (c *ClearCommand) Description() string {
	return "Clear the application content"
}

func (c *ClearCommand) Usage() string {
	return "/clear"
}

func (c *ClearCommand) Aliases() []string {
	return []string{"cls"}
}

func (c *ClearCommand) Execute(ctx context.Context, args []string, state ApplicationState) error {
	state.Clear()
	return nil
}
