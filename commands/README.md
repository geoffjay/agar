# Commands Package

The `commands` package provides a comprehensive slash command system for agar applications.

## Features

- **Built-in Commands**: `/exit`, `/help`, `/export`, `/import`, `/clear`
- **File-Based Commands**: Load commands from YAML files in `.agar/commands/`
- **Programmatic Registration**: Register custom commands via `RegisterCommand()`
- **Command Autocomplete**: Tab completion and command suggestions in prompt input
- **Thread-Safe Registry**: Concurrent-safe command management
- **Command Aliases**: Support for multiple names per command
- **Script Execution**: Execute shell scripts from file-based commands

## Quick Start

### Enable Commands in Your Application

Commands are enabled by default when creating an application:

```go
app := tui.NewApplication(tui.ApplicationConfig{
    Title:          "My App",
    Version:        "1.0.0",
    EnableCommands: true, // Default: true
})
```

### Execute Commands

Send a `CommandMsg` to execute a command:

```go
// In your Update() method
case tea.KeyMsg:
    if msg.String() == "enter" {
        input := getInput() // Get user input
        if strings.HasPrefix(input, "/") {
            return model, func() tea.Msg {
                return tui.NewCommandMsg(input)
            }
        }
    }
```

### Built-in Commands

The following commands are available by default:

| Command | Description | Usage |
|---------|-------------|-------|
| `/exit` | Exit the application | `/exit` |
| `/quit`, `/q` | Aliases for `/exit` | `/quit` |
| `/help` | List all available commands | `/help [command]` |
| `/h`, `/?` | Aliases for `/help` | `/h` |
| `/export` | Export application state | `/export [filename]` |
| `/save` | Alias for `/export` | `/save` |
| `/import` | Import application state | `/import <filename>` |
| `/load` | Alias for `/import` | `/load` |
| `/clear` | Clear application content | `/clear` |
| `/cls` | Alias for `/clear` | `/cls` |

## Custom Commands

### Programmatic Registration

Register commands in your application code:

```go
// Create a simple command
greetCmd := commands.NewCommandFunc(
    "greet",
    "Greet someone by name",
    "/greet <name>",
    func(ctx context.Context, args []string, state commands.ApplicationState) error {
        name := "World"
        if len(args) > 0 {
            name = args[0]
        }
        state.AddLine(fmt.Sprintf("Hello, %s!", name))
        return nil
    },
)

// Add aliases
greetCmd = greetCmd.WithAliases("hi", "hello")

// Register with the application
app.RegisterCommand(greetCmd)
```

### File-Based Commands

Create YAML files in `.agar/commands/`:

```yaml
# .agar/commands/time.yaml
name: time
description: Show the current time
usage: /time
aliases:
  - now
  - datetime
script: |
  #!/bin/bash
  date "+%Y-%m-%d %H:%M:%S"
```

File-based commands support:
- **Script Execution**: Run shell scripts with access to environment variables
- **Arguments**: Access command arguments via `$AGAR_COMMAND_ARGS` (JSON array)
- **Application State**: Access app state via `$AGAR_MODE`, `$AGAR_CONTENT` (JSON)
- **Custom Metadata**: Define custom environment variables

### Custom Command Paths

Specify custom paths to load commands from:

```go
app := tui.NewApplication(tui.ApplicationConfig{
    Title:        "My App",
    CommandPaths: []string{
        ".myapp/commands",
        "/etc/myapp/commands",
        os.ExpandEnv("$HOME/.config/myapp/commands"),
    },
})
```

## Command Autocomplete

The prompt input component supports command autocomplete:

```go
// Create command manager
cmdManager := commands.NewManager()
cmdManager.Initialize()

// Create prompt with command manager
prompt := tui.NewPromptInput("> ", "Type a command...", tui.SingleLineMode).
    WithHistory(true).
    WithCommandManager(cmdManager)
```

**Autocomplete Features:**
- Type `/` to see all available commands
- Commands appear with descriptions as you type
- Use `↑`/`↓` to navigate suggestions
- Press `Tab` to accept the selected completion
- Supports filtering by command name

## Advanced Usage

### Access Command Manager

Get direct access to the command manager for advanced operations:

```go
cmdManager := app.GetCommandManager()

// List all commands
commands := cmdManager.ListCommands()

// Get specific command
cmd, err := cmdManager.GetCommand("help")

// Check if command exists
if cmdManager.IsCommand("/help") {
    // ...
}

// Get completions
completions := cmdManager.GetCompletions("/he")
```

### Custom Command Implementation

Implement the `Command` interface for full control:

```go
type MyCommand struct {
    // Your fields
}

func (c *MyCommand) Name() string {
    return "mycommand"
}

func (c *MyCommand) Description() string {
    return "Does something custom"
}

func (c *MyCommand) Usage() string {
    return "/mycommand [options]"
}

func (c *MyCommand) Aliases() []string {
    return []string{"mc", "mycmd"}
}

func (c *MyCommand) Execute(ctx context.Context, args []string, state commands.ApplicationState) error {
    // Your implementation
    state.AddLine("Command executed!")
    return nil
}

// Register it
app.RegisterCommand(&MyCommand{})
```

### ApplicationState Interface

Commands interact with the application through the `ApplicationState` interface:

```go
type ApplicationState interface {
    GetContent() []string
    SetContent(lines []string)
    AddLine(line string)
    Clear()
    GetMode() string
    SetMode(mode string)
    GetMetadata() map[string]interface{}
    SetMetadata(key string, value interface{})
    Exit()
}
```

## Examples

See the [commands example](/examples/commands/main.go) for a complete demonstration.

## Testing

Run tests for the commands package:

```bash
go test ./commands/...
```

## Architecture

The command system consists of:

- **Command Interface**: Defines the contract for all commands
- **Registry**: Thread-safe storage for registered commands
- **Handler**: Parses and executes commands
- **Loader**: Loads commands from YAML files
- **Manager**: High-level API combining all components
- **ScriptHandler**: Executes shell scripts for file-based commands

## Best Practices

1. **Use Descriptive Names**: Command names should be clear and intuitive
2. **Provide Help Text**: Always include usage information
3. **Validate Arguments**: Check argument count and format
4. **Return Errors**: Return errors instead of panicking
5. **Use Aliases**: Provide common shortcuts (e.g., `/q` for `/quit`)
6. **Namespace Commands**: Use prefixes for plugin commands (e.g., `/git:status`)
7. **Test Commands**: Write tests for custom commands
8. **Document Behavior**: Document side effects and requirements

## Troubleshooting

### Commands Not Loading

1. Check that `EnableCommands` is `true` (default)
2. Verify command files are in `.agar/commands/` or custom paths
3. Ensure YAML files are properly formatted
4. Check file permissions

### Command Not Found

1. Use `/help` to list available commands
2. Check for typos in command name
3. Verify the command is registered
4. Check if command has been unregistered

### Autocomplete Not Working

1. Ensure prompt has command manager: `.WithCommandManager(manager)`
2. Verify manager is initialized: `manager.Initialize()`
3. Check that commands are registered

### Script Execution Fails

1. Verify script has proper shebang (`#!/bin/bash`)
2. Check that script is executable
3. Verify required tools are installed
4. Check environment variables

## License

This package is part of the Agar project.
