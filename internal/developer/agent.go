// Package developer implements the Developer Agent runner.
// It reads the requirements-spec artifact + platform context, composes a prompt,
// calls the LLM, parses the YAML response into an ImplementationPlan envelope,
// validates story traceability, writes the artifact, and advances the pipeline.
package developer

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/vitualizz/ai-software-delivery-team/internal/artifact"
	"github.com/vitualizz/ai-software-delivery-team/internal/config"
	"github.com/vitualizz/ai-software-delivery-team/internal/knowledge"
	"github.com/vitualizz/ai-software-delivery-team/internal/llm"
	"github.com/vitualizz/ai-software-delivery-team/internal/pipeline"
	"github.com/vitualizz/ai-software-delivery-team/internal/prompt"
	"github.com/vitualizz/ai-software-delivery-team/internal/requirements"
	"gopkg.in/yaml.v3"
)

// Agent orchestrates the developer workflow:
// read requirements-spec → compose prompt → LLM call → validate → write artifact → advance pipeline.
type Agent struct {
	registry prompt.SkillRegistry
	provider llm.Provider
	store    artifact.Store
	runner   pipeline.PipelineRunner
	reader   knowledge.Reader
}

// New constructs a Developer Agent with all required dependencies.
func New(
	registry prompt.SkillRegistry,
	provider llm.Provider,
	store artifact.Store,
	runner pipeline.PipelineRunner,
	reader knowledge.Reader,
) *Agent {
	return &Agent{
		registry: registry,
		provider: provider,
		store:    store,
		runner:   runner,
		reader:   reader,
	}
}

// Run orchestrates the full developer workflow for the given change.
//
// Returns ErrMissingRequirements when requirements-spec does not exist.
// Degrades gracefully when platform.yaml is absent (logs a warning, records absence).
func (a *Agent) Run(ctx context.Context, root config.Root, change string) error {
	// Step 1: Read requirements-spec. Hard fail if absent.
	if !a.store.Exists(change, "requirements-spec") {
		return ErrMissingRequirements
	}

	var specEnv artifact.Envelope[requirements.RequirementsSpec]
	if err := a.store.Read(ctx, change, "requirements-spec", &specEnv); err != nil {
		return fmt.Errorf("developer: read requirements-spec: %w", err)
	}

	// Step 2: Validate envelope schema version.
	if err := artifact.Validate(specEnv.EnvelopeHeader); err != nil {
		return fmt.Errorf("developer: requirements-spec validation: %w", err)
	}

	// Step 3: Read platform.yaml — degrade gracefully on error.
	// Use relative .asdt/-relative paths for input_refs (portable, not absolute).
	specRef := fmt.Sprintf(".asdt/artifacts/%s/requirements-spec.yaml", change)
	platformRef := ""
	plat, err := a.reader.Read(root)
	if err != nil {
		log.Printf("developer: platform.yaml not found, proceeding with degraded context: %v", err)
		plat = knowledge.Platform{}
	} else {
		platformRef = ".asdt/knowledge/platform.yaml"
	}

	// Step 4: Load role and skill fragments.
	roleFragment, err := a.registry.Role("developer")
	if err != nil {
		return fmt.Errorf("developer: load role: %w", err)
	}
	codeGenSkill, err := a.registry.Skill("code-generation")
	if err != nil {
		return fmt.Errorf("developer: load skill code-generation: %w", err)
	}
	testSkill, err := a.registry.Skill("test-writing")
	if err != nil {
		return fmt.Errorf("developer: load skill test-writing: %w", err)
	}

	// Step 5: Build artifact context from requirements-spec.
	artifactContext := buildArtifactContext(specEnv.Payload)
	platformContext := buildPlatformContext(plat)

	// Step 6: Compose prompt.
	composedPrompt, manifest := prompt.Compose(
		roleFragment,
		[]prompt.Fragment{codeGenSkill, testSkill},
		artifactContext,
		platformContext,
	)

	// Step 7: Call LLM.
	messages := []llm.Message{
		{Role: "system", Content: composedPrompt},
		{Role: "user", Content: "Produce an implementation plan for the requirements above."},
	}
	resp, err := a.provider.Complete(ctx, llm.Request{Messages: messages})
	if err != nil {
		return fmt.Errorf("developer: llm complete: %w", err)
	}

	// Step 8: Parse LLM YAML response into ImplementationPlan.
	plan, err := parsePlan(resp.Content)
	if err != nil {
		return fmt.Errorf("developer: parse plan: %w", err)
	}

	// Step 9: Validate traceability — every story ID must appear in at least one step.
	plan = validateTraceability(plan, specEnv.Payload)

	// Step 10: Build input_refs.
	inputRefs := []string{specRef}
	if platformRef != "" {
		inputRefs = append(inputRefs, platformRef)
	}

	// Step 11: Build envelope.
	env := artifact.Envelope[ImplementationPlan]{
		EnvelopeHeader: artifact.EnvelopeHeader{
			SchemaVersion: artifact.CurrentSchemaVersion,
			Agent:         "developer",
			ChangeID:      change,
			CreatedAt:     time.Now().UTC(),
			PromptVersion: manifest.Hash(),
			InputRefs:     inputRefs,
		},
		Payload: plan,
	}

	// Step 12: Write artifact.
	if err := a.store.Write(ctx, change, "implementation-plan", env); err != nil {
		return fmt.Errorf("developer: write artifact: %w", err)
	}

	// Step 13: Advance pipeline to PhasePlan.
	if _, err := a.runner.Advance(ctx, change, pipeline.PhasePlan); err != nil {
		return fmt.Errorf("developer: advance pipeline: %w", err)
	}

	return nil
}

