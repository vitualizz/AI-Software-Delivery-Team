package installer_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/vitualizz/asdt/internal/installer"
)

func TestDetect_BothAbsent(t *testing.T) {
	d := installer.AssistantDescriptor{
		ID:         "test-assistant",
		Name:       "Test",
		BinaryName: "binary-that-does-not-exist-xyz123",
		SkillsDir:  "/tmp/skills-dir-that-does-not-exist-xyz123",
	}
	binaryPresent, skillsPresent, err := installer.Detect(d)
	if err != nil {
		t.Fatalf("Detect returned unexpected error: %v", err)
	}
	if binaryPresent {
		t.Error("binaryPresent should be false when binary is not in PATH")
	}
	if skillsPresent {
		t.Error("skillsPresent should be false when SkillsDir does not exist")
	}
}

func TestDetect_BinaryPresentSkillsAbsent(t *testing.T) {
	// Use a real binary that is always present.
	d := installer.AssistantDescriptor{
		ID:         "test-assistant",
		Name:       "Test",
		BinaryName: "sh",
		SkillsDir:  "/tmp/skills-dir-that-does-not-exist-xyz123",
	}
	binaryPresent, skillsPresent, err := installer.Detect(d)
	if err != nil {
		t.Fatalf("Detect returned unexpected error: %v", err)
	}
	if !binaryPresent {
		t.Error("binaryPresent should be true when binary is in PATH")
	}
	if skillsPresent {
		t.Error("skillsPresent should be false when SkillsDir does not exist")
	}
}

func TestDetect_BothPresent(t *testing.T) {
	dir := t.TempDir()
	skillsDir := filepath.Join(dir, "skills")
	// ASDT's own consultant directory must exist — SkillsDir alone (the
	// shared assistant skills root) is NOT sufficient evidence that ASDT
	// itself is installed (see TestDetect_SkillsRootExistsButASDTAbsent).
	if err := os.MkdirAll(filepath.Join(skillsDir, "asdt"), 0o755); err != nil {
		t.Fatal(err)
	}

	d := installer.AssistantDescriptor{
		ID:         "test-assistant",
		Name:       "Test",
		BinaryName: "sh",
		SkillsDir:  skillsDir,
	}
	binaryPresent, skillsPresent, err := installer.Detect(d)
	if err != nil {
		t.Fatalf("Detect returned unexpected error: %v", err)
	}
	if !binaryPresent {
		t.Error("binaryPresent should be true")
	}
	if !skillsPresent {
		t.Error("skillsPresent should be true when {SkillsDir}/asdt exists")
	}
}

// TestDetect_SkillsRootExistsButASDTAbsent is a regression guard for the
// SkillsDir semantic shift (SkillsDir is now the assistant's shared skills
// ROOT, not an ASDT-specific directory — see internal/installer/detect.go
// doc comment). A shared skills root commonly exists for any user of the
// assistant regardless of whether ASDT was ever installed; skillsPresent
// must keep meaning "ASDT is installed", not "the assistant has a skills
// directory at all" — otherwise the TUI's "present"/"missing" status
// (internal/setup/views.go renderAssistantList) would misreport ASDT as
// present for users who simply have OTHER skills installed.
func TestDetect_SkillsRootExistsButASDTAbsent(t *testing.T) {
	dir := t.TempDir()
	skillsDir := filepath.Join(dir, "skills")
	// The shared skills root exists (e.g. other tools installed skills here)
	// but ASDT's own "asdt" directory does NOT.
	if err := os.MkdirAll(filepath.Join(skillsDir, "some-other-skill"), 0o755); err != nil {
		t.Fatal(err)
	}

	d := installer.AssistantDescriptor{
		ID:         "test-assistant",
		Name:       "Test",
		BinaryName: "sh",
		SkillsDir:  skillsDir,
	}
	_, skillsPresent, err := installer.Detect(d)
	if err != nil {
		t.Fatalf("Detect returned unexpected error: %v", err)
	}
	if skillsPresent {
		t.Error("skillsPresent should be false when the skills root exists but ASDT's own \"asdt\" directory does not — SkillsDir existing is no longer evidence that ASDT is installed")
	}
}
