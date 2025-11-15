package tools

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// GrepTool implements advanced pattern matching
type GrepTool struct{}

// GrepParams defines the parameters for the Grep tool
type GrepParams struct {
	Pattern      string   `json:"pattern"`
	Files        []string `json:"files"`
	Recursive    bool     `json:"recursive,omitempty"`
	IgnoreCase   bool     `json:"ignore_case,omitempty"`
	WordMatch    bool     `json:"word_match,omitempty"`    // Match whole words only
	InvertMatch  bool     `json:"invert_match,omitempty"`  // Show non-matching lines
	LineNumbers  bool     `json:"line_numbers,omitempty"`
	Count        bool     `json:"count,omitempty"`         // Only show counts
	MaxMatches   int      `json:"max_matches,omitempty"`
	OutputFormat string   `json:"output_format,omitempty"` // "text", "json", "csv"
}

// GrepResult represents the result of a grep operation
type GrepResult struct {
	Matches      []GrepMatch     `json:"matches,omitempty"`
	Statistics   *GrepStatistics `json:"statistics,omitempty"`
	TotalMatches int             `json:"total_matches"`
}

// GrepMatch represents a single grep match
type GrepMatch struct {
	File     string   `json:"file"`
	Line     int      `json:"line"`
	Content  string   `json:"content"`
	Captures []string `json:"captures,omitempty"` // Regex capture groups
}

// GrepStatistics provides statistics about the grep operation
type GrepStatistics struct {
	FilesSearched int            `json:"files_searched"`
	FilesMatched  int            `json:"files_matched"`
	TotalMatches  int            `json:"total_matches"`
	PatternCounts map[string]int `json:"pattern_counts,omitempty"`
}

// NewGrepTool creates a new Grep tool instance
func NewGrepTool() *GrepTool {
	return &GrepTool{}
}

// Name returns the tool's name
func (t *GrepTool) Name() string {
	return "grep"
}

// Description returns the tool's description
func (t *GrepTool) Description() string {
	return "Advanced pattern matching with capture groups, statistics, and flexible output formats"
}

// Schema returns the JSON schema for the tool's parameters
func (t *GrepTool) Schema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"pattern": map[string]interface{}{
				"type":        "string",
				"description": "Regular expression pattern to search for",
			},
			"files": map[string]interface{}{
				"type":        "array",
				"description": "Files or glob patterns to search",
				"items": map[string]interface{}{
					"type": "string",
				},
			},
			"recursive": map[string]interface{}{
				"type":        "boolean",
				"description": "Search directories recursively",
			},
			"ignore_case": map[string]interface{}{
				"type":        "boolean",
				"description": "Case-insensitive matching",
			},
			"word_match": map[string]interface{}{
				"type":        "boolean",
				"description": "Match whole words only",
			},
			"invert_match": map[string]interface{}{
				"type":        "boolean",
				"description": "Show lines that don't match the pattern",
			},
			"line_numbers": map[string]interface{}{
				"type":        "boolean",
				"description": "Include line numbers in results",
			},
			"count": map[string]interface{}{
				"type":        "boolean",
				"description": "Only return match counts, not actual matches",
			},
			"max_matches": map[string]interface{}{
				"type":        "integer",
				"description": "Maximum number of matches to return",
				"minimum":     0,
			},
			"output_format": map[string]interface{}{
				"type":        "string",
				"description": "Output format: text, json, or csv",
				"enum":        []string{"text", "json", "csv"},
			},
		},
		"required": []string{"pattern", "files"},
	}
}

// Validate checks if the parameters are valid
func (t *GrepTool) Validate(params json.RawMessage) error {
	var p GrepParams
	if err := json.Unmarshal(params, &p); err != nil {
		return fmt.Errorf("invalid parameters: %w", err)
	}

	if p.Pattern == "" {
		return fmt.Errorf("pattern is required")
	}

	if len(p.Files) == 0 {
		return fmt.Errorf("at least one file is required")
	}

	// Validate regex pattern
	pattern := p.Pattern
	if p.IgnoreCase {
		pattern = "(?i)" + pattern
	}
	if p.WordMatch {
		pattern = `\b` + pattern + `\b`
	}

	if _, err := regexp.Compile(pattern); err != nil {
		return fmt.Errorf("invalid regex pattern: %w", err)
	}

	if p.MaxMatches < 0 {
		return fmt.Errorf("max_matches must be non-negative")
	}

	if p.OutputFormat != "" && p.OutputFormat != "text" && p.OutputFormat != "json" && p.OutputFormat != "csv" {
		return fmt.Errorf("output_format must be 'text', 'json', or 'csv'")
	}

	return nil
}

