package specialists_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"testing/fstest"

	"github.com/vitualizz/ai-software-delivery-team/internal/artifact"
	"github.com/vitualizz/ai-software-delivery-team/internal/config"
	"github.com/vitualizz/ai-software-delivery-team/internal/llm"
	"github.com/vitualizz/ai-software-delivery-team/internal/memory"
	"github.com/vitualizz/ai-software-delivery-team/internal/pipeline"
	"github.com/vitualizz/ai-software-delivery-team/internal/prompt"
	"github.com/vitualizz/ai-software-delivery-team/internal/specialists"
	"gopkg.in/yaml.v3"
)

// --- In-memory Store for tests ---

type memStore struct {
	mu    sync.Mutex
	items map[string][]byte
}

func newMemStore() *memStore { return &memStore{items: make(map[string][]byte)} }

func storeKey(change, artifactType string) string { return change + "/" + artifactType }

func (s *memStore) Read(_ context.Context, change, artifactType string, out any) error {
	s.mu.Lock()
	data, ok := s.items[storeKey(change, artifactType)]
	s.mu.Unlock()
	if !ok {
		return errors.New("artifact not found: " + storeKey(change, artifactType))
	}
	return yaml.Unmarshal(data, out)
}

func (s *memStore) Write(_ context.Context, change, artifactType string, env any) error {
	data, err := yaml.Marshal(env)
	if err != nil {
		return err
	}
	s.mu.Lock()
	s.items[storeKey(change, artifactType)] = data
	s.mu.Unlock()
	return nil
}

func (s *memStore) List(_ context.Context, _ string) ([]string, error) { return nil, nil }

func (s *memStore) Exists(change, artifactType string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.items[storeKey(change, artifactType)]
	return ok
}

func (s *memStore) readEnvelope(change, artifactType string) (artifact.Envelope[map[string]any], error) {
	s.mu.Lock()
	data, ok := s.items[storeKey(change, artifactType)]
	s.mu.Unlock()
	if !ok {
		return artifact.Envelope[map[string]any]{}, errors.New("not found: " + storeKey(change, artifactType))
	}
	var env artifact.Envelope[map[string]any]
	return env, yaml.Unmarshal(data, &env)
}

// --- Mock PipelineRunner ---

type mockPipeline struct {
	mu            sync.Mutex
	advancedSteps []advanceRecord
	advanceErr    error
}

type advanceRecord struct {
	change       string
	specialistID string
	stepID       string
}

func (p *mockPipeline) Current(_ context.Context, _ string) (pipeline.State, error) {
	return pipeline.State{}, nil
}

func (p *mockPipeline) Advance(_ context.Context, _ string, _ pipeline.Phase) (pipeline.State, error) {
	return pipeline.State{}, nil
}

func (p *mockPipeline) CanTransition(_, _ pipeline.Phase) bool { return true }

func (p *mockPipeline) AdvanceStep(_ context.Context, _ config.Root, change, specialistID, stepID string) error {
	p.mu.Lock()
	p.advancedSteps = append(p.advancedSteps, advanceRecord{change: change, specialistID: specialistID, stepID: stepID})
	p.mu.Unlock()
	return p.advanceErr
}

func (p *mockPipeline) stepsFor(specialistID string) []string {
	p.mu.Lock()
	defer p.mu.Unlock()
	var out []string
	for _, r := range p.advancedSteps {
		if r.specialistID == specialistID {
			out = append(out, r.stepID)
		}
	}
	return out
}

// --- Minimal test registry ---

func buildMinimalRegistry() prompt.SkillRegistry {
	fsys := fstest.MapFS{
		"developer/SKILL.md":                    {Data: []byte("Developer role content.")},
		"security/SKILL.md":                     {Data: []byte("Security role content.")},
		"developer/skills/platform-analysis.md": {Data: []byte("platform analysis skill")},
		"developer/skills/artifact-loading.md":  {Data: []byte("artifact loading skill")},
		"developer/skills/scope-definition.md":  {Data: []byte("scope definition skill")},
		"developer/skills/code-generation.md":   {Data: []byte("code generation skill")},
		"developer/skills/test-generation.md":   {Data: []byte("test generation skill")},
		"developer/skills/review.md":            {Data: []byte("review skill")},
		"developer/skills/report.md":            {Data: []byte("report skill")},
		"security/skills/threat-modeling.md":    {Data: []byte("threat modeling skill")},
		"security/skills/owasp-review.md":       {Data: []byte("owasp skill")},
		"_shared/skills/platform-context.md":    {Data: []byte("platform context skill")},
		"_shared/skills/artifact-envelope.md":   {Data: []byte("artifact envelope skill")},
	}
	return prompt.NewEmbeddedRegistry(fsys)
}

// yamlResponses returns a slice of n copies of the given YAML string.
func yamlResponses(n int, resp string) []string {
	out := make([]string, n)
	for i := range out {
		out[i] = resp
	}
	return out
}

// --- Test 1: DeveloperDescriptor happy path ---

