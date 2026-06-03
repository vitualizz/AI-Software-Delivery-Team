// Package skill exposes the embedded prompt filesystem for the /asdt skill.
// The go:embed directive is valid here because prompts/ is a direct subdirectory
// of skill/, which is where this file lives.
package skill

import (
	"embed"
	"io/fs"
)

//go:embed prompts
var promptFS embed.FS

// PromptSubFS returns an fs.FS rooted at skill/prompts/.
// Callers use paths like "roles/requirements/role.md" directly.
// This is the production FS injected into prompt.EmbeddedRegistry.
func PromptSubFS() fs.FS {
	sub, err := fs.Sub(promptFS, "prompts")
	if err != nil {
		// This can only panic if the embed directive is misconfigured — a build-time error.
		panic("skill: failed to sub-FS at prompts: " + err.Error())
	}
	return sub
}
