package tools

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestSearchTool_Name(t *testing.T) {
	tool := NewSearchTool()
	if tool.Name() != "search" {
		t.Errorf("Expected name 'search', got '%s'", tool.Name())
	}
}

func TestSearchTool_Validate(t *testing.T) {
	tool := NewSearchTool()

	tests := []struct {
		name    string
		params  string
		wantErr bool
	}{
		{
			name:    "valid params",
			params:  `{"pattern": "test", "path": "/tmp"}`,
			wantErr: false,
		},
		{
			name:    "valid with regex",
			params:  `{"pattern": "func.*Error", "path": "/tmp"}`,
			wantErr: false,
		},
		{
			name:    "valid with context",
			params:  `{"pattern": "test", "path": "/tmp", "context": 2}`,
			wantErr: false,
		},
		{
			name:    "missing pattern",
			params:  `{"path": "/tmp"}`,
			wantErr: true,
		},
		{
			name:    "missing path",
			params:  `{"pattern": "test"}`,
			wantErr: true,
		},
		{
			name:    "invalid regex",
			params:  `{"pattern": "[invalid(", "path": "/tmp"}`,
			wantErr: true,
		},
		{
			name:    "negative context",
			params:  `{"pattern": "test", "path": "/tmp", "context": -1}`,
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

func TestSearchTool_Execute_SingleFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	content := "line1\ntest line 2\nline3\nanother test line\nline5"

	_ = os.WriteFile(testFile, []byte(content), 0644)

	tool := NewSearchTool()
	ctx := context.Background()

	params := map[string]interface{}{
		"pattern": "test",
		"path":    testFile,
	}
	paramsJSON, _ := json.Marshal(params)

	result, err := tool.Execute(ctx, paramsJSON)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	searchResult, ok := result.(*SearchResult)
	if !ok {
		t.Fatal("Result is not a SearchResult")
	}

	if searchResult.TotalMatches != 2 {
		t.Errorf("Expected 2 matches, got %d", searchResult.TotalMatches)
	}

	if searchResult.TotalFiles != 1 {
		t.Errorf("Expected 1 file searched, got %d", searchResult.TotalFiles)
	}

	// Verify match details
	if len(searchResult.Matches) != 2 {
		t.Fatalf("Expected 2 matches in array, got %d", len(searchResult.Matches))
	}

	firstMatch := searchResult.Matches[0]
	if firstMatch.Line != 2 {
		t.Errorf("Expected first match on line 2, got %d", firstMatch.Line)
	}

	if firstMatch.MatchText != "test" {
		t.Errorf("Expected match text 'test', got '%s'", firstMatch.MatchText)
	}
}

func TestSearchTool_Execute_WithContext(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	content := "line1\nline2\ntest line 3\nline4\nline5"

	_ = os.WriteFile(testFile, []byte(content), 0644)

	tool := NewSearchTool()
	ctx := context.Background()

	params := map[string]interface{}{
		"pattern": "test",
		"path":    testFile,
		"context": 1,
	}
	paramsJSON, _ := json.Marshal(params)

	result, err := tool.Execute(ctx, paramsJSON)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	searchResult := result.(*SearchResult)

	if len(searchResult.Matches) != 1 {
		t.Fatalf("Expected 1 match, got %d", len(searchResult.Matches))
	}

	match := searchResult.Matches[0]

	// Should have 3 context lines (1 before, match line, 1 after)
	if len(match.Context) != 3 {
		t.Errorf("Expected 3 context lines, got %d", len(match.Context))
	}

	if match.Context[0] != "line2" {
		t.Errorf("Expected first context line 'line2', got '%s'", match.Context[0])
	}

	if match.Context[1] != "test line 3" {
		t.Errorf("Expected match line 'test line 3', got '%s'", match.Context[1])
	}

	if match.Context[2] != "line4" {
		t.Errorf("Expected last context line 'line4', got '%s'", match.Context[2])
	}
}

func TestSearchTool_Execute_CaseInsensitive(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	content := "TEST\nTest\ntest\nTeSt"

	_ = os.WriteFile(testFile, []byte(content), 0644)

	tool := NewSearchTool()
	ctx := context.Background()

	params := map[string]interface{}{
		"pattern":     "test",
		"path":        testFile,
		"ignore_case": true,
	}
	paramsJSON, _ := json.Marshal(params)

	result, err := tool.Execute(ctx, paramsJSON)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	searchResult := result.(*SearchResult)

	if searchResult.TotalMatches != 4 {
		t.Errorf("Expected 4 matches (case insensitive), got %d", searchResult.TotalMatches)
	}
}

func TestSearchTool_Execute_Recursive(t *testing.T) {
	tmpDir := t.TempDir()

	// Create files in root
	_ = os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte("test in file 1"), 0644)

	// Create subdirectory with files
	subDir := filepath.Join(tmpDir, "subdir")
	_ = os.Mkdir(subDir, 0755)
	_ = os.WriteFile(filepath.Join(subDir, "file2.txt"), []byte("test in file 2"), 0644)

	// Create nested directory with files
	nestedDir := filepath.Join(subDir, "nested")
	_ = os.Mkdir(nestedDir, 0755)
	_ = os.WriteFile(filepath.Join(nestedDir, "file3.txt"), []byte("test in file 3"), 0644)

	tool := NewSearchTool()
	ctx := context.Background()

	params := map[string]interface{}{
		"pattern":   "test",
		"path":      tmpDir,
		"recursive": true,
	}
	paramsJSON, _ := json.Marshal(params)

	result, err := tool.Execute(ctx, paramsJSON)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	searchResult := result.(*SearchResult)

	if searchResult.TotalMatches != 3 {
		t.Errorf("Expected 3 matches, got %d", searchResult.TotalMatches)
	}

	if searchResult.TotalFiles != 3 {
		t.Errorf("Expected 3 files searched, got %d", searchResult.TotalFiles)
	}
}

func TestSearchTool_Execute_WithFilePattern(t *testing.T) {
	tmpDir := t.TempDir()

	_ = os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte("test content"), 0644)
	_ = os.WriteFile(filepath.Join(tmpDir, "file2.go"), []byte("test content"), 0644)
	_ = os.WriteFile(filepath.Join(tmpDir, "file3.txt"), []byte("test content"), 0644)

	tool := NewSearchTool()
	ctx := context.Background()

	params := map[string]interface{}{
		"pattern":      "test",
		"path":         tmpDir,
		"file_pattern": "*.txt",
	}
	paramsJSON, _ := json.Marshal(params)

	result, err := tool.Execute(ctx, paramsJSON)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	searchResult := result.(*SearchResult)

	if searchResult.TotalFiles != 2 {
		t.Errorf("Expected 2 txt files, got %d", searchResult.TotalFiles)
	}
}

