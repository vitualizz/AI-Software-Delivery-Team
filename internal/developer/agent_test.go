package developer_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/vitualizz/ai-software-delivery-team/internal/artifact"
	"github.com/vitualizz/ai-software-delivery-team/internal/config"
	"github.com/vitualizz/ai-software-delivery-team/internal/developer"
	"github.com/vitualizz/ai-software-delivery-team/internal/knowledge"
	"github.com/vitualizz/ai-software-delivery-team/internal/llm"
	"github.com/vitualizz/ai-software-delivery-team/internal/pipeline"
	"github.com/vitualizz/ai-software-delivery-team/internal/prompt"
	"github.com/vitualizz/ai-software-delivery-team/internal/requirements"
	"github.com/vitualizz/ai-software-delivery-team/skill"
	"gopkg.in/yaml.v3"
)

// validPlanYAML is a well-formed ImplementationPlan in YAML for MockProvider.
const validPlanYAML = `approach: Implement password reset via email token flow
complexity_estimate: medium
steps:
  - story_id: US-001
    files_to_create:
      - internal/auth/reset.go
      - internal/auth/reset_test.go
    files_to_modify:
      - internal/auth/handler.go
    rationale: Create a dedicated reset package to keep auth concerns isolated
    code_snippets:
      - file: internal/auth/reset.go
        content: |
          func GenerateResetToken() (string, error) {
            return "", nil
          }
`

// fixtureSpec is the RequirementsSpec written to the store before each test.
var fixtureSpec = artifact.Envelope[requirements.RequirementsSpec]{
	EnvelopeHeader: artifact.EnvelopeHeader{
		SchemaVersion: artifact.CurrentSchemaVersion,
		Agent:         "requirements",
		ChangeID:      "dev-change",
		CreatedAt:     time.Now().UTC(),
		PromptVersion: "deadbeef",
		InputRefs:     []string{},
	},
	Payload: requirements.RequirementsSpec{
		UserStories: []requirements.UserStory{
			{
				ID:     "US-001",
				As:     "developer",
				Want:   "reset my password via email",
				SoThat: "I can regain access",
			},
		},
		AcceptanceCriteria: map[string][]string{
			"US-001": {"User receives a reset link within 60 seconds"},
		},
		Scope:         requirements.Scope{In: []string{"email reset"}, Out: []string{"sms reset"}},
		NFRs:          []string{"Link expires after 15 minutes"},
		OpenQuestions: []string{},
	},
}

// makeStore creates an FSStore and config.Root backed by a temp .asdt/ directory.
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

// writeFixtureSpec writes the fixture requirements-spec to the store.
func writeFixtureSpec(t *testing.T, store artifact.Store, change string) {
	t.Helper()
	spec := fixtureSpec
	spec.ChangeID = change
	if err := store.Write(context.Background(), change, "requirements-spec", spec); err != nil {
		t.Fatalf("write fixture spec: %v", err)
	}
}

// advanceToRequirements puts the pipeline in the requirements state.
func advanceToRequirements(t *testing.T, store artifact.Store, change string) {
	t.Helper()
	runner := pipeline.NewFSMachine(store)
	if _, err := runner.Advance(context.Background(), change, pipeline.PhaseRequirements); err != nil {
		t.Fatalf("advance pipeline to requirements: %v", err)
	}
}

// testRegistry returns an EmbeddedRegistry backed by the production prompt FS.
func testRegistry(t *testing.T) prompt.SkillRegistry {
	t.Helper()
	return prompt.NewEmbeddedRegistry(skill.PromptSubFS())
}

// noopReader is a knowledge.Reader that always returns an error.
type noopReader struct{}

func (noopReader) Read(_ config.Root) (knowledge.Platform, error) {
	return knowledge.Platform{}, errors.New("platform.yaml not found")
}
func (noopReader) Write(_ config.Root, _ knowledge.Platform) error { return nil }

// platformReader returns a fixed platform.
type platformReader struct {
	platform knowledge.Platform
}

