# ADR-013 — Skill-Only Context Detection for asdt-init

Date: 2026-06-11
Status: Accepted

---

## Context

`asdt-init` Step 4 instructs the executing sub-agent to invoke the Go context
detector directly as a library:

```go
det := knowledge.DefaultContextDetector(primaryLang)
ctx, _ := det.DetectContext(projectRoot, knowledge.DetectConfig{...})
```

This is impossible from a target repository:

- `internal/knowledge` cannot be imported from outside this module — Go's
  `internal/` rule forbids it unconditionally.
- No distributed binary exposes it. The only binary is `cmd/asdt-tui` (the
  skill installer), and nothing in `cmd/` or `internal/installer/` calls
  `DetectContext` or `WriteContext`.
- ADR-009 wired the scanner to a `cmd/asdt` CLI that has since been removed.
  This is the second time the Go scanner has been orphaned by a binary
  removal.

In practice, the LLM running the skill hand-simulates the documented rules.
The determinism guarantee the design promises is therefore fiction at
runtime. Field incident: re-running `asdt-init` on `sperant-ia` (Python
`backend/` + Node `frontend/`) degraded `project-context.yaml` from good
values (`pytest`/high, `modular`/medium) to all-`unknown`, because the
emulated rules combine a root-only stack scan with go/node-only probes — and
the recalibration flow let `unknown/low` overwrite `detected/high` via a
blanket "accept all".

Human decision (Lee Palacios, 2026-06-11): **asdt-init must not depend on any
code, Go or otherwise. Go in this project exists exclusively for the TUI that
installs and customizes skills. On every other machine, only the skills run.**
This re-anchors on ADR-001: "The Go binary [...] cannot own any agent logic.
If the binary and the skill diverge on behavior, the skill wins."

## Decision

All project-context detection executes as **exact, bounded shell commands
documented in `skill/asdt-init/SKILL.md`** — runnable identically by any
model on any machine where the skills are installed. No Go invocation appears
anywhere in any skill.

1. **One rule = one command + one exact mapping table.** Every probe is
   expressed the way Step 1 already is: a single bounded command whose output
   maps to values with zero model judgment. This is the determinism mechanism
   that replaces compiled code.
2. **Deep stack scan (US-1).** The Step 1 marker scan deepens from `-d 1` to
   `-d 3` with a fixed exclusion list (`node_modules`, `.git`, `vendor`,
   `dist`, `build`, `.venv`, `target`) and a fixed result cap. Ordering is
   deterministic: shallowest path first, then lexicographic. `detected_stack`
   records every detected language; the primary language is the first entry
   under that ordering.
3. **Python probe rules (US-2).** `test_runner` and `naming_style` gain
   Python rules (pytest config detection, `snake_case` functions /
   `PascalCase` classes sampling), sourced from the same language→tooling
   table `platform.go` uses, ported into the skill as the single live copy.
4. **Recalibration downgrade guard (US-3).** "Accept all" never covers a
   field whose new confidence is lower than the old, whose new value is
   `unknown`, or whose old source is `manual` — those always require
   per-field confirmation. Before any overwrite, the previous file is
   preserved as `project-context.yaml.bak` (single rotating backup).
5. **Monorepo heuristic (US-4).** Two or more non-excluded directories at
   depth ≤ 2 each holding a distinct stack marker → `is_monorepo: true`,
   `detected`, `medium`.
6. **Negative-evidence confidence cap.** A value derived from the *absence*
   of evidence (e.g. `is_monorepo: false` because no marker was found) caps
   at `medium` confidence, never `high`.
7. **The orphaned Go detector is scheduled for removal.** `context_probe.go`,
   `context_detector.go`, and their tests leave the codebase in a follow-up
   change, after their rule tables are fully ported into `SKILL.md`. Go keeps
   only what the TUI installer needs.

## Alternatives Considered

**Binary subcommand (`asdt detect-context`) as the execution path** —
rejected. Violates ADR-001 (a binary owning agent logic), requires
distributing and version-managing a binary on every target machine, and this
exact approach has already failed twice: the ADR-009 CLI was removed, and the
library it stranded caused this incident.

**Hybrid: binary preferred, shell fallback** — rejected. Two code paths for
one contract invites version skew between installed binary and installed
skill, and still ships agent logic in the binary. The honest-degradation
semantics it offered (compiled evidence outranks shell evidence) are not
needed once shell commands ARE the contract: per the existing §4.4 table,
`detected` means "bounded command with direct file evidence", which the
shell commands satisfy.

**Status quo (document the Go library, let the model emulate it)** —
rejected. This is the bug: an unreachable implementation emulated by an LLM
and presented as deterministic detection.

## Consequences

- Detection works on any machine where the skills are installed — no
  toolchain, no binary, no version skew. Prompt-only contributions can change
  detection behavior (per ADR-001's consequences).
- `SKILL.md` becomes the single source of truth for detection rules; the
  doc/code drift class of bugs disappears with the dead code.
- Determinism is now enforced by command discipline (one bounded command,
  one exact mapping) instead of compiled code — weaker in the limit, but
  honest: it describes what actually executes.
- The Go unit tests covering the probes are deleted with the probes; rule
  correctness is validated by running `asdt-init` against reference repos
  (sperant-ia is the first acceptance case).
- Skill commands may only assume tools the runtimes guarantee; where `fd` is
  used, the documented command must state the equivalent `find` fallback.

## Related

- ADR-001 — Skill-First, Runtime-Agnostic Delivery: the doctrine this
  decision re-anchors on.
- ADR-009 — Deterministic Platform Initialization: its goals (deterministic,
  zero-token, bounded init) stand; its execution model (Go binary) is
  superseded by this ADR.
- Engram artifacts under `asdt-init-deep-detection/` (pm/* and architect/*):
  the incident analysis, user stories, and risk register behind this
  decision.
