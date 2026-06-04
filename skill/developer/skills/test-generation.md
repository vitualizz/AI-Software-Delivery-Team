# Test Generation — Developer Skill

## Purpose

This skill provides guidelines for generating high-quality, maintainable test cases. Apply these guidelines at Step 6 (Test Generation) of the Developer workflow when producing `test_snippets` for each implementation step.

The goal is tests that document intent, catch regressions, and are easy to extend — not tests that merely achieve coverage numbers.

---

## One Test Case Per Behavior

Each test case must verify exactly one observable behavior. Multiple assertions testing the same behavior in one case are fine; testing unrelated behaviors in one case is not.

- **Good**: "returns error when token is expired" verifies one failure mode.
- **Avoid**: "creates token, validates it, sends email, and marks it as used" — this is a scenario test disguised as a unit test.

---

## Table-Driven Tests

Group related test cases into a table (slice of structs). This pattern reduces boilerplate, makes the intent of each case explicit, and makes it trivial to add new edge cases.

```go
func TestValidateResetToken(t *testing.T) {
    tests := []struct {
        name    string
        token   ResetToken
        wantErr bool
        errMsg  string
    }{
        {
            name:    "valid token passes",
            token:   ResetToken{ID: "tok-1", ExpiresAt: time.Now().Add(15 * time.Minute)},
            wantErr: false,
        },
        {
            name:    "expired token returns error",
            token:   ResetToken{ID: "tok-1", ExpiresAt: time.Now().Add(-1 * time.Minute)},
            wantErr: true,
            errMsg:  "token expired",
        },
        {
            name:    "missing ID returns error",
            token:   ResetToken{ExpiresAt: time.Now().Add(15 * time.Minute)},
            wantErr: true,
            errMsg:  "token ID is required",
        },
    }

    for _, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {
            err := ValidateResetToken(tc.token)
            if tc.wantErr {
                if err == nil {
                    t.Fatal("expected error, got nil")
                }
                if !strings.Contains(err.Error(), tc.errMsg) {
                    t.Errorf("error %q does not contain %q", err.Error(), tc.errMsg)
                }
            } else if err != nil {
                t.Fatalf("unexpected error: %v", err)
            }
        })
    }
}
```

---

## Test Fixtures, Not Hardcoded Values

Extract repeated test setup into helper functions. Hardcoded values scattered across test cases make refactoring painful.

```go
// Good — fixture builder
func newValidResetToken(t *testing.T) ResetToken {
    t.Helper()
    return ResetToken{
        ID:        "test-tok-001",
        UserID:    "user-1",
        ExpiresAt: time.Now().Add(15 * time.Minute),
    }
}

// Usage
tok := newValidResetToken(t)
tok.ExpiresAt = time.Now().Add(-1 * time.Minute)  // mutate only what the test cares about
```

---

## Mock at the Boundary (Interfaces)

Never mock concrete types. Mock the interface the code depends on. Every injected interface from the `code-generation` skill is a seam for testing.

```go
// The code under test depends on this interface
type TokenStore interface {
    Save(ctx context.Context, token ResetToken) error
    FindByID(ctx context.Context, id string) (ResetToken, error)
}

// The mock implements the interface
type mockTokenStore struct {
    saveFn   func(ctx context.Context, token ResetToken) error
    findFn   func(ctx context.Context, id string) (ResetToken, error)
}

func (m *mockTokenStore) Save(ctx context.Context, token ResetToken) error {
    return m.saveFn(ctx, token)
}

func (m *mockTokenStore) FindByID(ctx context.Context, id string) (ResetToken, error) {
    return m.findFn(ctx, id)
}
```

---

## Test Behavior, Not Implementation

Tests should assert the OUTCOME observable from the public interface, not the internal mechanism:

- **Good**: assert that the returned artifact has the correct `schema_version` and `agent` fields.
- **Avoid**: assert that a specific private method was called, or that an internal counter was incremented.

If a test breaks when you rename a private variable without changing observable behaviour, the test is testing implementation.

---

## Meaningful Assertion Messages

Every failed assertion must produce a message that identifies what was expected, what was received, and in what context.

```go
// Good
if got.Status != "active" {
    t.Errorf("Status: got %q, want %q", got.Status, "active")
}

// Avoid
if got.Status != "active" {
    t.Error("wrong status")
}
```

---

## No Real Filesystem or Network in Unit Tests

- Use `t.TempDir()` for any test that writes files. The directory is automatically cleaned up.
- Use `fstest.MapFS` or `os.DirFS` on a temp directory for filesystem-backed tests.
- Never make real HTTP calls in unit tests. Use `httptest.NewServer` or a mock `Provider` interface.
- Never read from `$HOME` or system paths in tests — they produce non-reproducible results in CI.

---

## Coverage Requirement per Step

For every implementation step in the plan, generate at minimum:

1. One happy-path test (valid input → expected output)
2. One failure test per distinct error condition documented in the implementation step

When the implementation step produces an ASDT artifact, also test:
3. The artifact's envelope fields are all non-empty (`schema_version`, `agent`, `change_id`, `created_at`, `prompt_version`)
4. `open_items[]` is present in the payload (even if empty)

---

## Test Snippet Format in implementation-plan.yaml

When producing `test_snippets` in the plan:
- Set `file` to the test file path (e.g. `internal/auth/reset_test.go`).
- Include the full test function, not a fragment.
- Use the table-driven pattern for any step with more than one test case.
- Reference the implementation snippet's function/type names directly so the relationship is traceable.
