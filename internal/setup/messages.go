package setup

import (
	"io/fs"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/vitualizz/ai-software-delivery-team/internal/installer"
)

// InstallDoneMsg is sent by InstallCmd when the installer goroutine completes.
type InstallDoneMsg struct {
	Results []installer.InstallResult
}

// EngramCheckMsg is sent by EngramCheckCmd with the result of the PATH lookup.
type EngramCheckMsg struct {
	Found bool
}

// EngramCheckCmd returns a tea.Cmd that checks whether the engram binary is on
// PATH and sends EngramCheckMsg with the result.
func EngramCheckCmd() tea.Cmd {
	return func() tea.Msg {
		_, err := exec.LookPath("engram")
		return EngramCheckMsg{Found: err == nil}
	}
}

// InstallCmd returns a tea.Cmd that runs installer.Install asynchronously
// and wraps the results in InstallDoneMsg.
func InstallCmd(assistants []installer.AssistantDescriptor, provider installer.ProviderDescriptor, skillsFS fs.FS) tea.Cmd {
	return func() tea.Msg {
		results := installer.Install(assistants, provider, skillsFS)
		return InstallDoneMsg{Results: results}
	}
}
