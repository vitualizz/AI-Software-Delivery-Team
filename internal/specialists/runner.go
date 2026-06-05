package specialists

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/vitualizz/ai-software-delivery-team/internal/artifact"
	"github.com/vitualizz/ai-software-delivery-team/internal/config"
	"github.com/vitualizz/ai-software-delivery-team/internal/llm"
	"github.com/vitualizz/ai-software-delivery-team/internal/memory"
	"github.com/vitualizz/ai-software-delivery-team/internal/pipeline"
	"github.com/vitualizz/ai-software-delivery-team/internal/prompt"
	"gopkg.in/yaml.v3"
)

// ComposeFunc is the function signature for composing a prompt from fragments.
// It matches prompt.Compose so tests can inject a no-op alternative.
type ComposeFunc func(role prompt.Fragment, skills []prompt.Fragment, artifactCtx, platformCtx string) (string, prompt.Manifest)

// RunnerDeps holds all the dependencies required by a Runner.
// All fields are required except Memory (defaults to NullProvider) and Composer
// (defaults to prompt.Compose).
type RunnerDeps struct {
	Registry prompt.SkillRegistry
	Composer ComposeFunc
	Provider llm.Provider
	Store    artifact.Store
	Memory   memory.Provider
	Pipeline pipeline.PipelineRunner
}

// Runner executes any SpecialistDescriptor using the provided dependencies.
// It is the single generic execution engine; no per-specialist logic lives here.
type Runner struct {
	descriptor SpecialistDescriptor
	deps       RunnerDeps
}

// New constructs a Runner for the given descriptor.
// If deps.Memory is nil, NullProvider is used.
// If deps.Composer is nil, prompt.Compose is used.
func New(d SpecialistDescriptor, deps RunnerDeps) *Runner {
	if deps.Memory == nil {
		deps.Memory = &memory.NullProvider{}
	}
	if deps.Composer == nil {
		deps.Composer = prompt.Compose
	}
	return &Runner{descriptor: d, deps: deps}
}

// Run validates the descriptor and executes each workflow step in order.
// It returns the first step error, wrapped with the step ID.
func (r *Runner) Run(ctx context.Context, root config.Root, change string) error {
	if err := r.descriptor.Validate(); err != nil {
		return fmt.Errorf("specialist %s: invalid descriptor: %w", r.descriptor.ID, err)
	}

	for _, step := range r.descriptor.Workflow {
		if err := r.runStep(ctx, root, change, step); err != nil {
			return fmt.Errorf("step %s: %w", step.ID, err)
		}
	}
	return nil
}

// runStep executes a single workflow step.
func (r *Runner) runStep(ctx context.Context, root config.Root, change string, step WorkflowStep) error {
	// SkipIfInitialized gate: when the flag is set and platform-summary.yaml exists,
	// emit the summary as the step's output artifact without invoking the LLM.
	if step.SkipIfInitialized && step.OutputArtifact != "" {
		summaryFile := filepath.Join(root.Path(), "knowledge", "platform-summary.yaml")
		if data, err := os.ReadFile(summaryFile); err == nil {
			var payload map[string]any
			if unmarshalErr := yaml.Unmarshal(data, &payload); unmarshalErr == nil && payload != nil {
				payload["source"] = "platform-summary.yaml" // provenance marker
				payload["open_items"] = []any{}             // satisfy downstream expectations
				role, rerr := r.deps.Registry.Role(r.descriptor.ID)
				if rerr == nil {
					if werr := r.writeStepArtifact(ctx, root, change, step.OutputArtifact, payload, role, nil); werr != nil {
						return werr
					}
					return r.deps.Pipeline.AdvanceStep(ctx, root, change, r.descriptor.ID, step.ID)
				}
			}
			// malformed YAML or role lookup failed — fall through to normal LLM path
		}
		// summary file absent — fall through to normal LLM path
	}

	sharedFragments := r.loadSharedSkills()
	stepFragments := r.loadStepSkills(step)
	artifactCtx, openItems := r.loadStepInputs(ctx, root, change, step)
	platformCtx := r.loadPlatformContext(root)

	roleFragment, err := r.deps.Registry.Role(r.descriptor.ID)
	if err != nil {
		return err
	}

	allSkills := append(sharedFragments, stepFragments...) //nolint:gocritic
	composed, manifest := r.deps.Composer(roleFragment, allSkills, artifactCtx, platformCtx)
	composed += fmt.Sprintf("\n\n## Current Step: %s\n%s\n\nProduce YAML output for this step only.", step.ID, step.Description)

	response, err := r.deps.Provider.Complete(ctx, llm.Request{
		Messages: []llm.Message{{Role: "user", Content: composed}},
	})
	if err != nil {
		return err
	}

	payload := r.parseYAMLResponse(response.Content, openItems)

	if step.OutputArtifact != "" {
		if err := r.writeStepArtifact(ctx, root, change, step.OutputArtifact, payload, roleFragment, allSkills); err != nil {
			return err
		}
	} else if r.isLastStep(step) {
		if err := r.writeArtifacts(ctx, root, change, payload, roleFragment, allSkills, manifest); err != nil {
			return err
		}
	}

	if err := r.deps.Pipeline.AdvanceStep(ctx, root, change, r.descriptor.ID, step.ID); err != nil {
		return err
	}

	return nil
}

