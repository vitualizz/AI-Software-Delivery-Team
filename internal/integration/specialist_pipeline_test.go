//go:build integration

package integration_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/vitualizz/ai-software-delivery-team/internal/artifact"
	"github.com/vitualizz/ai-software-delivery-team/internal/config"
	"github.com/vitualizz/ai-software-delivery-team/internal/knowledge"
	"github.com/vitualizz/ai-software-delivery-team/internal/llm"
	"github.com/vitualizz/ai-software-delivery-team/internal/memory"
	"github.com/vitualizz/ai-software-delivery-team/internal/pipeline"
	"github.com/vitualizz/ai-software-delivery-team/internal/prompt"
	"github.com/vitualizz/ai-software-delivery-team/internal/specialists"
	"github.com/vitualizz/ai-software-delivery-team/skill"
)

// Scripted LLM responses — one per Developer specialist step.
var developerStepResponses = []string{
	// explore
	`files_to_understand: ["auth.go"]
patterns_to_follow: ["early-return"]
risks: []
open_questions: []
`,
	// spec
	`scope:
  in: ["login endpoint"]
  out: ["OAuth"]
requirements: ["password hashing"]
open_questions: []
`,
	// design
	`approach: JWT tokens
data_model:
  - name: User
    fields: ["id","email","password_hash"]
open_questions: []
`,
	// tasks
	`tasks:
  - id: T-1
    title: Add User model
    estimate: S
open_questions: []
`,
	// implement
	`steps:
  - story_ref: T-1
    title: User model
    files_to_create: ["internal/auth/user.go"]
    code_snippets:
      - file: user.go
        language: go
        content: "type User struct{ID string}"
`,
	// test
	`test_cases:
  - id: TC-1
    title: TestUser
    given: valid user
    when: created
    then: has id
`,
	// review
	`complexity_estimate: M
open_items: []
steps:
  - story_ref: T-1
    title: User model
    files_to_create: ["internal/auth/user.go"]
    rationale: core model
    code_snippets: []
    test_snippets: []
`,
}

// setupSpecialistProject creates a temp project dir with .asdt/ and platform.yaml.
func setupSpecialistProject(t *testing.T) (string, config.Root) {
	t.Helper()
	dir := t.TempDir()

	asdt := filepath.Join(dir, ".asdt")
	for _, sub := range []string{"artifacts", "knowledge"} {
		if err := os.MkdirAll(filepath.Join(asdt, sub), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", sub, err)
		}
	}

	root, err := config.Discover(dir)
	if err != nil {
		t.Fatalf("config.Discover: %v", err)
	}

	reader := knowledge.NewFSReader()
	platform := knowledge.Platform{
		SchemaVersion: "1",
		DetectedStack: []string{"go"},
		Conventions: knowledge.Conventions{
			FileStructure: "hexagonal",
		},
	}
	if err := reader.Write(root, platform); err != nil {
		t.Fatalf("write platform.yaml: %v", err)
	}

	return dir, root
}

