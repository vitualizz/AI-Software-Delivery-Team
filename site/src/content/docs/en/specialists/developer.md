---
title: Developer
description: Turns specs and designs into working code — implementation plans, production code, and test suites — the specialist to bring in once the shape of the solution is settled and it's time to build it.
order: 22
locale: en
---

# Developer (`/asdt-developer`)

> Turns specs and designs into working code — implementation plans, production code, and test suites — the specialist to bring in once the shape of the solution is settled and it's time to build it.

## What it does

The Developer Specialist transforms existing requirements, UX specs, and architecture decisions into a concrete implementation. It reads the codebase first (always), defines what will and won't be built, chooses the technical approach, breaks work into atomic tasks, and either produces an implementation plan or writes files directly to the host repository.

Two operating modes gate actual file writes. In **plan-only mode** (default), the specialist produces code as snippets in the knowledge base — no host repo changes. In **writing mode**, declared file targets are resolved and validated before any write happens — if a needed path is outside the declared targets, it stops and reports the issue rather than freelancing a write.

`explore` and `spec` are irrenunciable — they always run regardless of complexity. The `test` step is conditional: it only runs when `strict_tdd: true` is set in `.asdt/config.yaml`.

## When to invoke it

- The shape of the solution is already settled (requirements, architecture, or UX are defined)
- You need a concrete implementation plan with ordered tasks and file-level targets
- You want production code written directly to the codebase (writing mode, with declared targets)
- You're picking up from a prior Architect or PM artifact stored in the knowledge base

## Pipeline position

Typically runs **after Architect** (reads `architectural-decision` + `system-design`) and produces the final `dev-implementation` consumed by QA. Can run standalone with just a request description — it will explore and spec the problem itself without upstream artifacts. At `simple` complexity, it bypasses the Architect entirely.

## What it produces

`developer/dev-implementation` — the consolidated implementation artifact. In plan-only mode it contains ordered code snippets. In writing mode it contains the file manifest of what was written and why.

Consumed by: **QA** (reads the implementation to validate against acceptance criteria and produce test cases).

## Common patterns

```
/asdt-developer Implement the user profile component
# → Shape defined by prior UX/UI and Architect work, now time to build
```

```
/asdt-developer Add CSV export to the reporting dashboard
# → Standalone request — will explore, spec, and implement without prior artifacts
```

```
/asdt-developer Implement based on the Architect's ADR
# → Picks up architectural-decision from the knowledge base automatically
```

## Limits — what it does NOT do

- Does not produce architecture decisions or ADRs
- Does not write UX specs, wireframes, or component specs
- Does not produce test plans or quality reports (test step only generates test code, not plans)
- Never writes files outside declared targets in writing mode — stops and reports instead
- `explore` and `spec` cannot be skipped regardless of complexity
