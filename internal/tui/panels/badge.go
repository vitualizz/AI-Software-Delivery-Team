package panels

import "github.com/charmbracelet/lipgloss"

// Badge is a small value type rendering a bracketed, tinted label —
// mirrors the StatusIndicator precedent (constructor + Render() string,
// not a tea.Model). The caller supplies tone so Badge stays a pure
// rendering primitive with no hardcoded semantic-state mapping.
type Badge struct {
	label string
	tone  lipgloss.AdaptiveColor
}

// NewBadge constructs a badge with the given label and tone. The tone
// should be one of the package-level AdaptiveColor vars (e.g. ColorSuccess,
// ColorError, ColorInactive) — Badge never introduces its own hex literals.
func NewBadge(label string, tone lipgloss.AdaptiveColor) Badge {
	return Badge{label: label, tone: tone}
}

// Render returns the styled, bracketed badge string, e.g. "[present]"
// rendered in the badge's tone color.
func (b Badge) Render() string {
	return lipgloss.NewStyle().Foreground(b.tone).Render("[" + b.label + "]")
}
