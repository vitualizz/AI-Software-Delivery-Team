---
title: How It Works
description: End-to-end flow from feature request to shipped code — how ASDT orchestrates specialists through structured artifacts.
order: 5
locale: en
---

# How It Works

## The execution model

ASDT gives you structure without removing control. When you run a specialist, it orchestrates a focused sequence of steps — each one producing a single artifact, each one reading only what the previous step produced.

You invoke the specialists. ASDT never runs them automatically. That is intentional: every step where a human confirms a plan is a step where wrong assumptions get caught before they compound.

The `/asdt` advisor analyzes your request and suggests which specialists to involve and in what order. You confirm the plan, then run each command yourself.

## Knowledge flows forward automatically

Every specialist step produces one artifact — a structured document saved to the knowledge base with a stable key. The next specialist retrieves it by key. No manual context-passing. No copy-pasting between commands.

This means:

- **Specialists are decoupled.** The Developer reads the Architect's decision record as a document — not a shared variable, not a file import.
- **Artifacts survive sessions.** Run PM on Monday, continue with Architect on Thursday. The knowledge base holds the context.
- **Missing inputs degrade gracefully.** If an artifact is missing, the next specialist notes it in `open_items` and proceeds with whatever is available.

## Specialists adapt to complexity

Every specialist runs the right depth of steps for the complexity of the request. A quick bugfix runs fewer steps than a new authentication system. The table below shows the full step sequence at moderate complexity:

| Specialist | Steps |
|---|---|
| PM | `feature-intake` → `user-stories` → `scope-analysis` → `backlog-entry` |
| Architect | `load-constraints` → `evaluate-approaches` → `decision-record` → `system-design` → `risk-analysis` → `technical-handoff` |
| Developer | `explore` → `spec` → `design` → `tasks` → `implement` |
| QA | `load-requirements` → `ac-validation` → `edge-case-analysis` → `test-strategy` → `test-case-generation` → `quality-report` |
| Security | `threat-modeling` → `attack-surface` → `owasp-analysis` → `hardening-checklist` |
| UX/UI | `feature-brief` → `information-architecture` → `user-flows` → `component-mapping` → `responsive-strategy` → `ux-handoff` |

Steps run as isolated sub-agents — they don't share context with each other, which prevents earlier reasoning from leaking into later steps. Each step reads only its declared inputs, writes one artifact, and hands off.

## The human is always in the loop

ASDT enforces a soft gate at two points:

1. **After `/asdt`** — the pipeline advisor presents a routing plan and waits for confirmation before telling you which commands to run.
2. **Between specialists** — you choose when to run the next one. Nothing is automated.

This isn't a limitation. It's the design. AI-generated architecture decisions benefit from human review before a developer acts on them. QA plans benefit from human review before they define what "done" means. ASDT gives you the structure; you apply the judgment.
