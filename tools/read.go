package tools

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// ReadTool implements file reading functionality
type ReadTool struct{}

// ReadParams defines the parameters for the Read tool
type ReadParams struct {
	Path   string `json:"path"`
	Format string `json:"format,omitempty"` // "text", "binary", "auto"
	Offset int    `json:"offset,omitempty"` // Line offset for partial reads
	Limit  int    `json:"limit,omitempty"`  // Number of lines to read
}

// ReadResult represents the result of a read operation
type ReadResult struct {
	Content    string `json:"content"`
	Format     string `json:"format"`
	Lines      int    `json:"lines,omitempty"`
	Size       int64  `json:"size"`
	TotalLines int    `json:"total_lines,omitempty"`
}

// NewReadTool creates a new Read tool instance
func NewReadTool() *ReadTool {
	return &ReadTool{}
}

// Name returns the tool's name
func (t *ReadTool) Name() string {
	return "read"
}

// Description returns the tool's description
func (t *ReadTool) Description() string {
	return "Read files from the local filesystem with support for text and binary formats, line range selection, and encoding detection"
}

// Schema returns the JSON schema for the tool's parameters
func (t *ReadTool) Schema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"path": map[string]interface{}{
				"type":        "string",
				"description": "Path to the file to read",
			},
			"format": map[string]interface{}{
				"type":        "string",
				"description": "Format to read the file in: 'text', 'binary', or 'auto' (default)",
				"enum":        []string{"text", "binary", "auto"},
			},
			"offset": map[string]interface{}{
				"type":        "integer",
				"description": "Line offset for partial reads (text mode only)",
				"minimum":     0,
			},
			"limit": map[string]interface{}{
				"type":        "integer",
				"description": "Number of lines to read (text mode only, 0 = all)",
				"minimum":     0,
			},
		},
		"required": []string{"path"},
	}
}

// Validate checks if the parameters are valid
func (t *ReadTool) Validate(params json.RawMessage) error {
	var p ReadParams
	if err := json.Unmarshal(params, &p); err != nil {
		return fmt.Errorf("invalid parameters: %w", err)
	}

	if p.Path == "" {
		return fmt.Errorf("path is required")
	}

	if p.Format != "" && p.Format != "text" && p.Format != "binary" && p.Format != "auto" {
		return fmt.Errorf("format must be 'text', 'binary', or 'auto'")
	}

	if p.Offset < 0 {
		return fmt.Errorf("offset must be non-negative")
	}

	if p.Limit < 0 {
		return fmt.Errorf("limit must be non-negative")
	}

	return nil
}

// Execute runs the tool with the given parameters
func (t *ReadTool) Execute(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var p ReadParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}

	// Default format to auto
	if p.Format == "" {
		p.Format = "auto"
	}

	// Check if file exists
	info, err := os.Stat(p.Path)
	if err != nil {
		return nil, fmt.Errorf("cannot access file: %w", err)
	}

	if info.IsDir() {
		return nil, fmt.Errorf("path is a directory, not a file")
	}

	// Determine format
	format := p.Format
	if format == "auto" {
		format = detectFormat(p.Path)
	}

	// Read file based on format
	if format == "binary" {
		return t.readBinary(p.Path, info.Size())
	}

	return t.readText(p.Path, p.Offset, p.Limit, info.Size())
}

// detectFormat determines if a file should be read as text or binary
func detectFormat(path string) string {
	// Check extension
	ext := strings.ToLower(path)
	textExtensions := []string{".txt", ".md", ".json", ".yaml", ".yml", ".csv", ".xml", ".html", ".js", ".go", ".py", ".rb", ".java", ".c", ".cpp", ".h", ".sh", ".bash"}

	for _, textExt := range textExtensions {
		if strings.HasSuffix(ext, textExt) {
			return "text"
		}
	}

	// Read first few bytes to check for binary content
	data, err := os.ReadFile(path)
	if err != nil {
		return "text" // Default to text if we can't read
	}

	// Check for null bytes (common in binary files)
	for i := 0; i < len(data) && i < 512; i++ {
		if data[i] == 0 {
			return "binary"
		}
	}

	return "text"
}

// readText reads a file as text with optional line range
func (t *ReadTool) readText(path string, offset, limit int, size int64) (*ReadResult, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	content := string(data)
	lines := strings.Split(content, "\n")
	totalLines := len(lines)

	// Apply offset and limit if specified
	if offset > 0 || limit > 0 {
		start := offset
		end := len(lines)

		if start >= len(lines) {
			start = len(lines)
		}

		if limit > 0 {
			end = start + limit
			if end > len(lines) {
				end = len(lines)
			}
		}

		lines = lines[start:end]
		content = strings.Join(lines, "\n")
	}

	return &ReadResult{
		Content:    content,
		Format:     "text",
		Lines:      len(lines),
		Size:       size,
		TotalLines: totalLines,
	}, nil
}

// readBinary reads a file as binary and encodes it in base64
func (t *ReadTool) readBinary(path string, size int64) (*ReadResult, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	encoded := base64.StdEncoding.EncodeToString(data)

	return &ReadResult{
		Content: encoded,
		Format:  "binary",
		Size:    size,
	}, nil
}
