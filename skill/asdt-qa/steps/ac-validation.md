# AC Validation — QA Specialist

## Purpose
Critically review each acceptance criterion for quality. Surface problems BEFORE writing tests.
Bad ACs produce bad tests — fix the AC first.

## Inputs
- `qa/ac-list`: normalized acceptance criteria

Extract: acceptance_criteria[].

## Context budget
qa/ac-list: max 1,500 tokens.

## Processing
For each AC in ac-list, check these quality criteria:
1. ATOMIC: does this AC test exactly ONE behavior? (If it tests two things, split it.)
2. MEASURABLE: is the outcome observable and quantifiable? ("fast" is not measurable, "< 200ms" is.)
3. INDEPENDENT: can this AC be tested without relying on the outcome of another test?
4. COMPLETE: does it cover the success path AND specify error behavior?
5. UNAMBIGUOUS: would two different engineers write the same test for this AC?

For failing ACs:
- Write a corrected version of the AC.
- Tag it as "needs-revision" with your correction as a suggestion.
- Do NOT silently drop failing ACs — surface them as open_items.

## Output
Produces: `qa/ac-gaps`

Persist via mem_save under the output_topic_key in workflow.yaml; return envelope.

Schema:
```yaml
payload:
  validated_criteria:
    - id: "AC-001"
      status: "valid|needs-revision|invalid"
      issue: ""           # only if not valid
      suggested_revision: ""  # corrected AC text
  gap_count: 0
  open_items: []    # ACs that are invalid and need input from upstream specialists
```
