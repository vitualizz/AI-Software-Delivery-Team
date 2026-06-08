package pipeline

import "time"

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

// State is the v1 pipeline-state.yaml document (linear FSM).
// It matches the pipeline-state.schema.yaml contract exactly.
// Kept for backward compatibility — new code should use StateV2.
type State struct {
	SchemaVersion string       `yaml:"schema_version"`
	ChangeID      string       `yaml:"change_id"`
	CurrentState  Phase        `yaml:"current_state"`
	Transitions   []Transition `yaml:"transitions"`
}

// Transition records a single v1 state change event in the history.
type Transition struct {
	From      Phase     `yaml:"from"`
	To        Phase     `yaml:"to"`
	Timestamp time.Time `yaml:"timestamp"`
}

// --- v2 Specialist-scoped pipeline state ---

// SchemaVersionV2 is the schema_version value for v2 pipeline state documents.
const SchemaVersionV2 = "2"

// StateV2 is the per-specialist pipeline state document (schema_version "2").
// Each specialist has independent state; there is no global FSM.
type StateV2 struct {
	SchemaVersion string                     `yaml:"schema_version"`
	ChangeID      string                     `yaml:"change_id"`
	Specialists   map[string]SpecialistState `yaml:"specialists"`
}

// SpecialistState tracks the progress of a single specialist for one change.
type SpecialistState struct {
	CurrentStep      string       `yaml:"current_step"`
	StepsCompleted   []StepRecord `yaml:"steps_completed"`
	ArtifactsWritten []string     `yaml:"artifacts_written"`
}

// StepRecord is an entry in the append-only steps_completed log.
type StepRecord struct {
	ID        string    `yaml:"id"`
	Timestamp time.Time `yaml:"timestamp"`
}
