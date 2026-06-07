package prompt

import (
	"fmt"
	"io/fs"
	"path"
	"strings"

	"github.com/vitualizz/ai-software-delivery-team/skill"
)

// SkillRegistry is the port for resolving prompt fragments by name.
// Implementations may read from an embedded FS, an override directory,
// or a combination via OverrideResolver.
type SkillRegistry interface {
	// Role returns the role fragment for the given specialist or agent name.
	// For specialists: resolves skill/{id}/SKILL.md (body only, frontmatter stripped).
	// For legacy agents: resolves roles/{name}/role.md.
	Role(name string) (Fragment, error)

	// Skill returns a shared capability fragment by name.
	// Resolves from asdt-shared/skills/{name}.md (specialist layout) or
	// skills/{name}.md (legacy layout).
	Skill(name string) (Fragment, error)

	// ScopedSkill resolves a skill within an explicit specialist scope.
	// Tries {specialistID}/skills/{skillName}.md first, then falls back to Skill(skillName).
	ScopedSkill(specialistID, skillName string) (Fragment, error)

	// Version returns the version hash for a named fragment.
	// Returns an empty string if the fragment does not exist.
	Version(name string) (string, error)
}

// EmbeddedRegistry implements SkillRegistry by reading from an fs.FS.
// It supports two layouts:
//
// Specialist layout (new):
//
//	{id}/SKILL.md          — role fragments (frontmatter stripped)
//	{id}/skills/{name}.md  — scoped specialist skills
//	asdt-shared/skills/{name}.md — shared skills
//
// Legacy layout (backward compat):
//
//	roles/{name}/role.md
//	skills/{name}.md
type EmbeddedRegistry struct {
	fsys fs.FS
}

// NewEmbeddedRegistry constructs an EmbeddedRegistry backed by the given FS.
func NewEmbeddedRegistry(fsys fs.FS) *EmbeddedRegistry {
	return &EmbeddedRegistry{fsys: fsys}
}

// embeddedSpecialistDir maps a bare specialist ID (as carried in SKILL.md
// frontmatter, e.g. "developer", "security", "asdt-init") to the top-level
// directory name used in the embedded skill tree.
//
// The embedded tree's specialist directories are installed-name siblings
// (asdt-architect, asdt-developer, asdt-qa, asdt-security, asdt-ux-ui,
// asdt-init, ...) while the canonical specialist ID stays bare (the
// frontmatter "specialist-id" field, used throughout the pipeline/state
// machine). This mapping bridges the two:
//
//   - IDs already prefixed with "asdt-" (e.g. "asdt-init") map verbatim —
//     no double-prefixing.
//   - Every other bare ID is prefixed with "asdt-" to reach its embedded
//     directory (e.g. "developer" -> "asdt-developer").
//
// This is purely an embedded-FS path concern — it does not change the
// specialist ID surface used elsewhere (pipeline state, frontmatter, etc.).
func embeddedSpecialistDir(specialistID string) string {
	if strings.HasPrefix(specialistID, "asdt-") {
		return specialistID
	}
	return "asdt-" + specialistID
}

// Role resolves the role fragment. Resolution order:
//  1. {embeddedSpecialistDir(name)}/SKILL.md  (specialist layout — body only, frontmatter stripped)
//  2. roles/{name}/role.md  (legacy layout — full content)
func (r *EmbeddedRegistry) Role(name string) (Fragment, error) {
	// Try specialist layout first, mapping the bare specialist ID to its
	// embedded sibling directory name.
	specialist := path.Join(embeddedSpecialistDir(name), "SKILL.md")
	if content, err := fs.ReadFile(r.fsys, specialist); err == nil {
		return NewFragment(name, stripFrontmatter(string(content)), SourcePackaged), nil
	}

	// Fall back to legacy layout.
	legacy := path.Join("roles", name, "role.md")
	content, err := fs.ReadFile(r.fsys, legacy)
	if err != nil {
		return Fragment{}, fmt.Errorf("registry role %q: not found in specialist or legacy layout: %w", name, err)
	}
	return NewFragment(name, string(content), SourcePackaged), nil
}

// Skill resolves a shared skill by name. Resolution order:
//  1. asdt-shared/skills/{name}.md  (specialist layout)
//  2. skills/{name}.md              (legacy layout)
func (r *EmbeddedRegistry) Skill(name string) (Fragment, error) {
	// Try specialist asdt-shared layout first.
	shared := path.Join("asdt-shared", "skills", name+".md")
	if content, err := fs.ReadFile(r.fsys, shared); err == nil {
		return NewFragment(name, string(content), SourcePackaged), nil
	}

	// Fall back to legacy layout.
	legacy := path.Join("skills", name+".md")
	content, err := fs.ReadFile(r.fsys, legacy)
	if err != nil {
		return Fragment{}, fmt.Errorf("registry skill %q: not found in shared or legacy layout: %w", name, err)
	}
	return NewFragment(name, string(content), SourcePackaged), nil
}

// ScopedSkill resolves a skill scoped to a specific specialist. Resolution order:
//  1. {embeddedSpecialistDir(specialistID)}/skills/{skillName}.md  (specialist-specific)
//  2. Skill(skillName)                                             (shared fallback)
func (r *EmbeddedRegistry) ScopedSkill(specialistID, skillName string) (Fragment, error) {
	p := path.Join(embeddedSpecialistDir(specialistID), "skills", skillName+".md")
	if content, err := fs.ReadFile(r.fsys, p); err == nil {
		return NewFragment(skillName, string(content), SourcePackaged), nil
	}
	frag, err := r.Skill(skillName)
	if err != nil {
		return Fragment{}, fmt.Errorf("registry scoped skill %q/%q: %w", specialistID, skillName, err)
	}
	return frag, nil
}

// Version returns the version of a named fragment.
// It tries Role first, then Skill. Returns an error if not found.
func (r *EmbeddedRegistry) Version(name string) (string, error) {
	if frag, err := r.Role(name); err == nil {
		return frag.Version, nil
	}
	if frag, err := r.Skill(name); err == nil {
		return frag.Version, nil
	}
	return "", fmt.Errorf("registry version %q: fragment not found", name)
}

// stripFrontmatter removes a YAML frontmatter block (--- ... ---) from the
// beginning of a Markdown string. Returns the body unchanged if no frontmatter.
func stripFrontmatter(content string) string {
	content = strings.TrimSpace(content)
	if !strings.HasPrefix(content, "---") {
		return content
	}
	// Skip the opening "---" line.
	rest := content[3:]
	// Find the closing "---".
	end := strings.Index(rest, "---")
	if end == -1 {
		return content // no closing delimiter — return as-is
	}
	body := rest[end+3:]
	return strings.TrimSpace(body)
}

// DefaultEmbeddedRegistry returns an EmbeddedRegistry backed by the full
// skill/ embedded FS (skill.FS()). This is the registry used at runtime.
// Tests use NewEmbeddedRegistry with a custom fs.FS.
func DefaultEmbeddedRegistry() *EmbeddedRegistry {
	return NewEmbeddedRegistry(skill.FS())
}
