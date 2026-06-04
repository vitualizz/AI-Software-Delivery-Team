package memory

import (
	"context"
	"errors"
)

// ErrNotImplemented is returned by adapters that are registered but not yet wired.
var ErrNotImplemented = errors.New("memory: provider not implemented")

// EngramProvider is the opt-in Engram memory adapter.
// The stub returns ErrNotImplemented for every operation until the Engram
// MCP/CLI integration is wired. It exists so the composition root can register
// an Engram-backed provider behind config without the core ever importing Engram.
// Real implementation lands in a later change.
type EngramProvider struct {
	endpoint string
	project  string
}

// NewEngramProvider constructs an EngramProvider with the given endpoint and project.
func NewEngramProvider(endpoint, project string) *EngramProvider {
	return &EngramProvider{endpoint: endpoint, project: project}
}

func (e *EngramProvider) Load(_ context.Context, _ string) ([]byte, error) {
	return nil, ErrNotImplemented
}

func (e *EngramProvider) Save(_ context.Context, _ string, _ []byte) error {
	return ErrNotImplemented
}

func (e *EngramProvider) Name() string { return "engram" }

var _ Provider = (*EngramProvider)(nil)
