// Package memory defines the cross-session memory port and its adapters.
// The core never imports a concrete memory vendor; all memory backends are
// accessed exclusively through the Provider interface.
package memory

import (
	"context"
	"time"
)

// Provider is the port for cross-session organizational memory.
// All operations are best-effort: callers MUST NOT abort a run on error.
// Search returns an empty slice (not error) when nothing matches.
type Provider interface {
	Save(ctx context.Context, entry Entry) error
	Search(ctx context.Context, query string) ([]Entry, error)
	Get(ctx context.Context, id string) (*Entry, error)
	Name() string
}

// Entry is one structured knowledge record.
type Entry struct {
	// ID is set on Save: NullProvider=relative filepath, Engram=observation ID.
	ID       string
	Title    string
	Type     EntryType
	TopicKey string
	Content  EntryContent
	Metadata map[string]string
	SavedAt  time.Time
}

// EntryContent mirrors Engram's mem_save What/Why/Where/Learned contract.
type EntryContent struct {
	What    string `yaml:"what"`
	Why     string `yaml:"why"`
	Where   string `yaml:"where"`
	Learned string `yaml:"learned,omitempty"`
}

// EntryType is a typed string enum for the kind of knowledge record.
type EntryType string

const (
	EntryTypeDecision     EntryType = "decision"
	EntryTypeArchitecture EntryType = "architecture"
	EntryTypeDiscovery    EntryType = "discovery"
	EntryTypePattern      EntryType = "pattern"
	EntryTypeConfig       EntryType = "config"
	EntryTypePreference   EntryType = "preference"
)
