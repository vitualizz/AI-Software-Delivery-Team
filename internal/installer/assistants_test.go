package installer_test

import (
	"strings"
	"testing"

	"github.com/vitualizz/ai-software-delivery-team/internal/installer"
)

func TestDescriptors_Length(t *testing.T) {
	if len(installer.Descriptors) != 2 {
		t.Fatalf("expected 2 descriptors, got %d", len(installer.Descriptors))
	}
}

func TestDescriptors_Fields(t *testing.T) {
	for _, d := range installer.Descriptors {
		if d.ID == "" {
			t.Errorf("descriptor has empty ID")
		}
		if d.Name == "" {
			t.Errorf("descriptor %q has empty Name", d.ID)
		}
		if d.BinaryName == "" {
			t.Errorf("descriptor %q has empty BinaryName", d.ID)
		}
		if d.SkillsDir == "" {
			t.Errorf("descriptor %q has empty SkillsDir", d.ID)
		}
	}
}

// TestDescriptors_SkillsDirIsSiblingRoot verifies SkillsDir points at the
// assistant's skills ROOT (e.g. ~/.claude/skills), not a nested "asdt"
// subdirectory. The installer is responsible for deriving per-entry sibling
// destinations from this root — the descriptor itself must not pre-nest.
func TestDescriptors_SkillsDirIsSiblingRoot(t *testing.T) {
	for _, d := range installer.Descriptors {
		if strings.HasSuffix(d.SkillsDir, "/asdt") {
			t.Errorf("descriptor %q SkillsDir = %q, must NOT end in /asdt — it must be the skills root so the installer can map entries to top-level siblings", d.ID, d.SkillsDir)
		}
		if !strings.HasSuffix(d.SkillsDir, "/skills") {
			t.Errorf("descriptor %q SkillsDir = %q, want it to end in /skills (the assistant's skills root)", d.ID, d.SkillsDir)
		}
	}
}

func TestAssistantConstants(t *testing.T) {
	if installer.AssistantClaudeCode != "claude-code" {
		t.Errorf("AssistantClaudeCode = %q, want %q", installer.AssistantClaudeCode, "claude-code")
	}
	if installer.AssistantOpenCode != "opencode" {
		t.Errorf("AssistantOpenCode = %q, want %q", installer.AssistantOpenCode, "opencode")
	}
}
