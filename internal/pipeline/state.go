package pipeline

import "time"

// State is the full pipeline-state.yaml document.
// It matches the pipeline-state.schema.yaml contract exactly.
type State struct {
	SchemaVersion string       `yaml:"schema_version"`
	ChangeID      string       `yaml:"change_id"`
	CurrentState  Phase        `yaml:"current_state"`
	Transitions   []Transition `yaml:"transitions"`
}

// Transition records a single state change event in the history.
type Transition struct {
	From      Phase     `yaml:"from"`
	To        Phase     `yaml:"to"`
	Timestamp time.Time `yaml:"timestamp"`
}
