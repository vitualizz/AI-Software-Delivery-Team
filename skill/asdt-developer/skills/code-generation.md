# Code Generation — Developer Skill

## Purpose

This skill provides guidelines for generating idiomatic, production-quality code. Apply these guidelines at Step 5 (Code Generation) of the Developer workflow when writing code snippets or implementation instructions.

Platform conventions are provided by the `platform-context` shared skill. Always check platform context before generating code — the conventions there take precedence over these defaults.

---

## Match Existing Conventions

Before writing any code, review the platform context loaded at Step 2. Then:

- **Naming**: use the casing style already dominant in the project (snake_case, camelCase, PascalCase per layer). Never introduce a new casing style without justification.
- **File structure**: place new files in directories consistent with the existing layout. A feature in `src/features/` belongs there, not in `src/utils/`.
- **Imports**: prefer the project's established import style. If the project uses named exports, use named exports. If it uses default exports, match that pattern.
- **Libraries**: use libraries already in the dependency manifest (`go.mod`, `package.json`, etc.) before introducing new ones. Every new dependency is a cost.

If platform context is absent, infer conventions from the visible code and note the inference in the step's rationale.

---

## Prefer Composition Over Inheritance

- Favour small, focused interfaces over large class hierarchies.
- Embed/compose structs or objects instead of inheriting behaviour.
- In Go: use struct embedding and interface satisfaction. In TypeScript: use composition functions and `interface` declarations. In Python: use protocol classes and dataclasses.

---

## Early Return / Early Exit

Validate preconditions at the top of the function and return early on failure. This keeps the happy path unindented and readable.

```go
// Good
func ProcessOrder(order Order) error {
    if order.ID == "" {
        return errors.New("order ID is required")
    }
    if order.Total < 0 {
        return errors.New("order total cannot be negative")
    }
    // happy path here — not nested
    return process(order)
}

// Avoid
func ProcessOrder(order Order) error {
    if order.ID != "" {
        if order.Total >= 0 {
            return process(order)
        }
    }
    return errors.New("invalid order")
}
```

---

## Explicit Interfaces

Define interfaces at the point of use (the consumer), not at the point of definition (the implementor). Small, focused interfaces are easier to mock and satisfy.

```go
// Good — defined where used, minimal surface
type OrderStore interface {
    Save(ctx context.Context, order Order) error
    FindByID(ctx context.Context, id string) (Order, error)
}

// Avoid — large interface defined at the implementor
type Database interface {
    SaveOrder(...)
    FindOrder(...)
    SaveUser(...)
    FindUser(...)
    // 20 more methods...
}
```

---

## Dependency Injection — No Global State

All dependencies (database connections, HTTP clients, configuration, loggers) must be injected via constructor parameters or function arguments. Never use package-level variables for mutable state.

```go
// Good
type OrderService struct {
    store  OrderStore
    mailer Mailer
    clock  func() time.Time
}

func NewOrderService(store OrderStore, mailer Mailer) *OrderService {
    return &OrderService{store: store, mailer: mailer, clock: time.Now}
}

// Avoid
var globalDB *sql.DB  // injected via init() or set during startup
```

---

## Small, Focused Functions

A function should do one thing. If a function has more than one level of abstraction (e.g. it validates input AND formats output AND persists to storage), split it.

Rule of thumb: if you can describe what the function does with a single "and"-free sentence, its scope is appropriate.

---

## No Magic Numbers or Strings

All significant constants must be named. Magic numbers and inline strings make code fragile and unsearchable.

```go
// Good
const (
    maxRetries     = 3
    resetTokenTTL  = 15 * time.Minute
    emailSubject   = "Reset your password"
)

// Avoid
time.Sleep(900000000000)   // what is this?
if attempts > 3 { ... }   // where does 3 come from?
```

---

## Error Handling

- In Go: always check and wrap errors. Use `fmt.Errorf("context: %w", err)` for wrapping. Never ignore an error with `_`.
- In TypeScript: use `Result` types or explicit `try/catch` with typed error discrimination. Never swallow errors silently.
- In Python: use typed exceptions and do not use bare `except:` clauses.

Errors must carry enough context to locate the failure without a debugger.

---

## Code Snippet Format

When producing code snippets in an `implementation-plan.yaml`:
- Show the complete, relevant code unit (function, type, method) — not a fragment that requires the reader to guess the surrounding context.
- Include package/module declarations when introducing a new file.
- Add comments for non-obvious logic only. Do not comment what the code obviously does.
- If the snippet is an excerpt from a larger file, add `// ... (existing code)` markers to show where the snippet fits.
- Set `file` to the relative path from the project root (e.g. `internal/auth/reset.go`).
- Set `language` to the file's language identifier (e.g. `go`, `typescript`, `python`).
