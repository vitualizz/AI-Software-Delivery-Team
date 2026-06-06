package memory_test

import (
	"context"
	"strings"
	"testing"

	"github.com/vitualizz/ai-software-delivery-team/internal/memory"
)

// ------------------------------------------------------------------ InMemoryProvider --

func TestInMemoryProvider_Name(t *testing.T) {
	p := memory.NewInMemoryProvider()
	if got := p.Name(); got != "inmemory" {
		t.Errorf("Name: got %q, want %q", got, "inmemory")
	}
}

func TestInMemoryProvider_Save_RoundTrip(t *testing.T) {
	p := memory.NewInMemoryProvider()
	ctx := context.Background()

	entry := memory.Entry{
		Title:    "Architecture decision",
		Type:     memory.EntryTypeArchitecture,
		TopicKey: "arch/pattern",
		Content: memory.EntryContent{
			What:    "Use hexagonal architecture",
			Why:     "Testability and isolation",
			Where:   "internal/",
			Learned: "Avoid importing adapters from core",
		},
	}

	if err := p.Save(ctx, entry); err != nil {
		t.Fatalf("Save: %v", err)
	}

	// Get should return the saved entry by ID.
	// We need to find the ID — search for it first.
	results, err := p.Search(ctx, "hexagonal")
	if err != nil {
		t.Fatalf("Search after Save: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("Search: expected 1 result, got %d", len(results))
	}
	if results[0].ID == "" {
		t.Error("saved entry must have a non-empty ID")
	}
	if results[0].Title != entry.Title {
		t.Errorf("Title: got %q, want %q", results[0].Title, entry.Title)
	}
	if results[0].Content.Learned != entry.Content.Learned {
		t.Errorf("Content.Learned: got %q, want %q", results[0].Content.Learned, entry.Content.Learned)
	}
}

func TestInMemoryProvider_Get_ByID(t *testing.T) {
	p := memory.NewInMemoryProvider()
	ctx := context.Background()

	entry := memory.Entry{
		Title:    "Security finding",
		Type:     memory.EntryTypeDiscovery,
		TopicKey: "security/finding",
		Content: memory.EntryContent{
			What:  "SQL injection in user input",
			Why:   "Unparameterized queries",
			Where: "internal/db",
		},
	}

	if err := p.Save(ctx, entry); err != nil {
		t.Fatalf("Save: %v", err)
	}

	results, err := p.Search(ctx, "SQL injection")
	if err != nil || len(results) != 1 {
		t.Fatalf("Search: err=%v, len=%d", err, len(results))
	}
	id := results[0].ID

	got, err := p.Get(ctx, id)
	if err != nil {
		t.Fatalf("Get(%q): %v", id, err)
	}
	if got == nil {
		t.Fatal("Get: returned nil")
	}
	if got.Title != entry.Title {
		t.Errorf("Get: Title = %q, want %q", got.Title, entry.Title)
	}
	if got.Content.What != entry.Content.What {
		t.Errorf("Get: Content.What = %q, want %q", got.Content.What, entry.Content.What)
	}
}

func TestInMemoryProvider_Get_Missing(t *testing.T) {
	p := memory.NewInMemoryProvider()
	ctx := context.Background()

	_, err := p.Get(ctx, "mem-9999")
	if err == nil {
		t.Error("Get: expected error for unknown ID, got nil")
	}
}

func TestInMemoryProvider_Search_SubstringMatch(t *testing.T) {
	p := memory.NewInMemoryProvider()
	ctx := context.Background()

	_ = p.Save(ctx, memory.Entry{
		Title:   "Authentication decision",
		Type:    memory.EntryTypeDecision,
		Content: memory.EntryContent{What: "Use JWT for authentication sessions"},
	})
	_ = p.Save(ctx, memory.Entry{
		Title:   "Database schema",
		Type:    memory.EntryTypeArchitecture,
		Content: memory.EntryContent{What: "Postgres with JSONB columns"},
	})

	results, err := p.Search(ctx, "authentication")
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("Search: expected 1 result, got %d", len(results))
	}
	if !strings.Contains(results[0].Title, "Authentication") {
		t.Errorf("Search: unexpected result title %q", results[0].Title)
	}
}

func TestInMemoryProvider_Search_CaseInsensitive(t *testing.T) {
	p := memory.NewInMemoryProvider()
	ctx := context.Background()

	_ = p.Save(ctx, memory.Entry{
		Title:   "JWT Auth Pattern",
		Content: memory.EntryContent{What: "Use JWT tokens for stateless auth"},
	})

	// Query in all-caps should still match.
	results, err := p.Search(ctx, "JWT AUTH")
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("Search (case-insensitive): expected 1 result, got %d", len(results))
	}
}

func TestInMemoryProvider_Search_MatchesOnContent(t *testing.T) {
	p := memory.NewInMemoryProvider()
	ctx := context.Background()

	_ = p.Save(ctx, memory.Entry{
		Title:   "Unrelated title",
		Content: memory.EntryContent{What: "The real keyword is hexagonal architecture"},
	})

	results, err := p.Search(ctx, "hexagonal")
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("Search on Content.What: expected 1 result, got %d", len(results))
	}
}

func TestInMemoryProvider_Search_ReturnsEmpty_WhenNoMatch(t *testing.T) {
	p := memory.NewInMemoryProvider()
	ctx := context.Background()

	_ = p.Save(ctx, memory.Entry{
		Title:   "Database schema",
		Content: memory.EntryContent{What: "Postgres with JSONB columns"},
	})

	results, err := p.Search(ctx, "authentication jwt oauth")
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("Search: expected empty, got %d entries", len(results))
	}
}

func TestInMemoryProvider_Search_EmptyStore_ReturnsEmpty(t *testing.T) {
	p := memory.NewInMemoryProvider()
	ctx := context.Background()

	results, err := p.Search(ctx, "anything")
	if err != nil {
		t.Fatalf("Search on empty store: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected empty slice, got %d", len(results))
	}
}

func TestInMemoryProvider_Save_MultipleEntries(t *testing.T) {
	p := memory.NewInMemoryProvider()
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		_ = p.Save(ctx, memory.Entry{
			Title:   "entry",
			Content: memory.EntryContent{What: "shared keyword content"},
		})
	}

	results, err := p.Search(ctx, "shared keyword")
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) != 5 {
		t.Errorf("expected 5 results, got %d", len(results))
	}
}
