package installer_test

import (
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

func TestAssistantConstants(t *testing.T) {
	if installer.AssistantClaudeCode != "claude-code" {
		t.Errorf("AssistantClaudeCode = %q, want %q", installer.AssistantClaudeCode, "claude-code")
	}
	if installer.AssistantOpenCode != "opencode" {
		t.Errorf("AssistantOpenCode = %q, want %q", installer.AssistantOpenCode, "opencode")
	}
}
