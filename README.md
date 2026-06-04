# ASDT — AI Software Delivery Team

Artifact-first AI delivery pipeline. Requirements → Plan → Code, not just Code.

---

## The problem

Most AI coding tools take you from Idea → Prompt → Code. The reasoning that
produced that code evaporates when the chat ends. When requirements change, or a
reviewer asks "why was this designed this way?", there is no record. ASDT inserts
the steps that a real delivery team would take — and makes each step a durable,
human-readable artifact you can read, review, and replay. The conversation is
ephemeral; the artifacts are not.

---

## Three invariants

- **One entry point**: `/asdt <subcommand>` — one command, one dispatch table,
  no per-agent skills to install or manage.
- **One boundary**: everything lives under `.asdt/` — your project root stays
  clean. Uninstall with `rm -rf .asdt/`.
- **One behavior**: the same prompts, the same schemas, the same artifact output
  in Claude Code, OpenCode, and the standalone TUI.

---

## Quick start

```bash
# inside any project
/asdt knowledge              # scan project → .asdt/knowledge/platform.yaml
/asdt requirements "add user authentication with email and password"
/asdt develop                # reads requirements → .asdt/artifacts/{change}/implementation-plan.yaml
/asdt status                 # show pipeline state and transition history
```

---

## What gets produced

```
.asdt/
├── knowledge/
│   └── platform.yaml                   # detected stack, conventions, design fingerprint
└── artifacts/add-user-auth/
    ├── requirements-spec.yaml           # user stories, acceptance criteria, scope, NFRs
    ├── implementation-plan.yaml         # step-by-step plan with file names and code snippets
    └── pipeline-state.yaml             # current state + full transition history with timestamps
```

Every file is a plain YAML artifact you can open in any editor, commit to git,
diff in a PR, and replay in a future session.

---

## Architecture

ASDT ships as two independent artifacts that share one boundary (`.asdt/`):

1. **`skill/`** — the canonical, runtime-agnostic `/asdt` router. Pure prompt +
   file I/O. Runs identically in Claude Code and OpenCode. No Go dependency.
2. **Go binary** — an optional TUI that reads `.asdt/` for visualization. Built
   with hexagonal architecture (ports and adapters); depends on the skill's
   artifact schemas but owns none of the agent logic.

The shared contract is `Envelope[T]` — a typed YAML structure with a uniform
header (`schema_version`, `agent`, `change_id`, `created_at`, `prompt_version`,
`input_refs[]`) and a typed payload. Agents communicate only through these
files; there is no direct agent-to-agent messaging.

```
ai-software-delivery-team/
├── skill/                      # primary deliverable — runtime-agnostic /asdt package
│   ├── SKILL.md                # the /asdt router: dispatch table + invariants
│   └── prompts/
│       ├── roles/              # agent role prompts (requirements, developer, knowledge)
│       └── skills/             # reusable capability fragments (DRY prompt logic)
├── schemas/                    # canonical YAML schemas (single source of truth)
├── cmd/asdt/                   # TUI binary entry — composition root only, no logic
├── internal/
│   ├── artifact/               # Envelope[T], Store interface, FSStore adapter
│   ├── knowledge/              # platform.yaml detector and reader
│   ├── pipeline/               # sequential FSM, state machine, transition rules
│   ├── prompt/                 # layered composition engine, override resolver
│   ├── llm/                    # Provider interface, Anthropic/OpenAI/Mock adapters
│   ├── tui/                    # Bubbletea root model and panel components
│   └── config/                 # .asdt/ walk-up discovery, config.yaml R/W
├── testdata/                   # fixtures and golden files for Go tests
├── examples/                   # end-user-facing example runs
└── docs/
    ├── contributing.md         # how to add an agent or skill fragment
    ├── adr/                    # ADR-001..005 as individual files
    └── runtime-compat.md       # Claude Code vs OpenCode invocation mapping
```

---

## Install TUI (optional)

The Go TUI is optional. ASDT works without it through any AI coding assistant.

```bash
go install github.com/vitualizz/ai-software-delivery-team/cmd/asdt@latest
```

---

## Use in Claude Code

Copy `skill/` into your project's `.claude/skills/` directory, or install it
globally in `~/.claude/skills/`. The `/asdt` command becomes available
immediately in any Claude Code session.

```bash
cp -r skill/ ~/.claude/skills/asdt
```

## Use in OpenCode

Map `/asdt` to `skill/SKILL.md` in your OpenCode configuration:

```toml
[commands.asdt]
path = ".claude/skills/asdt/SKILL.md"
```

---

## Runtimes

Works in Claude Code, OpenCode, and as a standalone Go TUI. The behavior is
identical across all three — same prompts, same schemas, same artifact output.

---

## Contributing

See [docs/contributing.md](docs/contributing.md). You do not need to know Go to
contribute — prompt improvements and new agent definitions are first-class
contributions.

---

## License

MIT
