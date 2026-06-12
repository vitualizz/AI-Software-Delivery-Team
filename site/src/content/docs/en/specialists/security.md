---
title: Security
description: Hunts for the gaps an attacker would find first — threat models, OWASP reviews, hardening checklists — the specialist to bring in whenever auth, data handling, or external integrations are on the table, at any point in the pipeline.
order: 24
locale: en
---

# Security (`/asdt-security`)

> Hunts for the gaps an attacker would find first — threat models, OWASP reviews, hardening checklists — the specialist to bring in whenever auth, data handling, or external integrations are on the table, at any point in the pipeline.

## What it does

The Security Specialist performs threat modeling and security analysis using STRIDE and the OWASP Top 10. It maps the attack surface, identifies threats systematically, and produces a prioritized hardening checklist where every finding has a concrete, actionable mitigation — not "monitor it" or "add logging."

The critical invariant: **Security has no required predecessor.** It can run at any stage — on a fresh project with no prior artifacts, mid-development, or after launch. If upstream artifacts exist (architecture decisions, implementation), it reads them. If they don't, it works from the platform context and request alone, noting gaps in `open_items` and proceeding.

Depth is gated by `risk_surface`, not complexity. This is the only specialist where the question isn't "how complex is the feature?" but "how large is the attack surface?"

## When to invoke it

- Authentication, session management, or authorization is involved
- The feature handles or stores personally identifiable information
- External integrations, webhooks, or user-controlled URLs are present
- New API endpoints are being exposed publicly
- Any time before shipping to production when security hasn't been reviewed

## Pipeline position

**No required predecessor** — invoke at any point. For maximum impact, run it after the Architect produces `system-design` (Security can then analyze the API surface and service boundaries). For a quick threat model early in design, run it before architecture is finalized to surface design-level risks before they're baked in.

Its outputs (`security-findings` + `hardening-checklist`) are consumed by Developer and Architect to address mitigations.

## What it produces

Two final artifacts:

- **`security-findings`** — all findings with severity ratings (Critical/High/Medium/Low following CVSS-lite), CWE references, and concrete recommendations
- **`hardening-checklist`** — actionable items grouped by effort, with must-fix-before-launch vs. can-defer separation

Consumed by: **Developer** (to implement mitigations), **Architect** (to adjust design decisions that introduced structural risks).

## Common patterns

```
/asdt-security Audit the OAuth integration
# → External auth flow with token handling — high risk surface
```

```
/asdt-security Threat model the new payment webhook handler
# → User-controlled input hitting financial logic
```

```
/asdt-security Quick security pass before the v2 launch
# → No prior artifacts needed — runs from platform context alone
```

## Limits — what it does NOT do

- Does not write implementation code
- Does not produce architecture decisions or UX specs
- Does not produce test plans (though its findings inform what QA should cover)
- Every finding must have a concrete mitigation — "add monitoring" is not a mitigation
- Severity always follows CVSS-lite: Critical / High / Medium / Low — no other scale
