package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// DeleteTool implements file and directory deletion functionality
type DeleteTool struct{}

// DeleteParams defines the parameters for the Delete tool
type DeleteParams struct {
	Path      string `json:"path"`
	Recursive bool   `json:"recursive,omitempty"`
	Confirm   bool   `json:"confirm,omitempty"`
	DryRun    bool   `json:"dry_run,omitempty"`
}

// DeleteResult represents the result of a delete operation
type DeleteResult struct {
	Path      string   `json:"path"`
	Deleted   bool     `json:"deleted"`
	DryRun    bool     `json:"dry_run,omitempty"`
	FilesRemoved int   `json:"files_removed,omitempty"`
	Message   string   `json:"message,omitempty"`
	Items     []string `json:"items,omitempty"` // List of items that would be deleted (dry-run)
}

// NewDeleteTool creates a new Delete tool instance
func NewDeleteTool() *DeleteTool {
	return &DeleteTool{}
}

// Name returns the tool's name
func (t *DeleteTool) Name() string {
	return "delete"
}

// Description returns the tool's description
func (t *DeleteTool) Description() string {
	return "Delete files and directories with safety features including recursive deletion, dry-run mode, and confirmation prompts"
}

// Schema returns the JSON schema for the tool's parameters
func (t *DeleteTool) Schema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"path": map[string]interface{}{
				"type":        "string",
				"description": "Path to the file or directory to delete",
			},
			"recursive": map[string]interface{}{
				"type":        "boolean",
				"description": "Enable recursive deletion for directories",
			},
			"confirm": map[string]interface{}{
				"type":        "boolean",
				"description": "Require confirmation before deletion (currently always true for safety)",
			},
			"dry_run": map[string]interface{}{
				"type":        "boolean",
				"description": "Simulate deletion without actually removing files",
			},
		},
		"required": []string{"path"},
	}
}

// Validate checks if the parameters are valid
func (t *DeleteTool) Validate(params json.RawMessage) error {
	var p DeleteParams
	if err := json.Unmarshal(params, &p); err != nil {
		return fmt.Errorf("invalid parameters: %w", err)
	}

	if p.Path == "" {
		return fmt.Errorf("path is required")
	}

	return nil
}

// Execute runs the tool with the given parameters
func (t *DeleteTool) Execute(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var p DeleteParams
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

	// If it's a directory and recursive is not set, return error
	if info.IsDir() && !p.Recursive {
		return nil, fmt.Errorf("path is a directory, use recursive=true to delete directories")
	}

	// Dry run mode - just list what would be deleted
	if p.DryRun {
		items, count, err := t.listDeletionItems(p.Path, info.IsDir())
		if err != nil {
			return nil, fmt.Errorf("failed to list items: %w", err)
		}

		return &DeleteResult{
			Path:         p.Path,
			Deleted:      false,
			DryRun:       true,
			FilesRemoved: count,
			Items:        items,
			Message:      fmt.Sprintf("Would delete %d item(s)", count),
		}, nil
	}

	// Perform actual deletion
	var filesRemoved int
	if info.IsDir() {
		count, err := countFiles(p.Path)
		if err != nil {
			return nil, fmt.Errorf("failed to count files: %w", err)
		}
		filesRemoved = count

		if err := os.RemoveAll(p.Path); err != nil {
			return nil, fmt.Errorf("failed to delete directory: %w", err)
		}
	} else {
		filesRemoved = 1
		if err := os.Remove(p.Path); err != nil {
			return nil, fmt.Errorf("failed to delete file: %w", err)
		}
	}

	return &DeleteResult{
		Path:         p.Path,
		Deleted:      true,
		FilesRemoved: filesRemoved,
		Message:      fmt.Sprintf("Successfully deleted %d item(s)", filesRemoved),
	}, nil
}

// listDeletionItems returns a list of items that would be deleted
func (t *DeleteTool) listDeletionItems(path string, isDir bool) ([]string, int, error) {
	items := []string{}
	count := 0

	if !isDir {
		items = append(items, path)
		return items, 1, nil
	}

	err := filepath.Walk(path, func(walkPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if walkPath != path { // Don't include the root directory itself in the list
			items = append(items, walkPath)
			if !info.IsDir() {
				count++
			}
		}
		return nil
	})

	if err != nil {
		return nil, 0, err
	}

	// Add the directory itself
	items = append(items, path)
	count++ // Count the directory

	return items, count, nil
}

// countFiles counts the number of files in a directory recursively
func countFiles(path string) (int, error) {
	count := 0
	err := filepath.Walk(path, func(walkPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		count++
		return nil
	})
	return count, err
}
