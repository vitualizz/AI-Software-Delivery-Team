package memory

import (
	"context"
	"fmt"
	"strings"
)

// MCPCaller is the minimal port over an MCP client.
// One method, easy to mock in tests.
type MCPCaller interface {
	Call(ctx context.Context, tool string, args map[string]any) (map[string]any, error)
}

// EngramProvider is the opt-in Engram memory adapter.
// It delegates all operations to the Engram MCP via the injected MCPCaller port.
// EngramProvider never performs local filesystem operations.
type EngramProvider struct {
	project string
	caller  MCPCaller
}

// NewEngramProvider constructs an EngramProvider with the given project identifier
// and MCPCaller. The project is passed to every Engram operation to scope queries.
func NewEngramProvider(project string, caller MCPCaller) *EngramProvider {
	return &EngramProvider{project: project, caller: caller}
}

// Save delegates to mem_save with the entry's fields mapped to Engram arguments.
// It renders EntryContent in the Engram bold-label format (What/Why/Where/Learned).
// On success, it attempts to parse the observation ID from the response.
func (e *EngramProvider) Save(ctx context.Context, entry Entry) error {
	content := renderEntryContent(entry.Content)

	args := map[string]any{
		"title":         entry.Title,
		"type":          string(entry.Type),
		"topic_key":     entry.TopicKey,
		"project":       e.project,
		"content":       content,
		"capture_prompt": false,
	}

	_, err := e.caller.Call(ctx, "mem_save", args)
	if err != nil {
		return fmt.Errorf("memory: engram save: %w", err)
	}
	return nil
}

// Search delegates to mem_search and maps results to []Entry.
// Returns an empty (non-nil) slice on no results. Only transport errors propagate.
func (e *EngramProvider) Search(ctx context.Context, query string) ([]Entry, error) {
	args := map[string]any{
		"query":   query,
		"project": e.project,
	}

	resp, err := e.caller.Call(ctx, "mem_search", args)
	if err != nil {
		return nil, fmt.Errorf("memory: engram search: %w", err)
	}

	return mapSearchResponse(resp), nil
}

// Get delegates to mem_get_observation by ID and maps the result to *Entry.
func (e *EngramProvider) Get(ctx context.Context, id string) (*Entry, error) {
	args := map[string]any{
		"id": id,
	}

	resp, err := e.caller.Call(ctx, "mem_get_observation", args)
	if err != nil {
		return nil, fmt.Errorf("memory: engram get %q: %w", id, err)
	}

	if resp == nil {
		return nil, fmt.Errorf("memory: engram get %q: empty response", id)
	}

	entry := &Entry{ID: id}
	if title, ok := resp["title"].(string); ok {
		entry.Title = title
	}
	if t, ok := resp["type"].(string); ok {
		entry.Type = EntryType(t)
	}
	if tk, ok := resp["topic_key"].(string); ok {
		entry.TopicKey = tk
	}
	if content, ok := resp["content"].(string); ok {
		entry.Content = parseEntryContent(content)
	}
	return entry, nil
}

// Name returns "engram".
func (e *EngramProvider) Name() string { return "engram" }

// renderEntryContent formats EntryContent as Engram's bold-label format:
//
//	**What**: ...\n**Why**: ...\n**Where**: ...\n**Learned**: ...
func renderEntryContent(c EntryContent) string {
	var sb strings.Builder
	sb.WriteString("**What**: ")
	sb.WriteString(c.What)
	sb.WriteString("\n**Why**: ")
	sb.WriteString(c.Why)
	sb.WriteString("\n**Where**: ")
	sb.WriteString(c.Where)
	if c.Learned != "" {
		sb.WriteString("\n**Learned**: ")
		sb.WriteString(c.Learned)
	}
	return sb.String()
}

// parseEntryContent is a best-effort parser for the bold-label format.
// It extracts What/Why/Where/Learned from the rendered string.
func parseEntryContent(raw string) EntryContent {
	var c EntryContent
	lines := strings.Split(raw, "\n")
	for _, line := range lines {
		switch {
		case strings.HasPrefix(line, "**What**: "):
			c.What = strings.TrimPrefix(line, "**What**: ")
		case strings.HasPrefix(line, "**Why**: "):
			c.Why = strings.TrimPrefix(line, "**Why**: ")
		case strings.HasPrefix(line, "**Where**: "):
			c.Where = strings.TrimPrefix(line, "**Where**: ")
		case strings.HasPrefix(line, "**Learned**: "):
			c.Learned = strings.TrimPrefix(line, "**Learned**: ")
		}
	}
	return c
}

// mapSearchResponse extracts []Entry from a mem_search response.
// Tolerates missing or unexpected fields (best-effort mapping).
func mapSearchResponse(resp map[string]any) []Entry {
	// Engram search returns results under a "result" key or as a top-level slice.
	if resp == nil {
		return []Entry{}
	}

	var rawItems []any

	switch v := resp["results"].(type) {
	case []any:
		rawItems = v
	default:
		// Try top-level "result" string — no structured items available.
		return []Entry{}
	}

	entries := make([]Entry, 0, len(rawItems))
	for _, item := range rawItems {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		entry := Entry{}
		if id, ok := m["id"].(string); ok {
			entry.ID = id
		}
		if title, ok := m["title"].(string); ok {
			entry.Title = title
		}
		if t, ok := m["type"].(string); ok {
			entry.Type = EntryType(t)
		}
		if tk, ok := m["topic_key"].(string); ok {
			entry.TopicKey = tk
		}
		if content, ok := m["content"].(string); ok {
			entry.Content = parseEntryContent(content)
		}
		entries = append(entries, entry)
	}
	return entries
}

var _ Provider = (*EngramProvider)(nil)