// TestDeveloperSpecialistPipeline runs the full Developer pipeline with 7 isolated steps
// and asserts every intermediate and final artifact exists at the correct path.
func TestDeveloperSpecialistPipeline(t *testing.T) {
	ctx := context.Background()
	const change = "si-test"

	dir, root := setupSpecialistProject(t)
	asdt := filepath.Join(dir, ".asdt")

	store := artifact.NewFSStore(asdt)

	// Write a fixture requirements-spec so the explore step's Artifacts.Reads is satisfied.
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
			"user_stories":   []any{},
			"open_questions": []any{},
		},
	}
	if err := store.Write(ctx, change, "requirements-spec", fixtureEnv); err != nil {
		t.Fatalf("write fixture requirements-spec: %v", err)
	}

	// Wire real dependencies.
	reg := prompt.NewEmbeddedRegistry(skill.FS())
	d := specialists.DeveloperDescriptor()

	if len(developerStepResponses) != len(d.Workflow) {
		t.Fatalf("developerStepResponses count %d != workflow steps %d",
			len(developerStepResponses), len(d.Workflow))
	}

	provider := llm.NewMockProvider(llm.WithRawResponses(developerStepResponses...))
	machine := pipeline.NewFSMachine(store)

	deps := specialists.RunnerDeps{
		Registry: reg,
		Composer: prompt.Compose,
		Provider: provider,
		Store:    store,
		Memory:   memory.NewInMemoryProvider(),
		Pipeline: machine,
	}

	// Run the specialist.
	runner := specialists.New(d, deps)
	if err := runner.Run(ctx, root, change); err != nil {
		t.Fatalf("specialist Run: %v", err)
	}

	// Assert all 7 artifacts exist at their expected paths.
	expectedPaths := []struct {
		artifactType string
		relativePath string
	}{
		{"developer/dev-exploration", filepath.Join("developer", "dev-exploration.yaml")},
		{"developer/dev-spec", filepath.Join("developer", "dev-spec.yaml")},
		{"developer/dev-design", filepath.Join("developer", "dev-design.yaml")},
		{"developer/dev-tasks", filepath.Join("developer", "dev-tasks.yaml")},
		{"developer/dev-implementation", filepath.Join("developer", "dev-implementation.yaml")},
		{"developer/dev-tests", filepath.Join("developer", "dev-tests.yaml")},
		{"implementation-plan", "implementation-plan.yaml"},
	}

	for _, ep := range expectedPaths {
		// Check via Store.Exists.
		if !store.Exists(change, ep.artifactType) {
			t.Errorf("artifact %q not found via store", ep.artifactType)
		}

		// Check physical file on disk.
		fullPath := filepath.Join(asdt, "artifacts", change, ep.relativePath)
		if _, err := os.Stat(fullPath); err != nil {
			t.Errorf("file %q does not exist on disk: %v", fullPath, err)
		}

		// Assert each artifact unmarshals into Envelope[map[string]any] without error.
		var env artifact.Envelope[map[string]any]
		if err := store.Read(ctx, change, ep.artifactType, &env); err != nil {
			t.Errorf("artifact %q failed to unmarshal: %v", ep.artifactType, err)
			continue
		}
		if env.EnvelopeHeader.Agent != "developer" {
			t.Errorf("artifact %q: agent = %q, want %q", ep.artifactType, env.EnvelopeHeader.Agent, "developer")
		}
		if env.EnvelopeHeader.ChangeID != change {
			t.Errorf("artifact %q: change_id = %q, want %q", ep.artifactType, env.EnvelopeHeader.ChangeID, change)
		}
		if env.EnvelopeHeader.PromptVersion == "" {
			t.Errorf("artifact %q: prompt_version is empty", ep.artifactType)
		}
	}

	// Assert pipeline-state-v2.yaml has 7 step records for the developer specialist.
	stateKey := pipeline.ArtifactTypeV2
	if !store.Exists(change, stateKey) {
		t.Fatal("pipeline-state-v2 artifact does not exist")
	}

	var sv2 pipeline.PipelineStateV2
	if err := store.Read(ctx, change, stateKey, &sv2); err != nil {
		t.Fatalf("read pipeline-state-v2: %v", err)
	}

	devState, ok := sv2.Specialists["developer"]
	if !ok {
		t.Fatal("pipeline-state-v2 has no 'developer' entry")
	}
	if len(devState.StepsCompleted) != len(d.Workflow) {
		t.Errorf("developer specialist steps_completed = %d, want %d",
			len(devState.StepsCompleted), len(d.Workflow))
	}

	// Assert step IDs recorded match the descriptor order.
	for i, rec := range devState.StepsCompleted {
		wantID := d.Workflow[i].ID
		if rec.ID != wantID {
			t.Errorf("steps_completed[%d].id = %q, want %q", i, rec.ID, wantID)
		}
	}
}
