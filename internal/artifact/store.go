package artifact

import "context"

// Store is the driven port for reading and writing artifact envelopes
// under the .asdt/ boundary. Implementations are filesystem-backed (FSStore)
// or in-memory (for tests). The caller is responsible for the typed Envelope —
// this interface uses any to stay decoupled from generic instantiation limits.
type Store interface {
	// Read deserializes the artifact at change/artifactType into out.
	// out MUST be a pointer to an Envelope[T] or compatible struct.
	Read(ctx context.Context, change, artifactType string, out any) error

	// Write serializes env and persists it at change/artifactType.
	// env MUST be an Envelope[T] or compatible struct.
	Write(ctx context.Context, change, artifactType string, env any) error

	// List returns the artifact types available for a given change.
	List(ctx context.Context, change string) ([]string, error)

	// Exists returns true when change/artifactType has been written.
	Exists(change, artifactType string) bool
}
