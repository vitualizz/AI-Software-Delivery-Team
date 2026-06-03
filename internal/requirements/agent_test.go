package requirements_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/vitualizz/ai-software-delivery-team/internal/artifact"
	"github.com/vitualizz/ai-software-delivery-team/internal/config"
	"github.com/vitualizz/ai-software-delivery-team/internal/knowledge"
	"github.com/vitualizz/ai-software-delivery-team/internal/llm"
	"github.com/vitualizz/ai-software-delivery-team/internal/pipeline"
	"github.com/vitualizz/ai-software-delivery-team/internal/prompt"
	"github.com/vitualizz/ai-software-delivery-team/internal/requirements"
	"github.com/vitualizz/ai-software-delivery-team/skill"
	"gopkg.in/yaml.v3"
)

// validSpecYAML is a well-formed RequirementsSpec in YAML that the MockProvider returns.
const validSpecYAML = `user_stories:
  - id: US-001
    as: developer
    want: reset my password via email
    so_that: I can regain access when locked out
acceptance_criteria:
  US-001:
    - User receives a reset link within 60 seconds
    - Link expires after 15 minutes
scope:
  in:
    - Password reset email flow
    - Token generation and validation
  out:
    - SMS reset
    - Admin-initiated reset
nfrs:
  - Reset link must expire after 15 minutes
open_questions:
  - Should we support multiple emails per account?
`

// testRegistry builds a minimal test SkillRegistry backed by the embedded FS.
func testRegistry(t *testing.T) prompt.SkillRegistry {
	t.Helper()
	return prompt.NewEmbeddedRegistry(skill.PromptSubFS())
}

// makeStore creates an FSStore backed by a temporary .asdt/ directory.
// Returns the store and the config.Root for use in the test.
func makeStore(t *testing.T) (artifact.Store, config.Root) {
	t.Helper()
	dir := t.TempDir()
	asdt := filepath.Join(dir, ".asdt")
	if err := os.MkdirAll(asdt, 0o755); err != nil {
		t.Fatalf("mkdir .asdt: %v", err)
	}
	root, err := config.Discover(dir)
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}
	return artifact.NewFSStore(asdt), root
}

// noopReader is a knowledge.Reader that always returns an error (simulates missing platform.yaml).
type noopReader struct{}

func (noopReader) Read(_ config.Root) (knowledge.Platform, error) {
	return knowledge.Platform{}, errors.New("platform.yaml not found")
}
func (noopReader) Write(_ config.Root, _ knowledge.Platform) error { return nil }

// writingReader writes a platform.yaml to the root before tests need it.
type writingReader struct {
	platform knowledge.Platform
}

func (r writingReader) Read(_ config.Root) (knowledge.Platform, error) {
	return r.platform, nil
}
func (r writingReader) Write(_ config.Root, _ knowledge.Platform) error { return nil }

func TestAgent_HappyPath(t *testing.T) {
	store, root := makeStore(t)
	reg := testRegistry(t)
	mock := llm.NewMockProvider(llm.WithRawResponses(validSpecYAML))
	runner := pipeline.NewFSMachine(store)
	reader := writingReader{platform: knowledge.Platform{DetectedStack: []string{"go"}}}

	agent := requirements.New(reg, mock, store, runner, reader)

	const change = "test-change"
	const idea = "add password reset feature for users who forgot their credentials"

	if err := agent.Run(context.Background(), root, change, idea); err != nil {
		t.Fatalf("Run: %v", err)
	}

	// Verify artifact was written.
	var env artifact.Envelope[requirements.RequirementsSpec]
	if err := store.Read(context.Background(), change, "requirements-spec", &env); err != nil {
		t.Fatalf("Read artifact: %v", err)
	}

	// Check envelope header fields.
	if env.SchemaVersion == "" {
		t.Error("SchemaVersion must not be empty")
	}
	if env.Agent != "requirements" {
		t.Errorf("Agent: got %q, want %q", env.Agent, "requirements")
	}
	if env.ChangeID != change {
		t.Errorf("ChangeID: got %q, want %q", env.ChangeID, change)
	}
	if env.CreatedAt.IsZero() {
		t.Error("CreatedAt must be set")
	}
	if env.PromptVersion == "" {
		t.Error("PromptVersion (prompt_version) must not be empty")
	}

	// Check payload.
	if len(env.Payload.UserStories) == 0 {
		t.Error("expected at least one user story in payload")
	}

	// Verify pipeline advanced.
	state, err := runner.Current(context.Background(), change)
	if err != nil {
		t.Fatalf("Current: %v", err)
	}
	if state.CurrentState != pipeline.PhaseRequirements {
		t.Errorf("pipeline state: got %q, want %q", state.CurrentState, pipeline.PhaseRequirements)
	}

	// Verify MockProvider was called once.
	if len(mock.Calls) != 1 {
		t.Errorf("expected 1 LLM call, got %d", len(mock.Calls))
	}
}

