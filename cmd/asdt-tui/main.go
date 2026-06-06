// Package main is the composition root for the asdt-tui binary.
// It wires the embedded skill/ FS and launches the Bubbletea installer TUI.
package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/vitualizz/ai-software-delivery-team/internal/setup"
	"github.com/vitualizz/ai-software-delivery-team/skill"
)

func main() {
	skillsFS := skill.FS()

	model := setup.New(skillsFS)
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