// Execute runs the tool with the given parameters
func (t *GrepTool) Execute(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var p GrepParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}

	// Build regex pattern
	pattern := p.Pattern
	if p.IgnoreCase {
		pattern = "(?i)" + pattern
	}
	if p.WordMatch {
		pattern = `\b` + pattern + `\b`
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid regex pattern: %w", err)
	}

	// Expand file patterns
	var files []string
	for _, filePattern := range p.Files {
		if strings.Contains(filePattern, "*") {
			// It's a glob pattern
			matches, err := t.expandGlob(filePattern, p.Recursive)
			if err != nil {
				continue // Skip invalid patterns
			}
			files = append(files, matches...)
		} else {
			files = append(files, filePattern)
		}
	}

	// Search files
	var allMatches []GrepMatch
	stats := &GrepStatistics{
		PatternCounts: make(map[string]int),
	}

	filesMatched := make(map[string]bool)

	for _, file := range files {
		stats.FilesSearched++

		matches, err := t.grepFile(file, re, p)
		if err != nil {
			continue // Skip files we can't read
		}

		if len(matches) > 0 {
			filesMatched[file] = true
			allMatches = append(allMatches, matches...)

			// Update pattern counts
			for _, match := range matches {
				if len(match.Captures) > 0 {
					stats.PatternCounts[match.Captures[0]]++
				} else {
					stats.PatternCounts[match.Content]++
				}
			}
		}

		// Check max matches
		if p.MaxMatches > 0 && len(allMatches) >= p.MaxMatches {
			break
		}
	}

	// Trim to max matches
	if p.MaxMatches > 0 && len(allMatches) > p.MaxMatches {
		allMatches = allMatches[:p.MaxMatches]
	}

	stats.FilesMatched = len(filesMatched)
	stats.TotalMatches = len(allMatches)

	result := &GrepResult{
		Statistics:   stats,
		TotalMatches: len(allMatches),
	}

	// Only include matches if not count-only
	if !p.Count {
		result.Matches = allMatches
	}

	return result, nil
}

// grepFile searches a single file
func (t *GrepTool) grepFile(filePath string, re *regexp.Regexp, p GrepParams) ([]GrepMatch, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var matches []GrepMatch
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		matched := re.MatchString(line)

		// Invert match if requested
		if p.InvertMatch {
			matched = !matched
		}

		if matched {
			match := GrepMatch{
				File:    filePath,
				Content: line,
			}

			if p.LineNumbers {
				match.Line = lineNum
			}

			// Extract capture groups
			if !p.InvertMatch {
				captures := re.FindStringSubmatch(line)
				if len(captures) > 1 {
					match.Captures = captures[1:] // Skip full match
				}
			}

			matches = append(matches, match)
		}
	}

	return matches, scanner.Err()
}

// expandGlob expands a glob pattern to actual files
func (t *GrepTool) expandGlob(pattern string, recursive bool) ([]string, error) {
	// Handle ** patterns
	if strings.Contains(pattern, "**") {
		return t.expandRecursiveGlob(pattern)
	}

	// Use standard glob
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}

	// Filter out directories
	var files []string
	for _, match := range matches {
		info, err := os.Stat(match)
		if err != nil {
			continue
		}
		if !info.IsDir() {
			files = append(files, match)
		}
	}

	return files, nil
}

// expandRecursiveGlob handles ** patterns
func (t *GrepTool) expandRecursiveGlob(pattern string) ([]string, error) {
	parts := strings.Split(pattern, "**")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid ** pattern")
	}

	basePath := filepath.Clean(parts[0])
	suffix := strings.TrimPrefix(parts[1], string(filepath.Separator))

	var files []string

	err := filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		if suffix == "" {
			files = append(files, path)
			return nil
		}

		matched, _ := filepath.Match(suffix, filepath.Base(path))
		if matched {
			files = append(files, path)
		}

		return nil
	})

	return files, err
}
