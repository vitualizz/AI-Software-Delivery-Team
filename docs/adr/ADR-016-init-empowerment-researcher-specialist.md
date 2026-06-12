# ADR-016 — Init Empowerment: Three-Step Init Flow and the Researcher Specialist

Date: 2026-06-12
Status: Accepted

---

## Context

The incident chain: `general-purpose` fallback on a real `/asdt-init` run →
[ADR-015](ADR-015-init-delegation-conformance.md) ruled init delegates as a
single `builder` step → the owner pivoted (2026-06-12) to an
explore/clarify/write shape so init can PAUSE and ASK the human before writing
config. A second track surfaced in parallel: a Researcher specialist whose
discovery brief must reach PM's no-declared-inputs `feature-intake`.

The constraining physics of the host-agnostic harness:

- **Only `inline` steps run inside the interactive orchestrator turn.** They
  can pause for the human, but cannot persist an artifact.
- **Only `subagent` steps persist artifacts.** They cannot pause.

A single builder step therefore cannot both detect the stack AND stop to ask
the human a clarifying question. The work must split across the two execution
modes.

## Decision

### 1. init = explore · clarify · write

- **explore** — `subagent` / `analyst` / `haiku` → `init/stack-detection`
  carrying stack + per-field confidence + `ambiguities[]`. Read-only; never
  guesses.
- **clarify** — INLINE first-class named step: no `agent`, no `model`, no
  `topic_key`, no artifact. Asks ONE question at a time in prose; an explicit
  SKIP path for non-interactive harnesses fills per-field defaults; answers are
  injected into write's prompt.
- **write** — `subagent` / `builder` / `sonnet` ← `stack-detection` →
  `init/write-summary`. Writes the four `.asdt` files.

### 2. Researcher specialist (5 steps)

`context-recall` (INLINE — reuses `knowledge-recall.md` verbatim) ·
`divergent-ideation` (`analyst` / `sonnet`, `inputs: []`, trivial-eligible) ·
`feasibility-scan` · `discovery-brief` (references `report.md`) ·
`decision-preservation` (INLINE).

### 3. Handoff

The discovery brief is rendered as prose RAW REQUEST into PM's
`feature-intake` (`inputs: []` — zero contract change). The belt is one SOFT
`knowledge-recall` line, never a required input.

### 4. init is STANDALONE

init is a setup-class command, NOT routable in `skill/SKILL.md` §9.2. No
feature request ever routes to init. This is documented so future maintainers
do not "fix" the omission.

### 5. Step names

`explore` and `write` are ratified as the init step names.

### 6. Amendment mechanics

ADR-015's body is preserved and a banner added. ADRs are permanent records —
never silently rewritten.

### 7. R-001 reconciliation — write-boundary verification

Write-boundary verification replaces any checkpoint file. The
`### CLARIFY ANSWERS` block is REQUIRED in write's prompt even when empty
(`answers: {}`, `skipped: true|false`, `blocking_open_items: []`). Absent →
write HALTS with zero writes. `blocking_open_items[]` non-empty → HALT.
`write-summary.applied_answers[]` (field → value → origin `user|default`) is the
auditable outcome trail.

## Alternatives Considered

**clarify-as-implicit-glue** — rejected. Violates the mandated three-named-step
shape; invisible in `workflow.yaml`; untestable.

**builder-subagent questions-file round-trip** — rejected. Subagents cannot
pause; a stateful two-pass is forbidden.

**init as a routable §9.2 specialist** — rejected. The meta-orchestrator routes
feature requests; tiers are meaningless for one-shot setup.

## Consequences

**Positive:**

- Incident guardrails restored.
- Host-agnostic graceful degradation (non-interactive SKIP path).
- Zero-contract-change PM integration.
- Proven patterns reused (knowledge-recall, report, decision-preservation).
- Legible amendment chain.

**Negative:**

- clarify Q&A is unauditable by design (no transcript artifact).
- Two new `workflow.yaml` ↔ `SKILL.md` sync hazards.
- Registry triplication (three places to keep in sync).
- Raw-text handoff is lossy by design.
- Amendment chain adds reading overhead.
- init sits outside meta-routing — a discoverability cost.

**Debt:**

- Manual parity mirrors (no tooling).
- No clarify transcript artifact (a future opt-in remains possible).
- The SOFT belt line must not harden into a required input.

## Related

- **Amends [ADR-015](ADR-015-init-delegation-conformance.md)** — supersedes its
  REQ-1 single-builder-step sub-decision; the workflow.yaml-exists and
  analyst-default rulings remain in force.
- ADR-011 — Specialist Pipelines as Orchestration Plans: the orchestration
  model the new workflows obey.
- ADR-013 — Skill-Only Context Detection: init's probes run as bounded shell
  commands with no dependency on this repo's Go code.
