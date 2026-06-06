# AI Software Delivery Team

ASDT is a skill package manager for AI assistants. It installs a team of software specialists — architect, developer, QA, security, UX/UI — into Claude Code or OpenCode, giving the assistant a structured, repeatable workflow for delivering software features.

Without ASDT, AI assistants have no enforced discipline around how a feature gets designed, implemented, tested, or reviewed. ASDT installs that discipline as slash commands the assistant can invoke.

## Requirements

- Claude Code (`claude`) or OpenCode (`opencode`) installed
- Engram MCP server running (persistent memory provider)
- Go 1.22+ to build from source

## Installation

Install both binaries:

```sh
go install github.com/vitualizz/ai-software-delivery-team/cmd/asdt@latest
go install github.com/vitualizz/ai-software-delivery-team/cmd/asdt-tui@latest
```

## Getting Started

From the root of your software project:

**1. Initialize the knowledge base**

```sh
asdt init
```

Detects your project stack (Go, Node, Rust, Python, Ruby) and writes `.asdt/knowledge/platform.yaml` and `.asdt/knowledge/platform-summary.yaml`. No LLM involved — pure static analysis.

**2. Install skills into your AI assistant**

```sh
asdt-tui
```

Interactive TUI that lets you choose which AI assistant(s) to install into, and configures Engram as the memory provider. Skills are installed to `~/.claude/skills/asdt/` (Claude Code) or `~/.config/opencode/skills/asdt/` (OpenCode).

**3. Set your active change**

In `.asdt/config.yaml` at the project root:

```yaml
active_change: my-feature-name
memory:
  provider: engram
```

**4. Start the Engram MCP server**, then open your AI assistant in the project directory.

## Using the Specialists

Invoke from inside your AI assistant:

| Command | What it does | Produces |
|---|---|---|
| `/asdt` | Meta-orchestrator — analyzes your request, recommends which specialists to run and in what order | — |
| `/asdt:architect` | ADRs, system design, risk analysis | `architectural-decision.yaml`, `system-design.yaml` |
| `/asdt:developer` | Implementation plan with code | `implementation-plan.yaml` |
| `/asdt:qa` | Test plan | `test-plan.yaml` |
| `/asdt:security` | Threat model and hardening checklist | `security-findings.yaml`, `hardening-checklist.yaml` |
| `/asdt:ux-ui` | UX brief and component specs | `ux-brief.yaml`, `component-spec.yaml` |

Each specialist reads prior decisions from Engram memory, analyzes the current context, produces its YAML artifacts under `.asdt/artifacts/{change}/`, and saves the decision back to memory for the next specialist to build on.

Start with `/asdt` if you are unsure which specialist to invoke — it will tell you.

## Other Commands

```sh
asdt status   # Show the state of the active change
```

## Project Layout

```
cmd/
  asdt/        CLI entrypoint
  asdt-tui/    Installer TUI entrypoint
internal/
  installer/   Skill detection and installation logic
  setup/       Installer TUI (Bubbletea)
  tui/         Status observer: specialists panel + artifacts browser
  config/      Read/write .asdt/config.yaml, walk-up discovery
  knowledge/   Project stack detection (no LLM)
  artifact/    YAML artifact store under .asdt/artifacts/{change}/
  memory/      Provider interface for cross-session memory
  prompt/      Layered prompt assembly (role + skills + artifacts + platform)
  pipeline/    FSM tracking which specialists have run per change
skill/         SKILL.md files embedded in the binary, copied to the assistant on install
```
