// Package pipeline implements the sequential FSM for ASDT change delivery.
// Valid transitions are enforced by the specialist model (StateV2 / AdvanceStep).
// The Phase type and constants are kept for backward compatibility with v1 state files.
package pipeline

// Phase represents a pipeline stage as a string enum.
// Kept for backward compatibility with v1 pipeline-state.yaml files.
// New code should use the specialist model (StateV2 / AdvanceStep).
type Phase string

const (
	// PhaseRequirements is the initial phase after /asdt requirements runs.
	PhaseRequirements Phase = "requirements"

	// PhasePlan is set after /asdt develop produces an implementation plan.
	PhasePlan Phase = "plan"

	// PhaseImplement is set after the implementation phase completes.
	PhaseImplement Phase = "implement"

	// PhaseReview is the terminal phase before merging.
	PhaseReview Phase = "review"
)
