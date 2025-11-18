package commands

import (
	"context"
	"fmt"
	"strings"
)

// Handler executes commands from user input
type Handler struct {
	registry *Registry
}

// NewHandler creates a new command handler
func NewHandler(registry *Registry) *Handler {
	return &Handler{
		registry: registry,
	}
}

// Handle processes a command string and executes it
func (h *Handler) Handle(ctx context.Context, input string, state ApplicationState) error {
	// Remove leading/trailing whitespace
	input = strings.TrimSpace(input)

	// Check if input starts with "/"
	if !strings.HasPrefix(input, "/") {
		return fmt.Errorf("not a command: must start with '/'")
	}

	// Remove leading "/"
	input = strings.TrimPrefix(input, "/")

	// Parse command and arguments
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return fmt.Errorf("empty command")
	}

	cmdName := parts[0]
	args := parts[1:]

	// Get command from registry
	cmd, err := h.registry.Get(cmdName)
	if err != nil {
		return fmt.Errorf("unknown command: %s", cmdName)
	}

	// Execute command
	if err := cmd.Execute(ctx, args, state); err != nil {
		return fmt.Errorf("command execution failed: %w", err)
	}

	return nil
}

// IsCommand checks if the input is a command
func (h *Handler) IsCommand(input string) bool {
	return strings.HasPrefix(strings.TrimSpace(input), "/")
}

// GetCompletions returns command completions for the given input
func (h *Handler) GetCompletions(input string) []string {
	input = strings.TrimSpace(input)

	// If input doesn't start with "/", no completions
	if !strings.HasPrefix(input, "/") {
		return nil
	}

	// Remove leading "/"
	input = strings.TrimPrefix(input, "/")

	// Get only primary command names (not aliases)
	names := h.registry.CommandNames()

	// If input is empty, return all commands
	if input == "" {
		return names
	}

	// Filter commands that start with input
	var completions []string
	inputLower := strings.ToLower(input)
	for _, name := range names {
		if strings.HasPrefix(strings.ToLower(name), inputLower) {
			completions = append(completions, name)
		}
	}

	return completions
}

// GetCommandInfo returns detailed information about a command
func (h *Handler) GetCommandInfo(name string) (Command, error) {
	return h.registry.Get(name)
}
