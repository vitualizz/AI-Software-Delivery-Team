package panels

import "github.com/charmbracelet/lipgloss"

// StatusState represents the lifecycle state of a specialist step.
type StatusState int

// StatusState values for step lifecycle.
const (
	StatusIdle StatusState = iota
	StatusRunning
	StatusDone
	StatusError
)

// StatusIcon returns a unicode symbol for the given state.
func StatusIcon(state StatusState) string {
	switch state {
	case StatusRunning:
		return "\u25cc"
	case StatusDone:
		return "\u2713"
	case StatusError:
		return "\u2717"
	default:
		return "\u25cf"
	}
}

// StatusColor returns the palette color for the given state.
func StatusColor(state StatusState) lipgloss.AdaptiveColor {
	switch state {
	case StatusRunning:
		return ColorSecondary
	case StatusDone:
		return ColorSuccess
	case StatusError:
		return ColorError
	default:
		return ColorInactive
	}
}

// StatusIndicator renders a state icon with optional side padding.
type StatusIndicator struct {
	state   StatusState
	compact bool
}

// NewStatusIndicator creates an indicator for the given state.
func NewStatusIndicator(state StatusState) StatusIndicator {
	return StatusIndicator{state: state}
}

// Render returns the colored state icon string.
func (si StatusIndicator) Render() string {
	icon := StatusIcon(si.state)
	color := StatusColor(si.state)
	colored := lipgloss.NewStyle().Foreground(color).Render(icon)
	if si.compact {
		return colored
	}
	return " " + colored + " "
}

// FocusBorderStyle returns a border style with per-side colors when focused.
func FocusBorderStyle(focused bool) lipgloss.Style {
	if focused {
		return lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderTopForeground(ColorPrimary).
			BorderBottomForeground(ColorInactive).
			BorderLeftForeground(ColorSecondary).
			BorderRightForeground(ColorInactive)
	}
	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(ColorInactive)
}
