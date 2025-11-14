package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// IterationConfig controls the iterative Q&A behavior
type IterationConfig struct {
	MaxIterations    int
	IterationPrompt  string // What to ask between iterations
	CompletionPrompt string // How to ask if they're done
}

// IterativeQuestion represents a question in an iterative session
type IterativeQuestion struct {
	Question  string
	HelpText  string
	Iteration int // Which iteration this question is from
}

// IterativeFormModel represents an iterative Q&A form
type IterativeFormModel struct {
	title      string
	config     IterationConfig
	questions  []IterativeQuestion
	responses  []string
	currentIdx int
	iteration  int
	input      string
	err        error
	done       bool
	needsMore  bool // Whether user wants another iteration
	askingMore bool // Whether we're asking if they want more
	width      int  // Terminal width for text wrapping
}

// NewIterativeForm creates a new iterative form
func NewIterativeForm(title string, config IterationConfig) IterativeFormModel {
	return IterativeFormModel{
		title:      title,
		config:     config,
		questions:  []IterativeQuestion{},
		responses:  []string{},
		currentIdx: 0,
		iteration:  0,
		input:      "",
		done:       false,
		needsMore:  false,
		askingMore: false,
		width:      80, // Default width, will be updated by WindowSizeMsg
	}
}

// AddQuestions adds questions from a new iteration
func (m *IterativeFormModel) AddQuestions(questions []IterativeQuestion) {
	m.questions = append(m.questions, questions...)
}

// Init initializes the model
func (m IterativeFormModel) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m IterativeFormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.done = true
			return m, tea.Quit

		case "enter":
			return m.handleEnter()

		case "backspace":
			if len(m.input) > 0 {
				m.input = m.input[:len(m.input)-1]
			}

		default:
			m.input += msg.String()
		}
	}

	return m, nil
}

// handleEnter processes the enter key press
func (m IterativeFormModel) handleEnter() (tea.Model, tea.Cmd) {
	input := strings.TrimSpace(m.input)

	// Handle "do you need more info?" question
	if m.askingMore {
		lower := strings.ToLower(input)
		switch lower {
		case "yes", "y":
			m.needsMore = true
			m.askingMore = false
			m.done = true // Signal to caller to add more questions
			return m, tea.Quit
		case "no", "n":
			m.needsMore = false
			m.askingMore = false
			m.done = true
			return m, tea.Quit
		default:
			m.err = fmt.Errorf("please answer 'yes' or 'no'")
			m.input = ""
			return m, nil
		}
	}

	// Validate input
	if input == "" {
		m.err = fmt.Errorf("please provide an answer")
		return m, nil
	}

	// Store response
	m.responses = append(m.responses, input)
	m.err = nil
	m.input = ""
	m.currentIdx++

	// Check if we've answered all questions in current iteration
	if m.currentIdx >= len(m.questions) {
		// Check if we can do another iteration
		if m.iteration < m.config.MaxIterations-1 {
			m.askingMore = true
		} else {
			m.done = true
			return m, tea.Quit
		}
	}

	return m, nil
}

// View renders the UI
func (m IterativeFormModel) View() string {
	if m.done && !m.askingMore {
		if m.needsMore {
			return SuccessStyle.Render("✓ Gathering more information...\n")
		}
		return SuccessStyle.Render("✓ Information gathering complete!\n")
	}

	var b strings.Builder

	// Title
	b.WriteString(TitleStyle.Render(m.title))
	b.WriteString("\n\n")

	// Iteration indicator
	if m.config.MaxIterations > 1 {
		b.WriteString(HelpStyle.Render(fmt.Sprintf("Iteration %d of %d", m.iteration+1, m.config.MaxIterations)))
		b.WriteString("\n\n")
	}

	// If asking about more information
	if m.askingMore {
		b.WriteString(QuestionStyle.Render(m.config.CompletionPrompt))
		b.WriteString("\n")
		b.WriteString(HelpStyle.Render("(yes/no)"))
		b.WriteString("\n\n")
		b.WriteString(renderWrappedInput(m.input, m.width))
		b.WriteString("\n\n")

		if m.err != nil {
			b.WriteString(ErrorStyle.Render(fmt.Sprintf("⚠ %s", m.err.Error())))
			b.WriteString("\n\n")
		}

		// Show what we've gathered so far
		if len(m.responses) > 0 {
			b.WriteString("─────────────────────────────────\n")
			b.WriteString("Information gathered:\n")
			for i, resp := range m.responses {
				display := resp
				if len(display) > 60 {
					display = display[:60] + "..."
				}
				b.WriteString(fmt.Sprintf("%d. %s\n", i+1, display))
			}
		}
	} else if m.currentIdx < len(m.questions) {
		question := m.questions[m.currentIdx]

		// Progress
		b.WriteString(fmt.Sprintf("Question %d of %d (this iteration)\n\n", m.currentIdx+1-countQuestionsBeforeIteration(m.questions, m.iteration), countQuestionsInIteration(m.questions, m.iteration)))

		// Question text
		b.WriteString(QuestionStyle.Render(question.Question))
		b.WriteString("\n")

		// Help text
		if question.HelpText != "" {
			b.WriteString(HelpStyle.Render(question.HelpText))
			b.WriteString("\n")
		}

		b.WriteString("\n")

		// Input - use wrapped rendering
		b.WriteString(renderWrappedInput(m.input, m.width))
		b.WriteString("\n\n")

		// Error
		if m.err != nil {
			b.WriteString(ErrorStyle.Render(fmt.Sprintf("⚠ %s", m.err.Error())))
			b.WriteString("\n\n")
		}

		// Previous responses in this iteration
		startIdx := countQuestionsBeforeIteration(m.questions, m.iteration)
		if m.currentIdx > startIdx {
			b.WriteString("─────────────────────────────────\n")
			b.WriteString("Previous answers (this iteration):\n")
			for i := startIdx; i < m.currentIdx; i++ {
				resp := m.responses[i]
				if len(resp) > 50 {
					resp = resp[:50] + "..."
				}
				b.WriteString(fmt.Sprintf("%d. %s\n", i-startIdx+1, resp))
			}
		}
	}

	b.WriteString("\n")
	b.WriteString(HelpStyle.Render("Press Esc to cancel"))

	return b.String()
}