func TestAgent_ErrMissingIdea(t *testing.T) {
	store, root := makeStore(t)
	reg := testRegistry(t)
	mock := llm.NewMockProvider()
	runner := pipeline.NewFSMachine(store)

	agent := requirements.New(reg, mock, store, runner, noopReader{})

	err := agent.Run(context.Background(), root, "change", "")
	if !errors.Is(err, requirements.ErrMissingIdea) {
		t.Errorf("expected ErrMissingIdea, got: %v", err)
	}

	// LLM must not be called.
	if len(mock.Calls) != 0 {
		t.Errorf("expected 0 LLM calls, got %d", len(mock.Calls))
	}
}

func TestAgent_ErrAmbiguousIdea(t *testing.T) {
	store, root := makeStore(t)
	reg := testRegistry(t)
	mock := llm.NewMockProvider()
	runner := pipeline.NewFSMachine(store)

	agent := requirements.New(reg, mock, store, runner, noopReader{})

	// Fewer than 5 words.
	err := agent.Run(context.Background(), root, "change", "add login")
	var ambig requirements.ErrAmbiguousIdea
	if !errors.As(err, &ambig) {
		t.Errorf("expected ErrAmbiguousIdea, got: %v", err)
	}
	if ambig.Question == "" {
		t.Error("ErrAmbiguousIdea.Question must not be empty")
	}
}

func TestAgent_MissingPlatform_ProceedsWithWarning(t *testing.T) {
	store, root := makeStore(t)
	reg := testRegistry(t)
	mock := llm.NewMockProvider(llm.WithRawResponses(validSpecYAML))
	runner := pipeline.NewFSMachine(store)

	// noopReader simulates missing platform.yaml — agent must not error.
	agent := requirements.New(reg, mock, store, runner, noopReader{})

	const idea = "add password reset feature for users who forgot their credentials"
	err := agent.Run(context.Background(), root, "test-change", idea)
	if err != nil {
		t.Errorf("Run with missing platform.yaml should not error; got: %v", err)
	}
}

func TestAgent_PromptVersionSet(t *testing.T) {
	store, root := makeStore(t)
	reg := testRegistry(t)
	mock := llm.NewMockProvider(llm.WithRawResponses(validSpecYAML))
	runner := pipeline.NewFSMachine(store)
	reader := noopReader{}

	agent := requirements.New(reg, mock, store, runner, reader)

	const idea = "add password reset feature for users who need it badly"
	if err := agent.Run(context.Background(), root, "my-change", idea); err != nil {
		t.Fatalf("Run: %v", err)
	}

	// Read the written YAML directly to verify prompt_version is present.
	asdt := root.Path()
	artifactPath := filepath.Join(asdt, "artifacts", "my-change", "requirements-spec.yaml")
	data, err := os.ReadFile(artifactPath)
	if err != nil {
		t.Fatalf("ReadFile artifact: %v", err)
	}

	var raw map[string]any
	if err := yaml.Unmarshal(data, &raw); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	pv, ok := raw["prompt_version"]
	if !ok {
		t.Error("prompt_version field must be present in written YAML")
	}
	if pv == "" || pv == nil {
		t.Error("prompt_version must not be empty")
	}
}