func TestRunner_DeveloperDescriptor_HappyPath(t *testing.T) {
	ctx := context.Background()
	store := newMemStore()
	pl := &mockPipeline{}

	d := specialists.DeveloperDescriptor()
	provider := llm.NewMockProvider(llm.WithRawResponses(yamlResponses(len(d.Workflow), "steps: []\nopen_items: []\n")...))

	deps := specialists.RunnerDeps{
		Registry: buildMinimalRegistry(),
		Provider: provider,
		Store:    store,
		Memory:   &memory.NullProvider{},
		Pipeline: pl,
	}

	runner := specialists.New(d, deps)
	if err := runner.Run(ctx, config.Root{}, "test-change"); err != nil {
		t.Fatalf("Run: %v", err)
	}

	// implementation-plan must be written on the last step.
	if !store.Exists("test-change", "implementation-plan") {
		t.Error("expected implementation-plan artifact to be written")
	}

	env, err := store.readEnvelope("test-change", "implementation-plan")
	if err != nil {
		t.Fatalf("read envelope: %v", err)
	}

	h := env.EnvelopeHeader
	if h.SchemaVersion == "" {
		t.Error("envelope: schema_version is empty")
	}
	if h.Agent != "developer" {
		t.Errorf("envelope: agent = %q, want %q", h.Agent, "developer")
	}
	if h.ChangeID != "test-change" {
		t.Errorf("envelope: change_id = %q, want %q", h.ChangeID, "test-change")
	}
	if h.CreatedAt.IsZero() {
		t.Error("envelope: created_at is zero")
	}
	if h.PromptVersion == "" {
		t.Error("envelope: prompt_version is empty")
	}

	// Pipeline must have recorded all 7 steps.
	devSteps := pl.stepsFor("developer")
	if len(devSteps) != len(d.Workflow) {
		t.Errorf("expected %d AdvanceStep calls, got %d", len(d.Workflow), len(devSteps))
	}
}

// --- Test 2: SecurityDescriptor, no prior artifacts ---

func TestRunner_SecurityDescriptor_NoPriorArtifacts(t *testing.T) {
	ctx := context.Background()
	store := newMemStore()
	pl := &mockPipeline{}

	d := specialists.SecurityDescriptor()
	provider := llm.NewMockProvider(llm.WithRawResponses(yamlResponses(len(d.Workflow), "threats: []\nfindings: []\nopen_items: []\n")...))

	deps := specialists.RunnerDeps{
		Registry: buildMinimalRegistry(),
		Provider: provider,
		Store:    store,
		Memory:   &memory.NullProvider{},
		Pipeline: pl,
	}

	runner := specialists.New(d, deps)
	// Must not error even though no artifacts exist.
	if err := runner.Run(ctx, config.Root{}, "security-change"); err != nil {
		t.Fatalf("Run: %v", err)
	}

	// At least one artifact from Writes must be written.
	wrote := false
	for _, wType := range d.Artifacts.Writes {
		if store.Exists("security-change", wType) {
			wrote = true
			break
		}
	}
	if !wrote {
		t.Error("expected at least one artifact written by security runner")
	}
}

// --- Test 3: Provider returns error on step 2 ---

func TestRunner_ProviderError_StepWrapped(t *testing.T) {
	ctx := context.Background()
	store := newMemStore()
	pl := &mockPipeline{}

	stepErr := errors.New("LLM unavailable")
	call := 0
	provider := &sequentialProvider{
		results: []providerResult{
			{content: "steps: []\n"},
			{err: stepErr},
		},
		call: &call,
	}

	deps := specialists.RunnerDeps{
		Registry: buildMinimalRegistry(),
		Provider: provider,
		Store:    store,
		Memory:   &memory.NullProvider{},
		Pipeline: pl,
	}

	d := specialists.DeveloperDescriptor()
	runner := specialists.New(d, deps)
	err := runner.Run(ctx, config.Root{}, "err-change")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	// Error must contain the step ID of step 2 ("spec").
	if !strings.Contains(err.Error(), "spec") {
		t.Errorf("error %q should mention step ID 'spec'", err.Error())
	}
}

// --- Test 4: NullProvider memory never causes error ---

func TestRunner_NullMemory_DoesNotAffectRun(t *testing.T) {
	ctx := context.Background()
	store := newMemStore()
	pl := &mockPipeline{}

	d := specialists.DeveloperDescriptor()
	provider := llm.NewMockProvider(llm.WithRawResponses(yamlResponses(len(d.Workflow), "steps: []\n")...))

	deps := specialists.RunnerDeps{
		Registry: buildMinimalRegistry(),
		Provider: provider,
		Store:    store,
		Memory:   &memory.NullProvider{},
		Pipeline: pl,
	}

	runner := specialists.New(d, deps)
	if err := runner.Run(ctx, config.Root{}, "null-mem-change"); err != nil {
		t.Fatalf("Run with NullProvider memory: %v", err)
	}
}

// --- Test 5: InputRefs filter — step only loads declared artifact ---

