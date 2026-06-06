package setup_test

import (
	"fmt"
	"strings"
	"testing"
	"testing/fstest"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/vitualizz/ai-software-delivery-team/internal/installer"
	"github.com/vitualizz/ai-software-delivery-team/internal/setup"
)

func TestView_EngramMissingScreenShowsTitle(t *testing.T) {
	m := setup.New(fstest.MapFS{})
	// State starts at StateEngramMissing (zero value before Init fires).
	view := m.View()
	if !strings.Contains(view, "Engram Required") {
		t.Errorf("engram missing view should contain 'Engram Required', got:\n%s", view)
	}
}

func TestView_EngramMissingScreenShowsURL(t *testing.T) {
	m := setup.New(fstest.MapFS{})
	view := m.View()
	if !strings.Contains(view, "github.com/Gentleman-Programming/engram") {
		t.Errorf("engram missing view should contain URL, got:\n%s", view)
	}
}

func TestView_MainMenuContainsInstallOption(t *testing.T) {
	m := setup.New(fstest.MapFS{})
	m = sendEngramFound(t, m)
	view := m.View()
	if !strings.Contains(view, "Install / Update Skills") {
		t.Errorf("main menu view missing 'Install / Update Skills', got:\n%s", view)
	}
}

func TestView_AssistantListShowsBothNames(t *testing.T) {
	m := setup.New(fstest.MapFS{})
	m = sendEngramFound(t, m)
	m2 := updateKey(t, m, tea.KeyEnter) // → AssistantList
	view := m2.View()
	for _, d := range installer.Descriptors {
		if !strings.Contains(view, d.Name) {
			t.Errorf("assistant list missing %q, view:\n%s", d.Name, view)
		}
	}
}

func TestView_AssistantListSelectedItemHasCursor(t *testing.T) {
	m := setup.New(fstest.MapFS{})
	m = sendEngramFound(t, m)
	m2 := updateKey(t, m, tea.KeyEnter) // → AssistantList; cursor=0
	view := m2.View()
	if !strings.Contains(view, "►") {
		t.Errorf("assistant list missing cursor ►, view:\n%s", view)
	}
}

func TestView_DoneScreenShowsBothAssistants(t *testing.T) {
	successResult := installer.InstallResult{AssistantID: installer.AssistantClaudeCode, Err: nil}
	failResult := installer.InstallResult{AssistantID: installer.AssistantOpenCode, Err: fmt.Errorf("fail")}
	m := setup.New(fstest.MapFS{})
	next, _ := m.Update(setup.InstallDoneMsg{Results: []installer.InstallResult{successResult, failResult}})
	m2 := next.(setup.Model)
	view := m2.View()
	if !strings.Contains(view, "Claude Code") {
		t.Errorf("done screen missing 'Claude Code', view:\n%s", view)
	}
	if !strings.Contains(view, "OpenCode") {
		t.Errorf("done screen missing 'OpenCode', view:\n%s", view)
	}
}
