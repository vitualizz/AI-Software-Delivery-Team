// Package main is the composition root for the asdt binary.
// All dependency wiring happens here; no business logic lives here.
// The binary is a package manager: it validates config, installs skills/providers,
// and runs deterministic project analysis. Specialist execution lives in SKILL.md
// files — it is never invoked by this binary.
package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/vitualizz/ai-software-delivery-team/internal/artifact"
	"github.com/vitualizz/ai-software-delivery-team/internal/config"
	"github.com/vitualizz/ai-software-delivery-team/internal/knowledge"
	"github.com/vitualizz/ai-software-delivery-team/internal/pipeline"
)

// deprecated maps removed task-verb commands to guidance messages.
// These commands existed in the MVP pipeline model and have been superseded
// by the specialist model.
var deprecated = map[string]string{
	"requirements": "use /asdt:architect (for system requirements) or /asdt:ux-ui (for product requirements)",
	"develop":      "use /asdt:developer",
	"knowledge":    "platform context is now loaded automatically by each specialist's Platform Analysis step",
}

// validSpecialists lists the specialist IDs supported by this package manager.
// Specialist execution is handled by SKILL.md files in the AI assistant runtime —
// not by this binary. Use these IDs to invoke a specialist in your AI assistant:
// e.g. /asdt:developer, /asdt:architect, etc.
var validSpecialists = []string{
	"architect",
	"developer",
	"qa",
	"security",
	"ux-ui",
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

	// help does not require .asdt/ discovery.
	if subcmd == "help" {
		return printHelp()
	}

	// Deprecated commands: print guidance and exit 0.
	if msg, dep := deprecated[subcmd]; dep {
		fmt.Printf("warning: `asdt %s` is deprecated: %s\n", subcmd, msg)
		return nil
	}

	// Specialist commands: the binary no longer runs specialists.
	// Redirect to the AI assistant runtime.
	if isSpecialist(subcmd) {
		fmt.Printf("Run specialist %q via your AI assistant: /%s:%s\n", subcmd, "asdt", subcmd)
		fmt.Printf("The %q specialist is implemented in skill/%s/SKILL.md.\n", subcmd, subcmd)
		fmt.Printf("It requires a configured memory provider — set memory.provider in .asdt/config.yaml.\n")
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

	return fmt.Errorf("unknown command %q. Run `asdt help` for usage", subcmd)
}

// isSpecialist returns true when the given ID is a known specialist.
func isSpecialist(id string) bool {
	for _, s := range validSpecialists {
		if s == id {
			return true
		}
	}
	return false
}

// listSpecialists returns a sorted comma-separated list of specialist IDs.
func listSpecialists() string {
	ids := make([]string, len(validSpecialists))
	copy(ids, validSpecialists)
	sort.Strings(ids)
	return strings.Join(ids, ", ")
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
	fmt.Printf(`asdt — AI Software Delivery Team

Usage:
  asdt <command> [args]

Package manager commands:
  init                    Scan project and write platform-summary.yaml (deterministic, zero LLM tokens)
  status                  Show current pipeline state
  help                    Print this help message

Specialist invocation (via AI assistant — not this binary):
  %s

  To run a specialist, use your AI assistant:
    /asdt:developer --change add-auth
    /asdt:architect
    /asdt:security

  Specialists require a configured memory provider in .asdt/config.yaml.

Deprecated commands (use the specialist equivalents above):
  requirements            Replaced by: use /asdt:architect or /asdt:ux-ui
  develop                 Replaced by: /asdt:developer
  knowledge               Platform context is now auto-loaded by each specialist
`, "  "+listSpecialists())
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
