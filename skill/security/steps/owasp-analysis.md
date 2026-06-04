# OWASP Analysis — Security Specialist

## Purpose
Check the attack surface against the OWASP Top 10 (2021).
Systematic coverage prevents the "I forgot to check X" failure mode.

## Inputs
- `security/attack-surface`: entry points, trust boundaries, data flows

Extract: entry_points[], data_flows[].vulnerabilities_noted.

## Context budget
security/attack-surface (entry points + data flows): max 1,200 tokens.

## Processing
Check each OWASP Top 10 category against the attack surface. For each, state:
APPLICABLE / NOT APPLICABLE (with brief reason), and if applicable: MITIGATED / AT RISK.

A01 — Broken Access Control
- Check: authorization on every entry point, IDOR potential, path traversal

A02 — Cryptographic Failures
- Check: sensitive data in transit (TLS?), at rest (encryption?), in logs (redacted?)

A03 — Injection
- Check: SQL injection (parameterized queries?), XSS (output encoding?),
  command injection (shell calls?), template injection

A04 — Insecure Design
- Check: security not bolted on — was threat modeling done early?

A05 — Security Misconfiguration
- Check: default credentials, exposed debug endpoints, unnecessary features enabled,
  error messages exposing stack traces

A06 — Vulnerable Components
- Check: known CVEs in dependencies (from platform-summary stack)

A07 — Authentication Failures
- Check: session management, password policies, multi-factor, account lockout

A08 — Software Integrity Failures
- Check: unsigned updates, insecure deserialization, CI/CD pipeline integrity

A09 — Logging Failures
- Check: are security events logged? Are logs tamper-resistant?

A10 — SSRF
- Check: any user-controlled URLs fetched server-side?

## Output
Produces: `security/owasp-findings`

Schema:
```yaml
payload:
  findings:
    - id: "OF-001"
      owasp_category: "A01|A02|...|A10"
      title: ""
      description: ""
      severity: "Critical|High|Medium|Low"
      cwe: ""          # CWE reference number if applicable
      recommendation: ""
  not_applicable: []   # OWASP categories not relevant with brief reason
  open_items: []
```
