package memory

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode"

	"gopkg.in/yaml.v3"
)

// NullProvider is the default memory adapter.
// When constructed with NewNullProvider(root), it persists entries to the
// filesystem under {root}/.asdt/runs/{runID}/{slug}.yaml.
// The zero-value NullProvider degrades to a no-op so existing test fixtures
// that use memory.NullProvider{} continue to work without a root path.
type NullProvider struct {
	root  string // absolute path to the project root (parent of .asdt/)
	runID string // YYYYMMDD-HHMMSS, computed once at construction
}

// NewNullProvider constructs a NullProvider that persists entries under
// {root}/.asdt/runs/{runID}/. root should be the project root (parent of .asdt/).
func NewNullProvider(root string) *NullProvider {
	return &NullProvider{
		root:  root,
		runID: time.Now().UTC().Format("20060102-150405"),
	}
}

// Save marshals entry to YAML and writes it to {root}/.asdt/runs/{runID}/{slug}.yaml.
// The returned ID is the relative path from the .asdt directory: runs/{runID}/{slug}.yaml.
// When root is empty (zero-value) this is a no-op and returns nil.
func (p *NullProvider) Save(_ context.Context, entry Entry) error {
	if p.root == "" {
		return nil
	}

	dir := filepath.Join(p.root, ".asdt", "runs", p.runID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("memory: mkdir %s: %w", dir, err)
	}

	slug := topicKeySlug(entry.TopicKey)
	if slug == "" {
		slug = topicKeySlug(entry.Title)
	}
	if slug == "" {
		slug = "entry"
	}

	filename := slug + ".yaml"
	filePath := filepath.Join(dir, filename)

	if entry.SavedAt.IsZero() {
		entry.SavedAt = time.Now().UTC()
	}

	relID := filepath.Join("runs", p.runID, filename)
	entry.ID = relID

	data, err := yaml.Marshal(entry)
	if err != nil {
		return fmt.Errorf("memory: marshal entry: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0o644); err != nil {
		return fmt.Errorf("memory: write entry: %w", err)
	}

	return nil
}

// Search walks {root}/.asdt/runs/ and returns entries whose Title or Content.What
// contains any word from the query (case-insensitive substring match).
// Returns an empty (non-nil) slice when nothing matches. Individual file read/parse
// errors are silently skipped (degrade gracefully).
// When root is empty (zero-value) returns empty slice.
func (p *NullProvider) Search(_ context.Context, query string) ([]Entry, error) {
	if p.root == "" {
		return []Entry{}, nil
	}

	runsDir := filepath.Join(p.root, ".asdt", "runs")
	words := strings.Fields(strings.ToLower(query))
	if len(words) == 0 {
		return []Entry{}, nil
	}

	var results []Entry

	_ = filepath.WalkDir(runsDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // skip unreadable directories
		}
		if d.IsDir() || !strings.HasSuffix(path, ".yaml") {
			return nil
		}

		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return nil // degrade: skip unreadable files
		}

		var entry Entry
		if parseErr := yaml.Unmarshal(data, &entry); parseErr != nil {
			return nil // degrade: skip malformed files
		}

		haystack := strings.ToLower(entry.Title + " " + entry.Content.What)
		for _, word := range words {
			if strings.Contains(haystack, word) {
				results = append(results, entry)
				break
			}
		}
		return nil
	})

	return results, nil
}

// Get reads the entry at the given relative ID (relative from {root}/.asdt/).
// Example: id = "runs/20260605-120000/architect-constraints-analysis.yaml"
// When root is empty (zero-value) returns nil, error.
func (p *NullProvider) Get(_ context.Context, id string) (*Entry, error) {
	if p.root == "" {
		return nil, fmt.Errorf("memory: NullProvider has no root configured")
	}

	filePath := filepath.Join(p.root, ".asdt", id)
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("memory: entry %q not found: %w", id, err)
	}

	var entry Entry
	if err := yaml.Unmarshal(data, &entry); err != nil {
		return nil, fmt.Errorf("memory: parse entry %q: %w", id, err)
	}
	return &entry, nil
}

// Name returns "null".
func (p NullProvider) Name() string { return "null" }

// topicKeySlug converts a topic key like "architect/constraints-analysis" into
// a safe filename slug "architect-constraints-analysis" by replacing all
// non-alphanumeric characters with "-", collapsing consecutive dashes, and
// trimming leading/trailing dashes.
func topicKeySlug(s string) string {
	var b strings.Builder
	prevDash := false
	for _, r := range strings.ToLower(s) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
			prevDash = false
		} else {
			if !prevDash {
				b.WriteByte('-')
				prevDash = true
			}
		}
	}
	result := strings.Trim(b.String(), "-")
	return result
}

var _ Provider = (*NullProvider)(nil)
