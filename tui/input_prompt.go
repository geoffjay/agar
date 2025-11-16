package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// PromptMode represents the input mode for the prompt
type PromptMode int

const (
	SingleLineMode PromptMode = iota
	MultiLineMode
)

// PromptSubmitMsg is sent when the user submits input
type PromptSubmitMsg struct {
	Input string
}

// PromptModel represents a prompt input component with history
type PromptModel struct {
	prefix         string
	placeholder    string
	mode           PromptMode
	input          string
	history        []string
	historyIndex   int
	historyEnabled bool
	width          int
	onSubmit       func(string) // Optional callback
	done           bool
}

// NewPromptInput creates a new prompt input component
func NewPromptInput(prefix, placeholder string, mode PromptMode) PromptModel {
	return PromptModel{
		prefix:         prefix,
		placeholder:    placeholder,
		mode:           mode,
		input:          "",
		history:        make([]string, 0),
		historyIndex:   -1,
		historyEnabled: true,
		width:          80,
		onSubmit:       nil,
		done:           false,
	}
}

// WithHistory enables or disables history
func (m PromptModel) WithHistory(enabled bool) PromptModel {
	m.historyEnabled = enabled
	return m
}

// WithOnSubmit sets an optional callback function
func (m PromptModel) WithOnSubmit(callback func(string)) PromptModel {
	m.onSubmit = callback
	return m
}

// Init initializes the component
func (m PromptModel) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m PromptModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			if m.mode == SingleLineMode {
				// Single-line: Enter submits
				return m.submit()
			}
			// Multi-line: Enter adds newline
			m.input += "\n"

		case "ctrl+d":
			// Ctrl+D submits for multi-line
			if m.mode == MultiLineMode {
				return m.submit()
			}

		case "up":
			// Navigate history backwards (older)
			if m.historyEnabled && len(m.history) > 0 {
				if m.historyIndex == -1 {
					// First time pressing up, start at end of history
					m.historyIndex = len(m.history) - 1
					m.input = m.history[m.historyIndex]
				} else if m.historyIndex > 0 {
					m.historyIndex--
					m.input = m.history[m.historyIndex]
				}
			}

		case "down":
			// Navigate history forwards (newer)
			if m.historyEnabled && m.historyIndex != -1 {
				if m.historyIndex < len(m.history)-1 {
					m.historyIndex++
					m.input = m.history[m.historyIndex]
				} else {
					// Back to current (empty) input
					m.historyIndex = -1
					m.input = ""
				}
			}

		case "backspace":
			if len(m.input) > 0 {
				m.input = m.input[:len(m.input)-1]
			}

		case "ctrl+u":
			// Clear line
			m.input = ""

		case "ctrl+w":
			// Delete word
			m.input = deleteLastWord(m.input)

		default:
			// Add character to input
			if len(msg.String()) == 1 || msg.Type == tea.KeySpace || msg.Type == tea.KeyTab {
				if msg.Type == tea.KeySpace {
					m.input += " "
				} else if msg.Type == tea.KeyTab {
					m.input += "\t"
				} else {
					m.input += msg.String()
				}
			}
		}
	}

	return m, nil
}

// submit handles input submission
func (m PromptModel) submit() (tea.Model, tea.Cmd) {
	if strings.TrimSpace(m.input) == "" {
		return m, nil
	}

	input := m.input

	// Add to history
	if m.historyEnabled {
		m.history = append(m.history, input)
		m.historyIndex = -1
	}

	// Call callback if provided
	if m.onSubmit != nil {
		m.onSubmit(input)
	}

	// Clear input
	m.input = ""

	// Emit message for parent to handle
	return m, func() tea.Msg {
		return PromptSubmitMsg{Input: input}
	}
}

// View renders the component
func (m PromptModel) View() string {
	if m.done {
		return ""
	}

	var b strings.Builder

	// Show prompt prefix and input
	b.WriteString(m.prefix)

	if m.input == "" && m.placeholder != "" {
		// Show placeholder in dim style
		b.WriteString(HelpStyle.Render(m.placeholder))
	} else {
		b.WriteString(InputStyle.Render(m.input))
	}

	// Show cursor
	b.WriteString(InputStyle.Render("_"))

	b.WriteString("\n")

	// Help text based on mode
	if m.mode == SingleLineMode {
		b.WriteString(HelpStyle.Render("Enter to submit"))
	} else {
		b.WriteString(HelpStyle.Render("Enter for new line • Ctrl+D to submit"))
	}

	if m.historyEnabled && len(m.history) > 0 {
		b.WriteString(HelpStyle.Render(" • ↑↓ for history"))
	}

	b.WriteString(HelpStyle.Render(" • Ctrl+C to quit"))

	return b.String()
}

// GetInput returns the current input
func (m PromptModel) GetInput() string {
	return m.input
}

// GetHistory returns the input history
func (m PromptModel) GetHistory() []string {
	return m.history
}

// IsDone returns whether the prompt is done
func (m PromptModel) IsDone() bool {
	return m.done
}

// ClearHistory clears the input history
func (m *PromptModel) ClearHistory() {
	m.history = make([]string, 0)
	m.historyIndex = -1
}

// deleteLastWord removes the last word from the input
func deleteLastWord(input string) string {
	trimmed := strings.TrimRight(input, " \t")
	lastSpace := strings.LastIndexAny(trimmed, " \t")
	if lastSpace == -1 {
		return ""
	}
	return input[:lastSpace+1]
}