// GetResponses returns all collected responses
func (m IterativeFormModel) GetResponses() []string {
	return m.responses
}

// GetResponsesForIteration returns responses for a specific iteration
func (m IterativeFormModel) GetResponsesForIteration(iteration int) []string {
	var responses []string
	for i, q := range m.questions {
		if q.Iteration == iteration && i < len(m.responses) {
			responses = append(responses, m.responses[i])
		}
	}
	return responses
}

// IsDone returns whether the form is complete
func (m IterativeFormModel) IsDone() bool {
	return m.done && !m.needsMore
}

// NeedsMoreInfo returns whether user wants another iteration
func (m IterativeFormModel) NeedsMoreInfo() bool {
	return m.needsMore
}

// NextIteration prepares for the next iteration
func (m *IterativeFormModel) NextIteration() {
	m.iteration++
	m.askingMore = false
	m.needsMore = false
	m.done = false
}

// GetCurrentIteration returns the current iteration number
func (m IterativeFormModel) GetCurrentIteration() int {
	return m.iteration
}

// Helper functions

func countQuestionsBeforeIteration(questions []IterativeQuestion, iteration int) int {
	count := 0
	for _, q := range questions {
		if q.Iteration < iteration {
			count++
		}
	}
	return count
}

func countQuestionsInIteration(questions []IterativeQuestion, iteration int) int {
	count := 0
	for _, q := range questions {
		if q.Iteration == iteration {
			count++
		}
	}
	return count
}

// renderWrappedInput renders input text with proper wrapping
func renderWrappedInput(input string, termWidth int) string {
	if len(input) == 0 {
		return InputStyle.Render("> █")
	}

	maxWidth := max(termWidth-6, 40)

	// Split by user newlines
	userLines := strings.Split(input, "\n")
	var allLines []string

	for _, line := range userLines {
		if len(line) == 0 {
			allLines = append(allLines, "")
			continue
		}

		// Wrap long lines
		wrapped := wrapLineAtWords(line, maxWidth)
		allLines = append(allLines, wrapped...)
	}

	// Render with prefix and indentation
	var b strings.Builder
	for i, line := range allLines {
		if i == 0 {
			b.WriteString(InputStyle.Render("> " + line))
		} else {
			b.WriteString(InputStyle.Render("  " + line))
		}

		if i < len(allLines)-1 {
			b.WriteString("\n")
		}
	}

	b.WriteString(InputStyle.Render("█"))
	return b.String()
}

// wrapLineAtWords wraps a line at word boundaries
func wrapLineAtWords(line string, maxWidth int) []string {
	if len(line) <= maxWidth {
		return []string{line}
	}

	var wrapped []string
	remaining := line

	for len(remaining) > maxWidth {
		splitPos := -1
		for i := min(maxWidth-1, len(remaining)-1); i >= 0; i-- {
			if remaining[i] == ' ' {
				splitPos = i
				break
			}
		}

		if splitPos == -1 {
			splitPos = min(maxWidth, len(remaining))
		}

		wrapped = append(wrapped, remaining[:splitPos])
		remaining = strings.TrimLeft(remaining[splitPos:], " ")
	}

	if len(remaining) > 0 {
		wrapped = append(wrapped, remaining)
	}

	return wrapped
}
