package i18n

import (
	"os"
	"strings"

	"golang.org/x/text/language"
)

// supported lists the locales the application ships catalogs for.
// English must be first — language.NewMatcher uses position 0 as the fallback.
var supported = language.NewMatcher([]language.Tag{
	language.English,
	language.Spanish,
})

// detect returns the best-matching supported language tag from the system
// locale environment variables. Reads LC_ALL, LANG, and LANGUAGE in that
// order; falls back to English if none is parseable or supported.
func detect() language.Tag {
	for _, env := range []string{"LC_ALL", "LANG", "LANGUAGE"} {
		val := os.Getenv(env)
		if val == "" || val == "C" || val == "POSIX" {
			continue
		}
		// LANGUAGE can be a colon-separated priority list; take the first tag.
		if idx := strings.IndexByte(val, ':'); idx != -1 {
			val = val[:idx]
		}
		// Strip encoding suffix: "es_AR.UTF-8" → "es_AR".
		if idx := strings.IndexByte(val, '.'); idx != -1 {
			val = val[:idx]
		}
		// Normalize POSIX underscores to BCP 47 hyphens: "es_AR" → "es-AR".
		val = strings.ReplaceAll(val, "_", "-")
		tag, err := language.Parse(val)
		if err != nil {
			continue
		}
		matched, _, _ := supported.Match(tag)
		return matched
	}
	return language.English
}
