package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// GlobTool implements advanced pattern matching for finding files
type GlobTool struct{}

// GlobParams defines the parameters for the Glob tool
type GlobParams struct {
	Patterns       []string `json:"patterns"`                        // Array of glob patterns
	Path           string   `json:"path,omitempty"`                  // Base path (defaults to current dir)
	CaseSensitive  bool     `json:"case_sensitive,omitempty"`        // Case-sensitive matching
	FollowSymlinks bool     `json:"follow_symlinks,omitempty"`       // Follow symbolic links
	SortBy         string   `json:"sort_by,omitempty"`               // "modtime", "name", "size"
	SortOrder      string   `json:"sort_order,omitempty"`            // "asc", "desc"
	IncludeInfo    bool     `json:"include_info,omitempty"`          // Include file metadata
}

// GlobResult represents the result of a glob operation
type GlobResult struct {
	Matches []FileInfo `json:"matches"`
	Count   int        `json:"count"`
	Pattern string     `json:"pattern,omitempty"` // Combined pattern for reference
}

// NewGlobTool creates a new Glob tool instance
func NewGlobTool() *GlobTool {
	return &GlobTool{}
}

// Name returns the tool's name
func (t *GlobTool) Name() string {
	return "glob"
}

// Description returns the tool's description
func (t *GlobTool) Description() string {
	return "Find files using advanced glob pattern matching with support for multiple patterns, sorting, and detailed file information"
}

// Schema returns the JSON schema for the tool's parameters
func (t *GlobTool) Schema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"patterns": map[string]interface{}{
				"type":        "array",
				"description": "Array of glob patterns to match (e.g., ['**/*.go', '**/*_test.go'])",
				"items": map[string]interface{}{
					"type": "string",
				},
				"minItems": 1,
			},
			"path": map[string]interface{}{
				"type":        "string",
				"description": "Base path to search from (defaults to current directory)",
			},
			"case_sensitive": map[string]interface{}{
				"type":        "boolean",
				"description": "Enable case-sensitive pattern matching",
			},
			"follow_symlinks": map[string]interface{}{
				"type":        "boolean",
				"description": "Follow symbolic links during search",
			},
			"sort_by": map[string]interface{}{
				"type":        "string",
				"description": "Sort results by: 'modtime', 'name', or 'size'",
				"enum":        []string{"modtime", "name", "size"},
			},
			"sort_order": map[string]interface{}{
				"type":        "string",
				"description": "Sort order: 'asc' or 'desc'",
				"enum":        []string{"asc", "desc"},
			},
			"include_info": map[string]interface{}{
				"type":        "boolean",
				"description": "Include detailed file metadata in results",
			},
		},
		"required": []string{"patterns"},
	}
}

// Validate checks if the parameters are valid
func (t *GlobTool) Validate(params json.RawMessage) error {
	var p GlobParams
	if err := json.Unmarshal(params, &p); err != nil {
		return fmt.Errorf("invalid parameters: %w", err)
	}

	if len(p.Patterns) == 0 {
		return fmt.Errorf("at least one pattern is required")
	}

	if p.SortBy != "" && p.SortBy != "modtime" && p.SortBy != "name" && p.SortBy != "size" {
		return fmt.Errorf("sort_by must be 'modtime', 'name', or 'size'")
	}

	if p.SortOrder != "" && p.SortOrder != "asc" && p.SortOrder != "desc" {
		return fmt.Errorf("sort_order must be 'asc' or 'desc'")
	}

	return nil
}

