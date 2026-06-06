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

	cmd := setup.InstallCmd(assistants, provider, skillsFS)
	if cmd == nil {
		t.Error("InstallCmd returned nil tea.Cmd")
	}
}

func TestEngramCheckCmd_ReturnsNonNilCmd(t *testing.T) {
	cmd := setup.EngramCheckCmd()
	if cmd == nil {
		t.Error("EngramCheckCmd returned nil tea.Cmd")
	}
}
