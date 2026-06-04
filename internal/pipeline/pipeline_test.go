package pipeline_test

import (
	"context"
	"errors"
	"strings"
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

// TestFSMachine_InitialFields verifies that the first Advance creates a
// pipeline-state with the correct schema_version, change_id, and empty transitions.
func TestFSMachine_InitialFields(t *testing.T) {
	ctx := context.Background()
	m := newMachine(t)
	change := "init-fields"

	s, err := m.Advance(ctx, change, pipeline.PhaseRequirements)
	if err != nil {
		t.Fatalf("Advance: %v", err)
	}

	if s.SchemaVersion != "1" {
		t.Errorf("SchemaVersion: got %q, want %q", s.SchemaVersion, "1")
	}
	if s.ChangeID != change {
		t.Errorf("ChangeID: got %q, want %q", s.ChangeID, change)
	}
	if s.CurrentState != pipeline.PhaseRequirements {
		t.Errorf("CurrentState: got %q, want %q", s.CurrentState, pipeline.PhaseRequirements)
	}
	if len(s.Transitions) != 0 {
		t.Errorf("expected 0 transitions on initial state, got %d", len(s.Transitions))
	}
}

// TestFSMachine_AdvanceAppendsTransitions verifies that each valid Advance
// appends a new entry to transitions[] rather than replacing the list.
func TestFSMachine_AdvanceAppendsTransitions(t *testing.T) {
	ctx := context.Background()
	m := newMachine(t)
	change := "append-test"

	if _, err := m.Advance(ctx, change, pipeline.PhaseRequirements); err != nil {
		t.Fatalf("init: %v", err)
	}

	s1, err := m.Advance(ctx, change, pipeline.PhasePlan)
	if err != nil {
		t.Fatalf("requirements→plan: %v", err)
	}
	if len(s1.Transitions) != 1 {
		t.Errorf("after first advance: expected 1 transition, got %d", len(s1.Transitions))
	}
	if s1.Transitions[0].From != pipeline.PhaseRequirements {
		t.Errorf("Transitions[0].From: got %q, want %q", s1.Transitions[0].From, pipeline.PhaseRequirements)
	}
	if s1.Transitions[0].To != pipeline.PhasePlan {
		t.Errorf("Transitions[0].To: got %q, want %q", s1.Transitions[0].To, pipeline.PhasePlan)
	}

	s2, err := m.Advance(ctx, change, pipeline.PhaseImplement)
	if err != nil {
		t.Fatalf("plan→implement: %v", err)
	}
	if len(s2.Transitions) != 2 {
		t.Errorf("after second advance: expected 2 transitions, got %d", len(s2.Transitions))
	}
}

// TestIllegalTransition_ErrorContainsBothPhases verifies that the error from an
// illegal skip names both the current state and the attempted target state.
func TestIllegalTransition_ErrorContainsBothPhases(t *testing.T) {
	tests := []struct {
		name    string
		from    pipeline.Phase
		to      pipeline.Phase
		wantFrom string
		wantTo   string
	}{
		{
			name:    "requirements skip to implement",
			from:    pipeline.PhaseRequirements,
			to:      pipeline.PhaseImplement,
			wantFrom: "requirements",
			wantTo:   "implement",
		},
		{
			name:    "plan skip to review",
			from:    pipeline.PhasePlan,
			to:      pipeline.PhaseReview,
			wantFrom: "plan",
			wantTo:   "review",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			m := newMachine(t)
			change := "phase-names-" + tc.name

			// Advance to the starting state.
			if _, err := m.Advance(ctx, change, pipeline.PhaseRequirements); err != nil {
				t.Fatalf("init: %v", err)
			}
			if tc.from == pipeline.PhasePlan {
				if _, err := m.Advance(ctx, change, pipeline.PhasePlan); err != nil {
					t.Fatalf("advance to plan: %v", err)
				}
			}

			_, err := m.Advance(ctx, change, tc.to)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !errors.Is(err, pipeline.ErrIllegalTransition) {
				t.Errorf("expected ErrIllegalTransition, got: %v", err)
			}
			msg := err.Error()
			if !containsPhase(msg, tc.wantFrom) {
				t.Errorf("error %q should mention current phase %q", msg, tc.wantFrom)
			}
			if !containsPhase(msg, tc.wantTo) {
				t.Errorf("error %q should mention target phase %q", msg, tc.wantTo)
			}
		})
	}
}

// containsPhase checks whether a phase name appears in the error message.
func containsPhase(msg, phase string) bool {
	return strings.Contains(msg, phase)
}

// TestCanTransition_ReverseTransitions verifies that all reverse (backward)
// transitions return false from CanTransition.
func TestCanTransition_ReverseTransitions(t *testing.T) {
	m := newMachine(t)

	reverseEdges := [][2]pipeline.Phase{
		{pipeline.PhasePlan, pipeline.PhaseRequirements},
		{pipeline.PhaseImplement, pipeline.PhasePlan},
		{pipeline.PhaseImplement, pipeline.PhaseRequirements},
		{pipeline.PhaseReview, pipeline.PhaseImplement},
		{pipeline.PhaseReview, pipeline.PhasePlan},
		{pipeline.PhaseReview, pipeline.PhaseRequirements},
	}
	for _, e := range reverseEdges {
		if m.CanTransition(e[0], e[1]) {
			t.Errorf("CanTransition(%q→%q) should be false (reverse transition)", e[0], e[1])
		}
	}
}
