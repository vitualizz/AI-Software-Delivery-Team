// Package requirements implements the Requirements Agent runner.
// It composes a prompt, calls the LLM, parses the YAML response into a
// RequirementsSpec envelope, writes the artifact, and advances the pipeline.
package requirements

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
	"gopkg.in/yaml.v3"
)

// Agent orchestrates the requirements workflow:
// prompt composition → LLM call → artifact write → pipeline advance.
type Agent struct {
	registry prompt.SkillRegistry
	provider llm.Provider
	store    artifact.Store
	runner   pipeline.PipelineRunner
	reader   knowledge.Reader
}

// New constructs a Requirements Agent with all required dependencies.
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

// Run orchestrates the full requirements workflow for the given change and idea.
//
// Returns ErrMissingIdea when idea is empty.
// Returns ErrAmbiguousIdea when idea has fewer than 5 words.
// Degrades gracefully when platform.yaml is absent (logs a warning, continues).
func (a *Agent) Run(ctx context.Context, root config.Root, change, idea string) error {
	if idea == "" {
		return ErrMissingIdea
	}
	if len(strings.Fields(idea)) < 5 {
		return ErrAmbiguousIdea{
			Question: "Could you describe who needs this feature and what outcome they expect?",
		}
	}

	// Load platform context — degrade gracefully on error.
	plat, err := a.reader.Read(root)
	if err != nil {
		log.Printf("requirements: platform.yaml not found, proceeding without project context: %v", err)
		plat = knowledge.Platform{}
	}

	// Load role fragment.
	roleFragment, err := a.registry.Role("requirements")
	if err != nil {
		return fmt.Errorf("requirements: load role: %w", err)
	}

	// Load skill fragments.
	userStorySkill, err := a.registry.Skill("user-story-writing")
	if err != nil {
		return fmt.Errorf("requirements: load skill user-story-writing: %w", err)
	}
	scopeSkill, err := a.registry.Skill("scope-definition")
	if err != nil {
		return fmt.Errorf("requirements: load skill scope-definition: %w", err)
	}

	platformContext := buildPlatformContext(plat)

	// Compose prompt using the Compose function from the prompt package.
	composedPrompt, manifest := prompt.Compose(
		roleFragment,
		[]prompt.Fragment{userStorySkill, scopeSkill},
		"",
		platformContext,
	)

	// Call LLM.
	messages := []llm.Message{
		{Role: "system", Content: composedPrompt},
		{Role: "user", Content: idea},
	}
	resp, err := a.provider.Complete(ctx, llm.Request{Messages: messages})
	if err != nil {
		return fmt.Errorf("requirements: llm complete: %w", err)
	}

	// Parse LLM YAML response into RequirementsSpec.
	spec, err := parseSpec(resp.Content)
	if err != nil {
		return fmt.Errorf("requirements: parse spec: %w", err)
	}

	// Build the envelope.
	env := artifact.Envelope[RequirementsSpec]{
		EnvelopeHeader: artifact.EnvelopeHeader{
			SchemaVersion: artifact.CurrentSchemaVersion,
			Agent:         "requirements",
			ChangeID:      change,
			CreatedAt:     time.Now().UTC(),
			PromptVersion: manifest.Hash(),
			InputRefs:     []string{},
		},
		Payload: spec,
	}

	// Write artifact.
	if err := a.store.Write(ctx, change, "requirements-spec", env); err != nil {
		return fmt.Errorf("requirements: write artifact: %w", err)
	}

	// Advance pipeline to PhaseRequirements (initial state creation).
	if _, err := a.runner.Advance(ctx, change, pipeline.PhaseRequirements); err != nil {
		return fmt.Errorf("requirements: advance pipeline: %w", err)
	}

	return nil
}

// buildPlatformContext serializes the detected platform info as a human-readable
// context string for the LLM.
func buildPlatformContext(p knowledge.Platform) string {
	if len(p.DetectedStack) == 0 {
		return ""
	}
	return fmt.Sprintf("Detected stack: %s", strings.Join(p.DetectedStack, ", "))
}

// parseSpec unmarshals a YAML string from the LLM into a RequirementsSpec.
// The LLM may wrap the YAML in a markdown code block; this is stripped first.
func parseSpec(content string) (RequirementsSpec, error) {
	cleaned := stripMarkdownCodeBlock(content)
	var spec RequirementsSpec
	if err := yaml.Unmarshal([]byte(cleaned), &spec); err != nil {
		return RequirementsSpec{}, fmt.Errorf("unmarshal requirements spec: %w", err)
	}
	return spec, nil
}

// stripMarkdownCodeBlock removes ```yaml ... ``` fences if present.
func stripMarkdownCodeBlock(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "```") {
		// Find the end of the opening fence line.
		idx := strings.Index(s, "\n")
		if idx == -1 {
			return s
		}
		s = s[idx+1:]
		// Strip trailing ```.
		if end := strings.LastIndex(s, "```"); end != -1 {
			s = s[:end]
		}
	}
	return strings.TrimSpace(s)
}
