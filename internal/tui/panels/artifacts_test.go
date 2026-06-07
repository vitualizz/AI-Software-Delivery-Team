package panels_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/vitualizz/ai-software-delivery-team/internal/tui/panels"
)

// TestArtifactPanelEmptyFilesRendersPlaceholder verifies that a panel with no
// loaded files renders the "No artifacts found" placeholder.
func TestArtifactPanelEmptyFilesRendersPlaceholder(t *testing.T) {
	p := panels.NewArtifactPanel()
	view := p.View()

	if !strings.Contains(view, "No artifacts found") {
		t.Errorf("expected 'No artifacts found' in view, got: %q", view)
	}
}

// TestArtifactPanelRendersAllFiles verifies that a panel with 3 files renders
// all 3 file names in its view.
func TestArtifactPanelRendersAllFiles(t *testing.T) {
	p := panels.NewArtifactPanel()
	p, _ = p.UpdateSize(80, 40)

	files := []string{
		"/tmp/testasdt/requirements-spec.yaml",
		"/tmp/testasdt/implementation-plan.yaml",
		"/tmp/testasdt/pipeline-state.yaml",
	}
	p, _ = p.SetFiles(files)
	view := p.View()

	for _, f := range files {
		name := strings.TrimSuffix(filepath.Base(f), ".yaml")
		if !strings.Contains(view, name) {
			t.Errorf("expected %q in view, got: %q", name, view)
		}
	}
}

// TestArtifactPanelJKNavigation verifies that j/k key navigation changes
// the selected index.
func TestArtifactPanelJKNavigation(t *testing.T) {
	p := panels.NewArtifactPanel()
	p, _ = p.UpdateSize(80, 40)

	files := []string{
		"/tmp/test/alpha.yaml",
		"/tmp/test/beta.yaml",
		"/tmp/test/gamma.yaml",
	}
	p, _ = p.SetFiles(files)

	// Initial selection should be 0 — first item has the active indicator.
	view := p.View()
	if !strings.Contains(view, "▶ alpha") {
		t.Errorf("expected alpha to be selected initially, got: %q", view)
	}

	// Press j — should move to beta.
	p, _ = p.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	view = p.View()
	if !strings.Contains(view, "▶ beta") {
		t.Errorf("expected beta to be selected after j, got: %q", view)
	}

	// Press k — should go back to alpha.
	p, _ = p.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	view = p.View()
	if !strings.Contains(view, "▶ alpha") {
		t.Errorf("expected alpha to be selected after k, got: %q", view)
	}
}

// TestArtifactPanelEnterLoadsContent verifies that pressing Enter on a selected
// file loads its content into the viewport.
func TestArtifactPanelEnterLoadsContent(t *testing.T) {
	// Create a temp YAML file.
	dir := t.TempDir()
	yamlContent := "schema_version: \"1\"\nagent: test\nchange_id: my-change\n"
	filePath := filepath.Join(dir, "requirements-spec.yaml")
	if err := os.WriteFile(filePath, []byte(yamlContent), 0o644); err != nil {
		t.Fatal(err)
	}

	p := panels.NewArtifactPanel()
	p, _ = p.UpdateSize(80, 40)
	p, _ = p.SetFiles([]string{filePath})

	// Press Enter to load the file.
	p, _ = p.Update(tea.KeyMsg{Type: tea.KeyEnter})
	view := p.View()

	if !strings.Contains(view, "schema_version") {
		t.Errorf("expected YAML content in view after enter, got: %q", view)
	}
}

// TestArtifactPanelCompactMode verifies that at <=60 width the panel
// renders without errors in compact mode.
func TestArtifactPanelCompactMode(t *testing.T) {
	p := panels.NewArtifactPanel()
	p, _ = p.UpdateSize(50, 30)

	// With no files set, should show placeholder.
	view := p.View()

	if !strings.Contains(view, "Artifacts") {
		t.Errorf("expected Artifacts content in compact view, got: %q", view)
	}
}

// TestArtifactPanelFileTreeIndent verifies that hierarchical file paths
// render with tree indentation (box-drawing characters).
func TestArtifactPanelFileTreeIndent(t *testing.T) {
	p := panels.NewArtifactPanel()
	p, _ = p.UpdateSize(80, 40)
	p, _ = p.SetFiles([]string{
		"/root/src/a/file1.yaml",
		"/root/src/b/file2.yaml",
	})
	view := p.View()

	if !strings.Contains(view, "└──") {
		t.Errorf("expected tree indentation (├──/└──) for hierarchical paths, got: %q", view)
	}
}

// TestArtifactPanelSeparatorUsesThinnerChar verifies the separator between
// file list and content viewer uses the thinner '╌' character.
func TestArtifactPanelSeparatorUsesThinnerChar(t *testing.T) {
	p := panels.NewArtifactPanel()
	p, _ = p.UpdateSize(80, 40)
	p, _ = p.SetFiles([]string{"/tmp/test/file.yaml"})
	view := p.View()

	if !strings.Contains(view, "╌") {
		t.Errorf("expected thinner separator '╌' character, got: %q", view)
	}
}
