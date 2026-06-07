# Parallel Retrieval Mandate

**CRITICAL**: `mem_search` returns 300-char PREVIEWS, not full content. You MUST call
`mem_get_observation(id)` for EVERY declared input. **Skipping this produces wrong output.**

**Run all searches in parallel** — do NOT search sequentially:

```
for each input in step's declared inputs:
    mem_search(query: "{input topic_key}", project: "{project}") → save ID
```

Then **run all retrievals in parallel**:

```
for each saved ID:
    mem_get_observation(id: {saved_id}) → full content (REQUIRED)
```

Do NOT use search previews as source material.

## Fallback (unchanged)

If any input fails to resolve (mem_search returns no result or mem_get_observation errors),
note it in `open_items` and proceed with available context.

**Parallel batching MUST NOT introduce all-or-nothing failure.** If one input fails while
others succeed, the successfully resolved inputs are used normally, and the failed input
is recorded in `open_items`. The step continues with partial context — it does not abort.

## N=1 Degradation

A step with exactly one declared input degrades to a single `mem_search` + `mem_get_observation`
pair (no behavioral change required for single-input steps).
