package tui

import (
	"context"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/geoffjay/agar/commands"
)

// Application represents a complete TUI application with layout, panel, and footer
type Application struct {
	title          string
	version        string
	mode           string
	directory      string
	panel          *Panel
	footer         FooterModel
	layout         *Layout
	content        []string
	width          int
	height         int
	scrollOffset   int
	commandManager *commands.Manager
	metadata       map[string]interface{}
	shouldExit     bool
	ctx            context.Context
}

// ApplicationConfig holds configuration for creating a new Application
type ApplicationConfig struct {
	Title            string
	Version          string
	Mode             string
	Directory        string
	PanelMargin      int
	PanelPadding     int
	BorderStyle      BorderStyle
	CommandPaths     []string // Optional custom command paths
	EnableCommands   bool     // Enable slash command system (default: true)
}

// NewApplication creates a new application with the specified configuration
func NewApplication(config ApplicationConfig) *Application {
	// Create the panel for main content
	panel := NewPanel(config.PanelMargin, config.PanelPadding, config.BorderStyle)

	// Create the footer
	footer := NewFooter(config.Title, config.Version, config.Directory, config.Mode)

	// Create the layout (vertical)
	layout := NewLayout(Vertical)

	// Initialize command system if enabled (default: true)
	var cmdManager *commands.Manager
	if config.EnableCommands {
		if len(config.CommandPaths) > 0 {
			cmdManager = commands.NewManagerWithPaths(config.CommandPaths...)
		} else {
			cmdManager = commands.NewManager()
		}
		// Initialize command system (load built-in and file-based commands)
		// Errors are ignored as the command system will still work with built-in commands
		_ = cmdManager.Initialize()
	}

	app := &Application{
		title:          config.Title,
		version:        config.Version,
		mode:           config.Mode,
		directory:      config.Directory,
		panel:          panel,
		footer:         footer,
		layout:         layout,
		content:        make([]string, 0),
		width:          80,
		height:         24,
		scrollOffset:   0,
		commandManager: cmdManager,
		metadata:       make(map[string]interface{}),
		shouldExit:     false,
		ctx:            context.Background(),
	}

	return app
}

// Init initializes the application (implements tea.Model)
func (a *Application) Init() tea.Cmd {
	return nil
}

// Update handles messages (implements tea.Model)
func (a *Application) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Check if we should exit
	if a.shouldExit {
		return a, tea.Quit
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height

		// Update footer size
		updatedFooter, _ := a.footer.Update(msg)
		a.footer = updatedFooter.(FooterModel)

		// Update layout size
		a.layout.SetSize(a.width, a.height-1) // -1 for footer

		// Update panel size (take full height minus footer)
		a.panel.SetSize(a.width, a.height-1)

		// Update content display
		a.updatePanelContent()

		return a, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return a, tea.Quit
		}

	case CommandMsg:
		// Handle command execution
		if a.commandManager != nil {
			if err := a.commandManager.Handle(a.ctx, msg.Input, a); err != nil {
				a.AddLine("Error: " + err.Error())
			}
		}
		return a, nil
	}

	return a, nil
}

// View renders the application (implements tea.Model)
func (a *Application) View() string {
	var b strings.Builder

	// Calculate available height for content panel
	contentHeight := a.height - 1 // -1 for footer

	// Update panel size
	a.panel.SetSize(a.width, contentHeight)

	// Render the panel and remove any leading newlines from top margin
	panelOutput := a.panel.Render()
	panelOutput = strings.TrimLeft(panelOutput, "\n")

	// Count actual lines in panel output
	panelLines := strings.Count(panelOutput, "\n") + 1

	// Add the panel output
	b.WriteString(panelOutput)

	// Calculate how many lines we need to fill before the footer
	// We want: panelLines + fillLines + 1 (footer) = a.height
	fillLines := a.height - panelLines - 1

	// Add fill lines to push footer to bottom
	if fillLines > 0 {
		b.WriteString(strings.Repeat("\n", fillLines))
	}

	// Add newline before footer (this is the separator)
	b.WriteString("\n")

	// Render the footer
	b.WriteString(a.footer.View())

	return b.String()
}

