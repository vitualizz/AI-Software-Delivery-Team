// Package requirements implements the requirements-gathering agent.
// It elicits user stories, acceptance criteria, and scope from the user's idea.
package requirements

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

// UserStory is one story entry in the requirements spec.
type UserStory struct {
	ID     string `yaml:"id"`
	As     string `yaml:"as"`
	Want   string `yaml:"want"`
	SoThat string `yaml:"so_that"`
}

// RequirementsSpec is the payload written to requirements-spec.yaml.
type RequirementsSpec struct {
	UserStories        []UserStory         `yaml:"user_stories"`
	AcceptanceCriteria map[string][]string `yaml:"acceptance_criteria"`
	Scope              struct {
		In  []string `yaml:"in"`
		Out []string `yaml:"out"`
	} `yaml:"scope"`
	NFRs          []string `yaml:"nfrs"`
	OpenQuestions []string `yaml:"open_questions"`
}

// Agent runs the requirements-gathering workflow.
type Agent struct {
	registry prompt.SkillRegistry
	provider llm.Provider
	store    artifact.Store
	pipeline pipeline.PipelineRunner
	reader   *knowledge.FSReader
}

// New constructs a requirements Agent with the provided dependencies.
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

// Run elicits a RequirementsSpec from the LLM and writes requirements-spec.yaml.
func (a *Agent) Run(ctx context.Context, root config.Root, change, idea string) error {
	roleFragment, err := a.registry.Role("requirements")
	if err != nil {
		roleFragment = prompt.Fragment{Name: "requirements", Content: "You are a requirements analyst.", Version: "builtin"}
	}

	composed := fmt.Sprintf("%s\n\nIdea: %s\n\nProduce YAML output as a requirements spec.", roleFragment.Content, idea)

	resp, err := a.provider.Complete(ctx, llm.Request{
		Messages: []llm.Message{{Role: "user", Content: composed}},
	})
	if err != nil {
		return fmt.Errorf("requirements LLM: %w", err)
	}

	cleaned := stripMarkdownCodeBlock(resp.Content)
	var spec RequirementsSpec
	if err := yaml.Unmarshal([]byte(cleaned), &spec); err != nil {
		return fmt.Errorf("requirements unmarshal: %w", err)
	}

	env := artifact.Envelope[RequirementsSpec]{
		EnvelopeHeader: artifact.EnvelopeHeader{
			SchemaVersion: artifact.CurrentSchemaVersion,
			Agent:         "requirements",
			ChangeID:      change,
			CreatedAt:     time.Now().UTC(),
			PromptVersion: roleFragment.Version,
			InputRefs:     []string{},
		},
		Payload: spec,
	}

	if err := a.store.Write(ctx, change, "requirements-spec", env); err != nil {
		return fmt.Errorf("requirements write: %w", err)
	}

	// Write v1 pipeline state so FSMachine.Current returns PhaseRequirements.
	v1State := pipeline.State{
		SchemaVersion: "1",
		ChangeID:      change,
		CurrentState:  pipeline.PhaseRequirements,
		Transitions: []pipeline.Transition{
			{
				From:      "",
				To:        pipeline.PhaseRequirements,
				Timestamp: time.Now().UTC(),
			},
		},
	}
	if err := a.store.Write(ctx, change, "pipeline-state", v1State); err != nil {
		return fmt.Errorf("requirements pipeline-state write: %w", err)
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
