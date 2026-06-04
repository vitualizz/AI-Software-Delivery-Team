# ADR-006: Specialist Model

Date: 2026-06-04
Status: Accepted

## Context

The MVP modeled software delivery as a single linear FSM (requirements → plan →
implement → review) with task-verb subcommands and specialist knowledge baked into
Go packages (`internal/requirements`, `internal/developer`). Adding a new role
required a new Go package, a new struct, and a new switch arm in `cmd/asdt/main.go`
— not prompt authoring. The FSM hardcoded `requirements` as the only valid entry
point, so a UX designer or Security engineer had no place in the model without
extending the fixed graph.

This is the wrong conceptual model: real software delivery is performed by a team
of specialists, each owning an independent discipline. A security engineer does not
wait for a developer to finish before reviewing auth code. A UX designer does not
follow a requirements → plan workflow — they follow their own creative process.
Forcing every discipline into the same four-phase pipeline produces an awkward model
that fits none of them well.

Four concrete problems with the MVP model:
1. **Hardcoded FSM**: `validEdges` in `internal/pipeline/phase.go` was the only
   valid topology. A specialist that does not fit `requirements → plan → implement →
   review` had no valid pipeline position.
2. **Go package per specialist**: Adding a role = new Go package + new struct + new
   switch arm. This is code, not prompt authoring.
3. **Task-verb subcommands**: `asdt requirements`, `asdt develop` modeled phases, not
   roles. Users think "I need the developer to do X", not "I need to run the implement
   phase".
4. **No shared-skill architecture**: Common capabilities (platform context, artifact
   envelope, scope definition) were duplicated or implicit across specialists.

## Decision

A Specialist is a composable, independent unit defined by:

- **Identity**: `ID` (stable key), `Name` (human label), `Description` (trigger keywords for routing)
- **Workflow**: ordered steps SPECIFIC to that discipline — not a generic pipeline applied uniformly to all roles
- **Skill composition**: shared skills loaded for every step, plus specialist-scoped skills loaded per step
- **Artifact contract**: `Reads` (all soft — missing inputs degrade to `open_items[]`, never error), `Writes` (specialist-specific artifact types)
- **Independence guarantee**: any specialist may run first — there is no required predecessor

One generic `SpecialistRunner` in `internal/specialists/runner.go` executes any
`SpecialistDescriptor`. Adding a specialist requires exactly: one `SpecialistDescriptor`
value literal + one `skill/{id}/SKILL.md` tree. Zero new Go packages, zero new switch
arms.

The five built-in specialists are: Developer, UX/UI, Architect, QA, Security.

## Alternatives Considered

**Keep one Go package per specialist (status quo)** — rejected. Every new role
requires Go code changes. This violates the "prompt authoring, not Go development"
goal of ADR-001.

**Generalize the existing linear FSM to a configurable DAG shared by all specialists**
— rejected. A DAG still imposes a shared graph topology and a shared execution order.
A UX specialist does not do "requirements → plan → implement". The specialist model
gives each role its own workflow rather than fitting every role into one shared graph.

**A plugin system (Go plugins or external processes per specialist)** — rejected.
Runtime-loading fragility, massive complexity, and it breaks the single-binary +
runtime-agnostic invariant from ADR-001 for no benefit over data-driven descriptors.

**Event-sourced coordinator** — rejected. Over-engineering for a local-first MVP.
ASDT state is always knowable from one file per change; event sourcing adds
reconstruction complexity without adding value here.

## Consequences

Positive:
- Contributors can add specialists via prompt files alone — no Go expertise required for new roles.
- Each specialist is independently invocable (`/asdt:developer`, `/asdt:security`, etc.) without requiring a predecessor to run first.
- Specialist coupling is structurally impossible: specialists communicate only through artifact files under `.asdt/artifacts/{change}/`, never via Go imports.
- The meta-orchestrator (`skill/SKILL.md`) routes requests to specialists rather than executing any discipline-specific logic itself.

Negative:
- `SpecialistRunner` writes `Envelope[map[string]any]` — schema validation is
  deferred to the verify stage rather than happening at compile time via typed Go
  payload structs. Per-artifact schemas require explicit YAML schema files under
  `schemas/` for each new artifact type.
- Ordering correctness depends on `SpecialistDescriptor.Validate()` and the runner
  loop rather than an enforced edge map. A descriptor with an ill-ordered workflow
  produces a confusing run, not a startup error.
