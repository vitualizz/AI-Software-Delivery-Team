package pipeline_test

import (
	"context"
	"errors"
	"testing"

	"github.com/vitualizz/ai-software-delivery-team/internal/artifact"
	"github.com/vitualizz/ai-software-delivery-team/internal/pipeline"
)

func newMachine(t *testing.T) *pipeline.FSMachine {
	t.Helper()
	store := artifact.NewFSStore(t.TempDir())
	return pipeline.NewFSMachine(store)
}

// TestFullValidPath walks the entire requirements → plan → implement → review path.
func TestFullValidPath(t *testing.T) {
	ctx := context.Background()
	m := newMachine(t)
	change := "test-change"

	// Create initial state.
	s, err := m.Advance(ctx, change, pipeline.PhaseRequirements)
	if err != nil {
		t.Fatalf("initial requirements: %v", err)
	}
	if s.CurrentState != pipeline.PhaseRequirements {
		t.Errorf("state: got %q, want %q", s.CurrentState, pipeline.PhaseRequirements)
	}
	if len(s.Transitions) != 0 {
		t.Errorf("expected 0 transitions on init, got %d", len(s.Transitions))
	}

	// requirements → plan
	s, err = m.Advance(ctx, change, pipeline.PhasePlan)
	if err != nil {
		t.Fatalf("requirements→plan: %v", err)
	}
	if s.CurrentState != pipeline.PhasePlan {
		t.Errorf("state: got %q, want %q", s.CurrentState, pipeline.PhasePlan)
	}
	if len(s.Transitions) != 1 {
		t.Errorf("expected 1 transition, got %d", len(s.Transitions))
	}

	// plan → implement
	s, err = m.Advance(ctx, change, pipeline.PhaseImplement)
	if err != nil {
		t.Fatalf("plan→implement: %v", err)
	}

	// implement → review
	s, err = m.Advance(ctx, change, pipeline.PhaseReview)
	if err != nil {
		t.Fatalf("implement→review: %v", err)
	}
	if s.CurrentState != pipeline.PhaseReview {
		t.Errorf("state: got %q, want %q", s.CurrentState, pipeline.PhaseReview)
	}
	if len(s.Transitions) != 3 {
		t.Errorf("expected 3 transitions, got %d", len(s.Transitions))
	}
}

// TestIllegalSkip_RequirementsToImplement verifies that skipping plan is rejected.
func TestIllegalSkip_RequirementsToImplement(t *testing.T) {
	ctx := context.Background()
	m := newMachine(t)
	change := "skip-test"

	if _, err := m.Advance(ctx, change, pipeline.PhaseRequirements); err != nil {
		t.Fatalf("setup: %v", err)
	}

	_, err := m.Advance(ctx, change, pipeline.PhaseImplement)
	if !errors.Is(err, pipeline.ErrIllegalTransition) {
		t.Errorf("expected ErrIllegalTransition, got: %v", err)
	}
}

// TestIllegalSkip_RequirementsToReview verifies that skipping plan+implement is rejected.
func TestIllegalSkip_RequirementsToReview(t *testing.T) {
	ctx := context.Background()
	m := newMachine(t)
	change := "skip-test2"

	if _, err := m.Advance(ctx, change, pipeline.PhaseRequirements); err != nil {
		t.Fatalf("setup: %v", err)
	}

	_, err := m.Advance(ctx, change, pipeline.PhaseReview)
	if !errors.Is(err, pipeline.ErrIllegalTransition) {
		t.Errorf("expected ErrIllegalTransition, got: %v", err)
	}
}

// TestIllegalSkip_PlanToReview verifies that skipping implement is rejected.
func TestIllegalSkip_PlanToReview(t *testing.T) {
	ctx := context.Background()
	m := newMachine(t)
	change := "skip-test3"

	if _, err := m.Advance(ctx, change, pipeline.PhaseRequirements); err != nil {
		t.Fatalf("setup requirements: %v", err)
	}
	if _, err := m.Advance(ctx, change, pipeline.PhasePlan); err != nil {
		t.Fatalf("setup plan: %v", err)
	}

	_, err := m.Advance(ctx, change, pipeline.PhaseReview)
	if !errors.Is(err, pipeline.ErrIllegalTransition) {
		t.Errorf("expected ErrIllegalTransition, got: %v", err)
	}
}

// TestInitialStateCreation verifies that the pipeline-state.yaml is created
// when it does not yet exist and the first advance targets PhaseRequirements.
func TestInitialStateCreation(t *testing.T) {
	ctx := context.Background()
	m := newMachine(t)
	change := "fresh-change"

	// Current on non-existent file returns zero State without error.
	s, err := m.Current(ctx, change)
	if err != nil {
		t.Fatalf("Current on missing file: %v", err)
	}
	if s.ChangeID != "" {
		t.Errorf("expected empty ChangeID for missing file, got %q", s.ChangeID)
	}

	// Advance to requirements creates the file.
	s, err = m.Advance(ctx, change, pipeline.PhaseRequirements)
	if err != nil {
		t.Fatalf("Advance to requirements: %v", err)
	}
	if s.ChangeID != change {
		t.Errorf("ChangeID: got %q, want %q", s.ChangeID, change)
	}
	if s.SchemaVersion != "1" {
		t.Errorf("SchemaVersion: got %q, want %q", s.SchemaVersion, "1")
	}

	// Current now returns the persisted state.
	s2, err := m.Current(ctx, change)
	if err != nil {
		t.Fatalf("Current after init: %v", err)
	}
	if s2.CurrentState != pipeline.PhaseRequirements {
		t.Errorf("CurrentState: got %q, want %q", s2.CurrentState, pipeline.PhaseRequirements)
	}
}

// TestCanTransition covers all legal and illegal edge combinations.
func TestCanTransition(t *testing.T) {
	m := newMachine(t)

	legalEdges := [][2]pipeline.Phase{
		{pipeline.PhaseRequirements, pipeline.PhasePlan},
		{pipeline.PhasePlan, pipeline.PhaseImplement},
		{pipeline.PhaseImplement, pipeline.PhaseReview},
	}
	for _, e := range legalEdges {
		if !m.CanTransition(e[0], e[1]) {
			t.Errorf("CanTransition(%q, %q) should be true", e[0], e[1])
		}
	}

	illegalEdges := [][2]pipeline.Phase{
		{pipeline.PhaseRequirements, pipeline.PhaseImplement},
		{pipeline.PhaseRequirements, pipeline.PhaseReview},
		{pipeline.PhasePlan, pipeline.PhaseReview},
		{pipeline.PhasePlan, pipeline.PhaseRequirements},
		{pipeline.PhaseReview, pipeline.PhaseRequirements},
	}
	for _, e := range illegalEdges {
		if m.CanTransition(e[0], e[1]) {
			t.Errorf("CanTransition(%q, %q) should be false", e[0], e[1])
		}
	}
}
