// Package prompt implements the layered prompt composition engine.
// It manages a registry of role and skill fragments, resolves overrides,
// and composes them into a final prompt string with a version manifest.
//
// NOTE: go:embed wiring happens in T-1-5. Use SetEmbedFS to inject the embedded FS
// during initialization, allowing tests to use a plain os.DirFS.
package prompt

import (
	"crypto/sha256"
	"fmt"
)

// Source identifies where a Fragment was resolved from.
type Source string

const (
	// SourcePackaged means the fragment came from the embedded (go:embed) FS.
	SourcePackaged Source = "packaged"

	// SourceLocal means the fragment came from the project-local .asdt/prompts/ override.
	SourceLocal Source = "local"

	// SourceGlobal means the fragment came from the user-global ~/.config/asdt/prompts/ override.
	SourceGlobal Source = "global"
)

// Fragment is a single resolved prompt unit (role or skill) with provenance metadata.
type Fragment struct {
	// Name is the canonical identifier (e.g. "requirements", "user-story-writing").
	Name string

	// Content is the full text of the fragment.
	Content string

	// Source identifies where this fragment was resolved from.
	Source Source

	// Version is the SHA-256 of Content truncated to 8 hex characters.
	// Used in the Manifest for drift detection and prompt_version hashing.
	Version string
}

// NewFragment constructs a Fragment and computes its Version from content.
func NewFragment(name, content string, source Source) Fragment {
	return Fragment{
		Name:    name,
		Content: content,
		Source:  source,
		Version: fragmentVersion(content),
	}
}

// fragmentVersion returns the first 8 hex characters of the SHA-256 of content.
func fragmentVersion(content string) string {
	sum := sha256.Sum256([]byte(content))
	return fmt.Sprintf("%x", sum[:4]) // 4 bytes = 8 hex chars
}
