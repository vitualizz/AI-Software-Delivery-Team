# Skill: User Story Writing

## Purpose

This skill fragment provides guidelines for writing high-quality user stories that are clear, testable, and valuable. Apply these guidelines whenever producing user stories in a requirements spec.

---

## The INVEST Criteria

Every user story must satisfy all six INVEST properties:

| Property | Meaning | Failure example | Good example |
|----------|---------|-----------------|--------------|
| **Independent** | The story can be developed and tested without depending on another story being done first | "US-002 requires US-001's database table to exist" | Each story targets a distinct feature slice that can ship alone |
| **Negotiable** | The story captures a goal, not a specific implementation. The team decides HOW to implement it | "The reset link must use JWT with HS256 and expire in 15 minutes" | "Users can reset their password via email" |
| **Valuable** | The story delivers something a real user or stakeholder cares about. No story should exist only for technical reasons | "Create the database migration" | "Users can log in after resetting their password" |
| **Estimable** | The team can give a rough size estimate. Vague goals prevent estimation | "Make the app faster" | "Password reset emails arrive within 60 seconds" |
| **Small** | A single story fits in one iteration. Epics must be split | "Complete user authentication system" | "Users can request a password reset link" |
| **Testable** | The acceptance criteria can be verified by a QA engineer writing concrete test cases | "The UI should be intuitive" | "The reset form shows an error when the email is not registered" |

---

## Actor Identification

- Actors must be roles or personas, not system names. Use "registered user", "admin", "guest", "billing system" — not "the app", "the backend", or "it".
- When the idea implies multiple distinct actors interacting with the same feature (e.g. a user requests something, an admin approves it), write a separate story for each actor's interaction.
- Avoid "user" alone as an actor when a more specific role is evident from context.

---

## Story Format

Use this exact format for every story:

```
As a {actor},
I want {action — present tense, active voice},
so that {benefit — measurable outcome or value delivered}.
```

**Action guidance**:
- Use present tense: "I want to reset my password" not "I wanted" or "I will reset".
- Make the action specific: "I want to receive a password reset email" not "I want to deal with my password".
- Avoid describing the implementation in the action: "I want to click a button" → "I want to initiate a password reset".

**Benefit guidance**:
- The benefit should describe the outcome the actor gains, not what the system does.
- Weak: "so that the system updates the record" → Strong: "so that I can access my account again without contacting support".

---

## Splitting Epics

When an idea is broad (covers multiple user journeys, multiple actors, or a full feature area), split it into atomic stories before writing the spec:

1. Identify the distinct interactions (request, confirm, error, cancel, success).
2. Identify the distinct actors.
3. Write one story per interaction per actor.
4. If a story still feels too large after splitting, apply the vertical slice technique: split by data variant, access level, or delivery channel (e.g. "via email" vs "via SMS").

---

## Common Mistakes to Avoid

- **Compound stories**: "As a user, I want to reset my password AND update my profile..." → split into two stories.
- **Implementation stories**: "As a developer, I want to add a JWT middleware..." → not a user story; belongs in the implementation plan.
- **Benefit-free stories**: "As a user, I want a button." → always include a meaningful benefit.
- **Untestable criteria**: "The flow should be smooth." → replace with measurable behavior.