// Execute runs the tool with the given parameters
func (t *GlobTool) Execute(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var p GlobParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}

	// Default path to current directory
	if p.Path == "" {
		var err error
		p.Path, err = os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get current directory: %w", err)
		}
	}

	// Verify path exists
	if _, err := os.Stat(p.Path); err != nil {
		return nil, fmt.Errorf("path does not exist: %w", err)
	}

	// Collect matches from all patterns
	matchMap := make(map[string]FileInfo) // Use map to avoid duplicates

	for _, pattern := range p.Patterns {
		// Convert pattern to absolute if needed
		fullPattern := pattern
		if !filepath.IsAbs(pattern) {
			fullPattern = filepath.Join(p.Path, pattern)
		}

		matches, err := t.globPattern(fullPattern, p.CaseSensitive, p.FollowSymlinks, p.IncludeInfo)
		if err != nil {
			return nil, fmt.Errorf("failed to glob pattern '%s': %w", pattern, err)
		}

		for _, match := range matches {
			matchMap[match.Path] = match
		}
	}

	// Convert map to slice
	var matches []FileInfo
	for _, match := range matchMap {
		matches = append(matches, match)
	}

	// Sort results
	if p.SortBy != "" {
		t.sortMatches(matches, p.SortBy, p.SortOrder)
	} else {
		// Default sort by name ascending
		sort.Slice(matches, func(i, j int) bool {
			return matches[i].Name < matches[j].Name
		})
	}

	return &GlobResult{
		Matches: matches,
		Count:   len(matches),
		Pattern: strings.Join(p.Patterns, ", "),
	}, nil
}

// globPattern performs pattern matching using filepath.Glob with ** support
func (t *GlobTool) globPattern(pattern string, caseSensitive, followSymlinks, includeInfo bool) ([]FileInfo, error) {
	var matches []FileInfo

	// Check if pattern contains **
	if strings.Contains(pattern, "**") {
		// Use custom walking for ** patterns
		return t.globRecursive(pattern, caseSensitive, followSymlinks, includeInfo)
	}

	// Use standard filepath.Glob for simple patterns
	paths, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}

	for _, path := range paths {
		info, err := os.Stat(path)
		if err != nil {
			continue // Skip files we can't stat
		}

		fileInfo := FileInfo{
			Path:  path,
			Name:  filepath.Base(path),
			IsDir: info.IsDir(),
		}

		if includeInfo {
			fileInfo.Size = info.Size()
			fileInfo.Permissions = info.Mode().String()
			fileInfo.ModifiedTime = info.ModTime().Unix()
		}

		matches = append(matches, fileInfo)
	}

	return matches, nil
}

// globRecursive handles ** patterns by walking the directory tree
func (t *GlobTool) globRecursive(pattern string, caseSensitive, followSymlinks, includeInfo bool) ([]FileInfo, error) {
	var matches []FileInfo

	// Split pattern on **
	parts := strings.Split(pattern, "**")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid ** pattern: %s", pattern)
	}

	basePath := filepath.Clean(parts[0])
	suffix := strings.TrimPrefix(parts[1], string(filepath.Separator))

	err := filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}

		// Skip symlinks if not following them
		if !followSymlinks && info.Mode()&os.ModeSymlink != 0 {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Match suffix pattern
		relPath, err := filepath.Rel(basePath, path)
		if err != nil {
			return nil
		}

		matched := false
		if suffix == "" || suffix == "/" {
			matched = true
		} else {
			matched, _ = filepath.Match(suffix, filepath.Base(path))
			if !matched {
				// Try matching the full relative path
				matched, _ = filepath.Match(suffix, relPath)
			}
		}

		if matched {
			fileInfo := FileInfo{
				Path:  path,
				Name:  filepath.Base(path),
				IsDir: info.IsDir(),
			}

			if includeInfo {
				fileInfo.Size = info.Size()
				fileInfo.Permissions = info.Mode().String()
				fileInfo.ModifiedTime = info.ModTime().Unix()
			}

			matches = append(matches, fileInfo)
		}

		return nil
	})

	return matches, err
}

// sortMatches sorts the matches based on the specified criteria
func (t *GlobTool) sortMatches(matches []FileInfo, sortBy, sortOrder string) {
	ascending := sortOrder != "desc"

	switch sortBy {
	case "modtime":
		sort.Slice(matches, func(i, j int) bool {
			if ascending {
				return matches[i].ModifiedTime < matches[j].ModifiedTime
			}
			return matches[i].ModifiedTime > matches[j].ModifiedTime
		})
	case "size":
		sort.Slice(matches, func(i, j int) bool {
			if ascending {
				return matches[i].Size < matches[j].Size
			}
			return matches[i].Size > matches[j].Size
		})
	case "name":
		sort.Slice(matches, func(i, j int) bool {
			if ascending {
				return matches[i].Name < matches[j].Name
			}
			return matches[i].Name > matches[j].Name
		})
	}
}
