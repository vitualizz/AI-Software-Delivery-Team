package panels

import (
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/lipgloss"
)

// NewSpinner constructs an indeterminate spinner.Model using the Dot frame
// set, tinted with ColorSecondary — the cyan/sky pastel reserved for
// in-progress/running indicators (mirrors StatusColor's StatusRunning case).
//
// Exported (not package-private newSpinner) because internal/setup also needs
// an indeterminate spinner for StateInstalling and must not duplicate the
// frame-set/tint decision — a second hand-rolled construction would drift the
// moment this palette or spinner style changes.
func NewSpinner() spinner.Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(ColorSecondary)
	return s
}
