# Skill: Test Writing

## Purpose

This skill fragment provides guidelines for writing testable code and high-quality test cases. Apply these guidelines when designing implementation steps and when producing code snippets that include tests.

---

## Table-Driven Tests

Group related test cases into a table (slice of structs). This pattern reduces boilerplate, makes the intent of each case explicit, and makes it trivial to add new edge cases.

```go
func TestValidateOrder(t *testing.T) {
    tests := []struct {
        name    string
        order   Order
        wantErr bool
        errMsg  string
    }{
        {
            name:    "valid order passes validation",
            order:   Order{ID: "ord-1", Total: 10.00},
            wantErr: false,
        },
        {
            name:    "missing ID returns error",
            order:   Order{Total: 10.00},
            wantErr: true,
            errMsg:  "order ID is required",
        },
        {
            name:    "negative total returns error",
            order:   Order{ID: "ord-1", Total: -1},
            wantErr: true,
            errMsg:  "order total cannot be negative",
        },
    }

    for _, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {
            err := ValidateOrder(tc.order)
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

## Test One Behavior Per Case

Each test case must verify exactly one observable behavior. Multiple assertions testing the same behavior in one case are fine; testing unrelated behaviors in one case is not.

- **Good**: "returns error when ID is missing" verifies one failure mode.
- **Avoid**: "creates order, validates it, saves it, and sends confirmation email" — this is a scenario test disguised as a unit test.

---

## Use Test Fixtures, Not Hardcoded Values

Extract repeated test setup into helper functions or test fixture files. Hardcoded values scattered across test cases make refactoring painful.

```go
// Good — fixture builder
func newValidOrder(t *testing.T) Order {
    t.Helper()
    return Order{
        ID:    "test-ord-001",
        Total: 99.99,
        Items: []Item{{SKU: "WIDGET-1", Qty: 1}},
    }
}

// Usage
order := newValidOrder(t)
order.Total = -1  // mutate only what the test cares about
```

---

## Mock at the Boundary (Interfaces)

Never mock concrete types. Mock the interface the code depends on.

```go
// The code under test depends on this interface
type OrderStore interface {
    Save(ctx context.Context, order Order) error
}

// The mock implements the interface
type mockOrderStore struct {
    saveFn func(ctx context.Context, order Order) error
}

func (m *mockOrderStore) Save(ctx context.Context, order Order) error {
    return m.saveFn(ctx, order)
}

// The test injects the mock
func TestOrderService_Create(t *testing.T) {
    store := &mockOrderStore{
        saveFn: func(_ context.Context, _ Order) error { return nil },
    }
    svc := NewOrderService(store, nil)
    // ...
}
```

Apply this pattern whenever the code-generation skill uses dependency injection — every injected interface is a seam for testing.

---

## Test Behavior, Not Implementation

Tests should assert the OUTCOME observable from the public interface, not the internal mechanism:

- **Good**: assert that the returned artifact has the correct `schema_version` and `agent` fields.
- **Avoid**: assert that a specific private method was called, or that an internal counter was incremented.

If a test breaks when you rename a private variable without changing observable behaviour, the test is testing implementation.

---

## Meaningful Assertion Messages

Every failed assertion must produce a message that identifies:
1. What was expected
2. What was actually received
3. In what context (test name already provides this via `t.Run`)

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

## Tests Must Not Touch the Real Filesystem or Network

- Use `t.TempDir()` for any test that writes files. The directory is automatically cleaned up.
- Use `fstest.MapFS` or `os.DirFS` on a temp directory for filesystem-backed tests.
- Never make real HTTP calls in unit tests. Use `httptest.Server` or a mock `Provider` interface.
- Never read from `$HOME` or system paths in tests — they produce non-reproducible results in CI.

---

## Test the Happy Path AND the Failure Modes

For every public function, write at minimum:
1. One happy-path case (valid input → expected output)
2. One failure case per distinct error condition

Table-driven tests make this natural — each row is one scenario.
