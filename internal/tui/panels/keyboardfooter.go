package panels

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Hint represents a single keyboard shortcut with its description.
type Hint struct {
	Key         string
	Description string
}

// HintGroup is a labelled collection of related keyboard hints.
type HintGroup struct {
	Label string
	Hints []Hint
}

// RenderKeyboardFooter renders hint lines, scaled by terminal width.
func RenderKeyboardFooter(groups []HintGroup, width int) string {
	if width <= 50 || len(groups) == 0 {
		return ""
	}

	keyStyle := lipgloss.NewStyle().Foreground(ColorMuted)
	descStyle := lipgloss.NewStyle().Foreground(ColorPrimary)

	var parts []string
	for _, g := range groups {
		for _, h := range g.Hints {
			if width > 80 {
				parts = append(parts, keyStyle.Render(h.Key)+": "+descStyle.Render(h.Description))
			} else {
				parts = append(parts, keyStyle.Render(h.Key))
			}
		}
	}

	return strings.Join(parts, " | ")
}
