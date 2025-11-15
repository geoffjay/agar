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

// ListTool implements directory listing functionality
type ListTool struct{}

// ListParams defines the parameters for the List tool
type ListParams struct {
	Path      string   `json:"path"`
	Pattern   string   `json:"pattern,omitempty"`
	Recursive bool     `json:"recursive,omitempty"`
	Include   []string `json:"include,omitempty"` // File extensions to include
	Exclude   []string `json:"exclude,omitempty"` // File extensions to exclude
}

// ListResult represents the result of a list operation
type ListResult struct {
	Path  string     `json:"path"`
	Files []FileInfo `json:"files"`
	Count int        `json:"count"`
}

// NewListTool creates a new List tool instance
func NewListTool() *ListTool {
	return &ListTool{}
}

// Name returns the tool's name
func (t *ListTool) Name() string {
	return "list"
}

// Description returns the tool's description
func (t *ListTool) Description() string {
	return "List directory contents with filtering support including pattern matching, recursive listing, and file metadata"
}

// Schema returns the JSON schema for the tool's parameters
func (t *ListTool) Schema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"path": map[string]interface{}{
				"type":        "string",
				"description": "Path to the directory to list",
			},
			"pattern": map[string]interface{}{
				"type":        "string",
				"description": "Glob pattern to filter files (e.g., '*.txt')",
			},
			"recursive": map[string]interface{}{
				"type":        "boolean",
				"description": "List directories recursively",
			},
			"include": map[string]interface{}{
				"type":        "array",
				"description": "File extensions to include (e.g., ['.txt', '.md'])",
				"items": map[string]interface{}{
					"type": "string",
				},
			},
			"exclude": map[string]interface{}{
				"type":        "array",
				"description": "File extensions to exclude (e.g., ['.tmp', '.log'])",
				"items": map[string]interface{}{
					"type": "string",
				},
			},
		},
		"required": []string{"path"},
	}
}

// Validate checks if the parameters are valid
func (t *ListTool) Validate(params json.RawMessage) error {
	var p ListParams
	if err := json.Unmarshal(params, &p); err != nil {
		return fmt.Errorf("invalid parameters: %w", err)
	}

	if p.Path == "" {
		return fmt.Errorf("path is required")
	}

	return nil
}

// Execute runs the tool with the given parameters
func (t *ListTool) Execute(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var p ListParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}

	// Check if path exists
	info, err := os.Stat(p.Path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("path does not exist: %s", p.Path)
		}
		return nil, fmt.Errorf("cannot access path: %w", err)
	}

	if !info.IsDir() {
		return nil, fmt.Errorf("path is not a directory: %s", p.Path)
	}

	var files []FileInfo

	if p.Recursive {
		files, err = t.listRecursive(p.Path, p.Pattern, p.Include, p.Exclude)
	} else {
		files, err = t.listDirectory(p.Path, p.Pattern, p.Include, p.Exclude)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to list directory: %w", err)
	}

	// Sort by name
	sort.Slice(files, func(i, j int) bool {
		return files[i].Name < files[j].Name
	})

	return &ListResult{
		Path:  p.Path,
		Files: files,
		Count: len(files),
	}, nil
}

// listDirectory lists files in a single directory
func (t *ListTool) listDirectory(path, pattern string, include, exclude []string) ([]FileInfo, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	var files []FileInfo

	for _, entry := range entries {
		fullPath := filepath.Join(path, entry.Name())

		// Skip if doesn't match pattern
		if pattern != "" {
			matched, err := filepath.Match(pattern, entry.Name())
			if err != nil {
				return nil, fmt.Errorf("invalid pattern: %w", err)
			}
			if !matched {
				continue
			}
		}

		// Skip if not in include list
		if len(include) > 0 && !entry.IsDir() {
			if !matchesExtensions(entry.Name(), include) {
				continue
			}
		}

		// Skip if in exclude list
		if len(exclude) > 0 && !entry.IsDir() {
			if matchesExtensions(entry.Name(), exclude) {
				continue
			}
		}

		info, err := entry.Info()
		if err != nil {
			continue // Skip files we can't stat
		}

		fileInfo := FileInfo{
			Path:         fullPath,
			Name:         entry.Name(),
			Size:         info.Size(),
			IsDir:        entry.IsDir(),
			Permissions:  info.Mode().String(),
			ModifiedTime: info.ModTime().Unix(),
		}

		files = append(files, fileInfo)
	}

	return files, nil
}

// listRecursive lists files recursively
func (t *ListTool) listRecursive(path, pattern string, include, exclude []string) ([]FileInfo, error) {
	var files []FileInfo

	err := filepath.Walk(path, func(walkPath string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip files with errors
		}

		// Skip the root directory itself
		if walkPath == path {
			return nil
		}

		name := filepath.Base(walkPath)

		// Skip if doesn't match pattern
		if pattern != "" {
			matched, err := filepath.Match(pattern, name)
			if err != nil {
				return fmt.Errorf("invalid pattern: %w", err)
			}
			if !matched {
				return nil
			}
		}

		// Skip if not in include list
		if len(include) > 0 && !info.IsDir() {
			if !matchesExtensions(name, include) {
				return nil
			}
		}

		// Skip if in exclude list
		if len(exclude) > 0 && !info.IsDir() {
			if matchesExtensions(name, exclude) {
				return nil
			}
		}

		fileInfo := FileInfo{
			Path:         walkPath,
			Name:         name,
			Size:         info.Size(),
			IsDir:        info.IsDir(),
			Permissions:  info.Mode().String(),
			ModifiedTime: info.ModTime().Unix(),
		}

		files = append(files, fileInfo)
		return nil
	})

	return files, err
}

// matchesExtensions checks if a filename matches any of the given extensions
func matchesExtensions(filename string, extensions []string) bool {
	lower := strings.ToLower(filename)
	for _, ext := range extensions {
		if strings.HasSuffix(lower, strings.ToLower(ext)) {
			return true
		}
	}
	return false
}
