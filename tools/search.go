package tools

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

// SearchTool implements content-based file search
type SearchTool struct{}

// SearchParams defines the parameters for the Search tool
type SearchParams struct {
	Pattern     string   `json:"pattern"`              // Regex pattern
	Path        string   `json:"path"`
	Include     []string `json:"include,omitempty"`    // File patterns to include
	Exclude     []string `json:"exclude,omitempty"`    // File patterns to exclude
	Recursive   bool     `json:"recursive,omitempty"`
	IgnoreCase  bool     `json:"ignore_case,omitempty"`
	Context     int      `json:"context,omitempty"`    // Lines of context
	MaxResults  int      `json:"max_results,omitempty"`
	FilePattern string   `json:"file_pattern,omitempty"` // Glob pattern for files
}

// SearchResult represents the result of a search operation
type SearchResult struct {
	Matches      []SearchMatch `json:"matches"`
	TotalFiles   int           `json:"total_files"`
	TotalMatches int           `json:"total_matches"`
}

// SearchMatch represents a single search match
type SearchMatch struct {
	File      string   `json:"file"`
	Line      int      `json:"line"`
	Column    int      `json:"column"`
	MatchText string   `json:"match_text"`
	LineText  string   `json:"line_text"`
	Context   []string `json:"context,omitempty"`
}

// NewSearchTool creates a new Search tool instance
func NewSearchTool() *SearchTool {
	return &SearchTool{}
}

// Name returns the tool's name
func (t *SearchTool) Name() string {
	return "search"
}

// Description returns the tool's description
func (t *SearchTool) Description() string {
	return "Search for content in files using regular expressions with support for filtering, context lines, and recursive search"
}

// Schema returns the JSON schema for the tool's parameters
func (t *SearchTool) Schema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"pattern": map[string]interface{}{
				"type":        "string",
				"description": "Regular expression pattern to search for",
			},
			"path": map[string]interface{}{
				"type":        "string",
				"description": "Path to search (file or directory)",
			},
			"include": map[string]interface{}{
				"type":        "array",
				"description": "File patterns to include (e.g., ['.txt', '.md'])",
				"items": map[string]interface{}{
					"type": "string",
				},
			},
			"exclude": map[string]interface{}{
				"type":        "array",
				"description": "File patterns to exclude",
				"items": map[string]interface{}{
					"type": "string",
				},
			},
			"recursive": map[string]interface{}{
				"type":        "boolean",
				"description": "Search recursively in directories",
			},
			"ignore_case": map[string]interface{}{
				"type":        "boolean",
				"description": "Case-insensitive search",
			},
			"context": map[string]interface{}{
				"type":        "integer",
				"description": "Number of context lines to include around matches",
				"minimum":     0,
			},
			"max_results": map[string]interface{}{
				"type":        "integer",
				"description": "Maximum number of results to return (0 = unlimited)",
				"minimum":     0,
			},
			"file_pattern": map[string]interface{}{
				"type":        "string",
				"description": "Glob pattern for files to search (e.g., '*.go')",
			},
		},
		"required": []string{"pattern", "path"},
	}
}

// Validate checks if the parameters are valid
func (t *SearchTool) Validate(params json.RawMessage) error {
	var p SearchParams
	if err := json.Unmarshal(params, &p); err != nil {
		return fmt.Errorf("invalid parameters: %w", err)
	}

	if p.Pattern == "" {
		return fmt.Errorf("pattern is required")
	}

	if p.Path == "" {
		return fmt.Errorf("path is required")
	}

	// Validate regex pattern
	flags := ""
	if p.IgnoreCase {
		flags = "(?i)"
	}
	if _, err := regexp.Compile(flags + p.Pattern); err != nil {
		return fmt.Errorf("invalid regex pattern: %w", err)
	}

	if p.Context < 0 {
		return fmt.Errorf("context must be non-negative")
	}

	if p.MaxResults < 0 {
		return fmt.Errorf("max_results must be non-negative")
	}

	return nil
}