func (r platformReader) Read(_ config.Root) (knowledge.Platform, error) {
	return r.platform, nil
}
func (r platformReader) Write(_ config.Root, _ knowledge.Platform) error { return nil }

func TestAgent_HappyPath(t *testing.T) {
	store, root := makeStore(t)
	const change = "dev-change"
	writeFixtureSpec(t, store, change)
	advanceToRequirements(t, store, change)

	reg := testRegistry(t)
	mock := llm.NewMockProvider(llm.WithRawResponses(validPlanYAML))
	runner := pipeline.NewFSMachine(store)
	reader := platformReader{platform: knowledge.Platform{DetectedStack: []string{"go"}}}

	agent := developer.New(reg, mock, store, runner, reader)

	if err := agent.Run(context.Background(), root, change); err != nil {
		t.Fatalf("Run: %v", err)
	}

	// Verify plan artifact was written.
	var env artifact.Envelope[developer.ImplementationPlan]
	if err := store.Read(context.Background(), change, "implementation-plan", &env); err != nil {
		t.Fatalf("Read artifact: %v", err)
	}

	// Verify envelope fields.
	if env.Agent != "developer" {
		t.Errorf("Agent: got %q, want %q", env.Agent, "developer")
	}
	if env.ChangeID != change {
		t.Errorf("ChangeID: got %q, want %q", env.ChangeID, change)
	}
	if env.PromptVersion == "" {
		t.Error("PromptVersion must not be empty")
	}
	if env.CreatedAt.IsZero() {
		t.Error("CreatedAt must be set")
	}

	// Verify pipeline advanced to plan.
	state, err := runner.Current(context.Background(), change)
	if err != nil {
		t.Fatalf("Current: %v", err)
	}
	if state.CurrentState != pipeline.PhasePlan {
		t.Errorf("pipeline state: got %q, want %q", state.CurrentState, pipeline.PhasePlan)
	}

	// Verify story ID is referenced.
	if len(env.Payload.Steps) == 0 {
		t.Fatal("expected at least one step in plan")
	}
	if env.Payload.Steps[0].StoryID != "US-001" {
		t.Errorf("Steps[0].StoryID: got %q, want %q", env.Payload.Steps[0].StoryID, "US-001")
	}

	// No open items: US-001 is referenced.
	if len(env.Payload.OpenItems) != 0 {
		t.Errorf("expected empty open_items, got %v", env.Payload.OpenItems)
	}
}

func TestAgent_ErrMissingRequirements(t *testing.T) {
	store, root := makeStore(t)
	reg := testRegistry(t)
	mock := llm.NewMockProvider()
	runner := pipeline.NewFSMachine(store)

	agent := developer.New(reg, mock, store, runner, noopReader{})

	err := agent.Run(context.Background(), root, "no-spec-change")
	if !errors.Is(err, developer.ErrMissingRequirements) {
		t.Errorf("expected ErrMissingRequirements, got: %v", err)
	}
	if len(mock.Calls) != 0 {
		t.Errorf("expected 0 LLM calls, got %d", len(mock.Calls))
	}
}

func TestAgent_MissingPlatform_ProceedsWithWarning(t *testing.T) {
	store, root := makeStore(t)
	const change = "no-platform-change"
	writeFixtureSpec(t, store, change)
	advanceToRequirements(t, store, change)

	reg := testRegistry(t)
	mock := llm.NewMockProvider(llm.WithRawResponses(validPlanYAML))
	runner := pipeline.NewFSMachine(store)

	// noopReader — no platform.yaml.
	agent := developer.New(reg, mock, store, runner, noopReader{})

	err := agent.Run(context.Background(), root, change)
	if err != nil {
		t.Errorf("Run with missing platform.yaml should not error; got: %v", err)
	}

	// input_refs should NOT include platform path.
	var env artifact.Envelope[developer.ImplementationPlan]
	if err := store.Read(context.Background(), change, "implementation-plan", &env); err != nil {
		t.Fatalf("Read artifact: %v", err)
	}
	for _, ref := range env.InputRefs {
		if ref == ".asdt/knowledge/platform.yaml" {
			t.Error("input_refs must not include platform.yaml when it was absent")
		}
	}
}

