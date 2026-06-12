---
title: UX/UI Designer
description: Shapes how people actually experience the product — user flows, information architecture, component specs, responsive and accessibility strategy — the specialist to bring in before a single screen gets built.
order: 25
locale: en
---

# UX/UI Designer (`/asdt-ux-ui`)

> Shapes how people actually experience the product — user flows, information architecture, component specs, responsive and accessibility strategy — the specialist to bring in before a single screen gets built.

## What it does

The UX/UI Specialist transforms a feature brief into a structured UX specification that the Developer can implement without ambiguity. It identifies the primary actor and core problem first, then organizes content into an information architecture, maps the full interaction sequences (happy path + error paths + edge cases), catalogs which components to reuse, extend, or create from scratch, and — at complex tier — defines how each component behaves across breakpoints.

`ux-handoff` always runs regardless of complexity. That's the consolidation step that produces the two final artifacts consumed by Developer and Architect. `simple` and `moderate` tiers are intentionally identical — `responsive-strategy` is the only step that's gated to `complex`, because simple features rarely need explicit multi-breakpoint specs.

One hard dependency: `information-architecture` must run before `user-flows`. You cannot map interaction sequences before you know the content hierarchy and navigation path.

## When to invoke it

- A new screen, dialog, or feature-level UI needs to be designed
- User flows need to be mapped before architecture or implementation begins
- Component reuse strategy needs to be decided (extend existing vs. create new)
- Accessibility requirements need to be specified explicitly
- You want the Developer to receive a spec rather than infer the UX from the requirements

## Pipeline position

Works best **before Developer** — the `ux-brief` and `component-spec` are inputs the Developer reads to implement the UI correctly. Can run in parallel with Architect, since UX design and architecture decisions are largely independent. Running it after the Developer has already built a screen means the spec arrives too late to guide the implementation.

## What it produces

Two final artifacts:

- **`ux-brief`** — feature summary, primary actor, success criteria, user flows (happy path + decision points), information architecture
- **`component-spec`** — full component inventory: reused (with use case), extended (with changes needed), new (with reason, props, events, responsive behavior)

Consumed by: **Developer** (reads both to implement the UI), **Architect** (reads `ux-brief` to understand user flows when designing API contracts).

## Common patterns

```
/asdt-ux-ui Design the onboarding flow for new users
# → New multi-step UI — needs full IA + flows before any component work
```

```
/asdt-ux-ui Map the notification preferences screen
# → Existing UI pattern to extend — component-mapping will identify reuse opportunities
```

```
/asdt-ux-ui Spec the mobile layout for the dashboard
# → Complex tier with responsive-strategy for explicit breakpoint behavior
```

## Limits — what it does NOT do

- Does not write implementation code — only specifications and structure
- Does not produce architecture decisions or test plans
- Never proposes components inconsistent with the existing design system
- The generated UI must feel like it belongs to the existing application
- `information-architecture` cannot be skipped before `user-flows`
- `ux-handoff` always runs — consolidation is non-negotiable
