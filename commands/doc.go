// Package commands provides a comprehensive slash command system for agar applications.
//
// The command system includes:
//   - Built-in commands (/exit, /help, /export, /import, /clear)
//   - File-based command loading from .agar/commands/
//   - Programmatic command registration via RegisterCommand
//   - Command autocomplete support
//   - Thread-safe command registry
//
// # Basic Usage
//
// To use the command system in your application:
//
//	// Create application with commands enabled (default)
//	app := tui.NewApplication(tui.ApplicationConfig{
//	    Title: "My App",
//	    Version: "1.0.0",
//	    EnableCommands: true,
//	})
//
//	// Execute a command
//	app.Update(tui.NewCommandMsg("/help"))
//
// # Custom Commands
//
// Register custom commands programmatically:
//
//	cmd := commands.NewCommandFunc(
//	    "hello",
//	    "Say hello",
//	    "/hello [name]",
//	    func(ctx context.Context, args []string, state commands.ApplicationState) error {
//	        name := "World"
//	        if len(args) > 0 {
//	            name = args[0]
//	        }
//	        state.AddLine(fmt.Sprintf("Hello, %s!", name))
//	        return nil
//	    },
//	)
//	app.RegisterCommand(cmd)
//
// # File-Based Commands
//
// Create command files in .agar/commands/:
//
//	# .agar/commands/greet.yaml
//	name: greet
//	description: Greet the user
//	usage: /greet [name]
//	aliases:
//	  - hi
//	  - hello
//	script: |
//	  #!/bin/bash
//	  echo "Hello from script!"
//	  echo "Args: $AGAR_COMMAND_ARGS"
//
// # Custom Command Paths
//
// Specify custom command paths:
//
//	app := tui.NewApplication(tui.ApplicationConfig{
//	    Title: "My App",
//	    CommandPaths: []string{".myapp/commands", "/etc/myapp/commands"},
//	})
package commands
