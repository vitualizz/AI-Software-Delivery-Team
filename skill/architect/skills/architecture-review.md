# Architecture Review Guidelines

## Purpose

How to review existing architecture before making a change. Applied during Step 1 of the Architect workflow.

## Coupling Point Identification

Before proposing any structural change, map what is currently coupled to the affected area:

1. **Direct callers**: which packages or modules import the affected package? Use `rg` or the IDE's "Find Usages" to enumerate them.
2. **Data shape consumers**: which other components read or write the same data structures (DB tables, event schemas, API contracts)?
3. **Configuration dependents**: which config keys, environment variables, or feature flags does the current implementation consume?
4. **Test dependents**: which test files will break if the interface changes?

Document each coupling point in the `key_constraints` array in `system-design.yaml`.

## Blast Radius Analysis

Estimate the blast radius before proceeding:

| Blast radius | Definition | Action |
|---|---|---|
| **Contained** | Change affects ≤ 1 package; no external contracts change | Proceed; note in design |
| **Moderate** | Change affects 2–4 packages or one external contract | Document migration path in ADR |
| **Wide** | Change affects 5+ packages or multiple external contracts | Flag as High risk; require phased migration plan |
| **Cross-boundary** | Change affects a public API, event schema, or shared DB table | Flag as Critical risk; require backward-compat analysis |

## Backward-Compatibility Concerns

For any interface or data contract change:
- Can old and new versions coexist? If yes, document the coexistence window.
- Is there a version field in the schema? If not, note adding one in `open_items[]`.
- Are there clients outside this repository that consume the contract? If yes, escalate to `open_items[]`.

## Migration Path for Breaking Changes

When a breaking change is unavoidable:
1. Introduce the new interface alongside the old one (expand phase).
2. Migrate all callers to the new interface.
3. Delete the old interface (contract phase).

Document the expand–migrate–contract phases explicitly in the ADR consequences section.

## Review Checklist

Before writing the ADR, verify:
- [ ] Coupling points enumerated
- [ ] Blast radius classified
- [ ] Backward-compat implications documented
- [ ] Migration path defined if breaking changes are present
- [ ] Key constraints captured in `system-design.yaml`
