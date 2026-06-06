package setup_test

import (
	"testing"
	"testing/fstest"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/vitualizz/ai-software-delivery-team/internal/installer"
	"github.com/vitualizz/ai-software-delivery-team/internal/setup"
)

func TestNew_StartsAtMainMenu(t *testing.T) {
	m := setup.New(fstest.MapFS{}, makeCfgRoot(t))
	if m.State() != setup.StateMainMenu {
		t.Errorf("New() state = %v, want StateMainMenu", m.State())
	}
}

func TestUpdate_EnterOnInstallTransitionsToAssistantList(t *testing.T) {
	m := setup.New(fstest.MapFS{}, makeCfgRoot(t))
	// Cursor starts at 0 = "Install / Update Skills".
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	next, _ := m.Update(msg)
	m2 := next.(setup.Model)
	if m2.State() != setup.StateAssistantList {
		t.Errorf("after Enter at MainMenu: state = %v, want StateAssistantList", m2.State())
	}
}

func TestUpdate_ESCAtSelectAssistantsGoesBack(t *testing.T) {
	m := setup.New(fstest.MapFS{}, makeCfgRoot(t))
	// Navigate: MainMenu → AssistantList → SelectAssistants.
	m2 := updateKey(t, m, tea.KeyEnter) // → AssistantList
	m3 := updateKey(t, m2, tea.KeyEnter) // → SelectAssistants
	if m3.State() != setup.StateSelectAssistants {
		t.Fatalf("expected StateSelectAssistants, got %v", m3.State())
	}
	m4 := updateKey(t, m3, tea.KeyEsc)
	if m4.State() != setup.StateAssistantList {
		t.Errorf("ESC at SelectAssistants: state = %v, want StateAssistantList", m4.State())
	}
}

func TestUpdate_ESCAtMainMenuIsNoop(t *testing.T) {
	m := setup.New(fstest.MapFS{}, makeCfgRoot(t))
	m2 := updateKey(t, m, tea.KeyEsc)
	if m2.State() != setup.StateMainMenu {
		t.Errorf("ESC at MainMenu: state = %v, want StateMainMenu", m2.State())
	}
}

func TestUpdate_InstallDoneMsgTransitionsToDone(t *testing.T) {
	m := setup.New(fstest.MapFS{}, makeCfgRoot(t))
	// Force state to Installing directly via the message.
	next, _ := m.Update(setup.InstallDoneMsg{Results: []installer.InstallResult{}})
	m2 := next.(setup.Model)
	if m2.State() != setup.StateDone {
		t.Errorf("InstallDoneMsg: state = %v, want StateDone", m2.State())
	}
}

// updateKey sends a key press through Update and returns the new Model.
func updateKey(t *testing.T, m setup.Model, key tea.KeyType) setup.Model {
	t.Helper()
	msg := tea.KeyMsg{Type: key}
	next, _ := m.Update(msg)
	m2, ok := next.(setup.Model)
	if !ok {
		t.Fatalf("Update returned %T, want setup.Model", next)
	}
	return m2
}
