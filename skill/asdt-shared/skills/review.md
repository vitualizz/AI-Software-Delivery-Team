# Review — Shared Skill

## Purpose
Generic self-review step. Verify completeness, correctness, and convention compliance
of the work produced in this specialist run.

## Protocol
1. **Coverage check**: for each input artifact, verify every declared item is addressed
   in the output. List any gaps as `open_items`.
2. **Convention compliance**: cross-reference against `platform-summary`. Flag any
   output that violates naming conventions, file structure patterns, or library choices.
3. **Completeness check**: verify the output artifact has no empty required fields.
   Flag missing content as `open_items`.
4. **Self-correction**: if gaps are found, attempt to fill them before writing
   `open_items`. Only list items you cannot resolve in context.

## Output additions
Adds to the current step's payload:
- `review_notes: []` — observations (non-blocking)
- `open_items: []` — items that could not be resolved (blocking for next specialist)

## Scope limit
Review ONLY what is visible in the declared InputRefs for this step.
Do NOT re-run or re-reason about earlier steps.
Do NOT invent requirements not present in the inputs.
