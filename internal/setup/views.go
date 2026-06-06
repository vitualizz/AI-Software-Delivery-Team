package setup

import (
	"fmt"
	"strings"

	"github.com/vitualizz/ai-software-delivery-team/internal/installer"
)

const (
	ansiReset  = "\033[0m"
	ansiGreen  = "\033[32m"
	ansiRed    = "\033[31m"
	ansiDim    = "\033[2m"
	ansiBold   = "\033[1m"
	cursorChar = "►"
)

func renderState(m Model) string {
	switch m.state {
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

func renderMainMenu(m Model) string {
	var b strings.Builder
	b.WriteString(ansiBold + "ASDT Installer" + ansiReset + "\n\n")

	items := []string{"Install / Update Skills", "Quit"}
	for i, item := range items {
		if i == m.cursor {
			fmt.Fprintf(&b, "  %s%s %s%s\n", ansiBold, cursorChar, item, ansiReset)
		} else {
			fmt.Fprintf(&b, "  %s  %s%s\n", ansiDim, item, ansiReset)
		}
	}
	b.WriteString("\n[q] quit")
	return b.String()
}

func renderAssistantList(m Model) string {
	var b strings.Builder
	b.WriteString(ansiBold + "Installed Assistants" + ansiReset + "\n\n")

	for i, d := range installer.Descriptors {
		bp, sp, _ := installer.Detect(d)
		present := bp && sp

		label := ansiRed + "missing" + ansiReset
		if present {
			label = ansiGreen + "present" + ansiReset
		}

		cursor := "  "
		nameStyle := ansiDim
		if i == m.cursor {
			cursor = cursorChar + " "
			nameStyle = ansiBold
		}

		fmt.Fprintf(&b, "  %s%s%s%s [%s]\n", cursor, nameStyle, d.Name, ansiReset, label)
	}

	b.WriteString("\n[enter] continue  [esc] back  [q] quit")
	return b.String()
}

func renderSelectAssistants(m Model) string {
	var b strings.Builder
	b.WriteString(ansiBold + "Select Assistants to Install" + ansiReset + "\n\n")

	for i, d := range installer.Descriptors {
		check := "[ ]"
		if m.selected[i] {
			check = "[x]"
		}

		cursor := "  "
		nameStyle := ansiDim
		if i == m.cursor {
			cursor = cursorChar + " "
			nameStyle = ansiBold
		}

		fmt.Fprintf(&b, "  %s%s %s%s%s\n", cursor, check, nameStyle, d.Name, ansiReset)
	}

	b.WriteString("\n[space] toggle  [enter] confirm  [esc] back  [q] quit")
	return b.String()
}

func renderSelectProvider(m Model) string {
	var b strings.Builder
	b.WriteString(ansiBold + "Select Memory Provider" + ansiReset + "\n\n")

	for i, p := range installer.Providers {
		cursor := "  "
		nameStyle := ansiDim
		if i == m.cursor {
			cursor = cursorChar + " "
			nameStyle = ansiBold
		}

		fmt.Fprintf(&b, "  %s%s%s%s — %s\n", cursor, nameStyle, p.Name, ansiReset, p.Description)
	}

	b.WriteString("\n[enter] install  [esc] back  [q] quit")
	return b.String()
}

func renderInstalling() string {
	return ansiBold + "Installing..." + ansiReset + "\n"
}

func renderDone(m Model) string {
	var b strings.Builder
	b.WriteString(ansiBold + "Installation Complete" + ansiReset + "\n\n")

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
			fmt.Fprintf(&b, "  %s✓ %s%s\n", ansiGreen, name, ansiReset)
		} else {
			fmt.Fprintf(&b, "  %s✗ %s: %v%s\n", ansiRed, name, r.Err, ansiReset)
		}
	}

	b.WriteString("\n[enter/esc] back to menu  [q] quit")
	return b.String()
}
