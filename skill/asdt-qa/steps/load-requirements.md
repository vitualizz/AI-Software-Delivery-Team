# Load Requirements — QA Specialist

## Purpose
Extract and normalize acceptance criteria from whatever upstream artifacts exist.
QA starts from the deliverables of other specialists — not from the raw request.

## Inputs
- Any available upstream artifacts (use artifact-loading shared skill):
  - `requirements-spec` (if Requirements specialist ran)
  - `ux-brief` (if UX/UI specialist ran)
  - `architectural-decision` (if Architect specialist ran)
  - `dev-implementation` (if Developer specialist ran)
  - Raw request (fallback if no upstream artifacts)

Note: this step's `inputs:` list in `workflow.yaml` is empty by design — it has
no prior `subagent`-produced QA artifact to retrieve; it is QA's first generative
step and reads directly from upstream specialists' artifacts (or the raw request
as fallback).

## Context budget
Extract only: user_stories/acceptance_criteria/scope from each artifact.
Max 200 tokens per artifact. Max 1,500 tokens total.

## Processing
For each available upstream artifact:
1. Extract all acceptance criteria or testable requirements.
2. Normalize to a common format: Given/When/Then (write it if not already in that form).
3. Tag each AC with its source artifact.
4. Flag any AC that is NOT testable as-is (ambiguous, unmeasurable, subjective).
5. Count: how many ACs were found? How many are testable vs. need clarification?

If no upstream artifacts exist: derive ACs from the raw request by inferring
what "done" would look like for a typical user. Mark all as "inferred".

## Output
Produces: `qa/ac-list`

Persist via mem_save under the output_topic_key in workflow.yaml; return envelope.

Schema:
```yaml
payload:
  acceptance_criteria:
    - id: "AC-001"
      given: ""
      when: ""
      then: ""
      source: ""         # which upstream artifact
      testable: true
      testability_issue: ""  # only if testable: false
  total_count: 0
  testable_count: 0
  inferred: false
  open_items: []
```
