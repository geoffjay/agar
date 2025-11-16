package main

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/geoffjay/agar/tui"
)

// AppMode represents different application modes
type AppMode string

const (
	NormalMode AppMode = "NORMAL"
	EditMode   AppMode = "EDIT"
	InsertMode AppMode = "INSERT"
	VisualMode AppMode = "VISUAL"
)

// demoModel wraps the footer with a simple application
type demoModel struct {
	footer      tui.FooterModel
	mode        AppMode
	content     string
	height      int
	directories []string
	dirIndex    int
}

// ContentStyle defines the styling for the main content area
var ContentStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("252")).
	Padding(1, 2)

func initialModel() demoModel {
	cwd, err := os.Getwd()
	if err != nil {
		cwd = "/unknown"
	}

	directories := []string{
		cwd,
		"/home/user/projects",
		"/var/log/application",
		"/usr/local/bin",
	}

	footer := tui.NewFooter("MyApp", "1.2.3", directories[0], string(NormalMode))

	return demoModel{
		footer:      footer,
		mode:        NormalMode,
		content:     "Press keys to see the footer update dynamically!",
		height:      24,
		directories: directories,
		dirIndex:    0,
	}
}

func (m demoModel) Init() tea.Cmd {
	return nil
}

func (m demoModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.height = msg.Height
		// Forward the message to the footer
		updatedFooter, cmd := m.footer.Update(msg)
		m.footer = updatedFooter.(tui.FooterModel)
		return m, cmd

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "n":
			m.mode = NormalMode
			m.footer.SetMode(string(m.mode))
			m.content = "Switched to NORMAL mode"

		case "e":
			m.mode = EditMode
			m.footer.SetMode(string(m.mode))
			m.content = "Switched to EDIT mode"

		case "i":
			m.mode = InsertMode
			m.footer.SetMode(string(m.mode))
			m.content = "Switched to INSERT mode"

		case "v":
			m.mode = VisualMode
			m.footer.SetMode(string(m.mode))
			m.content = "Switched to VISUAL mode"

		case "d":
			// Cycle through directories
			m.dirIndex = (m.dirIndex + 1) % len(m.directories)
			m.footer.SetDirectory(m.directories[m.dirIndex])
			m.content = fmt.Sprintf("Changed directory to: %s", m.directories[m.dirIndex])

		case "t":
			// Toggle title
			if m.footer.GetTitle() == "MyApp" {
				m.footer.SetTitle("AgarDemo")
				m.content = "Changed title to: AgarDemo"
			} else {
				m.footer.SetTitle("MyApp")
				m.content = "Changed title to: MyApp"
			}

		case "u":
			// Toggle version
			if m.footer.GetVersion() == "1.2.3" {
				m.footer.SetVersion("2.0.0")
				m.content = "Updated version to: 2.0.0"
			} else {
				m.footer.SetVersion("1.2.3")
				m.content = "Updated version to: 1.2.3"
			}
		}
	}

	return m, nil
}

func (m demoModel) View() string {
	var b strings.Builder

	// Header
	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		Render("Footer Component Demo")
	b.WriteString(header)
	b.WriteString("\n\n")

	// Content area
	contentArea := ContentStyle.Render(m.content)
	b.WriteString(contentArea)
	b.WriteString("\n\n")

	// Help text
	helpText := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Render("Keys: [n] Normal • [e] Edit • [i] Insert • [v] Visual • [d] Cycle Directory • [t] Toggle Title • [u] Toggle Version • [q] Quit")
	b.WriteString(helpText)

	// Calculate how much space to fill before the footer
	// Count current lines (newlines + 1 for the first line)
	currentLines := strings.Count(b.String(), "\n") + 1
	// We need to fill enough newlines to push the footer to the very bottom
	// Total height - current lines - 1 (for the footer line itself)
	fillLines := m.height - currentLines - 1

	if fillLines > 0 {
		b.WriteString(strings.Repeat("\n", fillLines))
	}

	// Footer (always at the bottom)
	b.WriteString("\n")
	b.WriteString(m.footer.View())

	return b.String()
}

func main() {
	fmt.Println("=== Footer Component Example ===\n")
	fmt.Println("This example demonstrates the footer component that displays:")
	fmt.Println("  - Application title and version (left)")
	fmt.Println("  - Current directory (center)")
	fmt.Println("  - Current mode (right)")
	fmt.Println("\nThe footer is always 1 line tall and spans the full terminal width.")
	fmt.Println("\nPress Enter to start the demo...")
	fmt.Scanln()

	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n=== Demo completed! ===")
}
