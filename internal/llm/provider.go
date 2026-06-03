package llm

import (
	"context"
	"errors"
)

// ErrNotImplemented is returned by stub providers that have not yet been wired.
var ErrNotImplemented = errors.New("llm provider not implemented")

// Provider is the driven port for executing LLM completions.
// Streaming is the primary path; Complete is a convenience wrapper
// that aggregates all chunks into a single Response.
type Provider interface {
	// Stream sends req to the provider and returns a channel of incremental
	// Chunks. The channel is closed after the final chunk (Done == true).
	// The caller must drain the channel or cancel ctx to avoid goroutine leaks.
	Stream(ctx context.Context, req Request) (<-chan Chunk, error)

	// Complete sends req and returns the fully aggregated Response.
	// It is equivalent to consuming the Stream channel until Done.
	Complete(ctx context.Context, req Request) (Response, error)

	// Name returns a human-readable identifier for this provider.
	Name() string
}
