package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// FooterModel represents a footer component that spans the full width of the terminal
type FooterModel struct {
	title     string
	version   string
	directory string
	mode      string
	width     int
}

// FooterStyle defines the styling for the footer
var FooterStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("241")).
	Background(lipgloss.Color("235"))

// NewFooter creates a new footer component
func NewFooter(title, version, directory, mode string) FooterModel {
	return FooterModel{
		title:     title,
		version:   version,
		directory: directory,
		mode:      mode,
		width:     80, // Default width, will be updated on WindowSizeMsg
	}
}

// Init initializes the footer component
func (m FooterModel) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m FooterModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		return m, nil
	}

	return m, nil
}

// View renders the footer component
func (m FooterModel) View() string {
	// Combine title and version for left side
	leftSide := m.title
	if m.version != "" {
		leftSide += " v" + m.version
	}

	// Right side is the mode
	rightSide := m.mode

	// Calculate available space for the directory in the center
	leftLen := lipgloss.Width(leftSide)
	rightLen := lipgloss.Width(rightSide)

	// Add spacing (1 space on each side of the directory)
	availableSpace := m.width - leftLen - rightLen - 2

	// Truncate or pad the directory to fit
	centerContent := m.directory
	if availableSpace > 0 {
		if len(centerContent) > availableSpace {
			// Truncate with ellipsis if too long
			if availableSpace > 3 {
				centerContent = "..." + centerContent[len(centerContent)-(availableSpace-3):]
			} else {
				centerContent = strings.Repeat(".", availableSpace)
			}
		} else {
			// Center the directory text
			padding := (availableSpace - len(centerContent)) / 2
			centerContent = strings.Repeat(" ", padding) + centerContent + strings.Repeat(" ", availableSpace-padding-len(centerContent))
		}
	} else {
		centerContent = ""
	}

	// Build the footer line
	footerLine := leftSide + " " + centerContent + " " + rightSide

	// Ensure the line is exactly the width of the terminal
	if lipgloss.Width(footerLine) < m.width {
		footerLine += strings.Repeat(" ", m.width-lipgloss.Width(footerLine))
	} else if lipgloss.Width(footerLine) > m.width {
		footerLine = footerLine[:m.width]
	}

	return FooterStyle.Render(footerLine)
}

// SetTitle updates the footer title
func (m *FooterModel) SetTitle(title string) {
	m.title = title
}

// SetVersion updates the footer version
func (m *FooterModel) SetVersion(version string) {
	m.version = version
}

// SetDirectory updates the footer directory
func (m *FooterModel) SetDirectory(directory string) {
	m.directory = directory
}

// SetMode updates the footer mode
func (m *FooterModel) SetMode(mode string) {
	m.mode = mode
}

// GetTitle returns the current title
func (m FooterModel) GetTitle() string {
	return m.title
}

// GetVersion returns the current version
func (m FooterModel) GetVersion() string {
	return m.version
}

// GetDirectory returns the current directory
func (m FooterModel) GetDirectory() string {
	return m.directory
}

// GetMode returns the current mode
func (m FooterModel) GetMode() string {
	return m.mode
}
