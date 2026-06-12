---
title: Product Manager
description: Transforms raw feature requests into structured backlog entries with user stories, scope boundaries, and prioritization — the specialist to bring in before architecture or code when requirements need formalization.
order: 20
locale: en
---

# Product Manager (`/asdt-pm`)

> Transforms raw feature requests into structured backlog entries with user stories, scope boundaries, and prioritization — the specialist to bring in before architecture or code when requirements need formalization.

## What it does

The PM Specialist turns vague feature requests into a precise requirements artifact that every other specialist can consume without ambiguity. It extracts the core problem, identifies stakeholders, writes user stories with preliminary acceptance criteria, defines explicit scope boundaries, and consolidates everything into a single `backlog-entry` that flows downstream.

Two properties make the backlog-entry contract strict: scope boundaries are **mandatory** — a backlog-entry without explicit out-of-scope items is considered incomplete, because scope ambiguity is the root cause of most scope creep. And the acceptance criteria in the backlog-entry are high-level plain English conditions — **not** final testable criteria. QA formalizes them into Given/When/Then format.

The PM Specialist never writes architecture decisions, implementation code, or UX specs. Its one job is to make requirements unambiguous so that no downstream specialist has to guess.

## When to invoke it

- The request is phrased in vague or user-facing language ("add dark mode", "improve search")
- You need explicit user stories before the Architect or Developer gets involved
- Scope needs to be locked down before work starts to prevent mid-sprint expansion
- Multiple stakeholders are involved and their needs need to be reconciled

## Pipeline position

Works best as the **first** specialist in a pipeline — its `backlog-entry` is the primary requirements source for Architect, Developer, and QA. Running it after architecture is already decided risks scope/requirements drift. Can run standalone when you only need formalized requirements without proceeding further.

## What it produces

`pm/backlog-entry` — the canonical requirements artifact. Contains: feature name, executive summary, ordered user stories with acceptance criteria, full scope block (in/out of scope, integration points, risk flags), and open items for downstream specialists.

Consumed by: **Architect** (reads executive summary + scope), **Developer** (reads user stories + priority order), **QA** (reads user stories + acceptance criteria as the primary requirements source).

## Common patterns

```
/asdt-pm Add user authentication with email and password
# → Ambiguous requirements, needs scope before architecture
```

```
/asdt-pm Redesign the notification system
# → Multiple stakeholders with potentially competing needs
```

```
/asdt-pm Add CSV export to the reporting dashboard
# → Simple on the surface, but integration points and scope risk need mapping
```

## Limits — what it does NOT do

- Does not write architecture decisions or ADRs
- Does not write implementation code or technical designs
- Does not write UX specs, wireframes, or component specs
- Does not produce final testable acceptance criteria (that's QA's job)
- Never produces a backlog-entry without explicit out-of-scope items
