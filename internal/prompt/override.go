package prompt

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// OverrideResolver checks override directories before falling back to the
// packaged registry. Precedence order (first match wins):
//
//  1. Project-local:  .asdt/prompts/{role}/role.md  or  .asdt/prompts/skills/{name}.md
//  2. User-global:    ~/.config/asdt/prompts/{role}/role.md
//  3. Packaged:       embedded FS via the backing SkillRegistry
type OverrideResolver struct {
	localDir  string         // absolute path to .asdt/prompts/
	globalDir string         // absolute path to ~/.config/asdt/prompts/
	fallback  SkillRegistry  // packaged registry (EmbeddedRegistry)
}

// NewOverrideResolver constructs an OverrideResolver.
// localDir is the .asdt/prompts/ path; globalDir is ~/.config/asdt/prompts/.
// fallback is used when no override is found.
func NewOverrideResolver(localDir, globalDir string, fallback SkillRegistry) *OverrideResolver {
	return &OverrideResolver{
		localDir:  localDir,
		globalDir: globalDir,
		fallback:  fallback,
	}
}

// Role resolves the role fragment with override precedence.
func (r *OverrideResolver) Role(name string) (Fragment, error) {
	relPath := filepath.Join(name, "role.md")

	if f, err := readOverride(r.localDir, relPath, name, SourceLocal); err == nil {
		return f, nil
	}
	if f, err := readOverride(r.globalDir, relPath, name, SourceGlobal); err == nil {
		return f, nil
	}
	return r.fallback.Role(name)
}

// Skill resolves the skill fragment with override precedence.
// Checks local then global override directories before falling back to the packaged registry.
func (r *OverrideResolver) Skill(name string) (Fragment, error) {
	relPath := filepath.Join("skills", name+".md")

	if f, err := readOverride(r.localDir, relPath, name, SourceLocal); err == nil {
		return f, nil
	}
	if f, err := readOverride(r.globalDir, relPath, name, SourceGlobal); err == nil {
		return f, nil
	}
	return r.fallback.Skill(name)
}

// ScopedSkill resolves a skill for a specific specialist with override precedence.
// Override path: {specialistID}/skills/{skillName}.md in local then global dirs.
// Falls back to the packaged registry's ScopedSkill.
func (r *OverrideResolver) ScopedSkill(specialistID, skillName string) (Fragment, error) {
	relPath := filepath.Join(specialistID, "skills", skillName+".md")

	if f, err := readOverride(r.localDir, relPath, skillName, SourceLocal); err == nil {
		return f, nil
	}
	if f, err := readOverride(r.globalDir, relPath, skillName, SourceGlobal); err == nil {
		return f, nil
	}
	return r.fallback.ScopedSkill(specialistID, skillName)
}

// Version returns the version of the resolved fragment (whichever source wins).
func (r *OverrideResolver) Version(name string) (string, error) {
	if frag, err := r.Role(name); err == nil {
		return frag.Version, nil
	}
	if frag, err := r.Skill(name); err == nil {
		return frag.Version, nil
	}
	return "", fmt.Errorf("override resolver version %q: not found", name)
}

// readOverride attempts to read a file at dir/relPath.
// Returns an error when the file does not exist or cannot be read.
func readOverride(dir, relPath, name string, source Source) (Fragment, error) {
	if dir == "" {
		return Fragment{}, fmt.Errorf("override dir not set")
	}
	full := filepath.Join(dir, relPath)
	data, err := os.ReadFile(full)
	if err != nil {
		return Fragment{}, err
	}
	return NewFragment(name, string(data), source), nil
}

// DefaultGlobalDir returns the default user-global override directory.
// It expands ~/ to the user's home directory.
func DefaultGlobalDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "asdt", "prompts")
}

// Ensure OverrideResolver satisfies SkillRegistry.
var _ SkillRegistry = (*OverrideResolver)(nil)

// Ensure EmbeddedRegistry satisfies fs.FS-based reading.
var _ fs.FS = (fs.FS)(nil)
