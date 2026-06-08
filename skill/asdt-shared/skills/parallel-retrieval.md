# Orchestrator Fetch-Once Cache & Injected-Input Contract

**This is the SINGLE canonical home of the injected-input contract.** Every
specialist `SKILL.md` launch section and every step file's EXECUTOR header
points here rather than restating this content. Do not duplicate this
explanation inline anywhere else — edit it here and every pointer stays correct.

## Who this applies to

- **Orchestrator** (you, when launching `subagent` steps): you own the fetch-once
  cache and the injection. Follow the Cache Ledger Rule and the Injection Format
  below before building any sub-agent prompt.
- **Sub-agent** (you, when you ARE the launched step executor): your declared
  inputs arrive ALREADY INJECTED in your prompt as `### INPUT {topic_key}` blocks.
  Your inputs are injected — do NOT fetch them. Consume the injected content
  directly; **do NOT call `mem_search`/`mem_get_observation` to (re)retrieve
  your own declared inputs** — that work has already been done for you and
  repeating it wastes MCP round trips.

## Cache Ledger Rule (orchestrator)

Maintain a per-run map `topic_key -> resolved content`:

1. Before resolving any declared input, check whether its `topic_key` is already
   present in this run's ledger.
2. **First reference**: not present — call `mem_search(query: "{topic_key}", project: "{project}")`
   to get the ID, then `mem_get_observation(id)` for the full content (previews are
   NOT source material), and store the result under `topic_key` in the ledger.
3. **Every later reference**: present — reuse the stored value. Do **NOT** call
   `mem_search`/`mem_get_observation` again for that `topic_key`.
4. Net effect: each distinct `topic_key` is fetched **at most once per run**, no
   matter how many steps declare it as an input. `platform-summary` rides this
   same ledger — compute or retrieve it once, then serve every step that declares
   it from the cache.

## Injection Format (orchestrator builds, sub-agent consumes)

For each declared input, the orchestrator embeds ONE of these blocks directly in
the sub-agent's launch prompt — never a bare topic_key string for the sub-agent
to go fetch itself:

Resolved successfully:
```
### INPUT {topic_key}
{resolved full content}
```

Could not be resolved (`mem_search` returned nothing or `mem_get_observation` errored):
```
### INPUT {topic_key}: UNRESOLVED
(orchestrator could not fetch this input — record it in open_items and proceed)
```

The sub-agent reads these blocks as its source material. An `UNRESOLVED` block is
not a silent omission — it is an explicit instruction to record the gap and continue.

## Partial-Context Degradation (preserved, unchanged in spirit)

**Resolution failures MUST NOT cause all-or-nothing failure.** If the orchestrator
fails to resolve one declared input while others resolve fine:

- The successfully resolved inputs are injected and used normally.
- The failed input is injected as an `UNRESOLVED` block (see above).
- The sub-agent records the unresolved `topic_key` in `open_items` and proceeds
  with the partial context it has — it does not abort.

This is the same fallback contract the old self-fetch mandate guaranteed; only the
*mechanism* changed (orchestrator injects instead of sub-agent fetching), not the
degradation guarantee.

## N=1 Degradation

A step with exactly one declared input follows the exact same contract: the
orchestrator resolves that single `topic_key` through the cache (fetch once if
not already cached, reuse if it is) and injects either its `### INPUT` block or
its `UNRESOLVED` block. No special-casing — single-input steps are not exempt
from the fetch-once-and-inject rule, and they degrade the same way on failure.
