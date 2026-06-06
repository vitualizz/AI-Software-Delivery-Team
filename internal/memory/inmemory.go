package memory

import (
	"context"
	"fmt"
	"strings"
	"sync"
)

// InMemoryProvider is an in-memory implementation of the Provider interface.
// It is intended for use in tests only — it is never wired into the production binary.
// All operations are thread-safe via an internal RWMutex.
type InMemoryProvider struct {
	mu      sync.RWMutex
	entries map[string]Entry
	counter int
}

// NewInMemoryProvider constructs an empty InMemoryProvider ready for use in tests.
func NewInMemoryProvider() *InMemoryProvider {
	return &InMemoryProvider{
		entries: make(map[string]Entry),
	}
}

// Save stores the entry under a generated ID of the form "mem-N" (where N is a
// monotonically increasing counter). The ID is set on the entry before storage.
func (p *InMemoryProvider) Save(_ context.Context, entry Entry) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.counter++
	id := fmt.Sprintf("mem-%d", p.counter)
	entry.ID = id
	p.entries[id] = entry
	return nil
}

// Search returns all entries whose Title or Content.What contains any word from
// the query as a case-insensitive substring match. Returns a non-nil empty slice
// when nothing matches.
func (p *InMemoryProvider) Search(_ context.Context, query string) ([]Entry, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	words := strings.Fields(strings.ToLower(query))
	if len(words) == 0 {
		return []Entry{}, nil
	}

	var results []Entry
	for _, e := range p.entries {
		haystack := strings.ToLower(e.Title + " " + e.Content.What)
		for _, word := range words {
			if strings.Contains(haystack, word) {
				results = append(results, e)
				break
			}
		}
	}
	if results == nil {
		return []Entry{}, nil
	}
	return results, nil
}

// Get returns the entry stored under the given ID, or an error if it does not exist.
func (p *InMemoryProvider) Get(_ context.Context, id string) (*Entry, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	entry, ok := p.entries[id]
	if !ok {
		return nil, fmt.Errorf("memory: inmemory: entry %q not found", id)
	}
	copy := entry
	return &copy, nil
}

// Name returns "inmemory".
func (p *InMemoryProvider) Name() string { return "inmemory" }

var _ Provider = (*InMemoryProvider)(nil)
