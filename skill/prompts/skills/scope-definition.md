# Skill: Scope Definition

## Purpose

This skill fragment provides guidelines for defining explicit, unambiguous project scope. Apply these guidelines when producing the `scope`, `nfrs`, and `open_questions` sections of a requirements spec.

---

## Explicit In/Out Lists

Scope creep begins with ambiguity. Every spec must contain both:

- **`scope.in`**: A list of features, interactions, states, and behaviors that are explicitly within this change. Each item must be concrete enough that a developer can decide whether a given implementation task belongs here.
- **`scope.out`**: A list of features and behaviors that are adjacent to this change but explicitly excluded. Listing something as "out" prevents implicit assumption during development.

**Rule**: If a feature is not listed in `scope.in`, it is not in scope — even if it seems obviously related. If there is ambiguity about whether a feature is included, add it to `scope.out` explicitly, or move it to `open_questions`.

**Example — password reset feature**:

```yaml
scope:
  in:
    - Requesting a password reset via registered email address
    - Receiving a time-limited reset link by email
    - Setting a new password via the reset link
    - Invalidating the reset link after use or expiry
    - Error handling for unregistered email addresses
  out:
    - Two-factor authentication during reset
    - SMS-based password reset
    - Admin-initiated password reset on behalf of a user
    - Password strength enforcement (tracked separately)
    - Social login account recovery
```

---

## Dependency Identification

Before finalizing scope, scan for dependencies that the feature implicitly relies on:

- **Infrastructure**: Does this feature require an email sending service, a background job processor, a new database table?
- **Upstream features**: Does this feature require another feature to exist first (e.g. user registration before password reset)?
- **External systems**: Does this feature integrate with a third-party service (payment gateway, identity provider)?

Document discovered dependencies in `open_questions` if they are unresolved, or in `scope.in`/`scope.out` if they are settled.

---

## NFR Categories

Non-functional requirements (NFRs) are constraints on quality attributes. Check each category for applicability:

| Category | Example NFR |
|----------|------------|
| **Performance** | "Password reset emails must be delivered within 60 seconds of the request." |
| **Security** | "Reset tokens must be single-use and expire after 15 minutes." |
| **Accessibility** | "The reset form must be operable with keyboard-only navigation and screen readers (WCAG 2.1 AA)." |
| **Internationalisation (i18n)** | "Error messages must be translatable via the existing i18n system." |
| **Reliability** | "The feature must degrade gracefully if the email provider is unavailable." |
| **Data retention** | "Expired reset tokens must be purged within 24 hours." |

**Rules**:
- Only include NFRs that are directly implied by or highly relevant to the feature being specified.
- Do not invent NFRs for features that have no evident quality constraints.
- NFRs must be testable or measurable (not "it should be fast").

---

## Open Questions

Open questions are unresolved decisions that materially affect the design of this feature. They must be surfaced rather than assumed.

**When to add an open question** (not an assumption):
- The idea implies multiple valid implementation approaches with different tradeoffs.
- A business rule is unclear (e.g. "how long should reset tokens be valid?").
- An actor's permission level is ambiguous (e.g. "can guests trigger this flow?").
- A dependency is uncertain (e.g. "which email provider will be used?").

**Format**: Write each open question as a full interrogative sentence that a product manager or architect could answer directly.

**Example open questions**:
```yaml
open_questions:
  - "What is the expiry duration for reset tokens? (15 min, 1 hour, 24 hours?)"
  - "Should the system confirm whether the email is registered, or always show 'if registered, you will receive an email'?"
  - "Is rate-limiting on reset requests in scope for this iteration?"
```

**Do not** include questions that are already answered by the original idea or by obvious convention. Do not add open questions just to appear thorough — only list genuinely unresolved items.
