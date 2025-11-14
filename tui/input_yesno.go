package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// YesNoModel represents a yes/no input component
type YesNoModel struct {
	prompt   string
	helpText string
	selected bool // true = yes, false = no
	focused  int  // 0 = yes, 1 = no
	done     bool
}

// NewYesNoInput creates a new yes/no input component
func NewYesNoInput(prompt, helpText string) YesNoModel {
	return YesNoModel{
		prompt:   prompt,
		helpText: helpText,
		selected: true, // Default to yes
		focused:  0,
		done:     false,
	}
}

// Init initializes the component
func (m YesNoModel) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m YesNoModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.done {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit

		case "enter", " ":
			m.selected = (m.focused == 0)
			m.done = true
			return m, tea.Quit

		case "up", "k", "left", "h":
			m.focused = 0

		case "down", "j", "right", "l":
			m.focused = 1

		case "y", "Y":
			m.selected = true
			m.done = true
			return m, tea.Quit

		case "n", "N":
			m.selected = false
			m.done = true
			return m, tea.Quit
		}
	}

	return m, nil
}

// View renders the component
func (m YesNoModel) View() string {
	if m.done {
		return ""
	}

	var b strings.Builder

	// Prompt
	b.WriteString(QuestionStyle.Render(m.prompt))
	b.WriteString("\n")

	// Help text
	if m.helpText != "" {
		b.WriteString(HelpStyle.Render(m.helpText))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(HelpStyle.Render("───────────────────────────────────"))
	b.WriteString("\n\n")

	// Options
	yesStyle := lipgloss.NewStyle()
	noStyle := lipgloss.NewStyle()

	if m.focused == 0 {
		yesStyle = yesStyle.Foreground(lipgloss.Color("212")).Bold(true)
	}
	if m.focused == 1 {
		noStyle = noStyle.Foreground(lipgloss.Color("212")).Bold(true)
	}

	// Yes option
	if m.focused == 0 {
		b.WriteString(yesStyle.Render("[x] Yes"))
	} else {
		b.WriteString(yesStyle.Render("[ ] Yes"))
	}
	b.WriteString("\n")

	// No option
	if m.focused == 1 {
		b.WriteString(noStyle.Render("[x] No"))
	} else {
		b.WriteString(noStyle.Render("[ ] No"))
	}
	b.WriteString("\n\n")

	b.WriteString(HelpStyle.Render("───────────────────────────────────"))
	b.WriteString("\n\n")

	b.WriteString(HelpStyle.Render("↑/↓ or h/j/k/l to select • Enter/Space to confirm • y/n for quick answer • Esc to cancel"))

	return b.String()
}

// GetAnswer returns the selected answer
func (m YesNoModel) GetAnswer() bool {
	return m.selected
}

// GetAnswerString returns the answer as a string
func (m YesNoModel) GetAnswerString() string {
	if m.selected {
		return "yes"
	}
	return "no"
}

// IsDone returns whether the question has been answered
func (m YesNoModel) IsDone() bool {
	return m.done
}
