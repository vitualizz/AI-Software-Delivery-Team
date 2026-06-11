package installer

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// executorHeaderPath is the embedded path of the shared executor header that
// every generated agent definition bakes into its body. Its presence in the
// production embedded FS is guarded by skill/embedded_test.go.
const executorHeaderPath = "asdt-shared/skills/executor-header.md"

// agentTypeSpec describes one ASDT executor agent type. The same spec is
// rendered into each assistant's native agent-definition format.
type agentTypeSpec struct {
	ID          string   // "analyst" | "builder"; agent file name is "asdt-"+ID
	Description string   // one-line summary emitted into the frontmatter
	ClaudeTools []string // exact Claude Code tools allowlist, in emission order
	Constraints string   // English constraint prose appended after the executor header
}

// AgentTypeNames lists the canonical ASDT agent type IDs in render order.
// workflow.yaml `agent:` fields must use one of these values.
var AgentTypeNames = []string{"analyst", "builder"}

const analystConstraints = `## asdt-analyst constraints

You are a single-step ANALYSIS executor: you read, inspect, and reason, then
persist exactly one artifact. You never write to the working tree.

- Bash is INSPECTION-ONLY. Allowed command families: git log, git diff,
  git status, git show, git blame, ls, find, grep, wc, cat.
- Forbidden command families: anything that mutates state — creating, editing,
  moving, or deleting files; git add / git commit / git push; package installs;
  build or codegen output; redirections that write to disk.
- If the step you were given appears to require creating or editing files,
  STOP: do not attempt the write. Record the blocked work and its target paths
  in open_items — that work belongs to asdt-builder.
- You never delegate: do not call Agent/Task and do not run other steps.
`

const builderConstraints = `## asdt-builder constraints

You are a single-step IMPLEMENTATION executor: you create and edit files to
complete the one step you were given, then persist exactly one artifact.

- Create or edit files ONLY within the targets your step declares (its
  files_to_modify / allowed edit roots). If a needed change falls outside the
  declared targets, STOP that edit, do not write it, and record the path in
  open_items.
- You never delegate: do not call Agent/Task and do not run other steps.
`

// agentTypeSpecs holds the canonical specs in the same order as AgentTypeNames.
var agentTypeSpecs = []agentTypeSpec{
	{
		ID:          "analyst",
		Description: "Read-only ASDT step executor for analysis steps — inspects the repository with inspection-only Bash and persists one artifact, never writing files.",
		ClaudeTools: []string{
			"Read", "Glob", "Grep", "Bash",
			"mcp__plugin_engram_engram__mem_save",
			"mcp__plugin_engram_engram__mem_search",
			"mcp__plugin_engram_engram__mem_get_observation",
			"mcp__engram__mem_save",
			"mcp__engram__mem_search",
			"mcp__engram__mem_get_observation",
		},
		Constraints: analystConstraints,
	},
	{
		ID:          "builder",
		Description: "Write-capable ASDT step executor for implementation steps — creates and edits files within declared targets and persists one artifact, never delegating.",
		ClaudeTools: []string{
			"Read", "Glob", "Grep", "Bash", "Edit", "Write",
			"mcp__plugin_engram_engram__mem_save",
			"mcp__plugin_engram_engram__mem_search",
			"mcp__plugin_engram_engram__mem_get_observation",
			"mcp__plugin_engram_engram__mem_update",
			"mcp__engram__mem_save",
			"mcp__engram__mem_search",
			"mcp__engram__mem_get_observation",
			"mcp__engram__mem_update",
		},
		Constraints: builderConstraints,
	},
}

// analystBashAllowlist is the OpenCode bash permission whitelist for the
// read-only analyst agent, in emission order. The catch-all "*": deny is
// emitted last by writeOpenCodePermissions.
var analystBashAllowlist = []string{
	"git log*", "git diff*", "git status", "git show*", "git blame*",
	"ls*", "find*", "grep *", "wc *", "cat *",
}

// renderClaudeAgent renders one agent type as a Claude Code agent definition
// (~/.claude/agents/asdt-<id>.md). The tools line is ALWAYS emitted so the
// allowlist is explicit — omission would grant the harness default toolset.
func renderClaudeAgent(spec agentTypeSpec, executorHeader string) string {
	var b strings.Builder

	b.WriteString("---\n")
	b.WriteString("name: asdt-")
	b.WriteString(spec.ID)
	b.WriteString("\n")
	b.WriteString("description: ")
	b.WriteString(spec.Description)
	b.WriteString("\n")
	b.WriteString("tools: ")
	b.WriteString(strings.Join(spec.ClaudeTools, ", "))
	b.WriteString("\n")
	b.WriteString("---\n\n")
	writeAgentBody(&b, executorHeader, spec.Constraints)

	return b.String()
}

// renderOpenCodeAgent renders one agent type as an OpenCode subagent
// definition with a structural permission block.
func renderOpenCodeAgent(spec agentTypeSpec, executorHeader string) string {
	var b strings.Builder

	b.WriteString("---\n")
	b.WriteString("description: ")
	b.WriteString(spec.Description)
	b.WriteString("\n")
	b.WriteString("mode: subagent\n")
	b.WriteString("permission:\n")
	writeOpenCodePermissions(&b, spec.ID)
	b.WriteString("---\n\n")
	writeAgentBody(&b, executorHeader, spec.Constraints)

	return b.String()
}

