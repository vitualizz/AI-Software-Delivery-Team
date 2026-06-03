package developer

// ImplementationPlan is the typed payload for an implementation-plan artifact.
// It matches the implementation-plan.schema.yaml contract exactly.
type ImplementationPlan struct {
	Approach           string     `yaml:"approach,omitempty"`
	Steps              []PlanStep `yaml:"steps"`
	ComplexityEstimate string     `yaml:"complexity_estimate"`
	// OpenItems collects traceability warnings: story IDs from the spec that
	// are not referenced in any step. Populated during validation.
	OpenItems []string `yaml:"open_items,omitempty"`
}

// PlanStep is a single implementation step referencing one user story.
type PlanStep struct {
	StoryID        string        `yaml:"story_id"`
	FilesToCreate  []string      `yaml:"files_to_create"`
	FilesToModify  []string      `yaml:"files_to_modify"`
	Rationale      string        `yaml:"rationale"`
	CodeSnippets   []CodeSnippet `yaml:"code_snippets"`
}

// CodeSnippet is an inline code snippet attached to a PlanStep.
type CodeSnippet struct {
	File    string `yaml:"file"`
	Content string `yaml:"content"`
}
