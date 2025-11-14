package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// TextInputType represents the type of text input
type TextInputType int

const (
	SingleLine TextInputType = iota
	MultiLine
)

// TextModel represents a text input component
type TextModel struct {
	prompt    string
	helpText  string
	inputType TextInputType
	input     string
	width     int
	done      bool
}

// NewTextInput creates a new text input component
func NewTextInput(prompt, helpText string, inputType TextInputType) TextModel {
	return TextModel{
		prompt:    prompt,
		helpText:  helpText,
		inputType: inputType,
		input:     "",
		width:     80,
		done:      false,
	}
}

// Init initializes the component
func (m TextModel) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m TextModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.done {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit

		case "enter":
			if m.inputType == SingleLine {
				// Single-line: Enter submits
				if strings.TrimSpace(m.input) != "" {
					m.done = true
					return m, tea.Quit
				}
			} else {
				// Multi-line: Enter adds newline
				m.input += "\n"
			}

		case "ctrl+d":
			// Ctrl+D submits for multi-line
			if m.inputType == MultiLine && strings.TrimSpace(m.input) != "" {
				m.done = true
				return m, tea.Quit
			}

		case "backspace":
			if len(m.input) > 0 {
				m.input = m.input[:len(m.input)-1]
			}

		default:
			// Add character to input
			if len(msg.String()) == 1 {
				m.input += msg.String()
			}
		}
	}

	return m, nil
}

// View renders the component
func (m TextModel) View() string {
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

	// Input field
	b.WriteString(renderWrappedInput(m.input, m.width))
	b.WriteString("\n\n")

	b.WriteString(HelpStyle.Render("───────────────────────────────────"))
	b.WriteString("\n\n")

	// Help message
	if m.inputType == SingleLine {
		b.WriteString(HelpStyle.Render("Enter to submit • Esc to cancel"))
	} else {
		b.WriteString(HelpStyle.Render("Enter for new line • Ctrl+D to submit • Esc to cancel"))
	}

	return b.String()
}

// GetAnswer returns the input text
func (m TextModel) GetAnswer() string {
	return strings.TrimSpace(m.input)
}

// IsDone returns whether the question has been answered
func (m TextModel) IsDone() bool {
	return m.done
}
