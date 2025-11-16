package commands

import (
	"context"
	"strings"
	"testing"
)

type mockState struct {
	content  []string
	mode     string
	metadata map[string]interface{}
	exited   bool
}

func newMockState() *mockState {
	return &mockState{
		content:  []string{},
		mode:     "TEST",
		metadata: make(map[string]interface{}),
		exited:   false,
	}
}

func (m *mockState) GetContent() []string {
	return m.content
}

func (m *mockState) SetContent(lines []string) {
	m.content = lines
}

func (m *mockState) AddLine(line string) {
	m.content = append(m.content, line)
}

func (m *mockState) Clear() {
	m.content = []string{}
}

func (m *mockState) GetMode() string {
	return m.mode
}

func (m *mockState) SetMode(mode string) {
	m.mode = mode
}

func (m *mockState) GetMetadata() map[string]interface{} {
	return m.metadata
}

func (m *mockState) SetMetadata(key string, value interface{}) {
	m.metadata[key] = value
}

func (m *mockState) Exit() {
	m.exited = true
}

func TestHandler(t *testing.T) {
	registry := NewRegistry()
	handler := NewHandler(registry)

	// Register a test command
	executed := false
	cmd := NewCommandFunc(
		"test",
		"A test command",
		"/test",
		func(ctx context.Context, args []string, state ApplicationState) error {
			executed = true
			state.AddLine("Test executed")
			return nil
		},
	)

	if err := registry.Register(cmd); err != nil {
		t.Fatalf("Failed to register command: %v", err)
	}

	// Test command execution
	state := newMockState()
	if err := handler.Handle(context.Background(), "/test", state); err != nil {
		t.Errorf("Failed to handle command: %v", err)
	}

	if !executed {
		t.Error("Command was not executed")
	}

	if len(state.content) != 1 || state.content[0] != "Test executed" {
		t.Errorf("Expected 'Test executed', got %v", state.content)
	}
}

func TestHandlerWithArgs(t *testing.T) {
	registry := NewRegistry()
	handler := NewHandler(registry)

	// Register a command that uses arguments
	var receivedArgs []string
	cmd := NewCommandFunc(
		"echo",
		"Echo arguments",
		"/echo <args...>",
		func(ctx context.Context, args []string, state ApplicationState) error {
			receivedArgs = args
			state.AddLine(strings.Join(args, " "))
			return nil
		},
	)

	if err := registry.Register(cmd); err != nil {
		t.Fatalf("Failed to register command: %v", err)
	}

	// Test with arguments
	state := newMockState()
	if err := handler.Handle(context.Background(), "/echo hello world", state); err != nil {
		t.Errorf("Failed to handle command: %v", err)
	}

	if len(receivedArgs) != 2 || receivedArgs[0] != "hello" || receivedArgs[1] != "world" {
		t.Errorf("Expected ['hello', 'world'], got %v", receivedArgs)
	}
}

func TestHandlerInvalidCommands(t *testing.T) {
	registry := NewRegistry()
	handler := NewHandler(registry)
	state := newMockState()

	// Test command without leading slash
	err := handler.Handle(context.Background(), "test", state)
	if err == nil {
		t.Error("Expected error for command without leading slash")
	}

	// Test unknown command
	err = handler.Handle(context.Background(), "/unknown", state)
	if err == nil {
		t.Error("Expected error for unknown command")
	}

	// Test empty command
	err = handler.Handle(context.Background(), "/", state)
	if err == nil {
		t.Error("Expected error for empty command")
	}
}

func TestHandlerIsCommand(t *testing.T) {
	handler := NewHandler(NewRegistry())

	tests := []struct {
		input    string
		expected bool
	}{
		{"/test", true},
		{"/test arg", true},
		{"  /test  ", true},
		{"test", false},
		{"", false},
		{"  ", false},
	}

	for _, test := range tests {
		result := handler.IsCommand(test.input)
		if result != test.expected {
			t.Errorf("IsCommand(%q) = %v, expected %v", test.input, result, test.expected)
		}
	}
}

func TestHandlerGetCompletions(t *testing.T) {
	registry := NewRegistry()
	handler := NewHandler(registry)

	// Register some commands
	commands := []string{"help", "hello", "history", "test", "time"}
	for _, name := range commands {
		cmd := NewCommandFunc(
			name,
			"Test command",
			"/"+name,
			func(ctx context.Context, args []string, state ApplicationState) error {
				return nil
			},
		)
		if err := registry.Register(cmd); err != nil {
			t.Fatalf("Failed to register command: %v", err)
		}
	}

	// Test completions
	tests := []struct {
		input         string
		expectedCount int
		shouldContain []string
	}{
		{"/", 5, []string{"help", "hello", "history", "test", "time"}},
		{"/h", 3, []string{"hello", "help", "history"}},
		{"/he", 2, []string{"hello", "help"}},
		{"/hel", 2, []string{"hello", "help"}},
		{"/hello", 1, []string{"hello"}},
		{"/t", 2, []string{"test", "time"}},
		{"/x", 0, []string{}},
	}

	for _, test := range tests {
		result := handler.GetCompletions(test.input)
		if len(result) != test.expectedCount {
			t.Errorf("GetCompletions(%q) returned %d results, expected %d", test.input, len(result), test.expectedCount)
		}
		for _, expected := range test.shouldContain {
			if !contains(result, expected) {
				t.Errorf("GetCompletions(%q) missing expected value %q", test.input, expected)
			}
		}
	}

	// Test non-command input
	result := handler.GetCompletions("test")
	if result != nil {
		t.Errorf("GetCompletions('test') should return nil for non-command input, got %v", result)
	}
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
