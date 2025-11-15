package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// BorderStyle represents the style of border to use
type BorderStyle int

const (
	NoBorder BorderStyle = iota
	SingleBorder
	DoubleBorder
	RoundedBorder
)

// Panel represents a content area with configurable margin, padding, and borders
type Panel struct {
	content      string
	width        int
	height       int
	margin       int
	padding      int
	borderStyle  BorderStyle
	style        lipgloss.Style
	borderColor  lipgloss.Color
	bgColor      lipgloss.Color
}

// PanelStyle defines default styling for panels
var PanelStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("252"))

// NewPanel creates a new panel with the specified configuration
func NewPanel(margin, padding int, borderStyle BorderStyle) *Panel {
	return &Panel{
		content:      "",
		width:        80,
		height:       24,
		margin:       margin,
		padding:      padding,
		borderStyle:  borderStyle,
		style:        PanelStyle,
		borderColor:  lipgloss.Color("241"),
		bgColor:      lipgloss.Color(""),
	}
}

// SetContent sets the panel content
func (p *Panel) SetContent(content string) {
	p.content = content
}

// GetContent returns the panel content
func (p *Panel) GetContent() string {
	return p.content
}

// SetSize sets the panel dimensions
func (p *Panel) SetSize(width, height int) {
	p.width = width
	p.height = height
}

// SetMargin sets the panel margin
func (p *Panel) SetMargin(margin int) {
	p.margin = margin
}

// SetPadding sets the panel padding
func (p *Panel) SetPadding(padding int) {
	p.padding = padding
}

// SetBorderStyle sets the border style
func (p *Panel) SetBorderStyle(style BorderStyle) {
	p.borderStyle = style
}

// SetBorderColor sets the border color
func (p *Panel) SetBorderColor(color lipgloss.Color) {
	p.borderColor = color
}

// SetBackgroundColor sets the background color
func (p *Panel) SetBackgroundColor(color lipgloss.Color) {
	p.bgColor = color
}

// SetStyle sets the content style
func (p *Panel) SetStyle(style lipgloss.Style) {
	p.style = style
}

// GetWidth returns the total width including margin
func (p *Panel) GetWidth() int {
	return p.width
}

// GetHeight returns the total height including margin
func (p *Panel) GetHeight() int {
	return p.height
}

// GetContentWidth returns the available width for content
func (p *Panel) GetContentWidth() int {
	width := p.width - (p.margin * 2)
	if p.borderStyle != NoBorder {
		width -= 2 // Account for left and right borders
	}
	width -= (p.padding * 2)
	if width < 1 {
		width = 1
	}
	return width
}

// GetContentHeight returns the available height for content
func (p *Panel) GetContentHeight() int {
	height := p.height - (p.margin * 2)
	if p.borderStyle != NoBorder {
		height -= 2 // Account for top and bottom borders
	}
	height -= (p.padding * 2)
	if height < 1 {
		height = 1
	}
	return height
}

// Render renders the panel with all styling applied
func (p *Panel) Render() string {
	var b strings.Builder

	// Add top margin
	if p.margin > 0 {
		b.WriteString(strings.Repeat("\n", p.margin))
	}

	// Calculate inner dimensions
	contentWidth := p.GetContentWidth()
	contentHeight := p.GetContentHeight()

	// Split content into lines
	contentLines := strings.Split(p.content, "\n")

	// Prepare the border characters
	var topLeft, topRight, bottomLeft, bottomRight, horizontal, vertical string
	switch p.borderStyle {
	case SingleBorder:
		topLeft, topRight, bottomLeft, bottomRight = "┌", "┐", "└", "┘"
		horizontal, vertical = "─", "│"
	case DoubleBorder:
		topLeft, topRight, bottomLeft, bottomRight = "╔", "╗", "╚", "╝"
		horizontal, vertical = "═", "║"
	case RoundedBorder:
		topLeft, topRight, bottomLeft, bottomRight = "╭", "╮", "╰", "╯"
		horizontal, vertical = "─", "│"
	}

	borderStyle := lipgloss.NewStyle().Foreground(p.borderColor)

	// Build the panel content
	var panelLines []string

	// Top border
	if p.borderStyle != NoBorder {
		leftMargin := strings.Repeat(" ", p.margin)
		borderLine := leftMargin + borderStyle.Render(topLeft+strings.Repeat(horizontal, contentWidth+p.padding*2)+topRight)
		panelLines = append(panelLines, borderLine)
	}

	// Top padding
	if p.padding > 0 && p.borderStyle != NoBorder {
		for i := 0; i < p.padding; i++ {
			leftMargin := strings.Repeat(" ", p.margin)
			paddingLine := leftMargin + borderStyle.Render(vertical) + strings.Repeat(" ", contentWidth+p.padding*2) + borderStyle.Render(vertical)
			panelLines = append(panelLines, paddingLine)
		}
	} else if p.padding > 0 {
		for i := 0; i < p.padding; i++ {
			panelLines = append(panelLines, "")
		}
	}

	// Content lines
	for i := 0; i < contentHeight; i++ {
		var line string
		if i < len(contentLines) {
			line = contentLines[i]
		} else {
			line = ""
		}

		// Truncate or pad the line to fit
		if len(line) > contentWidth {
			line = line[:contentWidth]
		} else {
			line = line + strings.Repeat(" ", contentWidth-len(line))
		}

		// Apply content style
		styledLine := p.style.Render(line)

		// Add borders and padding
		leftMargin := strings.Repeat(" ", p.margin)
		leftPadding := strings.Repeat(" ", p.padding)
		rightPadding := strings.Repeat(" ", p.padding)

		if p.borderStyle != NoBorder {
			panelLines = append(panelLines, leftMargin+borderStyle.Render(vertical)+leftPadding+styledLine+rightPadding+borderStyle.Render(vertical))
		} else {
			panelLines = append(panelLines, leftMargin+leftPadding+styledLine+rightPadding)
		}
	}

	// Bottom padding
	if p.padding > 0 && p.borderStyle != NoBorder {
		for i := 0; i < p.padding; i++ {
			leftMargin := strings.Repeat(" ", p.margin)
			paddingLine := leftMargin + borderStyle.Render(vertical) + strings.Repeat(" ", contentWidth+p.padding*2) + borderStyle.Render(vertical)
			panelLines = append(panelLines, paddingLine)
		}
	} else if p.padding > 0 {
		for i := 0; i < p.padding; i++ {
			panelLines = append(panelLines, "")
		}
	}

	// Bottom border
	if p.borderStyle != NoBorder {
		leftMargin := strings.Repeat(" ", p.margin)
		borderLine := leftMargin + borderStyle.Render(bottomLeft+strings.Repeat(horizontal, contentWidth+p.padding*2)+bottomRight)
		panelLines = append(panelLines, borderLine)
	}

	// Join all lines
	b.WriteString(strings.Join(panelLines, "\n"))

	// Add bottom margin (if needed)
	// Note: bottom margin is typically not added as it's at the end

	return b.String()
}
