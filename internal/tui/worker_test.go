package tui_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/vitualizz/ai-software-delivery-team/internal/config"
	"github.com/vitualizz/ai-software-delivery-team/internal/tui"
	"gopkg.in/yaml.v3"
)

// TestLoadArtifactsCmdSuccess verifies that LoadArtifactsCmd returns an
// ArtifactListMsg with the correct file count from a fixture directory.
func TestLoadArtifactsCmdSuccess(t *testing.T) {
	// Build a real .asdt/ directory structure.
	dir := t.TempDir()
	asdt := filepath.Join(dir, ".asdt")
	change := "my-change"
	artifactsDir := filepath.Join(asdt, "artifacts", change)
	if err := os.MkdirAll(artifactsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Write two fixture YAML files.
	for _, name := range []string{"requirements-spec.yaml", "implementation-plan.yaml"} {
		data, _ := yaml.Marshal(map[string]string{"test": "data"})
		if err := os.WriteFile(filepath.Join(artifactsDir, name), data, 0o644); err != nil {
			t.Fatal(err)
		}
	}

	root, err := config.Discover(dir)
	if err != nil {
		t.Fatalf("Discover failed: %v", err)
	}

	cmd := tui.LoadArtifactsCmd(root, change)
	if cmd == nil {
		t.Fatal("LoadArtifactsCmd returned nil")
	}

	msg := cmd()
	listed, ok := msg.(tui.ArtifactListMsg)
	if !ok {
		t.Fatalf("expected ArtifactListMsg, got %T: %v", msg, msg)
	}
	if len(listed.Files) != 2 {
		t.Errorf("expected 2 files, got %d: %v", len(listed.Files), listed.Files)
	}
}

// TestLoadArtifactsCmdEmptyDir verifies that LoadArtifactsCmd returns an
// ArtifactListMsg with nil/empty files when the directory does not exist.
func TestLoadArtifactsCmdEmptyDir(t *testing.T) {
	dir := t.TempDir()
	asdt := filepath.Join(dir, ".asdt")
	if err := os.MkdirAll(asdt, 0o755); err != nil {
		t.Fatal(err)
	}

	root, err := config.Discover(dir)
	if err != nil {
		t.Fatalf("Discover failed: %v", err)
	}

	cmd := tui.LoadArtifactsCmd(root, "nonexistent-change")
	msg := cmd()

	listed, ok := msg.(tui.ArtifactListMsg)
	if !ok {
		t.Fatalf("expected ArtifactListMsg, got %T", msg)
	}
	if len(listed.Files) != 0 {
		t.Errorf("expected 0 files for missing directory, got %d", len(listed.Files))
	}
}

// TestTickCmdIsNonNil verifies that TickCmd returns a non-nil tea.Cmd.
func TestTickCmdIsNonNil(t *testing.T) {
	cmd := tui.TickCmd()
	if cmd == nil {
		t.Fatal("TickCmd returned nil")
	}
}

// TestTickCmdFiresTickMsg verifies that TickCmd eventually fires a TickMsg.
// We use a generous timeout since tea.Tick fires after 2 seconds.
func TestTickCmdFiresTickMsg(t *testing.T) {
	t.Skip("skipping slow tick test in normal suite — TickCmd fires after 2s")

	cmd := tui.TickCmd()
	done := make(chan struct{})
	go func() {
		msg := cmd()
		if _, ok := msg.(tui.TickMsg); !ok {
			t.Errorf("expected TickMsg, got %T", msg)
		}
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Error("TickCmd did not fire within 3 seconds")
	}
}
