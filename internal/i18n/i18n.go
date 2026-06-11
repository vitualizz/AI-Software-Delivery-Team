package i18n

import (
	"os"

	"golang.org/x/text/language"
)

// catalogs maps BCP 47 base language codes to their Catalog.
// To add a new language: add the catalog var and register it here.
var catalogs = map[string]Catalog{
	"en": English,
	"es": Spanish,
}

// resolve returns the supported language tag for the current environment.
// ASDT_LANG takes priority over system locale detection so it can be
// overridden without changing the system locale (e.g. ASDT_LANG=es).
func resolve() language.Tag {
	if override := os.Getenv("ASDT_LANG"); override != "" {
		if tag, err := language.Parse(override); err == nil {
			matched, _, _ := supported.Match(tag)
			return matched
		}
	}
	return detect()
}

// Active returns the Catalog for the user's locale.
func Active() Catalog {
	return catalogFor(resolve())
}

// ActiveCode returns the resolved base language code for the user's locale
// (e.g. "en", "es"). Codes without a registered catalog resolve to "en".
func ActiveCode() string {
	base, _ := resolve().Base()
	if _, ok := catalogs[base.String()]; ok {
		return base.String()
	}
	return "en"
}

// ForCode returns the Catalog registered for the given base language code,
// falling back to English for unknown codes. Useful when the language choice
// comes from persisted state or UI selection instead of env detection.
func ForCode(code string) Catalog {
	if c, ok := catalogs[code]; ok {
		return c
	}
	return English
}

// For returns the Catalog for the given language tag, falling back to English.
// Useful in tests and CLI flag handling to bypass env-var detection.
func For(tag language.Tag) Catalog { return catalogFor(tag) }

func catalogFor(tag language.Tag) Catalog {
	base, _ := tag.Base()
	if c, ok := catalogs[base.String()]; ok {
		return c
	}
	return English
}