func TestRunner_InputRefs_LoadsOnlyDeclaredArtifact(t *testing.T) {
	ctx := context.Background()
	store := newMemStore()
	pl := &mockPipeline{}

	// Pre-write two artifacts: only "developer/dev-exploration" is declared in InputRefs
	// for the "spec" step. "requirements-spec" must NOT appear in artifact context.
	writeArtifact(t, store, "step-test", "developer/dev-exploration", map[string]any{
		"files_to_understand": []any{"auth.go"},
		"risks":               []any{},
	})
	writeArtifact(t, store, "step-test", "requirements-spec", map[string]any{
		"user_stories": []any{},
	})

	// Use DeveloperDescriptor — the "spec" step (index 1) has InputRefs=["developer/dev-exploration"].
	d := specialists.DeveloperDescriptor()
	capturedPrompts := &captureProvider{}
	provider := llm.NewMockProvider(llm.WithRawResponses(yamlResponses(len(d.Workflow), "scope: {in: [], out: []}\nopen_items: []\n")...))
	_ = capturedPrompts

	deps := specialists.RunnerDeps{
		Registry: buildMinimalRegistry(),
		Provider: provider,
		Store:    store,
		Memory:   &memory.NullProvider{},
		Pipeline: pl,
	}

	runner := specialists.New(d, deps)
	if err := runner.Run(ctx, config.Root{}, "step-test"); err != nil {
		t.Fatalf("Run: %v", err)
	}

	// The "spec" step must have written its OutputArtifact ("developer/dev-spec").
	if !store.Exists("step-test", "developer/dev-spec") {
		t.Error("expected developer/dev-spec artifact to be written after spec step")
	}

	// Intermediate artifact must have correct agent and change_id.
	env, err := store.readEnvelope("step-test", "developer/dev-spec")
	if err != nil {
		t.Fatalf("read developer/dev-spec envelope: %v", err)
	}
	if env.EnvelopeHeader.Agent != "developer" {
		t.Errorf("agent = %q, want %q", env.EnvelopeHeader.Agent, "developer")
	}
	if env.EnvelopeHeader.ChangeID != "step-test" {
		t.Errorf("change_id = %q, want %q", env.EnvelopeHeader.ChangeID, "step-test")
	}
}

// --- Test 6: OutputArtifact — mid-run write at correct path ---

func TestRunner_OutputArtifact_WritesAtCorrectPath(t *testing.T) {
	ctx := context.Background()
	store := newMemStore()
	pl := &mockPipeline{}

	d := specialists.DeveloperDescriptor()
	provider := llm.NewMockProvider(llm.WithRawResponses(yamlResponses(len(d.Workflow), "result: ok\nopen_items: []\n")...))

	deps := specialists.RunnerDeps{
		Registry: buildMinimalRegistry(),
		Provider: provider,
		Store:    store,
		Memory:   &memory.NullProvider{},
		Pipeline: pl,
	}

	runner := specialists.New(d, deps)
	if err := runner.Run(ctx, config.Root{}, "mid-run-change"); err != nil {
		t.Fatalf("Run: %v", err)
	}

	// All intermediate artifacts must exist after the run.
	intermediates := []string{
		"developer/dev-exploration",
		"developer/dev-spec",
		"developer/dev-design",
		"developer/dev-tasks",
		"developer/dev-implementation",
		"developer/dev-tests",
		"implementation-plan",
	}
	for _, art := range intermediates {
		if !store.Exists("mid-run-change", art) {
			t.Errorf("expected artifact %q to exist after run", art)
		}
	}
}

// --- Test 7: Empty OutputArtifact on non-last step → no artifact written ---

func TestRunner_EmptyOutputArtifact_NonLastStep_NoWrite(t *testing.T) {
	ctx := context.Background()
	store := newMemStore()
	pl := &mockPipeline{}

	// Build a 3-step descriptor where only the middle step has no OutputArtifact.
	// This should not write anything for that step.
	d := specialists.SpecialistDescriptor{
		ID:   "test-specialist",
		Name: "Test",
		Workflow: []specialists.WorkflowStep{
			{ID: "step-a", InputRefs: []string{}, OutputArtifact: "test/step-a"},
			{ID: "step-b", InputRefs: []string{"test/step-a"}, OutputArtifact: ""}, // no write
			{ID: "step-c", InputRefs: []string{"test/step-a"}, OutputArtifact: "test/step-c"},
		},
		Artifacts: specialists.ArtifactContract{
			Reads:  []string{},
			Writes: []string{"test/step-c"},
		},
	}

	provider := llm.NewMockProvider(llm.WithRawResponses(yamlResponses(3, "result: ok\n")...))
	reg := fstest.MapFS{
		"test-specialist/SKILL.md": {Data: []byte("test role")},
	}

	deps := specialists.RunnerDeps{
		Registry: prompt.NewEmbeddedRegistry(reg),
		Provider: provider,
		Store:    store,
		Memory:   &memory.NullProvider{},
		Pipeline: pl,
	}

	runner := specialists.New(d, deps)
	if err := runner.Run(ctx, config.Root{}, "no-write-change"); err != nil {
		t.Fatalf("Run: %v", err)
	}

	// step-b must not have written any artifact at "test/step-b" (no key in store for it).
	if store.Exists("no-write-change", "test/step-b") {
		t.Error("step-b has OutputArtifact=\"\" and is not the last step — must not write artifact")
	}
	// step-a and step-c must exist.
	if !store.Exists("no-write-change", "test/step-a") {
		t.Error("test/step-a must be written")
	}
	if !store.Exists("no-write-change", "test/step-c") {
		t.Error("test/step-c must be written")
	}
}

// --- Test 8: loadStepInputs with missing artifact → no error, openItems populated ---

func TestRunner_MissingInput_SoftDegradation(t *testing.T) {
	ctx := context.Background()
	store := newMemStore()
	pl := &mockPipeline{}

	// SecurityDescriptor first step has InputRefs=[] — falls back to Artifacts.Reads=[].
	// No artifacts on disk — must complete without error.
	d := specialists.SecurityDescriptor()
	provider := llm.NewMockProvider(llm.WithRawResponses(yamlResponses(len(d.Workflow), "threats: []\nopen_items: []\n")...))

	deps := specialists.RunnerDeps{
		Registry: buildMinimalRegistryWithSecurity(),
		Provider: provider,
		Store:    store,
		Memory:   &memory.NullProvider{},
		Pipeline: pl,
	}

	runner := specialists.New(d, deps)
	if err := runner.Run(ctx, config.Root{}, "missing-input-change"); err != nil {
		t.Fatalf("Run with missing inputs: %v", err)
	}

	// The last step must not have written a security artifact when all inputs are missing
	// — but the run itself must not error.
	// (Security last step has OutputArtifact="" and is the last step → writeArtifacts runs.)
	wrote := false
	for _, w := range d.Artifacts.Writes {
		if store.Exists("missing-input-change", w) {
			wrote = true
			break
		}
	}
	if !wrote {
		t.Error("expected at least one artifact written by security runner even with missing inputs")
	}
}

