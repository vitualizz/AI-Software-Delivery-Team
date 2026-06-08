package setup_test

import (
	"errors"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/vitualizz/ai-software-delivery-team/internal/installer"
	"github.com/vitualizz/ai-software-delivery-team/internal/setup"
)

func TestNew_StartsAtEngramMissing(t *testing.T) {
	m := setup.New(fstest.MapFS{}, "dev")
	// Before Init() fires (no EngramCheckMsg yet), zero value is StateEngramMissing.
	if m.State() != setup.StateEngramMissing {
		t.Errorf("New() state = %v, want StateEngramMissing (%v)", m.State(), setup.StateEngramMissing)
	}
}

func TestUpdate_EngramCheckMsg_FoundTrue_TransitionsToMainMenu(t *testing.T) {
	m := setup.New(fstest.MapFS{}, "dev")
	next, _ := m.Update(setup.EngramCheckMsg{Found: true})
	m2 := next.(setup.Model)
	if m2.State() != setup.StateMainMenu {
		t.Errorf("EngramCheckMsg{Found:true}: state = %v, want StateMainMenu", m2.State())
	}
}

func TestUpdate_EngramCheckMsg_FoundFalse_StaysAtEngramMissing(t *testing.T) {
	m := setup.New(fstest.MapFS{}, "dev")
	next, _ := m.Update(setup.EngramCheckMsg{Found: false})
	m2 := next.(setup.Model)
	if m2.State() != setup.StateEngramMissing {
		t.Errorf("EngramCheckMsg{Found:false}: state = %v, want StateEngramMissing", m2.State())
	}
}

func TestUpdate_EngramMissing_NonQuitKeysAreNoop(t *testing.T) {
	m := setup.New(fstest.MapFS{}, "dev")
	// State starts at StateEngramMissing; send a key that should be no-op.
	m2 := updateKey(t, m, tea.KeyEnter)
	if m2.State() != setup.StateEngramMissing {
		t.Errorf("Enter in StateEngramMissing: state = %v, want StateEngramMissing", m2.State())
	}
}

func TestUpdate_EngramMissing_QQuits(t *testing.T) {
	m := setup.New(fstest.MapFS{}, "dev")
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
	_, cmd := m.Update(msg)
	if cmd == nil {
		t.Error("q in StateEngramMissing: expected non-nil cmd (tea.Quit), got nil")
	}
}

func TestUpdate_EngramMissing_CtrlCQuits(t *testing.T) {
	m := setup.New(fstest.MapFS{}, "dev")
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	if cmd == nil {
		t.Error("ctrl+c in StateEngramMissing: expected non-nil cmd (tea.Quit), got nil")
	}
}

func TestUpdate_SpaceToggleSelectsItem(t *testing.T) {
	m := setup.New(fstest.MapFS{}, "dev")
	// Navigate to StateSelectAssistants.
	m = sendEngramFound(t, m)         // → StateMainMenu
	m = updateKey(t, m, tea.KeyDown)  // cursor → 1 (Install)
	m = updateKey(t, m, tea.KeyEnter) // → StateAssistantList
	m = updateKey(t, m, tea.KeyEnter) // → StateSelectAssistants

	if m.State() != setup.StateSelectAssistants {
		t.Fatalf("expected StateSelectAssistants, got %v", m.State())
	}

	// Space should toggle cursor item on.
	m2 := updateKeyMsg(t, m, tea.KeyMsg{Type: tea.KeySpace})
	// Navigate to done to read selected — check via install sequence instead.
	// The simplest check: send space again to toggle off, verifying toggle behavior.
	m3 := updateKeyMsg(t, m2, tea.KeyMsg{Type: tea.KeySpace})
	_ = m3
	// If we get here without panic the space handler ran. Verify state unchanged.
	if m2.State() != setup.StateSelectAssistants {
		t.Errorf("after space: state = %v, want StateSelectAssistants", m2.State())
	}
}

