package installer_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/vitualizz/ai-software-delivery-team/internal/installer"
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
	if err := os.MkdirAll(skillsDir, 0o755); err != nil {
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
		t.Error("skillsPresent should be true when SkillsDir exists")
	}
}
