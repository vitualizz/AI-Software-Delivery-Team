# Artifact Loading — Developer Skill

## Purpose

Guide the Developer specialist through retrieving existing upstream artifacts from Engram, extracting relevant fields from each artifact type, and recording absent artifacts in `open_items[]`.

This skill is invoked at Step 1 (Artifact Loading) of the Developer workflow.

---

## Retrieving Artifacts

1. Retrieve the upstream artifacts for this change from Engram:
   - Call `mem_search(query: "{project}/{change}", project: "{project}")` to list all available artifacts for the change.
   - For each result that matches an expected artifact type (requirements-spec, system-design, ux-brief), call `mem_get_observation(id: {id})` to retrieve the full content.
2. Apply the extraction rules by artifact type (see below).
3. If no results are returned by `mem_search`, record in `open_items[]`:
   ```
   "No artifacts found in Engram for change '{change}' — proceeding with feature description only"
   ```
4. **Filesystem fallback (Engram outage only)**: If the Engram MCP server is unreachable, fall back to scanning `.asdt/artifacts/{change}/` for `.yaml` files. Treat any file found there as a best-effort substitute. Record in `open_items[]`:
   ```
   "Engram unavailable — loaded artifacts from filesystem fallback at .asdt/artifacts/{change}/; data may be stale"
   ```
   Do NOT use the filesystem path as a primary lookup strategy — Engram is authoritative.

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
