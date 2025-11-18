package app

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/geoffjay/agar/tools"
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

	// Create tool registry and register built-in tools
	toolRegistry := tools.NewToolRegistry()
	toolRegistry.Register(tools.NewReadTool())
	toolRegistry.Register(tools.NewWriteTool())
	toolRegistry.Register(tools.NewDeleteTool())
	toolRegistry.Register(tools.NewListTool())
	toolRegistry.Register(tools.NewGlobTool())
	toolRegistry.Register(tools.NewFetchTool())
	toolRegistry.Register(tools.NewDownloadTool())
	toolRegistry.Register(tools.NewSearchTool())
	toolRegistry.Register(tools.NewGrepTool())
	toolRegistry.Register(tools.NewShellTool())
	toolRegistry.Register(tools.NewTaskListTool())

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
		ToolRegistry:   toolRegistry,
	})

	// Register CLI-specific commands
	initCmd := NewInitCommand()
	if err := app.RegisterCommand(initCmd); err != nil {
		fmt.Printf("Warning: failed to register /init command: %v\n", err)
	}

	// Add welcome content
	app.AddLine("")
	app.AddLine("                      WELCOME TO AGAR CLI")
	app.AddLine("")
	app.AddLine("Agar is a comprehensive framework for building AI agent applications")
	app.AddLine("with TUI components and tool management.")
	app.AddLine("")
	app.AddLine("─────────────────────────────────────────────────────────────────────────────────────────")
	app.AddLine("")
	app.AddLine("GETTING STARTED")
	app.AddLine("")
	app.AddLine("  Type /help to see all available slash commands")
	app.AddLine("  Type /tools to see all available tools")
	app.AddLine("  Type /init <name> to create a new Agar project")
	app.AddLine("")
	app.AddLine("  Enter any text without a leading / to submit it as a prompt")
	app.AddLine("")
	app.AddLine("─────────────────────────────────────────────────────────────────────────────────────────")
	app.AddLine("")

	// Create prompt (no placeholder for cleaner look)
	cmdMgr := app.GetCommandManager()
	if cmdMgr != nil {
		app.AddLine(fmt.Sprintf("✓ Loaded %d slash commands and %d tools", len(cmdMgr.ListCommands()), toolRegistry.Count()))
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

		// Skip empty input
		if input == "" {
			updated, cmd := m.prompt.Update(msg)
			m.prompt = updated.(tui.PromptModel)
			return m, cmd
		}

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

		// Non-slash input: submit as prompt to AI agent
		m.app.AddLine("")
		m.app.AddLine(lipgloss.NewStyle().Foreground(lipgloss.Color("86")).Render("> " + input))
		m.app.AddLine("")

		// TODO: Submit to BAML/AI agent - for now just echo back
		m.app.AddLine("AI response for: " + input)
		m.app.AddLine("(Note: BAML integration pending)")
		m.app.AddLine("")

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
