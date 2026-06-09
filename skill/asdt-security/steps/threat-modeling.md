# Threat Modeling — Security Specialist

> **EXECUTOR**: You are the sub-agent assigned this single step. Do the work
> described here yourself and return. You are NOT the orchestrator: do NOT call
> Agent/Task/delegate, do NOT run other steps. Your inputs are INJECTED in
> this prompt by the orchestrator — do NOT fetch them. See
> `../asdt-shared/skills/parallel-retrieval.md` for the injected-input
> contract; if an input is marked UNRESOLVED, record it in `open_items` and
> proceed. Persist your one output via
> `mem_save` under the `output_topic_key` declared for this step in `workflow.yaml`,
> then return a structured summary envelope (status, summary, output topic_key, open_items).

## Purpose
Identify threats using the STRIDE methodology. STRIDE is systematic —
it catches threats that intuition misses.

## Inputs
- `platform-summary`: tech stack, component library (to know attack surface shape)
- Any available upstream artifacts (artifact-loading — all optional):
  - `system-design`: API surface, service boundaries
  - `architectural-decision`: chosen approach and consequences
  - `dev-implementation`: what code is being written

Retrieve via mem_search + mem_get_observation by topic_key.

Extract only: API surface entries, service boundaries, data model entities.
If no upstream artifacts: derive from platform-summary and the request.

## Context budget
platform-summary + upstream summaries: max 1,500 tokens.

## Processing
Apply STRIDE to the feature/system under analysis:

**S — Spoofing Identity**: Can an attacker pretend to be someone they're not?
- Check: authentication mechanisms, session handling, token validation

**T — Tampering with Data**: Can an attacker modify data in transit or at rest?
- Check: input validation, CSRF protection, data integrity checks, database constraints

**R — Repudiation**: Can a user deny having performed an action?
- Check: audit logging, non-repudiation controls, immutable logs

**I — Information Disclosure**: Can an attacker access data they shouldn't?
- Check: authorization checks, error messages (do they leak internals?), data exposure in APIs

**D — Denial of Service**: Can an attacker make the system unavailable?
- Check: rate limiting, resource exhaustion, input size limits, expensive operations

**E — Elevation of Privilege**: Can a low-privilege user gain higher privileges?
- Check: authorization bypass, insecure direct object references, privilege escalation paths

For each STRIDE category, identify threats specific to THIS feature.
If a category has no relevant threats, state "No applicable threats identified" — do not skip.

## Output
Produces: `security/stride-threats`

Persist via mem_save under the output_topic_key in workflow.yaml; return envelope.

Schema:
```yaml
payload:
  scope: ""    # what was analyzed (feature name + available context)
  threats:
    - id: "T-S-001"  # T-{stride-letter}-{number}
      stride_category: "S|T|R|I|D|E"
      title: ""
      description: ""
      affected_components: []
      severity: "Critical|High|Medium|Low"
  threat_count: 0
  open_items: []
```
