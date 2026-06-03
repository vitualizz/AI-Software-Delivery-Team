package llm_test

import (
	"context"
	"strings"
	"testing"

	"github.com/vitualizz/ai-software-delivery-team/internal/llm"
)

func TestMockProvider_Name(t *testing.T) {
	p := llm.NewMockProvider()
	if p.Name() != "mock" {
		t.Errorf("Name: got %q, want %q", p.Name(), "mock")
	}
}

// TestMockProvider_ScriptedResponsesInOrder verifies that Complete delivers
// scripted responses in order and repeats the last when the queue is exhausted.
func TestMockProvider_ScriptedResponsesInOrder(t *testing.T) {
	ctx := context.Background()
	p := llm.NewMockProvider(
		llm.WithScriptedResponses("first response", "second response"),
	)

	req := llm.Request{Messages: []llm.Message{{Role: "user", Content: "hi"}}}

	resp1, err := p.Complete(ctx, req)
	if err != nil {
		t.Fatalf("Complete 1: %v", err)
	}
	if resp1.Content != "first response" {
		t.Errorf("resp1: got %q, want %q", resp1.Content, "first response")
	}

	resp2, err := p.Complete(ctx, req)
	if err != nil {
		t.Fatalf("Complete 2: %v", err)
	}
	if resp2.Content != "second response" {
		t.Errorf("resp2: got %q, want %q", resp2.Content, "second response")
	}

	// Queue exhausted — repeats last.
	resp3, err := p.Complete(ctx, req)
	if err != nil {
		t.Fatalf("Complete 3: %v", err)
	}
	if resp3.Content != "second response" {
		t.Errorf("resp3 (repeat last): got %q, want %q", resp3.Content, "second response")
	}
}

// TestMockProvider_StreamDeliversChunks verifies that Stream sends individual
// chunks and terminates with a Done chunk.
func TestMockProvider_StreamDeliversChunks(t *testing.T) {
	ctx := context.Background()
	p := llm.NewMockProvider(llm.WithScriptedResponses("hello world"))

	req := llm.Request{Messages: []llm.Message{{Role: "user", Content: "go"}}}
	ch, err := p.Stream(ctx, req)
	if err != nil {
		t.Fatalf("Stream: %v", err)
	}

	var parts []string
	var doneCount int
	for c := range ch {
		if c.Done {
			doneCount++
		} else {
			parts = append(parts, c.Content)
		}
	}

	joined := strings.Join(parts, "")
	if !strings.Contains(joined, "hello") || !strings.Contains(joined, "world") {
		t.Errorf("Stream chunks reassembled to %q, expected 'hello world'", joined)
	}
	if doneCount != 1 {
		t.Errorf("expected exactly 1 Done chunk, got %d", doneCount)
	}
}

// TestMockProvider_CallRecorder verifies that all messages are recorded
// in Calls for test assertions.
func TestMockProvider_CallRecorder(t *testing.T) {
	ctx := context.Background()
	p := llm.NewMockProvider(llm.WithScriptedResponses("ok"))

	msgs := []llm.Message{
		{Role: "system", Content: "you are helpful"},
		{Role: "user", Content: "write code"},
	}
	if _, err := p.Complete(ctx, llm.Request{Messages: msgs}); err != nil {
		t.Fatalf("Complete: %v", err)
	}

	if len(p.Calls) != 1 {
		t.Fatalf("Calls len: got %d, want 1", len(p.Calls))
	}
	if len(p.Calls[0]) != 2 {
		t.Errorf("Calls[0] message count: got %d, want 2", len(p.Calls[0]))
	}
	if p.Calls[0][1].Content != "write code" {
		t.Errorf("Calls[0][1].Content: got %q, want %q", p.Calls[0][1].Content, "write code")
	}
}

// TestMockProvider_CompleteAggregatesStream verifies that Complete produces
// the same text as manually consuming the Stream channel.
func TestMockProvider_CompleteAggregatesStream(t *testing.T) {
	ctx := context.Background()
	script := "aggregate these words into one"

	// Stream path
	p1 := llm.NewMockProvider(llm.WithScriptedResponses(script))
	ch, err := p1.Stream(ctx, llm.Request{Messages: []llm.Message{{Role: "user", Content: "x"}}})
	if err != nil {
		t.Fatalf("Stream: %v", err)
	}
	var sb strings.Builder
	for c := range ch {
		if !c.Done {
			sb.WriteString(c.Content)
		}
	}
	fromStream := sb.String()

	// Complete path
	p2 := llm.NewMockProvider(llm.WithScriptedResponses(script))
	resp, err := p2.Complete(ctx, llm.Request{Messages: []llm.Message{{Role: "user", Content: "x"}}})
	if err != nil {
		t.Fatalf("Complete: %v", err)
	}

	if fromStream != resp.Content {
		t.Errorf("stream assembled %q, Complete returned %q", fromStream, resp.Content)
	}
}
