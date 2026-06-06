package setup

import (
	"fmt"
	"strings"

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

func renderEngramMissing() string {
	var b strings.Builder
	b.WriteString(styles.Default.Header.Render("Engram Required") + "\n\n")
	b.WriteString("ASDT requires the Engram MCP server to manage cross-session memory.\n")
	b.WriteString("Please install it before continuing.\n\n")
	b.WriteString("  " + styles.Default.Success.Render("https://github.com/Gentleman-Programming/engram") + "\n\n")
	b.WriteString("[q] quit")
	return b.String()
}

func renderMainMenu(m Model) string {
	var b strings.Builder
	b.WriteString(styles.Default.Header.Render("ASDT Installer") + "\n\n")

	items := []string{"Install / Update Skills", "Quit"}
	for i, item := range items {
		if i == m.cursor {
			fmt.Fprintf(&b, "  %s %s\n", cursorChar, styles.Default.Cursor.Render(item))
		} else {
			fmt.Fprintf(&b, "    %s\n", styles.Default.Dim.Render(item))
		}
	}
	b.WriteString("\n[q] quit")
	return b.String()
}

func renderAssistantList(m Model) string {
	var b strings.Builder
	b.WriteString(styles.Default.Header.Render("Installed Assistants") + "\n\n")

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

	b.WriteString("\n[enter] continue  [esc] back  [q] quit")
	return b.String()
}

func renderSelectAssistants(m Model) string {
	var b strings.Builder
	b.WriteString(styles.Default.Header.Render("Select Assistants to Install") + "\n\n")

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

	b.WriteString("\n[space] toggle  [enter] confirm  [esc] back  [q] quit")
	return b.String()
}

func renderSelectProvider(m Model) string {
	var b strings.Builder
	b.WriteString(styles.Default.Header.Render("Select Memory Provider") + "\n\n")

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

	b.WriteString("\n[enter] install  [esc] back  [q] quit")
	return b.String()
}

func renderInstalling() string {
	return styles.Default.Bold.Render("Installing...") + "\n"
}

func renderDone(m Model) string {
	var b strings.Builder
	b.WriteString(styles.Default.Header.Render("Installation Complete") + "\n\n")

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

	b.WriteString("\n[enter/esc] back to menu  [q] quit")
	return b.String()
}
