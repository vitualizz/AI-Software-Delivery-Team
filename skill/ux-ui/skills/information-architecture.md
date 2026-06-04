# Information Architecture Guidelines

## Purpose

Defines how content is structured and navigated within a feature. Applied during Step 3 of the UX/UI workflow.

## Content Hierarchy Principles

1. **Most-used actions surface first.** Primary CTA is always one tap or click away from the entry point.
2. **Group by user intent, not by data model.** Users think in tasks, not entities. Label sections by what the user is trying to do.
3. **Limit top-level sections to 5–7 items.** Beyond that, consider grouping or progressive disclosure.
4. **Destructive or irreversible actions go last** in any list and require a confirmation step.

## Navigation Patterns

| Pattern | When to use |
|---------|-------------|
| **Flat navigation** | 3–5 top-level destinations; equal weight; quick switching expected |
| **Hierarchical navigation** | Content has clear parent/child relationships; user drills down then returns |
| **Step-based navigation** | Sequential task with defined start/end; back/next controls; no jumping ahead |
| **Hub-and-spoke** | Central dashboard with isolated sub-tasks; user always returns to hub |

Always document the full navigation path from the app entry point to the feature. Example: `Home → Settings → Notifications → Add Rule`.

## Progressive Disclosure

- Show only what the user needs for the current decision.
- Hide advanced options behind an "Advanced" toggle or secondary panel.
- Never load a form with more than 7 visible fields without a strong justification.
- Multi-step forms must show a progress indicator when there are 3 or more steps.

## Scannable Layouts

- Lead with the most important information in the first visual zone (top-left for LTR languages).
- Use consistent vertical rhythm: headings, body, supporting text at predictable spacing intervals.
- Lists of more than 10 items need either pagination, infinite scroll with a boundary, or search/filter.
- Empty states are content: they must explain why the space is empty and offer a next action.

## Labeling Conventions

- Use **verb + noun** for action labels: "Add Rule", "Delete Account", "Export Report".
- Use **nouns** for navigation labels: "Settings", "Dashboard", "Reports".
- Avoid jargon. If a technical term is unavoidable, add a tooltip or inline description.
- Consistency over cleverness: use the same label for the same concept everywhere.

## Data Relationships

When documenting data relationships in `ux-brief.yaml`:
- List entities the user will see or modify (e.g., `User`, `Order`, `Tag`).
- Note cardinality: one-to-one, one-to-many, many-to-many.
- Flag any entity that requires a separate loading state (async fetch).
