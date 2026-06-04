//go:build parity

package specialists_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/vitualizz/ai-software-delivery-team/internal/artifact"
	"github.com/vitualizz/ai-software-delivery-team/internal/config"
	"github.com/vitualizz/ai-software-delivery-team/internal/llm"
	"github.com/vitualizz/ai-software-delivery-team/internal/memory"
	"github.com/vitualizz/ai-software-delivery-team/internal/pipeline"
	"github.com/vitualizz/ai-software-delivery-team/internal/prompt"
	"github.com/vitualizz/ai-software-delivery-team/internal/specialists"
	"github.com/vitualizz/ai-software-delivery-team/skill"
)

// parityResponse is the scripted YAML the MockProvider returns on the final
// self-review step. It must satisfy the implementation-plan schema contract
// verified by the parity test assertions.
const parityResponse = `complexity_estimate: M
open_items: []
steps:
  - story_ref: "US-001"
    title: "Implement password reset endpoint"
    files_to_create: ["internal/auth/reset.go"]
    files_to_modify: []
    rationale: "Handles the reset flow"
    code_snippets:
      - file: "internal/auth/reset.go"
        language: "go"
        content: "func ResetPassword(token string) error { return nil }"
    test_snippets:
      - file: "internal/auth/reset_test.go"
        content: "func TestResetPassword(t *testing.T) {}"
`

// fixtureRequirementsSpec is the requirements-spec envelope written before the
// parity run so the DeveloperDescriptor has an upstream artifact to read.
type fixtureRequirementsPayload struct {
	UserStories []struct {
		ID     string `yaml:"id"`
		As     string `yaml:"as"`
		Want   string `yaml:"want"`
		SoThat string `yaml:"so_that"`
	} `yaml:"user_stories"`
	Scope struct {
		In  []string `yaml:"in"`
		Out []string `yaml:"out"`
	} `yaml:"scope"`
}

// TestParityRunner_DeveloperDescriptor_MatchesSchemaContract wires real dependencies
// (FSStore, embedded registry, prompt.Compose, FSMachine, NullProvider) and verifies
// that the SpecialistRunner produces an implementation-plan artifact whose envelope
// header matches the contract previously enforced by internal/developer/agent.go.
func TestParityRunner_DeveloperDescriptor_MatchesSchemaContract(t *testing.T) {
	const change = "parity-change"

	// --- 1. Set up temp project root with .asdt/ structure. ---
	dir := t.TempDir()
	asdt := filepath.Join(dir, ".asdt")
	if err := os.MkdirAll(filepath.Join(asdt, "artifacts", change), 0o755); err != nil {
		t.Fatalf("mkdir artifacts: %v", err)
	}

	root, err := config.Discover(dir)
	if err != nil {
		t.Fatalf("config.Discover: %v", err)
	}

	store := artifact.NewFSStore(asdt)

	// --- 2. Write fixture requirements-spec. ---
	fixtureEnv := artifact.Envelope[map[string]any]{
		EnvelopeHeader: artifact.EnvelopeHeader{
			SchemaVersion: artifact.CurrentSchemaVersion,
			Agent:         "requirements",
			ChangeID:      change,
			CreatedAt:     time.Now().UTC(),
			PromptVersion: "deadbeef",
			InputRefs:     []string{},
		},
		Payload: map[string]any{
			"user_stories": []any{
				map[string]any{
					"id":      "US-001",
					"as":      "developer",
					"want":    "reset my password via email",
					"so_that": "I can regain access when locked out",
				},
			},
			"acceptance_criteria": map[string]any{
				"US-001": []any{
					"User receives a reset link within 60 seconds",
				},
			},
			"scope": map[string]any{
				"in":  []any{"Password reset email flow"},
				"out": []any{"SMS reset"},
			},
			"nfrs":           []any{"Reset link must expire after 15 minutes"},
			"open_questions": []any{},
		},
	}
	if err := store.Write(context.Background(), change, "requirements-spec", fixtureEnv); err != nil {
		t.Fatalf("write fixture requirements-spec: %v", err)
	}

	// --- 3. Wire real dependencies. ---
	reg := prompt.NewEmbeddedRegistry(skill.FS())

	// DeveloperDescriptor has 7 steps; provide the scripted response for the last
	// step (self-review) and a minimal valid YAML for the preceding 6 steps.
	d := specialists.DeveloperDescriptor()
	nSteps := len(d.Workflow)
	responses := make([]string, nSteps)
	for i := 0; i < nSteps-1; i++ {
		responses[i] = "steps: []\nopen_items: []\n"
	}
	responses[nSteps-1] = parityResponse

	provider := llm.NewMockProvider(llm.WithRawResponses(responses...))
	machine := pipeline.NewFSMachine(store)

	deps := specialists.RunnerDeps{
		Registry: reg,
		Composer: prompt.Compose,
		Provider: provider,
		Store:    store,
		Memory:   memory.NullProvider{},
		Pipeline: machine,
	}

	// --- 4. Run the specialist. ---
	runner := specialists.New(d, deps)
	if err := runner.Run(context.Background(), root, change); err != nil {
		t.Fatalf("Run: %v", err)
	}

	// --- 5. Read the written implementation-plan.yaml. ---
	if !store.Exists(change, "implementation-plan") {
		t.Fatal("implementation-plan artifact was not written")
	}

	var env artifact.Envelope[map[string]any]
	if err := store.Read(context.Background(), change, "implementation-plan", &env); err != nil {
		t.Fatalf("Read implementation-plan: %v", err)
	}

	// --- 6. Assert envelope header contract matches internal/developer/agent.go. ---

	if env.SchemaVersion != "1" {
		t.Errorf("SchemaVersion: got %q, want %q", env.SchemaVersion, "1")
	}
	if env.Agent != "developer" {
		t.Errorf("Agent: got %q, want %q", env.Agent, "developer")
	}
	if env.ChangeID != change {
		t.Errorf("ChangeID: got %q, want %q", env.ChangeID, change)
	}
	if env.PromptVersion == "" {
		t.Error("PromptVersion must not be empty")
	}

	// --- 7. Assert payload keys required by the developer schema contract. ---

	if env.Payload["steps"] == nil {
		t.Error("payload must contain a 'steps' key")
	}
	if _, ok := env.Payload["open_items"]; !ok {
		t.Error("payload must contain an 'open_items' key")
	}
}
