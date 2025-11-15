# AI Agent Tools Framework

## Overview

This document outlines the comprehensive tools framework for the Agar library, providing AI agents with essential capabilities for file operations, system interaction, web access, task management, and extensibility for custom tools.

## Core Tools

### File System Tools

#### 1. Read Tool

**Purpose**: Read files from the local filesystem
**Features**:

- Read text files with encoding detection
- Read binary files with base64 encoding
- Support for various file formats (txt, json, yaml, csv, md, etc.)
- Line range selection for large files
- Error handling with detailed messages

**Implementation Pattern**:

```go
type ReadTool struct {
    Path   string `json:"path"`
    Format string `json:"format,omitempty"` // "text", "binary", "auto"
    Offset int    `json:"offset,omitempty"` // Line offset for partial reads
    Limit  int    `json:"limit,omitempty"`  // Number of lines to read
}
```

#### 2. Write Tool

**Purpose**: Write content to files with various options
**Features**:

- Text and binary file writing
- Append mode support
- Create directories if they don't exist
- Backup existing files before overwriting
- Atomic writes with temporary files

**Implementation Pattern**:

```go
type WriteTool struct {
    Path     string `json:"path"`
    Content  string `json:"content"`
    Mode     string `json:"mode,omitempty"`     // "write", "append"
    Encoding string `json:"encoding,omitempty"` // "utf-8", "base64"
    Backup   bool   `json:"backup,omitempty"`
}
```

#### 3. Delete Tool

**Purpose**: Delete files and directories
**Features**:

- Secure deletion with confirmation prompts
- Recursive directory deletion
- Dry-run mode for safety
- Trash/recycle bin support

**Implementation Pattern**:

```go
type DeleteTool struct {
    Path      string `json:"path"`
    Recursive bool   `json:"recursive,omitempty"`
    Confirm   bool   `json:"confirm,omitempty"`
    DryRun    bool   `json:"dry_run,omitempty"`
}
```

#### 4. List Tool

**Purpose**: List directory contents with filtering
**Features**:

- Pattern-based filtering (glob)
- Recursive listing
- File metadata (size, permissions, modification time)
- Human-readable vs machine-readable output

**Implementation Pattern**:

```go
type ListTool struct {
    Path      string   `json:"path"`
    Pattern   string   `json:"pattern,omitempty"`
    Recursive bool     `json:"recursive,omitempty"`
    Include   []string `json:"include,omitempty"` // File extensions to include
    Exclude   []string `json:"exclude,omitempty"` // File extensions to exclude
}
```

#### 5. Glob Tool

**Purpose**: Find files using glob patterns with advanced matching capabilities
**Features**:

- Advanced glob pattern matching (**/\*.go, **/_test_.go)
- Multiple pattern support
- Case-sensitive/insensitive matching
- Follow/don't follow symlinks
- Detailed file information
- Sorted results by modification time or name

**Implementation Pattern**:

```go
type GlobTool struct {
    Patterns    []string `json:"patterns"`              // Array of glob patterns
    Path        string   `json:"path,omitempty"`        // Base path (defaults to current dir)
    CaseSensitive bool   `json:"case_sensitive,omitempty"`
    FollowSymlinks bool  `json:"follow_symlinks,omitempty"`
    SortBy      string   `json:"sort_by,omitempty"`     // "modtime", "name", "size"
    SortOrder   string   `json:"sort_order,omitempty"`  // "asc", "desc"
    IncludeInfo bool     `json:"include_info,omitempty"` // Include file metadata
}
```

### System Tools

#### 5. Shell Tool

**Purpose**: Execute shell commands with proper security measures
**Features**:

- Command execution with timeout
- Environment variable support
- Working directory specification
- Output streaming
- Error code capture
- Command safety validation

**Implementation Pattern**:

```go
type ShellTool struct {
    Command     string            `json:"command"`
    Args        []string         `json:"args,omitempty"`
    WorkingDir  string           `json:"working_dir,omitempty"`
    Environment map[string]string `json:"environment,omitempty"`
    Timeout     int              `json:"timeout,omitempty"` // seconds
    Shell       string           `json:"shell,omitempty"`   // "bash", "sh", "powershell"
}
```

#### 6. Environment Tool

**Purpose**: Manage environment variables and system information
**Features**:

