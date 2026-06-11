package setup_test

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"github.com/vitualizz/ai-software-delivery-team/internal/installer"
	"github.com/vitualizz/ai-software-delivery-team/internal/setup"
	"github.com/vitualizz/ai-software-delivery-team/internal/tui/panels"
)

// TestMain forces the English locale for the entire test binary so that
// view string assertions are deterministic regardless of the developer's
// system locale.
func TestMain(m *testing.M) {
	_ = os.Setenv("ASDT_LANG", "en")
	os.Exit(m.Run())
}

func TestView_MainMenuContainsInstallOption(t *testing.T) {
	m := setup.New(fstest.MapFS{}, "dev")
	m = sendEngramFound(t, m)
	view := m.View()
	if !strings.Contains(view, "Install / Update Skills") {
		t.Errorf("main menu view missing 'Install / Update Skills', got:\n%s", view)
	}
}

func TestView_MainMenuShowsAsdtTuiHeader(t *testing.T) {
	m := setup.New(fstest.MapFS{}, "dev")
	m = sendEngramFound(t, m)
	view := m.View()
	if !strings.Contains(view, "asdt-tui") {
		t.Errorf("main menu view missing 'asdt-tui' header, got:\n%s", view)
	}
}

func TestView_MainMenuShowsUpdateBanner(t *testing.T) {
	m := setup.New(fstest.MapFS{}, "v0.2.0")
	m = sendEngramFound(t, m)

	next, _ := m.Update(setup.UpdateCheckMsg{Current: "v0.2.0", Latest: "v0.3.0"})
	m2 := next.(setup.Model)

	view := m2.View()
	if !strings.Contains(view, "/releases") {
		t.Errorf("main menu view missing update banner releases URL, got:\n%s", view)
	}
	if !strings.Contains(view, "v0.3.0") {
		t.Errorf("main menu view missing latest version v0.3.0, got:\n%s", view)
	}
}

func TestView_SelectAssistantsShowsBothNames(t *testing.T) {
	m := setup.New(fstest.MapFS{}, "dev")
	m = sendEngramFound(t, m)         // no-op, already at MainMenu
	m = updateKey(t, m, tea.KeyEnter) // cursor-0 (Install) → StateEnvironmentCheck
	next, _ := m.Update(setup.EnvironmentCheckMsg{EngramFound: true})
	m = next.(setup.Model)
	m2 := updateKey(t, m, tea.KeyEnter) // preflightDone → StateSelectAssistants
	view := m2.View()
	for _, d := range installer.Descriptors {
		if !strings.Contains(view, d.Name) {
			t.Errorf("select assistants missing %q, view:\n%s", d.Name, view)
		}
	}
}

func TestView_SelectAssistantsSelectedItemHasCursor(t *testing.T) {
	m := setup.New(fstest.MapFS{}, "dev")
	m = sendEngramFound(t, m)         // no-op, already at MainMenu
	m = updateKey(t, m, tea.KeyEnter) // cursor-0 (Install) → StateEnvironmentCheck
	next, _ := m.Update(setup.EnvironmentCheckMsg{EngramFound: true})
	m = next.(setup.Model)
	m2 := updateKey(t, m, tea.KeyEnter) // preflightDone → StateSelectAssistants
	view := m2.View()
	if !strings.Contains(view, "►") {
		t.Errorf("select assistants missing cursor ►, view:\n%s", view)
	}
}

func TestView_SelectAssistantsUsesBadgeForStatus(t *testing.T) {
	// Force color so panels.Badge styling is distinguishable from plain-text fallback.
	prev := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.TrueColor)
	defer lipgloss.SetColorProfile(prev)

	m := setup.New(fstest.MapFS{}, "dev")
	m = sendEngramFound(t, m)         // no-op, already at MainMenu
	m = updateKey(t, m, tea.KeyEnter) // cursor-0 (Install) → StateEnvironmentCheck
	next, _ := m.Update(setup.EnvironmentCheckMsg{EngramFound: true})
	m = next.(setup.Model)
	m2 := updateKey(t, m, tea.KeyEnter) // preflightDone → StateSelectAssistants
	view := m2.View()

	wantPresent := panels.NewBadge("present", panels.ColorSuccess).Render()
	wantMissing := panels.NewBadge("missing", panels.ColorError).Render()

	if !strings.Contains(view, wantPresent) && !strings.Contains(view, wantMissing) {
		t.Errorf("expected select assistants to render status via panels.Badge ([present]/[missing] in tone color), got: %q", view)
	}
}

