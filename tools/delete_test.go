package tools

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestDeleteTool_Name(t *testing.T) {
	tool := NewDeleteTool()
	if tool.Name() != "delete" {
		t.Errorf("Expected name 'delete', got '%s'", tool.Name())
	}
}

func TestDeleteTool_Validate(t *testing.T) {
	tool := NewDeleteTool()

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
			name:    "valid with recursive",
			params:  `{"path": "/tmp/dir", "recursive": true}`,
			wantErr: false,
		},
		{
			name:    "valid with dry_run",
			params:  `{"path": "/tmp/test.txt", "dry_run": true}`,
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

func TestDeleteTool_Execute_File(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	// Create test file
	err := os.WriteFile(testFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tool := NewDeleteTool()
	ctx := context.Background()

	params := map[string]interface{}{
		"path": testFile,
	}
	paramsJSON, _ := json.Marshal(params)

	result, err := tool.Execute(ctx, paramsJSON)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	deleteResult, ok := result.(*DeleteResult)
	if !ok {
		t.Fatal("Result is not a DeleteResult")
	}

	if !deleteResult.Deleted {
		t.Error("Expected Deleted to be true")
	}

	if deleteResult.FilesRemoved != 1 {
		t.Errorf("Expected 1 file removed, got %d", deleteResult.FilesRemoved)
	}

	// Verify file is deleted
	if _, err := os.Stat(testFile); !os.IsNotExist(err) {
		t.Error("File should be deleted but still exists")
	}
}

func TestDeleteTool_Execute_NonExistentFile(t *testing.T) {
	tool := NewDeleteTool()
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

func TestDeleteTool_Execute_DirectoryWithoutRecursive(t *testing.T) {
	tmpDir := t.TempDir()
	testDir := filepath.Join(tmpDir, "testdir")

	// Create test directory
	err := os.Mkdir(testDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	tool := NewDeleteTool()
	ctx := context.Background()

	params := map[string]interface{}{
		"path": testDir,
	}
	paramsJSON, _ := json.Marshal(params)

	_, err = tool.Execute(ctx, paramsJSON)
	if err == nil {
		t.Error("Expected error when deleting directory without recursive flag")
	}
}

func TestDeleteTool_Execute_DirectoryRecursive(t *testing.T) {
	tmpDir := t.TempDir()
	testDir := filepath.Join(tmpDir, "testdir")

	// Create test directory with files
	err := os.Mkdir(testDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Create some files in the directory
	_ = os.WriteFile(filepath.Join(testDir, "file1.txt"), []byte("content1"), 0644)
	_ = os.WriteFile(filepath.Join(testDir, "file2.txt"), []byte("content2"), 0644)

	// Create a subdirectory with a file
	subDir := filepath.Join(testDir, "subdir")
	_ = os.Mkdir(subDir, 0755)
	_ = os.WriteFile(filepath.Join(subDir, "file3.txt"), []byte("content3"), 0644)

	tool := NewDeleteTool()
	ctx := context.Background()

	params := map[string]interface{}{
		"path":      testDir,
		"recursive": true,
	}
	paramsJSON, _ := json.Marshal(params)

	result, err := tool.Execute(ctx, paramsJSON)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	deleteResult := result.(*DeleteResult)

	if !deleteResult.Deleted {
		t.Error("Expected Deleted to be true")
	}

	if deleteResult.FilesRemoved <= 0 {
		t.Errorf("Expected some files removed, got %d", deleteResult.FilesRemoved)
	}

	// Verify directory is deleted
	if _, err := os.Stat(testDir); !os.IsNotExist(err) {
		t.Error("Directory should be deleted but still exists")
	}
}

func TestDeleteTool_Execute_DryRun_File(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	// Create test file
	err := os.WriteFile(testFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tool := NewDeleteTool()
	ctx := context.Background()

	params := map[string]interface{}{
		"path":    testFile,
		"dry_run": true,
	}
	paramsJSON, _ := json.Marshal(params)

	result, err := tool.Execute(ctx, paramsJSON)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	deleteResult := result.(*DeleteResult)

	if deleteResult.Deleted {
		t.Error("Expected Deleted to be false in dry-run mode")
	}

	if !deleteResult.DryRun {
		t.Error("Expected DryRun to be true")
	}

	if deleteResult.FilesRemoved != 1 {
		t.Errorf("Expected 1 file to be removed (dry-run), got %d", deleteResult.FilesRemoved)
	}

	if len(deleteResult.Items) != 1 {
		t.Errorf("Expected 1 item in list, got %d", len(deleteResult.Items))
	}

	// Verify file still exists
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Error("File should still exist in dry-run mode")
	}
}

func TestDeleteTool_Execute_DryRun_Directory(t *testing.T) {
	tmpDir := t.TempDir()
	testDir := filepath.Join(tmpDir, "testdir")

	// Create test directory with files
	err := os.Mkdir(testDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	_ = os.WriteFile(filepath.Join(testDir, "file1.txt"), []byte("content1"), 0644)
	_ = os.WriteFile(filepath.Join(testDir, "file2.txt"), []byte("content2"), 0644)

	tool := NewDeleteTool()
	ctx := context.Background()

	params := map[string]interface{}{
		"path":      testDir,
		"recursive": true,
		"dry_run":   true,
	}
	paramsJSON, _ := json.Marshal(params)

	result, err := tool.Execute(ctx, paramsJSON)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	deleteResult := result.(*DeleteResult)

	if deleteResult.Deleted {
		t.Error("Expected Deleted to be false in dry-run mode")
	}

	if !deleteResult.DryRun {
		t.Error("Expected DryRun to be true")
	}

	if deleteResult.FilesRemoved <= 0 {
		t.Errorf("Expected some files to be counted, got %d", deleteResult.FilesRemoved)
	}

	if len(deleteResult.Items) <= 0 {
		t.Error("Expected items in list")
	}

	// Verify directory still exists
	if _, err := os.Stat(testDir); os.IsNotExist(err) {
		t.Error("Directory should still exist in dry-run mode")
	}
}

func TestDeleteTool_Execute_EmptyDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	testDir := filepath.Join(tmpDir, "emptydir")

	// Create empty directory
	err := os.Mkdir(testDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	tool := NewDeleteTool()
	ctx := context.Background()

	params := map[string]interface{}{
		"path":      testDir,
		"recursive": true,
	}
	paramsJSON, _ := json.Marshal(params)

	result, err := tool.Execute(ctx, paramsJSON)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	deleteResult := result.(*DeleteResult)

	if !deleteResult.Deleted {
		t.Error("Expected Deleted to be true")
	}

	// Verify directory is deleted
	if _, err := os.Stat(testDir); !os.IsNotExist(err) {
		t.Error("Directory should be deleted but still exists")
	}
}
