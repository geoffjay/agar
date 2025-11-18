package tools

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestListTool_Name(t *testing.T) {
	tool := NewListTool()
	if tool.Name() != "list" {
		t.Errorf("Expected name 'list', got '%s'", tool.Name())
	}
}

func TestListTool_Validate(t *testing.T) {
	tool := NewListTool()

	tests := []struct {
		name    string
		params  string
		wantErr bool
	}{
		{
			name:    "valid params",
			params:  `{"path": "/tmp"}`,
			wantErr: false,
		},
		{
			name:    "valid with pattern",
			params:  `{"path": "/tmp", "pattern": "*.txt"}`,
			wantErr: false,
		},
		{
			name:    "valid with recursive",
			params:  `{"path": "/tmp", "recursive": true}`,
			wantErr: false,
		},
		{
			name:    "valid with include",
			params:  `{"path": "/tmp", "include": [".txt", ".md"]}`,
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

func TestListTool_Execute_BasicListing(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files
	_ = os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte("content1"), 0644)
	_ = os.WriteFile(filepath.Join(tmpDir, "file2.txt"), []byte("content2"), 0644)
	_ = os.WriteFile(filepath.Join(tmpDir, "file3.md"), []byte("content3"), 0644)

	tool := NewListTool()
	ctx := context.Background()

	params := map[string]interface{}{
		"path": tmpDir,
	}
	paramsJSON, _ := json.Marshal(params)

	result, err := tool.Execute(ctx, paramsJSON)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	listResult, ok := result.(*ListResult)
	if !ok {
		t.Fatal("Result is not a ListResult")
	}

	if listResult.Count != 3 {
		t.Errorf("Expected 3 files, got %d", listResult.Count)
	}

	if len(listResult.Files) != 3 {
		t.Errorf("Expected 3 files in list, got %d", len(listResult.Files))
	}
}

func TestListTool_Execute_WithPattern(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files
	_ = os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte("content1"), 0644)
	_ = os.WriteFile(filepath.Join(tmpDir, "file2.txt"), []byte("content2"), 0644)
	_ = os.WriteFile(filepath.Join(tmpDir, "file3.md"), []byte("content3"), 0644)

	tool := NewListTool()
	ctx := context.Background()

	params := map[string]interface{}{
		"path":    tmpDir,
		"pattern": "*.txt",
	}
	paramsJSON, _ := json.Marshal(params)

	result, err := tool.Execute(ctx, paramsJSON)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	listResult := result.(*ListResult)

	if listResult.Count != 2 {
		t.Errorf("Expected 2 files, got %d", listResult.Count)
	}

	// Verify all files are .txt
	for _, file := range listResult.Files {
		if filepath.Ext(file.Name) != ".txt" {
			t.Errorf("Expected only .txt files, got %s", file.Name)
		}
	}
}

func TestListTool_Execute_WithInclude(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files
	_ = os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte("content1"), 0644)
	_ = os.WriteFile(filepath.Join(tmpDir, "file2.md"), []byte("content2"), 0644)
	_ = os.WriteFile(filepath.Join(tmpDir, "file3.go"), []byte("content3"), 0644)

	tool := NewListTool()
	ctx := context.Background()

	params := map[string]interface{}{
		"path":    tmpDir,
		"include": []string{".txt", ".md"},
	}
	paramsJSON, _ := json.Marshal(params)

	result, err := tool.Execute(ctx, paramsJSON)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	listResult := result.(*ListResult)

	if listResult.Count != 2 {
		t.Errorf("Expected 2 files, got %d", listResult.Count)
	}
}

func TestListTool_Execute_WithExclude(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files
	_ = os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte("content1"), 0644)
	_ = os.WriteFile(filepath.Join(tmpDir, "file2.tmp"), []byte("content2"), 0644)
	_ = os.WriteFile(filepath.Join(tmpDir, "file3.log"), []byte("content3"), 0644)

	tool := NewListTool()
	ctx := context.Background()

	params := map[string]interface{}{
		"path":    tmpDir,
		"exclude": []string{".tmp", ".log"},
	}
	paramsJSON, _ := json.Marshal(params)

	result, err := tool.Execute(ctx, paramsJSON)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	listResult := result.(*ListResult)

	if listResult.Count != 1 {
		t.Errorf("Expected 1 file, got %d", listResult.Count)
	}

	if listResult.Files[0].Name != "file1.txt" {
		t.Errorf("Expected file1.txt, got %s", listResult.Files[0].Name)
	}
}

