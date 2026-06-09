package setup

import (
	"strings"

	"github.com/vitualizz/ai-software-delivery-team/internal/setup/components"
	"github.com/vitualizz/ai-software-delivery-team/internal/tui/panels"
)

// initialPreflightSections returns the seed state for the pre-flight check screen.
// All rows start in CheckStatusPending so the view immediately shows the spinner.
func initialPreflightSections() []components.SectionGroup {
	return []components.SectionGroup{
		{
			Title: "Your Environment",
			Rows: []components.CheckRow{
				{Label: "OS / Arch", Status: components.CheckStatusPending},
			},
		},
		{
			Title: "Memory Provider",
			Rows: []components.CheckRow{
				{Label: "Engram", Status: components.CheckStatusPending},
			},
		},
		{
			Title: "AI Enhancements",
			Rows: []components.CheckRow{
				{Label: "Codegraph", Status: components.CheckStatusPending, SoftWarn: true},
			},
		},
	}
}

func renderPreflightCheck(m Model) string {
	var b strings.Builder
	for _, sg := range m.preflightSections {
		b.WriteString(sg.Render(m.width))
		b.WriteString("\n")
	}

	var footer string
	switch {
	case !m.preflightDone:
		footer = panels.RenderKeyboardFooter([]panels.HintGroup{
			{Label: "Status", Hints: []panels.Hint{{Key: "checking", Description: "environment"}}},
		}, m.width)
	case m.engramMissing:
		footer = panels.RenderKeyboardFooter([]panels.HintGroup{
			{Label: "Required", Hints: []panels.Hint{
				{Key: "engram", Description: "required — install: https://github.com/Gentleman-Programming/engram"},
				{Key: "esc", Description: "back"},
			}},
		}, m.width)
	default:
		footer = panels.RenderKeyboardFooter([]panels.HintGroup{
			{Label: "Actions", Hints: []panels.Hint{
				{Key: "enter", Description: "continue"},
				{Key: "esc", Description: "back"},
			}},
		}, m.width)
	}
	return frame("Pre-flight Check", strings.TrimRight(b.String(), "\n"), footer, true)
}
