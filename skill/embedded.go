// Package skill exposes the embedded prompt filesystem for the /asdt skill tree.
// The unified FS contains specialist SKILL.md files, specialist-scoped skills,
// shared skills, and the legacy prompts/ subtree for backward compatibility.
package skill

import (
	"embed"
	"io/fs"
)

//go:embed SKILL.md _shared developer ux-ui architect qa security prompts
var skillFS embed.FS

// FS returns the full embedded skill tree rooted at skill/.
// Paths: "developer/SKILL.md", "_shared/skills/platform-context.md", etc.
// This is the production FS injected into prompt.EmbeddedRegistry.
func FS() fs.FS {
	return skillFS
}

// PromptSubFS returns an fs.FS rooted at skill/prompts/.
// Callers use paths like "roles/requirements/role.md" directly.
// Kept for backward compatibility during migration to the specialist layout.
// Will be removed in Phase 4 after internal/requirements/ and internal/developer/
// are deleted.
func PromptSubFS() fs.FS {
	sub, err := fs.Sub(skillFS, "prompts")
	if err != nil {
		// This can only panic if the embed directive is misconfigured — a build-time error.
		panic("skill: failed to sub-FS at prompts: " + err.Error())
	}
	return sub
}
