# Agar Project TODO

## Tools Framework - Remaining Implementation

### Phase 2: Web & Search Tools

#### Fetch Tool

**Purpose**: Fetch content from web resources

**Features**:

- HTTP/HTTPS request support
- Multiple HTTP methods (GET, POST, PUT, DELETE, PATCH)
- JSON/XML/HTML parsing and content-type negotiation
- Header management (custom headers, authentication headers)
- Authentication support (Basic, Bearer tokens, API keys)
- Rate limiting and retry logic
- Response caching
- Timeout configuration
- Content-type negotiation

**Implementation Pattern**:

```go
type FetchTool struct{}

type FetchParams struct {
    URL        string            `json:"url"`
    Method     string            `json:"method,omitempty"`      // GET, POST, PUT, DELETE, PATCH
    Headers    map[string]string `json:"headers,omitempty"`
    Body       string            `json:"body,omitempty"`
    Format     string            `json:"format,omitempty"`      // "text", "json", "html", "xml"
    Timeout    int               `json:"timeout,omitempty"`     // seconds
    MaxRetries int               `json:"max_retries,omitempty"`
    Auth       *AuthConfig       `json:"auth,omitempty"`
}

type AuthConfig struct {
    Type   string `json:"type"`   // "basic", "bearer", "apikey"
    User   string `json:"user,omitempty"`
    Pass   string `json:"pass,omitempty"`
    Token  string `json:"token,omitempty"`
    APIKey string `json:"apikey,omitempty"`
    Header string `json:"header,omitempty"` // Header name for API key
}

type FetchResult struct {
    StatusCode   int               `json:"status_code"`
    Headers      map[string]string `json:"headers"`
    Content      string            `json:"content"`
    ContentType  string            `json:"content_type"`
    Size         int64             `json:"size"`
    Duration     int64             `json:"duration_ms"`
    Cached       bool              `json:"cached,omitempty"`
    RedirectURL  string            `json:"redirect_url,omitempty"`
}
```

**Usage Example**:

```go
tool := tools.NewFetchTool()
params := json.RawMessage(`{
    "url": "https://api.example.com/data",
    "method": "GET",
    "headers": {
        "Accept": "application/json"
    },
    "timeout": 30
}`)
result, err := tool.Execute(ctx, params)
```

---

#### Download Tool

**Purpose**: Download files from URLs with resume support

**Features**:

- Resume interrupted downloads
- Progress tracking callbacks
- Integrity verification (MD5, SHA256 checksums)
- Large file streaming (memory-efficient)
- Output directory management
- Parallel chunk downloads
- Bandwidth limiting

**Implementation Pattern**:

```go
type DownloadTool struct{}

type DownloadParams struct {
    URL          string `json:"url"`
    OutputPath   string `json:"output_path"`
    Resume       bool   `json:"resume,omitempty"`
    Checksum     string `json:"checksum,omitempty"`     // Expected checksum
    ChecksumType string `json:"checksum_type,omitempty"` // "md5", "sha256"
    Timeout      int    `json:"timeout,omitempty"`       // seconds
    MaxRetries   int    `json:"max_retries,omitempty"`
    ChunkSize    int64  `json:"chunk_size,omitempty"`    // bytes
}

type DownloadResult struct {
    Path         string `json:"path"`
    Size         int64  `json:"size"`
    Checksum     string `json:"checksum"`
    ChecksumType string `json:"checksum_type"`
    Duration     int64  `json:"duration_ms"`
    Resumed      bool   `json:"resumed,omitempty"`
    BytesResumed int64  `json:"bytes_resumed,omitempty"`
}
```

**Usage Example**:

```go
tool := tools.NewDownloadTool()
params := json.RawMessage(`{
    "url": "https://example.com/largefile.zip",
    "output_path": "/downloads/file.zip",
    "resume": true,
    "checksum": "abc123...",
    "checksum_type": "sha256"
}`)
result, err := tool.Execute(ctx, params)
```

---

#### Search Tool

**Purpose**: Search for content in files and directories

**Features**:

- Content-based search with regular expressions
- File name pattern matching
- Recursive directory search
- Result filtering and sorting
- Context lines around matches (before/after)
- Case-sensitive/insensitive search
- Maximum result limits
- Search result highlighting

