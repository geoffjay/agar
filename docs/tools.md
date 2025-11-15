# AI Agent Tools Framework

## Overview

The Agar library provides a comprehensive tools framework for AI agents with essential capabilities for file operations, system interaction, and task management. All tools implement a consistent interface and include built-in security features.

## Core Interface

All tools implement the `Tool` interface:

```go
type Tool interface {
    Name() string
    Description() string
    Execute(ctx context.Context, params json.RawMessage) (interface{}, error)
    Validate(params json.RawMessage) error
    Schema() map[string]interface{}
}
```

## Tool Registry

The registry provides thread-safe management of tools:

```go
// Create and register tools
registry := tools.NewToolRegistry()
registry.Register(tools.NewReadTool())
registry.Register(tools.NewWriteTool())
registry.Register(tools.NewDeleteTool())
registry.Register(tools.NewListTool())
registry.Register(tools.NewGlobTool())
registry.Register(tools.NewShellTool())
registry.Register(tools.NewTaskListTool())

// Execute a tool
tool, err := registry.Get("read")
if err != nil {
    log.Fatal(err)
}

params := json.RawMessage(`{"path": "file.txt"}`)
result, err := tool.Execute(context.Background(), params)
```

## Available Tools

### File System Tools

#### Read Tool

Read files from the local filesystem with automatic format detection.

**Features**:
- Automatic format detection (text/binary based on file extension and content)
- Line range selection with offset and limit
- Base64 encoding for binary files
- Support for common text formats: txt, json, yaml, csv, md, html, js, go, py, rb, java, c, cpp, sh, etc.

**Parameters**:
```json
{
  "path": "string (required) - Path to the file to read",
  "format": "string (optional) - 'text', 'binary', or 'auto' (default: auto)",
  "offset": "integer (optional) - Line offset for partial reads (default: 0)",
  "limit": "integer (optional) - Number of lines to read, 0 = all (default: 0)"
}
```

**Usage Example**:
```go
tool := tools.NewReadTool()

// Read entire file
params := json.RawMessage(`{"path": "/path/to/file.txt"}`)
result, err := tool.Execute(ctx, params)

readResult := result.(*tools.ReadResult)
fmt.Println(readResult.Content)
fmt.Printf("Read %d of %d total lines\n", readResult.Lines, readResult.TotalLines)

// Read specific line range
params = json.RawMessage(`{
    "path": "/path/to/large-file.log",
    "offset": 100,
    "limit": 50
}`)
result, err = tool.Execute(ctx, params)

// Read binary file
params = json.RawMessage(`{
    "path": "/path/to/image.png",
    "format": "binary"
}`)
result, err = tool.Execute(ctx, params)
// Content will be base64 encoded
```

---

#### Write Tool

Write content to files with safety features including atomic writes and backups.

**Features**:
- Text and binary file writing
- Append mode for adding to existing files
- Automatic directory creation
- Optional backup of existing files before overwriting
- Atomic writes using temporary files
- Base64 decoding for binary content

**Parameters**:
```json
{
  "path": "string (required) - Path to the file to write",
  "content": "string (required) - Content to write",
  "mode": "string (optional) - 'write' or 'append' (default: write)",
  "encoding": "string (optional) - 'utf-8' or 'base64' (default: utf-8)",
  "backup": "boolean (optional) - Create backup before overwriting (default: false)"
}
```

**Usage Example**:
```go
tool := tools.NewWriteTool()

// Write text file
params := json.RawMessage(`{
    "path": "/path/to/file.txt",
    "content": "Hello, World!"
}`)
result, err := tool.Execute(ctx, params)

writeResult := result.(*tools.WriteResult)
fmt.Printf("Wrote %d bytes to %s\n", writeResult.BytesWritten, writeResult.Path)

// Write with backup
params = json.RawMessage(`{
    "path": "/path/to/config.json",
    "content": "{\"key\": \"value\"}",
    "backup": true
}`)
result, err = tool.Execute(ctx, params)
if writeResult.BackupPath != "" {
    fmt.Printf("Backup created at: %s\n", writeResult.BackupPath)
}

// Append to existing file
params = json.RawMessage(`{
    "path": "/path/to/log.txt",
    "content": "New log entry\n",
    "mode": "append"
}`)

// Write binary file (base64 encoded content)
params = json.RawMessage(`{
    "path": "/path/to/file.bin",
    "content": "AQIDBA==",
    "encoding": "base64"
}`)
```

