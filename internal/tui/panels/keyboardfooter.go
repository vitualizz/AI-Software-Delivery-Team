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

// compactFooterWidthThreshold is the width at or below which the footer
// renders in compact mode — key tokens only, no descriptions or labels.
const compactFooterWidthThreshold = 50

// RenderKeyboardFooter renders hint lines, scaled by terminal width.
func RenderKeyboardFooter(groups []HintGroup, width int) string {
	if len(groups) == 0 {
		return ""
	}

	if width <= compactFooterWidthThreshold {
		return renderCompactFooter(groups)
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

	return strings.Join(parts, " • ")
}

// renderCompactFooter renders a narrow-width footer with only key tokens
// (no labels or descriptions), joined by '|'. Used at width <=
// compactFooterWidthThreshold so the footer still hints at available
// shortcuts without consuming the limited horizontal space.
func renderCompactFooter(groups []HintGroup) string {
	keyStyle := lipgloss.NewStyle().Foreground(ColorMuted)

	var parts []string
	for _, g := range groups {
		for _, h := range g.Hints {
			parts = append(parts, keyStyle.Render(h.Key))
		}
	}

	return strings.Join(parts, "•")
}
