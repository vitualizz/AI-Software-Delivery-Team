package memory_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/vitualizz/ai-software-delivery-team/internal/memory"
)

// ------------------------------------------------------------------ helpers --

func tempRoot(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	return dir
}

// ------------------------------------------------------------------ NullProvider --

func TestNullProvider_Name(t *testing.T) {
	p := memory.NullProvider{}
	if got := p.Name(); got != "null" {
		t.Errorf("Name: got %q, want %q", got, "null")
	}
}

func TestNullProvider_Name_WithRoot(t *testing.T) {
	p := memory.NewNullProvider(tempRoot(t))
	if got := p.Name(); got != "null" {
		t.Errorf("Name: got %q, want %q", got, "null")
	}
}

func TestNullProvider_ZeroValue_Degrades(t *testing.T) {
	p := memory.NullProvider{}
	ctx := context.Background()

	// Save must not error (no-op).
	if err := p.Save(ctx, memory.Entry{Title: "test", Type: memory.EntryTypeDecision}); err != nil {
		t.Errorf("zero-value Save: unexpected error: %v", err)
	}

	// Search must return empty slice, not error.
	results, err := p.Search(ctx, "anything")
	if err != nil {
		t.Errorf("zero-value Search: unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("zero-value Search: expected empty, got %d entries", len(results))
	}

	// Get must return an error (no root configured).
	_, err = p.Get(ctx, "some-id")
	if err == nil {
		t.Error("zero-value Get: expected error, got nil")
	}
}

func TestNullProvider_Save_WritesYAML(t *testing.T) {
	root := tempRoot(t)
	p := memory.NewNullProvider(root)
	ctx := context.Background()

	entry := memory.Entry{
		Title:    "Auth decision",
		Type:     memory.EntryTypeDecision,
		TopicKey: "architect/constraints-analysis",
		Content: memory.EntryContent{
			What:  "Use JWT for authentication",
			Why:   "Stateless sessions required",
			Where: ".asdt/artifacts/auth-change",
		},
	}

	if err := p.Save(ctx, entry); err != nil {
		t.Fatalf("Save: %v", err)
	}

	// Verify a .yaml file was written under .asdt/runs/.
	runsDir := filepath.Join(root, ".asdt", "runs")
	var yamlFiles []string
	_ = filepath.WalkDir(runsDir, func(path string, d os.DirEntry, err error) error {
		if err == nil && !d.IsDir() && strings.HasSuffix(path, ".yaml") {
			yamlFiles = append(yamlFiles, path)
		}
		return nil
	})

	if len(yamlFiles) == 0 {
		t.Fatal("Save: expected at least one .yaml file under .asdt/runs/")
	}

	// The slug must contain "architect-constraints-analysis".
	found := false
	for _, f := range yamlFiles {
		if strings.Contains(filepath.Base(f), "architect-constraints-analysis") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Save: expected file with slug 'architect-constraints-analysis', got %v", yamlFiles)
	}

	// Read file contents and verify fields.
	data, err := os.ReadFile(yamlFiles[0])
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	content := string(data)
	for _, want := range []string{"Auth decision", "architect/constraints-analysis", "JWT", "Stateless"} {
		if !strings.Contains(content, want) {
			t.Errorf("YAML content: expected %q, not found in:\n%s", want, content)
		}
	}
}

func TestNullProvider_Search_ReturnsMatchingEntries(t *testing.T) {
	root := tempRoot(t)
	p := memory.NewNullProvider(root)
	ctx := context.Background()

	_ = p.Save(ctx, memory.Entry{
		Title:    "Authentication decision",
		Type:     memory.EntryTypeDecision,
		TopicKey: "auth/decision",
		Content:  memory.EntryContent{What: "Use JWT for authentication sessions"},
	})
	_ = p.Save(ctx, memory.Entry{
		Title:    "Database schema",
		Type:     memory.EntryTypeArchitecture,
		TopicKey: "db/schema",
		Content:  memory.EntryContent{What: "Postgres with JSONB columns"},
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

func TestNullProvider_Search_ReturnsEmpty_WhenNoMatch(t *testing.T) {
	root := tempRoot(t)
	p := memory.NewNullProvider(root)
	ctx := context.Background()

	_ = p.Save(ctx, memory.Entry{
		Title:   "Database schema",
		TopicKey: "db/schema",
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

func TestNullProvider_Get_ReturnsByID(t *testing.T) {
	root := tempRoot(t)
	p := memory.NewNullProvider(root)
	ctx := context.Background()

	entry := memory.Entry{
		Title:    "Security finding",
		Type:     memory.EntryTypeDiscovery,
		TopicKey: "security/finding",
		Content: memory.EntryContent{
			What:  "SQL injection in user input",
			Why:   "Unparameterized queries",
			Where: ".asdt/artifacts/security-change",
		},
	}
	if err := p.Save(ctx, entry); err != nil {
		t.Fatalf("Save: %v", err)
	}

	// Find the file that was written.
	runsDir := filepath.Join(root, ".asdt", "runs")
	var id string
	_ = filepath.WalkDir(runsDir, func(path string, d os.DirEntry, err error) error {
		if err == nil && !d.IsDir() && strings.HasSuffix(path, ".yaml") {
			// ID is relative to .asdt/.
			rel, relErr := filepath.Rel(filepath.Join(root, ".asdt"), path)
			if relErr == nil {
				id = rel
			}
		}
		return nil
	})

	if id == "" {
		t.Fatal("Get: no YAML file found to read back")
	}

	got, err := p.Get(ctx, id)
	if err != nil {
		t.Fatalf("Get(%q): %v", id, err)
	}
	if got == nil {
		t.Fatal("Get: returned nil entry")
	}
	if got.Title != entry.Title {
		t.Errorf("Get: Title = %q, want %q", got.Title, entry.Title)
	}
	if got.Content.What != entry.Content.What {
		t.Errorf("Get: Content.What = %q, want %q", got.Content.What, entry.Content.What)
	}
}

func TestNullProvider_Get_MissingFile_ReturnsError(t *testing.T) {
	root := tempRoot(t)
	p := memory.NewNullProvider(root)
	ctx := context.Background()

	_, err := p.Get(ctx, "runs/20000101-000000/nonexistent.yaml")
	if err == nil {
		t.Error("Get: expected error for missing file, got nil")
	}
}

func TestNullProvider_Roundtrip(t *testing.T) {
	root := tempRoot(t)
	p := memory.NewNullProvider(root)
	ctx := context.Background()

	original := memory.Entry{
		Title:    "Pattern: hexagonal architecture",
		Type:     memory.EntryTypePattern,
		TopicKey: "arch/pattern",
		Content: memory.EntryContent{
			What:    "All external I/O goes through ports",
			Why:     "Testability and isolation",
			Where:   "internal/",
			Learned: "Avoid importing adapters from core",
		},
	}

	if err := p.Save(ctx, original); err != nil {
		t.Fatalf("Save: %v", err)
	}

	results, err := p.Search(ctx, "hexagonal")
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("roundtrip Search: expected 1, got %d", len(results))
	}
	if results[0].Title != original.Title {
		t.Errorf("roundtrip: Title mismatch: got %q, want %q", results[0].Title, original.Title)
	}
	if results[0].Content.Learned != original.Content.Learned {
		t.Errorf("roundtrip: Learned mismatch: got %q, want %q", results[0].Content.Learned, original.Content.Learned)
	}
}

// ------------------------------------------------------------------ EngramProvider --

// mockCaller records calls made to it and returns pre-configured responses.
type mockCaller struct {
	calls    []calledWith
	response map[string]any
	err      error
}

type calledWith struct {
	tool string
	args map[string]any
}

func (m *mockCaller) Call(_ context.Context, tool string, args map[string]any) (map[string]any, error) {
	m.calls = append(m.calls, calledWith{tool: tool, args: args})
	return m.response, m.err
}

func TestEngramProvider_Name(t *testing.T) {
	p := memory.NewEngramProvider("test-project", &mockCaller{})
	if got := p.Name(); got != "engram" {
		t.Errorf("Name: got %q, want %q", got, "engram")
	}
}

func TestEngramProvider_DelegatesSave_CallsMCPCaller(t *testing.T) {
	caller := &mockCaller{response: map[string]any{"id": "42"}}
	p := memory.NewEngramProvider("my-project", caller)
	ctx := context.Background()

	entry := memory.Entry{
		Title:    "Architecture decision",
		Type:     memory.EntryTypeArchitecture,
		TopicKey: "runs/architect/my-change",
		Content: memory.EntryContent{
			What:  "Use hexagonal architecture",
			Why:   "Testability",
			Where: ".asdt/artifacts/my-change",
		},
	}

	if err := p.Save(ctx, entry); err != nil {
		t.Fatalf("Save: %v", err)
	}

	if len(caller.calls) != 1 {
		t.Fatalf("expected 1 Call, got %d", len(caller.calls))
	}
	call := caller.calls[0]
	if call.tool != "mem_save" {
		t.Errorf("tool = %q, want %q", call.tool, "mem_save")
	}
	if call.args["title"] != entry.Title {
		t.Errorf("args[title] = %v, want %q", call.args["title"], entry.Title)
	}
	if call.args["topic_key"] != entry.TopicKey {
		t.Errorf("args[topic_key] = %v, want %q", call.args["topic_key"], entry.TopicKey)
	}
	if call.args["project"] != "my-project" {
		t.Errorf("args[project] = %v, want %q", call.args["project"], "my-project")
	}
	// Content must contain bold-label format.
	content, _ := call.args["content"].(string)
	if !strings.Contains(content, "**What**:") {
		t.Errorf("content missing **What** label: %q", content)
	}
}

func TestEngramProvider_DelegatesSearch_CallsMCPCaller(t *testing.T) {
	caller := &mockCaller{
		response: map[string]any{
			"results": []any{
				map[string]any{
					"id":    "42",
					"title": "Prior JWT decision",
					"type":  "decision",
				},
			},
		},
	}
	p := memory.NewEngramProvider("my-project", caller)
	ctx := context.Background()

	results, err := p.Search(ctx, "postgresql")
	if err != nil {
		t.Fatalf("Search: %v", err)
	}

	if len(caller.calls) != 1 {
		t.Fatalf("expected 1 Call, got %d", len(caller.calls))
	}
	call := caller.calls[0]
	if call.tool != "mem_search" {
		t.Errorf("tool = %q, want %q", call.tool, "mem_search")
	}
	if call.args["query"] != "postgresql" {
		t.Errorf("args[query] = %v, want %q", call.args["query"], "postgresql")
	}
	if call.args["project"] != "my-project" {
		t.Errorf("args[project] = %v, want %q", call.args["project"], "my-project")
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
}

func TestEngramProvider_DelegatesGet_CallsMCPCaller(t *testing.T) {
	caller := &mockCaller{
		response: map[string]any{
			"title":   "Arch decision",
			"type":    "architecture",
			"content": "**What**: Use ports\n**Why**: Testability\n**Where**: internal/",
		},
	}
	p := memory.NewEngramProvider("my-project", caller)
	ctx := context.Background()

	entry, err := p.Get(ctx, "42")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if entry == nil {
		t.Fatal("Get: returned nil entry")
	}

	if len(caller.calls) != 1 {
		t.Fatalf("expected 1 Call, got %d", len(caller.calls))
	}
	call := caller.calls[0]
	if call.tool != "mem_get_observation" {
		t.Errorf("tool = %q, want %q", call.tool, "mem_get_observation")
	}
	if call.args["id"] != "42" {
		t.Errorf("args[id] = %v, want %q", call.args["id"], "42")
	}
	if entry.Title != "Arch decision" {
		t.Errorf("entry.Title = %q, want %q", entry.Title, "Arch decision")
	}
	if entry.Content.What != "Use ports" {
		t.Errorf("entry.Content.What = %q, want %q", entry.Content.What, "Use ports")
	}
}

func TestEngramProvider_Search_PropagatesTransportError(t *testing.T) {
	caller := &mockCaller{err: &testError{"transport failure"}}
	p := memory.NewEngramProvider("my-project", caller)

	_, err := p.Search(context.Background(), "anything")
	if err == nil {
		t.Error("Search: expected error on transport failure, got nil")
	}
}

// testError is a simple error type for tests.
type testError struct{ msg string }

func (e *testError) Error() string { return e.msg }
