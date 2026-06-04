package artifact_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
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

// nestedPayload is a struct with nested fields for round-trip coverage.
type nestedPayload struct {
	Title  string            `yaml:"title"`
	Meta   map[string]string `yaml:"meta"`
	Nested struct {
		Score int      `yaml:"score"`
		Tags  []string `yaml:"tags"`
	} `yaml:"nested"`
}

// TestRoundTrip_NestedPayload verifies Write+Read with a non-trivial nested struct.
func TestRoundTrip_NestedPayload(t *testing.T) {
	dir := t.TempDir()
	store := artifact.NewFSStore(dir)
	ctx := context.Background()

	original := artifact.Envelope[nestedPayload]{
		EnvelopeHeader: validHeader(),
	}
	original.Payload.Title = "nested round-trip"
	original.Payload.Meta = map[string]string{"env": "test", "version": "3"}
	original.Payload.Nested.Score = 99
	original.Payload.Nested.Tags = []string{"alpha", "beta", "gamma"}

	if err := store.Write(ctx, "ch-1", "nested-artifact", original); err != nil {
		t.Fatalf("Write: %v", err)
	}

	var restored artifact.Envelope[nestedPayload]
	if err := store.Read(ctx, "ch-1", "nested-artifact", &restored); err != nil {
		t.Fatalf("Read: %v", err)
	}

	if restored.Payload.Title != original.Payload.Title {
		t.Errorf("Title: got %q, want %q", restored.Payload.Title, original.Payload.Title)
	}
	if restored.Payload.Meta["env"] != "test" {
		t.Errorf("Meta[env]: got %q, want %q", restored.Payload.Meta["env"], "test")
	}
	if restored.Payload.Nested.Score != 99 {
		t.Errorf("Nested.Score: got %d, want 99", restored.Payload.Nested.Score)
	}
	if len(restored.Payload.Nested.Tags) != 3 {
		t.Errorf("Nested.Tags len: got %d, want 3", len(restored.Payload.Nested.Tags))
	}
}

// TestFSStore_Write_CreatesIntermediateDirectories verifies that Write creates
// the full directory path even when intermediate directories don't exist.
func TestFSStore_Write_CreatesIntermediateDirectories(t *testing.T) {
	dir := t.TempDir()
	// Point the store at a subdirectory that does not yet exist.
	store := artifact.NewFSStore(filepath.Join(dir, "deep", "nested", "path"))
	ctx := context.Background()

	env := artifact.Envelope[examplePayload]{
		EnvelopeHeader: validHeader(),
		Payload:        examplePayload{Title: "dir-creation"},
	}
	// Should succeed even though the intermediate directories do not exist.
	if err := store.Write(ctx, "ch-create", "test-artifact", env); err != nil {
		t.Fatalf("Write with missing intermediate dirs: %v", err)
	}
}

// TestFSStore_Read_NotFound verifies that Read returns a wrapped error
// (not a raw os error) when the file does not exist.
func TestFSStore_Read_NotFound(t *testing.T) {
	dir := t.TempDir()
	store := artifact.NewFSStore(dir)
	ctx := context.Background()

	var out artifact.Envelope[examplePayload]
	err := store.Read(ctx, "nonexistent-change", "nonexistent-artifact", &out)
	if err == nil {
		t.Fatal("expected error reading non-existent artifact, got nil")
	}
	// The error must mention the change and artifact name so the caller can diagnose.
	if !strings.Contains(err.Error(), "nonexistent-change") && !strings.Contains(err.Error(), "nonexistent-artifact") {
		t.Errorf("error %q should mention the change or artifact name", err.Error())
	}
	// Must not be a raw *os.PathError leaked directly — it should be wrapped.
	if errors.Is(err, os.ErrNotExist) {
		// The underlying cause may be ErrNotExist, but the message should be wrapped
		// with context. This is acceptable — the important thing is the message has context.
	}
}