func writeOpenCodePermissions(b *strings.Builder, id string) {
	if id == "builder" {
		b.WriteString("  edit: allow\n")
		b.WriteString("  task: deny\n")
		b.WriteString("  bash: allow\n")
		return
	}

	// analyst (and any future read-only type): deny everything except the
	// inspection-only bash whitelist.
	b.WriteString("  edit: deny\n")
	b.WriteString("  task: deny\n")
	b.WriteString("  bash:\n")
	for _, glob := range analystBashAllowlist {
		b.WriteString("    \"")
		b.WriteString(glob)
		b.WriteString("\": allow\n")
	}
	b.WriteString("    \"*\": deny\n")
}

// writeAgentBody emits the shared agent body: the executor header VERBATIM,
// one blank line, then the agent type's constraint prose.
func writeAgentBody(b *strings.Builder, executorHeader, constraints string) {
	b.WriteString(executorHeader)
	if !strings.HasSuffix(executorHeader, "\n") {
		b.WriteString("\n")
	}
	b.WriteString("\n")
	b.WriteString(constraints)
	if !strings.HasSuffix(constraints, "\n") {
		b.WriteString("\n")
	}
}

// AgentAdapterDescriptor describes how to generate executor agent definitions
// for one assistant, on top of the shared skill-tree copy.
type AgentAdapterDescriptor struct {
	AssistantID AssistantID
	Generate    func(skillsFS fs.FS, agentRoot string) ([]string, error)
}

// AgentAdapters lists assistants that receive generated executor agent
// definitions. Unlike CommandAdapters, BOTH assistants carry an entry —
// each has a native agents directory.
var AgentAdapters = []AgentAdapterDescriptor{
	{
		AssistantID: AssistantClaudeCode,
		Generate:    generateClaudeAgents,
	},
	{
		AssistantID: AssistantOpenCode,
		Generate:    generateOpenCodeAgents,
	},
}

func agentAdapterFor(id AssistantID) (AgentAdapterDescriptor, bool) {
	for _, adapter := range AgentAdapters {
		if adapter.AssistantID == id {
			return adapter, true
		}
	}
	return AgentAdapterDescriptor{}, false
}

// agentRootFor resolves the agents directory for one assistant, or "" when
// the assistant has no known agent root.
func agentRootFor(id AssistantID) string {
	switch id {
	case AssistantClaudeCode:
		home, _ := os.UserHomeDir()
		return home + "/.claude/agents"
	case AssistantOpenCode:
		if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
			return xdg + "/opencode/agents"
		}
		home, _ := os.UserHomeDir()
		return home + "/.config/opencode/agents"
	default:
		return ""
	}
}

func generateClaudeAgents(skillsFS fs.FS, agentRoot string) ([]string, error) {
	return generateAgentFiles(skillsFS, agentRoot, renderClaudeAgent)
}

func generateOpenCodeAgents(skillsFS fs.FS, agentRoot string) ([]string, error) {
	return generateAgentFiles(skillsFS, agentRoot, renderOpenCodeAgent)
}

// generateAgentFiles writes one agent definition file per agent type spec
// under agentRoot, baking the shared executor header into every body. A
// skillsFS without the executor header (a partial fixture) generates nothing
// — the embedded production FS is guaranteed to carry it. Per-file failures
// are isolated: the first error is reported, the remaining specs still write.
func generateAgentFiles(skillsFS fs.FS, agentRoot string, render func(agentTypeSpec, string) string) ([]string, error) {
	header, readErr := fs.ReadFile(skillsFS, executorHeaderPath)
	if readErr != nil {
		if os.IsNotExist(readErr) {
			return nil, nil
		}
		return nil, fmt.Errorf("read %s: %w", executorHeaderPath, readErr)
	}

	if mkErr := os.MkdirAll(agentRoot, 0o755); mkErr != nil {
		return nil, fmt.Errorf("mkdir %s: %w", agentRoot, mkErr)
	}

	var written []string
	var firstErr error

	for _, spec := range agentTypeSpecs {
		target := filepath.Join(agentRoot, "asdt-"+spec.ID+".md")
		content := render(spec, string(header))

		if writeErr := os.WriteFile(target, []byte(content), 0o644); writeErr != nil {
			if firstErr == nil {
				firstErr = fmt.Errorf("write %s: %w", target, writeErr)
			}
			continue
		}
		written = append(written, target)
	}

	return written, firstErr
}

// generateAgents mirrors generateCommands: resolve the adapter and root for
// the assistant (absence of either is the no-op), generate, fold written
// paths into result.WrittenCommands and any error into result.Err.
func generateAgents(assistant AssistantDescriptor, skillsFS fs.FS, result *InstallResult) {
	adapter, ok := agentAdapterFor(assistant.ID)
	if !ok {
		return
	}

	agentRoot := agentRootFor(assistant.ID)
	if agentRoot == "" {
		return
	}

	written, genErr := adapter.Generate(skillsFS, agentRoot)
	result.WrittenCommands = append(result.WrittenCommands, written...)
	if genErr != nil {
		result.Err = genErr
	}
}