// AddLine adds a single line to the content and auto-scrolls
func (a *Application) AddLine(line string) {
	a.content = append(a.content, line)
	a.updatePanelContent()
}

// AddLines adds multiple lines to the content and auto-scrolls
func (a *Application) AddLines(lines []string) {
	a.content = append(a.content, lines...)
	a.updatePanelContent()
}

// Clear clears all content
func (a *Application) Clear() {
	a.content = make([]string, 0)
	a.scrollOffset = 0
	a.updatePanelContent()
}

// SetContent replaces all content with the provided lines
func (a *Application) SetContent(lines []string) {
	a.content = lines
	a.updatePanelContent()
}

// GetContent returns all content lines
func (a *Application) GetContent() []string {
	return a.content
}

// SetTitle updates the application title
func (a *Application) SetTitle(title string) {
	a.title = title
	a.footer.SetTitle(title)
}

// SetVersion updates the application version
func (a *Application) SetVersion(version string) {
	a.version = version
	a.footer.SetVersion(version)
}

// SetMode updates the application mode
func (a *Application) SetMode(mode string) {
	a.mode = mode
	a.footer.SetMode(mode)
}

// SetDirectory updates the current directory
func (a *Application) SetDirectory(directory string) {
	a.directory = directory
	a.footer.SetDirectory(directory)
}

// GetPanel returns the panel for direct manipulation if needed
func (a *Application) GetPanel() *Panel {
	return a.panel
}

// GetFooter returns the footer for direct manipulation if needed
func (a *Application) GetFooter() *FooterModel {
	return &a.footer
}

// updatePanelContent updates the panel content with proper scrolling
func (a *Application) updatePanelContent() {
	contentHeight := a.panel.GetContentHeight()

	// Auto-scroll to bottom
	totalLines := len(a.content)
	if totalLines > contentHeight {
		a.scrollOffset = totalLines - contentHeight
	} else {
		a.scrollOffset = 0
	}

	// Get the visible lines
	visibleLines := a.getVisibleLines(contentHeight)

	// Update panel content
	a.panel.SetContent(strings.Join(visibleLines, "\n"))
}

// getVisibleLines returns the lines that should be visible based on scroll offset
func (a *Application) getVisibleLines(height int) []string {
	if len(a.content) == 0 {
		return []string{}
	}

	start := a.scrollOffset
	end := a.scrollOffset + height

	if start < 0 {
		start = 0
	}
	if end > len(a.content) {
		end = len(a.content)
	}
	if start > len(a.content) {
		start = len(a.content)
	}

	return a.content[start:end]
}

// ApplicationState interface implementation

// GetMode returns the current application mode
func (a *Application) GetMode() string {
	return a.mode
}

// GetMetadata returns application metadata
func (a *Application) GetMetadata() map[string]interface{} {
	return a.metadata
}

// SetMetadata sets application metadata
func (a *Application) SetMetadata(key string, value interface{}) {
	a.metadata[key] = value
}

// Exit signals the application to exit
func (a *Application) Exit() {
	a.shouldExit = true
}

// Command system helper methods

// RegisterCommand registers a custom command
func (a *Application) RegisterCommand(cmd commands.Command) error {
	if a.commandManager == nil {
		return nil // Commands not enabled
	}
	return a.commandManager.RegisterCommand(cmd)
}

// UnregisterCommand removes a command
func (a *Application) UnregisterCommand(name string) error {
	if a.commandManager == nil {
		return nil // Commands not enabled
	}
	return a.commandManager.UnregisterCommand(name)
}

// GetCommandManager returns the command manager for advanced usage
func (a *Application) GetCommandManager() *commands.Manager {
	return a.commandManager
}

// CommandMsg is a message for executing a slash command
type CommandMsg struct {
	Input string
}

// NewCommandMsg creates a new command message
func NewCommandMsg(input string) CommandMsg {
	return CommandMsg{Input: input}
}
