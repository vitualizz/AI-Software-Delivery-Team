# Threat Modeling Guidelines

## Purpose

STRIDE methodology for identifying and documenting threats. Applied during Step 1 of the Security workflow.

## STRIDE Categories

### S — Spoofing

An attacker impersonates a legitimate user, system, or service.

**Detection patterns**:
- Authentication mechanisms that can be bypassed or forged
- Predictable session tokens or API keys
- Missing sender verification on webhooks or callbacks
- Trust in client-supplied identity claims without server-side validation

**Common examples**: stolen JWT, session fixation, forged request origin header.

**Key question**: "Can an attacker claim to be someone they are not?"

### T — Tampering

An attacker modifies data in transit, at rest, or during processing.

**Detection patterns**:
- Unencrypted data transmission (HTTP instead of HTTPS)
- Missing integrity checks on stored data
- User-controlled input that flows into data mutations without validation
- Mass assignment vulnerabilities (binding all request fields to a model)
- Log injection (user input written to logs without sanitization)

**Key question**: "Can an attacker modify data they should not be able to change?"

### R — Repudiation

A user performs an action and later denies doing it, with no way to prove otherwise.

**Detection patterns**:
- Missing audit logs for sensitive operations
- Audit logs that can be deleted or modified by the subject of the audit
- No immutable record of state transitions
- Financial or legal operations without timestamped, attributable records

**Key question**: "Can a user deny an action they performed, and would we be unable to prove otherwise?"

### I — Information Disclosure

Sensitive data is exposed to unauthorized parties.

**Detection patterns**:
- Verbose error messages containing stack traces, SQL, or internal paths
- API endpoints that return more data than the caller is authorized to see
- Insecure direct object references (IDOR): user can access another user's resources by changing an ID
- Secrets in logs, URLs, or client-side code
- Missing field-level access control on GraphQL or API responses
- Caching of authenticated responses at a shared layer

**Key question**: "Can an attacker read data they should not be able to see?"

### D — Denial of Service

An attacker makes the system unavailable to legitimate users.

**Detection patterns**:
- Unauthenticated endpoints with expensive operations (large queries, file processing)
- Missing rate limiting on public endpoints
- Unbounded loops or recursion triggered by user input
- Resource exhaustion via large file uploads without limits
- Regular expression patterns vulnerable to ReDoS (catastrophic backtracking)
- Missing timeouts on outbound calls

**Key question**: "Can an attacker exhaust resources and make the system unavailable?"

### E — Elevation of Privilege

An attacker gains capabilities or access they should not have.

**Detection patterns**:
- Missing authorization checks (AuthN without AuthZ)
- Horizontal privilege escalation: user A can perform actions on user B's resources
- Vertical privilege escalation: a low-privilege user gains admin capabilities
- Insecure deserialization that can execute arbitrary code
- Path traversal allowing access to files outside the intended directory
- Indirect object references that bypass role checks

**Key question**: "Can an attacker do more than their role allows?"

## Trust Boundaries

A trust boundary is any point where data crosses from a lower-trust to a higher-trust zone, or vice versa.

Common trust boundaries:
- **External → Application**: HTTP request from the internet
- **Application → Database**: SQL query with user-derived parameters
- **Service A → Service B**: internal RPC or HTTP call
- **Browser → Server**: form submission, API call from frontend
- **Admin UI → Application**: elevated-privilege operation

For each trust boundary, document: what crosses it, in which direction, and what validation occurs at the boundary.

## Threat Trees (Text-Based)

When a threat has multiple attack paths, document it as a tree:

```
Goal: Bypass authentication
├── Path 1: Steal valid session token
│   ├── XSS to extract cookie
│   └── Man-in-the-middle on HTTP endpoint
├── Path 2: Forge authentication token
│   ├── Weak JWT secret (brute-force)
│   └── Algorithm confusion (RS256 → HS256)
└── Path 3: Account takeover via credential stuffing
    └── No rate limiting on /login
```

Document the highest-risk paths in the `threats[]` array of `threat-model.yaml`.

## Data Flow Diagrams (Text-Based)

For each significant data flow, document:

```
[User Browser] --> (HTTPS) --> [API Gateway] --> [Auth Middleware]
    --> [OrderService] --> [PostgreSQL: orders table]
    --> [EmailQueue] --> [EmailWorker] --> [SendGrid API]
```

Label each arrow with: protocol, trust level (authenticated/unauthenticated), and data sensitivity.
