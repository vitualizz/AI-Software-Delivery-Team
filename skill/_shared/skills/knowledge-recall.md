# Knowledge Recall — Shared Skill

## Purpose

Before complex reasoning, query organizational memory for prior decisions so this
step builds on — and does not contradict — past work. Accumulating knowledge across
changes is what makes ASDT a knowledge system rather than a stateless code generator.

## When to Use

Any step that makes an architectural, UX, security, or quality **decision**:

- First step of any specialist (organizational context sets the baseline)
- Steps named `evaluate-approaches`, `decision-record`, `threat-modeling`, `quality-report`
- Any step whose output will influence other specialists downstream

## Protocol

1. **Formulate a query** from the step's topic — use 3-5 words from the change
   name and specialist role (e.g. `"auth session strategy"`, `"ux payment flow"`).
2. Call `mem_search(query)` yourself and treat the top results as the Organizational
   Context for this step.
3. **If Organizational Context is present**:
   - Treat prior decisions as defaults — deviate only with explicit rationale.
   - If your current direction **conflicts** with a prior decision, surface it
     in `open_items` as: `"conflicts with prior decision: {title}"`.
   - Do not blindly repeat prior decisions — evaluate whether they still apply.
4. **If absent**, proceed normally — memory recall is best-effort and non-blocking.

## Context Budget

Organizational Context block: **max 300 tokens** (top 3 records only).

Keep the injected context within this budget yourself.

## Output

No artifact produced by this skill. It only enriches context for the step it precedes.
