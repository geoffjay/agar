package main

import (
	"context"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/geoffjay/agar/commands"
	"github.com/geoffjay/agar/tui"
)

// commandDemo demonstrates the command system with a prompt-based interface
type commandDemo struct {
	prompt     tui.PromptModel
	output     []string
	width      int
	height     int
	shouldExit bool
}

func main() {
	// Create a demo model
	model := newCommandDemo()

	// Run the program
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func newCommandDemo() *commandDemo {
	// Create command manager
	cmdManager := commands.NewManager()
	if err := cmdManager.Initialize(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to initialize commands: %v\n", err)
	}

	// Register a custom command
	greetCmd := commands.NewCommandFunc(
		"greet",
		"Greet someone by name",
		"/greet <name>",
		func(ctx context.Context, args []string, state commands.ApplicationState) error {
			name := "World"
			if len(args) > 0 {
				name = args[0]
			}
			state.AddLine(fmt.Sprintf("👋 Hello, %s!", name))
			return nil
		},
	).WithAliases("hi", "hello")

	if err := cmdManager.RegisterCommand(greetCmd); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to register greet command: %v\n", err)
	}

	// Register an echo command
	echoCmd := commands.NewCommandFunc(
		"echo",
		"Echo back the provided arguments",
		"/echo <message...>",
		func(ctx context.Context, args []string, state commands.ApplicationState) error {
			if len(args) == 0 {
				return fmt.Errorf("no message provided")
			}
			message := ""
			for i, arg := range args {
				if i > 0 {
					message += " "
				}
				message += arg
			}
			state.AddLine("Echo: " + message)
			return nil
		},
	)

	if err := cmdManager.RegisterCommand(echoCmd); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to register echo command: %v\n", err)
	}

	// Create prompt with command manager
	prompt := tui.NewPromptInput("> ", "Type a command (try /help)", tui.SingleLineMode).
		WithHistory(true).
		WithCommandManager(cmdManager)

	demo := &commandDemo{
		prompt: prompt,
		output: []string{
			"Command System Demo",
			"==================",
			"",
			"Welcome to the Agar command system demo!",
			"",
			"Try typing '/' to see available commands.",
			"Use Tab to autocomplete, ↑↓ to navigate suggestions.",
			"",
			"Available commands:",
			"  /help      - Show all available commands",
			"  /greet     - Greet someone",
			"  /echo      - Echo a message",
			"  /clear     - Clear the screen",
			"  /export    - Export state to file",
			"  /import    - Import state from file",
			"  /exit      - Exit the application",
			"",
		},
		width:      80,
		height:     24,
		shouldExit: false,
	}

	return demo
}

func (d *commandDemo) Init() tea.Cmd {
	return d.prompt.Init()
}

func (d *commandDemo) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if d.shouldExit {
		return d, tea.Quit
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		d.width = msg.Width
		d.height = msg.Height
		updatedPrompt, cmd := d.prompt.Update(msg)
		d.prompt = updatedPrompt.(tui.PromptModel)
		return d, cmd

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return d, tea.Quit
		}

	case tui.PromptSubmitMsg:
		// Handle submitted input
		input := msg.Input

		// Show the command in output
		d.output = append(d.output, "> "+input)

		// Check if it's a command
		if len(input) > 0 && input[0] == '/' {
			// Execute command
			if err := d.getCommandManager().Handle(context.Background(), input, d); err != nil {
				d.output = append(d.output, "Error: "+err.Error())
			}
		} else if input != "" {
			// Not a command, just echo it
			d.output = append(d.output, "Not a command. Try /help to see available commands.")
		}

		d.output = append(d.output, "")
		return d, nil
	}

	// Update prompt
	updatedPrompt, cmd := d.prompt.Update(msg)
	d.prompt = updatedPrompt.(tui.PromptModel)

	return d, cmd
}

func (d *commandDemo) View() string {
	// Calculate available height for output
	outputHeight := d.height - 5 // Reserve space for prompt and help

	// Get visible output lines (last N lines)
	visibleOutput := d.output
	if len(d.output) > outputHeight {
		visibleOutput = d.output[len(d.output)-outputHeight:]
	}

	// Build view
	view := ""
	for _, line := range visibleOutput {
		view += line + "\n"
	}

	view += "\n"
	view += d.prompt.View()

	return view
}

// ApplicationState interface implementation
func (d *commandDemo) GetContent() []string {
	return d.output
}

func (d *commandDemo) SetContent(lines []string) {
	d.output = lines
}

func (d *commandDemo) AddLine(line string) {
	d.output = append(d.output, line)
}

func (d *commandDemo) Clear() {
	d.output = []string{}
}

func (d *commandDemo) GetMode() string {
	return "DEMO"
}

func (d *commandDemo) SetMode(mode string) {
	// Not used in this demo
}

func (d *commandDemo) GetMetadata() map[string]interface{} {
	return make(map[string]interface{})
}

func (d *commandDemo) SetMetadata(key string, value interface{}) {
	// Not used in this demo
}

func (d *commandDemo) Exit() {
	d.shouldExit = true
}

// Helper to get command manager from prompt
func (d *commandDemo) getCommandManager() *commands.Manager {
	// Create a temporary prompt to extract the command manager
	// In a real application, you'd store this separately
	cmdManager := commands.NewManager()
	if err := cmdManager.Initialize(); err != nil {
		return cmdManager
	}

	// Re-register custom commands
	greetCmd := commands.NewCommandFunc(
		"greet",
		"Greet someone by name",
		"/greet <name>",
		func(ctx context.Context, args []string, state commands.ApplicationState) error {
			name := "World"
			if len(args) > 0 {
				name = args[0]
			}
			state.AddLine(fmt.Sprintf("👋 Hello, %s!", name))
			return nil
		},
	).WithAliases("hi", "hello")

	cmdManager.RegisterCommand(greetCmd)

	echoCmd := commands.NewCommandFunc(
		"echo",
		"Echo back the provided arguments",
		"/echo <message...>",
		func(ctx context.Context, args []string, state commands.ApplicationState) error {
			if len(args) == 0 {
				return fmt.Errorf("no message provided")
			}
			message := ""
			for i, arg := range args {
				if i > 0 {
					message += " "
				}
				message += arg
			}
			state.AddLine("Echo: " + message)
			return nil
		},
	)

	cmdManager.RegisterCommand(echoCmd)

	return cmdManager
}