func TestListTool_Execute_Recursive(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files in root
	_ = os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte("content1"), 0644)

	// Create subdirectory with files
	subDir := filepath.Join(tmpDir, "subdir")
	_ = os.Mkdir(subDir, 0755)
	_ = os.WriteFile(filepath.Join(subDir, "file2.txt"), []byte("content2"), 0644)

	// Create nested subdirectory with files
	nestedDir := filepath.Join(subDir, "nested")
	_ = os.Mkdir(nestedDir, 0755)
	_ = os.WriteFile(filepath.Join(nestedDir, "file3.txt"), []byte("content3"), 0644)

	tool := NewListTool()
	ctx := context.Background()

	params := map[string]interface{}{
		"path":      tmpDir,
		"recursive": true,
	}
	paramsJSON, _ := json.Marshal(params)

	result, err := tool.Execute(ctx, paramsJSON)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	listResult := result.(*ListResult)

	// Should include: file1.txt, subdir, file2.txt, nested, file3.txt = 5 items
	if listResult.Count < 5 {
		t.Errorf("Expected at least 5 items, got %d", listResult.Count)
	}
}

func TestListTool_Execute_RecursiveWithPattern(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files
	_ = os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte("content1"), 0644)
	_ = os.WriteFile(filepath.Join(tmpDir, "file2.md"), []byte("content2"), 0644)

	// Create subdirectory with files
	subDir := filepath.Join(tmpDir, "subdir")
	_ = os.Mkdir(subDir, 0755)
	_ = os.WriteFile(filepath.Join(subDir, "file3.txt"), []byte("content3"), 0644)
	_ = os.WriteFile(filepath.Join(subDir, "file4.md"), []byte("content4"), 0644)

	tool := NewListTool()
	ctx := context.Background()

	params := map[string]interface{}{
		"path":      tmpDir,
		"recursive": true,
		"pattern":   "*.txt",
	}
	paramsJSON, _ := json.Marshal(params)

	result, err := tool.Execute(ctx, paramsJSON)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	listResult := result.(*ListResult)

	if listResult.Count != 2 {
		t.Errorf("Expected 2 txt files, got %d", listResult.Count)
	}

	// Verify all are .txt files
	for _, file := range listResult.Files {
		if filepath.Ext(file.Name) != ".txt" {
			t.Errorf("Expected only .txt files, got %s", file.Name)
		}
	}
}

func TestListTool_Execute_NonExistentPath(t *testing.T) {
	tool := NewListTool()
	ctx := context.Background()

	params := map[string]interface{}{
		"path": "/nonexistent/directory",
	}
	paramsJSON, _ := json.Marshal(params)

	_, err := tool.Execute(ctx, paramsJSON)
	if err == nil {
		t.Error("Expected error for non-existent path")
	}
}

func TestListTool_Execute_NotADirectory(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "file.txt")
	_ = os.WriteFile(testFile, []byte("content"), 0644)

	tool := NewListTool()
	ctx := context.Background()

	params := map[string]interface{}{
		"path": testFile,
	}
	paramsJSON, _ := json.Marshal(params)

	_, err := tool.Execute(ctx, paramsJSON)
	if err == nil {
		t.Error("Expected error when path is not a directory")
	}
}

func TestListTool_Execute_EmptyDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	tool := NewListTool()
	ctx := context.Background()

	params := map[string]interface{}{
		"path": tmpDir,
	}
	paramsJSON, _ := json.Marshal(params)

	result, err := tool.Execute(ctx, paramsJSON)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	listResult := result.(*ListResult)

	if listResult.Count != 0 {
		t.Errorf("Expected 0 files in empty directory, got %d", listResult.Count)
	}
}

func TestListTool_Execute_FileMetadata(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	_ = os.WriteFile(testFile, []byte("test content"), 0644)

	tool := NewListTool()
	ctx := context.Background()

	params := map[string]interface{}{
		"path": tmpDir,
	}
	paramsJSON, _ := json.Marshal(params)

	result, err := tool.Execute(ctx, paramsJSON)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	listResult := result.(*ListResult)

	if len(listResult.Files) != 1 {
		t.Fatalf("Expected 1 file, got %d", len(listResult.Files))
	}

	file := listResult.Files[0]

	if file.Name != "test.txt" {
		t.Errorf("Expected name 'test.txt', got '%s'", file.Name)
	}

	if file.Size != 12 { // "test content" is 12 bytes
		t.Errorf("Expected size 12, got %d", file.Size)
	}

	if file.IsDir {
		t.Error("Expected IsDir to be false")
	}

	if file.Permissions == "" {
		t.Error("Expected permissions to be set")
	}

	if file.ModifiedTime == 0 {
		t.Error("Expected modified time to be set")
	}
}
