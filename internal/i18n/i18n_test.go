package i18n_test

import (
	"reflect"
	"testing"

	"github.com/vitualizz/ai-software-delivery-team/internal/i18n"
)

func TestActive_DefaultsToEnglish(t *testing.T) {
	t.Setenv("LC_ALL", "")
	t.Setenv("LANG", "")
	t.Setenv("LANGUAGE", "")
	t.Setenv("ASDT_LANG", "")

	got := i18n.Active()
	if got.Installer.BtnContinue != i18n.English.Installer.BtnContinue {
		t.Errorf("empty locale: expected English, got BtnContinue=%q", got.Installer.BtnContinue)
	}
}

func TestActive_SpanishFromLANG(t *testing.T) {
	t.Setenv("LC_ALL", "")
	t.Setenv("LANG", "es_AR.UTF-8")
	t.Setenv("ASDT_LANG", "")

	got := i18n.Active()
	if got.Installer.BtnContinue != i18n.Spanish.Installer.BtnContinue {
		t.Errorf("LANG=es_AR.UTF-8: expected Spanish, got BtnContinue=%q", got.Installer.BtnContinue)
	}
}

func TestActive_SpanishFromLCAll(t *testing.T) {
	t.Setenv("LC_ALL", "es_ES.UTF-8")
	t.Setenv("ASDT_LANG", "")

	got := i18n.Active()
	if got.Installer.BtnContinue != i18n.Spanish.Installer.BtnContinue {
		t.Errorf("LC_ALL=es_ES: expected Spanish, got BtnContinue=%q", got.Installer.BtnContinue)
	}
}

func TestActive_ASDTLANGOverridesSystem(t *testing.T) {
	t.Setenv("LC_ALL", "")
	t.Setenv("LANG", "en_US.UTF-8")
	t.Setenv("ASDT_LANG", "es")

	got := i18n.Active()
	if got.Installer.BtnContinue != i18n.Spanish.Installer.BtnContinue {
		t.Errorf("ASDT_LANG=es: expected Spanish, got BtnContinue=%q", got.Installer.BtnContinue)
	}
}

func TestActive_FallsBackToEnglishOnUnknownLocale(t *testing.T) {
	t.Setenv("LC_ALL", "")
	t.Setenv("LANG", "xx_XX.UTF-8")
	t.Setenv("ASDT_LANG", "")

	got := i18n.Active()
	if got.Installer.BtnContinue != i18n.English.Installer.BtnContinue {
		t.Errorf("unknown locale: expected English fallback, got BtnContinue=%q", got.Installer.BtnContinue)
	}
}

func TestActiveCode_ResolvesBaseCode(t *testing.T) {
	t.Setenv("LC_ALL", "")
	t.Setenv("LANG", "es_AR.UTF-8")
	t.Setenv("LANGUAGE", "")
	t.Setenv("ASDT_LANG", "")
	if got := i18n.ActiveCode(); got != "es" {
		t.Errorf("ActiveCode() with LANG=es_AR = %q, want %q", got, "es")
	}

	t.Setenv("LANG", "")
	if got := i18n.ActiveCode(); got != "en" {
		t.Errorf("ActiveCode() with empty locale = %q, want %q", got, "en")
	}
}

func TestForCode_KnownAndUnknown(t *testing.T) {
	if got := i18n.ForCode("es"); got.Installer.BtnContinue != i18n.Spanish.Installer.BtnContinue {
		t.Errorf("ForCode(\"es\"): expected Spanish, got BtnContinue=%q", got.Installer.BtnContinue)
	}
	if got := i18n.ForCode("xx"); got.Installer.BtnContinue != i18n.English.Installer.BtnContinue {
		t.Errorf("ForCode(\"xx\"): expected English fallback, got BtnContinue=%q", got.Installer.BtnContinue)
	}
}

func TestEnglishCatalogComplete(t *testing.T) {
	walkCatalog(t, "English", i18n.English)
}

func TestSpanishCatalogComplete(t *testing.T) {
	walkCatalog(t, "Spanish", i18n.Spanish)
}

// walkCatalog walks every top-level struct field of Catalog (Installer,
// Dashboard, Personas, and any future feature-area struct) and verifies that
// every string field has a non-empty value. A missing field in a new catalog
// is a compile-time zero-value that only surfaces at runtime — this test
// catches it early without needing one assertion helper per struct.
func walkCatalog(t *testing.T, locale string, c i18n.Catalog) {
	t.Helper()
	root := reflect.ValueOf(c)
	rootType := root.Type()
	for i := range root.NumField() {
		section := root.Field(i)
		if section.Kind() != reflect.Struct {
			continue
		}
		sectionName := rootType.Field(i).Name
		sectionType := section.Type()
		for j := range section.NumField() {
			if section.Field(j).Kind() == reflect.String && section.Field(j).String() == "" {
				t.Errorf("%s catalog: %s.%s is empty", locale, sectionName, sectionType.Field(j).Name)
			}
		}
	}
}

func TestPersonaDescription_KnownIDs(t *testing.T) {
	for _, id := range []string{"sky", "toffy", "atreus", "babi", "lee-palacios"} {
		if got := i18n.English.PersonaDescription(id); got == "" {
			t.Errorf("English.PersonaDescription(%q) = \"\", want non-empty", id)
		}
		if got := i18n.Spanish.PersonaDescription(id); got == "" {
			t.Errorf("Spanish.PersonaDescription(%q) = \"\", want non-empty", id)
		}
	}
	if en, es := i18n.English.PersonaDescription("sky"), i18n.Spanish.PersonaDescription("sky"); en == es {
		t.Errorf("Spanish persona description for sky should differ from English, both = %q", en)
	}
}

func TestPersonaDescription_EmptyFieldFallsBackToEnglish(t *testing.T) {
	incomplete := i18n.Spanish
	incomplete.Personas.Sky = "" // simulate a catalog missing one persona description
	got := incomplete.PersonaDescription("sky")
	want := i18n.English.Personas.Sky
	if got != want {
		t.Errorf("PersonaDescription with empty field = %q, want English fallback %q", got, want)
	}
}

func TestPersonaDescription_UnknownIDReturnsEmpty(t *testing.T) {
	if got := i18n.English.PersonaDescription("nyan-cat"); got != "" {
		t.Errorf("PersonaDescription(unknown) = %q, want \"\"", got)
	}
}
