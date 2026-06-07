package setup

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/vitualizz/ai-software-delivery-team/internal/installer"
	"github.com/vitualizz/ai-software-delivery-team/internal/setup/styles"
)

const cursorChar = "►"

func renderState(m Model) string {
	switch m.state {
	case StateEngramMissing:
		return renderEngramMissing()
	case StateMainMenu:
		return renderMainMenu(m)
	case StateAssistantList:
		return renderAssistantList(m)
	case StateSelectAssistants:
		return renderSelectAssistants(m)
	case StateSelectProvider:
		return renderSelectProvider(m)
	case StateInstalling:
		return renderInstalling()
	case StateDone:
		return renderDone(m)
	default:
		return ""
	}
}

// frame composes a screen from a title, a body block, and a footer hint,
// wrapping the title+body in the bordered Box and rendering the footer
// through StatusBar, joined vertically. Single source of the visual chrome.
func frame(title, body, footer string) string {
	header := styles.Default.Header.Render(title)
	inner := lipgloss.JoinVertical(lipgloss.Left, header, "", body)
	boxed := styles.Default.Box.Render(inner)
	bar := styles.Default.StatusBar.
		Width(lipgloss.Width(boxed)).
		Render(footer)
	return lipgloss.JoinVertical(lipgloss.Left, boxed, bar)
}

func renderEngramMissing() string {
	var b strings.Builder
	b.WriteString("ASDT requires the Engram MCP server to manage cross-session memory.\n")
	b.WriteString("Please install it before continuing.\n\n")
	b.WriteString("  " + styles.Default.Success.Render("https://github.com/Gentleman-Programming/engram"))
	return frame("Engram Required", b.String(), "[q] quit")
}

func renderMainMenu(m Model) string {
	var b strings.Builder

	items := []string{"Install / Update Skills", "Quit"}
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

	return frame("asdt-tui", strings.TrimRight(b.String(), "\n"), "[q] quit")
}

func renderAssistantList(m Model) string {
	var b strings.Builder

	for i, d := range installer.Descriptors {
		bp, sp, _ := installer.Detect(d)
		present := bp && sp

		var label string
		if present {
			label = styles.Default.Success.Render("present")
		} else {
			label = styles.Default.Error.Render("missing")
		}

		cursor := "  "
		var nameStr string
		if i == m.cursor {
			cursor = cursorChar + " "
			nameStr = styles.Default.Cursor.Render(d.Name)
		} else {
			nameStr = styles.Default.Dim.Render(d.Name)
		}

		fmt.Fprintf(&b, "  %s%s [%s]\n", cursor, nameStr, label)
	}

	return frame("Installed Assistants", strings.TrimRight(b.String(), "\n"), "[enter] continue  [esc] back  [q] quit")
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

	return frame("Select Assistants to Install", strings.TrimRight(b.String(), "\n"), "[space] toggle  [enter] confirm  [esc] back  [q] quit")
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

	return frame("Select Memory Provider", strings.TrimRight(b.String(), "\n"), "[enter] install  [esc] back  [q] quit")
}

func renderInstalling() string {
	return frame("Installing...", "", "[q] quit")
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

	return frame("Installation Complete", strings.TrimRight(b.String(), "\n"), "[enter/esc] back to menu  [q] quit")
}