// Execute runs the tool with the given parameters
func (t *SearchTool) Execute(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var p SearchParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}

	// Compile regex
	flags := ""
	if p.IgnoreCase {
		flags = "(?i)"
	}
	re, err := regexp.Compile(flags + p.Pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid regex pattern: %w", err)
	}

	// Check if path exists
	info, err := os.Stat(p.Path)
	if err != nil {
		return nil, fmt.Errorf("cannot access path: %w", err)
	}

	var matches []SearchMatch
	filesSearched := 0

	if info.IsDir() {
		matches, filesSearched, err = t.searchDirectory(p.Path, re, p)
	} else {
		var fileMatches []SearchMatch
		fileMatches, err = t.searchFile(p.Path, re, p.Context)
		if err == nil {
			matches = fileMatches
			filesSearched = 1
		}
	}

	if err != nil {
		return nil, err
	}

	// Apply max results limit
	if p.MaxResults > 0 && len(matches) > p.MaxResults {
		matches = matches[:p.MaxResults]
	}

	return &SearchResult{
		Matches:      matches,
		TotalFiles:   filesSearched,
		TotalMatches: len(matches),
	}, nil
}

// searchDirectory searches all files in a directory
func (t *SearchTool) searchDirectory(dirPath string, re *regexp.Regexp, p SearchParams) ([]SearchMatch, int, error) {
	var allMatches []SearchMatch
	filesSearched := 0

	walkFunc := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip files with errors
		}

		if info.IsDir() {
			if !p.Recursive && path != dirPath {
				return filepath.SkipDir
			}
			return nil
		}

		// Check file pattern
		if p.FilePattern != "" {
			matched, _ := filepath.Match(p.FilePattern, filepath.Base(path))
			if !matched {
				return nil
			}
		}

		// Check include/exclude
		if len(p.Include) > 0 && !matchesExtensions(filepath.Base(path), p.Include) {
			return nil
		}
		if len(p.Exclude) > 0 && matchesExtensions(filepath.Base(path), p.Exclude) {
			return nil
		}

		// Search the file
		matches, err := t.searchFile(path, re, p.Context)
		if err != nil {
			return nil // Skip files that can't be read
		}

		allMatches = append(allMatches, matches...)
		filesSearched++

		// Check max results
		if p.MaxResults > 0 && len(allMatches) >= p.MaxResults {
			return filepath.SkipAll
		}

		return nil
	}

	err := filepath.Walk(dirPath, walkFunc)
	if err != nil && err != filepath.SkipAll {
		return nil, 0, err
	}

	// Trim to max results
	if p.MaxResults > 0 && len(allMatches) > p.MaxResults {
		allMatches = allMatches[:p.MaxResults]
	}

	return allMatches, filesSearched, nil
}

// searchFile searches a single file for matches
func (t *SearchTool) searchFile(filePath string, re *regexp.Regexp, contextLines int) ([]SearchMatch, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var matches []SearchMatch
	var lines []string
	scanner := bufio.NewScanner(file)
	lineNum := 0

	// Read all lines for context
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	// Search each line
	for i, line := range lines {
		lineNum = i + 1

		if re.MatchString(line) {
			// Find match position
			loc := re.FindStringIndex(line)
			matchText := ""
			column := 0
			if loc != nil {
				matchText = line[loc[0]:loc[1]]
				column = loc[0] + 1
			}

			match := SearchMatch{
				File:      filePath,
				Line:      lineNum,
				Column:    column,
				MatchText: matchText,
				LineText:  line,
			}

			// Add context lines
			if contextLines > 0 {
				start := i - contextLines
				if start < 0 {
					start = 0
				}
				end := i + contextLines + 1
				if end > len(lines) {
					end = len(lines)
				}

				match.Context = make([]string, 0, end-start)
				for j := start; j < end; j++ {
					match.Context = append(match.Context, lines[j])
				}
			}

			matches = append(matches, match)
		}
	}

	return matches, nil
}
