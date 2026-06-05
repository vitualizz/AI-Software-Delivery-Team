// Package main is the composition root for the asdt binary.
// All dependency wiring happens here; no business logic lives here.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/vitualizz/ai-software-delivery-team/internal/artifact"
	"github.com/vitualizz/ai-software-delivery-team/internal/config"
	"github.com/vitualizz/ai-software-delivery-team/internal/knowledge"
	"github.com/vitualizz/ai-software-delivery-team/internal/llm"
	"github.com/vitualizz/ai-software-delivery-team/internal/memory"
	"github.com/vitualizz/ai-software-delivery-team/internal/pipeline"
	"github.com/vitualizz/ai-software-delivery-team/internal/prompt"
	"github.com/vitualizz/ai-software-delivery-team/internal/specialists"
)

// deprecated maps removed task-verb commands to guidance messages.
// These commands existed in the MVP pipeline model and have been superseded
// by the specialist model.
var deprecated = map[string]string{
	"requirements": "use /asdt:architect (for system requirements) or /asdt:ux-ui (for product requirements)",
	"develop":      "use /asdt:developer",
	"knowledge":    "platform context is now loaded automatically by each specialist's Platform Analysis step",
}

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

	// Deprecated commands: print guidance and exit 0.
	if msg, dep := deprecated[subcmd]; dep {
		fmt.Printf("warning: `asdt %s` is deprecated: %s\n", subcmd, msg)
		return nil
	}

	// init command: deterministic project-level platform scan (zero LLM tokens).
	// Handled before Discover so it works even without a pre-existing .asdt/ directory.
	if subcmd == "init" {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("init: get working directory: %w", err)
		}
		asdtDir := filepath.Join(cwd, ".asdt")
		if err := os.MkdirAll(asdtDir, 0o755); err != nil {
			return fmt.Errorf("init: create .asdt/: %w", err)
		}
		root, err := config.Discover(cwd)
		if err != nil {
			return fmt.Errorf("init: discover root: %w", err)
		}
		return runInit(ctx, root)
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
	runner := pipeline.NewFSMachine(store)

	// status command.
	if subcmd == "status" {
		return printStatus(root, cfg, store, runner, ctx)
	}

	// Build the specialist descriptor registry (data-driven dispatch table).
	descriptors := map[string]specialists.SpecialistDescriptor{
		"developer": specialists.DeveloperDescriptor(),
		"ux-ui":     specialists.UXUIDescriptor(),
		"architect": specialists.ArchitectDescriptor(),
		"qa":        specialists.QADescriptor(),
		"security":  specialists.SecurityDescriptor(),
	}

	d, ok := descriptors[subcmd]
	if !ok {
		return fmt.Errorf("unknown specialist %q. Valid: %s", subcmd, validSpecialists(descriptors))
	}

	deps := specialists.RunnerDeps{
		Registry: buildRegistry(root),
		Composer: prompt.Compose,
		Provider: buildProvider(),
		Store:    store,
		Pipeline: runner,
		Memory:   buildMemoryProvider(cfg),
	}
	change := resolveChange(rest, cfg)
	return specialists.New(d, deps).Run(ctx, root, change)
}

// buildRegistry returns an OverrideResolver that checks .asdt/prompts/ for local
// overrides before falling back to the production embedded FS.
func buildRegistry(root config.Root) prompt.SkillRegistry {
	localDir := strings.TrimSuffix(root.Path(), "/.asdt") + "/.asdt/prompts"
	globalDir := prompt.DefaultGlobalDir()
	embedded := prompt.DefaultEmbeddedRegistry()
	return prompt.NewOverrideResolver(localDir, globalDir, embedded)
}

// validSpecialists returns a sorted comma-separated list of specialist IDs.
func validSpecialists(m map[string]specialists.SpecialistDescriptor) string {
	ids := make([]string, 0, len(m))
	for id := range m {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return strings.Join(ids, ", ")
}

// buildMemoryProvider returns the configured memory.Provider.
// Defaults to NullProvider when config is absent or provider is unset.
func buildMemoryProvider(cfg config.Config) memory.Provider {
	switch cfg.Memory.Provider {
	case "engram":
		return memory.NewEngramProvider(cfg.Memory.Endpoint, cfg.Memory.Project)
	default:
		return memory.NullProvider{}
	}
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

// runInit scans the project root, writes platform.yaml, derives a deterministic
// PlatformSummary, and writes platform-summary.yaml. No LLM, no token cost.
func runInit(ctx context.Context, root config.Root) error {
	projectRoot := filepath.Dir(root.Path()) // .asdt/ parent == project root
	detector := knowledge.DefaultDetector()
	platform, err := detector.Detect(ctx, projectRoot)
	if err != nil {
		return fmt.Errorf("init: detect: %w", err)
	}
	reader := knowledge.NewFSReader()
	if err := reader.Write(root, platform); err != nil {
		return fmt.Errorf("init: write platform: %w", err)
	}
	summary := knowledge.DeriveSummary(platform)
	if err := reader.WriteSummary(root, summary); err != nil {
		return fmt.Errorf("init: write summary: %w", err)
	}
	fmt.Printf("Initialized .asdt/knowledge/\n")
	fmt.Printf("  stack:            %s\n", strings.Join(summary.Stack, ", "))
	fmt.Printf("  primary language: %s\n", summary.PrimaryLanguage)
	fmt.Printf("  package manager:  %s\n", summary.PackageManager)
	fmt.Printf("  test runner:      %s\n", summary.TestRunner)
	return nil
}

// printHelp prints CLI usage to stdout and returns nil.
func printHelp() error {
	fmt.Print(`asdt — AI Software Delivery Team

Usage:
  asdt <specialist> [args]

Specialists:
  developer               Generate implementation plan
  ux-ui                   Generate UX/UI spec and component spec
  architect               Generate architecture decision and system design
  qa                      Generate test plan and quality report
  security                Generate threat model and security findings

Other commands:
  init                    Scan project and write platform-summary.yaml (deterministic, zero LLM tokens)
  status                  Show current pipeline state
  help                    Print this help message

Flags:
  --change <name>         Use a named change (default: active_change from config, then "default")

Deprecated commands (use the specialist equivalents above):
  requirements            Replaced by: use /asdt:architect or /asdt:ux-ui
  develop                 Replaced by: /asdt:developer
  knowledge               Platform context is now auto-loaded by each specialist

Examples:
  asdt developer --change add-auth
  asdt security
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
		fmt.Printf("No pipeline state found for change %q. Run a specialist to start.\n", change)
		return nil
	}
	fmt.Printf("Change:       %s\n", state.ChangeID)
	fmt.Printf("Current:      %s\n", state.CurrentState)
	fmt.Printf("Transitions:  %d\n", len(state.Transitions))
	for _, t := range state.Transitions {
		fmt.Printf("  %s → %s  (%s)\n", t.From, t.To, t.Timestamp.Format("2006-01-02 15:04:05"))
	}
	_ = root // root used for discovery context
	_ = store
	return nil
}
