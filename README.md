# ASDT — AI Software Delivery Team

**A Software Delivery Knowledge System** — not an AI coding tool.

ASDT models the *knowledge* your software organization accumulates when it
delivers software well: architecture decisions, UX rationale, security posture,
test coverage strategies. Code is one output. Accumulated, searchable knowledge
is the primary asset.

---

## Why ASDT exists

Software organizations scale through accumulated knowledge, not through code.
Every architecture decision made, every threat modeled, every UX tradeoff
evaluated — these are the assets that let future changes happen faster and safer.
Traditional AI coding tools generate code and forget everything. ASDT generates
code *and* builds an organizational knowledge base that survives beyond any single
change.

---

## How ASDT differs from AI coding tools

| AI Coding Tool | ASDT |
|----------------|------|
| Chat history is the memory | Persistent knowledge records survive across changes |
| One AI doing everything | Domain specialists with isolated expertise |
| Context bloat over time | Each step gets only the context it needs |
| Code generation as the goal | Knowledge accumulation as the goal; code as one output |
| No traceability | Every artifact references its inputs and prompt version |

---

## How it works

**Your AI assistant is the runtime.** ASDT is a set of `SKILL.md` files that
Claude Code, OpenCode, Cursor, or any AI coding assistant reads and executes.
When you invoke a specialist, the AI assistant runs the workflow steps, calls
your memory provider (Engram) via MCP, and saves structured knowledge records —
no separate process, no binary required.

```
You → /asdt:architect "add auth"
         ↓
   AI assistant reads skill/architect/SKILL.md
         ↓
   Executes 7-step workflow
         ↓
   Saves decisions → Engram (mem_save)
         ↓
   Future runs query prior decisions → Engram (mem_search)
```

**A memory provider is required.** There is no filesystem fallback. Knowledge
that disappears between sessions is not organizational knowledge — it is a
transcript. Configure Engram (or an equivalent MCP memory server) before
running any specialist.

---

## The specialists

Each specialist is a **domain expert** with its own workflow, its own skills, and
its own artifacts. They communicate through the memory provider, not through
conversations.

| Command | Specialist | Produces |
|---------|-----------|----------|
| `asdt ux-ui` | UX/UI Designer | UX brief, component spec |
| `asdt architect` | Software Architect | Architecture decision, system design, risk register |
| `asdt developer` | Developer | Implementation plan |
| `asdt qa` | QA Engineer | Test plan, quality report |
| `asdt security` | Security Engineer | Threat model, security findings, hardening checklist |

**Why specialists exist**: Domain experts with isolated context produce better
decisions than a generalist with everything crammed into one conversation. An
architect thinking about tradeoffs should not also be worrying about OWASP
vulnerabilities. Isolation makes each specialist trustworthy in its domain.

---

## The skills

Skills are **context modules** loaded at specific workflow steps. Each skill
carries only the instructions a step needs — no more.

**Why skills exist**: Context isolation. Every token in a prompt influences the
output. A developer step writing code should not carry the UX specialist's
information architecture thinking — that would dilute it. Skills enforce that each
step operates with exactly the context it needs and nothing it does not.

Each specialist defines its pipeline in `workflow.yaml`:

```yaml
# skill/architect/workflow.yaml
steps:
  - name: platform-context
    skill: ../_shared/skills/platform-context.md
  - name: knowledge-recall
    skill: ../_shared/skills/knowledge-recall.md
  - name: constraints-analysis
    ...
```

The AI assistant reads `workflow.yaml` at runtime to know which steps to
execute and in what order.

---

## The memory layer

Memory is the **organizational knowledge accumulation** layer. After each
specialist run, ASDT records what was decided, why, and where. Future runs
query this layer to avoid contradicting prior decisions.

**Memory is not optional.** Configure a provider before running specialists.
ASDT supports any MCP-compatible memory server. Engram is the reference
implementation.

Add to `.asdt/config.yaml`:

```yaml
memory:
  provider: engram
  project: your-project-name
```

---

## Quick start

### 1. Install Engram

```bash
go install github.com/Gentleman-Programming/engram@latest
```

### 2. Configure your AI assistant

Add Engram as an MCP server. Example for Claude Code:

```bash
code --add-mcp '{"name":"engram","command":"engram","args":["mcp"]}'
```

### 3. Install ASDT skills

```bash
# Copy skills into your AI assistant's skill directory
cp -r skill/ ~/.claude/skills/asdt
```

### 4. Initialize your project

```bash
asdt init
# Scans your stack, writes .asdt/knowledge/platform-summary.yaml
```

### 5. Run a specialist

```
/asdt:architect "add auth"
/asdt:developer "add auth"
/asdt:security  "add auth"
```

---

## What gets recorded

Every specialist run saves structured knowledge to your memory provider:

```
Engram (mem_save):
  - What was decided and why (architectural-decision)
  - System design and data model (system-design)
  - Top risks and mitigations (risk-register)
  - Implementation plan with ordered tasks (implementation-plan)
  - Test strategy and test cases (test-plan)
  - OWASP findings and remediations (security-findings)
  - Ordered hardening actions (hardening-checklist)
```

Every record is searchable across sessions. The next specialist run queries
prior decisions before acting — so your architect never contradicts a security
finding made three changes ago.

---

## Architecture

```
skill/                    # THE PRODUCT: runtime-agnostic SKILL.md files
├── SKILL.md              # /asdt meta-orchestrator
├── _shared/skills/       # shared skills loaded per-step
│   ├── knowledge-recall.md       # query org memory before decisions
│   ├── decision-preservation.md  # record decisions to org memory
│   ├── platform-context.md
│   └── artifact-envelope.md
├── architect/
│   ├── SKILL.md          # specialist entry point
│   └── workflow.yaml     # step pipeline definition
├── developer/
├── ux-ui/
├── qa/
└── security/

cmd/asdt/                 # optional package manager binary
├── main.go               # install skills, configure providers, manage .asdt/config.yaml

internal/                 # Go packages for the package manager binary
├── config/               # read/write .asdt/config.yaml
├── knowledge/            # platform.yaml stack detector (zero LLM tokens)
└── tui/                  # Bubbletea TUI

schemas/                  # YAML schemas for all artifact types
docs/adr/                 # Architecture Decision Records (ADR-001 through ADR-010)
```

---

## The binary (optional)

The `asdt` binary is a **package manager and configurator** — it is not the
specialist runtime. The AI assistant is the runtime.

What the binary does:
- `asdt init` — detect your stack, write `platform-summary.yaml`
- `asdt install` — copy skills to the right path for your AI assistant
- `asdt update` — pull updated skills and memory provider

What the binary does NOT do:
- Run specialists (that is the AI assistant's job)
- Call LLMs
- Connect to Engram (the AI assistant handles MCP transport)

```bash
go install github.com/vitualizz/ai-software-delivery-team/cmd/asdt@latest
```

---

## Adding a specialist

```bash
# 1. Write the skill entry point
skill/my-specialist/SKILL.md

# 2. Define the pipeline
skill/my-specialist/workflow.yaml

# 3. Add any step-specific skills
skill/my-specialist/steps/
```

No Go code required. Specialists are purely SKILL.md + workflow.yaml.

---

## Use in Claude Code

```bash
cp -r skill/ ~/.claude/skills/asdt
```

Then invoke each specialist with `/asdt:developer`, `/asdt:architect`, etc.

---

## Contributing

See [docs/contributing.md](docs/contributing.md).

---

## License

MIT
