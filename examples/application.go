package main

import (
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/geoffjay/agar/tui"
)

// tickMsg is sent on every tick
type tickMsg time.Time

// demoApp wraps the Application for demonstration
type demoApp struct {
	app              *tui.Application
	lineCounter      int
	autoAdd          bool
	borderStyleIndex int
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second*2, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func initialApp() demoApp {
	cwd, err := os.Getwd()
	if err != nil {
		cwd = "/unknown"
	}

	// Create application with configuration
	app := tui.NewApplication(tui.ApplicationConfig{
		Title:        "LogViewer",
		Version:      "1.0.0",
		Mode:         "RUNNING",
		Directory:    cwd,
		PanelMargin:  1,
		PanelPadding: 1,
		BorderStyle:  tui.RoundedBorder,
	})

	// Add initial content
	app.AddLine("Welcome to the Application Demo!")
	app.AddLine("This demonstrates the tui.Application component.")
	app.AddLine("")
	app.AddLine("The main content area will auto-scroll as new lines are added.")
	app.AddLine("The footer remains fixed at the bottom.")
	app.AddLine("")
	app.AddLine("Press 'a' to toggle auto-add mode")
	app.AddLine("Press 'l' to manually add a line")
	app.AddLine("Press 'c' to clear content")
	app.AddLine("Press 'm' to cycle through modes")
	app.AddLine("Press 'b' to cycle through border styles")
	app.AddLine("Press 'q' to quit")
	app.AddLine("")
	app.AddLine("═══════════════════════════════════════════════════════════")
	app.AddLine("")

	return demoApp{
		app:              app,
		lineCounter:      1,
		autoAdd:          false,
		borderStyleIndex: 0, // Start with RoundedBorder
	}
}

func (d demoApp) Init() tea.Cmd {
	return tea.Batch(d.app.Init(), tickCmd())
}

func (d demoApp) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Forward window size updates to the application
		updatedApp, cmd := d.app.Update(msg)
		d.app = updatedApp.(*tui.Application)
		return d, cmd

	case tickMsg:
		if d.autoAdd {
			d.app.AddLine(fmt.Sprintf("[%s] Auto-generated log line #%d", time.Now().Format("15:04:05"), d.lineCounter))
			d.lineCounter++
		}
		return d, tickCmd()

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return d, tea.Quit

		case "a":
			d.autoAdd = !d.autoAdd
			if d.autoAdd {
				d.app.SetMode("AUTO-ADD")
				d.app.AddLine("")
				d.app.AddLine("▶ Auto-add mode ENABLED - New lines will be added every 2 seconds")
				d.app.AddLine("")
			} else {
				d.app.SetMode("MANUAL")
				d.app.AddLine("")
				d.app.AddLine("▪ Auto-add mode DISABLED")
				d.app.AddLine("")
			}

		case "l":
			d.app.AddLine(fmt.Sprintf("[%s] Manually added line #%d", time.Now().Format("15:04:05"), d.lineCounter))
			d.lineCounter++

		case "c":
			d.app.Clear()
			d.app.AddLine("Content cleared!")
			d.app.AddLine("")
			d.lineCounter = 1

		case "m":
			// Cycle through modes
			modes := []string{"RUNNING", "PAUSED", "STOPPED", "DEBUG"}
			currentMode := d.app.GetFooter().GetMode()
			nextIndex := 0
			for i, mode := range modes {
				if mode == currentMode {
					nextIndex = (i + 1) % len(modes)
					break
				}
			}
			d.app.SetMode(modes[nextIndex])
			d.app.AddLine(fmt.Sprintf("Mode changed to: %s", modes[nextIndex]))

		case "b":
			// Cycle through border styles
			panel := d.app.GetPanel()
			styles := []tui.BorderStyle{tui.RoundedBorder, tui.SingleBorder, tui.DoubleBorder, tui.NoBorder}
			styleNames := []string{"Rounded", "Single", "Double", "None"}

			// Cycle to next style
			d.borderStyleIndex = (d.borderStyleIndex + 1) % len(styles)

			panel.SetBorderStyle(styles[d.borderStyleIndex])
			d.app.AddLine(fmt.Sprintf("Border style changed to: %s", styleNames[d.borderStyleIndex]))

		case "t":
			// Toggle title
			if d.app.GetFooter().GetTitle() == "LogViewer" {
				d.app.SetTitle("AgarApp")
			} else {
				d.app.SetTitle("LogViewer")
			}
			d.app.AddLine(fmt.Sprintf("Title changed to: %s", d.app.GetFooter().GetTitle()))

		case "v":
			// Toggle version
			if d.app.GetFooter().GetVersion() == "1.0.0" {
				d.app.SetVersion("2.0.0")
			} else {
				d.app.SetVersion("1.0.0")
			}
			d.app.AddLine(fmt.Sprintf("Version changed to: %s", d.app.GetFooter().GetVersion()))
		}
	}

	return d, nil
}

func (d demoApp) View() string {
	return d.app.View()
}

func main() {
	fmt.Println("=== Application Component Example ===\n")
	fmt.Println("This example demonstrates the tui.Application component which includes:")
	fmt.Println("  - A content panel with configurable margin, padding, and borders")
	fmt.Println("  - Automatic scrolling when content exceeds available space")
	fmt.Println("  - A footer that stays fixed at the bottom")
	fmt.Println("  - Simple API for managing content (AddLine, Clear, etc.)")
	fmt.Println("\nInteractive controls:")
	fmt.Println("  [a] Toggle auto-add mode (adds lines every 2 seconds)")
	fmt.Println("  [l] Manually add a line")
	fmt.Println("  [c] Clear all content")
	fmt.Println("  [m] Cycle through modes")
	fmt.Println("  [b] Cycle through border styles")
	fmt.Println("  [t] Toggle title")
	fmt.Println("  [v] Toggle version")
	fmt.Println("  [q] Quit")
	fmt.Println("\nPress Enter to start the demo...")
	fmt.Scanln()

	p := tea.NewProgram(initialApp(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n=== Demo completed! ===")
}
