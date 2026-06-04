# Artifact Loading — Developer Skill

## Purpose

Guide the Developer specialist through scanning `.asdt/artifacts/{change}/` for existing upstream artifacts, extracting relevant fields from each artifact type, and recording absent artifacts in `open_items[]`.

This skill is invoked at Step 1 (Artifact Loading) of the Developer workflow.

---

## Scanning the Artifacts Directory

1. Resolve the ASDT root (walk up from CWD to find `.asdt/`).
2. Look for files at `.asdt/artifacts/{change}/`.
3. List every `.yaml` file found. If the directory does not exist or is empty, record in `open_items[]`:
   ```
   "No artifacts found under .asdt/artifacts/{change}/ — proceeding with feature description only"
   ```

---

## Extraction Rules by Artifact Type

### requirements-spec.yaml

Fields to extract:

| Field | Where to find it | What to do with it |
|---|---|---|
| `user_stories` | `payload.user_stories[]` | List each story ID and summary; use as `story_ref` in implementation steps |
| `scope.in` | `payload.scope.in[]` | Constrain the implementation plan to this list |
| `scope.out` | `payload.scope.out[]` | Record any overlap as an `open_items[]` entry |
| `nfrs` | `payload.nfrs[]` | Surface relevant NFRs (performance, security) in the plan |
| `open_questions` | `payload.open_questions[]` | Carry unresolved questions forward to the plan's `open_items[]` |

### system-design.yaml (from Architect specialist)

Fields to extract:

| Field | Where to find it | What to do with it |
|---|---|---|
| `decisions` | `payload.decisions[]` | Use as architectural constraints; do not contradict them |
| `components` | `payload.components[]` | Map implementation steps to declared components |
| `api_contracts` | `payload.api_contracts[]` | Generate implementation steps that satisfy the declared API surface |
| `risk_items` | `payload.risk_items[]` | Surface as `open_items[]` entries when they affect implementation |

### ux-brief.yaml (from UX/UI specialist)

Fields to extract:

| Field | Where to find it | What to do with it |
|---|---|---|
| `user_flows` | `payload.user_flows[]` | Map each flow step to an implementation step |
| `component_specs` | `payload.component_specs[]` | Generate implementation steps for each declared component |
| `interaction_notes` | `payload.interaction_notes[]` | Use as implementation constraints for UI code |

---

## Absent Artifact Protocol

When an expected artifact is not found, do NOT stop or error. Follow this protocol:

1. Add a note to `open_items[]` describing what was absent and what was assumed:
   ```yaml
   open_items:
     - "requirements-spec.yaml absent — proceeding with inferred scope from feature description"
     - "system-design.yaml absent — no architectural constraints applied; flag complex decisions as open_items"
   ```

2. Continue with whatever context is available (platform.yaml, visible code, the feature description provided at invocation).

3. Mark implementation steps that depend on absent artifacts with a note in their `rationale`:
   ```
   rationale: "Inferred from feature description — no requirements-spec present to confirm story coverage"
   ```

---

## Summary Guideline

After loading all found artifacts, produce an internal summary (not written to the artifact) before proceeding to Step 3:

```
Loaded:
  - requirements-spec.yaml: {N} user stories, scope {in/out counts}
  - platform.yaml: stack={stack}, conventions={summary}

Missing:
  - system-design.yaml
  - ux-brief.yaml

open_items to carry forward:
  - "system-design.yaml absent — ..."
  - "ux-brief.yaml absent — ..."
```

This summary becomes the grounding context for Steps 3–7.
