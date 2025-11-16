package commands

import (
	"fmt"
	"sort"
	"strings"
	"sync"
)

// Registry manages command registration and retrieval.
// It is thread-safe and supports command aliases.
type Registry struct {
	commands map[string]Command // Map of command name to command
	aliases  map[string]string  // Map of alias to command name
	mu       sync.RWMutex
}

// NewRegistry creates a new command registry
func NewRegistry() *Registry {
	return &Registry{
		commands: make(map[string]Command),
		aliases:  make(map[string]string),
	}
}

// Register adds a command to the registry
func (r *Registry) Register(cmd Command) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	name := strings.ToLower(cmd.Name())

	// Check if command already exists
	if _, exists := r.commands[name]; exists {
		return fmt.Errorf("command %q already registered", name)
	}

	// Register the command
	r.commands[name] = cmd

	// Register aliases
	for _, alias := range cmd.Aliases() {
		alias = strings.ToLower(alias)
		if _, exists := r.aliases[alias]; exists {
			return fmt.Errorf("alias %q already registered", alias)
		}
		r.aliases[alias] = name
	}

	return nil
}

// Unregister removes a command from the registry
func (r *Registry) Unregister(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	name = strings.ToLower(name)

	cmd, exists := r.commands[name]
	if !exists {
		return fmt.Errorf("command %q not found", name)
	}

	// Remove aliases
	for _, alias := range cmd.Aliases() {
		delete(r.aliases, strings.ToLower(alias))
	}

	// Remove command
	delete(r.commands, name)

	return nil
}

// Get retrieves a command by name or alias
func (r *Registry) Get(name string) (Command, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	name = strings.ToLower(name)

	// Check if it's an alias first
	if cmdName, isAlias := r.aliases[name]; isAlias {
		name = cmdName
	}

	cmd, exists := r.commands[name]
	if !exists {
		return nil, fmt.Errorf("command %q not found", name)
	}

	return cmd, nil
}

// Has checks if a command exists
func (r *Registry) Has(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	name = strings.ToLower(name)

	// Check command name
	if _, exists := r.commands[name]; exists {
		return true
	}

	// Check aliases
	if _, exists := r.aliases[name]; exists {
		return true
	}

	return false
}

// List returns all registered commands (sorted by name)
func (r *Registry) List() []Command {
	r.mu.RLock()
	defer r.mu.RUnlock()

	commands := make([]Command, 0, len(r.commands))
	for _, cmd := range r.commands {
		commands = append(commands, cmd)
	}

	// Sort by name
	sort.Slice(commands, func(i, j int) bool {
		return commands[i].Name() < commands[j].Name()
	})

	return commands
}

// Names returns all registered command names (including aliases)
func (r *Registry) Names() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.commands)+len(r.aliases))

	// Add command names
	for name := range r.commands {
		names = append(names, name)
	}

	// Add aliases
	for alias := range r.aliases {
		names = append(names, alias)
	}

	sort.Strings(names)
	return names
}

// Count returns the number of registered commands
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.commands)
}

// Clear removes all commands from the registry
func (r *Registry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.commands = make(map[string]Command)
	r.aliases = make(map[string]string)
}
