package tui

import (
	"strings"
)

// LayoutDirection represents the direction of the layout
type LayoutDirection int

const (
	Vertical LayoutDirection = iota
	Horizontal
)

// LayoutComponent represents a component that can be rendered in a layout
type LayoutComponent interface {
	Render() string
	GetWidth() int
	GetHeight() int
	SetSize(width, height int)
}

// Layout manages the arrangement of components
type Layout struct {
	direction  LayoutDirection
	components []LayoutComponent
	width      int
	height     int
}

// NewLayout creates a new layout with the specified direction
func NewLayout(direction LayoutDirection) *Layout {
	return &Layout{
		direction:  direction,
		components: make([]LayoutComponent, 0),
		width:      80,
		height:     24,
	}
}

// AddComponent adds a component to the layout
func (l *Layout) AddComponent(component LayoutComponent) {
	l.components = append(l.components, component)
}

// RemoveComponent removes a component from the layout
func (l *Layout) RemoveComponent(index int) {
	if index >= 0 && index < len(l.components) {
		l.components = append(l.components[:index], l.components[index+1:]...)
	}
}

// SetSize sets the layout dimensions
func (l *Layout) SetSize(width, height int) {
	l.width = width
	l.height = height
}

// GetWidth returns the layout width
func (l *Layout) GetWidth() int {
	return l.width
}

// GetHeight returns the layout height
func (l *Layout) GetHeight() int {
	return l.height
}

// Render renders the layout and all its components
func (l *Layout) Render() string {
	if len(l.components) == 0 {
		return ""
	}

	if l.direction == Vertical {
		return l.renderVertical()
	}
	return l.renderHorizontal()
}

// renderVertical renders components in a vertical stack
func (l *Layout) renderVertical() string {
	var b strings.Builder

	for i, component := range l.components {
		// Update component size to match layout width
		// Height is managed by the component itself
		component.SetSize(l.width, component.GetHeight())

		// Render the component
		rendered := component.Render()
		b.WriteString(rendered)

		// Add newline between components (but not after the last one)
		if i < len(l.components)-1 {
			b.WriteString("\n")
		}
	}

	return b.String()
}

// renderHorizontal renders components side by side
func (l *Layout) renderHorizontal() string {
	if len(l.components) == 0 {
		return ""
	}

	// Calculate width for each component
	componentWidth := l.width / len(l.components)

	// Get all rendered components
	renderedComponents := make([][]string, len(l.components))
	maxLines := 0

	for i, component := range l.components {
		component.SetSize(componentWidth, l.height)
		rendered := component.Render()
		lines := strings.Split(rendered, "\n")
		renderedComponents[i] = lines
		if len(lines) > maxLines {
			maxLines = len(lines)
		}
	}

	// Combine lines horizontally
	var b strings.Builder
	for lineIdx := 0; lineIdx < maxLines; lineIdx++ {
		for _, lines := range renderedComponents {
			var line string
			if lineIdx < len(lines) {
				line = lines[lineIdx]
			} else {
				line = strings.Repeat(" ", componentWidth)
			}

			// Ensure line is exactly componentWidth
			if len(line) < componentWidth {
				line += strings.Repeat(" ", componentWidth-len(line))
			} else if len(line) > componentWidth {
				line = line[:componentWidth]
			}

			b.WriteString(line)
		}
		if lineIdx < maxLines-1 {
			b.WriteString("\n")
		}
	}

	return b.String()
}
