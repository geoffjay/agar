package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

// BorderStyle represents the style of border to use
type BorderStyle int

const (
	NoBorder BorderStyle = iota
	SingleBorder
	DoubleBorder
	RoundedBorder
)

// BorderFlags represents which sides of the panel should have borders
type BorderFlags uint8

const (
	BorderTop    BorderFlags = 1 << iota // 0001
	BorderBottom                          // 0010
	BorderLeft                            // 0100
	BorderRight                           // 1000
	BorderAll    = BorderTop | BorderBottom | BorderLeft | BorderRight
)

// Panel represents a content area with configurable margin, padding, and borders
type Panel struct {
	content      string
	width        int
	height       int
	margin       int
	padding      int
	borderStyle  BorderStyle
	borderFlags  BorderFlags
	style        lipgloss.Style
	borderColor  lipgloss.Color
	bgColor      lipgloss.Color
}

// PanelStyle defines default styling for panels
var PanelStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("252"))

// NewPanel creates a new panel with the specified configuration
// By default, all borders are enabled if borderStyle is not NoBorder
func NewPanel(margin, padding int, borderStyle BorderStyle) *Panel {
	borderFlags := BorderFlags(0)
	if borderStyle != NoBorder {
		borderFlags = BorderAll
	}

	return &Panel{
		content:      "",
		width:        80,
		height:       24,
		margin:       margin,
		padding:      padding,
		borderStyle:  borderStyle,
		borderFlags:  borderFlags,
		style:        PanelStyle,
		borderColor:  lipgloss.Color("241"),
		bgColor:      lipgloss.Color(""),
	}
}

// NewPanelWithBorders creates a new panel with selective border configuration
func NewPanelWithBorders(margin, padding int, borderStyle BorderStyle, borderFlags BorderFlags) *Panel {
	return &Panel{
		content:      "",
		width:        80,
		height:       24,
		margin:       margin,
		padding:      padding,
		borderStyle:  borderStyle,
		borderFlags:  borderFlags,
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
	// Auto-enable all borders if switching from NoBorder to a border style
	if style != NoBorder && p.borderFlags == 0 {
		p.borderFlags = BorderAll
	}
}

// SetBorderFlags sets which sides should have borders
func (p *Panel) SetBorderFlags(flags BorderFlags) {
	p.borderFlags = flags
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
		if p.borderFlags&BorderLeft != 0 {
			width-- // Account for left border
		}
		if p.borderFlags&BorderRight != 0 {
			width-- // Account for right border
		}
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
		if p.borderFlags&BorderTop != 0 {
			height-- // Account for top border
		}
		if p.borderFlags&BorderBottom != 0 {
			height-- // Account for bottom border
		}
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

	// Check which borders are enabled
	hasTop := p.borderStyle != NoBorder && p.borderFlags&BorderTop != 0
	hasBottom := p.borderStyle != NoBorder && p.borderFlags&BorderBottom != 0
	hasLeft := p.borderStyle != NoBorder && p.borderFlags&BorderLeft != 0
	hasRight := p.borderStyle != NoBorder && p.borderFlags&BorderRight != 0

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

	leftMargin := strings.Repeat(" ", p.margin)

	// Top border
	if hasTop {
		var borderLine string
		borderLine = leftMargin

		// Top-left corner (only if both top and left are present)
		if hasLeft {
			borderLine += borderStyle.Render(topLeft)
		}

		// Top horizontal line
		borderLine += borderStyle.Render(strings.Repeat(horizontal, contentWidth+p.padding*2))

		// Top-right corner (only if both top and right are present)
		if hasRight {
			borderLine += borderStyle.Render(topRight)
		}

		panelLines = append(panelLines, borderLine)
	}

	// Top padding
	if p.padding > 0 {
		for i := 0; i < p.padding; i++ {
			var paddingLine string
			paddingLine = leftMargin

			if hasLeft {
				paddingLine += borderStyle.Render(vertical)
			}

			paddingLine += strings.Repeat(" ", contentWidth+p.padding*2)

			if hasRight {
				paddingLine += borderStyle.Render(vertical)
			}

			panelLines = append(panelLines, paddingLine)
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

		// Truncate or pad the line to fit (use visual width, not byte length)
		lineWidth := runewidth.StringWidth(line)
		if lineWidth > contentWidth {
			// Truncate to visual width
			line = runewidth.Truncate(line, contentWidth, "")
		} else if lineWidth < contentWidth {
			// Pad to visual width
			line = line + strings.Repeat(" ", contentWidth-lineWidth)
		}

		// Apply content style
		styledLine := p.style.Render(line)

		// Build the content line with borders and padding
		var contentLine string
		contentLine = leftMargin

		if hasLeft {
			contentLine += borderStyle.Render(vertical)
		}

		contentLine += strings.Repeat(" ", p.padding) + styledLine + strings.Repeat(" ", p.padding)

		if hasRight {
			contentLine += borderStyle.Render(vertical)
		}

		panelLines = append(panelLines, contentLine)
	}

	// Bottom padding
	if p.padding > 0 {
		for i := 0; i < p.padding; i++ {
			var paddingLine string
			paddingLine = leftMargin

			if hasLeft {
				paddingLine += borderStyle.Render(vertical)
			}

			paddingLine += strings.Repeat(" ", contentWidth+p.padding*2)

			if hasRight {
				paddingLine += borderStyle.Render(vertical)
			}

			panelLines = append(panelLines, paddingLine)
		}
	}

	// Bottom border
	if hasBottom {
		var borderLine string
		borderLine = leftMargin

		// Bottom-left corner (only if both bottom and left are present)
		if hasLeft {
			borderLine += borderStyle.Render(bottomLeft)
		}

		// Bottom horizontal line
		borderLine += borderStyle.Render(strings.Repeat(horizontal, contentWidth+p.padding*2))

		// Bottom-right corner (only if both bottom and right are present)
		if hasRight {
			borderLine += borderStyle.Render(bottomRight)
		}

		panelLines = append(panelLines, borderLine)
	}

	// Join all lines
	b.WriteString(strings.Join(panelLines, "\n"))

	return b.String()
}