- Get/set environment variables
- System information retrieval
- Path manipulation utilities
- Configuration file management

### Web Tools

#### 7. Fetch Tool

**Purpose**: Fetch content from web resources
**Features**:

- HTTP/HTTPS request support
- JSON/XML/HTML parsing
- Header management
- Authentication (Basic, Bearer, API keys)
- Rate limiting and retries
- Content-type negotiation

**Implementation Pattern**:

```go
type FetchTool struct {
    URL         string            `json:"url"`
    Method      string            `json:"method,omitempty"` // GET, POST, PUT, DELETE
    Headers     map[string]string `json:"headers,omitempty"`
    Body        string            `json:"body,omitempty"`
    Format      string            `json:"format,omitempty"`    // "text", "json", "html"
    Timeout     int               `json:"timeout,omitempty"`   // seconds
    MaxRetries  int               `json:"max_retries,omitempty"`
}
```

#### 8. Download Tool

**Purpose**: Download files from URLs
**Features**:

- Resume interrupted downloads
- Progress tracking
- Integrity verification (checksums)
- Large file streaming
- Output directory management

### Task Management Tools

#### 9. TaskList Tool

**Purpose**: Create and manage task lists
**Features**:

- Create hierarchical todo lists
- Task state management (pending, in_progress, completed, cancelled)
- Priority levels (high, medium, low)
- Task dependencies
- Progress tracking

**Implementation Pattern**:

```go
type TaskListTool struct {
    Action      string     `json:"action"`       // "create", "update", "list", "delete"
    ListID      string     `json:"list_id,omitempty"`
    TaskID      string     `json:"task_id,omitempty"`
    Title       string     `json:"title,omitempty"`
    Description string     `json:"description,omitempty"`
    Priority    string     `json:"priority,omitempty"`
    Status      string     `json:"status,omitempty"`
    ParentID    string     `json:"parent_id,omitempty"`
}
```

#### 10. TaskExec Tool

**Purpose**: Execute tasks with tracking and monitoring
**Features**:

- Task execution with progress updates
- Result capture and storage
- Error handling and retry logic
- Parallel task execution
- Task cancellation support

### Search and Analysis Tools

#### 11. Search Tool

**Purpose**: Search for content in files and directories
**Features**:

- Content-based search with regex
- File name pattern matching
- Recursive directory search
- Result filtering and sorting
- Context lines around matches

**Implementation Pattern**:

```go
type SearchTool struct {
    Pattern     string   `json:"pattern"`
    Path        string   `json:"path"`
    Include     []string `json:"include,omitempty"`     // File patterns to include
    Exclude     []string `json:"exclude,omitempty"`     // File patterns to exclude
    Recursive   bool     `json:"recursive,omitempty"`
    IgnoreCase  bool     `json:"ignore_case,omitempty"`
    Context     int      `json:"context,omitempty"`     // Lines of context around matches
    MaxResults  int      `json:"max_results,omitempty"`
}
```

#### 12. Grep Tool

**Purpose**: Advanced pattern matching and content extraction
**Features**:

- Regular expression support
- Multi-file searching
- Statistical analysis (match counts, patterns)
- Structured output (JSON/CSV)

### Data Processing Tools

#### 13. JSON Tool

**Purpose**: Parse, validate, and manipulate JSON data
**Features**:

- JSON parsing and validation
- Path-based querying (JSONPath)
- Data transformation and filtering
- Format conversion (JSON to CSV, YAML, etc.)

#### 14. CSV Tool

**Purpose**: Handle CSV data operations
**Features**:

- CSV parsing with various delimiters
- Data filtering and transformation
- Statistical operations
- Export to different formats

#### 15. Template Tool

**Purpose**: Text templating and code generation
**Features**:

- Template rendering with data
- Multiple template engines (Go templates, Jinja2-style)
- Conditional logic and loops
- Template inheritance

## Custom Tool Framework

### Tool Interface Design

All tools should implement a common interface for consistency:

```go
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
```

### Tool Registry

Implement a tool registry system for managing tools:

