//go:build integration

package integration_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

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

// scriptedRequirementsYAML is a realistic YAML requirements-spec payload
// that the mock LLM will return for the requirements agent call.
const scriptedRequirementsYAML = `
user_stories:
  - id: US-001
    as: authenticated user
    want: log in using email and password
    so_that: I can access my personal account securely
  - id: US-002
    as: new user
    want: register with email and password
    so_that: I can create an account and start using the application
acceptance_criteria:
  US-001:
    - A valid email and password combination grants access
    - An invalid combination shows a descriptive error message
    - Login form validates email format client-side
  US-002:
    - Registration requires email, password, and password confirmation
    - Password must be at least 8 characters
    - A confirmation email is sent after successful registration
scope:
  in:
    - email and password login form
    - user registration flow
    - password validation rules
  out:
    - OAuth/social login
    - two-factor authentication
    - password strength meter UI
nfrs:
  - Passwords must be hashed using bcrypt before storage
  - Login endpoint must rate-limit to 5 attempts per minute
open_questions: []
`

// scriptedImplementationPlanYAML is a realistic YAML implementation-plan payload
// that the mock LLM will return for the developer agent call.
// It references US-001 (matching the requirements spec above).
const scriptedImplementationPlanYAML = `
approach: Implement email/password authentication with bcrypt hashing and JWT tokens
steps:
  - story_id: US-001
    files_to_create:
      - internal/auth/handler.go
      - internal/auth/handler_test.go
    files_to_modify:
      - internal/router/router.go
    rationale: Add login endpoint that verifies bcrypt-hashed password and returns JWT
    code_snippets:
      - file: internal/auth/handler.go
        content: |
          func LoginHandler(w http.ResponseWriter, r *http.Request) {
            // validate credentials and return JWT
          }
  - story_id: US-002
    files_to_create:
      - internal/auth/register.go
      - internal/auth/register_test.go
    files_to_modify:
      - internal/router/router.go
    rationale: Add registration endpoint with email uniqueness check and bcrypt hashing
    code_snippets:
      - file: internal/auth/register.go
        content: |
          func RegisterHandler(w http.ResponseWriter, r *http.Request) {
            // hash password with bcrypt and store user
          }
complexity_estimate: medium
open_items: []
`

