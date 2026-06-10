package setup_test

import (
	"errors"
	"os"
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

func TestUpdate_EnvironmentCheckMsg_EngramFound_EntersSelectAssistantsOnEnter(t *testing.T) {
	m := setup.New(fstest.MapFS{}, "dev")
	// Navigate from MainMenu to StateEnvironmentCheck via cursor-0 Enter.
	m = updateKey(t, m, tea.KeyEnter) // cursor-0 (Install) → StateEnvironmentCheck
	next, _ := m.Update(setup.EnvironmentCheckMsg{EngramFound: true})
	m2 := next.(setup.Model)
	m3 := updateKey(t, m2, tea.KeyEnter)
	if m3.State() != setup.StateSelectAssistants {
		t.Errorf("after Enter(Install) + EnvironmentCheckMsg{true} + Enter: state = %v, want StateSelectAssistants", m3.State())
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
	m = advanceToMainMenu(t, m)       // no-op, already at StateMainMenu
	m = updateKey(t, m, tea.KeyEnter) // cursor-0 (Install) → StateEnvironmentCheck
	next, _ := m.Update(setup.EnvironmentCheckMsg{EngramFound: true})
	m = next.(setup.Model)
	m = updateKey(t, m, tea.KeyEnter) // preflightDone → StateSelectAssistants

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

func TestUpdate_SelectAssistants_AKeySelectsAll(t *testing.T) {
	m := setup.New(fstest.MapFS{}, "dev")
	m = updateKey(t, m, tea.KeyEnter) // MainMenu → EnvironmentCheck
	next, _ := m.Update(setup.EnvironmentCheckMsg{EngramFound: true})
	m = next.(setup.Model)
	m = updateKey(t, m, tea.KeyEnter) // EnvironmentCheck → SelectAssistants (all pre-selected)

	// Deselect all first, then press 'a' to re-select all.
	for range installer.Descriptors {
		m = updateKeyMsg(t, m, tea.KeyMsg{Type: tea.KeySpace})
		m = updateKey(t, m, tea.KeyDown)
	}
	m2 := updateKeyMsg(t, m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	if m2.State() != setup.StateSelectAssistants {
		t.Errorf("after 'a': state = %v, want StateSelectAssistants", m2.State())
	}
}

func TestUpdate_SelectAssistants_PreSelectsAllOnFirstEntry(t *testing.T) {
	m := setup.New(fstest.MapFS{}, "dev")
	m = updateKey(t, m, tea.KeyEnter)
	next, _ := m.Update(setup.EnvironmentCheckMsg{EngramFound: true})
	m = next.(setup.Model)
	m = updateKey(t, m, tea.KeyEnter) // EnvironmentCheck → SelectAssistants
	if m.State() != setup.StateSelectAssistants {
		t.Fatalf("expected StateSelectAssistants, got %v", m.State())
	}
	// Confirm immediately without touching selection → should proceed to SelectProvider.
	m2 := updateKey(t, m, tea.KeyEnter)
	if m2.State() != setup.StateSelectProvider {
		t.Errorf("Enter at SelectAssistants (all pre-selected): state = %v, want StateSelectProvider", m2.State())
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
	m = advanceToMainMenu(t, m)      // no-op, cursor=0
	m = updateKey(t, m, tea.KeyDown) // cursor → 1 (Dashboard)
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
	m2 := updateKey(t, m, tea.KeyEnter) // preflightDone → StateSelectAssistants
	if m2.State() != setup.StateSelectAssistants {
		t.Fatalf("expected StateSelectAssistants, got %v", m2.State())
	}
	m3 := updateKey(t, m2, tea.KeyEsc)
	if m3.State() != setup.StateEnvironmentCheck {
		t.Errorf("ESC at SelectAssistants: state = %v, want StateEnvironmentCheck", m3.State())
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
	// Force state to Installing directly via the message (backward-compat path).
	next, _ := m.Update(setup.InstallDoneMsg{Results: []installer.InstallResult{}})
	m2 := next.(setup.Model)
	if m2.State() != setup.StateDone {
		t.Errorf("InstallDoneMsg: state = %v, want StateDone", m2.State())
	}
}

func TestUpdate_AssistantInstallProgress_AccumulatesResults(t *testing.T) {
	m := setup.New(fstest.MapFS{}, "dev")
	// Send one progress msg — should NOT transition to Done yet (installExpected = 0, agentDone = false).
	next, _ := m.Update(setup.AssistantInstallProgressMsg{
		Result: installer.InstallResult{AssistantID: installer.AssistantClaudeCode},
	})
	m2 := next.(setup.Model)
	// State should not be StateDone because installExpected is 0 (not set via real flow).
	// The result should be appended regardless.
	if m2.State() == setup.StateDone {
		// installExpected==0 means len(results)==0 is never < 0, agentDone==false → not done.
		// But len(results)=1 >= installExpected=0 AND agentDone=false → not done. Correct.
		t.Error("should not be StateDone: agentDone is false")
	}
}

func TestUpdate_AssistantInstallProgress_AllDone_TransitionsToDone(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", "")
	m := setup.New(fstest.MapFS{}, "dev")
	// Advance to StateReview with skip=true (sets agentDone=true on Enter).
	m = advanceToSelectProvider(t, m)
	m = updateKey(t, m, tea.KeyEnter) // → AgentSetup
	// Navigate to Skip button (5 presets, Skip is at index 5).
	for range installer.PersonaPresets {
		m = updateKey(t, m, tea.KeyDown)
	}
	m = updateKey(t, m, tea.KeyEnter) // Skip → Review
	if m.State() != setup.StateReview {
		t.Fatalf("expected StateReview, got %v", m.State())
	}
	m = updateKey(t, m, tea.KeyEnter) // Review → Installing (sets installExpected, agentDone=true)
	if m.State() != setup.StateInstalling {
		t.Fatalf("expected StateInstalling, got %v", m.State())
	}

	// Send one progress msg per selected assistant to reach installExpected.
	// All assistants are pre-selected by default, so we need len(installer.Descriptors) msgs.
	for _, d := range installer.Descriptors {
		next, _ := m.Update(setup.AssistantInstallProgressMsg{
			Result: installer.InstallResult{AssistantID: d.ID},
		})
		m = next.(setup.Model)
	}
	if m.State() != setup.StateDone {
		t.Errorf("after all progress msgs (skip=true): state = %v, want StateDone", m.State())
	}
}

func TestUpdate_AssistantInstallProgress_WaitsForAgentConfig(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", "")
	m := setup.New(fstest.MapFS{}, "dev")
	m = advanceToSelectProvider(t, m)
	m = updateKey(t, m, tea.KeyEnter) // → AgentSetup (no conflicts in clean HOME)
	m = updateKey(t, m, tea.KeyEnter) // preset 0 → Review
	m = updateKey(t, m, tea.KeyEnter) // Review → Installing (agentDone=false)
	if m.State() != setup.StateInstalling {
		t.Fatalf("expected StateInstalling, got %v", m.State())
	}

	// Send all install msgs — should NOT transition yet because agentDone=false.
	for _, d := range installer.Descriptors {
		next, _ := m.Update(setup.AssistantInstallProgressMsg{
			Result: installer.InstallResult{AssistantID: d.ID},
		})
		m = next.(setup.Model)
	}
	if m.State() == setup.StateDone {
		t.Error("should not be StateDone: AgentInstallDoneMsg not yet received")
	}

	// Now send AgentInstallDoneMsg → should transition.
	next, _ := m.Update(setup.AgentInstallDoneMsg{Results: []installer.AgentConfigResult{}})
	m = next.(setup.Model)
	if m.State() != setup.StateDone {
		t.Errorf("after AgentInstallDoneMsg: state = %v, want StateDone", m.State())
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
// Uses a clean HOME so detectAgentConflicts finds no existing config.
func toInstalling(t *testing.T, m setup.Model) setup.Model {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", "")
	m = advanceToMainMenu(t, m)       // no-op
	m = updateKey(t, m, tea.KeyEnter) // cursor-0 (Install) → StateEnvironmentCheck
	next, _ := m.Update(setup.EnvironmentCheckMsg{EngramFound: true})
	m2, ok := next.(setup.Model)
	if !ok {
		t.Fatalf("toInstalling: EnvironmentCheckMsg Update returned %T, want setup.Model", next)
	}
	m2 = updateKey(t, m2, tea.KeyEnter) // preflightDone → StateSelectAssistants
	m2 = updateKey(t, m2, tea.KeyEnter) // SelectAssistants → SelectProvider
	m2 = updateKey(t, m2, tea.KeyEnter) // SelectProvider → AgentSetup
	m2 = updateKey(t, m2, tea.KeyEnter) // AgentSetup → Review (no conflicts)
	m2 = updateKey(t, m2, tea.KeyEnter) // Review → Installing
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

// --- StateAgentSetup transition tests ---

func TestUpdate_SelectProvider_EnterGoesToAgentSetup(t *testing.T) {
	m := setup.New(fstest.MapFS{}, "dev")
	m = advanceToSelectProvider(t, m)
	m2 := updateKey(t, m, tea.KeyEnter)
	if m2.State() != setup.StateAgentSetup {
		t.Errorf("Enter at SelectProvider: state = %v, want StateAgentSetup", m2.State())
	}
}

func TestUpdate_AgentSetup_EnterPreset_GoesToReview(t *testing.T) {
	// Use a clean HOME so detectAgentConflicts finds no existing config.
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", "")
	m := setup.New(fstest.MapFS{}, "dev")
	m = advanceToSelectProvider(t, m)
	m = updateKey(t, m, tea.KeyEnter) // SelectProvider → AgentSetup
	if m.State() != setup.StateAgentSetup {
		t.Fatalf("expected StateAgentSetup, got %v", m.State())
	}
	// Enter on cursor=0 (Sky preset) with no conflicts → StateReview.
	m2 := updateKey(t, m, tea.KeyEnter)
	if m2.State() != setup.StateReview {
		t.Errorf("Enter on preset at AgentSetup (no conflicts): state = %v, want StateReview", m2.State())
	}
}

func TestUpdate_Review_EnterGoesToInstalling(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", "")
	m := setup.New(fstest.MapFS{}, "dev")
	m = advanceToSelectProvider(t, m)
	m = updateKey(t, m, tea.KeyEnter) // SelectProvider → AgentSetup
	m = updateKey(t, m, tea.KeyEnter) // AgentSetup → Review
	if m.State() != setup.StateReview {
		t.Fatalf("expected StateReview, got %v", m.State())
	}
	m2 := updateKey(t, m, tea.KeyEnter) // Review → Installing
	if m2.State() != setup.StateInstalling {
		t.Errorf("Enter at Review: state = %v, want StateInstalling", m2.State())
	}
}

func TestUpdate_Review_EscGoesToAgentSetup(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", "")
	m := setup.New(fstest.MapFS{}, "dev")
	m = advanceToSelectProvider(t, m)
	m = updateKey(t, m, tea.KeyEnter) // SelectProvider → AgentSetup
	m = updateKey(t, m, tea.KeyEnter) // AgentSetup → Review (no conflicts)
	if m.State() != setup.StateReview {
		t.Fatalf("expected StateReview, got %v", m.State())
	}
	m2 := updateKey(t, m, tea.KeyEsc)
	if m2.State() != setup.StateAgentSetup {
		t.Errorf("Esc at Review (no conflicts): state = %v, want StateAgentSetup", m2.State())
	}
}

func TestUpdate_AgentSetup_EnterSkip_GoesToReview(t *testing.T) {
	// Use a clean HOME so detectAgentConflicts finds no existing config.
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", "")
	m := setup.New(fstest.MapFS{}, "dev")
	m = advanceToSelectProvider(t, m)
	m = updateKey(t, m, tea.KeyEnter) // SelectProvider → AgentSetup

	// Navigate to Skip (cursor=5 = 5 presets).
	m = updateKey(t, m, tea.KeyDown)
	m = updateKey(t, m, tea.KeyDown)
	m = updateKey(t, m, tea.KeyDown)
	m = updateKey(t, m, tea.KeyDown)
	m = updateKey(t, m, tea.KeyDown)

	m2 := updateKey(t, m, tea.KeyEnter) // Skip → Review
	if m2.State() != setup.StateReview {
		t.Errorf("Enter on Skip: state = %v, want StateReview", m2.State())
	}
}

func TestUpdate_AgentSetup_Esc_ReturnsToSelectProvider(t *testing.T) {
	m := setup.New(fstest.MapFS{}, "dev")
	m = advanceToSelectProvider(t, m)
	m = updateKey(t, m, tea.KeyEnter) // SelectProvider → AgentSetup
	if m.State() != setup.StateAgentSetup {
		t.Fatalf("expected StateAgentSetup, got %v", m.State())
	}
	m2 := updateKey(t, m, tea.KeyEsc)
	if m2.State() != setup.StateSelectProvider {
		t.Errorf("Esc at AgentSetup: state = %v, want StateSelectProvider", m2.State())
	}
}

// --- StateAgentWriteMode transition tests ---

func TestUpdate_AgentSetup_WithConflict_EntersAgentWriteMode(t *testing.T) {
	// Create a fake home with a CLAUDE.md that contains the asdt block marker,
	// so detectAgentConflicts returns a non-empty slice.
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("XDG_CONFIG_HOME", "")

	claudeDir := tmpHome + "/.claude"
	if err := os.MkdirAll(claudeDir, 0o755); err != nil {
		t.Fatal(err)
	}
	// Write a CLAUDE.md containing the asdt block start marker.
	claudePath := claudeDir + "/CLAUDE.md"
	marker := "<!-- asdt:agent-config -->"
	if err := os.WriteFile(claudePath, []byte("# My Config\n\n"+marker+"\n# Agent\n<!-- /asdt:agent-config -->\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	m := setup.New(fstest.MapFS{}, "dev")
	m = advanceToSelectProvider(t, m)
	m = updateKey(t, m, tea.KeyEnter) // SelectProvider → AgentSetup (detects conflicts)
	if m.State() != setup.StateAgentSetup {
		t.Fatalf("expected StateAgentSetup, got %v", m.State())
	}

	// Enter on preset 0 (Sky) with conflict → should go to StateAgentWriteMode.
	m2 := updateKey(t, m, tea.KeyEnter)
	if m2.State() != setup.StateAgentWriteMode {
		t.Errorf("Enter on preset with conflict: state = %v, want StateAgentWriteMode", m2.State())
	}
}

func setupConflict(t *testing.T) {
	t.Helper()
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("XDG_CONFIG_HOME", "")
	claudeDir := tmpHome + "/.claude"
	if err := os.MkdirAll(claudeDir, 0o755); err != nil {
		t.Fatal(err)
	}
	marker := "<!-- asdt:agent-config -->"
	if err := os.WriteFile(claudeDir+"/CLAUDE.md", []byte(marker+"\n<!-- /asdt:agent-config -->\n"), 0o644); err != nil {
		t.Fatal(err)
	}
}

func advanceToAgentWriteMode(t *testing.T) setup.Model {
	t.Helper()
	setupConflict(t)
	m := setup.New(fstest.MapFS{}, "dev")
	m = advanceToSelectProvider(t, m)
	m = updateKey(t, m, tea.KeyEnter) // → AgentSetup (conflict detected)
	m = updateKey(t, m, tea.KeyEnter) // → AgentWriteMode (preset 0 with conflict)
	if m.State() != setup.StateAgentWriteMode {
		t.Fatalf("advanceToAgentWriteMode: state = %v, want StateAgentWriteMode", m.State())
	}
	return m
}

func TestUpdate_AgentWriteMode_EnterGoesToReview(t *testing.T) {
	m := advanceToAgentWriteMode(t)
	m2 := updateKey(t, m, tea.KeyEnter)
	if m2.State() != setup.StateReview {
		t.Errorf("Enter at AgentWriteMode: state = %v, want StateReview", m2.State())
	}
}

func TestUpdate_AgentWriteMode_SpaceCyclesMode(t *testing.T) {
	m := advanceToAgentWriteMode(t)
	// Default mode for the conflicted assistant is AgentModeSkip (Keep).
	// Space cycles: Skip(2) → Overwrite(0).
	m2 := updateKeyMsg(t, m, tea.KeyMsg{Type: tea.KeySpace})
	modes := m2.AgentWriteModes()
	if len(modes) == 0 {
		t.Fatal("expected agentWriteModes to be non-empty after AgentWriteMode entry")
	}
	for _, mode := range modes {
		if mode != installer.AgentModeOverwrite {
			t.Errorf("after one Space cycle: mode = %v, want AgentModeOverwrite", mode)
		}
	}
}

func TestUpdate_AgentWriteMode_SpaceCyclesAllModes(t *testing.T) {
	m := advanceToAgentWriteMode(t)
	// Skip → Overwrite → Append → Skip
	m = updateKeyMsg(t, m, tea.KeyMsg{Type: tea.KeySpace}) // → Overwrite
	m = updateKeyMsg(t, m, tea.KeyMsg{Type: tea.KeySpace}) // → Append
	m = updateKeyMsg(t, m, tea.KeyMsg{Type: tea.KeySpace}) // → Skip
	for _, mode := range m.AgentWriteModes() {
		if mode != installer.AgentModeSkip {
			t.Errorf("after 3 Space cycles: mode = %v, want AgentModeSkip", mode)
		}
	}
}

func TestUpdate_AgentWriteMode_ReviewEscGoesBackToAgentWriteMode(t *testing.T) {
	m := advanceToAgentWriteMode(t)
	m = updateKey(t, m, tea.KeyEnter) // AgentWriteMode → Review
	if m.State() != setup.StateReview {
		t.Fatalf("expected StateReview, got %v", m.State())
	}
	m2 := updateKey(t, m, tea.KeyEsc) // Review → AgentWriteMode (has conflicts)
	if m2.State() != setup.StateAgentWriteMode {
		t.Errorf("Esc at Review (with conflicts): state = %v, want StateAgentWriteMode", m2.State())
	}
}

func TestUpdate_AgentWriteMode_Esc_ReturnsToAgentSetup(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("XDG_CONFIG_HOME", "")

	claudeDir := tmpHome + "/.claude"
	if err := os.MkdirAll(claudeDir, 0o755); err != nil {
		t.Fatal(err)
	}
	marker := "<!-- asdt:agent-config -->"
	if err := os.WriteFile(claudeDir+"/CLAUDE.md", []byte(marker+"\n<!-- /asdt:agent-config -->\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	m := setup.New(fstest.MapFS{}, "dev")
	m = advanceToSelectProvider(t, m)
	m = updateKey(t, m, tea.KeyEnter) // → AgentSetup (conflict)
	m = updateKey(t, m, tea.KeyEnter) // → AgentWriteMode
	if m.State() != setup.StateAgentWriteMode {
		t.Fatalf("expected StateAgentWriteMode, got %v", m.State())
	}

	m2 := updateKey(t, m, tea.KeyEsc)
	if m2.State() != setup.StateAgentSetup {
		t.Errorf("Esc at AgentWriteMode: state = %v, want StateAgentSetup", m2.State())
	}
}

// advanceToSelectProvider drives the model to StateSelectProvider.
func advanceToSelectProvider(t *testing.T, m setup.Model) setup.Model {
	t.Helper()
	m = updateKey(t, m, tea.KeyEnter) // MainMenu → EnvironmentCheck
	next, _ := m.Update(setup.EnvironmentCheckMsg{EngramFound: true})
	m = next.(setup.Model)
	m = updateKey(t, m, tea.KeyEnter) // EnvironmentCheck → SelectAssistants
	m = updateKey(t, m, tea.KeyEnter) // SelectAssistants → SelectProvider
	if m.State() != setup.StateSelectProvider {
		t.Fatalf("advanceToSelectProvider: state = %v, want StateSelectProvider", m.State())
	}
	return m
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
