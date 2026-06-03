// Package pipeline implements the sequential FSM for ASDT change delivery.
// Valid transitions: requirements → plan → implement → review.
// Any other transition is rejected with a descriptive error.
package pipeline

// Phase represents a pipeline stage as a string enum.
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

// validEdges defines the only legal phase transitions.
// requirements → plan → implement → review.
var validEdges = map[Phase]Phase{
	PhaseRequirements: PhasePlan,
	PhasePlan:         PhaseImplement,
	PhaseImplement:    PhaseReview,
}
