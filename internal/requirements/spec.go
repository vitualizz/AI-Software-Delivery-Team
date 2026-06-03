package requirements

// RequirementsSpec is the typed payload for a requirements-spec artifact.
// It matches the requirements-spec.schema.yaml contract exactly.
type RequirementsSpec struct {
	UserStories        []UserStory         `yaml:"user_stories"`
	AcceptanceCriteria map[string][]string `yaml:"acceptance_criteria"`
	Scope              Scope               `yaml:"scope"`
	NFRs               []string            `yaml:"nfrs"`
	OpenQuestions      []string            `yaml:"open_questions"`
}

// UserStory represents a single user story within a RequirementsSpec.
type UserStory struct {
	ID     string `yaml:"id"`
	As     string `yaml:"as"`
	Want   string `yaml:"want"`
	SoThat string `yaml:"so_that"`
}

// Scope defines the explicit in/out boundaries for a change.
type Scope struct {
	In  []string `yaml:"in"`
	Out []string `yaml:"out"`
}