---

#### Delete Tool

Delete files and directories with safety features including dry-run mode.

**Features**:
- File and directory deletion
- Recursive directory deletion
- Dry-run mode to preview what would be deleted
- Reports number of files/directories affected

**Parameters**:
```json
{
  "path": "string (required) - Path to delete",
  "recursive": "boolean (optional) - Enable recursive deletion for directories (default: false)",
  "confirm": "boolean (optional) - Reserved for future use",
  "dry_run": "boolean (optional) - Preview deletion without removing files (default: false)"
}
```

**Usage Example**:
```go
tool := tools.NewDeleteTool()

// Preview what would be deleted
params := json.RawMessage(`{
    "path": "/path/to/directory",
    "recursive": true,
    "dry_run": true
}`)
result, err := tool.Execute(ctx, params)

deleteResult := result.(*tools.DeleteResult)
fmt.Printf("Would delete %d items:\n", deleteResult.FilesRemoved)
for _, item := range deleteResult.Items {
    fmt.Printf("  - %s\n", item)
}

// Perform actual deletion
params = json.RawMessage(`{
    "path": "/path/to/directory",
    "recursive": true
}`)
result, err = tool.Execute(ctx, params)

// Delete single file
params = json.RawMessage(`{"path": "/path/to/file.txt"}`)
result, err = tool.Execute(ctx, params)
```

---

#### List Tool

List directory contents with filtering and metadata.

**Features**:
- Pattern-based filtering using glob syntax
- Recursive directory listing
- File metadata including size, permissions, and modification time
- Include/exclude filtering by file extensions
- Results sorted by name

**Parameters**:
```json
{
  "path": "string (required) - Directory path to list",
  "pattern": "string (optional) - Glob pattern to filter files (e.g., '*.txt')",
  "recursive": "boolean (optional) - List directories recursively (default: false)",
  "include": "array (optional) - File extensions to include (e.g., ['.txt', '.md'])",
  "exclude": "array (optional) - File extensions to exclude (e.g., ['.tmp', '.log'])"
}
```

**Usage Example**:
```go
tool := tools.NewListTool()

// List all files in directory
params := json.RawMessage(`{"path": "/path/to/directory"}`)
result, err := tool.Execute(ctx, params)

listResult := result.(*tools.ListResult)
fmt.Printf("Found %d files in %s\n", listResult.Count, listResult.Path)
for _, file := range listResult.Files {
    fmt.Printf("%s (%d bytes, %s, modified: %d)\n",
        file.Name, file.Size, file.Permissions, file.ModifiedTime)
}

// List with pattern filter
params = json.RawMessage(`{
    "path": "/path/to/directory",
    "pattern": "*.go"
}`)

// Recursive listing with extension filters
params = json.RawMessage(`{
    "path": "/path/to/project",
    "recursive": true,
    "include": [".go", ".mod", ".sum"],
    "exclude": [".tmp", ".bak"]
}`)
```

---

#### Glob Tool

Advanced file pattern matching with support for recursive patterns.

**Features**:
- Recursive pattern matching with `**` (e.g., `**/*.go`)
- Multiple pattern support
- Sort by name, size, or modification time
- Ascending or descending sort order
- Optional detailed file metadata
- Configurable symlink following

**Parameters**:
```json
{
  "patterns": "array (required) - Glob patterns to match (e.g., ['**/*.go', '*.txt'])",
  "path": "string (optional) - Base path to search from (default: current directory)",
  "case_sensitive": "boolean (optional) - Case-sensitive matching (default: false)",
  "follow_symlinks": "boolean (optional) - Follow symbolic links (default: false)",
  "sort_by": "string (optional) - Sort by 'name', 'size', or 'modtime'",
  "sort_order": "string (optional) - 'asc' or 'desc' (default: asc)",
  "include_info": "boolean (optional) - Include detailed metadata (default: false)"
}
```

