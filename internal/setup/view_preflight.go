package setup

import (
	"fmt"
	"strings"

	"github.com/vitualizz/asdt/internal/i18n"
	"github.com/vitualizz/asdt/internal/setup/components"
	"github.com/vitualizz/asdt/internal/setup/styles"
	"github.com/vitualizz/asdt/internal/tui/panels"
)

// initialPreflightSections returns the seed state for the pre-flight check screen.
// Section titles come from the active catalog; row labels are kept as constants
// because they also serve as lookup keys in rowHasStatus and in probe messages.
func initialPreflightSections(s i18n.InstallerStrings) []components.SectionGroup {
	return []components.SectionGroup{
		{
			Title: s.SectionYourEnvironment,
			Rows: []components.CheckRow{
				{Label: "OS / Arch", Status: components.CheckStatusPending},
				{Label: "Shell", Status: components.CheckStatusPending},
			},
		},
		{
			Title: s.SectionMemoryProvider,
			Rows: []components.CheckRow{
				{Label: "Engram", Status: components.CheckStatusPending},
			},
		},
		{
			Title: s.SectionAIEnhancements,
			Rows: []components.CheckRow{
				{Label: "Codegraph", Status: components.CheckStatusPending, SoftWarn: true},
			},
		},
	}
}

func renderPreflightCheck(m Model) string {
	s := m.catalog.Installer
	var b strings.Builder

	fmt.Fprintf(&b, "  %s\n\n", stepLine(s, 1, 6))

	for _, sg := range m.preflight.sections {
		b.WriteString(sg.Render(m.width))
		b.WriteString("\n")
	}

	if m.preflight.engramMissing && m.preflight.done {
		fmt.Fprintf(&b, "\n  %s\n", styles.Default.Warning.Render(s.PrefEngramRequired))
		fmt.Fprintf(&b, "  %s\n", styles.Default.Dim.Render(s.PrefEngramInstall))
		fmt.Fprintf(&b, "  %s\n", styles.Default.Dim.Render(s.PrefEngramRestart))
	}

	var footer string
	switch {
	case !m.preflight.done:
		footer = panels.RenderKeyboardFooter([]panels.HintGroup{
			{Label: s.HintGroupStatus, Hints: []panels.Hint{{Key: s.HintChecking, Description: s.HintEnvironment}}},
		}, m.width)
	case m.preflight.engramMissing:
		footer = panels.RenderKeyboardFooter([]panels.HintGroup{
			{Label: s.HintGroupRequired, Hints: []panels.Hint{
				{Key: "engram", Description: s.HintEngramRequired},
				{Key: "esc", Description: s.HintBack},
			}},
		}, m.width)
	default:
		footer = panels.RenderKeyboardFooter([]panels.HintGroup{
			{Label: s.HintGroupActions, Hints: []panels.Hint{
				{Key: "enter", Description: s.HintContinue},
				{Key: "esc", Description: s.HintBack},
			}},
		}, m.width)
	}
	return frame(s.TitlePreflightCheck, strings.TrimRight(b.String(), "\n"), footer, true)
}
