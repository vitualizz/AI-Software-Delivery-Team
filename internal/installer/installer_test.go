package installer_test

import (
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/vitualizz/ai-software-delivery-team/internal/config"
	"github.com/vitualizz/ai-software-delivery-team/internal/installer"
)

func makeConfigRoot(t *testing.T) config.Root {
	t.Helper()
	dir := t.TempDir()
	asdtDir := filepath.Join(dir, ".asdt")
	if err := os.MkdirAll(asdtDir, 0o755); err != nil {
		t.Fatal(err)
	}
	root, err := config.Discover(dir)
	if err != nil {
		t.Fatalf("config.Discover: %v", err)
	}
	return root
}

var testFS = fstest.MapFS{
	"architect/SKILL.md": &fstest.MapFile{Data: []byte("# Architect skill")},
	"developer/SKILL.md": &fstest.MapFile{Data: []byte("# Developer skill")},
}

func TestInstall_SuccessForTwoAssistants(t *testing.T) {
	dir1 := t.TempDir()
	dir2 := t.TempDir()
	assistants := []installer.AssistantDescriptor{
		{ID: "a1", Name: "A1", BinaryName: "sh", SkillsDir: filepath.Join(dir1, "skills")},
		{ID: "a2", Name: "A2", BinaryName: "sh", SkillsDir: filepath.Join(dir2, "skills")},
	}
	provider := installer.Providers[0] // engram
	root := makeConfigRoot(t)

	results := installer.Install(assistants, provider, testFS, root)

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	for i, r := range results {
		if r.Err != nil {
			t.Errorf("result[%d].Err = %v, want nil", i, r.Err)
		}
	}

	// Check files exist.
	checkFile(t, filepath.Join(dir1, "skills", "architect", "SKILL.md"))
	checkFile(t, filepath.Join(dir1, "skills", "developer", "SKILL.md"))
	checkFile(t, filepath.Join(dir2, "skills", "architect", "SKILL.md"))
	checkFile(t, filepath.Join(dir2, "skills", "developer", "SKILL.md"))
}

func TestInstall_PartialFailure(t *testing.T) {
	// Create an unwritable directory for the first assistant.
	dir1 := t.TempDir()
	unwritable := filepath.Join(dir1, "nope")
	if err := os.MkdirAll(unwritable, 0o555); err != nil { // read+execute, no write
		t.Fatal(err)
	}

	dir2 := t.TempDir()
	assistants := []installer.AssistantDescriptor{
		{ID: "a1", Name: "A1", BinaryName: "sh", SkillsDir: filepath.Join(unwritable, "skills")},
		{ID: "a2", Name: "A2", BinaryName: "sh", SkillsDir: filepath.Join(dir2, "skills")},
	}
	provider := installer.Providers[0]
	root := makeConfigRoot(t)

	results := installer.Install(assistants, provider, testFS, root)

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if results[0].Err == nil {
		t.Error("results[0].Err should be non-nil for unwritable target")
	}
	if results[1].Err != nil {
		t.Errorf("results[1].Err = %v, want nil", results[1].Err)
	}
}

func TestInstall_Idempotent(t *testing.T) {
	dir := t.TempDir()
	assistants := []installer.AssistantDescriptor{
		{ID: "a1", Name: "A1", BinaryName: "sh", SkillsDir: filepath.Join(dir, "skills")},
	}
	provider := installer.Providers[0]
	root := makeConfigRoot(t)

	results1 := installer.Install(assistants, provider, testFS, root)
	if results1[0].Err != nil {
		t.Fatalf("first install failed: %v", results1[0].Err)
	}

	results2 := installer.Install(assistants, provider, testFS, root)
	if results2[0].Err != nil {
		t.Errorf("second install (idempotent) failed: %v", results2[0].Err)
	}
}

func TestInstall_ConfigWritten(t *testing.T) {
	dir := t.TempDir()
	assistants := []installer.AssistantDescriptor{
		{ID: "a1", Name: "A1", BinaryName: "sh", SkillsDir: filepath.Join(dir, "skills")},
	}
	provider := installer.Providers[0] // engram
	root := makeConfigRoot(t)

	results := installer.Install(assistants, provider, testFS, root)
	if results[0].Err != nil {
		t.Fatalf("install failed: %v", results[0].Err)
	}

	cfg, err := config.Load(root)
	if err != nil {
		t.Fatalf("config.Load: %v", err)
	}
	if cfg.Memory.Provider != installer.ProviderEngram {
		t.Errorf("config.Memory.Provider = %q, want %q", cfg.Memory.Provider, installer.ProviderEngram)
	}
}

func TestInstall_PreservesExistingConfigFields(t *testing.T) {
	dir := t.TempDir()
	root := makeConfigRoot(t)

	// Pre-populate config with an existing field.
	existing, err := config.Load(root)
	if err != nil {
		t.Fatalf("config.Load: %v", err)
	}
	existing.ActiveChange = "my-existing-change"
	if err := config.Save(root, existing); err != nil {
		t.Fatalf("config.Save pre-populate: %v", err)
	}

	assistants := []installer.AssistantDescriptor{
		{ID: "a1", Name: "A1", BinaryName: "sh", SkillsDir: filepath.Join(dir, "skills")},
	}
	results := installer.Install(assistants, installer.Providers[0], testFS, root)
	if results[0].Err != nil {
		t.Fatalf("install failed: %v", results[0].Err)
	}

	cfg, err := config.Load(root)
	if err != nil {
		t.Fatalf("config.Load post-install: %v", err)
	}
	if cfg.ActiveChange != "my-existing-change" {
		t.Errorf("ActiveChange = %q, want %q", cfg.ActiveChange, "my-existing-change")
	}
	if cfg.Memory.Provider != installer.ProviderEngram {
		t.Errorf("Memory.Provider = %q, want %q", cfg.Memory.Provider, installer.ProviderEngram)
	}
}

func checkFile(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err != nil {
		t.Errorf("expected file %q to exist: %v", path, err)
	}
}
