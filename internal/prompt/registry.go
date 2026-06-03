package prompt

import (
	"fmt"
	"io/fs"
	"path"
)

// SkillRegistry is the port for resolving prompt fragments by name.
// Implementations may read from an embedded FS, an override directory,
// or a combination via OverrideResolver.
type SkillRegistry interface {
	// Role returns the role fragment for the given agent name.
	// e.g. Role("requirements") returns the PM/BA persona prompt.
	Role(name string) (Fragment, error)

	// Skill returns a capability fragment by name.
	// e.g. Skill("user-story-writing") returns the skill fragment.
	Skill(name string) (Fragment, error)

	// Version returns the version hash for a named fragment.
	// Returns an empty string if the fragment does not exist.
	Version(name string) (string, error)
}

// EmbeddedRegistry implements SkillRegistry by reading from an fs.FS.
// The FS is expected to contain:
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

// Role reads roles/{name}/role.md from the backing FS.
func (r *EmbeddedRegistry) Role(name string) (Fragment, error) {
	p := path.Join("roles", name, "role.md")
	content, err := fs.ReadFile(r.fsys, p)
	if err != nil {
		return Fragment{}, fmt.Errorf("registry role %q: %w", name, err)
	}
	return NewFragment(name, string(content), SourcePackaged), nil
}

// Skill reads skills/{name}.md from the backing FS.
func (r *EmbeddedRegistry) Skill(name string) (Fragment, error) {
	p := path.Join("skills", name+".md")
	content, err := fs.ReadFile(r.fsys, p)
	if err != nil {
		return Fragment{}, fmt.Errorf("registry skill %q: %w", name, err)
	}
	return NewFragment(name, string(content), SourcePackaged), nil
}

// Version returns the version of a named fragment.
// It tries Role first, then Skill. Returns an empty string on failure.
func (r *EmbeddedRegistry) Version(name string) (string, error) {
	if frag, err := r.Role(name); err == nil {
		return frag.Version, nil
	}
	if frag, err := r.Skill(name); err == nil {
		return frag.Version, nil
	}
	return "", fmt.Errorf("registry version %q: fragment not found", name)
}
