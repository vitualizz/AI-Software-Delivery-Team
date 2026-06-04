package tui

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/vitualizz/ai-software-delivery-team/internal/artifact"
	"github.com/vitualizz/ai-software-delivery-team/internal/config"
	"github.com/vitualizz/ai-software-delivery-team/internal/pipeline"
)

// LoadPipelineCmd reads pipeline-state.yaml via the given Store and returns
// a PipelineLoadedMsg on success or an ErrorMsg on failure.
// It is a proper tea.Cmd — it does not block the Update loop.
func LoadPipelineCmd(store artifact.Store, change string) tea.Cmd {
	return func() tea.Msg {
		if store == nil {
			return ErrorMsg{Err: fmt.Errorf("pipeline store not configured")}
		}
		var state pipeline.State
		if err := store.Read(context.Background(), change, "pipeline-state", &state); err != nil {
			return ErrorMsg{Err: err}
		}
		return PipelineLoadedMsg{State: state}
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
