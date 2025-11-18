package tools

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestShellTool_Name(t *testing.T) {
	tool := NewShellTool()
	if tool.Name() != "shell" {
		t.Errorf("Expected name 'shell', got '%s'", tool.Name())
	}
}

func TestShellTool_Validate(t *testing.T) {
	tool := NewShellTool()

	tests := []struct {
		name    string
		params  string
		wantErr bool
	}{
		{
			name:    "valid params",
			params:  `{"command": "echo hello"}`,
			wantErr: false,
		},
		{
			name:    "valid with args",
			params:  `{"command": "echo", "args": ["hello", "world"]}`,
			wantErr: false,
		},
		{
			name:    "valid with timeout",
			params:  `{"command": "sleep 1", "timeout": 5}`,
			wantErr: false,
		},
		{
			name:    "valid with shell",
			params:  `{"command": "echo hello", "shell": "bash"}`,
			wantErr: false,
		},
		{
			name:    "missing command",
			params:  `{}`,
			wantErr: true,
		},
		{
			name:    "empty command",
			params:  `{"command": ""}`,
			wantErr: true,
		},
		{
			name:    "dangerous command",
			params:  `{"command": "rm -rf /"}`,
			wantErr: true,
		},
		{
			name:    "invalid timeout",
			params:  `{"command": "echo", "timeout": 500}`,
			wantErr: true,
		},
		{
			name:    "invalid shell",
			params:  `{"command": "echo", "shell": "invalid"}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tool.Validate(json.RawMessage(tt.params))
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestShellTool_Execute_SimpleCommand(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping test on Windows")
	}

	tool := NewShellTool()
	ctx := context.Background()

	params := map[string]interface{}{
		"command": "echo hello",
	}
	paramsJSON, _ := json.Marshal(params)

	result, err := tool.Execute(ctx, paramsJSON)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	shellResult, ok := result.(*ShellResult)
	if !ok {
		t.Fatal("Result is not a ShellResult")
	}

	if shellResult.ExitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", shellResult.ExitCode)
	}

	if !strings.Contains(shellResult.Stdout, "hello") {
		t.Errorf("Expected stdout to contain 'hello', got '%s'", shellResult.Stdout)
	}
}

func TestShellTool_Execute_WithArgs(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping test on Windows")
	}

	tool := NewShellTool()
	ctx := context.Background()

	params := map[string]interface{}{
		"command": "echo",
		"args":    []string{"hello", "world"},
	}
	paramsJSON, _ := json.Marshal(params)

	result, err := tool.Execute(ctx, paramsJSON)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	shellResult := result.(*ShellResult)

	if shellResult.ExitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", shellResult.ExitCode)
	}

	output := strings.TrimSpace(shellResult.Stdout)
	if output != "hello world" {
		t.Errorf("Expected 'hello world', got '%s'", output)
	}
}

func TestShellTool_Execute_WithWorkingDir(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping test on Windows")
	}

	tmpDir := t.TempDir()

	tool := NewShellTool()
	ctx := context.Background()

	params := map[string]interface{}{
		"command":     "pwd",
		"working_dir": tmpDir,
	}
	paramsJSON, _ := json.Marshal(params)

	result, err := tool.Execute(ctx, paramsJSON)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	shellResult := result.(*ShellResult)

	if shellResult.ExitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", shellResult.ExitCode)
	}

	output := strings.TrimSpace(shellResult.Stdout)
	if output != tmpDir {
		t.Errorf("Expected working dir '%s', got '%s'", tmpDir, output)
	}
}

func TestShellTool_Execute_WithEnvironment(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping test on Windows")
	}

	tool := NewShellTool()
	ctx := context.Background()

	params := map[string]interface{}{
		"command": "echo $TEST_VAR",
		"shell":   "bash",
		"environment": map[string]string{
			"TEST_VAR": "test_value",
		},
	}
	paramsJSON, _ := json.Marshal(params)

	result, err := tool.Execute(ctx, paramsJSON)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	shellResult := result.(*ShellResult)

	if shellResult.ExitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", shellResult.ExitCode)
	}

	if !strings.Contains(shellResult.Stdout, "test_value") {
		t.Errorf("Expected stdout to contain 'test_value', got '%s'", shellResult.Stdout)
	}
}

func TestShellTool_Execute_WithShell(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping test on Windows")
	}

	tool := NewShellTool()
	ctx := context.Background()

	params := map[string]interface{}{
		"command": "echo hello && echo world",
		"shell":   "bash",
	}
	paramsJSON, _ := json.Marshal(params)

	result, err := tool.Execute(ctx, paramsJSON)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	shellResult := result.(*ShellResult)

	if shellResult.ExitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", shellResult.ExitCode)
	}

	output := shellResult.Stdout
	if !strings.Contains(output, "hello") || !strings.Contains(output, "world") {
		t.Errorf("Expected output to contain both 'hello' and 'world', got '%s'", output)
	}
}

func TestShellTool_Execute_NonZeroExit(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping test on Windows")
	}

	tool := NewShellTool()
	ctx := context.Background()

	params := map[string]interface{}{
		"command": "exit 42",
		"shell":   "bash",
	}
	paramsJSON, _ := json.Marshal(params)

	result, err := tool.Execute(ctx, paramsJSON)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	shellResult := result.(*ShellResult)

	if shellResult.ExitCode != 42 {
		t.Errorf("Expected exit code 42, got %d", shellResult.ExitCode)
	}
}

func TestShellTool_Execute_Timeout(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping test on Windows")
	}

	tool := NewShellTool()
	ctx := context.Background()

	params := map[string]interface{}{
		"command": "sleep 5",
		"timeout": 1,
	}
	paramsJSON, _ := json.Marshal(params)

	result, err := tool.Execute(ctx, paramsJSON)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	shellResult := result.(*ShellResult)

	if !shellResult.Timeout {
		t.Error("Expected timeout to be true")
	}

	if shellResult.ExitCode != -1 {
		t.Errorf("Expected exit code -1 for timeout, got %d", shellResult.ExitCode)
	}
}

func TestShellTool_Execute_StderrCapture(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping test on Windows")
	}

	tool := NewShellTool()
	ctx := context.Background()

	params := map[string]interface{}{
		"command": "echo error >&2",
		"shell":   "bash",
	}
	paramsJSON, _ := json.Marshal(params)

	result, err := tool.Execute(ctx, paramsJSON)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	shellResult := result.(*ShellResult)

	if !strings.Contains(shellResult.Stderr, "error") {
		t.Errorf("Expected stderr to contain 'error', got '%s'", shellResult.Stderr)
	}
}

func TestShellTool_Execute_Duration(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping test on Windows")
	}

	tool := NewShellTool()
	ctx := context.Background()

	params := map[string]interface{}{
		"command": "echo hello",
	}
	paramsJSON, _ := json.Marshal(params)

	result, err := tool.Execute(ctx, paramsJSON)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	shellResult := result.(*ShellResult)

	if shellResult.Duration < 0 {
		t.Error("Expected duration to be non-negative")
	}

	// Duration can be 0 for very fast commands in CI environments
	// Just verify it exists and is not negative
}

func TestShellTool_Execute_FileOperations(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping test on Windows")
	}

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	tool := NewShellTool()
	ctx := context.Background()

	// Create a file using shell
	params := map[string]interface{}{
		"command": "echo 'test content' > " + testFile,
		"shell":   "bash",
	}
	paramsJSON, _ := json.Marshal(params)

	result, err := tool.Execute(ctx, paramsJSON)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	shellResult := result.(*ShellResult)

	if shellResult.ExitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", shellResult.ExitCode)
	}

	// Verify file was created
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Error("Expected file to be created")
	}

	// Read file content
	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	if !strings.Contains(string(content), "test content") {
		t.Errorf("Expected file to contain 'test content', got '%s'", string(content))
	}
}
