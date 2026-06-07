// Package skill exposes the embedded prompt filesystem for the /asdt skill tree.
// The unified FS contains specialist SKILL.md files, specialist-scoped skills,
// and shared skills.
package skill

import (
	"embed"
	"io/fs"
)

//go:embed SKILL.md asdt-shared asdt-developer asdt-ux-ui asdt-architect asdt-qa asdt-security asdt-init
var skillFS embed.FS

// FS returns the full embedded skill tree rooted at skill/.
// Paths: "asdt-developer/SKILL.md", "asdt-shared/skills/platform-context.md", etc.
// This is the production FS injected into prompt.EmbeddedRegistry.
func FS() fs.FS {
	return skillFS
}

// PromptSubFS returns an fs.FS backed by the full skill FS.
// Kept for backward compatibility — callers that used the old prompts/ subtree
// should migrate to the specialist skill layout.
//
// Deprecated: use FS() directly.
func PromptSubFS() fs.FS {
	return skillFS
}