// loadSharedSkills resolves all shared skills declared on the descriptor.
// Skills that fail to resolve are silently dropped (degrade gracefully).
func (r *Runner) loadSharedSkills() []prompt.Fragment {
	frags := make([]prompt.Fragment, 0, len(r.descriptor.Skills))
	for _, name := range r.descriptor.Skills {
		frag, err := r.deps.Registry.Skill(name)
		if err != nil {
			continue // degrade: shared skill missing → skip, do not abort
		}
		frags = append(frags, frag)
	}
	return frags
}

// loadStepSkills resolves the specialist-scoped skills for a single step.
// Skills that fail to resolve are silently dropped.
func (r *Runner) loadStepSkills(step WorkflowStep) []prompt.Fragment {
	frags := make([]prompt.Fragment, 0, len(step.SkillRefs))
	for _, name := range step.SkillRefs {
		frag, err := r.deps.Registry.ScopedSkill(r.descriptor.ID, name)
		if err != nil {
			continue // degrade: step skill missing → skip, do not abort
		}
		frags = append(frags, frag)
	}
	return frags
}

// loadStepInputs loads only the artifact types declared in step.InputRefs.
// When InputRefs is empty, falls back to descriptor.Artifacts.Reads (backward compat).
// Missing inputs produce open_items notes but do not return an error.
func (r *Runner) loadStepInputs(ctx context.Context, root config.Root, change string, step WorkflowStep) (string, []string) {
	refs := step.InputRefs
	if len(refs) == 0 {
		refs = r.descriptor.Artifacts.Reads
	}
	var parts []string
	var missing []string
	for _, artifactType := range refs {
		if !r.deps.Store.Exists(change, artifactType) {
			missing = append(missing, fmt.Sprintf("%s.yaml absent — proceeding without it", artifactType))
			continue
		}
		var raw map[string]any
		if err := r.deps.Store.Read(ctx, change, artifactType, &raw); err != nil {
			missing = append(missing, fmt.Sprintf("%s.yaml unreadable: %v", artifactType, err))
			continue
		}
		data, _ := yaml.Marshal(raw)
		parts = append(parts, fmt.Sprintf("## Artifact: %s\n\n```yaml\n%s```", artifactType, string(data)))
	}
	_ = root // accepted for interface symmetry; currently unused
	return strings.Join(parts, "\n\n"), missing
}

// loadInputArtifacts pre-loads all soft-required artifact types declared in
// ArtifactContract.Reads. Missing inputs produce open_items notes but do not
// return an error.
//
// Deprecated: use loadStepInputs instead. Kept for backward compatibility.
func (r *Runner) loadInputArtifacts(ctx context.Context, root config.Root, change string) (context string, missing []string) {
	var parts []string
	for _, artifactType := range r.descriptor.Artifacts.Reads {
		if !r.deps.Store.Exists(change, artifactType) {
			missing = append(missing, fmt.Sprintf("%s.yaml absent — proceeding without it", artifactType))
			continue
		}
		var raw map[string]any
		if err := r.deps.Store.Read(ctx, change, artifactType, &raw); err != nil {
			missing = append(missing, fmt.Sprintf("%s.yaml unreadable: %v", artifactType, err))
			continue
		}
		data, _ := yaml.Marshal(raw)
		parts = append(parts, fmt.Sprintf("## Artifact: %s\n\n```yaml\n%s```", artifactType, string(data)))
	}
	_ = root // root is accepted for interface symmetry; currently unused
	return strings.Join(parts, "\n\n"), missing
}

// loadPlatformContext returns platform context for prompt injection.
// Prefers the deterministic project-level summary (platform-summary.yaml),
// falling back to raw platform.yaml, then to empty string.
func (r *Runner) loadPlatformContext(root config.Root) string {
	base := filepath.Join(root.Path(), "knowledge")
	// Prefer the deterministic project-level summary.
	if data, err := os.ReadFile(filepath.Join(base, "platform-summary.yaml")); err == nil {
		return fmt.Sprintf("## Platform Context (summary)\n\n```yaml\n%s```", string(data))
	}
	// Fall back to raw platform.yaml.
	if data, err := os.ReadFile(filepath.Join(base, "platform.yaml")); err == nil {
		return fmt.Sprintf("## Platform Context\n\n```yaml\n%s```", string(data))
	}
	return "" // both absent — degrade gracefully (unchanged contract)
}

