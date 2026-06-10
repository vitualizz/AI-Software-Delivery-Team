package setup

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/vitualizz/ai-software-delivery-team/internal/installer"
	"github.com/vitualizz/ai-software-delivery-team/internal/setup/styles"
	"github.com/vitualizz/ai-software-delivery-team/internal/tui/panels"
)

const cursorChar = "►"

func renderState(m Model) string {
	switch m.state {
	case StateEnvironmentCheck:
		return renderPreflightCheck(m)
	case StateDashboard:
		return renderDashboard(m)
	case StateMainMenu:
		return renderMainMenu(m)
	case StateAssistantList:
		return renderAssistantList(m)
	case StateSelectAssistants:
		return renderSelectAssistants(m)
	case StateSelectProvider:
		return renderSelectProvider(m)
	case StateAgentSetup:
		return renderAgentSetup(m)
	case StateAgentWriteMode:
		return renderAgentWriteMode(m)
	case StateInstalling:
		return renderInstalling(m)
	case StateDone:
		return renderDone(m)
	default:
		return ""
	}
}

// titleStyle returns an inline bold style with the primary panel color.
func titleStyle() lipgloss.Style {
	return lipgloss.NewStyle().Bold(true).Foreground(panels.ColorPrimary)
}

// frame composes a screen from a title, a body block, and a footer hint,
// wrapping the title+body in the bordered Box and rendering the footer
// through StatusBar, joined vertically. The focused flag controls border
// decoration via FocusBorderStyle.
func frame(title, body, footer string, focused bool) string {
	header := titleStyle().Render("▎ " + title)
	inner := lipgloss.JoinVertical(lipgloss.Left, header, "", body)
	boxed := panels.FocusBorderStyle(focused).
		Padding(1, 2).
		Render(inner)
	bar := styles.Default.StatusBar.
		Width(lipgloss.Width(boxed)).
		Render(footer)
	return lipgloss.JoinVertical(lipgloss.Left, boxed, bar)
}

func renderMainMenu(m Model) string {
	var b strings.Builder

	items := []string{"Install / Update Skills", "Dashboard", "Quit"}
	for i, item := range items {
		if i == m.cursor {
			fmt.Fprintf(&b, "  %s %s\n", cursorChar, styles.Default.Cursor.Render(item))
		} else {
			fmt.Fprintf(&b, "    %s\n", styles.Default.Dim.Render(item))
		}
	}

	if m.updateAvailable {
		fmt.Fprintf(&b, "\n%s\n",
			styles.Default.Cursor.Render(
				fmt.Sprintf("↑ asdt-tui %s available — https://github.com/vitualizz/ai-software-delivery-team/releases", m.latestVersion)))
	}

	hero := panels.RenderHero(m.currentVersion)
	menuStr := strings.TrimRight(b.String(), "\n")
	body := lipgloss.JoinVertical(lipgloss.Left, hero, "", menuStr)

	footer := panels.RenderKeyboardFooter([]panels.HintGroup{
		{Label: "Nav", Hints: []panels.Hint{{Key: "↑↓", Description: "navigate"}, {Key: "enter", Description: "select"}, {Key: "q", Description: "quit"}}},
	}, m.width)
	return frame("asdt-tui", body, footer, true)
}

func renderAssistantList(m Model) string {
	var b strings.Builder

	for i, d := range installer.Descriptors {
		bp, sp, _ := installer.Detect(d)
		present := bp && sp

		var badge panels.Badge
		if present {
			badge = panels.NewBadge("present", panels.ColorSuccess)
		} else {
			badge = panels.NewBadge("missing", panels.ColorError)
		}

		cursor := "  "
		var nameStr string
		if i == m.cursor {
			cursor = cursorChar + " "
			nameStr = styles.Default.Cursor.Render(d.Name)
		} else {
			nameStr = styles.Default.Dim.Render(d.Name)
		}

		fmt.Fprintf(&b, "  %s%s %s\n", cursor, nameStr, badge.Render())
	}

	footer := panels.RenderKeyboardFooter([]panels.HintGroup{
		{Label: "Actions", Hints: []panels.Hint{{Key: "enter", Description: "continue"}, {Key: "esc", Description: "back"}, {Key: "q", Description: "quit"}}},
	}, m.width)
	return frame("Installed Assistants", strings.TrimRight(b.String(), "\n"), footer, true)
}

