# ADR-009 — Deterministic project-level platform initialization

**Status**: Accepted

---

## Context

Each specialist run re-derived platform knowledge via an LLM `platform-analysis` step,
and `loadPlatformContext` read raw `platform.yaml` on every step. This caused:

- **Recurring token cost** — every specialist run burned LLM tokens to re-derive stack
  information that does not change between runs.
- **Non-deterministic output** — LLM-derived stack summaries varied between runs for the
  same codebase, producing inconsistent platform context across specialists.
- **Conceptual collision** — per-change `platform-summary.yaml` artifacts written under
  `.asdt/artifacts/{change}/` conflated change-scoped work with project-wide facts.

A fully tested `internal/knowledge/` Go scanner already existed but was unhooked after
its original CLI subcommand was removed.

---

## Decision

Introduce `asdt init` as a deterministic, zero-token project initialization command.

**`asdt init` does the following:**

1. Runs the existing `knowledge.DefaultDetector().Detect()` Go scanner against the
   project root (no LLM involved).
2. Writes `.asdt/knowledge/platform.yaml` with the raw scan output.
3. Derives a `PlatformSummary` via `knowledge.DeriveSummary()` — a pure function with
   a static lookup table (language → package manager + test runner).
4. Writes `.asdt/knowledge/platform-summary.yaml` — the canonical, project-level
   platform digest.

**Specialist integration:**

- `WorkflowStep` gains a `SkipIfInitialized bool` field (zero value = false = backward
  compatible).
- Steps flagged `SkipIfInitialized: true` skip their LLM call when
  `.asdt/knowledge/platform-summary.yaml` exists, emitting the summary file contents as
  the step's output artifact instead (with `source: platform-summary.yaml`).
- The following platform-scan steps are flagged:

  | Specialist | Step              |
  |------------|-------------------|
  | UX-UI      | `platform-analysis` |
  | Security   | `platform-analysis` |
  | Architect  | `platform-analysis` |

- `loadPlatformContext` is updated to prefer the deterministic summary, falling back to
  raw `platform.yaml`, then to empty string.

**Path isolation:**

`.asdt/knowledge/platform-summary.yaml` (project-level, written by `asdt init`) is
a distinct path from `.asdt/artifacts/{change}/platform-summary.yaml`
(change-scoped, written by LLM steps). No init operation writes into `artifacts/`.

---

## Consequences

### Positive

- **Zero token cost** for platform context after `asdt init` is run once per project.
- **Reproducible output** — the same project always produces the same `platform-summary.yaml`
  (modulo `detected_at` timestamp).
- **Reuses the existing tested knowledge scanner** — minimal new surface area (one pure
  derive function, one writer, one CLI case, one struct field, one runner branch).
- **Backward compatible** — absent summary → identical prior behavior (`SkipIfInitialized`
  zero value false; fallback chain intact).

### Negative / Tradeoffs

- **Stack-only summary this iteration** — `package_manager` and `test_runner` are derived
  from a static language-to-tooling table, not file inspection (yarn/pnpm detection,
  monorepo layouts, etc. are deferred to a future `--enrich` / file-walk enhancement).
- **Two platform files** — contributors must know that `platform-summary.yaml` is the
  canonical reuse source and that `platform.yaml` is the raw scan backing it.
- **Opt-in** — `asdt init` is not run automatically; teams must call it once per project
  (or add it to onboarding scripts).

---

## Alternatives Rejected

| Alternative | Reason rejected |
|-------------|-----------------|
| LLM enrichment during init | Blocked by unimplemented `AnthropicProvider`; defeats determinism goal; token cost remains. |
| Reuse change-scoped artifact | Wrong lifetime — per-change artifacts collide across changes and are deleted on change cleanup. |
| Keep re-scanning per run | The status quo cost this ADR removes; no path to deterministic, zero-token context. |

---

## Related

- ADR-006: Specialist model — introduced the platform-analysis step pattern this ADR optimizes.
- ADR-008: Context-isolated skill steps — the `SkipIfInitialized` flag follows the same
  context-isolation philosophy.
- `internal/knowledge/` — Go scanner package providing `DefaultDetector` and `DeriveSummary`.
- `cmd/asdt/main.go` — CLI wiring for `asdt init`.
