package llm

import (
	"context"
	"strings"
	"sync"
)

// MockProvider is a test-only Provider implementation that delivers
// scripted response strings and records all calls for assertion.
//
// Usage:
//
//	p := NewMockProvider(WithScriptedResponses("hello world", "second response"))
//	resp, _ := p.Complete(ctx, req)
//	// resp.Content == "hello world"
type MockProvider struct {
	mu       sync.Mutex
	scripts  []string // queued scripted responses
	index    int      // next script to deliver
	Calls    [][]Message
}

// MockOption is a functional option for MockProvider.
type MockOption func(*MockProvider)

// WithScriptedResponses enqueues scripted full-text responses.
// Each Complete/Stream call consumes one entry in order; the last
// entry is repeated when the queue is exhausted.
func WithScriptedResponses(responses ...string) MockOption {
	return func(p *MockProvider) {
		p.scripts = append(p.scripts, responses...)
	}
}

// NewMockProvider constructs a MockProvider with the given options.
func NewMockProvider(opts ...MockOption) *MockProvider {
	p := &MockProvider{}
	for _, o := range opts {
		o(p)
	}
	return p
}

// nextScript returns the next scripted response, repeating the last one
// when all scripts are consumed.
func (p *MockProvider) nextScript() string {
	if len(p.scripts) == 0 {
		return ""
	}
	if p.index >= len(p.scripts) {
		return p.scripts[len(p.scripts)-1]
	}
	s := p.scripts[p.index]
	p.index++
	return s
}

// Stream delivers the scripted response as a sequence of word-level Chunks,
// finishing with a Done chunk. Calls are recorded.
func (p *MockProvider) Stream(_ context.Context, req Request) (<-chan Chunk, error) {
	p.mu.Lock()
	p.Calls = append(p.Calls, req.Messages)
	script := p.nextScript()
	p.mu.Unlock()

	ch := make(chan Chunk)
	go func() {
		defer close(ch)
		words := strings.Fields(script)
		for i, w := range words {
			text := w
			if i < len(words)-1 {
				text += " "
			}
			ch <- Chunk{Content: text, Done: false}
		}
		ch <- Chunk{Done: true}
	}()
	return ch, nil
}

// Complete aggregates all chunks from Stream into a single Response.
func (p *MockProvider) Complete(ctx context.Context, req Request) (Response, error) {
	ch, err := p.Stream(ctx, req)
	if err != nil {
		return Response{}, err
	}
	var sb strings.Builder
	for c := range ch {
		if !c.Done {
			sb.WriteString(c.Content)
		}
	}
	return Response{Content: sb.String()}, nil
}

// Name returns the identifier for this provider.
func (p *MockProvider) Name() string { return "mock" }
