# Platform Context — Shared Skill

## Purpose

Inject the project's detected platform knowledge into any specialist's context. This skill is consumed by every specialist at the start of their workflow to ground code generation, component suggestions, and design decisions in the project's actual conventions.

---

## How to Find platform.yaml

Walk up from CWD until you find `.asdt/knowledge/platform.yaml`. This is the same nearest-ancestor search used for `.asdt/` itself.

Path: `.asdt/knowledge/platform.yaml` relative to the resolved ASDT root.

---

## What to Inject

When `platform.yaml` is found, extract and inject the following fields into the specialist's context (summarized to under 500 tokens):

| Field | What to inject |
|---|---|
| `detected_stack` | Languages, frameworks, runtimes detected (e.g. "Go 1.22, no frontend framework") |
| `conventions.naming` | Naming conventions in use (e.g. "snake_case for files, PascalCase for exported Go types") |
| `conventions.file_structure` | Directory layout pattern (e.g. "internal/ for packages, cmd/ for binaries") |
| `design_fingerprint` | Architectural pattern in use (e.g. "Hexagonal architecture, ports in internal/") |

Do not inject the entire `platform.yaml` verbatim. Summarize only the fields relevant to the specialist's current step.

---

## Graceful Degradation

If `platform.yaml` does not exist:

1. Do NOT halt the specialist workflow.
2. Record the absence in the artifact's `open_items[]`:
   ```
   "platform.yaml absent — conventions inferred from visible code patterns"
   ```
3. Proceed using conventions inferred from the code visible in the current context (file naming, import style, directory layout).

If `platform.yaml` exists but is partially populated (some fields empty or missing), inject only the fields that are present. Do not halt or record an error for missing optional fields.

---

## Injection Format

Build the injection from only the fields that are actually present — omit a line entirely when its source field is empty or missing. Never emit a label with nothing after it (`Architecture: ` on its own conveys nothing and still costs tokens):

```
Stack: {detected_stack values, comma-separated}
Conventions: {naming style, if present}{ | file structure note, if present}
Architecture: {design_fingerprint, only if present}
```

`Conventions` joins its two parts with ` | ` only when BOTH are present. If only one is present, emit that one alone with no separator. If neither is present, omit the `Conventions` line too.

Fully-populated example:
```
Stack: Go 1.22, no frontend framework
Conventions: PascalCase exported types, snake_case files | internal/ for packages, cmd/ for binaries
Architecture: Hexagonal — ports in internal/, adapters in cmd/
```

Partially-populated example — e.g. a `platform.yaml` straight out of `asdt init`, which only runs bounded presence checks and intentionally leaves `conventions.naming` and `design_fingerprint` for a dedicated future analysis step:
```
Stack: Go
Conventions: cmd/ for binaries, internal/ for private packages
```

---

## Conditional: Project Context

> Load this section only if `.asdt/knowledge/project-context.yaml` exists.
> If the file is absent, skip this entire section silently.
> If `schema_version` != `"1"`, skip and note `project-context.yaml: schema_version mismatch, skipped` in open_items.

The following fields describe the structural and stylistic context of this project,
as detected by `asdt init`. Each field carries a `source` (detected | inferred | manual)
and a `confidence` (high | medium | low). Fields with an empty `value` are omitted.

**Monorepo**: {{ is_monorepo.value }}  *({{ is_monorepo.source }}, {{ is_monorepo.confidence }})*
**Test runner**: {{ test_runner.value }}  *({{ test_runner.source }}, {{ test_runner.confidence }})*
**Naming style**: {{ naming_style.value }}  *({{ naming_style.source }}, {{ naming_style.confidence }})*
**Architectural style**: {{ architectural_style.value }}  *({{ architectural_style.source }}, {{ architectural_style.confidence }})*

When writing code or tests, treat `detected/high` fields as authoritative conventions.
Treat `inferred/medium` fields as likely conventions — confirm before diverging.
Treat `manual` fields as user-declared — never override without explicit user approval.

---

## Usage Note

This skill is always loaded by specialists via their `shared-skills` list. The specialist does not need to duplicate this logic — it calls this skill at its Platform Analysis step.
