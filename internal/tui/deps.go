// Package tui implements the Bubbletea TUI for observing .asdt/ state.
// It is a read-only delivery adapter — it never writes agent output.
package tui

import (
	"github.com/vitualizz/asdt/internal/artifact"
	"github.com/vitualizz/asdt/internal/config"
)

// Dependencies holds the injected ports the TUI requires.
// Populated by the composition root (cmd/asdt-tui/main.go) and passed to New.
type Dependencies struct {
	// ConfigRoot is the discovered .asdt/ boundary.
	ConfigRoot config.Root

	// Config holds the project-level settings (active change, etc.).
	Config config.Config

	// PipelineStore is the artifact.Store used to read pipeline-state.yaml.
	PipelineStore artifact.Store

	// Change is the name of the change to observe.
	// Falls back to Config.ActiveChange if empty.
	Change string
}