**Usage Example**:
```go
tool := tools.NewGlobTool()

// Find all Go files recursively
params := json.RawMessage(`{
    "patterns": ["**/*.go"],
    "path": "/path/to/project"
}`)
result, err := tool.Execute(ctx, params)

globResult := result.(*tools.GlobResult)
fmt.Printf("Found %d matches\n", globResult.Count)
for _, match := range globResult.Matches {
    fmt.Println(match.Path)
}

// Multiple patterns with sorting and metadata
params = json.RawMessage(`{
    "patterns": ["**/*.go", "**/*_test.go"],
    "path": "/path/to/project",
    "sort_by": "modtime",
    "sort_order": "desc",
    "include_info": true
}`)
result, err = tool.Execute(ctx, params)

globResult = result.(*tools.GlobResult)
for _, match := range globResult.Matches {
    fmt.Printf("%s (%d bytes, modified: %d)\n",
        match.Path, match.Size, match.ModifiedTime)
}

// Case-sensitive matching
params = json.RawMessage(`{
    "patterns": ["**/*.TXT"],
    "case_sensitive": true
}`)
```

---

### System Tools

#### Shell Tool

Execute shell commands with security measures and timeout protection.

**Features**:
- Command execution with configurable timeout
- Working directory specification
- Environment variable support
- Shell selection (bash, sh, powershell)
- Dangerous command blocking
- Exit code and output capture
- Separate stdout and stderr
- Execution duration tracking

**Parameters**:
```json
{
  "command": "string (required) - Command to execute",
  "args": "array (optional) - Command arguments",
  "working_dir": "string (optional) - Working directory for execution",
  "environment": "object (optional) - Environment variables as key-value pairs",
  "timeout": "integer (optional) - Timeout in seconds (default: 30, max: 300)",
  "shell": "string (optional) - Shell to use: 'bash', 'sh', or 'powershell'"
}
```

**Usage Example**:
```go
tool := tools.NewShellTool()

// Simple command execution
params := json.RawMessage(`{"command": "ls -la"}`)
result, err := tool.Execute(ctx, params)

shellResult := result.(*tools.ShellResult)
fmt.Printf("Exit Code: %d\n", shellResult.ExitCode)
fmt.Printf("Output:\n%s\n", shellResult.Stdout)
if shellResult.Stderr != "" {
    fmt.Printf("Errors:\n%s\n", shellResult.Stderr)
}

// Command with arguments and timeout
params = json.RawMessage(`{
    "command": "find",
    "args": ["/path/to/search", "-name", "*.go"],
    "timeout": 60
}`)

// With working directory and environment
params = json.RawMessage(`{
    "command": "make build",
    "working_dir": "/path/to/project",
    "environment": {
        "GOARCH": "amd64",
        "GOOS": "linux"
    },
    "timeout": 120
}`)

// Using specific shell for complex commands
params = json.RawMessage(`{
    "command": "for f in *.txt; do echo $f; done",
    "shell": "bash"
}`)

// Check for timeout
if shellResult.Timeout {
    fmt.Println("Command timed out")
}
```

**Security Notes**:
- Commands are validated before execution
- Dangerous patterns are blocked: `rm -rf /`, fork bombs, `mkfs`, `dd if=/dev/zero`
- Maximum timeout is enforced at 300 seconds
- Processes are properly cleaned up

---

### Task Management Tools

#### TaskList Tool

Create and manage hierarchical task lists with priority and status tracking.

**Features**:
- Task list creation and management
- Hierarchical tasks with parent/child relationships
- Priority levels: high, medium, low
- Status tracking: pending, in_progress, completed, cancelled
- CRUD operations: create, update, list, delete, get
- Automatic timestamps for created, updated, and completed
- UUID-based IDs for lists and tasks

