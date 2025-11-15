package tools

import (
	"context"
	"encoding/json"
)

// Tool interface that all tools must implement
type Tool interface {
	// Name returns the tool's name
	Name() string

	// Description returns a description of what the tool does
	Description() string

	// Execute runs the tool with the given parameters
	Execute(ctx context.Context, params json.RawMessage) (interface{}, error)

	// Validate checks if the parameters are valid
	Validate(params json.RawMessage) error

	// Schema returns the JSON schema for the tool's parameters
	Schema() map[string]interface{}
}

// ToolResult represents the result of a tool execution
type ToolResult struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// FileInfo represents metadata about a file
type FileInfo struct {
	Path         string `json:"path"`
	Name         string `json:"name"`
	Size         int64  `json:"size"`
	IsDir        bool   `json:"is_dir"`
	Permissions  string `json:"permissions"`
	ModifiedTime int64  `json:"modified_time"`
}
