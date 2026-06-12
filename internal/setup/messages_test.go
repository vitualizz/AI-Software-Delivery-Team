package setup_test

import (
	"io/fs"
	"testing"
	"testing/fstest"

	"github.com/vitualizz/ai-software-delivery-team/internal/installer"
	"github.com/vitualizz/ai-software-delivery-team/internal/setup"
)

func TestInstallDoneMsg_HasResultsField(t *testing.T) {
	results := []installer.InstallResult{
		{AssistantID: "test", Err: nil},
	}
	msg := setup.InstallDoneMsg{Results: results}
	if len(msg.Results) != 1 {
		t.Errorf("Results length = %d, want 1", len(msg.Results))
	}
}

func TestInstallCmd_ReturnsNonNilCmd(t *testing.T) {
	var skillsFS fs.FS = fstest.MapFS{}
	assistants := installer.Descriptors
	provider := installer.Providers[0]

	cmd := setup.InstallCmd(assistants, provider, skillsFS, "en", nil)
	if cmd == nil {
		t.Error("InstallCmd returned nil tea.Cmd")
	}
}

func TestAgentInstallCmd_ReturnsNonNilCmd(t *testing.T) {
	var skillsFS fs.FS = fstest.MapFS{}
	cmd := setup.AgentInstallCmd(installer.Descriptors, installer.PersonaPresets[0], true, map[string]installer.AgentWriteMode{}, skillsFS)
	if cmd == nil {
		t.Error("AgentInstallCmd returned nil tea.Cmd")
	}
}

func TestUpdateCheckCmd_ReturnsNonNilCmd(t *testing.T) {
	cmd := setup.UpdateCheckCmd("dev")
	if cmd == nil {
		t.Error("UpdateCheckCmd returned nil tea.Cmd")
	}
}