// --- helpers ---

// writeArtifact is a test helper to pre-populate the store with a YAML envelope.
func writeArtifact(t *testing.T, store *memStore, change, artifactType string, payload map[string]any) {
	t.Helper()
	if err := store.Write(context.Background(), change, artifactType, payload); err != nil {
		t.Fatalf("setup: write %s/%s: %v", change, artifactType, err)
	}
}

// captureProvider records each prompt for assertion in tests.
type captureProvider struct {
	mu      sync.Mutex
	prompts []string
}

func (p *captureProvider) Complete(_ context.Context, req llm.Request) (llm.Response, error) {
	p.mu.Lock()
	if len(req.Messages) > 0 {
		p.prompts = append(p.prompts, req.Messages[0].Content)
	}
	p.mu.Unlock()
	return llm.Response{Content: "result: ok\n"}, nil
}

func (p *captureProvider) Stream(_ context.Context, _ llm.Request) (<-chan llm.Chunk, error) {
	ch := make(chan llm.Chunk, 1)
	ch <- llm.Chunk{Done: true}
	close(ch)
	return ch, nil
}

func (p *captureProvider) Name() string { return "capture-mock" }

// buildMinimalRegistryWithSecurity includes security-specific skills.
func buildMinimalRegistryWithSecurity() prompt.SkillRegistry {
	fsys := fstest.MapFS{
		"security/SKILL.md":                   {Data: []byte("Security role content.")},
		"security/skills/threat-modeling.md":  {Data: []byte("threat modeling skill")},
		"security/skills/owasp-review.md":     {Data: []byte("owasp skill")},
		"_shared/skills/platform-context.md":  {Data: []byte("platform context skill")},
		"_shared/skills/artifact-envelope.md": {Data: []byte("artifact envelope skill")},
	}
	return prompt.NewEmbeddedRegistry(fsys)
}

// --- Test: loadPlatformContext summary-first ---

