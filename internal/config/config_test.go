package config_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/vitualizz/ai-software-delivery-team/internal/config"
)

// fixturesDir returns the absolute path to the testdata/projects directory.
func fixturesDir(t *testing.T) string {
	t.Helper()
	// Navigate from the package dir up to the module root, then into testdata.
	// The package is at internal/config; module root is ../../
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	return filepath.Join(wd, "..", "..", "testdata", "projects")
}

// TestDiscover_FoundTwoLevelsUp verifies that Discover finds .asdt/ two
// levels above the start directory (nested/src/pkg → nested/.asdt/).
func TestDiscover_FoundTwoLevelsUp(t *testing.T) {
	startDir := filepath.Join(fixturesDir(t), "nested", "src", "pkg")
	root, err := config.Discover(startDir)
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}
	want := filepath.Join(fixturesDir(t), "nested", ".asdt")
	if root.Path() != want {
		t.Errorf("Path: got %q, want %q", root.Path(), want)
	}
}

// TestDiscover_FoundAtRootLevel verifies that Discover works when .asdt/
// exists in the start directory itself.
func TestDiscover_FoundAtRootLevel(t *testing.T) {
	startDir := filepath.Join(fixturesDir(t), "nested")
	root, err := config.Discover(startDir)
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}
	want := filepath.Join(fixturesDir(t), "nested", ".asdt")
	if root.Path() != want {
		t.Errorf("Path: got %q, want %q", root.Path(), want)
	}
}

// TestDiscover_MonorepoNearestAncestor verifies that Discover returns the
// nearest .asdt/ (service-a, not service-b) when the start is inside service-a.
func TestDiscover_MonorepoNearestAncestor(t *testing.T) {
	startDir := filepath.Join(fixturesDir(t), "monorepo", "service-a", "src")
	root, err := config.Discover(startDir)
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}
	want := filepath.Join(fixturesDir(t), "monorepo", "service-a", ".asdt")
	if root.Path() != want {
		t.Errorf("Path: got %q, want %q", root.Path(), want)
	}
}

// TestDiscover_NotFound verifies that Discover returns ErrNotFound when
// no .asdt/ exists in any ancestor.
func TestDiscover_NotFound(t *testing.T) {
	// Use a temp dir with no .asdt/ anywhere in its tree.
	startDir := t.TempDir()
	_, err := config.Discover(startDir)
	if !errors.Is(err, config.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}
}

// TestLoadSave_RoundTrip verifies that Save and Load are inverses.
func TestLoadSave_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	asdt := filepath.Join(dir, ".asdt")
	if err := os.MkdirAll(asdt, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	root, err := config.Discover(dir)
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}

	original := config.Config{
		ActiveChange: "my-change",
		Defaults:     map[string]string{"provider": "mock"},
	}
	if err := config.Save(root, original); err != nil {
		t.Fatalf("Save: %v", err)
	}

	restored, err := config.Load(root)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if restored.ActiveChange != original.ActiveChange {
		t.Errorf("ActiveChange: got %q, want %q", restored.ActiveChange, original.ActiveChange)
	}
	if restored.Defaults["provider"] != original.Defaults["provider"] {
		t.Errorf("Defaults[provider]: got %q, want %q",
			restored.Defaults["provider"], original.Defaults["provider"])
	}
}

// TestLoad_MissingFile verifies that Load returns an empty Config (not an error)
// when config.yaml does not exist yet.
func TestLoad_MissingFile(t *testing.T) {
	dir := t.TempDir()
	asdt := filepath.Join(dir, ".asdt")
	if err := os.MkdirAll(asdt, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	root, err := config.Discover(dir)
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}

	cfg, err := config.Load(root)
	if err != nil {
		t.Errorf("Load of missing file should not error; got: %v", err)
	}
	if cfg.ActiveChange != "" {
		t.Errorf("expected empty Config, got ActiveChange=%q", cfg.ActiveChange)
	}
}
