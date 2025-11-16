package tools

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadTool_Name(t *testing.T) {
	tool := NewReadTool()
	if tool.Name() != "read" {
		t.Errorf("Expected name 'read', got '%s'", tool.Name())
	}
}

func TestReadTool_Description(t *testing.T) {
	tool := NewReadTool()
	desc := tool.Description()
	if desc == "" {
		t.Error("Description should not be empty")
	}
}

func TestReadTool_Schema(t *testing.T) {
	tool := NewReadTool()
	schema := tool.Schema()
	if schema == nil {
		t.Fatal("Schema should not be nil")
	}

	if schema["type"] != "object" {
		t.Error("Schema type should be 'object'")
	}
}

func TestReadTool_Validate(t *testing.T) {
	tool := NewReadTool()

	tests := []struct {
		name    string
		params  string
		wantErr bool
	}{
		{
			name:    "valid params",
			params:  `{"path": "/tmp/test.txt"}`,
			wantErr: false,
		},
		{
			name:    "valid with format",
			params:  `{"path": "/tmp/test.txt", "format": "text"}`,
			wantErr: false,
		},
		{
			name:    "valid with offset and limit",
			params:  `{"path": "/tmp/test.txt", "offset": 10, "limit": 5}`,
			wantErr: false,
		},
		{
			name:    "missing path",
			params:  `{}`,
			wantErr: true,
		},
		{
			name:    "empty path",
			params:  `{"path": ""}`,
			wantErr: true,
		},
		{
			name:    "invalid format",
			params:  `{"path": "/tmp/test.txt", "format": "invalid"}`,
			wantErr: true,
		},
		{
			name:    "negative offset",
			params:  `{"path": "/tmp/test.txt", "offset": -1}`,
			wantErr: true,
		},
		{
			name:    "negative limit",
			params:  `{"path": "/tmp/test.txt", "limit": -1}`,
			wantErr: true,
		},
		{
			name:    "invalid json",
			params:  `{invalid}`,
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

func TestReadTool_Execute_TextFile(t *testing.T) {
	// Create a temporary text file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	content := "line1\nline2\nline3\nline4\nline5"

	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tool := NewReadTool()
	ctx := context.Background()

	// Test basic read
	params := map[string]interface{}{
		"path": testFile,
	}
	paramsJSON, _ := json.Marshal(params)

	result, err := tool.Execute(ctx, paramsJSON)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	readResult, ok := result.(*ReadResult)
	if !ok {
		t.Fatal("Result is not a ReadResult")
	}

	if readResult.Format != "text" {
		t.Errorf("Expected format 'text', got '%s'", readResult.Format)
	}

	if readResult.Content != content {
		t.Errorf("Content mismatch.\nExpected: %s\nGot: %s", content, readResult.Content)
	}

	if readResult.TotalLines != 5 {
		t.Errorf("Expected 5 total lines, got %d", readResult.TotalLines)
	}
}

func TestReadTool_Execute_WithOffset(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	content := "line1\nline2\nline3\nline4\nline5"

	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tool := NewReadTool()
	ctx := context.Background()

	params := map[string]interface{}{
		"path":   testFile,
		"offset": 2,
	}
	paramsJSON, _ := json.Marshal(params)

	result, err := tool.Execute(ctx, paramsJSON)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	readResult := result.(*ReadResult)
	expected := "line3\nline4\nline5"

	if readResult.Content != expected {
		t.Errorf("Content mismatch.\nExpected: %s\nGot: %s", expected, readResult.Content)
	}
}

func TestReadTool_Execute_WithLimit(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	content := "line1\nline2\nline3\nline4\nline5"

	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tool := NewReadTool()
	ctx := context.Background()

	params := map[string]interface{}{
		"path":  testFile,
		"limit": 2,
	}
	paramsJSON, _ := json.Marshal(params)

	result, err := tool.Execute(ctx, paramsJSON)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	readResult := result.(*ReadResult)
	expected := "line1\nline2"

	if readResult.Content != expected {
		t.Errorf("Content mismatch.\nExpected: %s\nGot: %s", expected, readResult.Content)
	}

	if readResult.Lines != 2 {
		t.Errorf("Expected 2 lines, got %d", readResult.Lines)
	}
}

func TestReadTool_Execute_WithOffsetAndLimit(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	content := "line1\nline2\nline3\nline4\nline5"

	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tool := NewReadTool()
	ctx := context.Background()

	params := map[string]interface{}{
		"path":   testFile,
		"offset": 1,
		"limit":  2,
	}
	paramsJSON, _ := json.Marshal(params)

	result, err := tool.Execute(ctx, paramsJSON)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	readResult := result.(*ReadResult)
	expected := "line2\nline3"

	if readResult.Content != expected {
		t.Errorf("Content mismatch.\nExpected: %s\nGot: %s", expected, readResult.Content)
	}
}

func TestReadTool_Execute_BinaryFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.bin")
	binaryData := []byte{0x00, 0x01, 0x02, 0x03, 0xFF}

	err := os.WriteFile(testFile, binaryData, 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tool := NewReadTool()
	ctx := context.Background()

	params := map[string]interface{}{
		"path":   testFile,
		"format": "binary",
	}
	paramsJSON, _ := json.Marshal(params)

	result, err := tool.Execute(ctx, paramsJSON)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	readResult := result.(*ReadResult)

	if readResult.Format != "binary" {
		t.Errorf("Expected format 'binary', got '%s'", readResult.Format)
	}

	// Decode and verify
	decoded, err := base64.StdEncoding.DecodeString(readResult.Content)
	if err != nil {
		t.Fatalf("Failed to decode base64: %v", err)
	}

	if string(decoded) != string(binaryData) {
		t.Error("Binary data mismatch")
	}
}

func TestReadTool_Execute_NonExistentFile(t *testing.T) {
	tool := NewReadTool()
	ctx := context.Background()

	params := map[string]interface{}{
		"path": "/nonexistent/file.txt",
	}
	paramsJSON, _ := json.Marshal(params)

	_, err := tool.Execute(ctx, paramsJSON)
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

func TestReadTool_Execute_Directory(t *testing.T) {
	tmpDir := t.TempDir()

	tool := NewReadTool()
	ctx := context.Background()

	params := map[string]interface{}{
		"path": tmpDir,
	}
	paramsJSON, _ := json.Marshal(params)

	_, err := tool.Execute(ctx, paramsJSON)
	if err == nil {
		t.Error("Expected error when trying to read a directory")
	}

	if !strings.Contains(err.Error(), "directory") {
		t.Errorf("Expected error message to mention directory, got: %v", err)
	}
}

func TestReadTool_DetectFormat(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		want     string
	}{
		{"text file", "test.txt", "text"},
		{"markdown file", "README.md", "text"},
		{"json file", "config.json", "text"},
		{"go file", "main.go", "text"},
		{"no extension", "testfile", "text"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detectFormat(tt.filename)
			if got != tt.want {
				t.Errorf("detectFormat(%s) = %s, want %s", tt.filename, got, tt.want)
			}
		})
	}

	// Test binary detection with actual file
	t.Run("binary file with null bytes", func(t *testing.T) {
		tmpDir := t.TempDir()
		binFile := filepath.Join(tmpDir, "test.bin")
		// Write binary data with null bytes
		_ = os.WriteFile(binFile, []byte{0x00, 0x01, 0x02}, 0644)

		got := detectFormat(binFile)
		if got != "binary" {
			t.Errorf("detectFormat(%s) = %s, want binary", binFile, got)
		}
	})
}
