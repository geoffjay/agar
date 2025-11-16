package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/geoffjay/agar/tui"
)

// borderDemo demonstrates selective border functionality
type borderDemo struct {
	panels       []*tui.Panel
	currentPanel int
	width        int
	height       int
}

func initialBorderDemo() borderDemo {
	// Create panels with different border configurations
	panels := make([]*tui.Panel, 0)

	// Panel 1: All borders (default)
	p1 := tui.NewPanelWithBorders(1, 1, tui.RoundedBorder, tui.BorderAll)
	p1.SetContent("Panel 1: All Borders\n(BorderAll)")
	panels = append(panels, p1)

	// Panel 2: Only left border
	p2 := tui.NewPanelWithBorders(1, 1, tui.SingleBorder, tui.BorderLeft)
	p2.SetContent("Panel 2: Left Border Only\n(BorderLeft)")
	p2.SetBorderColor(lipgloss.Color("205"))
	panels = append(panels, p2)

	// Panel 3: Only right border
	p3 := tui.NewPanelWithBorders(1, 1, tui.DoubleBorder, tui.BorderRight)
	p3.SetContent("Panel 3: Right Border Only\n(BorderRight)")
	p3.SetBorderColor(lipgloss.Color("86"))
	panels = append(panels, p3)

	// Panel 4: Top and bottom borders
	p4 := tui.NewPanelWithBorders(1, 1, tui.RoundedBorder, tui.BorderTop|tui.BorderBottom)
	p4.SetContent("Panel 4: Top & Bottom\n(BorderTop | BorderBottom)")
	p4.SetBorderColor(lipgloss.Color("226"))
	panels = append(panels, p4)

	// Panel 5: Left and right borders
	p5 := tui.NewPanelWithBorders(1, 1, tui.SingleBorder, tui.BorderLeft|tui.BorderRight)
	p5.SetContent("Panel 5: Left & Right\n(BorderLeft | BorderRight)")
	p5.SetBorderColor(lipgloss.Color("51"))
	panels = append(panels, p5)

	// Panel 6: Top, left, and right (no bottom)
	p6 := tui.NewPanelWithBorders(1, 1, tui.DoubleBorder, tui.BorderTop|tui.BorderLeft|tui.BorderRight)
	p6.SetContent("Panel 6: Top, Left & Right\n(BorderTop | BorderLeft | BorderRight)")
	p6.SetBorderColor(lipgloss.Color("201"))
	panels = append(panels, p6)

	// Panel 7: Bottom, left, and right (no top)
	p7 := tui.NewPanelWithBorders(1, 1, tui.RoundedBorder, tui.BorderBottom|tui.BorderLeft|tui.BorderRight)
	p7.SetContent("Panel 7: Bottom, Left & Right\n(BorderBottom | BorderLeft | BorderRight)")
	p7.SetBorderColor(lipgloss.Color("46"))
	panels = append(panels, p7)

	// Panel 8: Only top border
	p8 := tui.NewPanelWithBorders(1, 1, tui.SingleBorder, tui.BorderTop)
	p8.SetContent("Panel 8: Top Border Only\n(BorderTop)")
	p8.SetBorderColor(lipgloss.Color("208"))
	panels = append(panels, p8)

	// Panel 9: Only bottom border
	p9 := tui.NewPanelWithBorders(1, 1, tui.DoubleBorder, tui.BorderBottom)
	p9.SetContent("Panel 9: Bottom Border Only\n(BorderBottom)")
	p9.SetBorderColor(lipgloss.Color("141"))
	panels = append(panels, p9)

	return borderDemo{
		panels:       panels,
		currentPanel: 0,
		width:        80,
		height:       24,
	}
}

func (d borderDemo) Init() tea.Cmd {
	return nil
}

func (d borderDemo) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		d.width = msg.Width
		d.height = msg.Height

		// Update all panel sizes
		// Each panel gets 1/3 of the width and appropriate height
		panelWidth := d.width / 3
		panelHeight := 6

		for _, panel := range d.panels {
			panel.SetSize(panelWidth, panelHeight)
		}

		return d, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return d, tea.Quit
		}
	}

	return d, nil
}

func (d borderDemo) View() string {
	// Title
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		Render("Border Flags Demo - Selective Panel Borders")

	// Instructions
	instructions := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Render("This demo shows panels with different border configurations using bitwise flags. Press 'q' to quit.")

	// Set panel sizes
	panelWidth := d.width / 3
	panelHeight := 6

	for _, panel := range d.panels {
		panel.SetSize(panelWidth, panelHeight)
	}

	// Render panels in a 3x3 grid
	var rows []string

	// Row 1: Panels 0, 1, 2
	row1 := lipgloss.JoinHorizontal(lipgloss.Top,
		d.panels[0].Render(),
		d.panels[1].Render(),
		d.panels[2].Render(),
	)
	rows = append(rows, row1)

	// Row 2: Panels 3, 4, 5
	row2 := lipgloss.JoinHorizontal(lipgloss.Top,
		d.panels[3].Render(),
		d.panels[4].Render(),
		d.panels[5].Render(),
	)
	rows = append(rows, row2)

	// Row 3: Panels 6, 7, 8
	row3 := lipgloss.JoinHorizontal(lipgloss.Top,
		d.panels[6].Render(),
		d.panels[7].Render(),
		d.panels[8].Render(),
	)
	rows = append(rows, row3)

	// Combine everything
	content := lipgloss.JoinVertical(lipgloss.Left,
		title,
		"",
		instructions,
		"",
		rows[0],
		rows[1],
		rows[2],
	)

	return content
}

func main() {
	fmt.Println("=== Border Flags Demo ===\n")
	fmt.Println("This example demonstrates selective border rendering using bitwise flags.")
	fmt.Println("You can specify which sides of a panel should have borders:")
	fmt.Println()
	fmt.Println("  BorderTop     - Top border only")
	fmt.Println("  BorderBottom  - Bottom border only")
	fmt.Println("  BorderLeft    - Left border only")
	fmt.Println("  BorderRight   - Right border only")
	fmt.Println("  BorderAll     - All borders (convenience constant)")
	fmt.Println()
	fmt.Println("You can combine flags using bitwise OR:")
	fmt.Println("  BorderLeft | BorderRight        - Left and right borders")
	fmt.Println("  BorderTop | BorderBottom        - Top and bottom borders")
	fmt.Println("  BorderTop | BorderLeft | BorderRight - Three sides (no bottom)")
	fmt.Println()
	fmt.Println("Press Enter to start the demo...")
	fmt.Scanln()

	p := tea.NewProgram(initialBorderDemo(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n=== Demo completed! ===")
	fmt.Println("\nExample usage:")
	fmt.Println("  // Create a panel with only left and right borders")
	fmt.Println("  panel := tui.NewPanelWithBorders(1, 1, tui.SingleBorder, tui.BorderLeft | tui.BorderRight)")
	fmt.Println()
	fmt.Println("  // You can also change borders on an existing panel")
	fmt.Println("  panel.SetBorderFlags(tui.BorderTop | tui.BorderBottom)")
	fmt.Println()
	fmt.Println("  // Change the border color")
	fmt.Println("  panel.SetBorderColor(lipgloss.Color(\"205\"))")
}
