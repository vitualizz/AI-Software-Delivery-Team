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
	case StateEngramMissing:
		return renderEngramMissing(m.width)
	case StateDashboard:
		return renderDashboard(m.width)
	case StateMainMenu:
		return renderMainMenu(m)
	case StateAssistantList:
		return renderAssistantList(m)
	case StateSelectAssistants:
		return renderSelectAssistants(m)
	case StateSelectProvider:
		return renderSelectProvider(m)
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

func renderEngramMissing(width int) string {
	var b strings.Builder
	b.WriteString("ASDT requires the Engram MCP server to manage cross-session memory.\n")
	b.WriteString("Please install it before continuing.\n\n")
	b.WriteString("  " + styles.Default.Success.Render("https://github.com/Gentleman-Programming/engram"))

	footer := panels.RenderKeyboardFooter([]panels.HintGroup{
		{Label: "Exit", Hints: []panels.Hint{{Key: "q", Description: "quit"}}},
	}, width)
	return frame("Engram Required", b.String(), footer, true)
}

func renderMainMenu(m Model) string {
	var b strings.Builder

	items := []string{"Dashboard", "Install / Update Skills", "Quit"}
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

	footer := panels.RenderKeyboardFooter([]panels.HintGroup{
		{Label: "Nav", Hints: []panels.Hint{{Key: "↑↓", Description: "navigate"}, {Key: "enter", Description: "select"}, {Key: "q", Description: "quit"}}},
	}, m.width)
	return frame("asdt-tui", strings.TrimRight(b.String(), "\n"), footer, true)
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

func renderInstalling(m Model) string {
	body := m.spinner.View() + " " + styles.Default.Dim.Render("Installing assistants and skills...")

	footer := panels.RenderKeyboardFooter([]panels.HintGroup{
		{Label: "Actions", Hints: []panels.Hint{{Key: "q", Description: "quit"}}},
	}, m.width)
	return frame("Installing...", body, footer, true)
}

func renderDashboard(width int) string {
	var b strings.Builder
	b.WriteString("  Dashboard — Coming Soon")

	footer := panels.RenderKeyboardFooter([]panels.HintGroup{
		{Label: "Nav", Hints: []panels.Hint{{Key: "esc", Description: "back to menu"}}},
	}, width)
	return frame("Dashboard", b.String(), footer, true)
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

	footer := panels.RenderKeyboardFooter([]panels.HintGroup{
		{Label: "Actions", Hints: []panels.Hint{{Key: "enter/esc", Description: "back to menu"}, {Key: "q", Description: "quit"}}},
	}, m.width)
	return frame("Installation Complete", strings.TrimRight(b.String(), "\n"), footer, true)
}