func TestView_SelectAssistantsShowsCheckboxes(t *testing.T) {
	m := setup.New(fstest.MapFS{}, "dev")
	m = sendEngramFound(t, m)
	m = updateKey(t, m, tea.KeyEnter) // → StateEnvironmentCheck
	next, _ := m.Update(setup.EnvironmentCheckMsg{EngramFound: true})
	m = next.(setup.Model)
	m2 := updateKey(t, m, tea.KeyEnter) // → StateSelectAssistants
	view := m2.View()
	if !strings.Contains(view, "[x]") {
		t.Errorf("select assistants missing pre-selected checkbox [x], view:\n%s", view)
	}
}

func TestView_DoneScreenShowsBothAssistants(t *testing.T) {
	successResult := installer.InstallResult{AssistantID: installer.AssistantClaudeCode, Err: nil}
	failResult := installer.InstallResult{AssistantID: installer.AssistantOpenCode, Err: fmt.Errorf("fail")}
	m := setup.New(fstest.MapFS{}, "dev")
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

// stateView drives the model from StateMainMenu to the requested ViewState by
// walking the same Update transitions the real TUI uses, and returns the
// rendered View() output for that state.
func stateView(t *testing.T, target string) string {
	t.Helper()

	m := setup.New(fstest.MapFS{}, "dev")

	if target == "PreflightCheck" {
		// Trigger preflight by pressing Enter at cursor-0 (Install).
		m = updateKey(t, m, tea.KeyEnter) // cursor-0 → StateEnvironmentCheck
		return m.View()
	}

	m = sendEngramFound(t, m) // no-op, already at MainMenu
	if target == "MainMenu" {
		return m.View()
	}

	// For any state that passes through SelectProvider → AgentSetup, use a
	// clean HOME so detectAgentConflicts returns empty (no conflicts).
	// AgentWriteMode intentionally sets up its own HOME with a conflict.
	if target != "AgentWriteMode" {
		t.Setenv("HOME", t.TempDir())
		t.Setenv("XDG_CONFIG_HOME", "")
	}

	// Install path: cursor-0 Enter → StateEnvironmentCheck → EnvironmentCheckMsg → Enter → StateSelectAssistants.
	m = updateKey(t, m, tea.KeyEnter) // cursor-0 (Install) → StateEnvironmentCheck
	next, _ := m.Update(setup.EnvironmentCheckMsg{EngramFound: true})
	m = next.(setup.Model)
	m = updateKey(t, m, tea.KeyEnter) // preflightDone → StateSelectAssistants
	if target == "SelectAssistants" {
		return m.View()
	}

	m = updateKey(t, m, tea.KeyEnter) // SelectAssistants → SelectProvider
	if target == "SelectProvider" {
		return m.View()
	}

	m = updateKey(t, m, tea.KeyEnter) // SelectProvider → AgentSetup
	if target == "AgentSetup" {
		return m.View()
	}

	m = updateKey(t, m, tea.KeyEnter) // AgentSetup → EmojiPref
	if target == "EmojiPref" {
		return m.View()
	}

	m = updateKey(t, m, tea.KeyEnter) // EmojiPref → Review
	if target == "Review" {
		return m.View()
	}

	if target == "AgentWriteMode" {
		// Need a conflict to enter StateAgentWriteMode. Set up a CLAUDE.md with
		// the asdt block marker so detectAgentConflicts returns a hit.
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
		// Re-drive from scratch with the new HOME so detectAgentConflicts fires.
		m2 := setup.New(fstest.MapFS{}, "dev")
		m2 = updateKey(t, m2, tea.KeyEnter) // MainMenu → EnvironmentCheck
		next2, _ := m2.Update(setup.EnvironmentCheckMsg{EngramFound: true})
		m2 = next2.(setup.Model)
		m2 = updateKey(t, m2, tea.KeyEnter) // EnvironmentCheck → SelectAssistants
		m2 = updateKey(t, m2, tea.KeyEnter) // SelectAssistants → SelectProvider
		m2 = updateKey(t, m2, tea.KeyEnter) // SelectProvider → AgentSetup (conflict detected)
		m2 = updateKey(t, m2, tea.KeyEnter) // AgentSetup preset → EmojiPref
		m2 = updateKey(t, m2, tea.KeyEnter) // EmojiPref → AgentWriteMode (conflict)
		return m2.View()
	}

	m = updateKey(t, m, tea.KeyEnter) // Review → Installing
	if target == "Installing" {
		return m.View()
	}

	t.Fatalf("stateView: unknown target %q", target)
	return ""
}

// TestView_AllStatesHaveBorder proves every one of the view states wraps
// its body in the rounded-border Box style (spec: "every screen's rendered
// output contains a bordered box").
func TestView_AllStatesHaveBorder(t *testing.T) {
	states := []string{
		"PreflightCheck",
		"MainMenu",
		"SelectAssistants",
		"SelectProvider",
		"AgentSetup",
		"EmojiPref",
		"Review",
		"AgentWriteMode",
		"Installing",
	}

	for _, state := range states {
		t.Run(state, func(t *testing.T) {
			view := stateView(t, state)
			if !strings.ContainsAny(view, "╭╮╰╯│") {
				t.Errorf("%s view missing rounded-border runes, got:\n%s", state, view)
			}
		})
	}

	t.Run("Dashboard", func(t *testing.T) {
		m := setup.New(fstest.MapFS{}, "dev")
		m = sendEngramFound(t, m)
		m = updateKey(t, m, tea.KeyDown)  // cursor → 1 (Dashboard)
		m = updateKey(t, m, tea.KeyEnter) // → StateDashboard
		view := m.View()
		if !strings.ContainsAny(view, "╭╮╰╯│") {
			t.Errorf("Dashboard view missing rounded-border runes, got:\n%s", view)
		}
	})

	t.Run("Done", func(t *testing.T) {
		successResult := installer.InstallResult{AssistantID: installer.AssistantClaudeCode, Err: nil}
		m := setup.New(fstest.MapFS{}, "dev")
		next, _ := m.Update(setup.InstallDoneMsg{Results: []installer.InstallResult{successResult}})
		m2 := next.(setup.Model)
		view := m2.View()
		if !strings.ContainsAny(view, "╭╮╰╯│") {
			t.Errorf("Done view missing rounded-border runes, got:\n%s", view)
		}
	})
}

// TestView_FooterRendersHintText proves the per-screen key-hint footer text
// is rendered inside the bordered box on every screen.
func TestView_FooterRendersHintText(t *testing.T) {
	cases := []struct {
		state string
		hint  string
	}{
		{"PreflightCheck", "checking"},
		{"MainMenu", "↑↓"},
		{"SelectAssistants", "space"},
		{"SelectProvider", "esc"},
		{"EmojiPref", "↑↓"},
		{"Review", "esc"},
		{"Installing", "q"},
	}

	for _, tc := range cases {
		t.Run(tc.state, func(t *testing.T) {
			view := stateView(t, tc.state)
			if !strings.Contains(view, tc.hint) {
				t.Errorf("%s view missing footer hint %q, got:\n%s", tc.state, tc.hint, view)
			}
		})
	}

	t.Run("Dashboard", func(t *testing.T) {
		m := setup.New(fstest.MapFS{}, "dev")
		m = sendEngramFound(t, m)
		m = updateKey(t, m, tea.KeyDown)  // cursor → 1 (Dashboard)
		m = updateKey(t, m, tea.KeyEnter) // → StateDashboard
		view := m.View()
		if !strings.Contains(view, "esc") {
			t.Errorf("Dashboard view missing footer hint %q, got:\n%s", "esc", view)
		}
	})

	t.Run("Done", func(t *testing.T) {
		successResult := installer.InstallResult{AssistantID: installer.AssistantClaudeCode, Err: nil}
		m := setup.New(fstest.MapFS{}, "dev")
		next, _ := m.Update(setup.InstallDoneMsg{Results: []installer.InstallResult{successResult}})
		m2 := next.(setup.Model)
		view := m2.View()
		if !strings.Contains(view, "enter/esc") {
			t.Errorf("Done view missing footer hint %q, got:\n%s", "enter/esc", view)
		}
	})
}

func TestView_DashboardShowsAssistantNames(t *testing.T) {
	m := setup.New(fstest.MapFS{}, "dev")
	m = sendEngramFound(t, m)
	m = updateKey(t, m, tea.KeyDown)  // cursor → 1 (Dashboard)
	m = updateKey(t, m, tea.KeyEnter) // → StateDashboard
	view := m.View()
	for _, d := range installer.Descriptors {
		if !strings.Contains(view, d.Name) {
			t.Errorf("dashboard view missing assistant name %q, got:\n%s", d.Name, view)
		}
	}
}

// TestView_InstallingShowsSpinner verifies that the StateInstalling screen
// renders an indeterminate spinner.Dot frame glyph alongside the existing
// progress indication — the visual cue that installation is actively running
// (T-030..T-039).
func TestView_InstallingShowsSpinner(t *testing.T) {
	view := stateView(t, "Installing")

	found := false
	for _, frame := range spinner.Dot.Frames {
		glyph := strings.TrimRight(frame, " ")
		if glyph == "" {
			continue
		}
		if strings.Contains(view, glyph) {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected Installing view to contain a spinner.Dot frame glyph, got:\n%s", view)
	}
}

func TestView_PreflightCheckShowsTitle(t *testing.T) {
	m := setup.New(fstest.MapFS{}, "dev")
	m = updateKey(t, m, tea.KeyEnter) // cursor-0 (Install) → StateEnvironmentCheck
	view := m.View()
	if !strings.Contains(view, "Pre-flight Check") {
		t.Errorf("preflight view should contain 'Pre-flight Check', got:\n%s", view)
	}
}

func TestView_PreflightCheckShowsSections(t *testing.T) {
	m := setup.New(fstest.MapFS{}, "dev")
	m = updateKey(t, m, tea.KeyEnter) // cursor-0 (Install) → StateEnvironmentCheck
	view := m.View()
	if !strings.Contains(view, "Memory Provider") {
		t.Errorf("preflight view missing 'Memory Provider' section, got:\n%s", view)
	}
}

func TestView_PreflightCheckShowsShellRow(t *testing.T) {
	m := setup.New(fstest.MapFS{}, "dev")
	m = updateKey(t, m, tea.KeyEnter) // cursor-0 (Install) → StateEnvironmentCheck
	view := m.View()
	if !strings.Contains(view, "Shell") {
		t.Errorf("preflight view missing 'Shell' row, got:\n%s", view)
	}
}

func TestView_ReviewShowsAssistantsAndProvider(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", "")
	view := stateView(t, "Review")
	for _, d := range installer.Descriptors {
		if !strings.Contains(view, d.Name) {
			t.Errorf("review view missing assistant name %q, got:\n%s", d.Name, view)
		}
	}
	if !strings.Contains(view, "Assistants") {
		t.Errorf("review view missing 'Assistants' label, got:\n%s", view)
	}
}

func TestView_SelectProviderShowsRadioIndicator(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", "")
	view := stateView(t, "SelectProvider")
	if !strings.Contains(view, "(•)") {
		t.Errorf("SelectProvider missing selected radio indicator '(•)', got:\n%s", view)
	}
	// Unselected radio only appears when there is more than one provider.
	if len(installer.Providers) > 1 && !strings.Contains(view, "( )") {
		t.Errorf("SelectProvider missing unselected radio indicator '( )', got:\n%s", view)
	}
}

func TestView_EmojiPrefShowsRadioOptionsAndSubtitle(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", "")
	view := stateView(t, "EmojiPref")
	if !strings.Contains(view, "Emoji Preference") {
		t.Errorf("EmojiPref missing title 'Emoji Preference', got:\n%s", view)
	}
	if !strings.Contains(view, "Should your assistants use emojis in their responses?") {
		t.Errorf("EmojiPref missing subtitle question, got:\n%s", view)
	}
	if !strings.Contains(view, "Yes — use emojis") {
		t.Errorf("EmojiPref missing Yes option, got:\n%s", view)
	}
	if !strings.Contains(view, "No — plain text only") {
		t.Errorf("EmojiPref missing No option, got:\n%s", view)
	}
	if !strings.Contains(view, "(•)") {
		t.Errorf("EmojiPref missing selected radio indicator '(•)', got:\n%s", view)
	}
	if !strings.Contains(view, "( )") {
		t.Errorf("EmojiPref missing unselected radio indicator '( )', got:\n%s", view)
	}
}

func TestView_EmojiPrefSharesStepFour(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", "")
	view := stateView(t, "EmojiPref")
	if !strings.Contains(view, "step 4 of 5") {
		t.Errorf("EmojiPref should share step 4 of 5 with AgentSetup, got:\n%s", view)
	}
}

func TestView_ReviewShowsEmojiRowWhenNotSkipped(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", "")
	view := stateView(t, "Review")
	if !strings.Contains(view, "Emojis:") {
		t.Errorf("review view missing 'Emojis:' label, got:\n%s", view)
	}
	if !strings.Contains(view, "Yes — use emojis") {
		t.Errorf("review view missing emoji value 'Yes — use emojis' (default), got:\n%s", view)
	}
}

func TestView_ReviewOmitsEmojiRowWhenSkipped(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", "")
	// Drive the skip path to Review: EmojiPref is bypassed entirely.
	m := setup.New(fstest.MapFS{}, "dev")
	m = updateKey(t, m, tea.KeyEnter) // MainMenu → EnvironmentCheck
	next, _ := m.Update(setup.EnvironmentCheckMsg{EngramFound: true})
	m = next.(setup.Model)
	m = updateKey(t, m, tea.KeyEnter) // EnvironmentCheck → SelectAssistants
	m = updateKey(t, m, tea.KeyEnter) // SelectAssistants → SelectProvider
	m = updateKey(t, m, tea.KeyEnter) // SelectProvider → AgentSetup
	for range installer.PersonaPresets {
		m = updateKey(t, m, tea.KeyDown) // navigate to [ Skip → ]
	}
	m = updateKey(t, m, tea.KeyEnter) // Skip → Review
	if m.State() != setup.StateReview {
		t.Fatalf("expected StateReview, got %v", m.State())
	}
	view := m.View()
	if strings.Contains(view, "Emojis:") {
		t.Errorf("review view must omit 'Emojis:' row when persona was skipped, got:\n%s", view)
	}
}

func TestView_AgentSetupShowsRadioIndicator(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", "")
	view := stateView(t, "AgentSetup")
	if !strings.Contains(view, "(•)") {
		t.Errorf("AgentSetup missing selected radio indicator '(•)', got:\n%s", view)
	}
	if !strings.Contains(view, "( )") {
		t.Errorf("AgentSetup missing unselected radio indicator '( )', got:\n%s", view)
	}
}

func TestView_PreflightCheckShowsEngramRecovery(t *testing.T) {
	m := setup.New(fstest.MapFS{}, "dev")
	m = updateKey(t, m, tea.KeyEnter) // → StateEnvironmentCheck
	// Inject EnvironmentCheckMsg with Engram missing
	next, _ := m.Update(setup.EnvironmentCheckMsg{EngramFound: false})
	m2 := next.(setup.Model)
	view := m2.View()
	if !strings.Contains(view, "Engram") {
		t.Errorf("preflight view missing Engram recovery section, got:\n%s", view)
	}
}

func TestView_AgentWriteModeShowsPerAssistantRows(t *testing.T) {
	view := stateView(t, "AgentWriteMode")
	// The view should show at least one mode label (Keep is default)
	if !strings.Contains(view, "Keep") && !strings.Contains(view, "Overwrite") && !strings.Contains(view, "Append") {
		t.Errorf("agent write mode view missing mode labels, got:\n%s", view)
	}
}

func TestView_DoneScreenShowsNextStepHint(t *testing.T) {
	successResult := installer.InstallResult{AssistantID: installer.AssistantClaudeCode, Err: nil}
	m := setup.New(fstest.MapFS{}, "dev")
	next, _ := m.Update(setup.InstallDoneMsg{Results: []installer.InstallResult{successResult}})
	m2 := next.(setup.Model)
	view := m2.View()
	if !strings.Contains(view, "/asdt") {
		t.Errorf("done screen missing next-step hint '/asdt', got:\n%s", view)
	}
}

// sendEngramFound is a no-op helper kept for caller compatibility.
// New() already starts at StateMainMenu, so no navigation is needed.
func sendEngramFound(_ *testing.T, m setup.Model) setup.Model {
	return m
}

// TestView_InstallingSpinnerTintedWithColorSecondary verifies that the
// rendered spinner glyph is styled with panels.ColorSecondary — the same
// cyan/sky pastel reserved for in-progress indicators across the dashboard
// (mirrors panels.NewSpinner's tint, see specialists panel wiring).
func TestView_InstallingSpinnerTintedWithColorSecondary(t *testing.T) {
	prev := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.TrueColor)
	defer lipgloss.SetColorProfile(prev)

	view := stateView(t, "Installing")

	tinted := lipgloss.NewStyle().Foreground(panels.ColorSecondary)
	found := false
	for _, frame := range spinner.Dot.Frames {
		if strings.Contains(view, tinted.Render(frame)) {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected Installing view to contain a spinner frame tinted with ColorSecondary, got:\n%s", view)
	}
}
