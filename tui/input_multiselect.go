package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// MultiSelectModel represents a multiple selection input component
type MultiSelectModel struct {
	prompt   string
	helpText string
	options  []string
	selected []bool
	cursor   int
	done     bool
}

// NewMultiSelectInput creates a new multi-select input component
func NewMultiSelectInput(prompt, helpText string, options []string) MultiSelectModel {
	selected := make([]bool, len(options))
	for i := range selected {
		selected[i] = false
	}
	return MultiSelectModel{
		prompt:   prompt,
		helpText: helpText,
		options:  options,
		selected: selected,
		cursor:   0,
		done:     false,
	}
}

// Init initializes the component
func (m MultiSelectModel) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m MultiSelectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.done {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit

		case "enter":
			m.done = true
			return m, tea.Quit

		case " ":
			// Toggle selection
			m.selected[m.cursor] = !m.selected[m.cursor]

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.options)-1 {
				m.cursor++
			}

		case "1", "2", "3", "4", "5", "6", "7", "8", "9":
			// Allow number shortcuts for options 1-9
			num := int(msg.String()[0] - '0')
			if num > 0 && num <= len(m.options) {
				// Toggle selection
				idx := num - 1
				m.selected[idx] = !m.selected[idx]
				m.cursor = idx
			}
		}
	}

	return m, nil
}

// View renders the component
func (m MultiSelectModel) View() string {
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

		if i == m.cursor {
			optionStyle = optionStyle.Foreground(lipgloss.Color("212")).Bold(true)
		}

		// Show checkmark for selected items
		if m.selected[i] {
			b.WriteString(optionStyle.Render(fmt.Sprintf("[✔] %s", option)))
		} else {
			b.WriteString(optionStyle.Render(fmt.Sprintf("[ ] %s", option)))
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(HelpStyle.Render("───────────────────────────────────"))
	b.WriteString("\n\n")

	// Help message
	helpMsg := "↑/↓ or j/k to navigate • Space to toggle • Enter to confirm"
	if len(m.options) <= 9 {
		helpMsg += " • 1-9 to toggle specific options"
	}
	helpMsg += " • Esc to cancel"
	b.WriteString(HelpStyle.Render(helpMsg))

	return b.String()
}

// GetAnswers returns all selected option texts
func (m MultiSelectModel) GetAnswers() []string {
	var answers []string
	for i, selected := range m.selected {
		if selected {
			answers = append(answers, m.options[i])
		}
	}
	return answers
}

// GetSelectedIndices returns the indices of all selected options
func (m MultiSelectModel) GetSelectedIndices() []int {
	var indices []int
	for i, selected := range m.selected {
		if selected {
			indices = append(indices, i)
		}
	}
	return indices
}

// IsDone returns whether the question has been answered
func (m MultiSelectModel) IsDone() bool {
	return m.done
}
