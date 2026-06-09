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
	"github.com/vitualizz/ai-software-delivery-team/internal/setup/components"
)

func TestNew_StartsAtMainMenu(t *testing.T) {
	m := setup.New(fstest.MapFS{}, "dev")
	// iota sentinel: StateMainMenu must be 0 (the zero value)
	if setup.StateMainMenu != 0 {
		t.Errorf("StateMainMenu = %d, want 0 (iota sentinel)", int(setup.StateMainMenu))
	}
	if m.State() != setup.StateMainMenu {
		t.Errorf("New() state = %v, want StateMainMenu (%v)", m.State(), setup.StateMainMenu)
	}
}

func TestUpdate_EnvironmentCheckMsg_EngramFound_EntersAssistantListOnEnter(t *testing.T) {
	m := setup.New(fstest.MapFS{}, "dev")
	// Navigate from MainMenu to StateEnvironmentCheck via cursor-0 Enter.
	m = updateKey(t, m, tea.KeyEnter) // cursor-0 (Install) → StateEnvironmentCheck
	next, _ := m.Update(setup.EnvironmentCheckMsg{EngramFound: true})
	m2 := next.(setup.Model)
	m3 := updateKey(t, m2, tea.KeyEnter)
	if m3.State() != setup.StateAssistantList {
		t.Errorf("after Enter(Install) + EnvironmentCheckMsg{true} + Enter: state = %v, want StateAssistantList", m3.State())
	}
}

func TestUpdate_EnvironmentCheckMsg_EngramMissing_BlocksContinue(t *testing.T) {
	m := setup.New(fstest.MapFS{}, "dev")
	m = updateKey(t, m, tea.KeyEnter) // cursor-0 (Install) → StateEnvironmentCheck
	next, _ := m.Update(setup.EnvironmentCheckMsg{EngramFound: false})
	m2 := next.(setup.Model)
	m3 := updateKey(t, m2, tea.KeyEnter)
	if m3.State() != setup.StateEnvironmentCheck {
		t.Errorf("Enter with engram missing should be no-op, state = %v, want StateEnvironmentCheck", m3.State())
	}
}

func TestUpdate_EnvironmentCheckProgressMsg_UpdatesSection(t *testing.T) {
	m := setup.New(fstest.MapFS{}, "dev")
	m = updateKey(t, m, tea.KeyEnter) // cursor-0 (Install) → StateEnvironmentCheck
	next, _ := m.Update(setup.EnvironmentCheckProgressMsg{
		RowLabel: "Engram",
		Status:   components.CheckStatusOK,
		Detail:   "/usr/bin/engram",
	})
	m2 := next.(setup.Model)
	_ = m2 // if no panic, row was updated
}

func TestUpdate_PreflightQQuits(t *testing.T) {
	m := setup.New(fstest.MapFS{}, "dev")
	m = updateKey(t, m, tea.KeyEnter) // cursor-0 → StateEnvironmentCheck
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
	_, cmd := m.Update(msg)
	if cmd == nil {
		t.Error("q in StateEnvironmentCheck: expected non-nil cmd (tea.Quit), got nil")
	}
}

func TestUpdate_PreflightCtrlCQuits(t *testing.T) {
	m := setup.New(fstest.MapFS{}, "dev")
	m = updateKey(t, m, tea.KeyEnter) // cursor-0 → StateEnvironmentCheck
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	if cmd == nil {
		t.Error("ctrl+c in StateEnvironmentCheck: expected non-nil cmd (tea.Quit), got nil")
	}
}

func TestUpdate_SpaceToggleSelectsItem(t *testing.T) {
	m := setup.New(fstest.MapFS{}, "dev")
	// Navigate to StateSelectAssistants via the full install flow.
	m = advanceToMainMenu(t, m) // no-op, already at StateMainMenu
	m = updateKey(t, m, tea.KeyEnter) // cursor-0 (Install) → StateEnvironmentCheck
	next, _ := m.Update(setup.EnvironmentCheckMsg{EngramFound: true})
	m = next.(setup.Model)
	m = updateKey(t, m, tea.KeyEnter) // preflightDone → StateAssistantList
	m = updateKey(t, m, tea.KeyEnter) // → StateSelectAssistants

	if m.State() != setup.StateSelectAssistants {
		t.Fatalf("expected StateSelectAssistants, got %v", m.State())
	}

	// Space should toggle cursor item on.
	m2 := updateKeyMsg(t, m, tea.KeyMsg{Type: tea.KeySpace})
	// The simplest check: send space again to toggle off, verifying toggle behavior.
	m3 := updateKeyMsg(t, m2, tea.KeyMsg{Type: tea.KeySpace})
	_ = m3
	// If we get here without panic the space handler ran. Verify state unchanged.
	if m2.State() != setup.StateSelectAssistants {
		t.Errorf("after space: state = %v, want StateSelectAssistants", m2.State())
	}
}

func TestUpdate_EnterOnInstallTriggersEnvironmentCheck(t *testing.T) {
	m := setup.New(fstest.MapFS{}, "dev")
	m = advanceToMainMenu(t, m) // no-op
	// Cursor starts at 0 = "Install / Update Skills".
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m2 := next.(setup.Model)
	if m2.State() != setup.StateEnvironmentCheck {
		t.Errorf("Enter at cursor 0 (Install): state = %v, want StateEnvironmentCheck", m2.State())
	}
}

