# User Stories — PM Specialist

## Purpose
Define user stories from the structured problem statement.
Each story captures WHO wants something, WHAT they want, and WHY they want it.

## Inputs
- `pm/feature-intake`

## Context budget
Extract: problem_statement, goal, stakeholders from intake.
Max 300 tokens.

## Processing
1. For each stakeholder in the intake, write user stories from their perspective.
2. Format: "As a [role], I want [action], so that [benefit]."
3. Assign each story a preliminary MoSCoW priority: Must / Should / Could / Won't.
4. For each story, write 1–3 high-level acceptance criteria. These are NOT Given/When/Then yet — QA formalizes them. Write them as plain English conditions: "User can X", "System prevents Y", "Error message shows when Z."
5. Estimate story size: small / medium / large (rough order of magnitude).
6. Flag dependencies between stories (a story that cannot be built before another).

## Output
Produces: `pm/user-stories`

Persist via mem_save under the output_topic_key in workflow.yaml; return envelope.

Schema:
```yaml
payload:
  user_stories:
    - id: "US-001"
      role: ""
      action: ""
      benefit: ""
      priority: "must | should | could | wont"
      size: "small | medium | large"
      acceptance_criteria:
        - ""               # plain English conditions — NOT Given/When/Then
      depends_on: []       # list of US IDs this story depends on
  total_count: 0
  must_count: 0
  open_items: []           # unresolved ambiguities from feature-intake that affect stories
```
