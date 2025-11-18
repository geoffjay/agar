// Package commands provides a slash command system for agar applications.
// Commands can be built-in, loaded from files, or registered programmatically.
package commands

import (
	"context"
	"encoding/json"
)

// Command represents a slash command that can be executed within an application.
type Command interface {
	// Name returns the command name (without the leading "/")
	Name() string

	// Description returns a brief description of what the command does
	Description() string

	// Usage returns usage instructions for the command
	Usage() string

	// Execute runs the command with the given arguments and application state
	Execute(ctx context.Context, args []string, state ApplicationState) error

	// Aliases returns alternative names for this command
	Aliases() []string
}

// ApplicationState provides access to application state for commands.
// This interface allows commands to interact with the application without
// tight coupling to the Application type.
type ApplicationState interface {
	// GetContent returns the current application content
	GetContent() []string

	// SetContent replaces all application content
	SetContent(lines []string)

	// AddLine appends a line to the application content
	AddLine(line string)

	// Clear removes all content
	Clear()

	// GetMode returns the current application mode
	GetMode() string

	// SetMode sets the application mode
	SetMode(mode string)

	// GetMetadata returns application metadata
	GetMetadata() map[string]interface{}

	// SetMetadata sets application metadata
	SetMetadata(key string, value interface{})

	// Exit signals the application to exit
	Exit()
}

// CommandFunc is a function type that implements the Command interface for simple commands
type CommandFunc struct {
	name        string
	description string
	usage       string
	aliases     []string
	fn          func(ctx context.Context, args []string, state ApplicationState) error
}

// NewCommandFunc creates a new command from a function
func NewCommandFunc(name, description, usage string, fn func(ctx context.Context, args []string, state ApplicationState) error) *CommandFunc {
	return &CommandFunc{
		name:        name,
		description: description,
		usage:       usage,
		aliases:     []string{},
		fn:          fn,
	}
}

// Name returns the command name
func (c *CommandFunc) Name() string {
	return c.name
}

// Description returns the command description
func (c *CommandFunc) Description() string {
	return c.description
}

// Usage returns the command usage
func (c *CommandFunc) Usage() string {
	return c.usage
}

// Execute runs the command function
func (c *CommandFunc) Execute(ctx context.Context, args []string, state ApplicationState) error {
	return c.fn(ctx, args, state)
}

// Aliases returns the command aliases
func (c *CommandFunc) Aliases() []string {
	return c.aliases
}

// WithAliases sets aliases for the command
func (c *CommandFunc) WithAliases(aliases ...string) *CommandFunc {
	c.aliases = aliases
	return c
}

// FileCommand represents a command loaded from a file
type FileCommand struct {
	NameValue        string            `json:"name" yaml:"name"`
	DescriptionValue string            `json:"description" yaml:"description"`
	UsageValue       string            `json:"usage" yaml:"usage"`
	AliasesValue     []string          `json:"aliases" yaml:"aliases"`
	Script           string            `json:"script" yaml:"script"`
	Metadata         map[string]string `json:"metadata" yaml:"metadata"`
}

// Name returns the command name
func (f *FileCommand) Name() string {
	return f.NameValue
}

// Description returns the command description
func (f *FileCommand) Description() string {
	return f.DescriptionValue
}

// Usage returns the command usage
func (f *FileCommand) Usage() string {
	return f.UsageValue
}

// Aliases returns the command aliases
func (f *FileCommand) Aliases() []string {
	if f.AliasesValue == nil {
		return []string{}
	}
	return f.AliasesValue
}

// Execute runs the command script
func (f *FileCommand) Execute(ctx context.Context, args []string, state ApplicationState) error {
	// For file-based commands, we'll execute the script
	// This will be implemented with proper script execution
	handler := &ScriptHandler{
		script: f.Script,
		env:    f.buildEnv(args, state),
	}
	return handler.Execute(ctx, state)
}

// buildEnv creates environment variables for script execution
func (f *FileCommand) buildEnv(args []string, state ApplicationState) map[string]string {
	env := make(map[string]string)

	// Add command metadata
	for k, v := range f.Metadata {
		env[k] = v
	}

	// Add command arguments
	if len(args) > 0 {
		argsJSON, _ := json.Marshal(args)
		env["AGAR_COMMAND_ARGS"] = string(argsJSON)
	}

	// Add application state
	env["AGAR_MODE"] = state.GetMode()
	contentJSON, _ := json.Marshal(state.GetContent())
	env["AGAR_CONTENT"] = string(contentJSON)

	return env
}
