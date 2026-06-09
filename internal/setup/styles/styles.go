// Package styles provides a centralized Lipgloss palette for the setup TUI.
package styles

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/vitualizz/ai-software-delivery-team/internal/tui/panels"
)

// Palette groups all named styles used by the installer TUI.
type Palette struct {
	// Cursor highlights the currently focused list item.
	Cursor lipgloss.Style
	// Success indicates a successful operation (green).
	Success lipgloss.Style
	// Error indicates a failure (red).
	Error lipgloss.Style
	// Warning indicates a cautionary condition (amber).
	Warning lipgloss.Style
	// Dim reduces visual prominence of secondary elements.
	Dim lipgloss.Style
	// Box wraps a screen body in a rounded border with padding.
	Box lipgloss.Style
	// StatusBar renders the per-screen key-hint footer with fg/bg colors.
	StatusBar lipgloss.Style
}

// Default is the package-level palette used by views.go.
var Default = Palette{
	Cursor:  lipgloss.NewStyle().Bold(true).Foreground(panels.ColorSecondary),
	Success: lipgloss.NewStyle().Foreground(panels.ColorSuccess),
	Error:   lipgloss.NewStyle().Foreground(panels.ColorError),
	Warning: lipgloss.NewStyle().Bold(true).Foreground(panels.ColorWarning),
	Dim:     lipgloss.NewStyle().Faint(true),
	Box: panels.FocusBorderStyle(true).
		Padding(1, 2),
	StatusBar: lipgloss.NewStyle().
		Foreground(panels.ColorPrimary).
		Background(panels.ColorInactive),
}