// buildArtifactContext serializes the requirements spec into a human-readable
// context string for the LLM prompt.
func buildArtifactContext(spec requirements.RequirementsSpec) string {
	if len(spec.UserStories) == 0 {
		return ""
	}
	var sb strings.Builder
	sb.WriteString("## User Stories\n")
	for _, us := range spec.UserStories {
		sb.WriteString(fmt.Sprintf("- [%s] As a %s, I want to %s, so that %s\n",
			us.ID, us.As, us.Want, us.SoThat))
	}
	return sb.String()
}

// buildPlatformContext serializes the platform info as a context string.
func buildPlatformContext(p knowledge.Platform) string {
	if len(p.DetectedStack) == 0 {
		return ""
	}
	return fmt.Sprintf("Detected stack: %s", strings.Join(p.DetectedStack, ", "))
}

// parsePlan unmarshals a YAML string from the LLM into an ImplementationPlan.
func parsePlan(content string) (ImplementationPlan, error) {
	cleaned := stripMarkdownCodeBlock(content)
	var plan ImplementationPlan
	if err := yaml.Unmarshal([]byte(cleaned), &plan); err != nil {
		return ImplementationPlan{}, fmt.Errorf("unmarshal implementation plan: %w", err)
	}
	return plan, nil
}

// stripMarkdownCodeBlock removes ```yaml ... ``` fences if present.
func stripMarkdownCodeBlock(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "```") {
		idx := strings.Index(s, "\n")
		if idx == -1 {
			return s
		}
		s = s[idx+1:]
		if end := strings.LastIndex(s, "```"); end != -1 {
			s = s[:end]
		}
	}
	return strings.TrimSpace(s)
}

// validateTraceability checks that every story ID from the spec appears in at
// least one plan step. Missing references are collected in plan.OpenItems.
func validateTraceability(plan ImplementationPlan, spec requirements.RequirementsSpec) ImplementationPlan {
	referenced := make(map[string]bool)
	for _, step := range plan.Steps {
		referenced[step.StoryID] = true
	}

	for _, story := range spec.UserStories {
		if !referenced[story.ID] {
			plan.OpenItems = append(plan.OpenItems,
				fmt.Sprintf("story %s not referenced in any implementation step", story.ID))
		}
	}
	return plan
}
