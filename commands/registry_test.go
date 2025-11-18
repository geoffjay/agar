package commands

import (
	"context"
	"testing"
)

func TestRegistry(t *testing.T) {
	registry := NewRegistry()

	// Test empty registry
	if registry.Count() != 0 {
		t.Errorf("Expected empty registry, got %d commands", registry.Count())
	}

	// Create a test command
	cmd := NewCommandFunc(
		"test",
		"A test command",
		"/test",
		func(ctx context.Context, args []string, state ApplicationState) error {
			return nil
		},
	)

	// Test registration
	if err := registry.Register(cmd); err != nil {
		t.Errorf("Failed to register command: %v", err)
	}

	if registry.Count() != 1 {
		t.Errorf("Expected 1 command, got %d", registry.Count())
	}

	// Test retrieval
	retrieved, err := registry.Get("test")
	if err != nil {
		t.Errorf("Failed to get command: %v", err)
	}

	if retrieved.Name() != "test" {
		t.Errorf("Expected command name 'test', got '%s'", retrieved.Name())
	}

	// Test duplicate registration
	if err := registry.Register(cmd); err == nil {
		t.Error("Expected error when registering duplicate command")
	}

	// Test Has
	if !registry.Has("test") {
		t.Error("Registry should have 'test' command")
	}

	if registry.Has("nonexistent") {
		t.Error("Registry should not have 'nonexistent' command")
	}

	// Test unregistration
	if err := registry.Unregister("test"); err != nil {
		t.Errorf("Failed to unregister command: %v", err)
	}

	if registry.Count() != 0 {
		t.Errorf("Expected empty registry after unregister, got %d", registry.Count())
	}
}

func TestRegistryAliases(t *testing.T) {
	registry := NewRegistry()

	cmd := NewCommandFunc(
		"test",
		"A test command",
		"/test",
		func(ctx context.Context, args []string, state ApplicationState) error {
			return nil
		},
	).WithAliases("t", "tst")

	if err := registry.Register(cmd); err != nil {
		t.Fatalf("Failed to register command: %v", err)
	}

	// Test retrieval by name
	retrieved, err := registry.Get("test")
	if err != nil {
		t.Errorf("Failed to get command by name: %v", err)
	}
	if retrieved.Name() != "test" {
		t.Errorf("Expected command name 'test', got '%s'", retrieved.Name())
	}

	// Test retrieval by alias
	retrieved, err = registry.Get("t")
	if err != nil {
		t.Errorf("Failed to get command by alias 't': %v", err)
	}
	if retrieved.Name() != "test" {
		t.Errorf("Expected command name 'test', got '%s'", retrieved.Name())
	}

	retrieved, err = registry.Get("tst")
	if err != nil {
		t.Errorf("Failed to get command by alias 'tst': %v", err)
	}
	if retrieved.Name() != "test" {
		t.Errorf("Expected command name 'test', got '%s'", retrieved.Name())
	}

	// Test Has with aliases
	if !registry.Has("t") {
		t.Error("Registry should have alias 't'")
	}
	if !registry.Has("tst") {
		t.Error("Registry should have alias 'tst'")
	}
}

func TestRegistryList(t *testing.T) {
	registry := NewRegistry()

	// Register multiple commands
	for i := 0; i < 5; i++ {
		cmd := NewCommandFunc(
			string(rune('a'+i)),
			"Test command",
			"/cmd",
			func(ctx context.Context, args []string, state ApplicationState) error {
				return nil
			},
		)
		if err := registry.Register(cmd); err != nil {
			t.Fatalf("Failed to register command: %v", err)
		}
	}

	// Test List
	commands := registry.List()
	if len(commands) != 5 {
		t.Errorf("Expected 5 commands, got %d", len(commands))
	}

	// Check that commands are sorted
	for i := 0; i < len(commands)-1; i++ {
		if commands[i].Name() > commands[i+1].Name() {
			t.Error("Commands are not sorted")
			break
		}
	}
}

func TestRegistryConcurrency(t *testing.T) {
	registry := NewRegistry()

	// Test concurrent registration
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(n int) {
			cmd := NewCommandFunc(
				string(rune('a'+n)),
				"Test command",
				"/cmd",
				func(ctx context.Context, args []string, state ApplicationState) error {
					return nil
				},
			)
			_ = registry.Register(cmd) // Ignore errors in concurrent test
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify all commands were registered
	if registry.Count() != 10 {
		t.Errorf("Expected 10 commands, got %d", registry.Count())
	}
}
