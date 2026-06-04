package pipeline

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/vitualizz/ai-software-delivery-team/internal/artifact"
	"github.com/vitualizz/ai-software-delivery-team/internal/config"
)

const (
	artifactType  = "pipeline-state"
	schemaVersion = "1"
)

// PipelineRunner is the port for reading and advancing pipeline state.
type PipelineRunner interface {
	// Current returns the current pipeline State for the given change.
	// If no state file exists, it returns a zero State without error.
	Current(ctx context.Context, change string) (State, error)

	// Advance moves the pipeline from its current state to `to`.
	// Returns an error if the transition is illegal.
	// Deprecated: use AdvanceStep for new specialist-model code.
	Advance(ctx context.Context, change string, to Phase) (State, error)

	// CanTransition returns true when transitioning from → to is a legal edge.
	CanTransition(from, to Phase) bool

	// AdvanceStep records completion of a workflow step for one specialist.
	// It appends to steps_completed (never overwrites), sets current_step,
	// and creates the specialist state if absent. Creates the state file on
	// first use with schema_version "2".
	// root is accepted for interface uniformity but the FSMachine ignores it
	// because the store is already bound to the project root.
	AdvanceStep(ctx context.Context, root config.Root, change, specialistID, stepID string) error
}

// FSMachine is the filesystem-backed implementation of PipelineRunner.
// It reads and writes pipeline-state.yaml via the artifact.Store port.
type FSMachine struct {
	store artifact.Store
}

// NewFSMachine constructs an FSMachine backed by the provided store.
func NewFSMachine(store artifact.Store) *FSMachine {
	return &FSMachine{store: store}
}

// CanTransition returns true when from → to is a legal pipeline edge.
func (m *FSMachine) CanTransition(from, to Phase) bool {
	expected, ok := validEdges[from]
	return ok && expected == to
}

// Current reads the pipeline-state.yaml for the given change.
// If the file does not exist, it returns a zero-value State.
func (m *FSMachine) Current(ctx context.Context, change string) (State, error) {
	if !m.store.Exists(change, artifactType) {
		return State{}, nil
	}
	var s State
	if err := m.store.Read(ctx, change, artifactType, &s); err != nil {
		return State{}, fmt.Errorf("pipeline current %s: %w", change, err)
	}
	return s, nil
}

// Advance reads the current state, validates the transition, appends it to
// the history, sets the new current state, and writes the file back via Store.
// It creates the pipeline-state.yaml when it does not yet exist (initial state).
func (m *FSMachine) Advance(ctx context.Context, change string, to Phase) (State, error) {
	current, err := m.Current(ctx, change)
	if err != nil {
		return State{}, err
	}

	// Handle initial state creation.
	if current.ChangeID == "" {
		// No file yet — bootstrapping. Only PhaseRequirements is valid as first state.
		if to != PhaseRequirements {
			return State{}, fmt.Errorf(
				"pipeline advance: cannot start at %q; initial state must be %q",
				to, PhaseRequirements,
			)
		}
		s := State{
			SchemaVersion: schemaVersion,
			ChangeID:      change,
			CurrentState:  PhaseRequirements,
			Transitions:   []Transition{},
		}
		if err := m.store.Write(ctx, change, artifactType, s); err != nil {
			return State{}, fmt.Errorf("pipeline advance write initial: %w", err)
		}
		return s, nil
	}

	from := current.CurrentState
	if !m.CanTransition(from, to) {
		expected, _ := validEdges[from]
		return State{}, fmt.Errorf(
			"%w: cannot advance from %q to %q (expected next: %q)",
			ErrIllegalTransition, from, to, expected,
		)
	}

	current.CurrentState = to
	current.Transitions = append(current.Transitions, Transition{
		From:      from,
		To:        to,
		Timestamp: time.Now().UTC(),
	})

	if err := m.store.Write(ctx, change, artifactType, current); err != nil {
		return State{}, fmt.Errorf("pipeline advance write: %w", err)
	}
	return current, nil
}

// ErrIllegalTransition is returned when an advance attempt violates the FSM edges.
var ErrIllegalTransition = errors.New("illegal pipeline transition")

// ArtifactTypeV2 is the artifact type key for the specialist-scoped pipeline state.
const ArtifactTypeV2 = "pipeline-state-v2"

// AdvanceStep records a specialist step completion in PipelineStateV2.
// It reads the current v2 state (or creates it), appends the step, and writes back.
func (m *FSMachine) AdvanceStep(_ context.Context, _ config.Root, change, specialistID, stepID string) error {
	ctx := context.Background()

	var sv2 PipelineStateV2
	if m.store.Exists(change, ArtifactTypeV2) {
		if err := m.store.Read(ctx, change, ArtifactTypeV2, &sv2); err != nil {
			return fmt.Errorf("pipeline advance-step read %s: %w", change, err)
		}
	}

	// Bootstrap on first use.
	if sv2.SchemaVersion == "" {
		sv2 = PipelineStateV2{
			SchemaVersion: SchemaVersionV2,
			ChangeID:      change,
			Specialists:   make(map[string]SpecialistState),
		}
	}
	if sv2.Specialists == nil {
		sv2.Specialists = make(map[string]SpecialistState)
	}

	sp := sv2.Specialists[specialistID]
	sp.StepsCompleted = append(sp.StepsCompleted, StepRecord{
		ID:        stepID,
		Timestamp: time.Now().UTC(),
	})
	sp.CurrentStep = stepID
	sv2.Specialists[specialistID] = sp

	if err := m.store.Write(ctx, change, ArtifactTypeV2, sv2); err != nil {
		return fmt.Errorf("pipeline advance-step write %s: %w", change, err)
	}
	return nil
}
