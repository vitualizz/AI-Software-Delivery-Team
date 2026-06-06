// Package styles provides a centralized Lipgloss palette for the setup TUI.
package styles

import "github.com/charmbracelet/lipgloss"

// Palette groups all named styles used by the installer TUI.
type Palette struct {
	// Header is bold with Catppuccin Mauve — used for screen titles.
	Header lipgloss.Style
	// Cursor highlights the currently focused list item.
	Cursor lipgloss.Style
	// Selected is used for items marked with [x].
	Selected lipgloss.Style
	// Unselected dims items that are not in focus.
	Unselected lipgloss.Style
	// Success indicates a successful operation (green).
	Success lipgloss.Style
	// Error indicates a failure (red).
	Error lipgloss.Style
	// Label is a neutral white-ish text style for item names.
	Label lipgloss.Style
	// Description provides secondary, greyed-out contextual text.
	Description lipgloss.Style
	// Dim reduces visual prominence of secondary elements.
	Dim lipgloss.Style
	// Bold emphasizes text.
	Bold lipgloss.Style
}

// Default is the package-level palette used by views.go.
var Default = Palette{
	Header:      lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#cba6f7")),
	Cursor:      lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#b4befe")),
	Selected:    lipgloss.NewStyle().Foreground(lipgloss.Color("#b4befe")),
	Unselected:  lipgloss.NewStyle().Foreground(lipgloss.Color("#6c7086")),
	Success:     lipgloss.NewStyle().Foreground(lipgloss.Color("#a6e3a1")),
	Error:       lipgloss.NewStyle().Foreground(lipgloss.Color("#f38ba8")),
	Label:       lipgloss.NewStyle().Foreground(lipgloss.Color("#cba6f7")),
	Description: lipgloss.NewStyle().Faint(true).Foreground(lipgloss.Color("#6c7086")),
	Dim:         lipgloss.NewStyle().Faint(true),
	Bold:        lipgloss.NewStyle().Bold(true),
}
