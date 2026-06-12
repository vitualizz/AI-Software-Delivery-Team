# ADR-015 — Make asdt-init Delegation Conformant via a Net-New workflow.yaml and an Explicit analyst Default

> **Amended by ADR-016** — the single-builder-step sub-decision (REQ-1) is superseded by ADR-016's three-step explore/clarify/write flow; the workflow.yaml-exists and analyst-default rulings remain in force.

Date: 2026-06-12
Status: Accepted (amended by ADR-016)

---

## Context

During a real `/asdt-init` recalibration run, the orchestrator launched the
sub-agent as `general-purpose` instead of `asdt-analyst` / `asdt-builder`,
silently dropping the tool guardrails those agent types carry.

Root cause: `asdt-init` had no `workflow.yaml`. The specialist-header routing
rule (`agent: analyst` → `asdt-analyst`, `agent: builder` → `asdt-builder`,
`general-purpose` ONLY as an availability fallback) had no entry to read, and
the harness fell back to its default agent type.

This was a structural asymmetry: all six other specialists ship a
`workflow.yaml`; init did not. The routing rule cannot resolve an `agent:`
field that no file declares.

## Decision

**Approach A — make init conformant with the specialist pattern:**

1. Add a net-new `skill/asdt-init/workflow.yaml` so init's delegation is
   `agent:`-routed exactly like every other specialist.
2. Add an explicit default rule to `parallel-retrieval.md`: a `subagent` step
   WITHOUT an `agent:` field defaults to `analyst` (read-only, safe) — never
   silently `general-purpose`.

### Sub-decisions

- **REQ-1** *(amended by ADR-016)*: init delegates as ONE `builder` step
  (detect + write + confirm). The Engram gate stays inline with the
  orchestrator — it inspects its OWN tool list and is undelegatable.
- **REQ-5a**: strip the duplicate `## Identity` heading from
  `agents-template.md` (the persona files keep it).
- **REQ-3**: write the Delegation Contract section user-facing, matching the
  Non-Negotiables voice.
- **Riders**: REQ-4 (PM row missing from the specialists table), REQ-5b
  (Session Protocol mentions `project-context.yaml`), REQ-5c (executor-header
  note for ad-hoc delegations).

## Alternatives Considered

**B — inline agent declaration in SKILL.md prose without a workflow.yaml** —
rejected. Leaves init a permanent special case; the `agent:` route still has
no entry to resolve, so this papers over the root cause instead of removing it.

**C — Go change / new agent type** — rejected. Hits two hard constraints:
`AgentTypeNames` is hardcoded to `[analyst, builder]`, and the executor header
is baked at install time. High coupling and low reversibility for what is a
config-and-docs gap.

## Consequences

**Positive:**

- init is conformant with the specialist pattern — the special case is killed.
- The analyst-default rule guards every future workflow author.
- Purely additive: no Go, fully reversible.

**Negative:**

- Two coordinated edits must land together.
- The `parallel-retrieval` wording must not perturb the executor-header
  injection logic.
- Template changes require a reinstall to reach live configs.

**Debt:**

- Registry/template information is duplicated across files (manual sync).

> **Implementation note**: the riders (the `parallel-retrieval` analyst-default
> line, the `## Identity` dedup, the Delegation Contract, the PM row) were
> *decided* here but are *implemented* under the `agent-template-reinforcement`
> change, not `init-empowerment`. This ADR records the decision; do not claim
> the riders ship with the init-empowerment change.

## Related

- ADR-016 — Init Empowerment: amends this ADR (REQ-1 superseded; the
  workflow.yaml-exists and analyst-default rulings remain in force).
- ADR-014 — SKILL.md Self-Load + Inline Orchestrator Gate: established the
  specialist-header routing rule this ADR makes init obey.
