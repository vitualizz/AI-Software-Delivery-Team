# AI Software Delivery Team

ASDT is a skill package manager for AI assistants. It installs a team of software specialists — architect, developer, QA, security, UX/UI — into Claude Code or OpenCode, giving the assistant a structured, repeatable workflow for delivering software features.

Without ASDT, AI assistants have no enforced discipline around how a feature gets designed, implemented, tested, or reviewed. ASDT installs that discipline as slash commands the assistant can invoke.

## Requirements

- Claude Code (`claude`) or OpenCode (`opencode`) installed
- [Engram MCP server](https://github.com/Gentleman-Programming/engram) installed and running (persistent memory provider)
- Go 1.22+ to install from source

## Installation

```sh
curl -fsSL https://raw.githubusercontent.com/vitualizz/ai-software-delivery-team/main/install.sh | bash
```

Downloads the pre-built binary for your platform (Linux/macOS, x86_64/arm64) and installs it to `~/.local/bin/`. No Go required.

## Getting Started

**1. Install skills into your AI assistant**

```sh
asdt-tui
```

Interactive TUI that checks Engram is installed, lets you choose which AI assistant(s) to target, and copies the ASDT skills into them. Each skill (the `asdt` consultant, every specialist, and the shared fragment library) is installed as its own top-level sibling directory directly under `~/.claude/skills/` (Claude Code) or `~/.config/opencode/skills/` (OpenCode) — e.g. `~/.claude/skills/asdt-architect/`, `~/.claude/skills/asdt-shared/` — so each specialist is independently invocable (`asdt:architect`, `asdt:developer`, …) alongside `/asdt`.

**2. Initialize your project**

Open your AI assistant in the project directory and run:

```
/asdt:init
```

The assistant will detect your project stack, ask a few configuration questions, and write `.asdt/config.yaml` and `.asdt/knowledge/platform.yaml`.

**3. Start the Engram MCP server**, then use the specialists.

## Using the Specialists

Invoke from inside your AI assistant:

| Command | What it does | Produces |
|---|---|---|
| `/asdt` | Meta-orchestrator — analyzes your request, recommends which specialists to run and in what order | — |
| `/asdt:init` | Initialize ASDT for the project — detects stack, writes config | `.asdt/config.yaml`, `platform.yaml` |
| `/asdt:architect` | ADRs, system design, risk analysis | `architectural-decision.yaml`, `system-design.yaml` |
| `/asdt:developer` | Implementation plan with code | `implementation-plan.yaml` |
| `/asdt:qa` | Test plan | `test-plan.yaml` |
| `/asdt:security` | Threat model and hardening checklist | `security-findings.yaml`, `hardening-checklist.yaml` |
| `/asdt:ux-ui` | UX brief and component specs | `ux-brief.yaml`, `component-spec.yaml` |

Each specialist reads prior decisions from Engram memory, analyzes the current context, produces its artifacts, and saves the decision back to memory for the next specialist to build on.

Start with `/asdt` if you are unsure which specialist to invoke — it will tell you.

## Project Layout

```
cmd/
  asdt-tui/    Installer TUI — the only user-facing binary
internal/
  installer/   Skill detection and installation logic
  setup/       Installer TUI (Bubbletea + Lipgloss styles)
  setup/styles/ Centralized color palette and style definitions
  tui/         Status observer: specialists panel + artifacts browser
  config/      Read/write .asdt/config.yaml, walk-up discovery
  knowledge/   Project stack detection (no LLM)
  artifact/    YAML artifact store under .asdt/artifacts/{change}/
  memory/      Provider interface for cross-session memory
  prompt/      Layered prompt assembly (role + skills + artifacts + platform)
  pipeline/    FSM tracking which specialists have run per change
skill/         SKILL.md files embedded in the binary, copied to the assistant on install
```