**Parameters**:
```json
{
  "action": "string (required) - 'create', 'update', 'list', 'delete', or 'get'",
  "list_id": "string (optional) - Task list ID",
  "task_id": "string (optional) - Task ID",
  "title": "string (optional) - Title for task or list",
  "description": "string (optional) - Task description",
  "priority": "string (optional) - 'high', 'medium', or 'low' (default: medium)",
  "status": "string (optional) - 'pending', 'in_progress', 'completed', or 'cancelled'",
  "parent_id": "string (optional) - Parent task ID for hierarchical tasks"
}
```

**Usage Examples**:

**Create a task list**:
```go
tool := tools.NewTaskListTool()
params := json.RawMessage(`{
    "action": "create",
    "title": "Project Tasks"
}`)
result, err := tool.Execute(ctx, params)

taskResult := result.(*tools.TaskListResult)
listID := taskResult.ListID
fmt.Printf("Created list: %s\n", listID)
```

**Create tasks**:
```go
// High priority task
params := json.RawMessage(fmt.Sprintf(`{
    "action": "create",
    "list_id": "%s",
    "title": "Implement authentication",
    "description": "Add OAuth2 support",
    "priority": "high"
}`, listID))
result, err := tool.Execute(ctx, params)

taskResult := result.(*tools.TaskListResult)
taskID := taskResult.TaskID

// Task with default priority (medium)
params = json.RawMessage(fmt.Sprintf(`{
    "action": "create",
    "list_id": "%s",
    "title": "Write documentation"
}`, listID))
```

**Update task status**:
```go
params := json.RawMessage(fmt.Sprintf(`{
    "action": "update",
    "list_id": "%s",
    "task_id": "%s",
    "status": "in_progress"
}`, listID, taskID))
result, err := tool.Execute(ctx, params)

// Mark as completed
params = json.RawMessage(fmt.Sprintf(`{
    "action": "update",
    "list_id": "%s",
    "task_id": "%s",
    "status": "completed"
}`, listID, taskID))
```

**List tasks**:
```go
// List all task lists
params := json.RawMessage(`{"action": "list"}`)
result, err := tool.Execute(ctx, params)

taskResult := result.(*tools.TaskListResult)
for _, list := range taskResult.Lists {
    fmt.Printf("%s: %s (%d tasks)\n", list.ID, list.Title, list.TaskCount)
}

// List tasks in a specific list
params = json.RawMessage(fmt.Sprintf(`{
    "action": "list",
    "list_id": "%s"
}`, listID))
result, err = tool.Execute(ctx, params)

taskResult = result.(*tools.TaskListResult)
for _, task := range taskResult.Tasks {
    fmt.Printf("[%s] %s - %s (%s)\n",
        task.Priority, task.Title, task.Status, task.Description)
}
```

**Get specific task**:
```go
params := json.RawMessage(fmt.Sprintf(`{
    "action": "get",
    "list_id": "%s",
    "task_id": "%s"
}`, listID, taskID))
result, err := tool.Execute(ctx, params)

taskResult := result.(*tools.TaskListResult)
task := taskResult.Task
fmt.Printf("Task: %s\nStatus: %s\nPriority: %s\n",
    task.Title, task.Status, task.Priority)
```

**Hierarchical tasks**:
```go
// Create parent task
params := json.RawMessage(fmt.Sprintf(`{
    "action": "create",
    "list_id": "%s",
    "title": "Feature: User Management"
}`, listID))
parentResult, _ := tool.Execute(ctx, params)
parentID := parentResult.(*tools.TaskListResult).TaskID

// Create subtasks
params = json.RawMessage(fmt.Sprintf(`{
    "action": "create",
    "list_id": "%s",
    "title": "Design database schema",
    "parent_id": "%s",
    "priority": "high"
}`, listID, parentID))

params = json.RawMessage(fmt.Sprintf(`{
    "action": "create",
    "list_id": "%s",
    "title": "Implement API endpoints",
    "parent_id": "%s"
}`, listID, parentID))
```

**Delete tasks and lists**:
```go
// Delete a task
params := json.RawMessage(fmt.Sprintf(`{
    "action": "delete",
    "list_id": "%s",
    "task_id": "%s"
}`, listID, taskID))

// Delete entire list
params = json.RawMessage(fmt.Sprintf(`{
    "action": "delete",
    "list_id": "%s"
}`, listID))
```