func renderSelectAssistants(m Model) string {
	var b strings.Builder

	for i, d := range installer.Descriptors {
		check := "[ ]"
		if m.selected[i] {
			check = "[x]"
		}

		cursor := "  "
		var nameStr string
		if i == m.cursor {
			cursor = cursorChar + " "
			nameStr = styles.Default.Cursor.Render(d.Name)
		} else {
			nameStr = styles.Default.Dim.Render(d.Name)
		}

		fmt.Fprintf(&b, "  %s%s %s\n", cursor, check, nameStr)
	}

	footer := panels.RenderKeyboardFooter([]panels.HintGroup{
		{Label: "Actions", Hints: []panels.Hint{{Key: "space", Description: "toggle"}, {Key: "enter", Description: "confirm"}, {Key: "esc", Description: "back"}, {Key: "q", Description: "quit"}}},
	}, m.width)
	return frame("Select Assistants to Install", strings.TrimRight(b.String(), "\n"), footer, true)
}

func renderSelectProvider(m Model) string {
	var b strings.Builder

	for i, p := range installer.Providers {
		cursor := "  "
		var nameStr string
		if i == m.cursor {
			cursor = cursorChar + " "
			nameStr = styles.Default.Cursor.Render(p.Name)
		} else {
			nameStr = styles.Default.Dim.Render(p.Name)
		}

		fmt.Fprintf(&b, "  %s%s — %s\n", cursor, nameStr, p.Description)
	}

	footer := panels.RenderKeyboardFooter([]panels.HintGroup{
		{Label: "Actions", Hints: []panels.Hint{{Key: "enter", Description: "install"}, {Key: "esc", Description: "back"}, {Key: "q", Description: "quit"}}},
	}, m.width)
	return frame("Select Memory Provider", strings.TrimRight(b.String(), "\n"), footer, true)
}

func renderAgentSetup(m Model) string {
	var b strings.Builder

	// Always show a notice that this writes to global assistant config.
	if len(m.agentConflicts) > 0 {
		fmt.Fprintf(&b, "  %s\n",
			styles.Default.Warning.Render("⚠  Existing global agent config detected — will be overwritten."))
		fmt.Fprintf(&b, "  %s\n\n",
			styles.Default.Dim.Render("   Proceed at your own risk."))
	} else {
		fmt.Fprintf(&b, "  %s\n\n",
			styles.Default.Dim.Render("This will write to your global AI assistant config."))
	}

	// 4 presets + Skip.
	type option struct {
		name        string
		description string
	}
	options := make([]option, 0, len(installer.PersonaPresets)+1)
	for _, p := range installer.PersonaPresets {
		options = append(options, option{name: p.Name, description: p.Description})
	}
	options = append(options, option{name: "Skip", description: "Continue without configuring agent persona."})

	for i, opt := range options {
		cursor := "  "
		var nameStr string
		if i == m.cursor {
			cursor = cursorChar + " "
			nameStr = styles.Default.Cursor.Render(opt.name)
		} else {
			nameStr = styles.Default.Dim.Render(opt.name)
		}
		fmt.Fprintf(&b, "  %s%s — %s\n", cursor, nameStr, opt.description)
	}

	subtitle := styles.Default.Dim.Render("Configure how AI assistants behave across all tools")
	body := lipgloss.JoinVertical(lipgloss.Left, subtitle, "", strings.TrimRight(b.String(), "\n"))

	footer := panels.RenderKeyboardFooter([]panels.HintGroup{
		{Label: "Actions", Hints: []panels.Hint{{Key: "↑↓", Description: "navigate"}, {Key: "enter", Description: "select"}, {Key: "esc", Description: "back"}, {Key: "q", Description: "quit"}}},
	}, m.width)
	return frame("Agent Persona", body, footer, true)
}

func renderAgentWriteMode(m Model) string {
	var b strings.Builder

	type option struct {
		name        string
		description string
	}
	options := []option{
		{name: "Overwrite", description: "Replace the existing config entirely"},
		{name: "Append", description: "Add the new config at the end of the file"},
		{name: "Do nothing", description: "Leave the existing config untouched"},
	}

	for i, opt := range options {
		cursor := "  "
		var nameStr string
		if i == m.cursor {
			cursor = cursorChar + " "
			nameStr = styles.Default.Cursor.Render(opt.name)
		} else {
			nameStr = styles.Default.Dim.Render(opt.name)
		}
		fmt.Fprintf(&b, "  %s%s — %s\n", cursor, nameStr, opt.description)
	}

	subtitle := styles.Default.Dim.Render("Choose how to handle the existing agent config")
	body := lipgloss.JoinVertical(lipgloss.Left, subtitle, "", strings.TrimRight(b.String(), "\n"))

	footer := panels.RenderKeyboardFooter([]panels.HintGroup{
		{Label: "Actions", Hints: []panels.Hint{{Key: "↑↓", Description: "navigate"}, {Key: "enter", Description: "select"}, {Key: "esc", Description: "back"}}},
	}, m.width)
	return frame("Existing Config Detected", body, footer, true)
}