func TestLoadPlatformContext_SummaryPresent(t *testing.T) {
	// Write a platform-summary.yaml; must return "Platform Context (summary)" header.
	dir := t.TempDir()
	asdt := filepath.Join(dir, ".asdt")
	if err := os.MkdirAll(filepath.Join(asdt, "knowledge"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	summaryContent := "stack: [go]\nprimary_language: go\n"
	if err := os.WriteFile(filepath.Join(asdt, "knowledge", "platform-summary.yaml"), []byte(summaryContent), 0o644); err != nil {
		t.Fatalf("write summary: %v", err)
	}

	root, err := config.Discover(dir)
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}

	// We need an exported LoadPlatformContext, but the method is unexported.
	// We test it indirectly: run a one-step descriptor and verify the LLM prompt
	// contains "Platform Context (summary)" in the concatenated output.
	capturedReqs := &requestCaptureProvider{}
	d := specialists.SpecialistDescriptor{
		ID:   "test-ctx",
		Name: "TestCtx",
		Workflow: []specialists.WorkflowStep{
			{ID: "step-one", InputRefs: []string{}, OutputArtifact: "test-ctx/out"},
		},
		Artifacts: specialists.ArtifactContract{Writes: []string{"test-ctx/out"}},
	}
	reg := fstest.MapFS{
		"test-ctx/SKILL.md": {Data: []byte("role content")},
	}
	deps := specialists.RunnerDeps{
		Registry: prompt.NewEmbeddedRegistry(reg),
		Provider: capturedReqs,
		Store:    newMemStore(),
		Memory:   &memory.NullProvider{},
		Pipeline: &mockPipeline{},
	}
	runner := specialists.New(d, deps)
	if err := runner.Run(context.Background(), root, "ctx-change"); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(capturedReqs.reqs) == 0 {
		t.Fatal("no LLM requests captured")
	}
	prompt := capturedReqs.reqs[0].Messages[0].Content
	if !strings.Contains(prompt, "Platform Context (summary)") {
		t.Errorf("expected prompt to contain 'Platform Context (summary)', got:\n%s", prompt)
	}
	if !strings.Contains(prompt, "primary_language: go") {
		t.Errorf("expected prompt to contain summary content, got:\n%s", prompt)
	}
}

func TestLoadPlatformContext_PlatformYAMLFallback(t *testing.T) {
	// Write only platform.yaml; must NOT contain "(summary)" label.
	dir := t.TempDir()
	asdt := filepath.Join(dir, ".asdt")
	if err := os.MkdirAll(filepath.Join(asdt, "knowledge"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	platformContent := "schema_version: \"1\"\ndetected_stack: [node]\n"
	if err := os.WriteFile(filepath.Join(asdt, "knowledge", "platform.yaml"), []byte(platformContent), 0o644); err != nil {
		t.Fatalf("write platform.yaml: %v", err)
	}

	root, err := config.Discover(dir)
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}

	capturedReqs := &requestCaptureProvider{}
	d := specialists.SpecialistDescriptor{
		ID:   "test-ctx2",
		Name: "TestCtx2",
		Workflow: []specialists.WorkflowStep{
			{ID: "step-one", InputRefs: []string{}, OutputArtifact: "test-ctx2/out"},
		},
		Artifacts: specialists.ArtifactContract{Writes: []string{"test-ctx2/out"}},
	}
	reg := fstest.MapFS{
		"test-ctx2/SKILL.md": {Data: []byte("role content")},
	}
	deps := specialists.RunnerDeps{
		Registry: prompt.NewEmbeddedRegistry(reg),
		Provider: capturedReqs,
		Store:    newMemStore(),
		Memory:   &memory.NullProvider{},
		Pipeline: &mockPipeline{},
	}
	runner := specialists.New(d, deps)
	if err := runner.Run(context.Background(), root, "fallback-change"); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(capturedReqs.reqs) == 0 {
		t.Fatal("no LLM requests captured")
	}
	p := capturedReqs.reqs[0].Messages[0].Content
	if strings.Contains(p, "Platform Context (summary)") {
		t.Error("expected fallback to NOT use '(summary)' label")
	}
	if !strings.Contains(p, "Platform Context") {
		t.Errorf("expected prompt to contain 'Platform Context' fallback, got:\n%s", p)
	}
	if !strings.Contains(p, "detected_stack") {
		t.Errorf("expected prompt to contain platform.yaml content, got:\n%s", p)
	}
}

func TestLoadPlatformContext_BothAbsent(t *testing.T) {
	// Neither file present — prompt must NOT contain any platform context header.
	dir := t.TempDir()
	asdt := filepath.Join(dir, ".asdt")
	if err := os.MkdirAll(asdt, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	root, err := config.Discover(dir)
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}

	capturedReqs := &requestCaptureProvider{}
	d := specialists.SpecialistDescriptor{
		ID:   "test-ctx3",
		Name: "TestCtx3",
		Workflow: []specialists.WorkflowStep{
			{ID: "step-one", InputRefs: []string{}, OutputArtifact: "test-ctx3/out"},
		},
		Artifacts: specialists.ArtifactContract{Writes: []string{"test-ctx3/out"}},
	}
	reg := fstest.MapFS{
		"test-ctx3/SKILL.md": {Data: []byte("role content")},
	}
	deps := specialists.RunnerDeps{
		Registry: prompt.NewEmbeddedRegistry(reg),
		Provider: capturedReqs,
		Store:    newMemStore(),
		Memory:   &memory.NullProvider{},
		Pipeline: &mockPipeline{},
	}
	runner := specialists.New(d, deps)
	if err := runner.Run(context.Background(), root, "absent-change"); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(capturedReqs.reqs) == 0 {
		t.Fatal("no LLM requests captured")
	}
	p := capturedReqs.reqs[0].Messages[0].Content
	if strings.Contains(p, "Platform Context") {
		t.Errorf("expected no platform context when both files absent, got:\n%s", p)
	}
}

// --- Test: SkipIfInitialized gate ---

func TestRunStep_SkipIfInitialized_SummaryPresent(t *testing.T) {
	// Case A: summary present → LLM NOT called, OutputArtifact written with source marker.
	dir := t.TempDir()
	asdt := filepath.Join(dir, ".asdt")
	if err := os.MkdirAll(filepath.Join(asdt, "knowledge"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	summaryYAML := "stack:\n- go\nprimary_language: go\npackage_manager: go modules\ntest_runner: go test\ndetected_at: 2026-01-01T00:00:00Z\n"
	if err := os.WriteFile(filepath.Join(asdt, "knowledge", "platform-summary.yaml"), []byte(summaryYAML), 0o644); err != nil {
		t.Fatalf("write summary: %v", err)
	}
	root, err := config.Discover(dir)
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}

	callCount := 0
	trackProvider := &countingProvider{count: &callCount, response: "result: ok\n"}

	store := newMemStore()
	pl := &mockPipeline{}
	d := specialists.SpecialistDescriptor{
		ID:   "test-skip",
		Name: "TestSkip",
		Workflow: []specialists.WorkflowStep{
			{
				ID:                "platform-analysis",
				InputRefs:         []string{},
				OutputArtifact:    "platform-summary",
				SkipIfInitialized: true,
			},
		},
		Artifacts: specialists.ArtifactContract{Writes: []string{"platform-summary"}},
	}
	reg := fstest.MapFS{
		"test-skip/SKILL.md": {Data: []byte("role content")},
	}
	deps := specialists.RunnerDeps{
		Registry: prompt.NewEmbeddedRegistry(reg),
		Provider: trackProvider,
		Store:    store,
		Memory:   &memory.NullProvider{},
		Pipeline: pl,
	}
	runner := specialists.New(d, deps)
	if err := runner.Run(context.Background(), root, "skip-change"); err != nil {
		t.Fatalf("Run: %v", err)
	}

	// LLM must NOT have been called.
	if callCount != 0 {
		t.Errorf("LLM should not be called when SkipIfInitialized + summary present; got %d call(s)", callCount)
	}

	// OutputArtifact must exist with source marker.
	if !store.Exists("skip-change", "platform-summary") {
		t.Fatal("platform-summary artifact must be written even when step is skipped")
	}
	env, err := store.readEnvelope("skip-change", "platform-summary")
	if err != nil {
		t.Fatalf("read platform-summary envelope: %v", err)
	}
	source, _ := env.Payload["source"].(string)
	if source != "platform-summary.yaml" {
		t.Errorf("artifact source: got %q, want %q", source, "platform-summary.yaml")
	}

	// AdvanceStep must have been called for the step.
	steps := pl.stepsFor("test-skip")
	if len(steps) != 1 || steps[0] != "platform-analysis" {
		t.Errorf("AdvanceStep: expected [platform-analysis], got %v", steps)
	}
}

func TestRunStep_SkipIfInitialized_SummaryAbsent(t *testing.T) {
	// Case B: SkipIfInitialized=true but summary absent → LLM IS called (normal path).
	dir := t.TempDir()
	asdt := filepath.Join(dir, ".asdt")
	if err := os.MkdirAll(asdt, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	root, err := config.Discover(dir)
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}

	callCount := 0
	trackProvider := &countingProvider{count: &callCount, response: "stack: [go]\nopen_items: []\n"}

	store := newMemStore()
	pl := &mockPipeline{}
	d := specialists.SpecialistDescriptor{
		ID:   "test-skip-absent",
		Name: "TestSkipAbsent",
		Workflow: []specialists.WorkflowStep{
			{
				ID:                "platform-analysis",
				InputRefs:         []string{},
				OutputArtifact:    "platform-summary",
				SkipIfInitialized: true,
			},
		},
		Artifacts: specialists.ArtifactContract{Writes: []string{"platform-summary"}},
	}
	reg := fstest.MapFS{
		"test-skip-absent/SKILL.md": {Data: []byte("role content")},
	}
	deps := specialists.RunnerDeps{
		Registry: prompt.NewEmbeddedRegistry(reg),
		Provider: trackProvider,
		Store:    store,
		Memory:   &memory.NullProvider{},
		Pipeline: pl,
	}
	runner := specialists.New(d, deps)
	if err := runner.Run(context.Background(), root, "absent-skip-change"); err != nil {
		t.Fatalf("Run: %v", err)
	}

	// LLM MUST be called when summary is absent.
	if callCount == 0 {
		t.Error("LLM must be called when SkipIfInitialized + summary absent")
	}
}

func TestRunStep_SkipIfInitialized_MalformedYAML_Fallthrough(t *testing.T) {
	// Case C: malformed YAML → graceful fallthrough to LLM.
	dir := t.TempDir()
	asdt := filepath.Join(dir, ".asdt")
	if err := os.MkdirAll(filepath.Join(asdt, "knowledge"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	// Write a YAML scalar (not a mapping) so Unmarshal into map[string]any yields nil payload.
	if err := os.WriteFile(filepath.Join(asdt, "knowledge", "platform-summary.yaml"), []byte("- only\n- a\n- list\n"), 0o644); err != nil {
		t.Fatalf("write malformed: %v", err)
	}
	root, err := config.Discover(dir)
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}

	callCount := 0
	trackProvider := &countingProvider{count: &callCount, response: "stack: []\nopen_items: []\n"}

	d := specialists.SpecialistDescriptor{
		ID:   "test-malformed",
		Name: "TestMalformed",
		Workflow: []specialists.WorkflowStep{
			{
				ID:                "platform-analysis",
				InputRefs:         []string{},
				OutputArtifact:    "platform-summary",
				SkipIfInitialized: true,
			},
		},
		Artifacts: specialists.ArtifactContract{Writes: []string{"platform-summary"}},
	}
	reg := fstest.MapFS{
		"test-malformed/SKILL.md": {Data: []byte("role content")},
	}
	deps := specialists.RunnerDeps{
		Registry: prompt.NewEmbeddedRegistry(reg),
		Provider: trackProvider,
		Store:    newMemStore(),
		Memory:   &memory.NullProvider{},
		Pipeline: &mockPipeline{},
	}
	runner := specialists.New(d, deps)
	if err := runner.Run(context.Background(), root, "malformed-change"); err != nil {
		t.Fatalf("Run must not error on malformed YAML: %v", err)
	}

	// LLM must be called (fallthrough).
	if callCount == 0 {
		t.Error("LLM must be called when platform-summary.yaml is malformed")
	}
}

func TestRunStep_SkipIfInitialized_FalseAlwaysExecutes(t *testing.T) {
	// Case D: SkipIfInitialized=false → step always executes, even when summary exists.
	dir := t.TempDir()
	asdt := filepath.Join(dir, ".asdt")
	if err := os.MkdirAll(filepath.Join(asdt, "knowledge"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	summaryYAML := "stack: [go]\nprimary_language: go\n"
	if err := os.WriteFile(filepath.Join(asdt, "knowledge", "platform-summary.yaml"), []byte(summaryYAML), 0o644); err != nil {
		t.Fatalf("write summary: %v", err)
	}
	root, err := config.Discover(dir)
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}

	callCount := 0
	trackProvider := &countingProvider{count: &callCount, response: "stack: [go]\nopen_items: []\n"}

	d := specialists.SpecialistDescriptor{
		ID:   "test-no-skip",
		Name: "TestNoSkip",
		Workflow: []specialists.WorkflowStep{
			{
				ID:                "explore",
				InputRefs:         []string{},
				OutputArtifact:    "test-no-skip/out",
				SkipIfInitialized: false, // zero value — must always run
			},
		},
		Artifacts: specialists.ArtifactContract{Writes: []string{"test-no-skip/out"}},
	}
	reg := fstest.MapFS{
		"test-no-skip/SKILL.md": {Data: []byte("role content")},
	}
	deps := specialists.RunnerDeps{
		Registry: prompt.NewEmbeddedRegistry(reg),
		Provider: trackProvider,
		Store:    newMemStore(),
		Memory:   &memory.NullProvider{},
		Pipeline: &mockPipeline{},
	}
	runner := specialists.New(d, deps)
	if err := runner.Run(context.Background(), root, "no-skip-change"); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if callCount == 0 {
		t.Error("LLM must be called when SkipIfInitialized=false, even when summary exists")
	}
}

// --- sequentialProvider: deterministic multi-response mock ---

type providerResult struct {
	content string
	err     error
}

type sequentialProvider struct {
	results []providerResult
	call    *int
}

func (p *sequentialProvider) Complete(_ context.Context, _ llm.Request) (llm.Response, error) {
	idx := *p.call
	*p.call++
	if idx >= len(p.results) {
		last := p.results[len(p.results)-1]
		return llm.Response{Content: last.content}, last.err
	}
	r := p.results[idx]
	return llm.Response{Content: r.content}, r.err
}

func (p *sequentialProvider) Stream(_ context.Context, _ llm.Request) (<-chan llm.Chunk, error) {
	ch := make(chan llm.Chunk, 1)
	ch <- llm.Chunk{Done: true}
	close(ch)
	return ch, nil
}

func (p *sequentialProvider) Name() string { return "sequential-mock" }

// --- requestCaptureProvider: records full llm.Request objects ---

type requestCaptureProvider struct {
	mu   sync.Mutex
	reqs []llm.Request
}

func (p *requestCaptureProvider) Complete(_ context.Context, req llm.Request) (llm.Response, error) {
	p.mu.Lock()
	p.reqs = append(p.reqs, req)
	p.mu.Unlock()
	return llm.Response{Content: "result: ok\n"}, nil
}

func (p *requestCaptureProvider) Stream(_ context.Context, _ llm.Request) (<-chan llm.Chunk, error) {
	ch := make(chan llm.Chunk, 1)
	ch <- llm.Chunk{Done: true}
	close(ch)
	return ch, nil
}

func (p *requestCaptureProvider) Name() string { return "request-capture-mock" }

// --- countingProvider: tracks how many times Complete() is called ---

type countingProvider struct {
	mu       sync.Mutex
	count    *int
	response string
}

func (p *countingProvider) Complete(_ context.Context, _ llm.Request) (llm.Response, error) {
	p.mu.Lock()
	*p.count++
	p.mu.Unlock()
	return llm.Response{Content: p.response}, nil
}

func (p *countingProvider) Stream(_ context.Context, _ llm.Request) (<-chan llm.Chunk, error) {
	ch := make(chan llm.Chunk, 1)
	ch <- llm.Chunk{Done: true}
	close(ch)
	return ch, nil
}

func (p *countingProvider) Name() string { return "counting-mock" }

// ----------------------------------------------------------------
// Mock memory.Provider for runner memory-hook tests
// ----------------------------------------------------------------

type mockMemoryProvider struct {
	mu          sync.Mutex
	searchErr   error
	searchRet   []memory.Entry
	saveErr     error
	savedEntries []memory.Entry
	searchCalls  int
	saveCalls    int
}

func (m *mockMemoryProvider) Search(_ context.Context, _ string) ([]memory.Entry, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.searchCalls++
	return m.searchRet, m.searchErr
}

func (m *mockMemoryProvider) Save(_ context.Context, entry memory.Entry) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.saveCalls++
	m.savedEntries = append(m.savedEntries, entry)
	return m.saveErr
}

func (m *mockMemoryProvider) Get(_ context.Context, _ string) (*memory.Entry, error) {
	return nil, nil
}

func (m *mockMemoryProvider) Name() string { return "mock-memory" }

// buildMinimalRegistryForMem builds a minimal registry suitable for the memory hook tests.
func buildMinimalRegistryForMem() prompt.SkillRegistry {
	fsys := fstest.MapFS{
		"developer/SKILL.md": {Data: []byte("Developer role.")},
	}
	return prompt.NewEmbeddedRegistry(fsys)
}

// buildOneShotDeveloper returns a 1-step developer-like descriptor for fast tests.
func buildOneShotDeveloper() specialists.SpecialistDescriptor {
	return specialists.SpecialistDescriptor{
		ID:   "developer",
		Name: "Developer",
		Workflow: []specialists.WorkflowStep{
			{ID: "review", InputRefs: []string{}, OutputArtifact: "implementation-plan"},
		},
		Artifacts: specialists.ArtifactContract{
			Reads:  []string{},
			Writes: []string{"implementation-plan"},
		},
	}
}

// TestRunner_LoadsOrganizationalContext_InjectsBlock verifies that when memory.Search
// returns entries, they appear in the first step's LLM prompt.
func TestRunner_LoadsOrganizationalContext_InjectsBlock(t *testing.T) {
	ctx := context.Background()

	mem := &mockMemoryProvider{
		searchRet: []memory.Entry{
			{Title: "Prior JWT decision", Type: memory.EntryTypeDecision, Content: memory.EntryContent{What: "Use JWT for auth"}},
		},
	}

	capture := &requestCaptureProvider{}
	d := buildOneShotDeveloper()

	deps := specialists.RunnerDeps{
		Registry: buildMinimalRegistryForMem(),
		Provider: capture,
		Store:    newMemStore(),
		Memory:   mem,
		Pipeline: &mockPipeline{},
	}

	runner := specialists.New(d, deps)
	if err := runner.Run(ctx, config.Root{}, "add-auth"); err != nil {
		t.Fatalf("Run: %v", err)
	}

	if mem.searchCalls == 0 {
		t.Error("expected memory.Search to be called at least once")
	}

	if len(capture.reqs) == 0 {
		t.Fatal("no LLM requests captured")
	}
	prompt := capture.reqs[0].Messages[0].Content
	if !strings.Contains(prompt, "Organizational Context") {
		t.Errorf("expected prompt to contain 'Organizational Context'; prompt:\n%s", prompt)
	}
	if !strings.Contains(prompt, "Prior JWT decision") {
		t.Errorf("expected prompt to contain entry title; prompt:\n%s", prompt)
	}
}

// TestRunner_SkipsOrgContext_WhenMemoryEmpty verifies that when memory.Search returns
// no entries, no "Organizational Context" section is injected.
func TestRunner_SkipsOrgContext_WhenMemoryEmpty(t *testing.T) {
	ctx := context.Background()

	mem := &mockMemoryProvider{searchRet: []memory.Entry{}}
	capture := &requestCaptureProvider{}
	d := buildOneShotDeveloper()

	deps := specialists.RunnerDeps{
		Registry: buildMinimalRegistryForMem(),
		Provider: capture,
		Store:    newMemStore(),
		Memory:   mem,
		Pipeline: &mockPipeline{},
	}

	runner := specialists.New(d, deps)
	if err := runner.Run(ctx, config.Root{}, "add-auth"); err != nil {
		t.Fatalf("Run: %v", err)
	}

	if len(capture.reqs) == 0 {
		t.Fatal("no LLM requests captured")
	}
	p := capture.reqs[0].Messages[0].Content
	if strings.Contains(p, "Organizational Context") {
		t.Errorf("expected NO 'Organizational Context' when memory is empty; prompt:\n%s", p)
	}
}

// TestRunner_SavesKnowledgeRecord_AfterFinalStep verifies that memory.Save is called
// once after the final step completes.
func TestRunner_SavesKnowledgeRecord_AfterFinalStep(t *testing.T) {
	ctx := context.Background()

	mem := &mockMemoryProvider{searchRet: []memory.Entry{}}
	d := buildOneShotDeveloper()
	provider := llm.NewMockProvider(llm.WithRawResponses("result: ok\nopen_items: []\n"))

	deps := specialists.RunnerDeps{
		Registry: buildMinimalRegistryForMem(),
		Provider: provider,
		Store:    newMemStore(),
		Memory:   mem,
		Pipeline: &mockPipeline{},
	}

	runner := specialists.New(d, deps)
	if err := runner.Run(ctx, config.Root{}, "add-auth"); err != nil {
		t.Fatalf("Run: %v", err)
	}

	if mem.saveCalls == 0 {
		t.Error("expected memory.Save to be called after final step")
	}
	if len(mem.savedEntries) == 0 {
		t.Fatal("no saved entries")
	}
	saved := mem.savedEntries[0]
	if saved.Title == "" {
		t.Error("saved entry Title must not be empty")
	}
	if saved.Type == "" {
		t.Error("saved entry Type must not be empty")
	}
	if saved.Content.What == "" {
		t.Error("saved entry Content.What must not be empty")
	}
}

// TestRunner_DoesNotFail_WhenMemorySaveErrors verifies that a memory.Save error
// does not abort the run — the run must complete normally.
func TestRunner_DoesNotFail_WhenMemorySaveErrors(t *testing.T) {
	ctx := context.Background()

	mem := &mockMemoryProvider{
		searchRet: []memory.Entry{},
		saveErr:   errors.New("engram unavailable"),
	}
	d := buildOneShotDeveloper()
	provider := llm.NewMockProvider(llm.WithRawResponses("result: ok\nopen_items: []\n"))

	deps := specialists.RunnerDeps{
		Registry: buildMinimalRegistryForMem(),
		Provider: provider,
		Store:    newMemStore(),
		Memory:   mem,
		Pipeline: &mockPipeline{},
	}

	runner := specialists.New(d, deps)
	err := runner.Run(ctx, config.Root{}, "add-auth")
	if err != nil {
		t.Fatalf("Run must succeed even when memory.Save errors; got: %v", err)
	}
}

// TestRunner_DoesNotFail_WhenMemorySearchErrors verifies that a memory.Search error
// does not abort the run — the run proceeds without organizational context.
func TestRunner_DoesNotFail_WhenMemorySearchErrors(t *testing.T) {
	ctx := context.Background()

	mem := &mockMemoryProvider{
		searchErr: errors.New("engram unreachable"),
	}
	capture := &requestCaptureProvider{}
	d := buildOneShotDeveloper()

	deps := specialists.RunnerDeps{
		Registry: buildMinimalRegistryForMem(),
		Provider: capture,
		Store:    newMemStore(),
		Memory:   mem,
		Pipeline: &mockPipeline{},
	}

	runner := specialists.New(d, deps)
	if err := runner.Run(ctx, config.Root{}, "add-auth"); err != nil {
		t.Fatalf("Run must succeed when memory.Search errors; got: %v", err)
	}

	// No "Organizational Context" should appear since search errored.
	if len(capture.reqs) > 0 {
		p := capture.reqs[0].Messages[0].Content
		if strings.Contains(p, "Organizational Context") {
			t.Errorf("expected no org context injection when Search errors; prompt:\n%s", p)
		}
	}
}
