package memory

import "context"

// NullProvider is the default memory adapter.
// It does nothing: Load returns nil,nil; Save is a no-op.
// Memory is optional infrastructure — NullProvider keeps the default path
// fully offline and deterministic.
type NullProvider struct{}

func (NullProvider) Load(_ context.Context, _ string) ([]byte, error) { return nil, nil }
func (NullProvider) Save(_ context.Context, _ string, _ []byte) error  { return nil }
func (NullProvider) Name() string                                       { return "null" }

var _ Provider = NullProvider{}
