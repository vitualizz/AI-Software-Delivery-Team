package setup

import (
	"io/fs"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/vitualizz/ai-software-delivery-team/internal/config"
	"github.com/vitualizz/ai-software-delivery-team/internal/installer"
)

// InstallDoneMsg is sent by installCmd when the installer goroutine completes.
type InstallDoneMsg struct {
	Results []installer.InstallResult
}

// InstallCmd returns a tea.Cmd that runs installer.Install asynchronously
// and wraps the results in InstallDoneMsg.
func InstallCmd(assistants []installer.AssistantDescriptor, provider installer.ProviderDescriptor, skillsFS fs.FS, cfgRoot config.Root) tea.Cmd {
	return func() tea.Msg {
		results := installer.Install(assistants, provider, skillsFS, cfgRoot)
		return InstallDoneMsg{Results: results}
	}
}