func TestUpdate_EnterOnInstallTransitionsToAssistantList(t *testing.T) {
	m := setup.New(fstest.MapFS{}, "dev")
	m = sendEngramFound(t, m) // → StateMainMenu
	// Cursor starts at 0 = "Dashboard". Move down to cursor 1 = "Install / Update Skills".
	m = updateKey(t, m, tea.KeyDown)
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	next, _ := m.Update(msg)
	m2 := next.(setup.Model)
	if m2.State() != setup.StateAssistantList {
		t.Errorf("after Enter at cursor 1 (Install): state = %v, want StateAssistantList", m2.State())
	}
}

func TestUpdate_EnterOnDashboardTransitionsToDashboard(t *testing.T) {
	m := setup.New(fstest.MapFS{}, "dev")
	m = sendEngramFound(t, m) // → StateMainMenu, cursor=0
	// Cursor at 0 = "Dashboard".
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m2 := next.(setup.Model)
	if m2.State() != setup.StateDashboard {
		t.Errorf("after Enter at cursor 0 (Dashboard): state = %v, want StateDashboard", m2.State())
	}
}

func TestUpdate_EscFromDashboardReturnsToMenu(t *testing.T) {
	m := setup.New(fstest.MapFS{}, "dev")
	m = sendEngramFound(t, m)         // → StateMainMenu
	m = updateKey(t, m, tea.KeyEnter) // → StateDashboard
	m2 := updateKey(t, m, tea.KeyEsc) // Esc → back
	if m2.State() != setup.StateMainMenu {
		t.Errorf("Esc from Dashboard: state = %v, want StateMainMenu", m2.State())
	}
}

func TestUpdate_DashboardShowsComingSoon(t *testing.T) {
	m := setup.New(fstest.MapFS{}, "dev")
	m = sendEngramFound(t, m)         // → StateMainMenu
	m = updateKey(t, m, tea.KeyEnter) // → StateDashboard
	view := m.View()
	if !strings.Contains(view, "Coming Soon") {
		t.Errorf("Dashboard view missing 'Coming Soon', got:\n%s", view)
	}
}

func TestUpdate_ESCAtSelectAssistantsGoesBack(t *testing.T) {
	m := setup.New(fstest.MapFS{}, "dev")
	m = sendEngramFound(t, m) // → StateMainMenu
	// Navigate: MainMenu → AssistantList → SelectAssistants.
	m = updateKey(t, m, tea.KeyDown)     // cursor → 1 (Install)
	m2 := updateKey(t, m, tea.KeyEnter)  // → AssistantList
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
	m := setup.New(fstest.MapFS{}, "dev")
	m = sendEngramFound(t, m) // → StateMainMenu
	m2 := updateKey(t, m, tea.KeyEsc)
	if m2.State() != setup.StateMainMenu {
		t.Errorf("ESC at MainMenu: state = %v, want StateMainMenu", m2.State())
	}
}

func TestUpdate_InstallDoneMsgTransitionsToDone(t *testing.T) {
	m := setup.New(fstest.MapFS{}, "dev")
	// Force state to Installing directly via the message.
	next, _ := m.Update(setup.InstallDoneMsg{Results: []installer.InstallResult{}})
	m2 := next.(setup.Model)
	if m2.State() != setup.StateDone {
		t.Errorf("InstallDoneMsg: state = %v, want StateDone", m2.State())
	}
}

func TestUpdate_UpdateCheckMsg_Newer_SetsBanner(t *testing.T) {
	m := setup.New(fstest.MapFS{}, "v0.2.0")
	m = sendEngramFound(t, m) // → StateMainMenu

	next, _ := m.Update(setup.UpdateCheckMsg{Current: "v0.2.0", Latest: "v0.3.0"})
	m2 := next.(setup.Model)

	view := m2.View()
	if !strings.Contains(view, "v0.3.0") {
		t.Errorf("expected view to contain latest tag %q, got:\n%s", "v0.3.0", view)
	}
	if !strings.Contains(view, "/releases") {
		t.Errorf("expected view to contain releases URL, got:\n%s", view)
	}
}

func TestUpdate_UpdateCheckMsg_Error_NoBanner(t *testing.T) {
	m := setup.New(fstest.MapFS{}, "v0.2.0")
	m = sendEngramFound(t, m) // → StateMainMenu

	next, _ := m.Update(setup.UpdateCheckMsg{Err: errors.New("boom")})
	m2 := next.(setup.Model)

	view := m2.View()
	if strings.Contains(view, "/releases") {
		t.Errorf("expected no banner on error, got:\n%s", view)
	}
}

