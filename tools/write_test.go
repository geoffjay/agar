package tools

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestWriteTool_Name(t *testing.T) {
	tool := NewWriteTool()
	if tool.Name() != "write" {
		t.Errorf("Expected name 'write', got '%s'", tool.Name())
	}
}

func TestWriteTool_Validate(t *testing.T) {
	tool := NewWriteTool()

	tests := []struct {
		name    string
		params  string
		wantErr bool
	}{
		{
			name:    "valid params",
			params:  `{"path": "/tmp/test.txt", "content": "hello"}`,
			wantErr: false,
		},
		{
			name:    "valid with mode",
			params:  `{"path": "/tmp/test.txt", "content": "hello", "mode": "append"}`,
			wantErr: false,
		},
		{
			name:    "valid with encoding",
			params:  `{"path": "/tmp/test.txt", "content": "aGVsbG8=", "encoding": "base64"}`,
			wantErr: false,
		},
		{
			name:    "missing path",
			params:  `{"content": "hello"}`,
			wantErr: true,
		},
		{
			name:    "empty content is valid",
			params:  `{"path": "/tmp/test.txt", "content": ""}`,
			wantErr: false,
		},
		{
			name:    "invalid mode",
			params:  `{"path": "/tmp/test.txt", "content": "hello", "mode": "invalid"}`,
			wantErr: true,
		},
		{
			name:    "invalid encoding",
			params:  `{"path": "/tmp/test.txt", "content": "hello", "encoding": "invalid"}`,
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

func TestWriteTool_Execute_Write(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	tool := NewWriteTool()
	ctx := context.Background()

	content := "Hello, World!"
	params := map[string]interface{}{
		"path":    testFile,
		"content": content,
	}
	paramsJSON, _ := json.Marshal(params)

	result, err := tool.Execute(ctx, paramsJSON)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	writeResult, ok := result.(*WriteResult)
	if !ok {
		t.Fatal("Result is not a WriteResult")
	}

	if writeResult.Path != testFile {
		t.Errorf("Expected path '%s', got '%s'", testFile, writeResult.Path)
	}

	if writeResult.BytesWritten != int64(len(content)) {
		t.Errorf("Expected %d bytes written, got %d", len(content), writeResult.BytesWritten)
	}

	// Verify file contents
	data, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	if string(data) != content {
		t.Errorf("File content mismatch.\nExpected: %s\nGot: %s", content, string(data))
	}
}

func TestWriteTool_Execute_Append(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	tool := NewWriteTool()
	ctx := context.Background()

	// Write initial content
	initialContent := "Hello"
	params1 := map[string]interface{}{
		"path":    testFile,
		"content": initialContent,
	}
	paramsJSON1, _ := json.Marshal(params1)

	_, err := tool.Execute(ctx, paramsJSON1)
	if err != nil {
		t.Fatalf("First write failed: %v", err)
	}

	// Append more content
	appendContent := ", World!"
	params2 := map[string]interface{}{
		"path":    testFile,
		"content": appendContent,
		"mode":    "append",
	}
	paramsJSON2, _ := json.Marshal(params2)

	result, err := tool.Execute(ctx, paramsJSON2)
	if err != nil {
		t.Fatalf("Append failed: %v", err)
	}

	writeResult := result.(*WriteResult)
	if writeResult.BytesWritten != int64(len(appendContent)) {
		t.Errorf("Expected %d bytes written, got %d", len(appendContent), writeResult.BytesWritten)
	}

	// Verify file contents
	data, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	expected := initialContent + appendContent
	if string(data) != expected {
		t.Errorf("File content mismatch.\nExpected: %s\nGot: %s", expected, string(data))
	}
}

func TestWriteTool_Execute_Base64(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.bin")

	tool := NewWriteTool()
	ctx := context.Background()

	// Binary data
	binaryData := []byte{0x00, 0x01, 0x02, 0x03, 0xFF}
	encoded := base64.StdEncoding.EncodeToString(binaryData)

	params := map[string]interface{}{
		"path":     testFile,
		"content":  encoded,
		"encoding": "base64",
	}
	paramsJSON, _ := json.Marshal(params)

	result, err := tool.Execute(ctx, paramsJSON)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	writeResult := result.(*WriteResult)
	if writeResult.BytesWritten != int64(len(binaryData)) {
		t.Errorf("Expected %d bytes written, got %d", len(binaryData), writeResult.BytesWritten)
	}

	// Verify file contents
	data, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	if string(data) != string(binaryData) {
		t.Error("Binary data mismatch")
	}
}

func TestWriteTool_Execute_Backup(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	tool := NewWriteTool()
	ctx := context.Background()

	// Write initial content
	initialContent := "Initial content"
	params1 := map[string]interface{}{
		"path":    testFile,
		"content": initialContent,
	}
	paramsJSON1, _ := json.Marshal(params1)

	_, err := tool.Execute(ctx, paramsJSON1)
	if err != nil {
		t.Fatalf("First write failed: %v", err)
	}

	// Overwrite with backup
	newContent := "New content"
	params2 := map[string]interface{}{
		"path":    testFile,
		"content": newContent,
		"backup":  true,
	}
	paramsJSON2, _ := json.Marshal(params2)

	result, err := tool.Execute(ctx, paramsJSON2)
	if err != nil {
		t.Fatalf("Second write failed: %v", err)
	}

	writeResult := result.(*WriteResult)
	if writeResult.BackupPath == "" {
		t.Error("Expected backup path to be set")
	}

	// Verify backup file exists and has original content
	backupData, err := os.ReadFile(writeResult.BackupPath)
	if err != nil {
		t.Fatalf("Failed to read backup file: %v", err)
	}

	if string(backupData) != initialContent {
		t.Errorf("Backup content mismatch.\nExpected: %s\nGot: %s", initialContent, string(backupData))
	}

	// Verify main file has new content
	data, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	if string(data) != newContent {
		t.Errorf("File content mismatch.\nExpected: %s\nGot: %s", newContent, string(data))
	}
}

func TestWriteTool_Execute_CreateDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "subdir", "nested", "test.txt")

	tool := NewWriteTool()
	ctx := context.Background()

	content := "Test content"
	params := map[string]interface{}{
		"path":    testFile,
		"content": content,
	}
	paramsJSON, _ := json.Marshal(params)

	result, err := tool.Execute(ctx, paramsJSON)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	writeResult := result.(*WriteResult)
	if writeResult.BytesWritten != int64(len(content)) {
		t.Errorf("Expected %d bytes written, got %d", len(content), writeResult.BytesWritten)
	}

	// Verify file exists
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Error("File was not created")
	}

	// Verify content
	data, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	if string(data) != content {
		t.Errorf("File content mismatch.\nExpected: %s\nGot: %s", content, string(data))
	}
}
