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

func TestEnglishCatalogComplete(t *testing.T) {
	assertInstallerComplete(t, "English", i18n.English)
}

func TestSpanishCatalogComplete(t *testing.T) {
	assertInstallerComplete(t, "Spanish", i18n.Spanish)
}

// assertInstallerComplete verifies that every string field in InstallerStrings
// has a non-empty value. A missing field in a new catalog is a compile-time
// zero-value that only surfaces at runtime — this test catches it early.
func assertInstallerComplete(t *testing.T, name string, c i18n.Catalog) {
	t.Helper()
	v := reflect.ValueOf(c.Installer)
	typ := v.Type()
	for i := range v.NumField() {
		if v.Field(i).Kind() == reflect.String && v.Field(i).String() == "" {
			t.Errorf("%s catalog: InstallerStrings.%s is empty", name, typ.Field(i).Name)
		}
	}
}
