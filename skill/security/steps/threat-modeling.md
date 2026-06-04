# Threat Modeling — Security Specialist

## Purpose
Identify threats using the STRIDE methodology. STRIDE is systematic —
it catches threats that intuition misses.

## Inputs
- `platform-summary`: tech stack, component library (to know attack surface shape)
- Any available upstream artifacts (artifact-loading — all optional):
  - `system-design`: API surface, service boundaries
  - `architectural-decision`: chosen approach and consequences
  - `implementation-plan`: what code is being written

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