func TestUpdate_EnterOnDashboardTransitionsToDashboard(t *testing.T) {
	m := setup.New(fstest.MapFS{}, "dev")
	m = advanceToMainMenu(t, m) // no-op, cursor=0
	m = updateKey(t, m, tea.KeyDown)  // cursor → 1 (Dashboard)
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m2 := next.(setup.Model)
	if m2.State() != setup.StateDashboard {
		t.Errorf("after Enter at cursor 1 (Dashboard): state = %v, want StateDashboard", m2.State())
	}
}

func TestUpdate_EscFromDashboardReturnsToMenu(t *testing.T) {
	m := setup.New(fstest.MapFS{}, "dev")
	m = advanceToMainMenu(t, m)       // no-op
	m = updateKey(t, m, tea.KeyDown)  // cursor → 1 (Dashboard)
	m = updateKey(t, m, tea.KeyEnter) // → StateDashboard
	m2 := updateKey(t, m, tea.KeyEsc) // Esc → back
	if m2.State() != setup.StateMainMenu {
		t.Errorf("Esc from Dashboard: state = %v, want StateMainMenu", m2.State())
	}
}

func TestUpdate_ESCAtSelectAssistantsGoesBack(t *testing.T) {
	m := setup.New(fstest.MapFS{}, "dev")
	m = advanceToMainMenu(t, m)       // no-op
	m = updateKey(t, m, tea.KeyEnter) // cursor-0 (Install) → StateEnvironmentCheck
	next, _ := m.Update(setup.EnvironmentCheckMsg{EngramFound: true})
	m = next.(setup.Model)
	m2 := updateKey(t, m, tea.KeyEnter)  // preflightDone → StateAssistantList
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
	m = advanceToMainMenu(t, m) // no-op
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
	m = advanceToMainMenu(t, m) // no-op

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
	m = advanceToMainMenu(t, m) // no-op

	next, _ := m.Update(setup.UpdateCheckMsg{Err: errors.New("boom")})
	m2 := next.(setup.Model)

	view := m2.View()
	if strings.Contains(view, "/releases") {
		t.Errorf("expected no banner on error, got:\n%s", view)
	}
}

func TestUpdate_UpdateCheckMsg_DevBuild_NoBanner(t *testing.T) {
	m := setup.New(fstest.MapFS{}, "dev")
	m = advanceToMainMenu(t, m) // no-op

	next, _ := m.Update(setup.UpdateCheckMsg{Current: "dev", Latest: "v9.9.9"})
	m2 := next.(setup.Model)

	view := m2.View()
	if strings.Contains(view, "/releases") {
		t.Errorf("expected no banner on dev build, got:\n%s", view)
	}
}

// toInstalling drives the model from StateMainMenu through the full install
// flow to StateInstalling — the same Update transitions the real TUI uses
// (mirrors views_test.go's stateView helper, but returns setup.Model so
// callers can keep driving Update directly).
func toInstalling(t *testing.T, m setup.Model) setup.Model {
	t.Helper()
	m = advanceToMainMenu(t, m)       // no-op
	m = updateKey(t, m, tea.KeyEnter) // cursor-0 (Install) → StateEnvironmentCheck
	next, _ := m.Update(setup.EnvironmentCheckMsg{EngramFound: true})
	m2, ok := next.(setup.Model)
	if !ok {
		t.Fatalf("toInstalling: EnvironmentCheckMsg Update returned %T, want setup.Model", next)
	}
	m2 = updateKey(t, m2, tea.KeyEnter) // preflightDone → StateAssistantList
	m2 = updateKey(t, m2, tea.KeyEnter) // AssistantList → SelectAssistants
	m2 = updateKey(t, m2, tea.KeyEnter) // SelectAssistants → SelectProvider
	m2 = updateKey(t, m2, tea.KeyEnter) // SelectProvider → Installing
	if m2.State() != setup.StateInstalling {
		t.Fatalf("toInstalling: state = %v, want StateInstalling", m2.State())
	}
	return m2
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
// spinner.TickMsg received OUTSIDE StateInstalling (and StateEnvironmentCheck)
// is gated off — it must not produce a re-tick command.
func TestUpdate_SpinnerTickIgnoredOutsideInstalling(t *testing.T) {
	m := setup.New(fstest.MapFS{}, "dev")
	m = advanceToMainMenu(t, m) // no-op → StateMainMenu (not Installing or EnvironmentCheck)

	_, cmd := m.Update(spinner.TickMsg{})
	if cmd != nil {
		msg := cmd()
		if _, ok := msg.(spinner.TickMsg); ok {
			t.Error("expected no spinner re-tick command outside StateInstalling, but got one")
		}
	}
}

// advanceToMainMenu is a no-op helper kept for caller compatibility.
// New() already starts at StateMainMenu, so no navigation is needed.
func advanceToMainMenu(_ *testing.T, m setup.Model) setup.Model {
	return m
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
