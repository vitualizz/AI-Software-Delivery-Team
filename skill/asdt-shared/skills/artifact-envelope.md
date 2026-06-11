# Artifact Envelope — Shared Skill

## Purpose

Define the contract for writing valid ASDT artifacts. Every specialist MUST produce artifacts using this envelope structure. Missing any required field is a validation failure — do not write the artifact.

---

## Required Envelope Structure

Every ASDT artifact MUST conform to this YAML structure:

```yaml
schema_version: "1"
agent: {specialist-id}
change_id: {change}
created_at: {ISO 8601 timestamp, e.g. "2026-06-04T10:30:00Z"}
prompt_version: {version hash of the prompt fragments used}
input_refs:
  - {Engram topic_key of each artifact read: "{project}/{change}/{specialist}/{artifact-type}" — per ADR-011}
payload:
  # specialist-specific fields
  open_items: []  # always present; populated when expected inputs were absent
```

---

## Field Definitions

| Field | Type | Required | Description |
|---|---|---|---|
| `schema_version` | string | Yes | Always `"1"` for current schema |
| `agent` | string | Yes | The specialist ID (e.g. `"developer"`, `"security"`) — must match the specialist's `specialist-id` frontmatter field |
| `change_id` | string | Yes | The active change name (e.g. `"add-password-reset"`) |
| `created_at` | string | Yes | ISO 8601 UTC timestamp of when the artifact was written |
| `prompt_version` | string | Yes | Hash or version identifier of the prompt fragments active at write time (for drift detection) |
| `input_refs` | []string | Yes | List of Engram topic_keys (`{project}/{change}/{specialist}/{artifact-type}`, per ADR-011) for every artifact read as input. Filesystem knowledge files (e.g. `.asdt/knowledge/platform.yaml`) keep their relative-path form — the only non-topic-key allowance. Empty list `[]` when no inputs were read |
| `payload` | map | Yes | Specialist-specific content. Must contain at minimum an `open_items` key |

---

## The open_items[] Contract

`open_items` is a required field inside `payload`. It is an array of strings, each describing what was missing and what was assumed instead.

**When open_items is empty**: all expected inputs were present and valid.

**When open_items contains entries**: the specialist ran with degraded inputs. Each entry describes what was absent and what assumption was made.

Entry format:
```
"{artifact or resource} absent — {what was assumed or how the specialist proceeded}"
```

Examples:
```yaml
open_items:
  - "requirements-spec.yaml absent — proceeding with inferred scope from feature description"
  - "platform.yaml absent — conventions inferred from visible code patterns"
  - "ux-brief.yaml absent — no user flow constraints applied to implementation plan"
```

**Rule**: Never omit `open_items` from `payload`. Use an empty array `[]` when nothing is missing.

---

## Validation Rules

Before writing any artifact, validate:

1. `schema_version` must be the string `"1"` — not the number 1, not `"2"`, not empty
2. `agent` must be non-empty and must match the running specialist's ID
3. `change_id` must be non-empty
4. `created_at` must be a valid ISO 8601 timestamp (UTC preferred)
5. `prompt_version` must be non-empty (use a hash of the assembled prompt or a monotonic identifier)
6. `input_refs` must be present (empty list is valid when no inputs were read)
7. `payload` must be present and must contain `open_items`

If any validation rule fails, do NOT write the artifact. Record the failure and stop.

---

## Input Refs Convention

`input_refs` contains the Engram topic_key of each artifact read as input, in the canonical form `{project}/{change}/{specialist}/{artifact-type}` (per ADR-011 — Engram is the authoritative artifact store). Examples:

```yaml
input_refs:
  - my-app/add-password-reset/pm/requirements-spec
  - my-app/add-password-reset/ux-ui/ux-brief
  - knowledge/platform.yaml
```

Filesystem knowledge files under `.asdt/` (e.g. `.asdt/knowledge/platform.yaml`) are the ONLY non-topic-key allowance — reference them by their relative path from the `.asdt/` root, as in the last example above.

When a specialist expects an artifact but it is absent (degraded run), do NOT include the absent reference in `input_refs`. Record the absence in `open_items[]` instead.

---

## Complete Example

```yaml
schema_version: "1"
agent: developer
change_id: add-password-reset
created_at: "2026-06-04T10:30:00Z"
prompt_version: "sha256:a1b2c3d4"
input_refs:
  - my-app/add-password-reset/pm/requirements-spec
  - knowledge/platform.yaml
payload:
  complexity_estimate: M
  open_items:
    - "ux-brief.yaml absent — no user flow constraints applied to implementation steps"
  steps:
    - story_ref: "REQ-001"
      title: "Implement password reset request endpoint"
      files_to_create:
        - internal/auth/reset.go
      files_to_modify:
        - internal/auth/handler.go
      rationale: "New endpoint needed to accept reset requests and trigger email delivery"
```
