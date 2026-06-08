# Hardening Checklist — Security Specialist

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
Produce actionable recommendations from all security findings.
Apply the report shared skill. Every finding becomes a concrete action item.

## Inputs
- `security/stride-threats`: STRIDE threats with severity
- `security/owasp-findings`: OWASP findings with recommendations

Retrieve via mem_search + mem_get_observation by topic_key.

Apply context-extraction: from stride-threats keep Critical + High severity only.
From owasp-findings keep all findings sorted by severity.

## Context budget
stride-threats (Critical/High) + owasp-findings: max 1,500 tokens.

## Processing
Apply the `report` shared skill:
1. Deduplicate: merge overlapping findings from STRIDE and OWASP.
2. Prioritize: Critical first, then High, then Medium, then Low.
3. For each finding, write ONE concrete action item:
   - Not "fix authentication" — write "Add rate limiting to /login endpoint: max 5
     attempts per IP per 15 minutes, return HTTP 429 with Retry-After header"
4. Group by implementation effort: Quick wins (< 1h), Medium (1h-1 day), Significant (> 1 day).
5. Write a security posture summary: what is the overall risk level? What must be fixed
   before launch vs. what can be deferred?

## Output
Produces: `security-findings` (final) and `hardening-checklist` (final)

This step produces TWO final artifacts (single-artifact shape, unlike qa's
`quality-report` which only LOOKS compound — confirmed by reading this section
directly). Persist `security-findings` via mem_save under this step's
`output_topic_key` in workflow.yaml; persist the second final artifact
`hardening-checklist` under its own distinct per-type topic_key (see the NOTE
on this step's workflow.yaml entry — `{project}/{change}/security/hardening-checklist`,
which collides with neither the primary key nor any intermediate artifact name);
return envelope covering both persisted keys.

security-findings schema:
```yaml
payload:
  findings:
    - id: "SF-001"
      severity: "Critical|High|Medium|Low"
      title: ""
      description: ""
      cwe: ""
      recommendation: ""
  overall_risk: "Critical|High|Medium|Low"
  open_items: []
```

hardening-checklist schema:
```yaml
payload:
  quick_wins:       # < 1h effort
    - item: ""
      priority: ""
      finding_ref: ""
  medium_effort:    # 1h-1 day
    - item: ""
      priority: ""
      finding_ref: ""
  significant:      # > 1 day
    - item: ""
      priority: ""
      finding_ref: ""
  must_fix_before_launch: []
  can_defer: []
  open_items: []
```
