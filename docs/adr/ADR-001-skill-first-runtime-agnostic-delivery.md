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
