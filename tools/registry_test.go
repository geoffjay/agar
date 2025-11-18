package tools

import (
	"context"
	"encoding/json"
	"testing"
)

// Mock tool for testing
type mockTool struct {
	name string
}

func (m *mockTool) Name() string {
	return m.name
}

func (m *mockTool) Description() string {
	return "Mock tool for testing"
}

func (m *mockTool) Execute(ctx context.Context, params json.RawMessage) (interface{}, error) {
	return "mock result", nil
}

func (m *mockTool) Validate(params json.RawMessage) error {
	return nil
}

func (m *mockTool) Schema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
	}
}

func TestNewToolRegistry(t *testing.T) {
	registry := NewToolRegistry()
	if registry == nil {
		t.Fatal("NewToolRegistry returned nil")
	}
	if registry.Count() != 0 {
		t.Errorf("Expected 0 tools, got %d", registry.Count())
	}
}

func TestRegister(t *testing.T) {
	registry := NewToolRegistry()
	tool := &mockTool{name: "test"}

	err := registry.Register(tool)
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	if registry.Count() != 1 {
		t.Errorf("Expected 1 tool, got %d", registry.Count())
	}

	// Try to register the same tool again
	err = registry.Register(tool)
	if err == nil {
		t.Error("Expected error when registering duplicate tool")
	}
}

func TestGet(t *testing.T) {
	registry := NewToolRegistry()
	tool := &mockTool{name: "test"}

	_ = registry.Register(tool)

	retrieved, err := registry.Get("test")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if retrieved.Name() != "test" {
		t.Errorf("Expected tool name 'test', got '%s'", retrieved.Name())
	}

	// Try to get a non-existent tool
	_, err = registry.Get("nonexistent")
	if err == nil {
		t.Error("Expected error when getting non-existent tool")
	}
}

func TestUnregister(t *testing.T) {
	registry := NewToolRegistry()
	tool := &mockTool{name: "test"}

	_ = registry.Register(tool)

	err := registry.Unregister("test")
	if err != nil {
		t.Fatalf("Unregister failed: %v", err)
	}

	if registry.Count() != 0 {
		t.Errorf("Expected 0 tools, got %d", registry.Count())
	}

	// Try to unregister a non-existent tool
	err = registry.Unregister("nonexistent")
	if err == nil {
		t.Error("Expected error when unregistering non-existent tool")
	}
}

func TestList(t *testing.T) {
	registry := NewToolRegistry()
	tool1 := &mockTool{name: "tool1"}
	tool2 := &mockTool{name: "tool2"}

	_ = registry.Register(tool1)
	_ = registry.Register(tool2)

	list := registry.List()
	if len(list) != 2 {
		t.Errorf("Expected 2 tools in list, got %d", len(list))
	}

	// Check that both tools are in the list
	found := make(map[string]bool)
	for _, name := range list {
		found[name] = true
	}

	if !found["tool1"] || !found["tool2"] {
		t.Error("Expected both tool1 and tool2 in list")
	}
}

func TestConcurrency(t *testing.T) {
	registry := NewToolRegistry()

	// Register tools concurrently
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(n int) {
			tool := &mockTool{name: string(rune('a' + n))}
			_ = registry.Register(tool)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	if registry.Count() != 10 {
		t.Errorf("Expected 10 tools, got %d", registry.Count())
	}
}
