# ASDT — AI Software Delivery Team

*Specialist-first AI delivery. Decisions preserved, not just code generated.*

---

## The philosophy

Most AI tools collapse your entire team into a single chat. A PM, architect,
developer, QA engineer, and security analyst don't do the same job — and neither
should your AI. ASDT models a real software delivery organization: each
specialist owns its own discipline, its own workflow, and its own artifacts.
Agents communicate through files, not conversations. The result is a durable,
reviewable, replayable record of every decision made.

---

## Three invariants

- **One boundary**: everything lives under `.asdt/` — your project root stays
  clean. Uninstall with `rm -rf .asdt/`.
- **One behavior**: same specialists in Claude Code, OpenCode, and the TUI.
- **No required order**: any specialist can run first — they're independent
  professionals.

---

## The specialists

| Command | Role | Produces |
|---------|------|----------|
| `/asdt:ux-ui` | UX/UI Specialist | `ux-brief.yaml`, `component-spec.yaml` |
| `/asdt:architect` | Architect | `architectural-decision.yaml`, `system-design.yaml` |
| `/asdt:developer` | Developer | `implementation-plan.yaml` |
| `/asdt:qa` | QA Engineer | `test-plan.yaml`, `test-cases.yaml` |
| `/asdt:security` | Security Specialist | `threat-model.yaml`, `security-findings.yaml` |
| `/asdt "request"` | Meta-orchestrator | Suggests which specialists to run |

---

## Quick start

```bash
# Ask the meta-orchestrator to suggest a plan
/asdt "add AI Reports Dashboard"

# Then run each suggested specialist
/asdt:ux-ui       # Platform Analysis → IA → User Flows → Component Spec
/asdt:architect   # Constraints → ADR → System Design → Risk Analysis
/asdt:developer   # Loads all prior artifacts → Implementation Plan + Code

# Or run Security at any point — no predecessor required
/asdt:security    # Threat Modeling → OWASP Analysis → Hardening Checklist
```

---

## What gets produced

```
.asdt/
├── config.yaml
├── knowledge/
│   └── platform.yaml              # tech stack, conventions, design fingerprint
└── artifacts/add-ai-reports/
    ├── ux-brief.yaml              # user flows, IA, component mapping
    ├── component-spec.yaml        # reused/new components
    ├── architectural-decision.yaml # ADR with alternatives considered
    ├── system-design.yaml         # data model, API surface, risks
    ├── implementation-plan.yaml   # step-by-step plan + code snippets
    ├── test-plan.yaml             # test strategy + test cases
    ├── security-findings.yaml     # threats, OWASP findings, hardening checklist
    └── pipeline-state.yaml        # per-specialist step history
```

Every file is a plain YAML artifact you can open in any editor, commit to git,
diff in a PR, and replay in a future session.

---

## Platform awareness

ASDT is designed for *existing systems*, not greenfield projects. Every
specialist loads `platform.yaml` before doing any work — it contains your tech
stack, naming conventions, existing components, and design patterns. The goal
isn't generating software. The goal is extending your software consistently, in
a way that feels like it belongs there.

---

## Architecture

```
skill/                  ← THE PRODUCT: runtime-agnostic specialist SKILL.md files
├── SKILL.md            ← /asdt meta-orchestrator
├── _shared/skills/     ← platform-context, artifact-envelope, scope-definition
├── developer/          ← /asdt:developer (7-step workflow)
├── ux-ui/              ← /asdt:ux-ui (7-step workflow)
├── architect/          ← /asdt:architect (7-step workflow)
├── qa/                 ← /asdt:qa (6-step workflow)
└── security/           ← /asdt:security (5-step, no required predecessor)

internal/               ← Go packages for the optional TUI binary
├── specialists/        ← SpecialistDescriptor + generic Runner (the core abstraction)
├── artifact/           ← Envelope[T], FSStore
├── memory/             ← MemoryProvider interface, NullProvider, EngramProvider stub
├── pipeline/           ← PipelineStateV2, AdvanceStep
├── prompt/             ← layered composition, ScopedSkill registry
├── knowledge/          ← platform.yaml detector
├── llm/                ← Provider interface + Anthropic/OpenAI/Mock
└── tui/                ← Bubbletea TUI (optional)

schemas/                ← YAML schemas for all artifact types
docs/adr/               ← Architecture Decision Records (ADR-001 through ADR-007)
```

---

## Adding a specialist

```
# 1. Write the skill file
skill/my-specialist/SKILL.md

# 2. Add a descriptor (one struct literal in Go)
func MySpecialistDescriptor() SpecialistDescriptor { ... }

# 3. Register it
descriptors["my-specialist"] = specialists.MySpecialistDescriptor()
```

Zero new Go packages. The specialist system is data-driven.

---

## Install the TUI (optional)

The Go TUI is optional. ASDT works without it through any AI coding assistant.

```bash
go install github.com/vitualizz/ai-software-delivery-team/cmd/asdt@latest
```

---

## Use in Claude Code

Install `skill/` as a skill package. Each `/asdt:{specialist}` SKILL.md is
independently invocable.

```bash
cp -r skill/ ~/.claude/skills/asdt
```

## Use in OpenCode

Map `/asdt` and each specialist to the corresponding SKILL.md in your OpenCode
configuration:

```toml
[commands.asdt]
path = ".claude/skills/asdt/SKILL.md"

[commands."asdt:developer"]
path = ".claude/skills/asdt/developer/SKILL.md"
```

---

## Memory

By default, no memory is used (`NullProvider`). Configure Engram by adding to
`.asdt/config.yaml`:

```yaml
memory:
  provider: engram
```

---

## Contributing

See [docs/contributing.md](docs/contributing.md).

---

## License

MIT
