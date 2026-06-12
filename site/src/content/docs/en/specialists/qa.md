---
title: QA Engineer
description: Builds the safety net before code ships — test plans, acceptance criteria validation, edge case analysis, and quality reports — the specialist to bring in when "it works on my machine" isn't good enough.
order: 23
locale: en
---

# QA Engineer (`/asdt-qa`)

> Builds the safety net before code ships — test plans, acceptance criteria validation, edge case analysis, and quality reports — the specialist to bring in when "it works on my machine" isn't good enough.

## What it does

The QA Specialist validates acceptance criteria, discovers edge cases systematically, defines the testing strategy across the pyramid (unit / integration / e2e), and produces a quality report with a ship-readiness verdict. It starts from whatever upstream artifacts exist — developer implementation, architecture decisions, or raw requirements — and normalizes them into a testable AC list before writing a single test case.

`ac-validation` always runs regardless of complexity — AC gaps must be surfaced, not silently ignored. A bad AC produces a bad test; the QA Specialist fixes the AC first, then generates test cases against the corrected version.

The QA Specialist is not trivial-eligible. At trivial complexity it falls back to `simple`, because no dependency-complete step set exists below that level.

## When to invoke it

- Code is ready for review and you need a quality gate before it ships
- Acceptance criteria exist but haven't been formally validated (atomicity, measurability, independence)
- You want systematic edge case coverage, not just happy-path tests
- You need a structured test plan that a developer can implement without guessing

## Pipeline position

Typically runs **after Developer** (reads `dev-implementation`) and is the final sign-off before code merges. Can run earlier — against PM's `backlog-entry` or Architect's `architectural-decision` — to catch AC quality issues before implementation starts. That early pass saves far more time than finding gaps after the code is written.

## What it produces

`test-plan` — the final quality artifact and sign-off. Contains: test summary (unit/integration/e2e counts), AC coverage percentage, any uncovered AC gaps, the quality verdict with rationale, and the full test case list.

Consumed by: **Developer** (to implement the test suite), used as the sign-off artifact before code merges.

## Common patterns

```
/asdt-qa Review the checkout flow for edge cases
# → Happy-path is tested but boundary conditions and error paths need coverage
```

```
/asdt-qa Validate acceptance criteria before implementation starts
# → Run QA on PM's backlog-entry to catch AC quality issues early
```

```
/asdt-qa Build a test plan for the authentication module
# → Full test pyramid strategy for security-sensitive code
```

## Limits — what it does NOT do

- Does not write implementation code
- Does not write architecture decisions or UX specs
- `ac-validation` cannot be skipped — AC gaps must always be surfaced
- `test-strategy` is a required input for test case generation at moderate+ — never omit it
- Test cases are specifications (Given/When/Then) — not executable code
