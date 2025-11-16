package tools

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestGrepTool_Name(t *testing.T) {
	tool := NewGrepTool()
	if tool.Name() != "grep" {
		t.Errorf("Expected name 'grep', got '%s'", tool.Name())
	}
}

func TestGrepTool_Validate(t *testing.T) {
	tool := NewGrepTool()

	tests := []struct {
		name    string
		params  string
		wantErr bool
	}{
		{
			name:    "valid params",
			params:  `{"pattern": "test", "files": ["file.txt"]}`,
			wantErr: false,
		},
		{
			name:    "valid with options",
			params:  `{"pattern": "test", "files": ["*.txt"], "ignore_case": true, "line_numbers": true}`,
			wantErr: false,
		},
		{
			name:    "missing pattern",
			params:  `{"files": ["file.txt"]}`,
			wantErr: true,
		},
		{
			name:    "missing files",
			params:  `{"pattern": "test"}`,
			wantErr: true,
		},
		{
			name:    "empty files array",
			params:  `{"pattern": "test", "files": []}`,
			wantErr: true,
		},
		{
			name:    "invalid regex",
			params:  `{"pattern": "[invalid(", "files": ["file.txt"]}`,
			wantErr: true,
		},
		{
			name:    "invalid output_format",
			params:  `{"pattern": "test", "files": ["file.txt"], "output_format": "invalid"}`,
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

func TestGrepTool_Execute_BasicSearch(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	content := "line1\ntest line\nline3\nanother test\nline5"

	_ = os.WriteFile(testFile, []byte(content), 0644)

	tool := NewGrepTool()
	ctx := context.Background()

	params := map[string]interface{}{
		"pattern": "test",
		"files":   []string{testFile},
	}
	paramsJSON, _ := json.Marshal(params)

	result, err := tool.Execute(ctx, paramsJSON)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	grepResult, ok := result.(*GrepResult)
	if !ok {
		t.Fatal("Result is not a GrepResult")
	}

	if grepResult.TotalMatches != 2 {
		t.Errorf("Expected 2 matches, got %d", grepResult.TotalMatches)
	}

	if grepResult.Statistics.FilesSearched != 1 {
		t.Errorf("Expected 1 file searched, got %d", grepResult.Statistics.FilesSearched)
	}
}

func TestGrepTool_Execute_WithLineNumbers(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	content := "line1\ntest line 2\nline3"

	_ = os.WriteFile(testFile, []byte(content), 0644)

	tool := NewGrepTool()
	ctx := context.Background()

	params := map[string]interface{}{
		"pattern":      "test",
		"files":        []string{testFile},
		"line_numbers": true,
	}
	paramsJSON, _ := json.Marshal(params)

	result, err := tool.Execute(ctx, paramsJSON)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	grepResult := result.(*GrepResult)

	if len(grepResult.Matches) != 1 {
		t.Fatalf("Expected 1 match, got %d", len(grepResult.Matches))
	}

	match := grepResult.Matches[0]

	if match.Line != 2 {
		t.Errorf("Expected line number 2, got %d", match.Line)
	}
}

func TestGrepTool_Execute_CaseInsensitive(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	content := "TEST\nTest\ntest\nTeSt"

	_ = os.WriteFile(testFile, []byte(content), 0644)

	tool := NewGrepTool()
	ctx := context.Background()

	params := map[string]interface{}{
		"pattern":     "test",
		"files":       []string{testFile},
		"ignore_case": true,
	}
	paramsJSON, _ := json.Marshal(params)

	result, err := tool.Execute(ctx, paramsJSON)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	grepResult := result.(*GrepResult)

	if grepResult.TotalMatches != 4 {
		t.Errorf("Expected 4 matches (case insensitive), got %d", grepResult.TotalMatches)
	}
}

func TestGrepTool_Execute_WordMatch(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	content := "test\ntesting\ntest123\nthe test"

	_ = os.WriteFile(testFile, []byte(content), 0644)

	tool := NewGrepTool()
	ctx := context.Background()

	params := map[string]interface{}{
		"pattern":    "test",
		"files":      []string{testFile},
		"word_match": true,
	}
	paramsJSON, _ := json.Marshal(params)

	result, err := tool.Execute(ctx, paramsJSON)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	grepResult := result.(*GrepResult)

	// Should only match "test" and "the test" (whole words), not "testing" or "test123"
	if grepResult.TotalMatches != 2 {
		t.Errorf("Expected 2 word matches, got %d", grepResult.TotalMatches)
	}
}

func TestGrepTool_Execute_InvertMatch(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	content := "line1\ntest line\nline3\nline4"

	_ = os.WriteFile(testFile, []byte(content), 0644)

	tool := NewGrepTool()
	ctx := context.Background()

	params := map[string]interface{}{
		"pattern":      "test",
		"files":        []string{testFile},
		"invert_match": true,
	}
	paramsJSON, _ := json.Marshal(params)

	result, err := tool.Execute(ctx, paramsJSON)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	grepResult := result.(*GrepResult)

	// Should match lines without "test" (line1, line3, line4)
	if grepResult.TotalMatches != 3 {
		t.Errorf("Expected 3 non-matching lines, got %d", grepResult.TotalMatches)
	}
}

func TestGrepTool_Execute_CaptureGroups(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	content := "user: john, age: 30\nuser: jane, age: 25"

	_ = os.WriteFile(testFile, []byte(content), 0644)

	tool := NewGrepTool()
	ctx := context.Background()

	params := map[string]interface{}{
		"pattern": `user: (\w+), age: (\d+)`,
		"files":   []string{testFile},
	}
	paramsJSON, _ := json.Marshal(params)

	result, err := tool.Execute(ctx, paramsJSON)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	grepResult := result.(*GrepResult)

	if len(grepResult.Matches) != 2 {
		t.Fatalf("Expected 2 matches, got %d", len(grepResult.Matches))
	}

	firstMatch := grepResult.Matches[0]

	if len(firstMatch.Captures) != 2 {
		t.Errorf("Expected 2 capture groups, got %d", len(firstMatch.Captures))
	}

	if firstMatch.Captures[0] != "john" {
		t.Errorf("Expected first capture 'john', got '%s'", firstMatch.Captures[0])
	}

	if firstMatch.Captures[1] != "30" {
		t.Errorf("Expected second capture '30', got '%s'", firstMatch.Captures[1])
	}
}

func TestGrepTool_Execute_CountOnly(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	content := "test\ntest\ntest"

	_ = os.WriteFile(testFile, []byte(content), 0644)

	tool := NewGrepTool()
	ctx := context.Background()

	params := map[string]interface{}{
		"pattern": "test",
		"files":   []string{testFile},
		"count":   true,
	}
	paramsJSON, _ := json.Marshal(params)

	result, err := tool.Execute(ctx, paramsJSON)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	grepResult := result.(*GrepResult)

	if grepResult.TotalMatches != 3 {
		t.Errorf("Expected 3 matches, got %d", grepResult.TotalMatches)
	}

	// Matches array should be empty in count mode
	if len(grepResult.Matches) != 0 {
		t.Errorf("Expected empty matches array in count mode, got %d", len(grepResult.Matches))
	}
}

func TestGrepTool_Execute_MaxMatches(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	content := "test\ntest\ntest\ntest\ntest"

	_ = os.WriteFile(testFile, []byte(content), 0644)

	tool := NewGrepTool()
	ctx := context.Background()

	params := map[string]interface{}{
		"pattern":     "test",
		"files":       []string{testFile},
		"max_matches": 2,
	}
	paramsJSON, _ := json.Marshal(params)

	result, err := tool.Execute(ctx, paramsJSON)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	grepResult := result.(*GrepResult)

	if grepResult.TotalMatches != 2 {
		t.Errorf("Expected 2 matches (max_matches), got %d", grepResult.TotalMatches)
	}
}

func TestGrepTool_Execute_Statistics(t *testing.T) {
	tmpDir := t.TempDir()

	// Create multiple files
	_ = os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte("test\ntest"), 0644)
	_ = os.WriteFile(filepath.Join(tmpDir, "file2.txt"), []byte("no match"), 0644)
	_ = os.WriteFile(filepath.Join(tmpDir, "file3.txt"), []byte("test"), 0644)

	tool := NewGrepTool()
	ctx := context.Background()

	params := map[string]interface{}{
		"pattern": "test",
		"files":   []string{filepath.Join(tmpDir, "*.txt")},
	}
	paramsJSON, _ := json.Marshal(params)

	result, err := tool.Execute(ctx, paramsJSON)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	grepResult := result.(*GrepResult)

	if grepResult.Statistics == nil {
		t.Fatal("Expected statistics to be set")
	}

	stats := grepResult.Statistics

	if stats.FilesSearched != 3 {
		t.Errorf("Expected 3 files searched, got %d", stats.FilesSearched)
	}

	if stats.FilesMatched != 2 {
		t.Errorf("Expected 2 files matched, got %d", stats.FilesMatched)
	}

	if stats.TotalMatches != 3 {
		t.Errorf("Expected 3 total matches, got %d", stats.TotalMatches)
	}
}

func TestGrepTool_Execute_MultipleFiles(t *testing.T) {
	tmpDir := t.TempDir()

	file1 := filepath.Join(tmpDir, "file1.txt")
	file2 := filepath.Join(tmpDir, "file2.txt")

	_ = os.WriteFile(file1, []byte("test in file 1"), 0644)
	_ = os.WriteFile(file2, []byte("test in file 2"), 0644)

	tool := NewGrepTool()
	ctx := context.Background()

	params := map[string]interface{}{
		"pattern": "test",
		"files":   []string{file1, file2},
	}
	paramsJSON, _ := json.Marshal(params)

	result, err := tool.Execute(ctx, paramsJSON)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	grepResult := result.(*GrepResult)

	if grepResult.TotalMatches != 2 {
		t.Errorf("Expected 2 matches, got %d", grepResult.TotalMatches)
	}

	// Verify files are different
	files := make(map[string]bool)
	for _, match := range grepResult.Matches {
		files[match.File] = true
	}

	if len(files) != 2 {
		t.Errorf("Expected matches from 2 different files, got %d", len(files))
	}
}

func TestGrepTool_Execute_GlobPattern(t *testing.T) {
	tmpDir := t.TempDir()

	_ = os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte("test"), 0644)
	_ = os.WriteFile(filepath.Join(tmpDir, "file2.txt"), []byte("test"), 0644)
	_ = os.WriteFile(filepath.Join(tmpDir, "file3.md"), []byte("test"), 0644)

	tool := NewGrepTool()
	ctx := context.Background()

	params := map[string]interface{}{
		"pattern": "test",
		"files":   []string{filepath.Join(tmpDir, "*.txt")},
	}
	paramsJSON, _ := json.Marshal(params)

	result, err := tool.Execute(ctx, paramsJSON)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	grepResult := result.(*GrepResult)

	if grepResult.Statistics.FilesSearched != 2 {
		t.Errorf("Expected 2 txt files searched, got %d", grepResult.Statistics.FilesSearched)
	}
}

func TestGrepTool_Execute_NoMatches(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	content := "line1\nline2\nline3"

	_ = os.WriteFile(testFile, []byte(content), 0644)

	tool := NewGrepTool()
	ctx := context.Background()

	params := map[string]interface{}{
		"pattern": "notfound",
		"files":   []string{testFile},
	}
	paramsJSON, _ := json.Marshal(params)

	result, err := tool.Execute(ctx, paramsJSON)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	grepResult := result.(*GrepResult)

	if grepResult.TotalMatches != 0 {
		t.Errorf("Expected 0 matches, got %d", grepResult.TotalMatches)
	}
}

func TestGrepTool_Execute_PatternCounts(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	content := "error: connection failed\nerror: timeout\nwarning: deprecated"

	_ = os.WriteFile(testFile, []byte(content), 0644)

	tool := NewGrepTool()
	ctx := context.Background()

	params := map[string]interface{}{
		"pattern": `(error|warning)`,
		"files":   []string{testFile},
	}
	paramsJSON, _ := json.Marshal(params)

	result, err := tool.Execute(ctx, paramsJSON)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	grepResult := result.(*GrepResult)

	if grepResult.Statistics.PatternCounts["error"] != 2 {
		t.Errorf("Expected 2 'error' matches, got %d", grepResult.Statistics.PatternCounts["error"])
	}

	if grepResult.Statistics.PatternCounts["warning"] != 1 {
		t.Errorf("Expected 1 'warning' match, got %d", grepResult.Statistics.PatternCounts["warning"])
	}
}

func TestGrepTool_Execute_MultipleCaptures(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	content := "Name: John, Age: 30\nName: Jane, Age: 25"

	_ = os.WriteFile(testFile, []byte(content), 0644)

	tool := NewGrepTool()
	ctx := context.Background()

	params := map[string]interface{}{
		"pattern": `Name: (\w+), Age: (\d+)`,
		"files":   []string{testFile},
	}
	paramsJSON, _ := json.Marshal(params)

	result, err := tool.Execute(ctx, paramsJSON)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	grepResult := result.(*GrepResult)

	if len(grepResult.Matches) != 2 {
		t.Fatalf("Expected 2 matches, got %d", len(grepResult.Matches))
	}

	// Check first match captures
	if len(grepResult.Matches[0].Captures) != 2 {
		t.Errorf("Expected 2 captures, got %d", len(grepResult.Matches[0].Captures))
	}

	if grepResult.Matches[0].Captures[0] != "John" {
		t.Errorf("Expected first capture 'John', got '%s'", grepResult.Matches[0].Captures[0])
	}

	if grepResult.Matches[0].Captures[1] != "30" {
		t.Errorf("Expected second capture '30', got '%s'", grepResult.Matches[0].Captures[1])
	}
}