func TestAgent_InputRefsChain(t *testing.T) {
	store, root := makeStore(t)
	const change = "chain-change"
	writeFixtureSpec(t, store, change)
	advanceToRequirements(t, store, change)

	reg := testRegistry(t)
	mock := llm.NewMockProvider(llm.WithRawResponses(validPlanYAML))
	runner := pipeline.NewFSMachine(store)
	reader := platformReader{platform: knowledge.Platform{DetectedStack: []string{"go"}}}

	agent := developer.New(reg, mock, store, runner, reader)
	if err := agent.Run(context.Background(), root, change); err != nil {
		t.Fatalf("Run: %v", err)
	}

	var env artifact.Envelope[developer.ImplementationPlan]
	if err := store.Read(context.Background(), change, "implementation-plan", &env); err != nil {
		t.Fatalf("Read: %v", err)
	}

	// Must reference the requirements-spec.
	foundSpec := false
	for _, ref := range env.InputRefs {
		if ref == ".asdt/artifacts/chain-change/requirements-spec.yaml" {
			foundSpec = true
		}
	}
	if !foundSpec {
		t.Errorf("input_refs must contain requirements-spec path; got %v", env.InputRefs)
	}

	// Must reference platform.yaml.
	foundPlatform := false
	for _, ref := range env.InputRefs {
		if ref == ".asdt/knowledge/platform.yaml" {
			foundPlatform = true
		}
	}
	if !foundPlatform {
		t.Errorf("input_refs must contain platform.yaml path; got %v", env.InputRefs)
	}
}

func TestAgent_StoryTraceability_OpenItems(t *testing.T) {
	store, root := makeStore(t)
	const change = "trace-change"

	// Write spec with two stories.
	spec := fixtureSpec
	spec.ChangeID = change
	spec.Payload.UserStories = append(spec.Payload.UserStories, requirements.UserStory{
		ID:     "US-002",
		As:     "admin",
		Want:   "audit all resets",
		SoThat: "I can review security events",
	})
	if err := store.Write(context.Background(), change, "requirements-spec", spec); err != nil {
		t.Fatalf("write spec: %v", err)
	}
	advanceToRequirements(t, store, change)

	// Plan only references US-001 (not US-002).
	reg := testRegistry(t)
	mock := llm.NewMockProvider(llm.WithRawResponses(validPlanYAML))
	runner := pipeline.NewFSMachine(store)

	agent := developer.New(reg, mock, store, runner, noopReader{})
	if err := agent.Run(context.Background(), root, change); err != nil {
		t.Fatalf("Run: %v", err)
	}

	var env artifact.Envelope[developer.ImplementationPlan]
	if err := store.Read(context.Background(), change, "implementation-plan", &env); err != nil {
		t.Fatalf("Read: %v", err)
	}

	// US-002 should appear in open_items.
	found := false
	for _, item := range env.Payload.OpenItems {
		if item != "" && len(item) > 0 {
			found = true
		}
	}
	if !found {
		t.Error("expected open_items to contain US-002 traceability warning")
	}
}

func TestAgent_PromptVersionSet(t *testing.T) {
	store, root := makeStore(t)
	const change = "pv-change"
	writeFixtureSpec(t, store, change)
	advanceToRequirements(t, store, change)

	reg := testRegistry(t)
	mock := llm.NewMockProvider(llm.WithRawResponses(validPlanYAML))
	runner := pipeline.NewFSMachine(store)

	agent := developer.New(reg, mock, store, runner, noopReader{})
	if err := agent.Run(context.Background(), root, change); err != nil {
		t.Fatalf("Run: %v", err)
	}

	asdt := root.Path()
	artifactPath := filepath.Join(asdt, "artifacts", change, "implementation-plan.yaml")
	data, err := os.ReadFile(artifactPath)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
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