func renderInstalling(m Model) string {
	body := m.spinner.View() + " " + styles.Default.Dim.Render("Installing assistants and skills...")

	footer := panels.RenderKeyboardFooter([]panels.HintGroup{
		{Label: "Actions", Hints: []panels.Hint{{Key: "q", Description: "quit"}}},
	}, m.width)
	return frame("Installing...", body, footer, true)
}

func renderDashboard(m Model) string {
	var b strings.Builder
	for _, d := range installer.Descriptors {
		b.WriteString(renderSummaryLine(d))
		b.WriteString("\n")
	}

	footer := panels.RenderKeyboardFooter([]panels.HintGroup{
		{Label: "Nav", Hints: []panels.Hint{{Key: "esc", Description: "back to menu"}}},
	}, m.width)
	return frame("Dashboard", strings.TrimRight(b.String(), "\n"), footer, true)
}

// renderSummaryLine renders a single assistant summary line for the dashboard,
// showing the assistant name and binary/skills presence badges.
func renderSummaryLine(d installer.AssistantDescriptor) string {
	bp, sp, _ := installer.Detect(d)
	var parts []string
	parts = append(parts, d.Name)
	if bp {
		parts = append(parts, panels.NewBadge("binary", panels.ColorSuccess).Render())
	} else {
		parts = append(parts, panels.NewBadge("binary", panels.ColorError).Render())
	}
	if sp {
		parts = append(parts, panels.NewBadge("skills", panels.ColorSuccess).Render())
	} else {
		parts = append(parts, panels.NewBadge("skills", panels.ColorError).Render())
	}
	return "  " + strings.Join(parts, " ")
}

func renderDone(m Model) string {
	var b strings.Builder

	nameFor := make(map[installer.AssistantID]string, len(installer.Descriptors))
	for _, d := range installer.Descriptors {
		nameFor[d.ID] = d.Name
	}

	for _, r := range m.results {
		name := nameFor[r.AssistantID]
		if name == "" {
			name = string(r.AssistantID)
		}
		if r.Err == nil {
			fmt.Fprintf(&b, "  %s\n", styles.Default.Success.Render("✓ "+name))
		} else {
			fmt.Fprintf(&b, "  %s\n", styles.Default.Error.Render(fmt.Sprintf("✗ %s: %v", name, r.Err)))
		}
	}

	if len(m.agentResults) > 0 {
		fmt.Fprintf(&b, "\n  %s\n", styles.Default.Dim.Render("Agent Config:"))
		for _, r := range m.agentResults {
			name := nameFor[r.AssistantID]
			if name == "" {
				name = string(r.AssistantID)
			}
			switch {
			case r.Err != nil:
				fmt.Fprintf(&b, "  %s\n", styles.Default.Error.Render(fmt.Sprintf("✗ %s: %v", name, r.Err)))
			case r.Skipped:
				fmt.Fprintf(&b, "  %s\n", styles.Default.Dim.Render("– "+name+": skipped (existing config kept)"))
			default:
				fmt.Fprintf(&b, "  %s\n", styles.Default.Success.Render("✓ "+name+" agent config written"))
			}
			// Note about OpenCode override behavior.
			if r.AssistantID == installer.AssistantOpenCode && r.Err == nil && !r.Skipped {
				fmt.Fprintf(&b, "  %s\n", styles.Default.Dim.Render("  Note: OpenCode reads this as a global override for all projects."))
			}
		}
	}

	footer := panels.RenderKeyboardFooter([]panels.HintGroup{
		{Label: "Actions", Hints: []panels.Hint{{Key: "enter/esc", Description: "back to menu"}, {Key: "q", Description: "quit"}}},
	}, m.width)
	return frame("Installation Complete", strings.TrimRight(b.String(), "\n"), footer, true)
}