// parseYAMLResponse unmarshals the LLM YAML response into a generic map.
// On parse failure, returns a map with the raw response and open_items populated.
func (r *Runner) parseYAMLResponse(response string, openItems []string) map[string]any {
	cleaned := stripMarkdownCodeBlock(response)
	var payload map[string]any
	if err := yaml.Unmarshal([]byte(cleaned), &payload); err != nil || payload == nil {
		payload = map[string]any{
			"raw_response": response,
		}
	}
	return r.injectOpenItems(payload, openItems)
}

// injectOpenItems merges openItems into payload["open_items"].
// Existing open_items entries are preserved; new entries are appended.
func (r *Runner) injectOpenItems(payload map[string]any, openItems []string) map[string]any {
	if len(openItems) == 0 {
		return payload
	}
	existing, _ := payload["open_items"].([]any)
	for _, item := range openItems {
		existing = append(existing, item)
	}
	payload["open_items"] = existing
	return payload
}

// isLastStep returns true when step is the final step in the workflow.
func (r *Runner) isLastStep(step WorkflowStep) bool {
	workflow := r.descriptor.Workflow
	if len(workflow) == 0 {
		return false
	}
	return workflow[len(workflow)-1].ID == step.ID
}

// writeArtifacts writes an Envelope[map[string]any] for each artifact type
// declared in ArtifactContract.Writes.
func (r *Runner) writeArtifacts(
	ctx context.Context,
	root config.Root,
	change string,
	payload map[string]any,
	role prompt.Fragment,
	skills []prompt.Fragment,
	manifest prompt.Manifest,
) error {
	inputRefs := r.buildInputRefs(root, change)
	promptVersion := r.computePromptVersion(role, skills)

	for _, artifactType := range r.descriptor.Artifacts.Writes {
		env := artifact.Envelope[map[string]any]{
			EnvelopeHeader: artifact.EnvelopeHeader{
				SchemaVersion: artifact.CurrentSchemaVersion,
				Agent:         r.descriptor.ID,
				ChangeID:      change,
				CreatedAt:     time.Now().UTC(),
				PromptVersion: promptVersion,
				InputRefs:     inputRefs,
			},
			Payload: payload,
		}
		if err := r.deps.Store.Write(ctx, change, artifactType, env); err != nil {
			return fmt.Errorf("write artifact %s: %w", artifactType, err)
		}
	}
	_ = manifest // manifest is available for callers that need the full version map
	return nil
}

// writeStepArtifact writes a single Envelope[map[string]any] for the given artifactType.
// It is called for every step whose OutputArtifact field is non-empty, enabling
// per-step artifact writes (context isolation). The artifact IS the handoff to the
// next step — not conversation history.
func (r *Runner) writeStepArtifact(
	ctx context.Context,
	root config.Root,
	change, artifactType string,
	payload map[string]any,
	role prompt.Fragment,
	skills []prompt.Fragment,
) error {
	env := artifact.Envelope[map[string]any]{
		EnvelopeHeader: artifact.EnvelopeHeader{
			SchemaVersion: artifact.CurrentSchemaVersion,
			Agent:         r.descriptor.ID,
			ChangeID:      change,
			CreatedAt:     time.Now().UTC(),
			PromptVersion: r.computePromptVersion(role, skills),
			InputRefs:     r.buildInputRefs(root, change),
		},
		Payload: payload,
	}
	if err := r.deps.Store.Write(ctx, change, artifactType, env); err != nil {
		return fmt.Errorf("write artifact %s: %w", artifactType, err)
	}
	return nil
}

// computePromptVersion returns an 8-character SHA-256 hash of the concatenated
// fragment versions. This is the value written to EnvelopeHeader.PromptVersion.
func (r *Runner) computePromptVersion(role prompt.Fragment, skills []prompt.Fragment) string {
	var sb strings.Builder
	sb.WriteString(role.Version)
	for _, s := range skills {
		sb.WriteString(s.Version)
	}
	sum := sha256.Sum256([]byte(sb.String()))
	return fmt.Sprintf("%x", sum[:4]) // 8 hex chars
}

// buildInputRefs returns relative paths to all input artifacts that exist.
func (r *Runner) buildInputRefs(root config.Root, change string) []string {
	var refs []string
	for _, artifactType := range r.descriptor.Artifacts.Reads {
		if r.deps.Store.Exists(change, artifactType) {
			refs = append(refs, fmt.Sprintf(".asdt/artifacts/%s/%s.yaml", change, artifactType))
		}
	}
	platformPath := filepath.Join(root.Path(), "knowledge", "platform.yaml")
	if _, err := os.Stat(platformPath); err == nil {
		refs = append(refs, ".asdt/knowledge/platform.yaml")
	}
	return refs
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
