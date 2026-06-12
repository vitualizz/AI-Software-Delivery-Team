# Divergent Ideation — Researcher Specialist

## Purpose
Frame the problem and generate divergent candidate directions — WITHOUT
converging. This is the Researcher's first step: it opens the solution space
wide so the later steps have real options to assess and narrow. Deliberately
generative, never selective.

## Inputs
- `inputs: []` — there are no upstream artifacts; the raw problem statement from
  the user is the only source.
- Memory context from `context-recall` (injected inline — prior discovery,
  similar problem framings, and decisions enrich the framing but never constrain
  it).

## Context budget
Raw problem statement + injected recall context: max ~600 tokens. Keep the
framing tight so the generative work has room.

## Processing
1. Restate the problem as a `problem_framing` — one or two sentences capturing
   the pain or opportunity in neutral terms, NOT a solution.
2. Generate at least one `Idea` (aim for several when the space allows). Each
   idea is a distinct direction, not a variation in wording.
3. Do NOT converge. Do NOT rank, score, or pick a favourite — that is the job of
   `feasibility-scan` and `discovery-brief`. Mixing convergence in here collapses
   the space prematurely.
4. If the problem description is empty or too thin to frame, ask ONCE for a
   clearer statement — never fabricate a problem to ideate against.

Each `Idea.id` is a **stable snake_case slug** that names the direction (e.g.
`idea_inline_clarify`). Downstream steps reference ideas by this id as a foreign
key, so it must be stable and unique within this artifact.

## Output
Produces: `researcher/ideation`

Persist via mem_save under the output_topic_key in workflow.yaml; return envelope.

Schema:
```yaml
payload:
  problem_framing: ""         # neutral one/two-sentence statement of the pain or opportunity
  ideas:
    - id: ""                  # stable snake_case slug, e.g. idea_inline_clarify
      what: ""                # the direction in one sentence
      why: ""                 # why it could address the problem_framing
      theme: ""               # optional grouping label
  open_items: []
```

**Idea** (value-object):
```yaml
id: ""        # stable snake_case slug — the FK used by feasibility-scan and discovery-brief
what: ""      # the candidate direction, one sentence
why: ""       # the rationale linking it to problem_framing
theme: ""     # optional — a grouping label when several ideas share a theme
```