```go
type ToolRegistry struct {
    tools map[string]Tool
    mu    sync.RWMutex
}

func NewToolRegistry() *ToolRegistry {
    return &ToolRegistry{
        tools: make(map[string]Tool),
    }
}

func (r *ToolRegistry) Register(tool Tool) error {
    r.mu.Lock()
    defer r.mu.Unlock()

    if _, exists := r.tools[tool.Name()]; exists {
        return fmt.Errorf("tool %s already registered", tool.Name())
    }

    r.tools[tool.Name()] = tool
    return nil
}

func (r *ToolRegistry) Get(name string) (Tool, error) {
    r.mu.RLock()
    defer r.mu.RUnlock()

    tool, exists := r.tools[name]
    if !exists {
        return nil, fmt.Errorf("tool %s not found", name)
    }

    return tool, nil
}
```

### Security and Safety Features

1. **Tool Validation**

   - Parameter validation before execution
   - Input sanitization
   - Sandbox execution where appropriate

2. **Access Control**

   - Tool permission levels
   - Resource limits (file size, execution time)
   - User confirmation for destructive operations

3. **Error Handling**
   - Comprehensive error reporting
   - Graceful failure modes
   - Operation rollback capabilities

### Custom Tool Creation

Users can create custom tools by implementing the Tool interface:

```go
// Example: Custom Database Query Tool
type DatabaseQueryTool struct {
    connectionString string
}

func (t *DatabaseQueryTool) Name() string {
    return "database_query"
}

func (t *DatabaseQueryTool) Description() string {
    return "Execute SQL queries against a database"
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
                "description": "Query parameters",
            },
        },
        "required": []string{"query"},
    }
}

func (t *DatabaseQueryTool) Execute(ctx context.Context, params json.RawMessage) (interface{}, error) {
    // Implementation here
}
```

### Integration with TUI Applications

Tools should integrate seamlessly with the existing TUI framework:

1. **Tool Selection Interface**

   - Interactive tool picker
   - Tool categorization
   - Quick access to frequently used tools

2. **Parameter Input**

   - Dynamic forms based on tool schema
   - JSON parameter entry
   - Parameter validation and suggestions

3. **Result Display**
   - Formatted output display
   - Result export options
   - Progress indicators for long operations

## Best Practices

### Tool Implementation

1. **Error Messages**: Provide clear, actionable error messages
2. **Timeouts**: Implement reasonable timeouts for long operations
3. **Progress**: Show progress for long-running operations
4. **Safety**: Always use safe defaults (dry-run mode, confirmations)
5. **Consistency**: Follow consistent parameter naming and structure

### Security Considerations

1. **Path Traversal**: Sanitize file paths to prevent directory traversal
2. **Command Injection**: Validate and sanitize shell commands
3. **Resource Limits**: Implement limits on file sizes, execution time, memory usage
4. **Authentication**: Handle credentials securely
5. **Logging**: Log tool usage for security monitoring

### Performance Optimization

1. **Streaming**: Use streaming for large file operations
2. **Caching**: Cache results for frequently accessed data
3. **Batching**: Support batch operations for multiple files
4. **Parallelism**: Use goroutines for independent operations

## Implementation Roadmap

### Phase 1: Core Tools (Priority: High)

- Read, Write, Delete, List tools
- Shell tool with basic security
- TaskList management

### Phase 2: Web & Search (Priority: Medium)

- Fetch and Download tools
- Search and Grep tools
- Environment management

### Phase 3: Data Processing (Priority: Medium)

- JSON and CSV tools
- Template processing
- Advanced file operations

### Phase 4: Framework & Customization (Priority: Low)

- Custom tool framework
- Tool marketplace/registry
- Advanced security features

## Usage Examples

```go
// Example usage in an AI agent application
registry := tools.NewToolRegistry()

// Register core tools
registry.Register(tools.NewReadTool())
registry.Register(tools.NewWriteTool())
registry.Register(tools.NewTaskListTool())

// Execute a tool
readTool, _ := registry.Get("read")
params := []byte(`{"path": "config.json", "format": "json"}`)
result, err := readTool.Execute(ctx, params)
```

## Future Enhancements

1. **Plugin System**: Allow custom tools as external plugins
2. **Tool Chaining**: Enable tool workflows and pipelines
3. **AI Integration**: Tools that leverage AI capabilities
4. **Collaboration**: Multi-agent tool sharing and coordination
5. **Metrics**: Tool usage analytics and performance metrics

This framework provides a solid foundation for building AI agents with comprehensive tool capabilities while maintaining security, extensibility, and ease of use.
