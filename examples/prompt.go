package main

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/geoffjay/agar/tui"
)

// demoModel wraps the prompt with response display
type demoModel struct {
	prompt    tui.PromptModel
	responses []string
	mode      tui.PromptMode
}

func initialModel() demoModel {
	prompt := tui.NewPromptInput("> ", "Type a message...", tui.SingleLineMode).
		WithHistory(true)

	return demoModel{
		prompt:    prompt,
		responses: make([]string, 0),
		mode:      tui.SingleLineMode,
	}
}

func (m demoModel) Init() tea.Cmd {
	return nil
}

func (m demoModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "ctrl+q":
			return m, tea.Quit

		case "ctrl+t":
			// Toggle between single-line and multi-line
			if m.mode == tui.SingleLineMode {
				m.mode = tui.MultiLineMode
				m.prompt = tui.NewPromptInput(">> ", "Type a multi-line message...", tui.MultiLineMode).
					WithHistory(true)
				m.responses = append(m.responses, "[Switched to multi-line mode]")
			} else {
				m.mode = tui.SingleLineMode
				m.prompt = tui.NewPromptInput("> ", "Type a message...", tui.SingleLineMode).
					WithHistory(true)
				m.responses = append(m.responses, "[Switched to single-line mode]")
			}
			return m, nil

		case "ctrl+l":
			// Clear responses
			m.responses = make([]string, 0)
			return m, nil

		case "ctrl+h":
			// Show history
			history := m.prompt.GetHistory()
			if len(history) == 0 {
				m.responses = append(m.responses, "[No history]")
			} else {
				m.responses = append(m.responses, fmt.Sprintf("[History: %d items]", len(history)))
				for i, item := range history {
					m.responses = append(m.responses, fmt.Sprintf("  %d: %s", i+1, item))
				}
			}
			return m, nil
		}

	case tui.PromptSubmitMsg:
		// Handle submitted input
		response := fmt.Sprintf("You said: %s", msg.Input)
		m.responses = append(m.responses, response)

		// Echo mode - just show what was typed
		// In a real app, you might process the input, call tools, etc.
		if strings.HasPrefix(msg.Input, "/") {
			// Simulate command handling
			cmd := strings.TrimPrefix(msg.Input, "/")
			m.responses = append(m.responses, fmt.Sprintf("[Command: %s]", cmd))
		}

		// Update the prompt model
		updated, cmd := m.prompt.Update(msg)
		m.prompt = updated.(tui.PromptModel)
		return m, cmd
	}

	// Forward other messages to prompt
	updated, cmd := m.prompt.Update(msg)
	m.prompt = updated.(tui.PromptModel)
	return m, cmd
}

func (m demoModel) View() string {
	var b strings.Builder

	// Header
	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		Render("Prompt Component Demo")
	b.WriteString(header)
	b.WriteString("\n\n")

	// Current mode
	modeText := "single-line"
	if m.mode == tui.MultiLineMode {
		modeText = "multi-line"
	}
	modeDisplay := lipgloss.NewStyle().
		Foreground(lipgloss.Color("86")).
		Render(fmt.Sprintf("Mode: %s", modeText))
	b.WriteString(modeDisplay)
	b.WriteString("\n\n")

	// Responses (last 10)
	responseStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252"))

	start := 0
	if len(m.responses) > 10 {
		start = len(m.responses) - 10
	}

	for _, response := range m.responses[start:] {
		b.WriteString(responseStyle.Render(response))
		b.WriteString("\n")
	}

	if len(m.responses) > 0 {
		b.WriteString("\n")
	}

	// Prompt
	b.WriteString(m.prompt.View())
	b.WriteString("\n\n")

	// Help
	helpText := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Render("Commands: [Ctrl+T] Toggle mode • [Ctrl+L] Clear • [Ctrl+H] Show history • [Ctrl+C] Quit")
	b.WriteString(helpText)

	return b.String()
}

func main() {
	fmt.Println("=== Prompt Component Example ===\n")
	fmt.Println("This example demonstrates the prompt input component with:")
	fmt.Println("  - Single-line and multi-line modes")
	fmt.Println("  - Input history (use ↑↓ arrows)")
	fmt.Println("  - Message-based submission")
	fmt.Println("  - Configurable prefix")
	fmt.Println()
	fmt.Println("Try typing messages and navigating history with arrow keys.")
	fmt.Println()
	fmt.Println("Press Enter to start...")
	fmt.Scanln()

	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n=== Demo completed! ===")
}
