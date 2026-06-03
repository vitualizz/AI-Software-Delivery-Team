package artifact_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/vitualizz/ai-software-delivery-team/internal/artifact"
)

// examplePayload is a simple struct used for round-trip tests.
type examplePayload struct {
	Title   string   `yaml:"title"`
	Tags    []string `yaml:"tags"`
	Version int      `yaml:"version"`
}

func validHeader() artifact.EnvelopeHeader {
	return artifact.EnvelopeHeader{
		SchemaVersion: artifact.CurrentSchemaVersion,
		Agent:         "test-agent",
		ChangeID:      "change-001",
		CreatedAt:     time.Now().UTC(),
		PromptVersion: "abc12345",
		InputRefs:     []string{},
	}
}

// TestRoundTrip verifies that an Envelope[T] serializes and deserializes
// correctly via FSStore.
func TestRoundTrip(t *testing.T) {
	dir := t.TempDir()
	store := artifact.NewFSStore(dir)
	ctx := context.Background()

	original := artifact.Envelope[examplePayload]{
		EnvelopeHeader: validHeader(),
		Payload: examplePayload{
			Title:   "hello world",
			Tags:    []string{"go", "test"},
			Version: 42,
		},
	}

	if err := store.Write(ctx, "my-change", "test-artifact", original); err != nil {
		t.Fatalf("Write: %v", err)
	}

	var restored artifact.Envelope[examplePayload]
	if err := store.Read(ctx, "my-change", "test-artifact", &restored); err != nil {
		t.Fatalf("Read: %v", err)
	}

	if restored.Agent != original.Agent {
		t.Errorf("Agent: got %q, want %q", restored.Agent, original.Agent)
	}
	if restored.ChangeID != original.ChangeID {
		t.Errorf("ChangeID: got %q, want %q", restored.ChangeID, original.ChangeID)
	}
	if restored.SchemaVersion != original.SchemaVersion {
		t.Errorf("SchemaVersion: got %q, want %q", restored.SchemaVersion, original.SchemaVersion)
	}
	if restored.Payload.Title != original.Payload.Title {
		t.Errorf("Payload.Title: got %q, want %q", restored.Payload.Title, original.Payload.Title)
	}
	if restored.Payload.Version != original.Payload.Version {
		t.Errorf("Payload.Version: got %d, want %d", restored.Payload.Version, original.Payload.Version)
	}
}

// TestExistsAndList verifies Store.Exists and Store.List behavior.
func TestExistsAndList(t *testing.T) {
	dir := t.TempDir()
	store := artifact.NewFSStore(dir)
	ctx := context.Background()

	if store.Exists("change-x", "artifact-a") {
		t.Fatal("Exists should be false before writing")
	}

	env := artifact.Envelope[examplePayload]{
		EnvelopeHeader: validHeader(),
		Payload:        examplePayload{Title: "a"},
	}
	if err := store.Write(ctx, "change-x", "artifact-a", env); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if err := store.Write(ctx, "change-x", "artifact-b", env); err != nil {
		t.Fatalf("Write: %v", err)
	}

	if !store.Exists("change-x", "artifact-a") {
		t.Fatal("Exists should be true after writing")
	}

	list, err := store.List(ctx, "change-x")
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(list) != 2 {
		t.Errorf("List len: got %d, want 2; items: %v", len(list), list)
	}
}

// TestValidate_AllFieldsPresent verifies that a complete header passes validation.
func TestValidate_AllFieldsPresent(t *testing.T) {
	h := validHeader()
	if err := artifact.Validate(h); err != nil {
		t.Errorf("expected nil, got: %v", err)
	}
}

// TestValidate_MissingFields verifies that each missing required field
// is reported individually in the error.
func TestValidate_MissingFields(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*artifact.EnvelopeHeader)
		wantErr string
	}{
		{
			name:    "missing schema_version",
			mutate:  func(h *artifact.EnvelopeHeader) { h.SchemaVersion = "" },
			wantErr: "schema_version",
		},
		{
			name:    "missing agent",
			mutate:  func(h *artifact.EnvelopeHeader) { h.Agent = "" },
			wantErr: "agent",
		},
		{
			name:    "missing change_id",
			mutate:  func(h *artifact.EnvelopeHeader) { h.ChangeID = "" },
			wantErr: "change_id",
		},
		{
			name:    "missing created_at",
			mutate:  func(h *artifact.EnvelopeHeader) { h.CreatedAt = time.Time{} },
			wantErr: "created_at",
		},
		{
			name:    "missing prompt_version",
			mutate:  func(h *artifact.EnvelopeHeader) { h.PromptVersion = "" },
			wantErr: "prompt_version",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := validHeader()
			tc.mutate(&h)
			err := artifact.Validate(h)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Errorf("error %q does not mention field %q", err.Error(), tc.wantErr)
			}
		})
	}
}

// TestValidate_SchemaVersionMismatch verifies that a wrong schema_version
// returns ErrSchemaVersionMismatch with both versions in the message.
func TestValidate_SchemaVersionMismatch(t *testing.T) {
	h := validHeader()
	h.SchemaVersion = "99"
	err := artifact.Validate(h)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, artifact.ErrSchemaVersionMismatch) {
		t.Errorf("expected ErrSchemaVersionMismatch, got: %v", err)
	}
	if !strings.Contains(err.Error(), "99") {
		t.Errorf("error should mention the bad version: %v", err)
	}
	if !strings.Contains(err.Error(), artifact.CurrentSchemaVersion) {
		t.Errorf("error should mention the expected version: %v", err)
	}
}
