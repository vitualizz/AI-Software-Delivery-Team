package setup_test

import (
	"strings"
	"testing"
	"testing/fstest"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/vitualizz/ai-software-delivery-team/internal/setup"
)

// Subtitle questions asserted below pin the live catalog switch: the English
// and Spanish strings must both be reachable from the language screen.
const (
	languageSubtitleEN = "Which language should the installer use?"
	languageSubtitleES = "¿En qué idioma querés usar el instalador?"
)

// toLanguageSelect drives the model from MainMenu into StateLanguageSelect
// (Enter on cursor-0 Install).
func toLanguageSelect(t *testing.T, m setup.Model) setup.Model {
	t.Helper()
	m = updateKey(t, m, tea.KeyEnter) // cursor-0 (Install) → StateLanguageSelect
	if m.State() != setup.StateLanguageSelect {
		t.Fatalf("toLanguageSelect: state = %v, want StateLanguageSelect", m.State())
	}
	return m
}

func TestNew_LanguageDefaultsFromActiveCode(t *testing.T) {
	// TestMain pins ASDT_LANG=en for the binary; override per-case.
	t.Setenv("ASDT_LANG", "es")
	m := setup.New(fstest.MapFS{}, "dev")
	if got := m.LanguageCode(); got != "es" {
		t.Errorf("New() with ASDT_LANG=es: LanguageCode() = %q, want %q", got, "es")
	}

	t.Setenv("ASDT_LANG", "en")
	m = setup.New(fstest.MapFS{}, "dev")
	if got := m.LanguageCode(); got != "en" {
		t.Errorf("New() with ASDT_LANG=en: LanguageCode() = %q, want %q", got, "en")
	}
}

func TestUpdate_MainMenuInstall_EntersLanguageSelect(t *testing.T) {
	m := setup.New(fstest.MapFS{}, "dev")
	next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m2 := next.(setup.Model)
	if m2.State() != setup.StateLanguageSelect {
		t.Errorf("Enter at cursor 0 (Install): state = %v, want StateLanguageSelect", m2.State())
	}
	if cmd == nil {
		t.Error("Enter at cursor 0 (Install): expected LanguagePrefCmd, got nil cmd")
	}
}

func TestUpdate_LanguagePrefMsg_PreselectsPersistedLanguage(t *testing.T) {
	m := setup.New(fstest.MapFS{}, "dev")
	m = toLanguageSelect(t, m)

	next, _ := m.Update(setup.LanguagePrefMsg{Code: "es"})
	m2 := next.(setup.Model)
	if got := m2.LanguageCode(); got != "es" {
		t.Errorf("after LanguagePrefMsg{es}: LanguageCode() = %q, want %q", got, "es")
	}
	if view := m2.View(); !strings.Contains(view, languageSubtitleES) {
		t.Errorf("after LanguagePrefMsg{es}: view should use the Spanish catalog, got:\n%s", view)
	}
}

func TestUpdate_LanguagePrefMsg_IgnoredAfterUserTouched(t *testing.T) {
	m := setup.New(fstest.MapFS{}, "dev")
	m = toLanguageSelect(t, m)

	m = updateKey(t, m, tea.KeyDown) // user explicitly selects Español → touched
	if got := m.LanguageCode(); got != "es" {
		t.Fatalf("after Down: LanguageCode() = %q, want %q", got, "es")
	}

	next, _ := m.Update(setup.LanguagePrefMsg{Code: "en"})
	m2 := next.(setup.Model)
	if got := m2.LanguageCode(); got != "es" {
		t.Errorf("late LanguagePrefMsg must be ignored after touch: LanguageCode() = %q, want %q", got, "es")
	}
}

func TestUpdate_LanguageSelect_UpDownSwitchCatalogLive(t *testing.T) {
	m := setup.New(fstest.MapFS{}, "dev")
	m = toLanguageSelect(t, m)

	if view := m.View(); !strings.Contains(view, languageSubtitleEN) {
		t.Fatalf("language screen should start with the English subtitle, got:\n%s", view)
	}

	m = updateKey(t, m, tea.KeyDown) // → Español
	if view := m.View(); !strings.Contains(view, languageSubtitleES) {
		t.Errorf("after Down (Español): view missing Spanish subtitle, got:\n%s", view)
	}

	m = updateKey(t, m, tea.KeyUp) // → English again
	if view := m.View(); !strings.Contains(view, languageSubtitleEN) {
		t.Errorf("after Up (English): view missing English subtitle, got:\n%s", view)
	}
}

func TestUpdate_LanguageSelect_EnterStartsPreflight(t *testing.T) {
	m := setup.New(fstest.MapFS{}, "dev")
	m = toLanguageSelect(t, m)

	next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m2 := next.(setup.Model)
	if m2.State() != setup.StateEnvironmentCheck {
		t.Errorf("Enter at LanguageSelect: state = %v, want StateEnvironmentCheck", m2.State())
	}
	if cmd == nil {
		t.Error("Enter at LanguageSelect: expected EnvironmentCheckCmd batch, got nil cmd")
	}
}

func TestUpdate_LanguageSelect_EscReturnsToMainMenu(t *testing.T) {
	m := setup.New(fstest.MapFS{}, "dev")
	m = toLanguageSelect(t, m)

	m2 := updateKey(t, m, tea.KeyEsc)
	if m2.State() != setup.StateMainMenu {
		t.Errorf("Esc at LanguageSelect: state = %v, want StateMainMenu", m2.State())
	}
}

func TestView_LanguageSelectHasNoStepIndicator(t *testing.T) {
	m := setup.New(fstest.MapFS{}, "dev")
	m = toLanguageSelect(t, m)

	view := m.View()
	if strings.Contains(view, "step") {
		t.Errorf("language screen must not show a step indicator (unnumbered, like MainMenu), got:\n%s", view)
	}
	for _, label := range []string{"English", "Español"} {
		if !strings.Contains(view, label) {
			t.Errorf("language screen missing native option label %q, got:\n%s", label, view)
		}
	}
	if !strings.Contains(view, "(•)") {
		t.Errorf("language screen missing selected radio indicator '(•)', got:\n%s", view)
	}
}