// setupProject creates a temporary project directory with a .asdt/ structure
// and a fake platform.yaml. Returns the config.Root.
func setupProject(t *testing.T) (string, config.Root) {
	t.Helper()
	dir := t.TempDir()

	asdt := filepath.Join(dir, ".asdt")
	if err := os.MkdirAll(filepath.Join(asdt, "artifacts"), 0o755); err != nil {
		t.Fatalf("mkdir artifacts: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(asdt, "knowledge"), 0o755); err != nil {
		t.Fatalf("mkdir knowledge: %v", err)
	}

	root, err := config.Discover(dir)
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}

	// Write a fake platform.yaml.
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

// TestFullRequirementsToDevPipeline runs the complete requirements → developer
// pipeline end-to-end using MockProvider with scripted responses.
func TestFullRequirementsToDevPipeline(t *testing.T) {
	ctx := context.Background()
	const changeID = "test-change"

	_, root := setupProject(t)

	// Wire all real dependencies with manual DI.
	store := artifact.NewFSStore(root.Path())
	machine := pipeline.NewFSMachine(store)
	registry := prompt.NewEmbeddedRegistry(skill.PromptSubFS())
	reader := knowledge.NewFSReader()

	// MockProvider scripted with two raw responses:
	//   1st Complete call → requirements spec YAML
	//   2nd Complete call → implementation plan YAML
	provider := llm.NewMockProvider(
		llm.WithRawResponses(scriptedRequirementsYAML, scriptedImplementationPlanYAML),
	)

	// --- Step 1: Run requirements agent ---
	reqAgent := requirements.New(registry, provider, store, machine, reader)
	idea := "add user authentication with email and password login form"
	if err := reqAgent.Run(ctx, root, changeID, idea); err != nil {
		t.Fatalf("requirements.Run: %v", err)
	}

	// Assert: requirements-spec.yaml exists.
	if !store.Exists(changeID, "requirements-spec") {
		t.Fatal("requirements-spec.yaml was not created")
	}

	// Assert: file is valid YAML that unmarshals into Envelope[RequirementsSpec].
	var specEnv artifact.Envelope[requirements.RequirementsSpec]
	if err := store.Read(ctx, changeID, "requirements-spec", &specEnv); err != nil {
		t.Fatalf("Read requirements-spec: %v", err)
	}

	// Assert: all required envelope header fields are non-empty.
	if err := artifact.Validate(specEnv.EnvelopeHeader); err != nil {
		t.Errorf("requirements-spec envelope invalid: %v", err)
	}

	// Assert: prompt_version is set.
	if specEnv.PromptVersion == "" {
		t.Error("requirements-spec: prompt_version must not be empty")
	}

	// Assert: UserStories is non-empty.
	if len(specEnv.Payload.UserStories) == 0 {
		t.Error("requirements-spec: UserStories must not be empty")
	}

	// Assert: pipeline-state.yaml shows current_state == requirements.
	pipelineState, err := machine.Current(ctx, changeID)
	if err != nil {
		t.Fatalf("pipeline Current: %v", err)
	}
	if pipelineState.CurrentState != pipeline.PhaseRequirements {
		t.Errorf("pipeline state: got %q, want %q", pipelineState.CurrentState, pipeline.PhaseRequirements)
	}

	// Collect story IDs from spec for traceability check later.
	specStoryIDs := make(map[string]bool)
	for _, s := range specEnv.Payload.UserStories {
		specStoryIDs[s.ID] = true
	}

	// --- Step 2: Run developer agent ---
	devAgent := developer.New(registry, provider, store, machine, reader)
	if err := devAgent.Run(ctx, root, changeID); err != nil {
		t.Fatalf("developer.Run: %v", err)
	}

	// Assert: implementation-plan.yaml exists.
	if !store.Exists(changeID, "implementation-plan") {
		t.Fatal("implementation-plan.yaml was not created")
	}

	// Assert: file unmarshals into Envelope[ImplementationPlan].
	var planEnv artifact.Envelope[developer.ImplementationPlan]
	if err := store.Read(ctx, changeID, "implementation-plan", &planEnv); err != nil {
		t.Fatalf("Read implementation-plan: %v", err)
	}

	// Assert: envelope header is valid.
	if err := artifact.Validate(planEnv.EnvelopeHeader); err != nil {
		t.Errorf("implementation-plan envelope invalid: %v", err)
	}

	// Assert: at least one step references a story ID from the requirements spec.
	foundRef := false
	for _, step := range planEnv.Payload.Steps {
		if specStoryIDs[step.StoryID] {
			foundRef = true
			break
		}
	}
	if !foundRef {
		t.Errorf("implementation-plan: no step references a story ID from requirements-spec (spec IDs: %v)", specStoryIDs)
	}

	// Assert: input_refs includes the requirements-spec path.
	hasSpecRef := false
	for _, ref := range planEnv.InputRefs {
		if containsPath(ref, "requirements-spec") {
			hasSpecRef = true
			break
		}
	}
	if !hasSpecRef {
		t.Errorf("implementation-plan: input_refs %v does not include requirements-spec path", planEnv.InputRefs)
	}

	// Assert: pipeline-state.yaml current_state is now "plan".
	pipelineState2, err := machine.Current(ctx, changeID)
	if err != nil {
		t.Fatalf("pipeline Current (after dev): %v", err)
	}
	if pipelineState2.CurrentState != pipeline.PhasePlan {
		t.Errorf("pipeline state after dev: got %q, want %q", pipelineState2.CurrentState, pipeline.PhasePlan)
	}

	// Verify pipeline-state.yaml is a real file on disk by reading it raw.
	stateFilePath := filepath.Join(root.Path(), "artifacts", changeID, "pipeline-state.yaml")
	stateData, err := os.ReadFile(stateFilePath)
	if err != nil {
		t.Fatalf("read pipeline-state.yaml: %v", err)
	}
	var rawState map[string]interface{}
	if err := yaml.Unmarshal(stateData, &rawState); err != nil {
		t.Fatalf("unmarshal pipeline-state.yaml: %v", err)
	}
	if rawState["current_state"] != "plan" {
		t.Errorf("pipeline-state.yaml current_state: got %v, want 'plan'", rawState["current_state"])
	}
}

// containsPath returns true when the string contains the given path fragment.
func containsPath(s, fragment string) bool {
	for i := 0; i+len(fragment) <= len(s); i++ {
		if s[i:i+len(fragment)] == fragment {
			return true
		}
	}
	return false
}
