// Package memory defines the cross-session memory port and its adapters.
// The core never imports a concrete memory vendor; all memory backends are
// accessed exclusively through the Provider interface.
package memory

import "context"

// Provider is the port for cross-session memory.
// Missing key: Load returns nil,nil (not an error).
// All operations are best-effort in the application layer — callers must not
// abort a run when a Provider returns an error.
type Provider interface {
	Load(ctx context.Context, key string) ([]byte, error)
	Save(ctx context.Context, key string, data []byte) error
	Name() string
}
