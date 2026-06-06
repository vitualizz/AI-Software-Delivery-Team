package memory_test

import (
	"context"
	"strings"
	"testing"

	"github.com/vitualizz/ai-software-delivery-team/internal/memory"
)

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
