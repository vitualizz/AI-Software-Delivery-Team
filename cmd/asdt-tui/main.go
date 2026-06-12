// Package main is the composition root for the asdt-tui binary.
// It wires the embedded skill/ FS and launches the Bubbletea installer TUI.
package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/vitualizz/asdt/internal/setup"
	"github.com/vitualizz/asdt/skill"
)

// version is the running binary's version string. It defaults to "dev" for
// local builds and is stamped at release time via linker flags
// (-X main.version={{ .Version }}, see .goreleaser.yaml).
var version = "dev"

func main() {
	skillsFS := skill.FS()

	model := setup.New(skillsFS, version)
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
