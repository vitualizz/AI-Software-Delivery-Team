// Package main is the composition root for the asdt-tui binary.
// It wires the embedded skill/ FS and launches the Bubbletea installer TUI.
package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/vitualizz/ai-software-delivery-team/internal/config"
	"github.com/vitualizz/ai-software-delivery-team/internal/setup"
	"github.com/vitualizz/ai-software-delivery-team/skill"
)

func main() {
	skillsFS := skill.FS()

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	cfgRoot, err := config.Discover(cwd)
	if err != nil {
		// No .asdt/ found — create a zero root in cwd for the TUI to use.
		asdtPath := cwd + "/.asdt"
		_ = os.MkdirAll(asdtPath, 0o755)
		cfgRoot, err = config.Discover(cwd)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	}

	model := setup.New(skillsFS, cfgRoot)
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
