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

// Scripted responses — one per Developer step, each valid YAML for that step's output.

const parityExploreResponse = `files_to_understand:
  - auth.go
patterns_to_follow:
  - early-return
risks: []
open_questions: []
`

const paritySpecResponse = `scope:
  in:
    - login endpoint
  out:
    - OAuth
requirements:
  - password hashing
open_questions: []
`

const parityDesignResponse = `approach: JWT tokens
data_model:
  - name: User
    fields:
      - id
      - email
      - password_hash
open_questions: []
`

const parityTasksResponse = `tasks:
  - id: T-1
    title: Add User model
    estimate: S
open_questions: []
`

const parityImplementResponse = `steps:
  - story_ref: T-1
    title: User model
    files_to_create:
      - internal/auth/user.go
    code_snippets:
      - file: user.go
        language: go
        content: "type User struct{ID string}"
`

const parityTestResponse = `test_cases:
  - id: TC-1
    title: TestUser
    given: valid user
    when: created
    then: has id
`

const parityReviewResponse = `complexity_estimate: M
open_items: []
steps:
  - story_ref: T-1
    title: User model
    files_to_create:
      - internal/auth/user.go
    rationale: core model
    code_snippets: []
    test_snippets: []
`

// TestParityRunner_DeveloperDescriptor_PerStepArtifactChain verifies:
//  1. Every Developer step writes its declared OutputArtifact to the correct path.
//  2. The final artifact (implementation-plan) is written after the review step.
//  3. All 7 envelope headers carry correct agent, change_id, and prompt_version.
//  4. Intermediate artifacts live under .asdt/artifacts/{change}/developer/*.yaml.
func TestParityRunner_DeveloperDescriptor_PerStepArtifactChain(t *testing.T) {
	const change = "parity-change"

	// 1. Set up temp project root with .asdt/ structure.
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

	// 2. Write fixture requirements-spec so the explore step has an upstream artifact.
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
			"open_questions": []any{},
		},
	}
	if err := store.Write(context.Background(), change, "requirements-spec", fixtureEnv); err != nil {
		t.Fatalf("write fixture requirements-spec: %v", err)
	}

	// 3. Wire real dependencies with scripted responses (one per Developer step).
	reg := prompt.NewEmbeddedRegistry(skill.FS())
	d := specialists.DeveloperDescriptor()

	responses := []string{
		parityExploreResponse,
		paritySpecResponse,
		parityDesignResponse,
		parityTasksResponse,
		parityImplementResponse,
		parityTestResponse,
		parityReviewResponse,
	}
	if len(responses) != len(d.Workflow) {
		t.Fatalf("responses count %d != workflow steps %d", len(responses), len(d.Workflow))
	}

	provider := llm.NewMockProvider(llm.WithRawResponses(responses...))
	machine := pipeline.NewFSMachine(store)

	deps := specialists.RunnerDeps{
		Registry: reg,
		Composer: prompt.Compose,
		Provider: provider,
		Store:    store,
		Memory:   memory.NewInMemoryProvider(),
		Pipeline: machine,
	}

	// 4. Run the specialist.
	runner := specialists.New(d, deps)
	if err := runner.Run(context.Background(), root, change); err != nil {
		t.Fatalf("Run: %v", err)
	}

	// 5. Assert all 7 artifacts exist at the correct paths.
	expectedArtifacts := []string{
		"developer/dev-exploration",
		"developer/dev-spec",
		"developer/dev-design",
		"developer/dev-tasks",
		"developer/dev-implementation",
		"developer/dev-tests",
		"implementation-plan",
	}

	for _, artifactType := range expectedArtifacts {
		if !store.Exists(change, artifactType) {
			t.Errorf("artifact %q does not exist", artifactType)
		}
	}

	// 6. Assert intermediate artifacts are at developer/ subdirectory.
	intermediateDir := filepath.Join(asdt, "artifacts", change, "developer")
	intermediateFiles := []string{
		"dev-exploration.yaml",
		"dev-spec.yaml",
		"dev-design.yaml",
		"dev-tasks.yaml",
		"dev-implementation.yaml",
		"dev-tests.yaml",
	}
	for _, fname := range intermediateFiles {
		p := filepath.Join(intermediateDir, fname)
		if _, err := os.Stat(p); err != nil {
			t.Errorf("intermediate artifact file %q does not exist: %v", p, err)
		}
	}

	// 7. Assert final artifact is at the top level (not under developer/).
	finalPath := filepath.Join(asdt, "artifacts", change, "implementation-plan.yaml")
	if _, err := os.Stat(finalPath); err != nil {
		t.Errorf("final artifact implementation-plan.yaml does not exist: %v", err)
	}

	// 8. Assert envelope headers on each artifact.
	for _, artifactType := range expectedArtifacts {
		var env artifact.Envelope[map[string]any]
		if err := store.Read(context.Background(), change, artifactType, &env); err != nil {
			t.Fatalf("read %q envelope: %v", artifactType, err)
		}

		h := env.EnvelopeHeader
		if h.SchemaVersion == "" {
			t.Errorf("%s: schema_version is empty", artifactType)
		}
		if h.Agent != "developer" {
			t.Errorf("%s: agent = %q, want %q", artifactType, h.Agent, "developer")
		}
		if h.ChangeID != change {
			t.Errorf("%s: change_id = %q, want %q", artifactType, h.ChangeID, change)
		}
		if h.PromptVersion == "" {
			t.Errorf("%s: prompt_version is empty", artifactType)
		}
		if h.CreatedAt.IsZero() {
			t.Errorf("%s: created_at is zero", artifactType)
		}
	}
}
