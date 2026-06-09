package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/vitualizz/ai-software-delivery-team/internal/tui/panels"
)

// SectionGroup groups related CheckRows under a titled header in the
// pre-flight check screen.
type SectionGroup struct {
	Title string
	Rows  []CheckRow
}

// Render returns the styled multi-line representation of this section.
func (g SectionGroup) Render(width int) string {
	var b strings.Builder
	titleLine := lipgloss.NewStyle().Bold(true).Foreground(panels.ColorPrimary).Render("  " + g.Title)
	b.WriteString(titleLine + "\n")
	for _, row := range g.Rows {
		b.WriteString(row.Render(width) + "\n")
	}
	return b.String()
}
