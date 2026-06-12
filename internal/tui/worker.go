package tui

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/vitualizz/asdt/internal/artifact"
	"github.com/vitualizz/asdt/internal/config"
	"github.com/vitualizz/asdt/internal/pipeline"
)

// LoadSpecialistsCmd reads pipeline-state.yaml as a StateV2 document
// and returns a SpecialistsLoadedMsg on success. If the file is absent or cannot
// be decoded as v2, it returns SpecialistsLoadedMsg with a nil State so the panel
// renders the "No specialists have run yet" placeholder gracefully.
// It is a proper tea.Cmd — it does not block the Update loop.
func LoadSpecialistsCmd(store artifact.Store, change string) tea.Cmd {
	return func() tea.Msg {
		if store == nil {
			return SpecialistsLoadedMsg{State: nil}
		}
		var state pipeline.StateV2
		if err := store.Read(context.Background(), change, "pipeline-state", &state); err != nil {
			// Missing or unreadable — return nil state so panel shows placeholder.
			return SpecialistsLoadedMsg{State: nil}
		}
		if state.SchemaVersion != pipeline.SchemaVersionV2 {
			// Wrong schema version — not a v2 document.
			return SpecialistsLoadedMsg{State: nil}
		}
		return SpecialistsLoadedMsg{State: &state}
	}
}

// LoadArtifactsCmd lists the artifact files under .asdt/artifacts/{change}/
// and returns an ArtifactListMsg with the file paths.
func LoadArtifactsCmd(root config.Root, change string) tea.Cmd {
	return func() tea.Msg {
		if root.Path() == "" {
			return ArtifactListMsg{Files: nil}
		}
		dir := filepath.Join(root.Path(), "artifacts", change)
		entries, err := os.ReadDir(dir)
		if err != nil {
			if os.IsNotExist(err) {
				return ArtifactListMsg{Files: nil}
			}
			return ErrorMsg{Err: fmt.Errorf("load artifacts: %w", err)}
		}
		var files []string
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".yaml") {
				files = append(files, filepath.Join(dir, e.Name()))
			}
		}
		return ArtifactListMsg{Files: files}
	}
}

// TickCmd returns a tea.Cmd that fires a TickMsg after 2 seconds.
// The root model re-issues it on every tick to create a polling loop.
func TickCmd() tea.Cmd {
	return tea.Tick(2*time.Second, func(_ time.Time) tea.Msg {
		return TickMsg{}
	})
}
