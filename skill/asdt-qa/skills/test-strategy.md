# Test Strategy Guidelines

## Purpose

Test pyramid principles and when to use each test type. Applied during Step 4 of the QA workflow.

## Test Pyramid

```
        /\
       /  \
      / E2E \        ← Few; slow; expensive; high confidence on user flows
     /--------\
    /Integration\    ← Some; test service boundaries and external deps
   /--------------\
  /   Unit Tests   \ ← Many; fast; cheap; test business logic in isolation
 /------------------\
```

Rules:
- Most tests should be unit tests.
- Integration tests cover only service boundaries — not re-testing what unit tests already cover.
- E2E tests cover only critical user journeys — not every feature.

## When to Write Unit Tests

Write a unit test when:
- The code contains conditional logic (if/switch).
- The code transforms data (parsing, formatting, calculation).
- The code validates input.
- The code implements a domain rule.
- The behavior must remain stable under refactoring.

Do NOT write unit tests for:
- Framework boilerplate (router registration, dependency wiring).
- Pure configuration (no logic, just values).
- Code that only delegates to another well-tested component.

## When to Write Integration Tests

Write an integration test when:
- The code crosses a process boundary (HTTP, RPC, message queue).
- The code interacts with a real persistence layer (DB query, file system).
- The behavior depends on the combined behavior of two or more components.
- Mocking would hide the real contract of an external dependency.

Keep integration tests focused on the boundary, not on business logic already covered by unit tests.

## When to Write E2E Tests

Write an E2E test when:
- The user flow spans multiple services or pages.
- The feature is critical to revenue or user retention.
- Regression of this flow would not be caught by unit or integration tests.

Limit E2E tests to the minimum set of critical paths. Each E2E test should take less than 30 seconds to run.

## Test Naming Conventions

Pattern: `{Unit/Context}_{Scenario}_{ExpectedOutcome}`

Examples:
- `UserService_CreateUser_ReturnsIDOnSuccess`
- `UserService_CreateUser_ErrorsOnDuplicateEmail`
- `LoginFlow_ValidCredentials_RedirectsToDashboard`

For table-driven tests (Go, Jest parameterized): name the case description in the table row, not the test function.

## Fixture Design

- Fixtures must be the minimum valid state for the test — no extra data.
- Use builder patterns or factory functions for creating test entities.
- Never share mutable state between tests.
- Name fixtures descriptively: `validUser`, `expiredToken`, `emptyCart` — not `user1`, `data`, `obj`.

## Test Isolation

- Each test sets up its own state and tears it down after.
- Tests must not depend on execution order.
- Tests must not depend on external services (use fakes/stubs for integration boundaries in unit tests).
- Database tests: use transactions rolled back after each test, or a per-test schema/database.

## Flaky Test Prevention

Common flaky test causes and mitigations:

| Cause | Mitigation |
|-------|------------|
| Time-dependent assertions | Use fixed clocks / mock time |
| Network calls in unit tests | Use stubs/mocks; never call real network |
| Race conditions | Use proper synchronization primitives; avoid sleep() |
| Order-dependent state | Reset state in teardown; use per-test isolation |
| Non-deterministic data | Use fixed seeds for random generators |
| Polling without timeout | Always set a deadline; fail if deadline exceeded |

A flaky test is worse than no test — it erodes trust in the entire suite. Fix or delete flaky tests immediately.

## Coverage Targets

| Layer | Target |
|-------|--------|
| Business logic (domain) | 90%+ |
| Application services | 80%+ |
| HTTP handlers / adapters | 70%+ |
| Integration tests | Focus on contract, not percentage |
| E2E | Critical paths only |

Coverage is a proxy, not a goal. 80% with meaningful tests beats 100% with trivial tests.
