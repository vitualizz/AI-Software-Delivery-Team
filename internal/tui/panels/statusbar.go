package panels

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// StatusBar is a tea.Model-compatible bottom bar showing change and specialist.
type StatusBar struct {
	change     string
	specialist string
	width      int
	compact    bool
}

// NewStatusBar returns a zero-value StatusBar.
func NewStatusBar() StatusBar {
	return StatusBar{}
}

// Init returns nil — StatusBar is a no-op Bubbletea model.
func (sb StatusBar) Init() tea.Cmd { return nil }

// Update is a pass-through that returns the bar unchanged.
func (sb StatusBar) Update(_ tea.Msg) (StatusBar, tea.Cmd) {
	return sb, nil
}

// View renders the status bar with responsive detail levels.
func (sb StatusBar) View() string {
	var line string
	switch {
	case sb.width > 80:
		line = fmt.Sprintf(" change: %s  specialist: %s  [tab] switch panel  [q] quit", sb.change, sb.specialist)
	case sb.width > 50:
		line = fmt.Sprintf(" change: %s  specialist: %s", sb.change, sb.specialist)
	default:
		line = fmt.Sprintf(" change: %s", sb.change)
		if sb.compact {
			line += "  (more available — widen window)"
		}
	}

	style := lipgloss.NewStyle().
		Foreground(ColorOnInactive).
		Background(ColorInactive).
		Width(sb.width)
	return style.Render(line)
}

// UpdateSize stores the available width for text truncation.
func (sb StatusBar) UpdateSize(width, _ int) (StatusBar, tea.Cmd) {
	sb.width = width
	return sb, nil
}

// SetChange updates the active change identifier.
func (sb *StatusBar) SetChange(c string) { sb.change = c }

// SetSpecialist updates the selected specialist name.
func (sb *StatusBar) SetSpecialist(s string) { sb.specialist = s }

// SetCompact marks whether the dashboard is in compact layout mode (<50
// cols, where the artifacts panel is hidden entirely). When true, the
// narrowest status-bar tier surfaces an inline "more available" cue so the
// user knows there's hidden content beyond the visible specialists panel.
func (sb *StatusBar) SetCompact(c bool) { sb.compact = c }