**Implementation Pattern**:

```go
type SearchTool struct{}

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

type SearchResult struct {
    Matches      []SearchMatch `json:"matches"`
    TotalFiles   int           `json:"total_files"`
    TotalMatches int           `json:"total_matches"`
}

type SearchMatch struct {
    File      string   `json:"file"`
    Line      int      `json:"line"`
    Column    int      `json:"column"`
    MatchText string   `json:"match_text"`
    LineText  string   `json:"line_text"`
    Context   []string `json:"context,omitempty"`
}
```

**Usage Example**:

```go
tool := tools.NewSearchTool()
params := json.RawMessage(`{
    "pattern": "func.*Error",
    "path": "/path/to/project",
    "recursive": true,
    "include": [".go"],
    "context": 2
}`)
result, err := tool.Execute(ctx, params)
```

---

#### Grep Tool

**Purpose**: Advanced pattern matching and content extraction

**Features**:

- Regular expression support with capture groups
- Multi-file searching
- Statistical analysis (match counts, pattern frequency)
- Structured output (JSON/CSV formats)
- Line number reporting
- Inverted matching (lines that don't match)
- Word boundary matching
- Multi-line pattern support

**Implementation Pattern**:

```go
type GrepTool struct{}

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

type GrepResult struct {
    Matches      []GrepMatch     `json:"matches,omitempty"`
    Statistics   *GrepStatistics `json:"statistics,omitempty"`
    TotalMatches int             `json:"total_matches"`
}

type GrepMatch struct {
    File     string   `json:"file"`
    Line     int      `json:"line"`
    Content  string   `json:"content"`
    Captures []string `json:"captures,omitempty"` // Regex capture groups
}

type GrepStatistics struct {
    FilesSearched int            `json:"files_searched"`
    FilesMatched  int            `json:"files_matched"`
    TotalMatches  int            `json:"total_matches"`
    PatternCounts map[string]int `json:"pattern_counts,omitempty"`
}
```

**Usage Example**:

```go
tool := tools.NewGrepTool()
params := json.RawMessage(`{
    "pattern": "TODO|FIXME",
    "files": ["**/*.go"],
    "recursive": true,
    "line_numbers": true,
    "count": false
}`)
result, err := tool.Execute(ctx, params)
```

---

### Phase 3: Data Processing Tools

#### JSON Tool

**Purpose**: Parse, validate, and manipulate JSON data

**Features**:

- JSON parsing and validation
- JSONPath querying for data extraction
- Data transformation and filtering
- Format conversion (JSON to CSV, YAML, XML)
- Schema validation
- Pretty printing and minification
- Merge and diff operations
- Query result caching

**Implementation Pattern**:

```go
type JSONTool struct{}

type JSONParams struct {
    Action    string                 `json:"action"` // "parse", "query", "validate", "transform", "convert"
    Data      string                 `json:"data,omitempty"`
    Path      string                 `json:"path,omitempty"`     // File path or JSONPath query
    Query     string                 `json:"query,omitempty"`    // JSONPath expression
    Schema    string                 `json:"schema,omitempty"`   // JSON Schema for validation
    Transform map[string]interface{} `json:"transform,omitempty"` // Transformation rules
    Format    string                 `json:"format,omitempty"`   // Output format
    Pretty    bool                   `json:"pretty,omitempty"`
}

type JSONResult struct {
    Action    string      `json:"action"`
    Data      interface{} `json:"data,omitempty"`
    Valid     bool        `json:"valid,omitempty"`
    Errors    []string    `json:"errors,omitempty"`
    Converted string      `json:"converted,omitempty"`
}
```

**Usage Example**:

```go
tool := tools.NewJSONTool()

// Parse and query
params := json.RawMessage(`{
    "action": "query",
    "path": "/path/to/data.json",
    "query": "$.users[?(@.age > 25)].name"
}`)

// Validate against schema
params = json.RawMessage(`{
    "action": "validate",
    "data": "{\"name\": \"John\"}",
    "schema": "{\"type\": \"object\", \"required\": [\"name\"]}"
}`)

// Convert to CSV
params = json.RawMessage(`{
    "action": "convert",
    "path": "/path/to/data.json",
    "format": "csv"
}`)
```

---

#### CSV Tool

**Purpose**: Handle CSV data operations

**Features**:

- CSV parsing with configurable delimiters
- Header detection and validation
- Data filtering and transformation
- Column selection and reordering
- Statistical operations (sum, avg, min, max)
- Export to different formats (JSON, XML)
- Merge multiple CSV files
- Data type inference

**Implementation Pattern**:

```go
type CSVTool struct{}

type CSVParams struct {
    Action    string   `json:"action"` // "parse", "filter", "transform", "stats", "convert"
    Path      string   `json:"path,omitempty"`
    Data      string   `json:"data,omitempty"`
    Delimiter string   `json:"delimiter,omitempty"` // Default: ","
    HasHeader bool     `json:"has_header,omitempty"`
    Columns   []string `json:"columns,omitempty"`   // Columns to select
    Filter    string   `json:"filter,omitempty"`    // Filter expression
    Format    string   `json:"format,omitempty"`    // Output format
}

type CSVResult struct {
    Action     string                 `json:"action"`
    Data       []map[string]string    `json:"data,omitempty"`
    Statistics map[string]interface{} `json:"statistics,omitempty"`
    RowCount   int                    `json:"row_count"`
    Converted  string                 `json:"converted,omitempty"`
}
```

**Usage Example**:

```go
tool := tools.NewCSVTool()

// Parse CSV
params := json.RawMessage(`{
    "action": "parse",
    "path": "/path/to/data.csv",
    "has_header": true
}`)

// Filter rows
params = json.RawMessage(`{
    "action": "filter",
    "path": "/path/to/data.csv",
    "filter": "age > 25 AND city == 'NYC'"
}`)

// Calculate statistics
params = json.RawMessage(`{
    "action": "stats",
    "path": "/path/to/data.csv",
    "columns": ["age", "salary"]
}`)

// Convert to JSON
params = json.RawMessage(`{
    "action": "convert",
    "path": "/path/to/data.csv",
    "format": "json"
}`)
```

---

#### Template Tool

**Purpose**: Text templating and code generation

**Features**:

- Template rendering with Go templates
- Data binding from JSON/YAML
- Conditional logic and loops
- Template inheritance and includes
- Custom function support
- Multiple template engines (Go templates, text/template)
- Template validation
- Partial rendering

**Implementation Pattern**:

```go
type TemplateTool struct{}

type TemplateParams struct {
    Template     string                 `json:"template"`           // Template string
    TemplateFile string                 `json:"template_file,omitempty"` // Or path to template
    Data         map[string]interface{} `json:"data"`              // Data to render
    DataFile     string                 `json:"data_file,omitempty"`     // Or path to data file
    Functions    map[string]string      `json:"functions,omitempty"`     // Custom functions
    Partials     map[string]string      `json:"partials,omitempty"`      // Partial templates
    Engine       string                 `json:"engine,omitempty"`        // "go", "text"
}

type TemplateResult struct {
    Output   string   `json:"output"`
    Errors   []string `json:"errors,omitempty"`
    Warnings []string `json:"warnings,omitempty"`
}
```

**Usage Example**:

```go
tool := tools.NewTemplateTool()

// Basic template rendering
params := json.RawMessage(`{
    "template": "Hello, {{.Name}}! You are {{.Age}} years old.",
    "data": {
        "Name": "Alice",
        "Age": 30
    }
}`)

// From files
params = json.RawMessage(`{
    "template_file": "/templates/email.tmpl",
    "data_file": "/data/users.json"
}`)

// With loops and conditionals
params = json.RawMessage(`{
    "template": "{{range .Items}}{{if gt .Price 100}}{{.Name}}: ${{.Price}}{{end}}{{end}}",
    "data": {
        "Items": [
            {"Name": "Widget", "Price": 150},
            {"Name": "Gadget", "Price": 50}
        ]
    }
}`)
```

---

### Phase 4: Advanced Features

#### Environment Tool

**Purpose**: Manage environment variables and system information

**Features**:

- Get/set environment variables
- List all environment variables
- System information retrieval (OS, arch, hostname, user)
- Path manipulation utilities (join, split, clean)
- Configuration file management
- Environment variable expansion

**Implementation Pattern**:

```go
type EnvironmentTool struct{}

type EnvironmentParams struct {
    Action    string            `json:"action"` // "get", "set", "list", "unset", "info", "path"
    Variable  string            `json:"variable,omitempty"`
    Value     string            `json:"value,omitempty"`
    Variables map[string]string `json:"variables,omitempty"` // For bulk operations
    PathOp    string            `json:"path_op,omitempty"`   // "join", "split", "clean", "abs"
    Paths     []string          `json:"paths,omitempty"`
}

type EnvironmentResult struct {
    Action      string            `json:"action"`
    Variable    string            `json:"variable,omitempty"`
    Value       string            `json:"value,omitempty"`
    Variables   map[string]string `json:"variables,omitempty"`
    SystemInfo  *SystemInfo       `json:"system_info,omitempty"`
    PathResult  string            `json:"path_result,omitempty"`
}

type SystemInfo struct {
    OS           string `json:"os"`
    Arch         string `json:"arch"`
    Hostname     string `json:"hostname"`
    User         string `json:"user"`
    HomeDir      string `json:"home_dir"`
    TempDir      string `json:"temp_dir"`
    NumCPU       int    `json:"num_cpu"`
    GOROOT       string `json:"goroot,omitempty"`
    GOPATH       string `json:"gopath,omitempty"`
}
```

**Usage Example**:

```go
tool := tools.NewEnvironmentTool()

// Get variable
params := json.RawMessage(`{
    "action": "get",
    "variable": "PATH"
}`)

// Set variable
params = json.RawMessage(`{
    "action": "set",
    "variable": "MY_VAR",
    "value": "my_value"
}`)

// Get system info
params = json.RawMessage(`{
    "action": "info"
}`)

// Path operations
params = json.RawMessage(`{
    "action": "path",
    "path_op": "join",
    "paths": ["/home/user", "documents", "file.txt"]
}`)
```

---

#### TaskExec Tool

**Purpose**: Execute tasks with tracking and monitoring

**Features**:

- Task execution with real-time progress updates
- Result capture and storage
- Error handling and automatic retry logic
- Parallel task execution
- Task cancellation support
- Task dependencies and ordering
- Execution history and logging
- Resource usage tracking

**Implementation Pattern**:

```go
type TaskExecTool struct{}

type TaskExecParams struct {
    TaskID       string            `json:"task_id"`
    Command      string            `json:"command,omitempty"`
    Script       string            `json:"script,omitempty"`
    Environment  map[string]string `json:"environment,omitempty"`
    Timeout      int               `json:"timeout,omitempty"`
    Retries      int               `json:"retries,omitempty"`
    RetryDelay   int               `json:"retry_delay,omitempty"` // seconds
    Dependencies []string          `json:"dependencies,omitempty"` // Task IDs
    Parallel     bool              `json:"parallel,omitempty"`
}

type TaskExecResult struct {
    TaskID       string          `json:"task_id"`
    Status       string          `json:"status"` // "running", "completed", "failed", "cancelled"
    Output       string          `json:"output,omitempty"`
    Error        string          `json:"error,omitempty"`
    ExitCode     int             `json:"exit_code,omitempty"`
    StartTime    int64           `json:"start_time"`
    EndTime      int64           `json:"end_time,omitempty"`
    Duration     int64           `json:"duration_ms,omitempty"`
    Retries      int             `json:"retries,omitempty"`
    ResourceUse  *ResourceUsage  `json:"resource_usage,omitempty"`
}

type ResourceUsage struct {
    CPUPercent     float64 `json:"cpu_percent"`
    MemoryBytes    int64   `json:"memory_bytes"`
    DiskReadBytes  int64   `json:"disk_read_bytes"`
    DiskWriteBytes int64   `json:"disk_write_bytes"`
}
```

**Usage Example**:

```go
tool := tools.NewTaskExecTool()

// Execute with retries
params := json.RawMessage(`{
    "task_id": "build-123",
    "command": "make build",
    "timeout": 300,
    "retries": 3,
    "retry_delay": 5
}`)

// Execute with dependencies
params = json.RawMessage(`{
    "task_id": "deploy-456",
    "command": "make deploy",
    "dependencies": ["build-123", "test-789"]
}`)
```

---

### Phase 5: Framework Extensions

#### Plugin System

**Purpose**: Allow custom tools as external plugins

**Features**:

- Dynamic tool loading from shared libraries (.so, .dll, .dylib)
- Plugin discovery and automatic registration
- Version compatibility checking
- Plugin sandboxing and isolation
- Hot reload support
- Plugin marketplace/registry
- Dependency management

**Implementation Pattern**:

```go
type PluginManager struct {
    plugins map[string]*Plugin
    loader  *PluginLoader
}

type Plugin struct {
    Name        string
    Version     string
    Path        string
    Tool        Tool
    Metadata    PluginMetadata
}

type PluginMetadata struct {
    Author       string
    Description  string
    Dependencies []string
    MinVersion   string
    MaxVersion   string
    Homepage     string
}

type PluginLoader struct {
    searchPaths []string
    cache       map[string]*Plugin
}
```

**Usage Example**:

```go
manager := tools.NewPluginManager()

// Load plugin from file
plugin, err := manager.LoadPlugin("/plugins/myplugin.so")

// Auto-discover plugins
manager.DiscoverPlugins("/plugins")

// Use plugin tool
tool := plugin.Tool
result, err := tool.Execute(ctx, params)
```

---

#### Tool Chaining / Pipeline System

**Purpose**: Enable tool workflows and pipelines

**Features**:

- Sequential tool execution with data flow
- Data passing between tools
- Conditional execution
- Error handling in pipelines
- Pipeline templates and reusability
- Variable interpolation
- Parallel step execution
- Pipeline validation

**Implementation Pattern**:

```go
type Pipeline struct {
    Name        string
    Description string
    Steps       []PipelineStep
    Variables   map[string]interface{}
}

type PipelineStep struct {
    Name      string          `json:"name"`
    Tool      string          `json:"tool"`
    Params    json.RawMessage `json:"params"`
    OutputVar string          `json:"output_var,omitempty"` // Store result in variable
    Condition string          `json:"condition,omitempty"`  // Execute if condition is true
    OnError   string          `json:"on_error,omitempty"`   // "continue", "abort", "retry"
    Retry     int             `json:"retry,omitempty"`
    Parallel  bool            `json:"parallel,omitempty"`
}

type PipelineResult struct {
    Name       string                 `json:"name"`
    Status     string                 `json:"status"` // "completed", "failed", "partial"
    Steps      []StepResult           `json:"steps"`
    Variables  map[string]interface{} `json:"variables"`
    Duration   int64                  `json:"duration_ms"`
}

type StepResult struct {
    Name     string      `json:"name"`
    Status   string      `json:"status"`
    Output   interface{} `json:"output,omitempty"`
    Error    string      `json:"error,omitempty"`
    Duration int64       `json:"duration_ms"`
}
```

**Usage Example**:

```go
pipeline := &Pipeline{
    Name: "process-logs",
    Steps: []PipelineStep{
        {
            Name:      "read-logs",
            Tool:      "read",
            Params:    json.RawMessage(`{"path": "app.log"}`),
            OutputVar: "logs",
        },
        {
            Name:      "find-errors",
            Tool:      "grep",
            Params:    json.RawMessage(`{"pattern": "ERROR", "data": "{{.logs}}"}`),
            OutputVar: "errors",
        },
        {
            Name:      "save-errors",
            Tool:      "write",
            Params:    json.RawMessage(`{"path": "errors.log", "content": "{{.errors}}"}`),
            Condition: "len(.errors) > 0",
        },
    },
}

executor := tools.NewPipelineExecutor()
result, err := executor.Execute(ctx, pipeline)
```

---

#### Metrics and Monitoring

**Purpose**: Tool usage analytics and performance monitoring

**Features**:

- Tool execution tracking
- Performance profiling (execution time, memory usage)
- Error rate monitoring
- Resource usage statistics
- Custom metrics and tags
- Export to monitoring systems (Prometheus, StatsD)
- Alerting thresholds
- Historical data retention

**Implementation Pattern**:

```go
type MetricsCollector struct {
    metrics  map[string]*ToolMetrics
    exporters []MetricsExporter
}

type ToolMetrics struct {
    ToolName        string
    ExecutionCount  int64
    TotalDuration   int64
    AverageDuration int64
    MinDuration     int64
    MaxDuration     int64
    ErrorCount      int64
    SuccessCount    int64
    LastExecuted    int64
    ResourceUsage   AggregateResourceUsage
}

type AggregateResourceUsage struct {
    TotalCPUTime      int64
    AverageCPUTime    int64
    PeakMemoryBytes   int64
    AverageMemoryBytes int64
    TotalDiskIO       int64
}

type MetricsExporter interface {
    Export(metrics map[string]*ToolMetrics) error
}
```

**Usage Example**:

```go
collector := tools.NewMetricsCollector()

// Wrap tool execution
result, err := collector.ExecuteWithMetrics(tool, params)

// Get metrics for a tool
metrics := collector.GetMetrics("read")
fmt.Printf("Read tool executed %d times, avg duration: %dms\n",
    metrics.ExecutionCount, metrics.AverageDuration)

// Export to Prometheus
prometheusExporter := tools.NewPrometheusExporter(":9090")
collector.AddExporter(prometheusExporter)
collector.Export()
```

---

## Implementation Guidelines

### For Each New Tool

When implementing tools from this TODO list:

1. **Interface Compliance**

   - Implement all Tool interface methods (Name, Description, Execute, Validate, Schema)
   - Follow existing naming conventions (e.g., NewXxxTool() constructor)
   - Use consistent parameter and result structures

2. **Parameter Validation**

   - Validate all required parameters in Validate() method
   - Check parameter types and ranges
   - Provide clear, actionable validation error messages
   - Use JSON schema in Schema() method

3. **Error Handling**

   - Return descriptive errors with context
   - Use error wrapping with fmt.Errorf and %w
   - Handle edge cases gracefully
   - Clean up resources on error

4. **Security**

   - Sanitize all inputs (especially paths and commands)
   - Implement timeouts for potentially long operations
   - Validate paths to prevent directory traversal
   - Block dangerous operations (similar to Shell tool)
   - Use safe defaults

5. **Testing**

   - Write comprehensive unit tests (aim for >80% coverage)
   - Test happy path and error cases
   - Test edge cases and boundary conditions
   - Test with various parameter combinations
   - Use table-driven tests for multiple scenarios

6. **Documentation**
   - Add tool to docs/tools.md when implemented
   - Provide clear usage examples
   - Document security considerations
   - Include parameter descriptions with types and defaults

### Priority Order

Recommended implementation order based on dependencies and utility:

**High Priority** (Most Useful Immediately):

1. Environment Tool - System interaction needed by many applications
2. Search Tool - Complements existing Glob/List tools
3. Grep Tool - Advanced pattern matching

**Medium Priority** (Extends Capabilities): 4. Fetch Tool - Web access for modern applications 5. JSON Tool - Data processing for APIs and configs 6. Download Tool - File retrieval from web

**Lower Priority** (Specialized Use Cases): 7. CSV Tool - Specific data format handling 8. Template Tool - Code generation support 9. TaskExec Tool - Advanced task execution

**Future/Advanced** (Framework Extensions): 10. Plugin System - Extensibility framework 11. Tool Chaining - Workflow automation 12. Metrics and Monitoring - Observability

### Testing Requirements

Each tool should have:

- Minimum 10-15 unit tests
- Tests for all parameter combinations
- Error case coverage
- Edge case testing
- Table-driven test structure where appropriate
- Overall >80% code coverage target

### Example Test Structure

```go
func TestNewTool_Name(t *testing.T) {
    tool := tools.NewXxxTool()
    if tool.Name() != "expected_name" {
        t.Errorf("Expected name 'expected_name', got '%s'", tool.Name())
    }
}

func TestNewTool_Validate(t *testing.T) {
    tool := tools.NewXxxTool()

    tests := []struct {
        name    string
        params  string
        wantErr bool
    }{
        {"valid params", `{"param": "value"}`, false},
        {"missing required", `{}`, true},
        {"invalid type", `{"param": 123}`, true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tool.Validate(json.RawMessage(tt.params))
            if (err != nil) != tt.wantErr {
                t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}

func TestNewTool_Execute(t *testing.T) {
    // Test actual execution
}
```
