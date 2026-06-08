# ADR-002: `.asdt/` as the Clean Boundary

Date: 2026-06-04
Status: Accepted

## Context

ASDT operates inside existing projects. It must not pollute the project root,
interfere with existing files, or require any changes to the host project's
structure. Developers should be able to adopt ASDT in an existing repository
without a migration step and remove it without a trace.

Git's own discovery model (`.git/` walk-up from CWD) is the proven pattern for
per-project tooling state. Every developer already understands it.

## Decision

All ASDT state — artifacts, knowledge, config, pipeline state, and prompt
overrides — lives exclusively under a single `.asdt/` directory at the project
root. Discovery uses git-style CWD walk-up: start from the current working
directory, walk up toward the filesystem root, stop at the first directory
containing `.asdt/`. If none is found, create `.asdt/` at the nearest project
marker (`.git`, `go.mod`, `package.json`).

Nothing is ever written outside `.asdt/`. This is an invariant, not a
convention.

## Alternatives Considered

**`~/.config/asdt/projects/{hash}/`** — rejected. State is not project-local;
breaks on multi-project setups where the same change name could collide across
projects; requires a hashing scheme that obscures which project owns which
state.

**`.agent/`** — rejected. Conflicts with other tooling (some AI-assistant
frameworks also claim `.agent/`). ASDT needs a name it owns unambiguously.

**Inline in project files** — rejected. Contaminates the project's own
structure; not separable; cannot be `.gitignore`d without affecting project
files.

**Session/runtime memory only** — rejected. Evaporates when the chat ends.
ASDT's core thesis is that artifacts are durable records, not ephemeral context.

## Consequences

- Trivially uninstallable: `rm -rf .asdt/` leaves the project exactly as it was.
- Monorepo-friendly: each sub-project can have its own `.asdt/` if needed.
- Teams can choose to commit `.asdt/` for shared artifact history or add it to
  `.gitignore` for local-only use — ASDT does not enforce either.
- The TUI has exactly one place to look for all state.
- All paths in artifact envelopes are `.asdt/`-relative, making them portable
  across machines and clones.

## Addendum (2026-06-07): Scope clarification — bookkeeping vs. host-source writes

This ADR's "nothing is ever written outside `.asdt/`" invariant governs ASDT's OWN tool
bookkeeping: config, knowledge cache, pipeline state, prompt overrides, and artifact handoff
(the last now superseded by Engram per ADR-011). Its purpose is uninstallability
(`rm -rf .asdt/` leaves the project untouched) for ASDT's state — NOT a prohibition on a
specialist whose declared role is to modify the host source tree.

The `asdt-developer` specialist, in writing mode, writes real code to the host project's
source tree. Those writes are NOT governed by this ADR; they are governed by `asdt-developer`'s
own write-scope contract (declared edit roots resolved from `dev-tasks`/`dev-design`, validated
before writing, STOP-on-out-of-scope — the sdd-apply model). Such host-source writes are a
deliberate, scoped product of an approved task, not ASDT bookkeeping, and are reversible via
normal version control rather than via `.asdt/` removal.

In short: `.asdt/`-only = ASDT's tool state. Declared edit roots = a code-producing specialist's
host writes. These are two distinct write surfaces; this ADR governs only the first.
