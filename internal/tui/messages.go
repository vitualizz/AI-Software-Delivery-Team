package tui

import "github.com/vitualizz/asdt/internal/pipeline"

// SpecialistsLoadedMsg is sent when pipeline-state.yaml (v2) is successfully read.
type SpecialistsLoadedMsg struct {
	State *pipeline.StateV2
}

// ArtifactListMsg is sent when the artifact directory listing completes.
type ArtifactListMsg struct {
	Files []string
}

// ArtifactContentMsg is sent when a specific artifact file's YAML is loaded.
type ArtifactContentMsg struct {
	Path    string
	Content string
}

// ErrorMsg is sent when any background command encounters an error.
type ErrorMsg struct {
	Err error
}

// TickMsg is sent by TickCmd to trigger a polling reload.
type TickMsg struct{}
