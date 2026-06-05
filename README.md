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

## The specialists

Each specialist is a **domain expert** with its own workflow, its own skills, and
its own artifacts. They communicate through files, not conversations.

| Command | Specialist | Produces |
|---------|-----------|----------|
| `asdt ux-ui` | UX/UI Designer | `ux-brief.yaml`, `component-spec.yaml` |
| `asdt architect` | Software Architect | `architectural-decision.yaml`, `system-design.yaml`, `risk-register.yaml` |
| `asdt developer` | Developer | `implementation-plan.yaml` |
| `asdt qa` | QA Engineer | `test-plan.yaml`, `quality-report.yaml` |
| `asdt security` | Security Engineer | `threat-model.yaml`, `security-findings.yaml`, `hardening-checklist.yaml` |

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

---

## The artifacts

Artifacts are YAML files written to `.asdt/artifacts/{change}/`. Each artifact is
the **primary communication medium** between specialists: an architect produces
`architectural-decision.yaml`, a developer reads it. No conversation, no ambiguity.

**Why artifacts exist**: Specialists do not share conversations. They share
documents. Artifacts are structured, versioned, traceable documents — every one
carries a prompt version hash and references its inputs. You can `git diff` them,
review them in a PR, and replay them months later.

---

## The memory layer

Memory is the **organizational knowledge accumulation** layer. After each
specialist run, ASDT records what was decided, why, and where. Future runs query
this layer to avoid contradicting prior decisions.

**Why memory exists**: Knowledge that disappears when a session ends is not
organizational knowledge — it is a transcript. The memory layer (backed by Engram
or the local `.asdt/runs/` filesystem) turns specialist output into durable
institutional memory that accumulates across changes.

---

## Quick start

```bash
# Initialize the project — scans your stack, writes platform-summary.yaml
asdt init

# Run a specialist for a change
asdt architect --change add-auth
asdt developer --change add-auth
asdt security  --change add-auth
```

---

## What gets produced

```
.asdt/
├── config.yaml                         # memory provider, active change
├── knowledge/
│   ├── platform.yaml                   # detected stack, conventions
│   └── platform-summary.yaml          # deterministic summary (zero LLM tokens)
├── runs/                               # knowledge timeline (NullProvider default)
│   └── 20260605-120000/
│       └── architect-add-auth.yaml    # entry: what was decided and why
└── artifacts/add-auth/
    ├── architectural-decision.yaml    # ADR with alternatives considered
    ├── system-design.yaml             # data model, API surface, risks
    ├── risk-register.yaml             # top 3-5 risks + mitigations
    ├── implementation-plan.yaml       # ordered tasks + code snippets
    ├── test-plan.yaml                 # test strategy + test cases
    ├── security-findings.yaml         # OWASP findings + remediations
    ├── hardening-checklist.yaml       # ordered hardening actions
    └── pipeline-state.yaml            # per-specialist step history
```

Every file is plain YAML: open it in any editor, commit it to git, diff it in a
PR, and replay it in a future session.

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
├── developer/            # 7-step workflow
├── ux-ui/                # 7-step workflow
├── architect/            # 7-step workflow
├── qa/                   # 6-step workflow
└── security/             # 5-step workflow (no required predecessor)

internal/                 # Go packages for the optional binary
├── specialists/          # SpecialistDescriptor + generic Runner
├── memory/               # Provider port: NullProvider (FS) + EngramProvider (MCP)
├── artifact/             # Envelope[T], FSStore
├── pipeline/             # PipelineStateV2, AdvanceStep
├── prompt/               # layered composition, ScopedSkill registry
├── knowledge/            # platform.yaml detector
├── llm/                  # Provider interface + Mock
└── tui/                  # Bubbletea TUI (optional)

schemas/                  # YAML schemas for all artifact types
docs/adr/                 # Architecture Decision Records (ADR-001 through ADR-010)
```

---

## Memory configuration

By default, runs are stored in `.asdt/runs/` (NullProvider — no external
dependencies). To use Engram for cross-session semantic search, add to
`.asdt/config.yaml`:

```yaml
memory:
  provider: engram
  project: your-project-name
```

---

## Adding a specialist

```bash
# 1. Write the skill file
skill/my-specialist/SKILL.md

# 2. Add a descriptor (one struct literal)
func MySpecialistDescriptor() SpecialistDescriptor { ... }

# 3. Register it
descriptors["my-specialist"] = specialists.MySpecialistDescriptor()
```

Zero new Go packages. The specialist system is data-driven.

---

## Install the binary (optional)

ASDT works without the binary through any AI coding assistant. The binary is the
TUI frontend for the same specialist system.

```bash
go install github.com/vitualizz/ai-software-delivery-team/cmd/asdt@latest
```

---

## Use in Claude Code

Install `skill/` as a skill package:

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