func TestUpdate_UpdateCheckMsg_DevBuild_NoBanner(t *testing.T) {
	m := setup.New(fstest.MapFS{}, "dev")
	m = sendEngramFound(t, m) // → StateMainMenu

	next, _ := m.Update(setup.UpdateCheckMsg{Current: "dev", Latest: "v9.9.9"})
	m2 := next.(setup.Model)

	view := m2.View()
	if strings.Contains(view, "/releases") {
		t.Errorf("expected no banner on dev build, got:\n%s", view)
	}
}

// toInstalling drives the model from StateEngramMissing through the menu
// flow to StateInstalling — the same Update transitions the real TUI uses
// (mirrors views_test.go's stateView helper, but returns setup.Model so
// callers can keep driving Update directly).
func toInstalling(t *testing.T, m setup.Model) setup.Model {
	t.Helper()
	m = sendEngramFound(t, m)
	m = updateKey(t, m, tea.KeyDown)  // cursor → 1 (Install)
	m = updateKey(t, m, tea.KeyEnter) // MainMenu → AssistantList
	m = updateKey(t, m, tea.KeyEnter) // AssistantList → SelectAssistants
	m = updateKey(t, m, tea.KeyEnter) // SelectAssistants → SelectProvider
	m = updateKey(t, m, tea.KeyEnter) // SelectProvider → Installing
	if m.State() != setup.StateInstalling {
		t.Fatalf("toInstalling: state = %v, want StateInstalling", m.State())
	}
	return m
}

// TestUpdate_SpinnerTickAdvancesWhileInstalling verifies that a
// spinner.TickMsg received while in StateInstalling is routed to the
// embedded spinner and produces a re-tick command — keeping the
// indeterminate spinner animating for the duration of the install.
func TestUpdate_SpinnerTickAdvancesWhileInstalling(t *testing.T) {
	m := setup.New(fstest.MapFS{}, "dev")
	m = toInstalling(t, m)

	next, cmd := m.Update(spinner.TickMsg{})
	if cmd == nil {
		t.Fatal("expected a re-tick command while StateInstalling, got nil")
	}
	msg := cmd()
	if _, ok := msg.(spinner.TickMsg); !ok {
		t.Errorf("expected re-emitted command to produce spinner.TickMsg, got %T", msg)
	}

	m2, ok := next.(setup.Model)
	if !ok {
		t.Fatalf("Update returned %T, want setup.Model", next)
	}
	if m2.State() != setup.StateInstalling {
		t.Errorf("expected state to remain StateInstalling after spinner tick, got %v", m2.State())
	}
}

// TestUpdate_SpinnerTickIgnoredOutsideInstalling verifies that a
// spinner.TickMsg received OUTSIDE StateInstalling is gated off — it must not
// produce a re-tick command, so the spinner loop does not run (and animate)
// in screens where it has no visual representation.
func TestUpdate_SpinnerTickIgnoredOutsideInstalling(t *testing.T) {
	m := setup.New(fstest.MapFS{}, "dev")
	m = sendEngramFound(t, m) // → StateMainMenu (not Installing)

	_, cmd := m.Update(spinner.TickMsg{})
	if cmd != nil {
		msg := cmd()
		if _, ok := msg.(spinner.TickMsg); ok {
			t.Error("expected no spinner re-tick command outside StateInstalling, but got one")
		}
	}
}

// sendEngramFound sends EngramCheckMsg{Found:true} and returns updated model.
func sendEngramFound(t *testing.T, m setup.Model) setup.Model {
	t.Helper()
	next, _ := m.Update(setup.EngramCheckMsg{Found: true})
	m2, ok := next.(setup.Model)
	if !ok {
		t.Fatalf("Update returned %T, want setup.Model", next)
	}
	return m2
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

// updateKeyMsg sends a full KeyMsg through Update and returns the new Model.
func updateKeyMsg(t *testing.T, m setup.Model, msg tea.KeyMsg) setup.Model {
	t.Helper()
	next, _ := m.Update(msg)
	m2, ok := next.(setup.Model)
	if !ok {
		t.Fatalf("Update returned %T, want setup.Model", next)
	}
	return m2
}