---

## Security Features

### Input Validation
All tools perform strict parameter validation before execution:
- Type checking and format validation
- Range validation for numeric parameters
- Path sanitization to prevent directory traversal attacks
- Required parameter enforcement

### Shell Command Safety
The Shell tool includes multiple security layers:
- **Dangerous pattern blocking**: Prevents execution of commands like `rm -rf /`, fork bombs, `mkfs`, and `dd if=/dev/zero`
- **Timeout enforcement**: Maximum 300-second timeout to prevent runaway processes
- **Process cleanup**: Ensures proper cleanup of child processes
- **No privilege escalation**: Runs commands with current user permissions

### File Operations Safety
File system tools implement safety features:
- **Backup support**: Write tool can backup files before overwriting
- **Dry-run mode**: Delete tool supports preview mode
- **Atomic writes**: Write tool uses temporary files and atomic rename operations
- **Directory creation**: Automatic parent directory creation with appropriate permissions

### Error Handling
- Comprehensive error messages with context
- Graceful failure modes with proper cleanup
- Resource cleanup on error conditions
- Context-aware error reporting

## Best Practices

### Parameter Validation

Always validate parameters before execution:

```go
tool := tools.NewReadTool()

// Validate before executing
if err := tool.Validate(params); err != nil {
    log.Printf("Invalid parameters: %v", err)
    return
}

result, err := tool.Execute(ctx, params)
if err != nil {
    log.Printf("Execution failed: %v", err)
    return
}
```

### Use Dry-Run for Destructive Operations

Test destructive operations before executing:

```go
deleteTool := tools.NewDeleteTool()

// First, preview what would be deleted
dryRunParams := json.RawMessage(`{
    "path": "/path/to/delete",
    "recursive": true,
    "dry_run": true
}`)
result, _ := deleteTool.Execute(ctx, dryRunParams)

deleteResult := result.(*tools.DeleteResult)
fmt.Printf("Will delete %d items\n", deleteResult.FilesRemoved)

// After user confirmation, perform actual deletion
actualParams := json.RawMessage(`{
    "path": "/path/to/delete",
    "recursive": true
}`)
result, err := deleteTool.Execute(ctx, actualParams)
```

### Set Appropriate Timeouts

Configure timeouts based on expected execution time:

```go
shellTool := tools.NewShellTool()

// Short timeout for quick commands
quickParams := json.RawMessage(`{
    "command": "ls -la",
    "timeout": 5
}`)

// Longer timeout for builds
buildParams := json.RawMessage(`{
    "command": "make build",
    "timeout": 180
}`)
```

### Use Context Cancellation

Support cancellation for long-running operations:

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

result, err := tool.Execute(ctx, params)
if err != nil {
    if ctx.Err() == context.DeadlineExceeded {
        log.Println("Operation timed out")
    } else {
        log.Printf("Operation failed: %v", err)
    }
}
```

### Registry Pattern

Initialize the registry once and reuse:

```go
// Application initialization
var toolRegistry *tools.ToolRegistry

func init() {
    toolRegistry = tools.NewToolRegistry()
    toolRegistry.Register(tools.NewReadTool())
    toolRegistry.Register(tools.NewWriteTool())
    toolRegistry.Register(tools.NewDeleteTool())
    toolRegistry.Register(tools.NewListTool())
    toolRegistry.Register(tools.NewGlobTool())
    toolRegistry.Register(tools.NewShellTool())
    toolRegistry.Register(tools.NewTaskListTool())
}

