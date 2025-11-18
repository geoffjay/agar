package tools

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestGlobTool_Name(t *testing.T) {
	tool := NewGlobTool()
	if tool.Name() != "glob" {
		t.Errorf("Expected name 'glob', got '%s'", tool.Name())
	}
}

func TestGlobTool_Validate(t *testing.T) {
	tool := NewGlobTool()

	tests := []struct {
		name    string
		params  string
		wantErr bool
	}{
		{
			name:    "valid params",
			params:  `{"patterns": ["*.txt"]}`,
			wantErr: false,
		},
		{
			name:    "valid multiple patterns",
			params:  `{"patterns": ["*.txt", "*.md"]}`,
			wantErr: false,
		},
		{
			name:    "valid with path",
			params:  `{"patterns": ["*.txt"], "path": "/tmp"}`,
			wantErr: false,
		},
		{
			name:    "valid with sort",
			params:  `{"patterns": ["*.txt"], "sort_by": "modtime", "sort_order": "desc"}`,
			wantErr: false,
		},
		{
			name:    "missing patterns",
			params:  `{}`,
			wantErr: true,
		},
		{
			name:    "empty patterns array",
			params:  `{"patterns": []}`,
			wantErr: true,
		},
		{
			name:    "invalid sort_by",
			params:  `{"patterns": ["*.txt"], "sort_by": "invalid"}`,
			wantErr: true,
		},
		{
			name:    "invalid sort_order",
			params:  `{"patterns": ["*.txt"], "sort_order": "invalid"}`,
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

func TestGlobTool_Execute_SimplePattern(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files
	_ = os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte("content1"), 0644)
	_ = os.WriteFile(filepath.Join(tmpDir, "file2.txt"), []byte("content2"), 0644)
	_ = os.WriteFile(filepath.Join(tmpDir, "file3.md"), []byte("content3"), 0644)

	tool := NewGlobTool()
	ctx := context.Background()

	params := map[string]interface{}{
		"patterns": []string{"*.txt"},
		"path":     tmpDir,
	}
	paramsJSON, _ := json.Marshal(params)

	result, err := tool.Execute(ctx, paramsJSON)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	globResult, ok := result.(*GlobResult)
	if !ok {
		t.Fatal("Result is not a GlobResult")
	}

	if globResult.Count != 2 {
		t.Errorf("Expected 2 matches, got %d", globResult.Count)
	}
}

func TestGlobTool_Execute_MultiplePatterns(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files
	_ = os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte("content1"), 0644)
	_ = os.WriteFile(filepath.Join(tmpDir, "file2.md"), []byte("content2"), 0644)
	_ = os.WriteFile(filepath.Join(tmpDir, "file3.go"), []byte("content3"), 0644)

	tool := NewGlobTool()
	ctx := context.Background()

	params := map[string]interface{}{
		"patterns": []string{"*.txt", "*.md"},
		"path":     tmpDir,
	}
	paramsJSON, _ := json.Marshal(params)

	result, err := tool.Execute(ctx, paramsJSON)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	globResult := result.(*GlobResult)

	if globResult.Count != 2 {
		t.Errorf("Expected 2 matches, got %d", globResult.Count)
	}
}

func TestGlobTool_Execute_RecursivePattern(t *testing.T) {
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

	tool := NewGlobTool()
	ctx := context.Background()

	params := map[string]interface{}{
		"patterns": []string{"**/*.txt"},
		"path":     tmpDir,
	}
	paramsJSON, _ := json.Marshal(params)

	result, err := tool.Execute(ctx, paramsJSON)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	globResult := result.(*GlobResult)

	if globResult.Count < 3 {
		t.Errorf("Expected at least 3 matches, got %d", globResult.Count)
	}
}

func TestGlobTool_Execute_WithIncludeInfo(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	_ = os.WriteFile(testFile, []byte("test content"), 0644)

	tool := NewGlobTool()
	ctx := context.Background()

	params := map[string]interface{}{
		"patterns":     []string{"*.txt"},
		"path":         tmpDir,
		"include_info": true,
	}
	paramsJSON, _ := json.Marshal(params)

	result, err := tool.Execute(ctx, paramsJSON)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	globResult := result.(*GlobResult)

	if len(globResult.Matches) != 1 {
		t.Fatalf("Expected 1 match, got %d", len(globResult.Matches))
	}

	match := globResult.Matches[0]

	if match.Size == 0 {
		t.Error("Expected size to be set when include_info is true")
	}

	if match.Permissions == "" {
		t.Error("Expected permissions to be set when include_info is true")
	}

	if match.ModifiedTime == 0 {
		t.Error("Expected modified time to be set when include_info is true")
	}
}

func TestGlobTool_Execute_SortByName(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files
	_ = os.WriteFile(filepath.Join(tmpDir, "charlie.txt"), []byte("content"), 0644)
	_ = os.WriteFile(filepath.Join(tmpDir, "alpha.txt"), []byte("content"), 0644)
	_ = os.WriteFile(filepath.Join(tmpDir, "bravo.txt"), []byte("content"), 0644)

	tool := NewGlobTool()
	ctx := context.Background()

	params := map[string]interface{}{
		"patterns":   []string{"*.txt"},
		"path":       tmpDir,
		"sort_by":    "name",
		"sort_order": "asc",
	}
	paramsJSON, _ := json.Marshal(params)

	result, err := tool.Execute(ctx, paramsJSON)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	globResult := result.(*GlobResult)

	if len(globResult.Matches) != 3 {
		t.Fatalf("Expected 3 matches, got %d", len(globResult.Matches))
	}

	// Verify sorted order
	expected := []string{"alpha.txt", "bravo.txt", "charlie.txt"}
	for i, match := range globResult.Matches {
		if match.Name != expected[i] {
			t.Errorf("Expected match %d to be %s, got %s", i, expected[i], match.Name)
		}
	}
}

func TestGlobTool_Execute_SortBySize(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files with different sizes
	_ = os.WriteFile(filepath.Join(tmpDir, "small.txt"), []byte("a"), 0644)          // 1 byte
	_ = os.WriteFile(filepath.Join(tmpDir, "large.txt"), []byte("aaaaaa"), 0644)     // 6 bytes
	_ = os.WriteFile(filepath.Join(tmpDir, "medium.txt"), []byte("aaa"), 0644)       // 3 bytes

	tool := NewGlobTool()
	ctx := context.Background()

	params := map[string]interface{}{
		"patterns":     []string{"*.txt"},
		"path":         tmpDir,
		"sort_by":      "size",
		"sort_order":   "asc",
		"include_info": true,
	}
	paramsJSON, _ := json.Marshal(params)

	result, err := tool.Execute(ctx, paramsJSON)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	globResult := result.(*GlobResult)

	if len(globResult.Matches) != 3 {
		t.Fatalf("Expected 3 matches, got %d", len(globResult.Matches))
	}

	// Verify sorted by size ascending
	if globResult.Matches[0].Name != "small.txt" {
		t.Errorf("Expected first match to be small.txt, got %s", globResult.Matches[0].Name)
	}

	if globResult.Matches[2].Name != "large.txt" {
		t.Errorf("Expected last match to be large.txt, got %s", globResult.Matches[2].Name)
	}
}

func TestGlobTool_Execute_SortDescending(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files
	_ = os.WriteFile(filepath.Join(tmpDir, "a.txt"), []byte("content"), 0644)
	_ = os.WriteFile(filepath.Join(tmpDir, "b.txt"), []byte("content"), 0644)
	_ = os.WriteFile(filepath.Join(tmpDir, "c.txt"), []byte("content"), 0644)

	tool := NewGlobTool()
	ctx := context.Background()

	params := map[string]interface{}{
		"patterns":   []string{"*.txt"},
		"path":       tmpDir,
		"sort_by":    "name",
		"sort_order": "desc",
	}
	paramsJSON, _ := json.Marshal(params)

	result, err := tool.Execute(ctx, paramsJSON)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	globResult := result.(*GlobResult)

	// Verify sorted descending
	if globResult.Matches[0].Name != "c.txt" {
		t.Errorf("Expected first match to be c.txt, got %s", globResult.Matches[0].Name)
	}

	if globResult.Matches[2].Name != "a.txt" {
		t.Errorf("Expected last match to be a.txt, got %s", globResult.Matches[2].Name)
	}
}

func TestGlobTool_Execute_NoMatches(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a file that won't match
	_ = os.WriteFile(filepath.Join(tmpDir, "file.txt"), []byte("content"), 0644)

	tool := NewGlobTool()
	ctx := context.Background()

	params := map[string]interface{}{
		"patterns": []string{"*.md"},
		"path":     tmpDir,
	}
	paramsJSON, _ := json.Marshal(params)

	result, err := tool.Execute(ctx, paramsJSON)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	globResult := result.(*GlobResult)

	if globResult.Count != 0 {
		t.Errorf("Expected 0 matches, got %d", globResult.Count)
	}
}
