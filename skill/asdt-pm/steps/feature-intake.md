# Feature Intake — PM Specialist

## Purpose
Extract and normalize the raw feature request into a structured problem statement.
This is the PM's first step — it transforms user language into a structured form
that downstream steps can process.

## Inputs
- Raw feature request from the user (always available — this step has `inputs: []`)
- Memory context from `knowledge-recall` (injected inline — check for prior similar features)

## Processing
1. Extract the core problem: what pain or gap is the user describing?
2. Identify the goal: what does "done" look like from the user's perspective?
3. Identify stakeholders: who benefits from or is affected by this feature?
4. Capture the trigger: why is this being requested now?
5. Flag ambiguities: list any terms or requirements that need clarification before user stories can be written.

If a prior similar feature exists in memory (from knowledge-recall): note it and highlight what is different about this new request.

## Output
Produces: `pm/feature-intake`

Persist via mem_save under the output_topic_key in workflow.yaml; return envelope.

Schema:
```yaml
payload:
  feature_name: ""
  problem_statement: ""
  goal: ""
  stakeholders: []
  trigger: ""              # why now?
  ambiguities: []          # open questions that need clarification before user stories
  prior_feature_ref: ""    # mem observation ID if a similar prior feature was found
```
