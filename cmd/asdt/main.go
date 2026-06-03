// Package main is the composition root for the asdt binary.
// All dependency wiring happens here; no business logic lives here.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/vitualizz/ai-software-delivery-team/internal/artifact"
	"github.com/vitualizz/ai-software-delivery-team/internal/config"
	"github.com/vitualizz/ai-software-delivery-team/internal/developer"
	"github.com/vitualizz/ai-software-delivery-team/internal/knowledge"
	"github.com/vitualizz/ai-software-delivery-team/internal/llm"
	"github.com/vitualizz/ai-software-delivery-team/internal/pipeline"
	"github.com/vitualizz/ai-software-delivery-team/internal/prompt"
	"github.com/vitualizz/ai-software-delivery-team/internal/requirements"
)

func main() {
	if err := run(context.Background(), os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return printHelp()
	}
	subcmd := args[0]
	rest := args[1:]

	// help does not require .asdt/ discovery.
	if subcmd == "help" {
		return printHelp()
	}

	// All other commands require a project root.
	root, err := config.Discover(".")
	if err != nil {
		return fmt.Errorf("no .asdt/ found: run from inside a project or create .asdt/ to initialize: %w", err)
	}

	cfg, err := config.Load(root)
	if err != nil {
		cfg = config.Config{} // use defaults on load failure
	}

	// Build shared dependencies.
	store := artifact.NewFSStore(root.Path())
	reg := prompt.DefaultEmbeddedRegistry()
	runner := pipeline.NewFSMachine(store)
	knowledgeReader := knowledge.NewFSReader()

	switch subcmd {
	case "knowledge":
		return runKnowledge(ctx, root, knowledgeReader)

	case "requirements":
		if len(rest) == 0 {
			return fmt.Errorf("usage: asdt requirements <idea>")
		}
		idea := strings.Join(rest, " ")
		change := resolveChange(rest, cfg)
		provider := buildProvider()
		agent := requirements.New(reg, provider, store, runner, knowledgeReader)
		return agent.Run(ctx, root, change, idea)

	case "develop":
		change := resolveChange(rest, cfg)
		provider := buildProvider()
		agent := developer.New(reg, provider, store, runner, knowledgeReader)
		return agent.Run(ctx, root, change)

	case "status":
		return printStatus(root, cfg, store, runner, ctx)

	default:
		return fmt.Errorf("unknown subcommand: %q. Valid: knowledge, requirements, develop, status, help", subcmd)
	}
}

// runKnowledge scans the project root and writes platform.yaml.
func runKnowledge(ctx context.Context, root config.Root, reader knowledge.Reader) error {
	det := knowledge.DefaultDetector()
	// The project root is one level up from the .asdt/ directory.
	projectRoot := strings.TrimSuffix(root.Path(), "/.asdt")
	p, err := det.Detect(ctx, projectRoot)
	if err != nil {
		return fmt.Errorf("knowledge detect: %w", err)
	}
	if err := reader.Write(root, p); err != nil {
		return fmt.Errorf("knowledge write: %w", err)
	}
	fmt.Printf("knowledge: detected stack %v\n", p.DetectedStack)
	return nil
}

// resolveChange parses --change <name> from args, falls back to cfg.ActiveChange,
// then falls back to "default".
func resolveChange(args []string, cfg config.Config) string {
	for i := 0; i < len(args)-1; i++ {
		if args[i] == "--change" {
			return args[i+1]
		}
	}
	if cfg.ActiveChange != "" {
		return cfg.ActiveChange
	}
	return "default"
}

// buildProvider returns a real LLM provider when ANTHROPIC_API_KEY is set,
// otherwise returns a MockProvider with a stub response for local development.
func buildProvider() llm.Provider {
	if os.Getenv("ANTHROPIC_API_KEY") != "" {
		log.Println("Anthropic provider not yet implemented, using mock")
	}
	return llm.NewMockProvider(llm.WithScriptedResponses(
		"mock LLM response — set ANTHROPIC_API_KEY and implement AnthropicProvider for production use",
	))
}

// printHelp prints CLI usage to stdout and returns nil.
func printHelp() error {
	fmt.Print(`asdt — AI Software Delivery Team

Usage:
  asdt <subcommand> [args]

Subcommands:
  knowledge               Scan project and write .asdt/knowledge/platform.yaml
  requirements <idea>     Generate requirements-spec from a free-text idea
  develop                 Generate implementation-plan from requirements-spec
  status                  Show current pipeline state
  help                    Print this help message

Flags:
  --change <name>         Use a named change (default: active_change from config, then "default")

Examples:
  asdt knowledge
  asdt requirements "add user password reset for users who forgot credentials"
  asdt develop
  asdt status
`)
	return nil
}

// printStatus reads the pipeline state and prints it to stdout.
func printStatus(root config.Root, cfg config.Config, store artifact.Store, runner pipeline.PipelineRunner, ctx context.Context) error {
	change := cfg.ActiveChange
	if change == "" {
		change = "default"
	}
	state, err := runner.Current(ctx, change)
	if err != nil {
		return fmt.Errorf("status: read pipeline: %w", err)
	}
	if state.ChangeID == "" {
		fmt.Printf("No pipeline state found for change %q. Run `asdt requirements` to start.\n", change)
		return nil
	}
	fmt.Printf("Change:       %s\n", state.ChangeID)
	fmt.Printf("Current:      %s\n", state.CurrentState)
	fmt.Printf("Transitions:  %d\n", len(state.Transitions))
	for _, t := range state.Transitions {
		fmt.Printf("  %s → %s  (%s)\n", t.From, t.To, t.Timestamp.Format("2006-01-02 15:04:05"))
	}
	_ = root // root used for discovery context
	return nil
}
