package app

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/geoffjay/agar/tui"
)

// cliModel wraps the Application and Prompt components
type cliModel struct {
	app    *tui.Application
	prompt tui.PromptModel
	width  int
	height int
}

// RunTUI launches the TUI application
func RunTUI() error {
	cwd, err := os.Getwd()
	if err != nil {
		cwd = "/"
	}

	// Create TUI application
	app := tui.NewApplication(tui.ApplicationConfig{
		Title:          "Agar",
		Version:        "0.1.0",
		Mode:           "INTERACTIVE",
		Directory:      cwd,
		PanelMargin:    0,
		PanelPadding:   1,
		BorderStyle:    tui.NoBorder,
		EnableCommands: true, // Explicitly enable command system
	})

	// Add welcome content
	app.AddLine("")
	app.AddLine("                      WELCOME TO AGAR CLI")
	app.AddLine("")
	app.AddLine("Agar is a comprehensive framework for building AI agent applications")
	app.AddLine("with TUI components and tool management.")
	app.AddLine("")
	app.AddLine("─────────────────────────────────────────────────────────────────────────────────────────")
	app.AddLine("")
	app.AddLine("AVAILABLE COMMANDS")
	app.AddLine("")
	app.AddLine("  help              Show this help message")
	app.AddLine("  tools             List available tools")
	app.AddLine("  components        List TUI components")
	app.AddLine("  init <name>       Initialize a new Agar project (requires ANTHROPIC_API_KEY)")
	app.AddLine("  clear             Clear the screen")
	app.AddLine("  exit              Exit the CLI")
	app.AddLine("")
	app.AddLine("─────────────────────────────────────────────────────────────────────────────────────────")
	app.AddLine("")
	app.AddLine("Type a command below or press Ctrl+C to exit.")
	app.AddLine("")

	// Create prompt (no placeholder for cleaner look)
	cmdMgr := app.GetCommandManager()
	if cmdMgr != nil {
		app.AddLine(fmt.Sprintf("Command system loaded with %d commands", len(cmdMgr.ListCommands())))
		app.AddLine("")
	} else {
		app.AddLine("Warning: Command system not initialized")
		app.AddLine("")
	}

	prompt := tui.NewPromptInput("λ ", "", tui.SingleLineMode).
		WithHistory(true).
		WithCommandManager(cmdMgr)

	model := cliModel{
		app:    app,
		prompt: prompt,
		width:  80,
		height: 24,
	}

	// Run the application
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("failed to run TUI: %w", err)
	}

	return nil
}

func (m cliModel) Init() tea.Cmd {
	return nil
}

func (m cliModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Update application size (leave room for prompt and separator)
		appHeight := m.height - 3 // -3 for prompt (1) + separator (1) + footer (1)
		updatedApp, _ := m.app.Update(tea.WindowSizeMsg{Width: m.width, Height: appHeight})
		m.app = updatedApp.(*tui.Application)

		return m, nil

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

	case tui.PromptSubmitMsg:
		input := strings.TrimSpace(msg.Input)

		// Check if this is a slash command
		if strings.HasPrefix(input, "/") {
			// Echo the command to the output
			m.app.AddLine("")
			m.app.AddLine(lipgloss.NewStyle().Foreground(lipgloss.Color("86")).Render("> " + input))

			// Forward to Application's command system
			updatedApp, appCmd := m.app.Update(tui.NewCommandMsg(input))
			m.app = updatedApp.(*tui.Application)

			// Update prompt to clear it
			updated, cmd := m.prompt.Update(msg)
			m.prompt = updated.(tui.PromptModel)

			// Combine commands (appCmd may be tea.Quit if exit was called)
			return m, tea.Batch(appCmd, cmd)
		}

		// Process non-slash commands with the hardcoded handler
		m.handleCommand(input)

		// Update prompt to clear it
		updated, cmd := m.prompt.Update(msg)
		m.prompt = updated.(tui.PromptModel)
		return m, cmd
	}

	// Forward messages to prompt
	updated, cmd := m.prompt.Update(msg)
	m.prompt = updated.(tui.PromptModel)
	return m, cmd
}

func (m *cliModel) handleCommand(input string) {
	input = strings.TrimSpace(input)
	if input == "" {
		return
	}

	parts := strings.Fields(input)
	command := parts[0]

	m.app.AddLine("")
	m.app.AddLine(lipgloss.NewStyle().Foreground(lipgloss.Color("86")).Render("> " + input))

	switch command {
	case "help":
		m.showHelp()

	case "tools":
		m.showTools()

	case "components":
		m.showComponents()

	case "init":
		if len(parts) < 2 {
			m.app.AddLine(lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render("Error: init requires a project name"))
			m.app.AddLine("Usage: init <project-name>")
		} else {
			projectName := parts[1]
			m.app.AddLine(fmt.Sprintf("Initializing project '%s'...", projectName))
			m.app.AddLine("Note: Set ANTHROPIC_API_KEY environment variable for AI features")
			m.app.AddLine("Tip: Use 'agar init' command outside TUI for full functionality")
		}

	case "clear":
		m.app.Clear()
		m.app.AddLine("Screen cleared. Type 'help' for available commands.")

	case "exit", "quit":
		m.app.AddLine("Goodbye!")

	default:
		m.app.AddLine(lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render(fmt.Sprintf("Unknown command: %s", command)))
		m.app.AddLine("Type 'help' to see available commands")
	}

	m.app.AddLine("")
}

