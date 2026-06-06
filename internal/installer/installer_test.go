package installer_test

import (
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/vitualizz/ai-software-delivery-team/internal/installer"
)

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

	results := installer.Install(assistants, provider, testFS)

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

	results := installer.Install(assistants, provider, testFS)

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

	results1 := installer.Install(assistants, provider, testFS)
	if results1[0].Err != nil {
		t.Fatalf("first install failed: %v", results1[0].Err)
	}

	results2 := installer.Install(assistants, provider, testFS)
	if results2[0].Err != nil {
		t.Errorf("second install (idempotent) failed: %v", results2[0].Err)
	}
}

func checkFile(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err != nil {
		t.Errorf("expected file %q to exist: %v", path, err)
	}
}
