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

// Active returns the Catalog for the user's locale.
// ASDT_LANG takes priority over system locale detection so it can be
// overridden without changing the system locale (e.g. ASDT_LANG=es).
func Active() Catalog {
	if override := os.Getenv("ASDT_LANG"); override != "" {
		if tag, err := language.Parse(override); err == nil {
			matched, _, _ := supported.Match(tag)
			return catalogFor(matched)
		}
	}
	return catalogFor(detect())
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
