package panels

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// PanelHeader is a tea.Model-compatible title bar for specialist/artifact panels.
type PanelHeader struct {
	title   string
	count   int
	width   int
	focused bool
}

// NewPanelHeader creates a header with the given title and no count.
func NewPanelHeader(title string) PanelHeader {
	return PanelHeader{title: title, count: -1}
}

// Init returns nil — PanelHeader is a no-op Bubbletea model.
func (ph PanelHeader) Init() tea.Cmd { return nil }

// Update is a pass-through that returns the header unchanged.
func (ph PanelHeader) Update(_ tea.Msg) (PanelHeader, tea.Cmd) {
	return ph, nil
}

// View renders the header title with optional count and truncation.
func (ph PanelHeader) View() string {
	header := "\u258e " + ph.title
	if ph.count >= 0 && ph.width >= 20 {
		header += fmt.Sprintf(" (%d)", ph.count)
	}

	style := lipgloss.NewStyle().Bold(true)
	if ph.focused {
		style = style.Foreground(ColorPrimary)
	} else {
		style = style.Foreground(ColorInactive)
	}

	if ph.width > 0 && ph.width < 15 {
		runes := []rune(header)
		maxLen := ph.width - 2
		if maxLen < 1 {
			maxLen = 1
		}
		if len(runes) > maxLen {
			header = string(runes[:maxLen]) + "\u2026"
		}
	}

	return style.Render(header)
}

// UpdateSize sets the available width for truncation calculations.
func (ph PanelHeader) UpdateSize(width, _ int) (PanelHeader, tea.Cmd) {
	ph.width = width
	return ph, nil
}

// SetFocused marks the header as focused, changing its foreground color.
func (ph *PanelHeader) SetFocused(f bool) { ph.focused = f }

// SetCount updates the displayed item count.
func (ph *PanelHeader) SetCount(n int) { ph.count = n }
