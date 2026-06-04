// Package developer implements the developer agent.
// It reads a requirements spec and produces an implementation plan.
package developer

import (
	"context"
	"fmt"
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

// ImplementationStep is one ordered step in the implementation plan.
type ImplementationStep struct {
	StoryID        string       `yaml:"story_id"`
	FilesToCreate  []string     `yaml:"files_to_create"`
	FilesToModify  []string     `yaml:"files_to_modify"`
	Rationale      string       `yaml:"rationale"`
	CodeSnippets   []CodeSnippet `yaml:"code_snippets"`
}

// CodeSnippet is a code fragment attached to a step.
type CodeSnippet struct {
	File    string `yaml:"file"`
	Content string `yaml:"content"`
}

// ImplementationPlan is the payload written to implementation-plan.yaml.
type ImplementationPlan struct {
	Approach          string               `yaml:"approach"`
	Steps             []ImplementationStep `yaml:"steps"`
	ComplexityEstimate string              `yaml:"complexity_estimate"`
	OpenItems         []string             `yaml:"open_items"`
}

// Agent runs the developer workflow.
type Agent struct {
	registry prompt.SkillRegistry
	provider llm.Provider
	store    artifact.Store
	pipeline pipeline.PipelineRunner
	reader   *knowledge.FSReader
}

// New constructs a developer Agent with the provided dependencies.
func New(
	registry prompt.SkillRegistry,
	provider llm.Provider,
	store artifact.Store,
	machine pipeline.PipelineRunner,
	reader *knowledge.FSReader,
) *Agent {
	return &Agent{
		registry: registry,
		provider: provider,
		store:    store,
		pipeline: machine,
		reader:   reader,
	}
}

// Run reads the requirements spec, calls the LLM, and writes implementation-plan.yaml.
func (a *Agent) Run(ctx context.Context, root config.Root, change string) error {
	roleFragment, err := a.registry.Role("developer")
	if err != nil {
		roleFragment = prompt.Fragment{Name: "developer", Content: "You are a senior software developer.", Version: "builtin"}
	}

	// Load requirements-spec if present.
	inputRefs := []string{}
	artifactCtx := ""
	specKey := "requirements-spec"
	if a.store.Exists(change, specKey) {
		var raw map[string]any
		if err := a.store.Read(ctx, change, specKey, &raw); err == nil {
			data, _ := yaml.Marshal(raw)
			artifactCtx = fmt.Sprintf("## requirements-spec\n\n```yaml\n%s```", string(data))
			inputRefs = append(inputRefs, fmt.Sprintf(".asdt/artifacts/%s/%s.yaml", change, specKey))
		}
	}

	composed := fmt.Sprintf("%s\n\n%s\n\nProduce YAML output as an implementation plan.", roleFragment.Content, artifactCtx)

	resp, err := a.provider.Complete(ctx, llm.Request{
		Messages: []llm.Message{{Role: "user", Content: composed}},
	})
	if err != nil {
		return fmt.Errorf("developer LLM: %w", err)
	}

	cleaned := stripMarkdownCodeBlock(resp.Content)
	var plan ImplementationPlan
	if err := yaml.Unmarshal([]byte(cleaned), &plan); err != nil {
		return fmt.Errorf("developer unmarshal: %w", err)
	}

	env := artifact.Envelope[ImplementationPlan]{
		EnvelopeHeader: artifact.EnvelopeHeader{
			SchemaVersion: artifact.CurrentSchemaVersion,
			Agent:         "developer",
			ChangeID:      change,
			CreatedAt:     time.Now().UTC(),
			PromptVersion: roleFragment.Version,
			InputRefs:     inputRefs,
		},
		Payload: plan,
	}

	if err := a.store.Write(ctx, change, "implementation-plan", env); err != nil {
		return fmt.Errorf("developer write: %w", err)
	}

	// Write v1 pipeline state so FSMachine.Current returns PhasePlan.
	v1State := pipeline.State{
		SchemaVersion: "1",
		ChangeID:      change,
		CurrentState:  pipeline.PhasePlan,
		Transitions: []pipeline.Transition{
			{
				From:      pipeline.PhaseRequirements,
				To:        pipeline.PhasePlan,
				Timestamp: time.Now().UTC(),
			},
		},
	}
	if err := a.store.Write(ctx, change, "pipeline-state", v1State); err != nil {
		return fmt.Errorf("developer pipeline-state write: %w", err)
	}

	return nil
}

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