// Use throughout application
func executeTool(name string, params json.RawMessage) (interface{}, error) {
    tool, err := toolRegistry.Get(name)
    if err != nil {
        return nil, fmt.Errorf("tool not found: %w", err)
    }

    if err := tool.Validate(params); err != nil {
        return nil, fmt.Errorf("validation failed: %w", err)
    }

    return tool.Execute(context.Background(), params)
}
```

### Error Handling Pattern

Implement comprehensive error handling:

```go
func safeToolExecution(tool tools.Tool, params json.RawMessage) (interface{}, error) {
    // Validate parameters
    if err := tool.Validate(params); err != nil {
        return nil, fmt.Errorf("parameter validation failed: %w", err)
    }

    // Create context with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
    defer cancel()

    // Execute tool
    result, err := tool.Execute(ctx, params)
    if err != nil {
        // Check for context errors
        if ctx.Err() == context.DeadlineExceeded {
            return nil, fmt.Errorf("operation timed out")
        }
        if ctx.Err() == context.Canceled {
            return nil, fmt.Errorf("operation canceled")
        }
        return nil, fmt.Errorf("execution failed: %w", err)
    }

    return result, nil
}
```

## Creating Custom Tools

Implement the `Tool` interface to create custom tools:

```go
package mytools

import (
    "context"
    "encoding/json"
    "fmt"

    "github.com/geoffjay/agar/tools"
)

// DatabaseQueryTool executes SQL queries
type DatabaseQueryTool struct {
    connectionString string
}

func NewDatabaseQueryTool(connStr string) *DatabaseQueryTool {
    return &DatabaseQueryTool{
        connectionString: connStr,
    }
}

func (t *DatabaseQueryTool) Name() string {
    return "database_query"
}

func (t *DatabaseQueryTool) Description() string {
    return "Execute SQL queries against a database and return results"
}

func (t *DatabaseQueryTool) Schema() map[string]interface{} {
    return map[string]interface{}{
        "type": "object",
        "properties": map[string]interface{}{
            "query": map[string]interface{}{
                "type":        "string",
                "description": "SQL query to execute",
            },
            "params": map[string]interface{}{
                "type":        "array",
                "description": "Query parameters for prepared statements",
                "items": map[string]interface{}{
                    "type": "string",
                },
            },
            "timeout": map[string]interface{}{
                "type":        "integer",
                "description": "Query timeout in seconds (default: 30)",
                "minimum":     1,
                "maximum":     300,
            },
        },
        "required": []string{"query"},
    }
}

func (t *DatabaseQueryTool) Validate(params json.RawMessage) error {
    var p struct {
        Query   string   `json:"query"`
        Params  []string `json:"params"`
        Timeout int      `json:"timeout"`
    }

    if err := json.Unmarshal(params, &p); err != nil {
        return fmt.Errorf("invalid parameters: %w", err)
    }

    if p.Query == "" {
        return fmt.Errorf("query is required")
    }

    if p.Timeout < 0 || p.Timeout > 300 {
        return fmt.Errorf("timeout must be between 1 and 300 seconds")
    }

    return nil
}

func (t *DatabaseQueryTool) Execute(ctx context.Context, params json.RawMessage) (interface{}, error) {
    var p struct {
        Query   string   `json:"query"`
        Params  []string `json:"params"`
        Timeout int      `json:"timeout"`
    }

    if err := json.Unmarshal(params, &p); err != nil {
        return nil, fmt.Errorf("invalid parameters: %w", err)
    }

    // Set default timeout
    if p.Timeout == 0 {
        p.Timeout = 30
    }

    // Your database query implementation here
    // ...

    return result, nil
}

// Register and use the custom tool
func main() {
    registry := tools.NewToolRegistry()

    // Register custom tool
    dbTool := NewDatabaseQueryTool("postgresql://localhost/mydb")
    registry.Register(dbTool)

    // Use the tool
    params := json.RawMessage(`{
        "query": "SELECT * FROM users WHERE age > $1",
        "params": ["25"],
        "timeout": 60
    }`)

    tool, _ := registry.Get("database_query")
    result, err := tool.Execute(context.Background(), params)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Query results: %+v\n", result)
}
```

## Dependencies

The tools package requires:

```go
require (
    github.com/google/uuid v1.6.0  // For TaskList UUID generation
)
```

The TUI framework (for UI integration) requires:

```go
require (
    github.com/charmbracelet/bubbletea v0.25.0
    github.com/charmbracelet/lipgloss v0.9.1
)
```
