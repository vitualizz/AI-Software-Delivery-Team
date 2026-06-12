---
title: Architect
description: Makes architecture decisions and produces ADRs, system design, and API design artifacts — the specialist to bring in when a choice will shape service boundaries, data models, or scalability for the long haul.
order: 21
locale: en
---

# Architect (`/asdt-architect`)

> Makes architecture decisions and produces ADRs, system design, and API design artifacts — the specialist to bring in when a choice will shape service boundaries, data models, or scalability for the long haul.

## What it does

The Architect Specialist makes the technical decisions that everything else is built on. It evaluates competing approaches, documents the chosen path as an Architecture Decision Record (ADR), and produces a concrete system design with data models, API surfaces, and service boundaries — all before a single line of implementation code is written.

Every decision produced by the Architect comes with alternatives considered and consequences documented — including negative consequences. A decision record with only positive consequences is incomplete. This forces honest trade-off analysis instead of post-hoc justification.

The Architect Specialist never writes implementation code, UX specs, or test plans. Its one job is to make the structural decision that the Developer can build against without ambiguity.

## When to invoke it

- A decision will shape service boundaries, data models, or scalability beyond the current feature
- The technical approach is non-obvious and has meaningful trade-offs between at least two viable options
- A cross-cutting concern (caching strategy, auth model, event bus) needs a documented decision
- You want a formal ADR to explain to future engineers why the code is the way it is

## Pipeline position

Typically runs **after PM** (reads `backlog-entry`) and **before Developer** (Developer reads `architectural-decision` + `system-design`). At `simple` complexity it is not called at all — the Developer handles it directly. At `trivial`, it runs a single `load-constraints` consult. At `moderate` and `complex`, it runs a full evaluation and produces the complete architectural artifact set.

## What it produces

Two final artifacts consumed by downstream specialists:

- **`architectural-decision`** — the ADR with full context, decision, alternatives, consequences, and key constraints the Developer must not violate
- **`system-design`** — data model, API surface, service boundaries, key sequence, and top risks

Consumed by: **Developer** (reads both), **QA** (reads `architectural-decision` to understand design context).

## Common patterns

```
/asdt-architect Design the rate-limiting strategy for the public API
# → Cross-cutting concern that will affect every endpoint
```

```
/asdt-architect Choose the event sourcing approach for the order pipeline
# → Non-reversible structural decision with meaningful trade-offs
```

```
/asdt-architect ADR for switching from REST to GraphQL on the mobile client
# → External contract change that needs documented rationale
```

## Limits — what it does NOT do

- Does not write implementation code
- Does not write UX specs or wireframes
- Does not produce test plans or acceptance criteria
- Never skips alternatives — every decision record requires them
- Does not design in isolation — always accounts for existing platform constraints
- System design is always incomplete without both a data model AND an API surface
