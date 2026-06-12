---
title: Researcher
description: Explores fuzzy problems and opportunities through divergent ideation and feasibility scanning, converging on one recommended direction — the specialist to bring in before requirements exist, when you don't yet know what to build.
order: 26
locale: en
---

# Researcher (`/asdt-researcher`)

> Explores fuzzy problems and opportunities through divergent ideation and feasibility scanning, converging on one recommended direction — the specialist to bring in before requirements exist, when you don't yet know what to build.

## What it does

The Researcher Specialist diverges before PM converges. It takes a fuzzy problem or opportunity and runs a structured discovery sequence: frame the problem and generate deliberately divergent candidate directions, assess each one with a feasibility verdict grounded in evidence, then converge on a single recommended direction packaged as a `discovery-brief`.

Two properties keep the contract honest: ideation is **generative, never selective** — the ideation step produces candidates without ranking them, so promising-but-unusual directions survive long enough to be assessed. And the brief recommends exactly **one** direction with explicit rationale; candidates that didn't make the cut are recorded as won't-do entries that seed PM's out-of-scope list.

The Researcher is analyst-only — it never writes the filesystem. Its one job is to turn "we don't know what to build" into a feasibility-grounded recommendation that PM can treat as a well-formed raw request.

## When to invoke it

- The problem or opportunity is fuzzy ("we're losing users somewhere", "costs feel too high")
- The direction is unclear — multiple plausible solutions exist and nobody has compared them
- You need a feasibility-grounded recommendation **before** writing requirements
- You're weighing build-vs-buy or approach-vs-approach trade-offs at the idea stage

## Pipeline position

The only **pre-PM** specialist in the pipeline — it runs before requirements exist. Its `discovery-brief` summary and recommended direction are rendered as prose and handed to `/asdt-pm` feature-intake as the raw request, so PM starts from an explored, feasibility-checked direction instead of a guess. Can run standalone when you only need structured exploration without proceeding to requirements.

## What it produces

**researcher/ideation** — the framed problem plus divergent candidate directions, unranked by design. **researcher/feasibility** — a green/yellow/red verdict per candidate with supporting evidence and effort estimates. **researcher/discovery-brief** — the converged recommendation: one direction, rationale, feasibility notes, and won't-do candidates.

Consumed by: **PM** (reads the discovery-brief as its raw request; `wont_candidates` seed the backlog-entry's out-of-scope list). At the trivial tier only the ideation artifact is produced.

## Common patterns

```
/asdt-researcher We're losing users during onboarding but don't know why
# → Fuzzy problem, needs framing and candidate directions before requirements
```

```
/asdt-researcher Native mobile app or PWA for offline support?
# → Competing directions, needs feasibility verdicts before committing
```

```
/asdt-researcher Explore ways to cut our infra costs
# → Open opportunity, needs divergent ideation before anyone picks a lane
```

## Limits — what it does NOT do

- Does not write requirements or user stories (that's PM's job)
- Does not write architecture decisions or ADRs
- Does not write implementation code or tests
- Never acts as a builder — analyst-only, never writes the filesystem
- Never replaces PM — its brief feeds PM intake, it doesn't skip it
- Ideation never ranks candidates — convergence happens only in the brief, which recommends exactly ONE direction
