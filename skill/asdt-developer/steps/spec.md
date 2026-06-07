# Spec — Developer Specialist

> **EXECUTOR**: You are the sub-agent assigned this single step. Do the work
> described here yourself and return. You are NOT the orchestrator: do NOT call
> Agent/Task/delegate, do NOT run other steps. Retrieve every input named under
> `## Inputs` via `mem_search` (by its topic_key) then `mem_get_observation` —
> do not assume it is already in your context. Persist your one output via
> `mem_save` under the `output_topic_key` declared for this step in `workflow.yaml`,
> then return a structured summary envelope (status, summary, output topic_key, open_items).

## Purpose
Define exactly what needs to be built: in-scope, out-of-scope, and acceptance criteria.

## Inputs
- Request: the original feature description
- `developer/dev-exploration`: files to understand, patterns, open questions

Retrieve via mem_search + mem_get_observation by topic_key.

Extract from dev-exploration: `open_questions` (answer them here), `patterns_to_follow`.

## Context budget
Request + dev-exploration summary: max 1,500 tokens.

## Processing
1. Answer each `open_question` from the exploration step.
2. Define the scope boundary: what IS included and what is explicitly NOT included.
3. Write acceptance criteria (Given/When/Then format, max 5 criteria).
4. List technical requirements (NFRs: performance targets, error handling expectations).

Do NOT design the technical approach. Do NOT write code. Only define what to build.

## Output
Produces: `developer/dev-spec`

Persist via mem_save under the output_topic_key in workflow.yaml; return envelope.

Schema:
```yaml
payload:
  scope:
    in: []
    out: []
  acceptance_criteria:
    - given: ""
      when: ""
      then: ""
  technical_requirements: []
  open_questions_answered: {}  # map of question → answer
  open_items: []
```
