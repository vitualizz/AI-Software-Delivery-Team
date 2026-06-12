# ADR-014 — SKILL.md Self-Load of specialist-header + Inline Orchestrator Gate

Date: 2026-06-12
Status: Accepted

---

## Context

Every specialist `SKILL.md` declares a `shared-skills:` frontmatter key whose
first entry is `specialist-header`, and until this change each body opened
with a "Fallback guard" blockquote that aborted "if `specialist-header` was
not loaded before this file". Both rested on the assumption that some loader
resolves `shared-skills:` and injects `specialist-header.md` into context
ahead of the `SKILL.md` body.

No such loader exists:

- **Claude Code injects only the `SKILL.md` body** when a skill is invoked.
  It does not interpret custom frontmatter keys; `shared-skills:` is inert
  metadata to it.
- **The installer copies the skill tree verbatim.** Nothing in
  `internal/installer/` expands, inlines, or rewires shared-skill
  references at install time.
- **`internal/installer/adapters.go` (`parseSpecialistFrontmatter`) reads
  only `specialist-id`, `name`, and `description`** to generate the OpenCode
  wrapper. `shared-skills:` is never parsed by any code path.

Consequently the ORCHESTRATOR GATE, the Prerequisites check, and the
Tailored Workflow detection living in `specialist-header.md` never reached
the orchestrator. The Fallback guard could not fire either — the model has
no way to observe that a file "was not loaded", so in practice specialists
ran without the gate and were prone to executing `subagent` steps inline.

A previous remediation attempt was recorded only in Engram under the label
"ADR-011": it proposed solving the duplication by *transcluding* the header
via the `shared-skills` mechanism — the exact mechanism that does not exist.
That Engram-only record also **collides in number with the on-disk
[ADR-011](ADR-011-specialist-pipelines-as-orchestration-plans.md)**
(Specialist Pipelines as Orchestration Plans), which is the authoritative
ADR-011 and whose Decision item 1 already states that the gate lives inline
in each `SKILL.md`. This ADR restores the on-disk ADR-011 position and
supersedes the Engram-only record.

## Decision

1. **Self-load via explicit Read.** The first content of every specialist
   `SKILL.md` body is a `FIRST ACTION` blockquote instructing the assistant
   to Read `../asdt-shared/skills/specialist-header.md` and `./workflow.yaml`
   before acting, and to re-read them whenever their content can no longer
   be recalled (e.g. after a context compaction). Explicit tool use by the
   model replaces the imaginary loader.
2. **Minimal inline gate.** Immediately below, a compact `ORCHESTRATOR GATE
   (inline copy)` blockquote carries the survival-critical rule — the calling
   assistant is the sole orchestrator; `subagent` steps launch via the native
   delegation primitive, never inline; sub-agents answer to their injected
   executor header. This copy is deliberately small (~5 wrapped lines) so it
   survives compaction; the full protocol remains ONLY in
   `specialist-header.md`.
3. **Byte-identical across specialists.** The replacement block is identical
   in all six specialist `SKILL.md` files (relative paths resolve from every
   specialist directory), keeping drift detectable by simple diffing.
4. **`shared-skills:` is kept, reclassified as metadata.** The key remains in
   frontmatter as documentation of related skills, but no document may claim
   it is a loading mechanism. `skill/asdt-shared/skills/README.md`,
   `platform-context.md`, and the site contributing guide are corrected
   accordingly; the real injection mechanisms are the FIRST ACTION Read
   (specialist-header) and per-step `reference_skills` in `workflow.yaml`.

## Alternatives Considered

**Make the installer expand `shared-skills:` into the SKILL.md body** —
rejected. Ships agent logic in the binary (violates ADR-001), creates skew
between source skills and installed skills, and does nothing for runtimes
that load skills without the installer.

**Inline the full specialist-header into all six SKILL.md files** —
rejected. Six divergent copies of ~50 lines is the duplication problem the
header file was created to solve; the minimal-inline-copy + canonical-file
split keeps one source of truth while guaranteeing the gate is present.

**Rename or remove the `shared-skills:` frontmatter key** — rejected for
this change. It is harmless as metadata, removal would churn all six files
plus templates and docs, and a future loader could legitimately adopt it.

## Consequences

- The gate and prerequisites now actually execute: they travel in the
  injected body (inline copy) and via an explicit Read the model performs
  itself (full version).
- Compaction resilience is explicit — the FIRST ACTION instruction mandates
  re-reading when header content is no longer recallable.
- Documentation no longer advertises a fictional loading mechanism, closing
  a doc/runtime drift of the same class ADR-013 closed for context
  detection.
- The Engram-only "ADR-011" transclusion record is superseded; the number
  collision is documented in the ADR index Gap Notes so future readers do
  not mistake it for the on-disk ADR-011.

## Related

- ADR-011 — Specialist Pipelines as Orchestration Plans: Decision item 1
  (inline gate in each SKILL.md) is restored by this ADR.
- ADR-001 — Skill-First, Runtime-Agnostic Delivery: why the fix lives in
  prompts, not in the installer binary.
- ADR-013 — Skill-Only Context Detection: precedent for removing behavior
  that documentation promised but no runtime executed.
