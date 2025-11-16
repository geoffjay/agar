package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Loader loads commands from files
type Loader struct {
	paths []string
}

// NewLoader creates a new command loader
func NewLoader(paths ...string) *Loader {
	return &Loader{
		paths: paths,
	}
}

// LoadAll loads all commands from the configured paths
func (l *Loader) LoadAll() ([]Command, error) {
	var commands []Command

	for _, path := range l.paths {
		cmds, err := l.loadFromPath(path)
		if err != nil {
			// Log error but continue loading from other paths
			fmt.Fprintf(os.Stderr, "Warning: failed to load commands from %s: %v\n", path, err)
			continue
		}
		commands = append(commands, cmds...)
	}

	return commands, nil
}

// loadFromPath loads commands from a single path
func (l *Loader) loadFromPath(path string) ([]Command, error) {
	// Check if path exists
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // Path doesn't exist, skip silently
		}
		return nil, fmt.Errorf("failed to stat path: %w", err)
	}

	// If it's a file, load single command
	if !info.IsDir() {
		cmd, err := l.loadFromFile(path)
		if err != nil {
			return nil, err
		}
		return []Command{cmd}, nil
	}

	// If it's a directory, load all command files
	return l.loadFromDir(path)
}

// loadFromDir loads all commands from a directory
func (l *Loader) loadFromDir(dir string) ([]Command, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	var commands []Command
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Only process .yaml and .yml files
		name := entry.Name()
		if !strings.HasSuffix(name, ".yaml") && !strings.HasSuffix(name, ".yml") {
			continue
		}

		path := filepath.Join(dir, name)
		cmd, err := l.loadFromFile(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to load command from %s: %v\n", path, err)
			continue
		}

		commands = append(commands, cmd)
	}

	return commands, nil
}

// loadFromFile loads a single command from a file
func (l *Loader) loadFromFile(path string) (Command, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var cmd FileCommand
	if err := yaml.Unmarshal(data, &cmd); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Validate command
	if err := l.validate(&cmd); err != nil {
		return nil, fmt.Errorf("invalid command: %w", err)
	}

	return &cmd, nil
}

// validate checks if a command is valid
func (l *Loader) validate(cmd *FileCommand) error {
	if cmd.NameValue == "" {
		return fmt.Errorf("command name is required")
	}

	if cmd.DescriptionValue == "" {
		return fmt.Errorf("command description is required")
	}

	if cmd.Script == "" {
		return fmt.Errorf("command script is required")
	}

	return nil
}

// DefaultCommandPaths returns the default paths to search for commands
func DefaultCommandPaths() []string {
	paths := []string{}

	// Check for .agar/commands in current directory
	if fileExists(".agar/commands") {
		paths = append(paths, ".agar/commands")
	}

	// Check for ~/.agar/commands
	home, err := os.UserHomeDir()
	if err == nil {
		homeCommands := filepath.Join(home, ".agar", "commands")
		if fileExists(homeCommands) {
			paths = append(paths, homeCommands)
		}
	}

	return paths
}
