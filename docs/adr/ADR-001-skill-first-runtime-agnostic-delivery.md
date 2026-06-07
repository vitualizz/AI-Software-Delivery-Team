# ADR-001: Skill-First, Runtime-Agnostic Delivery

Date: 2026-06-04
Status: Accepted

## Context

ASDT needs to work in Claude Code, OpenCode, and as a TUI binary without code
duplication. The temptation is to build a Go binary first and wrap it. But the
primary users interact through AI coding assistants, not terminals.

Coding assistants share exactly two capabilities: completion (call the LLM) and
file I/O (read/write the project). Any design that requires a third capability
is not portable. ASDT's thesis is that the delivery pipeline — requirements,
planning, implementation review — is expressible entirely in those two
primitives.

## Decision

The `skill/SKILL.md` router IS the product. The Go TUI is an optional layer.
All core behavior — subcommand routing, prompt composition, artifact I/O — is
expressed as prompt instructions operating on plain YAML files. No
runtime-specific APIs are used at any point in the agent workflows.

The Go binary can read and display `.asdt/` artifacts but cannot own any agent
logic. If the binary and the skill diverge on behavior, the skill wins.

## Alternatives Considered

**Go binary as primary deliverable** — rejected. Requires installation before
first use; breaks on machines without Go; forces every contributor to know Go
even for prompt-only improvements.

**One skill file per agent** — rejected. Pollutes the skill namespace, breaks
the single-entry-point invariant (`/asdt <subcommand>`), and makes the dispatch
table implicit (scattered across many files) rather than explicit (one table in
`SKILL.md`).

**OpenAI function-calling protocol** — rejected. Vendor lock-in. Claude Code
and OpenCode do not share a function-calling API. File I/O is the only portable
inter-agent contract.

## Consequences

- Contributors can add new agents by writing a prompt file and adding one row to
  the dispatch table in `SKILL.md` — zero Go required.
- Cross-runtime parity is testable: the same prompt inputs must produce the same
  artifact schema on both runtimes.
- The Go binary can be adopted incrementally without blocking the skill from
  shipping.
- Prompt improvements are first-class contributions, as valuable as code changes.

## Addendum: Command-Palette Discoverability (2026-06-07)

Status: Accepted (addendum to the original decision above)

### Context

OpenCode surfaces a "/" command palette backed by Markdown files under a
dedicated `commands/` directory; Claude Code discovers specialists
directly from the copied skill tree and needs no equivalent file. This
difference was not anticipated by the original two-primitives framing:
it is not a new CAPABILITY one runtime has and the other lacks (both
runtimes still only need completion and file I/O), but a difference in
the FILE LAYOUT each runtime expects in order to make an installed
specialist discoverable through its native UI.

### Decision

Command-palette discoverability is handled as an ADDITIVE, per-assistant
`CommandAdapter` registry (`internal/installer/adapters.go`), mirroring
the existing `Providers` flat-registry idiom. For each assistant that
needs extra discoverability artifacts, the adapter generates them by
reading the SAME embedded `SKILL.md` frontmatter already used to install
the skill tree — writing one deterministic wrapper file per specialist,
purely via file I/O, with zero hardcoded per-specialist tables. An
assistant absent from the registry (Claude Code) gets no extra
artifacts; absence IS the no-op, exactly as `Providers` carries no
placeholder entry for "no provider".

This is a REFINEMENT of the original decision, not a departure from it:
the two primitives — completion and file I/O — remain the complete and
sufficient basis for the entire pipeline. What this addendum recognizes
is that "file I/O" includes writing runtime-specific DISCOVERABILITY
artifacts, not just the specialist content itself.

### Alternatives Considered

**A third "registration" primitive** — rejected. Generating a
command-palette wrapper is still pure file I/O (read frontmatter, write
a `.md` file); inventing a primitive for it would imply a capability
neither runtime actually requires beyond what ADR-001 already grants,
and would invite scope creep toward runtime-specific APIs the original
decision explicitly rejects.

**A new companion ADR** — rejected. A new ADR would wrongly imply a new
architectural pillar sitting alongside the skill-first decision, when in
fact this is a narrow, additive refinement of HOW file I/O is applied
for one runtime's UI. Amending ADR-001 keeps the "exactly two
primitives" thesis as the single source of truth and records this as
what it is: a clarification, not a new pillar.

**A no-op `CommandAdapter` entry for Claude Code** — rejected, for the
same reason `Providers` carries no "null provider": an entry that exists
solely to do nothing contradicts this codebase's zero-special-casing,
absence-is-the-no-op idiom.

### Consequences

- The "exactly two primitives" thesis from the original decision stands
  unchanged — this addendum documents a file-layout NUANCE within file
  I/O, not a new capability.
- New assistants that need their own discoverability artifacts are
  supported by appending ONE entry to `CommandAdapters` — zero edits to
  `Install`, `installOne`, `copyEntry`, `writeSkillFile`, or any existing
  registry entry.
- Generated wrapper content derives 100% structurally from each
  specialist's own `SKILL.md` frontmatter (`description`, `name`,
  `specialist-id`); there are no hardcoded per-specialist tables to keep
  in sync.
- **Test limitation**: there is no live OpenCode runtime in this
  repository or its CI. Correctness of the generated wrappers (frontmatter
  shape, `agent: build`, `subtask: false`, deterministic content, and
  idempotent re-writes) is validated STRUCTURALLY — by asserting on the
  generated file paths and content byte-for-byte — not by exercising
  OpenCode's actual command-palette UI end to end.