func (m *cliModel) showHelp() {
	m.app.AddLine("Available commands:")
	m.app.AddLine("  help              Show this help message")
	m.app.AddLine("  tools             List available tools")
	m.app.AddLine("  components        List TUI components")
	m.app.AddLine("  init <name>       Initialize a new project")
	m.app.AddLine("  clear             Clear the screen")
	m.app.AddLine("  exit              Exit the CLI")
}

func (m *cliModel) showTools() {
	m.app.AddLine("Available Tools (11):")
	m.app.AddLine("")
	m.app.AddLine("File Operations:")
	m.app.AddLine("  • read      - Read files with format detection")
	m.app.AddLine("  • write     - Write files with backup support")
	m.app.AddLine("  • delete    - Delete files/directories")
	m.app.AddLine("  • list      - List directory contents")
	m.app.AddLine("  • glob      - Pattern matching with ** support")
	m.app.AddLine("")
	m.app.AddLine("Web Access:")
	m.app.AddLine("  • fetch     - HTTP requests with auth")
	m.app.AddLine("  • download  - Download files with resume")
	m.app.AddLine("")
	m.app.AddLine("Search:")
	m.app.AddLine("  • search    - Regex content search")
	m.app.AddLine("  • grep      - Advanced pattern matching")
	m.app.AddLine("")
	m.app.AddLine("System:")
	m.app.AddLine("  • shell     - Execute shell commands")
	m.app.AddLine("  • tasklist  - Manage task lists")
}

func (m *cliModel) showComponents() {
	m.app.AddLine("Available TUI Components:")
	m.app.AddLine("")
	m.app.AddLine("  Application   Complete app framework with layout")
	m.app.AddLine("  Panel         Configurable content areas with borders")
	m.app.AddLine("  Footer        Status bar component")
	m.app.AddLine("  Layout        Vertical/horizontal containers")
	m.app.AddLine("  Prompt        Interactive input with history")
	m.app.AddLine("")
	m.app.AddLine("Input Components:")
	m.app.AddLine("  • Text        - Single/multi-line text input")
	m.app.AddLine("  • YesNo       - Yes/No questions")
	m.app.AddLine("  • Options     - Single selection from list")
	m.app.AddLine("  • MultiSelect - Multiple selections")
	m.app.AddLine("")
	m.app.AddLine("Form Components:")
	m.app.AddLine("  • IterativeForm - Q&A sessions")
}

func (m cliModel) View() string {
	// Calculate panel height with margins: 1 top + panel + 1 bottom + footer
	topMargin := 1
	bottomMargin := 1
	footerHeight := 1
	panelHeight := m.height - topMargin - bottomMargin - footerHeight

	// Update panel size
	panel := m.app.GetPanel()
	panel.SetSize(m.width, panelHeight)

	// Get content lines
	contentLines := m.app.GetContent()

	// Calculate available lines for content (leave room for prompt)
	promptViewLines := strings.Count(m.prompt.View(), "\n") + 1
	separatorLines := 1
	availableContentLines := panel.GetContentHeight() - promptViewLines - separatorLines

	// Get visible content (auto-scroll to bottom)
	start := 0
	if len(contentLines) > availableContentLines {
		start = len(contentLines) - availableContentLines
	}
	visibleContent := contentLines[start:]

	// Build panel content: content + separator + prompt
	var panelContent strings.Builder
	panelContent.WriteString(strings.Join(visibleContent, "\n"))
	panelContent.WriteString("\n")
	panelContent.WriteString(strings.Repeat("─", panel.GetContentWidth()))
	panelContent.WriteString("\n")
	panelContent.WriteString(m.prompt.View())

	panel.SetContent(panelContent.String())

	// Render panel (trim leading newlines from margin)
	panelRendered := panel.Render()
	panelRendered = strings.TrimLeft(panelRendered, "\n")

	// Build output
	var output strings.Builder

	// Top margin (1 line)
	output.WriteString("\n")

	// Panel
	output.WriteString(panelRendered)

	// Count total lines so far (top margin + panel)
	currentOutput := output.String()
	currentLines := strings.Count(currentOutput, "\n") + 1

	// Calculate how many newlines needed to reach m.height
	// We want footer on line m.height, so we need (m.height - currentLines - 1) newlines
	fillLines := m.height - currentLines - 1

	// Add fill lines (this includes the bottom margin)
	if fillLines > 0 {
		output.WriteString(strings.Repeat("\n", fillLines))
	}

	// Add final newline before footer
	output.WriteString("\n")

	// Footer (on line m.height)
	output.WriteString(m.app.GetFooter().View())

	return output.String()
}
