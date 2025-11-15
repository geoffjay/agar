package tools

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// WriteTool implements file writing functionality
type WriteTool struct{}

// WriteParams defines the parameters for the Write tool
type WriteParams struct {
	Path     string `json:"path"`
	Content  string `json:"content"`
	Mode     string `json:"mode,omitempty"`     // "write", "append"
	Encoding string `json:"encoding,omitempty"` // "utf-8", "base64"
	Backup   bool   `json:"backup,omitempty"`
}

// WriteResult represents the result of a write operation
type WriteResult struct {
	Path       string `json:"path"`
	BytesWritten int64  `json:"bytes_written"`
	BackupPath string `json:"backup_path,omitempty"`
}

// NewWriteTool creates a new Write tool instance
func NewWriteTool() *WriteTool {
	return &WriteTool{}
}

// Name returns the tool's name
func (t *WriteTool) Name() string {
	return "write"
}

// Description returns the tool's description
func (t *WriteTool) Description() string {
	return "Write content to files with support for text and binary formats, append mode, automatic directory creation, and backup functionality"
}

// Schema returns the JSON schema for the tool's parameters
func (t *WriteTool) Schema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"path": map[string]interface{}{
				"type":        "string",
				"description": "Path to the file to write",
			},
			"content": map[string]interface{}{
				"type":        "string",
				"description": "Content to write to the file",
			},
			"mode": map[string]interface{}{
				"type":        "string",
				"description": "Write mode: 'write' (default) or 'append'",
				"enum":        []string{"write", "append"},
			},
			"encoding": map[string]interface{}{
				"type":        "string",
				"description": "Content encoding: 'utf-8' (default) or 'base64'",
				"enum":        []string{"utf-8", "base64"},
			},
			"backup": map[string]interface{}{
				"type":        "boolean",
				"description": "Create a backup of existing file before overwriting",
			},
		},
		"required": []string{"path", "content"},
	}
}

// Validate checks if the parameters are valid
func (t *WriteTool) Validate(params json.RawMessage) error {
	var p WriteParams
	if err := json.Unmarshal(params, &p); err != nil {
		return fmt.Errorf("invalid parameters: %w", err)
	}

	if p.Path == "" {
		return fmt.Errorf("path is required")
	}

	// Note: Content can be empty string, which is valid (creates empty file)

	if p.Mode != "" && p.Mode != "write" && p.Mode != "append" {
		return fmt.Errorf("mode must be 'write' or 'append'")
	}

	if p.Encoding != "" && p.Encoding != "utf-8" && p.Encoding != "base64" {
		return fmt.Errorf("encoding must be 'utf-8' or 'base64'")
	}

	return nil
}

// Execute runs the tool with the given parameters
func (t *WriteTool) Execute(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var p WriteParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}

	// Default values
	if p.Mode == "" {
		p.Mode = "write"
	}
	if p.Encoding == "" {
		p.Encoding = "utf-8"
	}

	// Ensure directory exists
	dir := filepath.Dir(p.Path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	// Backup existing file if requested
	var backupPath string
	if p.Backup && p.Mode == "write" {
		if _, err := os.Stat(p.Path); err == nil {
			backupPath = p.Path + ".backup"
			if err := copyFile(p.Path, backupPath); err != nil {
				return nil, fmt.Errorf("failed to create backup: %w", err)
			}
		}
	}

	// Decode content if needed
	var data []byte
	var err error
	if p.Encoding == "base64" {
		data, err = base64.StdEncoding.DecodeString(p.Content)
		if err != nil {
			return nil, fmt.Errorf("failed to decode base64 content: %w", err)
		}
	} else {
		data = []byte(p.Content)
	}

	// Write or append to file
	var bytesWritten int64
	if p.Mode == "append" {
		file, err := os.OpenFile(p.Path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return nil, fmt.Errorf("failed to open file for append: %w", err)
		}
		defer file.Close()

		n, err := file.Write(data)
		if err != nil {
			return nil, fmt.Errorf("failed to append to file: %w", err)
		}
		bytesWritten = int64(n)
	} else {
		// Use a temporary file for atomic writes
		tmpFile := p.Path + ".tmp"
		if err := os.WriteFile(tmpFile, data, 0644); err != nil {
			return nil, fmt.Errorf("failed to write to temporary file: %w", err)
		}

		// Rename temp file to target (atomic operation on most systems)
		if err := os.Rename(tmpFile, p.Path); err != nil {
			os.Remove(tmpFile) // Clean up temp file on error
			return nil, fmt.Errorf("failed to rename temporary file: %w", err)
		}

		bytesWritten = int64(len(data))
	}

	result := &WriteResult{
		Path:         p.Path,
		BytesWritten: bytesWritten,
		BackupPath:   backupPath,
	}

	return result, nil
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}
