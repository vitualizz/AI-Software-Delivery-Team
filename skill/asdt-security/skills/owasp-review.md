# OWASP Top 10 Review Guidelines

## Purpose

OWASP Top 10 2021 checklist with detection patterns and remediation guidance. Applied during Step 3 of the Security workflow.

---

## A01:2021 — Broken Access Control

**Description**: Restrictions on what authenticated users can do are not properly enforced.

**Detection patterns**:
- Direct object references using predictable IDs (e.g., `/orders/42`) without ownership checks
- Accessing admin-only endpoints without role verification
- CORS misconfiguration allowing requests from untrusted origins
- Missing authorization check after authentication middleware
- Force-browsing to pages that are hidden but not protected

**Stack-specific concerns**:
- Go: missing middleware on router groups; `context.Value` abuse bypassing auth checks
- Node/Express: missing `next()`-based auth checks in middleware chains
- React: client-side route guards not backed by server-side checks (security theater)

**Remediation**: enforce authorization on every server-side handler; use a centralized authorization module; log access denials.

---

## A02:2021 — Cryptographic Failures

**Description**: Sensitive data is exposed due to missing or weak cryptography.

**Detection patterns**:
- HTTP (not HTTPS) for any data transmission
- Passwords stored as plain text or with weak hashing (MD5, SHA-1)
- Encryption keys hardcoded in source code or config files in version control
- Symmetric encryption without authenticated encryption (AES-CBC without MAC)
- Sensitive data (PII, tokens) in URLs, logs, or browser history

**Stack-specific concerns**:
- Go: use `bcrypt` or `argon2` for password hashing; use `crypto/tls` for TLS config; avoid deprecated `crypto/des`
- All stacks: secrets must come from environment variables or a secrets manager — never from committed files

**Remediation**: use strong password hashing (bcrypt cost >= 12); TLS 1.2+ only; encrypt PII at rest; rotate secrets regularly.

---

## A03:2021 — Injection

**Description**: User-supplied data is sent to an interpreter as part of a command or query.

**Detection patterns**:
- String concatenation to build SQL queries (classic SQLi)
- Dynamic command execution with user-supplied arguments (`exec.Command` with user input)
- Template injection (user input rendered into server-side templates)
- LDAP, XPath, NoSQL injection (same root cause: unsanitized input to a query language)
- Log injection (user input written to logs without sanitization)

**Stack-specific concerns**:
- Go: use `database/sql` parameterized queries; never `fmt.Sprintf` into SQL
- Node: use parameterized queries with pg/mysql drivers; avoid `eval()`
- All ORMs: use query builder methods, not raw SQL interpolation

**Remediation**: parameterized queries everywhere; never interpolate user input into queries; validate and sanitize on input.

---

## A04:2021 — Insecure Design

**Description**: Missing or insufficient security controls in the architecture and design phase.

**Detection patterns**:
- No threat model produced for the feature
- Business logic flaws: e.g., a discount coupon that can be applied multiple times
- Missing rate limiting design for sensitive operations
- No account lockout or CAPTCHA design for brute-forceable endpoints

**Remediation**: threat modeling before implementation; security requirements treated as first-class acceptance criteria.

---

## A05:2021 — Security Misconfiguration

**Description**: Security settings are defined, implemented, or maintained improperly.

**Detection patterns**:
- Default credentials unchanged
- Unnecessary features enabled (debug endpoints, verbose error pages in production)
- Missing security headers: `Content-Security-Policy`, `X-Frame-Options`, `X-Content-Type-Options`, `Strict-Transport-Security`
- Overly permissive CORS: `Access-Control-Allow-Origin: *` on authenticated endpoints
- Directory listing enabled on web server
- Stack traces returned to clients on 500 errors

**Remediation**: security headers on all responses; no debug endpoints in production; error responses never expose internals.

---

## A06:2021 — Vulnerable and Outdated Components

**Detection patterns**:
- Dependencies with known CVEs (check with `govulncheck`, `npm audit`, `trivy`)
- Unmaintained dependencies (last commit > 2 years, no security patches)
- Components running as root or with excessive permissions

**Remediation**: run dependency audit in CI; subscribe to security advisories for critical dependencies; automate dependency updates.

---

## A07:2021 — Identification and Authentication Failures

**Detection patterns**:
- Weak passwords permitted (no minimum length, complexity rules)
- No account lockout or rate limiting on login endpoint
- Predictable or short-lived session tokens
- Session not invalidated on logout
- Insecure "forgot password" flow (predictable tokens, tokens not expiring)
- Missing MFA for sensitive operations

**Stack-specific concerns**:
- Go JWT: verify `alg` header is expected algorithm; use `none` algorithm rejection; validate `exp` and `iss` claims
- All stacks: invalidate sessions server-side on logout (JWTs need a denylist or short expiry)

**Remediation**: rate limit authentication endpoints; use secure session management; enforce password policy; invalidate tokens on logout.

---

## A08:2021 — Software and Data Integrity Failures

**Detection patterns**:
- Insecure deserialization: deserializing untrusted data that can trigger code execution
- CI/CD pipeline that can be modified by external pull requests without review
- Software updates fetched over HTTP without integrity checks
- Plugins or extensions loaded from untrusted sources

**Remediation**: verify integrity of software updates; review CI pipeline permissions; avoid deserializing untrusted objects.

---

## A09:2021 — Security Logging and Monitoring Failures

**Detection patterns**:
- Authentication failures not logged
- Audit log for sensitive operations (data deletion, permission changes, payments) not present
- Logs that can be modified or deleted by application users
- No alerts for repeated failures or anomalous access patterns
- PII or secrets in log output

**Remediation**: log all authentication events, authorization failures, and sensitive mutations; store logs in append-only, off-system storage; alert on anomaly thresholds.

---

## A10:2021 — Server-Side Request Forgery (SSRF)

**Description**: The application fetches a remote resource using user-supplied URL.

**Detection patterns**:
- Webhook registration that accepts arbitrary URLs
- URL preview or screenshot generation features
- File import from URL
- Internal service URLs accessible from user-supplied input
- DNS rebinding attacks via user-controlled hostnames

**Remediation**: validate and allowlist URL schemes and hosts; block requests to private IP ranges (RFC 1918) and localhost; use a dedicated egress proxy.
