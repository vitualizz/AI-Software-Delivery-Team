# Edge Case Analysis Guidelines

## Purpose

Systematic discovery of edge cases before test cases are written. Applied during Step 3 of the QA workflow.

## Boundary Values

For any input with a defined range or limit, test:
- **Minimum valid value**
- **Maximum valid value**
- **One below minimum** (should fail)
- **One above maximum** (should fail)
- **Zero or empty** (is it valid? what does the system do?)
- **Negative values** (if the domain implies non-negative)

Example: if an API accepts `quantity` between 1 and 100:
- Test `quantity = 1`, `quantity = 100` (boundaries, must pass)
- Test `quantity = 0`, `quantity = 101` (outside bounds, must fail)
- Test `quantity = -1` (negative, must fail)
- Test omitting `quantity` entirely (missing required field, must fail)

## Null / Empty / Missing Inputs

For every input field, explicitly test:
- `null` or `nil` — is it treated as absent or as an explicit null?
- Empty string `""` — is it equivalent to absent, or a distinct state?
- Empty collection `[]` — does the code handle zero-length gracefully?
- Whitespace-only string `"   "` — is it trimmed and treated as empty?
- Missing key in a JSON/YAML payload — does the parser use the default or error?

Document the expected behavior for each case in the test case `then` field — never assume.

## Concurrent Access Scenarios

When a feature involves shared mutable state:
- **Double submit**: user clicks submit twice in quick succession. What happens on the second request?
- **Optimistic locking conflict**: two users edit the same record simultaneously. Which wins? Does the loser get an error?
- **Race on creation**: two requests create a resource with the same unique key at the same time. Is only one created?
- **Read-during-write**: a reader fetches a record while a writer is partially updating it. Is a partial state ever visible?

Flag these as integration or E2E test cases. Unit tests cannot catch concurrency issues.

## Permission Edge Cases

- A user with the minimum required role performs the action — should succeed.
- A user with one role below the required role attempts the action — should be denied.
- An unauthenticated request — should be rejected with `401`.
- A user performing an action on a resource they own vs. one owned by another user.
- An admin performing an action on behalf of another user — is this allowed?
- A token that was valid but has been revoked mid-session.

## Network Failure Scenarios

- The downstream service times out. Does the feature degrade gracefully?
- The downstream service returns a 500. Is the error propagated correctly?
- The connection is dropped mid-request. Is the operation atomic, or can it leave state inconsistent?
- Retry logic: if the client retries, is the operation idempotent?

Document expected behavior: fail-open, fail-closed, or retry with backoff.

## Large Data Sets

- A list that contains 0 items — empty state rendered correctly?
- A list that contains 1 item — singular vs. plural handling?
- A list that contains the maximum paginated page size exactly.
- A list that exceeds the maximum page size — is pagination applied?
- A search query that returns no results.
- A search query that returns thousands of results — performance and UI handling.

## Timezone and Locale Edge Cases

- Timestamps stored in UTC but displayed in local time — is the conversion correct?
- Daylight saving time transitions — does a "next day" calculation jump correctly?
- Locales with non-Gregorian calendars — is the date picker functional?
- Number formatting: `1,000.00` vs. `1.000,00` — does the parser handle both?
- RTL languages: does the layout flip correctly?

## Edge Case Documentation Format

For each edge case in `test-plan.yaml`:
```yaml
- id: "TC-0XX"
  title: "Empty cart checkout attempt"
  given: "A user has an empty shopping cart"
  when: "They click the checkout button"
  then: "The checkout button is disabled and a message 'Add items to your cart to continue' is displayed"
  type: "e2e"
```

Always document WHY this edge case matters — what could go wrong if it is not handled.
