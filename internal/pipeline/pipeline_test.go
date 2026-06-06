package pipeline_test

import (
	"context"
	"testing"

	"github.com/vitualizz/ai-software-delivery-team/internal/artifact"
	"github.com/vitualizz/ai-software-delivery-team/internal/config"
	"github.com/vitualizz/ai-software-delivery-team/internal/pipeline"
)

func newMachine(t *testing.T) *pipeline.FSMachine {
	t.Helper()
	store := artifact.NewFSStore(t.TempDir())
	return pipeline.NewFSMachine(store)
}

// zeroRoot returns a zero-value config.Root. FSMachine.AdvanceStep ignores it.
func zeroRoot() config.Root {
	return config.Root{}
}

// --- Pipeline state v2 / AdvanceStep tests ---

// TestAdvanceStep_CreatesV2StateOnFirstCall verifies that AdvanceStep creates
// a v2 pipeline-state file when one does not yet exist.
func TestAdvanceStep_CreatesV2StateOnFirstCall(t *testing.T) {
	ctx := context.Background()
	m := newMachine(t)
	change := "v2-init"

	err := m.AdvanceStep(ctx, zeroRoot(), change, "developer", "artifact-loading")
	if err != nil {
		t.Fatalf("AdvanceStep: %v", err)
	}
}

// TestAdvanceStep_TwoSpecialistsAdvanceIndependently verifies that two specialists
// record progress independently without interfering with each other.
func TestAdvanceStep_TwoSpecialistsAdvanceIndependently(t *testing.T) {
	ctx := context.Background()
	m := newMachine(t)
	change := "v2-independent"

	if err := m.AdvanceStep(ctx, zeroRoot(), change, "developer", "artifact-loading"); err != nil {
		t.Fatalf("developer step 1: %v", err)
	}
	if err := m.AdvanceStep(ctx, zeroRoot(), change, "security", "threat-modeling"); err != nil {
		t.Fatalf("security step 1: %v", err)
	}
	if err := m.AdvanceStep(ctx, zeroRoot(), change, "developer", "platform-context"); err != nil {
		t.Fatalf("developer step 2: %v", err)
	}
}

// TestAdvanceStep_StepsCompletedGrows verifies that StepsCompleted appends rather
// than overwrites — calling AdvanceStep twice grows the slice to length 2.
func TestAdvanceStep_StepsCompletedGrows(t *testing.T) {
	ctx := context.Background()
	store := artifact.NewFSStore(t.TempDir())
	m := pipeline.NewFSMachine(store)
	change := "v2-grows"

	if err := m.AdvanceStep(ctx, zeroRoot(), change, "developer", "artifact-loading"); err != nil {
		t.Fatalf("step 1: %v", err)
	}
	if err := m.AdvanceStep(ctx, zeroRoot(), change, "developer", "platform-context"); err != nil {
		t.Fatalf("step 2: %v", err)
	}

	// Read the raw v2 state and verify two steps are recorded.
	var sv2 pipeline.StateV2
	if err := store.Read(ctx, change, pipeline.ArtifactTypeV2, &sv2); err != nil {
		t.Fatalf("read v2 state: %v", err)
	}
	devState := sv2.Specialists["developer"]
	if len(devState.StepsCompleted) != 2 {
		t.Errorf("expected 2 steps completed, got %d", len(devState.StepsCompleted))
	}
	if devState.CurrentStep != "platform-context" {
		t.Errorf("CurrentStep: got %q, want %q", devState.CurrentStep, "platform-context")
	}
}

// TestCurrentOnMissingFile verifies that Current returns a zero State (no error)
// when the pipeline-state file has not been created yet.
func TestCurrentOnMissingFile(t *testing.T) {
	ctx := context.Background()
	m := newMachine(t)

	s, err := m.Current(ctx, "nonexistent-change")
	if err != nil {
		t.Fatalf("Current on missing file: %v", err)
	}
	if s.ChangeID != "" {
		t.Errorf("expected empty ChangeID for missing file, got %q", s.ChangeID)
	}
}
