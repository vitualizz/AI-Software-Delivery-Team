---
title: Specialist Model
description: How ASDT models software delivery as a team of independent specialists, each owning a discipline.
order: 6
locale: en
---

# Specialist Model

## Why specialists, not a pipeline

The first version of ASDT modeled software delivery as a fixed four-phase FSM: `requirements → plan → implement → review`. Adding a new role required a new Go package, a new struct, and a new switch arm — code, not prompt authoring. The FSM hardcoded `requirements` as the only valid entry point, so a security engineer or UX designer had no valid place in the model without restructuring the entire graph.

This is the wrong model. Real software delivery is performed by a team of specialists, each owning an independent discipline. A security engineer doesn't wait for a developer to finish before reviewing auth code. A UX designer doesn't follow a requirements → plan workflow — they follow their own creative process.

**ADR-006** formalized the switch: a Specialist is a composable, independent unit defined by its identity, its own workflow steps, its artifact contract, and an independence guarantee — any specialist may run first.

## What defines a specialist

A specialist has four parts:

**Identity** — a stable `id` (e.g. `developer`), a human name, and a description that the pipeline advisor uses to route requests.

**Workflow** — an ordered list of steps specific to that discipline. The Developer runs `explore → spec → design → implement`. The UX/UI specialist runs `feature-brief → information-architecture → user-flows → component-mapping → ux-handoff`. These are not the same pipeline applied to different names — each specialist's workflow reflects how that discipline actually works.

**Skill composition** — shared skills loaded for every step (platform context, artifact envelope, scope definition), plus specialist-scoped skills loaded per step (threat modeling for Security, code generation for Developer). Capabilities are mixed in rather than inherited.

**Artifact contract** — what the specialist reads (`inputs`) and writes (`outputs`). Inputs are soft: a missing input degrades to an `open_items[]` note, never an error. Outputs have stable `topic_key` values so other specialists can retrieve them by key.

## Adding a specialist

Adding a new specialist requires exactly two things:

1. One `SpecialistDescriptor` value literal in the registry
2. One `skill/{id}/SKILL.md` tree with step files

Zero new Go packages, zero new switch arms. The `asdt-*` embed glob in `skill/embedded.go` picks up any directory matching the pattern and ships it in the next build. See [Contributing](/docs/contributing) for the full authoring contract.

## The independence guarantee

Any specialist may run first — there is no required predecessor. If the Developer finds no Architect artifact in Engram, it proceeds with `open_items: ["architect/adr not found"]` and makes reasonable assumptions. The resulting implementation artifact is less precise than if the Architect had run first, but it's valid output.

This design choice prioritizes flexibility over correctness guarantees. You can always run specialists out of order. ASDT trusts you to decide when to involve each discipline.
