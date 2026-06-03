// Package artifact defines the ArtifactEnvelope contract, the Store port,
// and the FSStore adapter for reading/writing artifacts under .asdt/.
// This package imports only stdlib and gopkg.in/yaml.v3.
package artifact

import "time"

// CurrentSchemaVersion is the schema version this build understands.
// Consumers reject envelopes with a different version.
const CurrentSchemaVersion = "1"

// EnvelopeHeader is the uniform metadata written at the root of every
// artifact YAML file. It is shared across all agent types.
type EnvelopeHeader struct {
	SchemaVersion string    `yaml:"schema_version"`
	Agent         string    `yaml:"agent"`
	ChangeID      string    `yaml:"change_id"`
	CreatedAt     time.Time `yaml:"created_at"`
	PromptVersion string    `yaml:"prompt_version"`
	InputRefs     []string  `yaml:"input_refs"`
}

// Envelope wraps any typed payload with the uniform header.
// YAML inline flattens the header fields to the root level so the file
// reads naturally without a nested "header" key.
type Envelope[T any] struct {
	EnvelopeHeader `yaml:",inline"`
	Payload        T `yaml:"payload"`
}
