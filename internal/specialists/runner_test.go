package specialists_test

import (
	"context"
	"errors"
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
		"developer/SKILL.md":                   {Data: []byte("Developer role content.")},
		"security/SKILL.md":                    {Data: []byte("Security role content.")},
		"developer/skills/artifact-loading.md": {Data: []byte("artifact loading skill")},
		"developer/skills/code-generation.md":  {Data: []byte("code generation skill")},
		"developer/skills/test-generation.md":  {Data: []byte("test generation skill")},
		"security/skills/threat-modeling.md":   {Data: []byte("threat modeling skill")},
		"security/skills/owasp-review.md":      {Data: []byte("owasp skill")},
		"_shared/skills/platform-context.md":   {Data: []byte("platform context skill")},
		"_shared/skills/artifact-envelope.md":  {Data: []byte("artifact envelope skill")},
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
		Memory:   memory.NullProvider{},
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
		Memory:   memory.NullProvider{},
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
		Memory:   memory.NullProvider{},
		Pipeline: pl,
	}

	d := specialists.DeveloperDescriptor()
	runner := specialists.New(d, deps)
	err := runner.Run(ctx, config.Root{}, "err-change")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	// Error must contain the step ID of step 2 ("platform-context").
	if !strings.Contains(err.Error(), "platform-context") {
		t.Errorf("error %q should mention step ID 'platform-context'", err.Error())
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
		Memory:   memory.NullProvider{},
		Pipeline: pl,
	}

	runner := specialists.New(d, deps)
	if err := runner.Run(ctx, config.Root{}, "null-mem-change"); err != nil {
		t.Fatalf("Run with NullProvider memory: %v", err)
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
