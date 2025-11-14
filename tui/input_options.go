package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// OptionsModel represents a multiple choice input component
type OptionsModel struct {
	prompt   string
	helpText string
	options  []string
	selected int
	done     bool
}

// NewOptionsInput creates a new options input component
func NewOptionsInput(prompt, helpText string, options []string) OptionsModel {
	return OptionsModel{
		prompt:   prompt,
		helpText: helpText,
		options:  options,
		selected: 0,
		done:     false,
	}
}

// Init initializes the component
func (m OptionsModel) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m OptionsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.done {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit

		case "enter", " ":
			m.done = true
			return m, tea.Quit

		case "up", "k":
			if m.selected > 0 {
				m.selected--
			}

		case "down", "j":
			if m.selected < len(m.options)-1 {
				m.selected++
			}

		case "1", "2", "3", "4", "5", "6", "7", "8", "9":
			// Allow number shortcuts for options 1-9
			num := int(msg.String()[0] - '0')
			if num > 0 && num <= len(m.options) {
				m.selected = num - 1
				m.done = true
				return m, tea.Quit
			}
		}
	}

	return m, nil
}

// View renders the component
func (m OptionsModel) View() string {
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
	for i, option := range m.options {
		optionStyle := lipgloss.NewStyle()

		if i == m.selected {
			optionStyle = optionStyle.Foreground(lipgloss.Color("212")).Bold(true)
			b.WriteString(optionStyle.Render(fmt.Sprintf("[x] %s", option)))
		} else {
			b.WriteString(optionStyle.Render(fmt.Sprintf("[ ] %s", option)))
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(HelpStyle.Render("───────────────────────────────────"))
	b.WriteString("\n\n")

	// Help message
	helpMsg := "↑/↓ or j/k to select • Enter/Space to confirm"
	if len(m.options) <= 9 {
		helpMsg += " • 1-9 for quick select"
	}
	helpMsg += " • Esc to cancel"
	b.WriteString(HelpStyle.Render(helpMsg))

	return b.String()
}

// GetAnswer returns the selected option text
func (m OptionsModel) GetAnswer() string {
	if m.selected >= 0 && m.selected < len(m.options) {
		return m.options[m.selected]
	}
	return ""
}

// GetSelectedIndex returns the index of the selected option
func (m OptionsModel) GetSelectedIndex() int {
	return m.selected
}

// IsDone returns whether the question has been answered
func (m OptionsModel) IsDone() bool {
	return m.done
}
