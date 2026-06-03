package prompt

import (
	"crypto/sha256"
	"fmt"
	"sort"
	"strings"
)

// Manifest maps fragment name to version hash.
// It is returned by Compose and written to EnvelopeHeader.PromptVersion
// so any artifact can be reproduced and drift can be detected.
type Manifest struct {
	fragments map[string]string
}

// NewManifest creates an empty Manifest.
func NewManifest() Manifest {
	return Manifest{fragments: make(map[string]string)}
}

// Set records the version for a named fragment.
func (m *Manifest) Set(name, version string) {
	if m.fragments == nil {
		m.fragments = make(map[string]string)
	}
	m.fragments[name] = version
}

// Get returns the version for a named fragment, or empty string if absent.
func (m Manifest) Get(name string) string {
	return m.fragments[name]
}

// All returns a copy of the fragment→version map.
func (m Manifest) All() map[string]string {
	out := make(map[string]string, len(m.fragments))
	for k, v := range m.fragments {
		out[k] = v
	}
	return out
}

// Hash returns a stable SHA-256 hash of all fragment versions, sorted by name.
// Same fragment set + same versions always produces the same hash.
func (m Manifest) Hash() string {
	keys := make([]string, 0, len(m.fragments))
	for k := range m.fragments {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var sb strings.Builder
	for _, k := range keys {
		sb.WriteString(k)
		sb.WriteByte(':')
		sb.WriteString(m.fragments[k])
		sb.WriteByte('\n')
	}
	sum := sha256.Sum256([]byte(sb.String()))
	return fmt.Sprintf("%x", sum[:4]) // 8 hex chars
}
