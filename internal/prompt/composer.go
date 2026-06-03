package prompt

import "strings"

// Compose assembles the final prompt from layers in the specified order:
//  1. Role fragment (persona + workflow)
//  2. Skill fragments (capability fragments, in order)
//  3. Artifact context (serialized input artifacts)
//  4. Platform context (platform.yaml summary)
//
// It returns the composed prompt string and a Manifest of all fragment versions.
//
// Context budget: this implementation concatenates layers naively.
// TODO(T-1-5): add token estimation (chars/4) and progressive summarization
// for lower-priority layers when budget is exceeded.
func Compose(role Fragment, skills []Fragment, artifactContext string, platformContext string) (string, Manifest) {
	manifest := NewManifest()

	var parts []string

	// Layer 1: role
	parts = append(parts, role.Content)
	manifest.Set(role.Name, role.Version)

	// Layer 2: skill fragments
	for _, s := range skills {
		parts = append(parts, s.Content)
		manifest.Set(s.Name, s.Version)
	}

	// Layer 3: artifact context
	if artifactContext != "" {
		parts = append(parts, artifactContext)
	}

	// Layer 4: platform context
	if platformContext != "" {
		parts = append(parts, platformContext)
	}

	return strings.Join(parts, "\n\n---\n\n"), manifest
}
