// Package pipeline implements the sequential FSM for ASDT change delivery.
// Valid transitions are enforced by the specialist model (StateV2 / AdvanceStep).
// The Phase type and constants are kept for backward compatibility with v1 state files.
package pipeline

import (
	"context"
	"fmt"
	"time"

	"github.com/vitualizz/asdt/internal/artifact"
	"github.com/vitualizz/asdt/internal/config"
)

const (
	artifactType  = "pipeline-state"
	schemaVersion = "1"
)

// Runner is the port for reading and advancing pipeline state.
type Runner interface {
	// Current returns the current pipeline State for the given change.
	// If no state file exists, it returns a zero State without error.
	Current(ctx context.Context, change string) (State, error)

	// AdvanceStep records completion of a workflow step for one specialist.
	// It appends to steps_completed (never overwrites), sets current_step,
	// and creates the specialist state if absent. Creates the state file on
	// first use with schema_version "2".
	// root is accepted for interface uniformity but the FSMachine ignores it
	// because the store is already bound to the project root.
	AdvanceStep(ctx context.Context, root config.Root, change, specialistID, stepID string) error
}

// FSMachine is the filesystem-backed implementation of Runner.
// It reads and writes pipeline-state.yaml via the artifact.Store port.
type FSMachine struct {
	store artifact.Store
}

// NewFSMachine constructs an FSMachine backed by the provided store.
func NewFSMachine(store artifact.Store) *FSMachine {
	return &FSMachine{store: store}
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

// ArtifactTypeV2 is the artifact type key for the specialist-scoped pipeline state.
const ArtifactTypeV2 = "pipeline-state-v2"

// AdvanceStep records a specialist step completion in StateV2.
// It reads the current v2 state (or creates it), appends the step, and writes back.
func (m *FSMachine) AdvanceStep(_ context.Context, _ config.Root, change, specialistID, stepID string) error {
	ctx := context.Background()

	var sv2 StateV2
	if m.store.Exists(change, ArtifactTypeV2) {
		if err := m.store.Read(ctx, change, ArtifactTypeV2, &sv2); err != nil {
			return fmt.Errorf("pipeline advance-step read %s: %w", change, err)
		}
	}

	// Bootstrap on first use.
	if sv2.SchemaVersion == "" {
		sv2 = StateV2{
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