func TestSearchTool_Execute_WithIncludeExclude(t *testing.T) {
	tmpDir := t.TempDir()

	_ = os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte("test content"), 0644)
	_ = os.WriteFile(filepath.Join(tmpDir, "file2.md"), []byte("test content"), 0644)
	_ = os.WriteFile(filepath.Join(tmpDir, "file3.tmp"), []byte("test content"), 0644)

	tool := NewSearchTool()
	ctx := context.Background()

	params := map[string]interface{}{
		"pattern": "test",
		"path":    tmpDir,
		"include": []string{".txt", ".md"},
		"exclude": []string{".tmp"},
	}
	paramsJSON, _ := json.Marshal(params)

	result, err := tool.Execute(ctx, paramsJSON)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	searchResult := result.(*SearchResult)

	if searchResult.TotalFiles != 2 {
		t.Errorf("Expected 2 files, got %d", searchResult.TotalFiles)
	}
}

func TestSearchTool_Execute_MaxResults(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	content := "test\ntest\ntest\ntest\ntest"

	_ = os.WriteFile(testFile, []byte(content), 0644)

	tool := NewSearchTool()
	ctx := context.Background()

	params := map[string]interface{}{
		"pattern":     "test",
		"path":        testFile,
		"max_results": 2,
	}
	paramsJSON, _ := json.Marshal(params)

	result, err := tool.Execute(ctx, paramsJSON)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	searchResult := result.(*SearchResult)

	if searchResult.TotalMatches != 2 {
		t.Errorf("Expected 2 matches (max_results), got %d", searchResult.TotalMatches)
	}
}

func TestSearchTool_Execute_RegexPattern(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")
	content := "func main() {}\nfunc testFunc() {}\nvar x = 10\nfunc anotherFunc() {}"

	_ = os.WriteFile(testFile, []byte(content), 0644)

	tool := NewSearchTool()
	ctx := context.Background()

	params := map[string]interface{}{
		"pattern": `func\s+\w+`,
		"path":    testFile,
	}
	paramsJSON, _ := json.Marshal(params)

	result, err := tool.Execute(ctx, paramsJSON)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	searchResult := result.(*SearchResult)

	if searchResult.TotalMatches != 3 {
		t.Errorf("Expected 3 function matches, got %d", searchResult.TotalMatches)
	}
}

func TestSearchTool_Execute_NoMatches(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	content := "line1\nline2\nline3"

	_ = os.WriteFile(testFile, []byte(content), 0644)

	tool := NewSearchTool()
	ctx := context.Background()

	params := map[string]interface{}{
		"pattern": "notfound",
		"path":    testFile,
	}
	paramsJSON, _ := json.Marshal(params)

	result, err := tool.Execute(ctx, paramsJSON)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	searchResult := result.(*SearchResult)

	if searchResult.TotalMatches != 0 {
		t.Errorf("Expected 0 matches, got %d", searchResult.TotalMatches)
	}
}

func TestSearchTool_Execute_MatchPosition(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	content := "This is a test line"

	_ = os.WriteFile(testFile, []byte(content), 0644)

	tool := NewSearchTool()
	ctx := context.Background()

	params := map[string]interface{}{
		"pattern": "test",
		"path":    testFile,
	}
	paramsJSON, _ := json.Marshal(params)

	result, err := tool.Execute(ctx, paramsJSON)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	searchResult := result.(*SearchResult)

	if len(searchResult.Matches) != 1 {
		t.Fatalf("Expected 1 match, got %d", len(searchResult.Matches))
	}

	match := searchResult.Matches[0]

	if match.Column != 11 {
		t.Errorf("Expected column 11, got %d", match.Column)
	}

	if match.MatchText != "test" {
		t.Errorf("Expected match text 'test', got '%s'", match.MatchText)
	}
}
