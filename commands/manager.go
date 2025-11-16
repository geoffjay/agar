package commands

import (
	"context"
	"fmt"
)

// Manager provides a high-level interface for managing the command system.
// It combines the registry, handler, and loader for easy integration.
type Manager struct {
	registry *Registry
	handler  *Handler
	loader   *Loader
}

// NewManager creates a new command manager with default configuration
func NewManager() *Manager {
	registry := NewRegistry()
	handler := NewHandler(registry)
	loader := NewLoader(DefaultCommandPaths()...)

	return &Manager{
		registry: registry,
		handler:  handler,
		loader:   loader,
	}
}

// NewManagerWithPaths creates a new command manager with custom command paths
func NewManagerWithPaths(paths ...string) *Manager {
	registry := NewRegistry()
	handler := NewHandler(registry)
	loader := NewLoader(paths...)

	return &Manager{
		registry: registry,
		handler:  handler,
		loader:   loader,
	}
}

// Initialize loads built-in and file-based commands
func (m *Manager) Initialize() error {
	// Register built-in commands
	if err := RegisterBuiltinCommands(m.registry); err != nil {
		return fmt.Errorf("failed to register builtin commands: %w", err)
	}

	// Load file-based commands
	commands, err := m.loader.LoadAll()
	if err != nil {
		return fmt.Errorf("failed to load file-based commands: %w", err)
	}

	// Register loaded commands
	for _, cmd := range commands {
		if err := m.registry.Register(cmd); err != nil {
			// Log warning but continue
			fmt.Printf("Warning: failed to register command %q: %v\n", cmd.Name(), err)
		}
	}

	return nil
}

// RegisterCommand registers a custom command
func (m *Manager) RegisterCommand(cmd Command) error {
	return m.registry.Register(cmd)
}

// UnregisterCommand removes a command
func (m *Manager) UnregisterCommand(name string) error {
	return m.registry.Unregister(name)
}

// Handle executes a command
func (m *Manager) Handle(ctx context.Context, input string, state ApplicationState) error {
	return m.handler.Handle(ctx, input, state)
}

// IsCommand checks if the input is a command
func (m *Manager) IsCommand(input string) bool {
	return m.handler.IsCommand(input)
}

// GetCompletions returns command completions
func (m *Manager) GetCompletions(input string) []string {
	return m.handler.GetCompletions(input)
}

// GetCommand returns a command by name
func (m *Manager) GetCommand(name string) (Command, error) {
	return m.registry.Get(name)
}

// ListCommands returns all registered commands
func (m *Manager) ListCommands() []Command {
	return m.registry.List()
}

// AddCommandPath adds a path to search for commands
func (m *Manager) AddCommandPath(path string) {
	m.loader.paths = append(m.loader.paths, path)
}

// ReloadCommands reloads all file-based commands
func (m *Manager) ReloadCommands() error {
	// Clear existing file-based commands (keep built-in)
	// This is a simplification - in practice you might want to track which
	// commands are file-based vs built-in
	commands, err := m.loader.LoadAll()
	if err != nil {
		return fmt.Errorf("failed to reload commands: %w", err)
	}

	// Re-register loaded commands
	for _, cmd := range commands {
		// Unregister if exists (ignore errors)
		_ = m.registry.Unregister(cmd.Name())

		// Register new version
		if err := m.registry.Register(cmd); err != nil {
			fmt.Printf("Warning: failed to register command %q: %v\n", cmd.Name(), err)
		}
	}

	return nil
}
