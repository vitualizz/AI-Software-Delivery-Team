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
