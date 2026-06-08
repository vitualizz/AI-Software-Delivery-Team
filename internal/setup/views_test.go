package setup_test

import (
	"fmt"
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

func TestView_EngramMissingScreenShowsTitle(t *testing.T) {
	m := setup.New(fstest.MapFS{}, "dev")
	// State starts at StateEngramMissing (zero value before Init fires).
	view := m.View()
	if !strings.Contains(view, "Engram Required") {
		t.Errorf("engram missing view should contain 'Engram Required', got:\n%s", view)
	}
}

func TestView_EngramMissingScreenShowsURL(t *testing.T) {
	m := setup.New(fstest.MapFS{}, "dev")
	view := m.View()
	if !strings.Contains(view, "github.com/Gentleman-Programming/engram") {
		t.Errorf("engram missing view should contain URL, got:\n%s", view)
	}
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

func TestView_AssistantListShowsBothNames(t *testing.T) {
	m := setup.New(fstest.MapFS{}, "dev")
	m = sendEngramFound(t, m)
	m = updateKey(t, m, tea.KeyDown)    // cursor → 1 (Install)
	m2 := updateKey(t, m, tea.KeyEnter) // → AssistantList
	view := m2.View()
	for _, d := range installer.Descriptors {
		if !strings.Contains(view, d.Name) {
			t.Errorf("assistant list missing %q, view:\n%s", d.Name, view)
		}
	}
}

func TestView_AssistantListSelectedItemHasCursor(t *testing.T) {
	m := setup.New(fstest.MapFS{}, "dev")
	m = sendEngramFound(t, m)
	m = updateKey(t, m, tea.KeyDown)    // cursor → 1 (Install)
	m2 := updateKey(t, m, tea.KeyEnter) // → AssistantList
	view := m2.View()
	if !strings.Contains(view, "►") {
		t.Errorf("assistant list missing cursor ►, view:\n%s", view)
	}
}

func TestView_AssistantListUsesBadgeForStatus(t *testing.T) {
	// Force a color-capable profile: in Ascii (no-color) mode, the prior
	// ad-hoc `fmt.Sprintf("[%s]", styles.Default.Success.Render("present"))`
	// and panels.NewBadge(...).Render() both degrade to the same plain
	// "[present]" string — the assertion would pass either way and prove
	// nothing about which rendering path is actually used. With color
	// enabled, panels.Badge styles the brackets *inside* the colored span
	// ("\x1b[...][present]\x1b[0m"), while the old call site styled only the
	// label and wrapped brackets outside ("[\x1b[...]present\x1b[0m]") — a
	// real, observable difference that proves adoption (not duplication).
	prev := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.TrueColor)
	defer lipgloss.SetColorProfile(prev)

	m := setup.New(fstest.MapFS{}, "dev")
	m = sendEngramFound(t, m)
	m = updateKey(t, m, tea.KeyDown)    // cursor → 1 (Install)
	m2 := updateKey(t, m, tea.KeyEnter) // → AssistantList
	view := m2.View()

	wantPresent := panels.NewBadge("present", panels.ColorSuccess).Render()
	wantMissing := panels.NewBadge("missing", panels.ColorError).Render()

	if !strings.Contains(view, wantPresent) && !strings.Contains(view, wantMissing) {
		t.Errorf("expected assistant list to render status via panels.Badge ([present]/[missing] in tone color), got: %q", view)
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

// stateView drives the model from StateEngramMissing to the requested
// ViewState by walking the same Update transitions the real TUI uses, and
// returns the rendered View() output for that state.
func stateView(t *testing.T, target string) string {
	t.Helper()

	m := setup.New(fstest.MapFS{}, "dev")
	view := m.View()
	if target == "EngramMissing" {
		return view
	}

	m = sendEngramFound(t, m)
	if target == "MainMenu" {
		return m.View()
	}

	m = updateKey(t, m, tea.KeyDown)  // cursor → 1 (Install)
	m = updateKey(t, m, tea.KeyEnter) // MainMenu → AssistantList
	if target == "AssistantList" {
		return m.View()
	}

	m = updateKey(t, m, tea.KeyEnter) // AssistantList → SelectAssistants
	if target == "SelectAssistants" {
		return m.View()
	}

	m = updateKey(t, m, tea.KeyEnter) // SelectAssistants → SelectProvider
	if target == "SelectProvider" {
		return m.View()
	}

	m = updateKey(t, m, tea.KeyEnter) // SelectProvider → Installing
	if target == "Installing" {
		return m.View()
	}

	t.Fatalf("stateView: unknown target %q", target)
	return ""
}

// TestView_AllStatesHaveBorder proves every one of the 7 view states wraps
// its body in the rounded-border Box style (spec: "every screen's rendered
// output contains a bordered box").
func TestView_AllStatesHaveBorder(t *testing.T) {
	states := []string{
		"EngramMissing",
		"MainMenu",
		"AssistantList",
		"SelectAssistants",
		"SelectProvider",
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
		m = updateKey(t, m, tea.KeyEnter) // cursor 0 → StateDashboard
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
// survives StatusBar styling — i.e. it wasn't dropped when raw "\n[...] quit"
// literals were replaced (spec: "every screen's footer is rendered via
// StatusBar styling").
func TestView_FooterRendersHintText(t *testing.T) {
	cases := []struct {
		state string
		hint  string
	}{
		{"EngramMissing", "q"},
		{"MainMenu", "↑↓"},
		{"AssistantList", "enter"},
		{"SelectAssistants", "space"},
		{"SelectProvider", "esc"},
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
		m = updateKey(t, m, tea.KeyEnter) // cursor 0 → StateDashboard
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

func TestView_DashboardShowsComingSoon(t *testing.T) {
	m := setup.New(fstest.MapFS{}, "dev")
	m = sendEngramFound(t, m)
	m = updateKey(t, m, tea.KeyEnter) // cursor 0 → StateDashboard
	view := m.View()
	if !strings.Contains(view, "Coming Soon") {
		t.Errorf("Dashboard view missing 'Coming Soon', got:\n%s", view)
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
