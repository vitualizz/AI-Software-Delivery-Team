package config_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
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

// TestLoadSave_ZeroValues verifies that a Config with zero-value fields
// round-trips correctly through Save+Load without data loss.
func TestLoadSave_ZeroValues(t *testing.T) {
	dir := t.TempDir()
	asdt := filepath.Join(dir, ".asdt")
	if err := os.MkdirAll(asdt, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	root, err := config.Discover(dir)
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}

	// Zero-value Config: no active change, nil defaults.
	original := config.Config{}
	if err := config.Save(root, original); err != nil {
		t.Fatalf("Save: %v", err)
	}

	restored, err := config.Load(root)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if restored.ActiveChange != "" {
		t.Errorf("ActiveChange: got %q, want empty", restored.ActiveChange)
	}
	if len(restored.Defaults) != 0 {
		t.Errorf("Defaults: got %v, want empty", restored.Defaults)
	}
}

// ------------------------------------------------------------------ Validate --

func TestMemoryConfig_Validate_ErrorOnEmpty(t *testing.T) {
	mc := config.MemoryConfig{Provider: ""}
	if err := mc.Validate(); err == nil {
		t.Error("expected error when Provider is empty, got nil")
	} else if !strings.Contains(err.Error(), "memory.provider") {
		t.Errorf("error message should mention memory.provider; got: %q", err.Error())
	}
}

func TestMemoryConfig_Validate_PassWhenSet(t *testing.T) {
	mc := config.MemoryConfig{Provider: "engram"}
	if err := mc.Validate(); err != nil {
		t.Errorf("expected no error when Provider is set, got: %v", err)
	}
}

func TestConfig_Validate_Delegates(t *testing.T) {
	// Config with empty memory provider should propagate MemoryConfig error.
	cfg := config.Config{}
	if err := cfg.Validate(); err == nil {
		t.Error("Config.Validate: expected error when memory.provider is empty, got nil")
	}
	// Config with memory provider set should pass.
	cfg.Memory = config.MemoryConfig{Provider: "engram"}
	if err := cfg.Validate(); err != nil {
		t.Errorf("Config.Validate: expected no error when memory.provider is set, got: %v", err)
	}
}

// TestDiscover_StopsAtFilesystemRoot verifies that Discover returns ErrNotFound
// when walking from a temp directory that has no .asdt/ anywhere in its ancestry.
// This exercises the "reached root" stop condition.
func TestDiscover_StopsAtFilesystemRoot(t *testing.T) {
	// t.TempDir() is always under /tmp or equivalent — no .asdt/ will exist there.
	startDir := t.TempDir()
	_, err := config.Discover(startDir)
	if !errors.Is(err, config.ErrNotFound) {
		t.Errorf("expected ErrNotFound when no .asdt/ exists; got: %v", err)
	}
}

// TestDiscover_NotFound_ErrorMessageIsHelpful verifies that the ErrNotFound error
// contains guidance for the user.
func TestDiscover_NotFound_ErrorMessageIsHelpful(t *testing.T) {
	startDir := t.TempDir()
	_, err := config.Discover(startDir)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	msg := err.Error()
	if !strings.Contains(msg, ".asdt") {
		t.Errorf("error message should mention .asdt: %q", msg)
	}
}
