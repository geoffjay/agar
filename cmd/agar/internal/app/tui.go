package app

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/geoffjay/agar/tui"
)

// RunTUI launches the TUI application
func RunTUI() error {
	cwd, err := os.Getwd()
	if err != nil {
		cwd = "/"
	}

	// Create TUI application
	app := tui.NewApplication(tui.ApplicationConfig{
		Title:        "Agar",
		Version:      "0.1.0",
		Mode:         "READY",
		Directory:    cwd,
		PanelMargin:  1,
		PanelPadding: 1,
		BorderStyle:  tui.RoundedBorder,
	})

	// Add welcome content
	app.AddLine("╔══════════════════════════════════════════════════════════════╗")
	app.AddLine("║                   Welcome to Agar CLI                       ║")
	app.AddLine("╚══════════════════════════════════════════════════════════════╝")
	app.AddLine("")
	app.AddLine("Agar is a comprehensive framework for building AI agent")
	app.AddLine("applications with TUI components and tool management.")
	app.AddLine("")
	app.AddLine("═══════════════════════════════════════════════════════════════")
	app.AddLine("")
	app.AddLine("Available Tools (11):")
	app.AddLine("  • File Operations: read, write, delete, list, glob")
	app.AddLine("  • Web Access: fetch, download")
	app.AddLine("  • Search: search, grep")
	app.AddLine("  • System: shell, tasklist")
	app.AddLine("")
	app.AddLine("TUI Components:")
	app.AddLine("  • Application - Complete app framework with panels")
	app.AddLine("  • Panel - Configurable content areas")
	app.AddLine("  • Footer - Status bar component")
	app.AddLine("  • Layout - Vertical/horizontal containers")
	app.AddLine("  • Input Components - Text, YesNo, Options, MultiSelect")
	app.AddLine("  • Iterative Forms - Q&A sessions")
	app.AddLine("")
	app.AddLine("═══════════════════════════════════════════════════════════════")
	app.AddLine("")
	app.AddLine("Quick Start:")
	app.AddLine("  1. Import the library:")
	app.AddLine("     import \"github.com/geoffjay/agar/tools\"")
	app.AddLine("     import \"github.com/geoffjay/agar/tui\"")
	app.AddLine("")
	app.AddLine("  2. Create a tool registry:")
	app.AddLine("     registry := tools.NewToolRegistry()")
	app.AddLine("     registry.Register(tools.NewReadTool())")
	app.AddLine("")
	app.AddLine("  3. Create a TUI application:")
	app.AddLine("     app := tui.NewApplication(tui.ApplicationConfig{...})")
	app.AddLine("")
	app.AddLine("═══════════════════════════════════════════════════════════════")
	app.AddLine("")
	app.AddLine("Documentation:")
	app.AddLine("  • Tools: docs/tools.md")
	app.AddLine("  • Repository: github.com/geoffjay/agar")
	app.AddLine("")
	app.AddLine("Press 'q' to quit")
	app.AddLine("")

	// Run the application
	p := tea.NewProgram(app, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("failed to run TUI: %w", err)
	}

	return nil
}
