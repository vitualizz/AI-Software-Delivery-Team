# Accessibility Guidelines

## Purpose

WCAG 2.1 AA baseline requirements and implementation patterns. Applied during Step 6 of the UX/UI workflow.

## Baseline Standard

Target: **WCAG 2.1 Level AA**. Every new component must meet this baseline before handoff. Level AAA is aspirational — document AAA-relevant decisions in `open_items[]`.

## Semantic HTML Structure

- Use landmark elements: `<header>`, `<nav>`, `<main>`, `<aside>`, `<footer>`.
- One `<main>` per page.
- Headings (`h1`–`h6`) must form a logical outline — never skip levels.
- Use `<button>` for actions and `<a>` for navigation. Never use `<div>` as an interactive element.
- Use `<ul>` / `<ol>` for lists of items. Use `<dl>` for term/description pairs.
- Use `<table>` only for tabular data with proper `<thead>`, `<tbody>`, `scope` attributes.
- Form controls must be associated with a `<label>` via `for`/`id` or `aria-labelledby`.

## Keyboard Navigation Requirements

- All interactive elements must be reachable and operable via keyboard alone.
- Tab order must match visual reading order (left-to-right, top-to-bottom for LTR).
- No keyboard trap: pressing `Tab` or `Escape` must always allow the user to exit a component.
- Custom components that have no native keyboard behavior (e.g., custom dropdown, date picker) must implement ARIA keyboard patterns per the [APG](https://www.w3.org/WAI/ARIA/apg/).
- Keyboard shortcuts must not override browser or OS defaults.

## ARIA Labels

- Use `aria-label` when a visible text label does not exist or is insufficient.
- Use `aria-labelledby` to associate an element with an existing visible heading.
- Use `aria-describedby` for supplementary descriptions (e.g., input hints, error messages).
- Use `role` only when a native HTML element cannot provide the correct semantics.
- Do NOT add ARIA roles to elements that already have correct implicit roles (`<button role="button">` is redundant).
- Live regions: use `aria-live="polite"` for non-urgent updates; `aria-live="assertive"` only for critical alerts.

## Color Contrast Minimums

| Text type | Minimum ratio |
|-----------|--------------|
| Normal text (< 18pt / < 14pt bold) | 4.5:1 |
| Large text (≥ 18pt / ≥ 14pt bold) | 3:1 |
| UI components and graphical objects | 3:1 |
| Disabled states | No requirement (but note in open_items) |

Tools: [WebAIM Contrast Checker](https://webaim.org/resources/contrastchecker/), [Colour Contrast Analyser](https://www.tpgi.com/color-contrast-checker/).

Flag any color pair in `component-spec.yaml` that you cannot verify against `platform.yaml` tokens.

## Focus Management

- Visible focus indicator is required on all focusable elements. Default browser outline is acceptable; custom styles must meet 3:1 contrast ratio.
- Modal dialogs: on open, move focus to the first focusable element inside; on close, return focus to the trigger that opened it.
- Dynamic content (route changes, in-page navigation): move focus to the new page heading or a skip-link target.
- Toasts and notifications: do not steal focus; use `aria-live` regions instead.

## Images and Media

- Decorative images: `alt=""`.
- Informative images: `alt` must convey the same information as the image.
- Complex images (charts, diagrams): provide a long description via `aria-describedby` pointing to a text block.
- Video: captions required. Audio description required if visual content conveys meaning not in audio.
- Icon-only buttons: must have `aria-label` with a descriptive action.

## Component Checklist (per new component)

Before marking a component as complete in `component-spec.yaml`, verify:

- [ ] Operates with keyboard alone
- [ ] Color contrast passes for all text/UI pairs
- [ ] All images have appropriate `alt` text
- [ ] Focus order is logical
- [ ] Form inputs have associated labels
- [ ] Interactive state changes are communicated to screen readers
- [ ] No content depends solely on color to convey meaning
