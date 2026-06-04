# Acceptance Criteria Guidelines

## Purpose

How to write and validate acceptance criteria. Applied during Step 2 of the QA workflow.

## Given/When/Then Format

All acceptance criteria must be written in Given/When/Then format:
- **Given**: the precondition — system state and context before the action
- **When**: the action or event that triggers the behavior
- **Then**: the observable outcome — what the system does in response

Example:
```
Given a registered user with a verified email address
When they submit the login form with valid credentials
Then they are redirected to the dashboard and a session token is issued
```

## Atomic Criteria

Each acceptance criterion must describe **one behavior**. If you find "and" in a Then clause that joins two different behaviors, split it into two AC.

Violating example: "Then the order is saved AND a confirmation email is sent AND the inventory is decremented."
Correct: three separate AC.

Exception: related outcomes of a single operation may stay together if they are always true together (e.g., "record is saved AND audit log entry is created" — these are one atomic concern).

## Measurable Outcomes

Vague language is not acceptable in acceptance criteria:

| Vague | Measurable replacement |
|-------|----------------------|
| "loads quickly" | "page renders within 200ms at p95 under 100 concurrent users" |
| "displays correctly on mobile" | "renders without horizontal scroll at 320px viewport width" |
| "handles large files" | "accepts files up to 50MB without timeout or memory error" |
| "appropriate error message" | "displays 'Invalid email format' when the email field fails RFC 5322 validation" |

Always require a specific threshold, count, or observable state change.

## Negative Cases

Every functional AC must be paired with at least one negative case:
- What happens when the precondition is NOT met?
- What happens when the input is invalid?
- What happens when a dependency is unavailable?

Negative cases are written as separate AC entries with their own Given/When/Then.

## Explicit Non-Goals

Each AC set should include a non-goals section stating what is explicitly out of scope. This prevents scope creep during implementation and testing.

Example non-goal: "Out of scope: rate limiting on the login endpoint. Tracked separately in TICKET-456."

## Validation Checklist

For each acceptance criterion, verify:
- [ ] Written in Given/When/Then format
- [ ] Covers exactly one behavior (atomic)
- [ ] Outcome is measurable with a specific threshold or observable state
- [ ] Negative case exists for each functional AC
- [ ] No ambiguous language (fast, easy, appropriate, etc.)
- [ ] Non-goals documented for the AC set

## AC Gap Classification

When flagging an AC gap in `test-plan.yaml`, classify it:

| Gap type | Description | Blocking? |
|----------|-------------|-----------|
| `untestable` | Criterion cannot be observed by any test | Yes — must be rewritten |
| `ambiguous` | Multiple valid interpretations exist | Yes — must be clarified |
| `incomplete` | Missing precondition, action, or outcome | Yes — must be completed |
| `missing-negative` | No failure path specified | No — add negative AC |
| `missing-nonfunc` | Performance/security threshold not stated | Depends on criticality |
